package query

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"dealscanner/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CashConvertersDiscoveredSummary struct {
	CanonicalURL   string
	ExternalItemID *string
	Title          string
	Price          float64
	Shipping       float64
	TotalPrice     float64
	ImageURL       string
}

type CashConvertersSummaryUpsertResult struct {
	ListingIDs        []string
	NewListingIDs     []string
	ChangedListingIDs []string
	HadPreviouslySeen bool
}

type CcDetailInput struct {
	CanonicalUrl string
	Title        string
	Price        float64
	Shipping     float64
	TotalPrice   float64
	ImageUrl     string
	Description  string
	Availability string
}

type CashConvertersCatalogSearchInput struct {
	Text          string
	MinPrice      *float64
	MaxPrice      *float64
	AvailableOnly bool
	Limit         int
	Offset        int
	Required      string
	RequiredMode  string
	Excluded      string
	ExcludedMode  string
	Sort          string // "newest", "price_asc", "price_desc", "relevance"
}

type CashConvertersCatalogSearchResult struct {
	ListingID    string     `gorm:"column:listingId"`
	Title        string     `gorm:"column:title"`
	Description  string     `gorm:"column:description"`
	CanonicalURL string     `gorm:"column:canonicalUrl"`
	ImageURL     *string    `gorm:"column:imageUrl"`
	TotalPrice   *float64   `gorm:"column:totalPrice"`
	Availability string     `gorm:"column:availability"`
	SourceStatus string     `gorm:"column:sourceStatus"`
	LastSeenAt   time.Time  `gorm:"column:lastSeenAt"`
	LastDetailAt *time.Time `gorm:"column:lastDetailAt"`
}

func (q *Query) GetCashConvertersScanState(ctx context.Context) (*models.CashConvertersScanState, error) {
	db := q.db.WithContext(ctx)
	var state models.CashConvertersScanState
	err := db.Where("id = ?", "cashConverters").First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		initial := models.CashConvertersScanState{
			ID:                        "cashConverters",
			CurrentSweepID:            1,
			TailNextPage:              1,
			LastKnownNonEmptyPage:     0,
			ConsecutiveEmptyTailPages: 0,
			UpdatedAt:                 time.Now().UTC(),
		}
		if err := db.Create(&initial).Error; err != nil {
			return nil, fmt.Errorf("create initial cash converters scan state: %w", err)
		}
		return &initial, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get cash converters scan state: %w", err)
	}
	return &state, nil
}

func (q *Query) UpdateCashConvertersScanState(ctx context.Context, state *models.CashConvertersScanState) error {
	state.UpdatedAt = time.Now().UTC()
	return q.db.WithContext(ctx).Save(state).Error
}

func (q *Query) UpsertCashConvertersSummaries(
	ctx context.Context,
	summaries []CashConvertersDiscoveredSummary,
	sweepID int64,
	page int,
	observedAt time.Time,
) (result CashConvertersSummaryUpsertResult, err error) {
	if len(summaries) == 0 {
		return result, nil
	}

	err = q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, sum := range summaries {
			listingID := StableListingId("cashConverters", sum.CanonicalURL)
			result.ListingIDs = append(result.ListingIDs, listingID)

			var existingListing models.Listing
			res := tx.Where("id = ?", listingID).First(&existingListing)

			if errors.Is(res.Error, gorm.ErrRecordNotFound) {
				// New listing
				lst := models.Listing{
					ID:           listingID,
					Source:       "cashConverters",
					ExternalId:   sum.ExternalItemID,
					CanonicalUrl: sum.CanonicalURL,
					Title:        sum.Title,
					ImageUrl:     &sum.ImageURL,
					Availability: ccStringPtr("available"),
					FirstSeenAt:  observedAt,
					LastSeenAt:   observedAt,
				}
				if err := tx.Create(&lst).Error; err != nil {
					return fmt.Errorf("create listing %s: %w", listingID, err)
				}

				obs := models.ListingObservation{
					ID:           uuid.NewString(),
					ListingId:    listingID,
					Source:       "cashConverters",
					ObservedAt:   observedAt,
					Price:        &sum.Price,
					Shipping:     &sum.Shipping,
					TotalPrice:   &sum.TotalPrice,
					Title:        sum.Title,
					ImageUrl:     &sum.ImageURL,
					Availability: ccStringPtr("available"),
				}
				if err := tx.Create(&obs).Error; err != nil {
					return fmt.Errorf("create observation for %s: %w", listingID, err)
				}

				meta := models.CashConvertersListingMeta{
					ListingID:        listingID,
					ExternalItemID:   sum.ExternalItemID,
					FirstSeenSweepID: sweepID,
					LastSeenSweepID:  sweepID,
					LastSeenPage:     page,
					LastListedAt:     observedAt,
					SourceStatus:     "available",
					CreatedAt:        observedAt,
					UpdatedAt:        observedAt,
				}
				if err := tx.Create(&meta).Error; err != nil {
					return fmt.Errorf("create metadata for %s: %w", listingID, err)
				}

				doc := models.CashConvertersSearchDoc{
					ListingID:    listingID,
					Title:        sum.Title,
					Description:  "",
					CanonicalURL: sum.CanonicalURL,
					ImageURL:     &sum.ImageURL,
					TotalPrice:   &sum.TotalPrice,
					Availability: "available",
					SourceStatus: "available",
					LastSeenAt:   observedAt,
					UpdatedAt:    observedAt,
				}
				if err := tx.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "listingId"}},
					DoUpdates: clause.AssignmentColumns([]string{
						"title",
						"canonicalUrl",
						"imageUrl",
						"totalPrice",
						"availability",
						"sourceStatus",
						"lastSeenAt",
						"updatedAt",
					}),
				}).Create(&doc).Error; err != nil {
					return fmt.Errorf("upsert search doc for %s: %w", listingID, err)
				}

				updateFtsSearchDoc(tx, &doc)

				result.NewListingIDs = append(result.NewListingIDs, listingID)
				result.ChangedListingIDs = append(result.ChangedListingIDs, listingID)
			} else if res.Error == nil {
				// Existing listing
				result.HadPreviouslySeen = true
				var previousDoc models.CashConvertersSearchDoc
				docResult := tx.Where("listingId = ?", listingID).First(&previousDoc)
				if docResult.Error != nil && !errors.Is(docResult.Error, gorm.ErrRecordNotFound) {
					return fmt.Errorf("load previous search doc %s: %w", listingID, docResult.Error)
				}
				materiallyChanged := existingListing.Title != sum.Title ||
					existingListing.ImageUrl == nil || *existingListing.ImageUrl != sum.ImageURL ||
					existingListing.Availability == nil || *existingListing.Availability != "available" ||
					errors.Is(docResult.Error, gorm.ErrRecordNotFound) ||
					previousDoc.TotalPrice == nil || *previousDoc.TotalPrice != sum.TotalPrice ||
					previousDoc.SourceStatus != "available"
				updates := map[string]interface{}{
					"lastSeenAt":   observedAt,
					"availability": "available",
					"title":        sum.Title,
					"imageUrl":     sum.ImageURL,
				}
				if sum.ExternalItemID != nil {
					updates["externalId"] = sum.ExternalItemID
				}
				if err := tx.Model(&existingListing).Updates(updates).Error; err != nil {
					return fmt.Errorf("update listing %s: %w", listingID, err)
				}

				metaUpdates := map[string]interface{}{
					"lastSeenSweepId":    sweepID,
					"lastSeenPage":       page,
					"lastListedAt":       observedAt,
					"sourceStatus":       "available",
					"missingSuspectedAt": nil,
					"missingVerifiedAt":  nil,
					"updatedAt":          observedAt,
				}
				if sum.ExternalItemID != nil {
					metaUpdates["externalItemId"] = sum.ExternalItemID
				}
				if err := tx.Model(&models.CashConvertersListingMeta{}).Where("listingId = ?", listingID).Updates(metaUpdates).Error; err != nil {
					return fmt.Errorf("update metadata %s: %w", listingID, err)
				}

				docUpdates := map[string]interface{}{
					"title":        sum.Title,
					"imageUrl":     sum.ImageURL,
					"totalPrice":   sum.TotalPrice,
					"availability": "available",
					"sourceStatus": "available",
					"lastSeenAt":   observedAt,
					"updatedAt":    observedAt,
				}
				if err := tx.Model(&models.CashConvertersSearchDoc{}).Where("listingId = ?", listingID).Updates(docUpdates).Error; err != nil {
					return fmt.Errorf("update search doc %s: %w", listingID, err)
				}

				var currentDoc models.CashConvertersSearchDoc
				if tx.Where("listingId = ?", listingID).First(&currentDoc).Error == nil {
					updateFtsSearchDoc(tx, &currentDoc)
				}

				if materiallyChanged {
					result.ChangedListingIDs = append(result.ChangedListingIDs, listingID)
				}
			} else {
				return res.Error
			}
		}
		return nil
	})

	return result, err
}

func (q *Query) EnqueueCashConvertersDetailJobs(ctx context.Context, listingID string, firstSeenAt time.Time) error {
	stages := []struct {
		stage string
		dueAt time.Time
	}{
		{stage: "initial", dueAt: firstSeenAt},
		{stage: "plus_5m", dueAt: firstSeenAt.Add(5 * time.Minute)},
		{stage: "plus_10m", dueAt: firstSeenAt.Add(10 * time.Minute)},
		{stage: "plus_60m", dueAt: firstSeenAt.Add(60 * time.Minute)},
		{stage: "plus_300m", dueAt: firstSeenAt.Add(300 * time.Minute)},
	}

	db := q.db.WithContext(ctx)
	for _, s := range stages {
		job := models.CashConvertersDetailJob{
			ID:        uuid.NewString(),
			ListingID: listingID,
			DueAt:     s.dueAt,
			Stage:     s.stage,
			CreatedAt: time.Now().UTC(),
			UpdatedAt: time.Now().UTC(),
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&job).Error; err != nil {
			return err
		}
	}
	return nil
}

func (q *Query) ListDueCashConvertersDetailJobs(ctx context.Context, limit int, now time.Time) ([]models.CashConvertersDetailJob, error) {
	var jobs []models.CashConvertersDetailJob
	err := q.db.WithContext(ctx).
		Where("completedAt IS NULL AND dueAt <= ?", now).
		Order("dueAt ASC").
		Limit(limit).
		Find(&jobs).Error
	return jobs, err
}

func (q *Query) CompleteCashConvertersDetailJob(ctx context.Context, id string, now time.Time) error {
	return q.db.WithContext(ctx).Model(&models.CashConvertersDetailJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"completedAt": now,
			"updatedAt":   now,
		}).Error
}

func (q *Query) RetryCashConvertersDetailJob(ctx context.Context, id string, nextDue time.Time, lastErr string) error {
	now := time.Now().UTC()
	var job models.CashConvertersDetailJob
	db := q.db.WithContext(ctx)
	if err := db.Where("id = ?", id).First(&job).Error; err != nil {
		return err
	}

	attempts := job.Attempts + 1
	updates := map[string]interface{}{
		"attempts":      attempts,
		"lastAttemptAt": now,
		"dueAt":         nextDue,
		"lastError":     lastErr,
		"updatedAt":     now,
	}
	if attempts >= 3 {
		updates["completedAt"] = now
	}
	return db.Model(&job).Updates(updates).Error
}

func (q *Query) UpdateCashConvertersDetail(
	ctx context.Context,
	listingID string,
	detail CcDetailInput,
	fetchedAt time.Time,
) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		listingUpdates := map[string]interface{}{
			"description":  detail.Description,
			"availability": detail.Availability,
			"lastDetailAt": fetchedAt,
		}
		if detail.Title != "" {
			listingUpdates["title"] = detail.Title
		}
		if detail.ImageUrl != "" {
			listingUpdates["imageUrl"] = detail.ImageUrl
		}
		if err := tx.Model(&models.Listing{}).Where("id = ?", listingID).Updates(listingUpdates).Error; err != nil {
			return err
		}

		metaUpdates := map[string]interface{}{
			"lastDetailStatus": detail.Availability,
			"updatedAt":        fetchedAt,
			"detailFetchCount": gorm.Expr("detailFetchCount + 1"),
		}
		if err := tx.Model(&models.CashConvertersListingMeta{}).Where("listingId = ?", listingID).Updates(metaUpdates).Error; err != nil {
			return fmt.Errorf("update Cash Converters metadata for %s: %w", listingID, err)
		}

		docUpdates := map[string]interface{}{
			"description":  detail.Description,
			"availability": detail.Availability,
			"lastDetailAt": fetchedAt,
			"updatedAt":    fetchedAt,
		}
		if detail.Title != "" {
			docUpdates["title"] = detail.Title
		}
		if detail.ImageUrl != "" {
			docUpdates["imageUrl"] = detail.ImageUrl
		}
		if detail.TotalPrice > 0 {
			docUpdates["totalPrice"] = detail.TotalPrice
		}
		if err := tx.Model(&models.CashConvertersSearchDoc{}).Where("listingId = ?", listingID).Updates(docUpdates).Error; err != nil {
			return err
		}

		var doc models.CashConvertersSearchDoc
		docResult := tx.Where("listingId = ?", listingID).First(&doc)
		if docResult.Error == nil {
			updateFtsSearchDoc(tx, &doc)
		} else if !errors.Is(docResult.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("reload Cash Converters search document: %w", docResult.Error)
		}
		return nil
	})
}

func (q *Query) RefreshCashConvertersSearchDoc(ctx context.Context, listingID string) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lst models.Listing
		if err := tx.Where("id = ?", listingID).First(&lst).Error; err != nil {
			return err
		}
		var meta models.CashConvertersListingMeta
		metaResult := tx.Where("listingId = ?", listingID).First(&meta)
		if metaResult.Error != nil && !errors.Is(metaResult.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load Cash Converters metadata: %w", metaResult.Error)
		}

		desc := ""
		if lst.Description != nil {
			desc = *lst.Description
		}
		avail := "available"
		if lst.Availability != nil {
			avail = *lst.Availability
		}
		status := "available"
		if meta.SourceStatus != "" {
			status = meta.SourceStatus
		}

		var obs models.ListingObservation
		obsResult := tx.Where("listingId = ?", listingID).Order("observedAt DESC").First(&obs)
		if obsResult.Error != nil && !errors.Is(obsResult.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("load Cash Converters observation: %w", obsResult.Error)
		}

		doc := models.CashConvertersSearchDoc{
			ListingID:    listingID,
			Title:        lst.Title,
			Description:  desc,
			CanonicalURL: lst.CanonicalUrl,
			ImageURL:     lst.ImageUrl,
			TotalPrice:   obs.TotalPrice,
			Availability: avail,
			SourceStatus: status,
			LastSeenAt:   lst.LastSeenAt,
			LastDetailAt: lst.LastDetailAt,
			UpdatedAt:    time.Now().UTC(),
		}

		if err := tx.Save(&doc).Error; err != nil {
			return err
		}
		updateFtsSearchDoc(tx, &doc)
		return nil
	})
}

func (q *Query) SearchCashConvertersCatalog(ctx context.Context, input CashConvertersCatalogSearchInput) ([]CashConvertersCatalogSearchResult, error) {
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	trimmedText := strings.TrimSpace(input.Text)

	if trimmedText != "" {
		results, err := q.searchCatalogFTS(ctx, trimmedText, input, limit)
		if err == nil {
			return results, nil
		}
	}

	return q.searchCatalogLIKE(ctx, trimmedText, input, limit)
}

func (q *Query) searchCatalogFTS(ctx context.Context, ftsQuery string, input CashConvertersCatalogSearchInput, limit int) ([]CashConvertersCatalogSearchResult, error) {
	db := q.db.WithContext(ctx)

	cleanQuery := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return ' '
	}, ftsQuery)
	words := strings.Fields(cleanQuery)
	if len(words) == 0 {
		return nil, fmt.Errorf("empty fts query")
	}

	ftsExpr := strings.Join(words, " AND ")

	query := db.Table("CashConvertersSearchFts f").
		Select("d.listingId, d.title, d.description, d.canonicalUrl, d.imageUrl, d.totalPrice, d.availability, d.sourceStatus, d.lastSeenAt, d.lastDetailAt").
		Joins("JOIN CashConvertersSearchDoc d ON d.listingId = f.listingId").
		Where("CashConvertersSearchFts MATCH ?", ftsExpr)

	query = applyCashConvertersCatalogFilters(query, input)

	switch input.Sort {
	case "price_asc":
		query = query.Order("d.totalPrice ASC")
	case "price_desc":
		query = query.Order("d.totalPrice DESC")
	case "newest":
		query = query.Order("d.lastSeenAt DESC")
	case "relevance":
		query = query.Order("bm25(CashConvertersSearchFts), d.lastSeenAt DESC")
	default:
		query = query.Order("d.lastSeenAt DESC")
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}
	query = query.Limit(limit).Offset(offset)

	var results []CashConvertersCatalogSearchResult
	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (q *Query) searchCatalogLIKE(ctx context.Context, text string, input CashConvertersCatalogSearchInput, limit int) ([]CashConvertersCatalogSearchResult, error) {
	db := q.db.WithContext(ctx)

	query := db.Table("CashConvertersSearchDoc d").
		Select("d.listingId, d.title, d.description, d.canonicalUrl, d.imageUrl, d.totalPrice, d.availability, d.sourceStatus, d.lastSeenAt, d.lastDetailAt")

	if text != "" {
		likePattern := "%" + strings.ToLower(text) + "%"
		query = query.Where("lower(d.title || ' ' || d.description) LIKE ?", likePattern)
	}

	query = applyCashConvertersCatalogFilters(query, input)

	switch input.Sort {
	case "price_asc":
		query = query.Order("d.totalPrice ASC")
	case "price_desc":
		query = query.Order("d.totalPrice DESC")
	case "newest", "relevance":
		query = query.Order("d.lastSeenAt DESC")
	default:
		query = query.Order("d.lastSeenAt DESC")
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}
	query = query.Limit(limit).Offset(offset)

	var results []CashConvertersCatalogSearchResult
	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (q *Query) CountCashConvertersCatalog(ctx context.Context, input CashConvertersCatalogSearchInput) (int64, error) {
	trimmedText := strings.TrimSpace(input.Text)
	if trimmedText != "" {
		count, err := q.countCatalogFTS(ctx, trimmedText, input)
		if err == nil {
			return count, nil
		}
	}
	return q.countCatalogLIKE(ctx, trimmedText, input)
}

func (q *Query) countCatalogFTS(ctx context.Context, text string, input CashConvertersCatalogSearchInput) (int64, error) {
	cleanQuery := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
			return r
		}
		return ' '
	}, text)
	words := strings.Fields(cleanQuery)
	if len(words) == 0 {
		return 0, fmt.Errorf("empty fts query")
	}

	query := q.db.WithContext(ctx).
		Table("CashConvertersSearchFts f").
		Joins("JOIN CashConvertersSearchDoc d ON d.listingId = f.listingId").
		Where("CashConvertersSearchFts MATCH ?", strings.Join(words, " AND "))
	query = applyCashConvertersCatalogFilters(query, input)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (q *Query) countCatalogLIKE(ctx context.Context, text string, input CashConvertersCatalogSearchInput) (int64, error) {
	query := q.db.WithContext(ctx).Table("CashConvertersSearchDoc d")
	if text != "" {
		likePattern := "%" + strings.ToLower(text) + "%"
		query = query.Where("lower(d.title || ' ' || d.description) LIKE ?", likePattern)
	}
	query = applyCashConvertersCatalogFilters(query, input)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func applyCashConvertersCatalogFilters(query *gorm.DB, input CashConvertersCatalogSearchInput) *gorm.DB {
	if input.MinPrice != nil {
		query = query.Where("d.totalPrice >= ?", *input.MinPrice)
	}
	if input.MaxPrice != nil {
		query = query.Where("d.totalPrice <= ?", *input.MaxPrice)
	}
	if input.AvailableOnly {
		query = query.Where("d.availability = 'available' AND d.sourceStatus = 'available'")
	}
	query = applyCatalogPhraseFilter(query, ParsePhraseList(input.Required), input.RequiredMode, false)
	query = applyCatalogPhraseFilter(query, ParsePhraseList(input.Excluded), input.ExcludedMode, true)
	return query
}

func applyCatalogPhraseFilter(query *gorm.DB, phrases []string, mode string, negate bool) *gorm.DB {
	if len(phrases) == 0 {
		return query
	}

	wholeWords := strings.HasSuffix(mode, "_words") || mode == "all" || mode == "any" || mode == ""
	matchAny := strings.HasPrefix(mode, "any")
	predicates := make([]string, 0, len(phrases))
	args := make([]interface{}, 0, len(phrases))
	for _, phrase := range phrases {
		if wholeWords {
			words := strings.Fields(strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' {
					return r
				}
				return ' '
			}, phrase))
			if len(words) == 0 {
				continue
			}
			predicates = append(predicates, "d.listingId IN (SELECT listingId FROM CashConvertersSearchFts WHERE CashConvertersSearchFts MATCH ?)")
			args = append(args, `"`+strings.Join(words, " ")+`"`)
		} else {
			predicates = append(predicates, "lower(d.title || ' ' || d.description) LIKE ?")
			args = append(args, "%"+strings.ToLower(phrase)+"%")
		}
	}
	if len(predicates) == 0 {
		return query
	}

	joiner := " AND "
	if matchAny {
		joiner = " OR "
	}
	condition := "(" + strings.Join(predicates, joiner) + ")"
	if negate {
		condition = "NOT " + condition
	}
	return query.Where(condition, args...)
}

func (q *Query) MarkCashConvertersMissingCandidates(ctx context.Context, completedSweepID int64, now time.Time) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.CashConvertersListingMeta{}).
			Where("lastSeenSweepId < ? AND sourceStatus = ?", completedSweepID, "available").
			Updates(map[string]interface{}{
				"sourceStatus":       "suspected_missing",
				"missingSuspectedAt": now,
				"updatedAt":          now,
			}).Error; err != nil {
			return fmt.Errorf("mark Cash Converters missing candidates: %w", err)
		}

		if err := tx.Model(&models.CashConvertersSearchDoc{}).
			Where("sourceStatus = ? AND listingId IN (SELECT listingId FROM CashConvertersListingMeta WHERE sourceStatus = ?)", "available", "suspected_missing").
			Update("sourceStatus", "suspected_missing").Error; err != nil {
			return fmt.Errorf("update Cash Converters missing search documents: %w", err)
		}
		return nil
	})
}

func (q *Query) ListCashConvertersMissingVerificationsDue(ctx context.Context, limit int, now time.Time) ([]models.Listing, error) {
	var listings []models.Listing
	retryCutoff := now.Add(-10 * time.Minute)
	err := q.db.WithContext(ctx).
		Table("Listing").
		Joins("JOIN CashConvertersListingMeta m ON m.listingId = Listing.id").
		Where("m.sourceStatus = ? AND m.missingVerifiedAt IS NULL", "suspected_missing").
		Where("m.missingVerificationCount = 0 OR m.updatedAt <= ?", retryCutoff).
		Limit(limit).
		Find(&listings).Error
	return listings, err
}

func (q *Query) ConfirmCashConvertersMissing(ctx context.Context, listingID string, now time.Time) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.CashConvertersListingMeta{}).
			Where("listingId = ?", listingID).
			Updates(map[string]interface{}{
				"sourceStatus":             "confirmed_missing",
				"missingVerifiedAt":        now,
				"missingVerificationCount": gorm.Expr("missingVerificationCount + 1"),
				"updatedAt":                now,
			}).Error; err != nil {
			return fmt.Errorf("confirm Cash Converters metadata missing for %s: %w", listingID, err)
		}

		if err := tx.Model(&models.Listing{}).
			Where("id = ?", listingID).
			Updates(map[string]interface{}{
				"availability":  "unavailable",
				"unavailableAt": now,
			}).Error; err != nil {
			return fmt.Errorf("mark Cash Converters listing unavailable for %s: %w", listingID, err)
		}

		if err := tx.Model(&models.CashConvertersSearchDoc{}).
			Where("listingId = ?", listingID).
			Updates(map[string]interface{}{
				"sourceStatus": "confirmed_missing",
				"availability": "unavailable",
				"updatedAt":    now,
			}).Error; err != nil {
			return fmt.Errorf("mark Cash Converters search document unavailable for %s: %w", listingID, err)
		}
		return nil
	})
}

func (q *Query) RecordCashConvertersMissingVerification(
	ctx context.Context,
	listingID string,
	now time.Time,
) (bool, error) {
	var confirmed bool
	err := q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var meta models.CashConvertersListingMeta
		if err := tx.Where("listingId = ?", listingID).First(&meta).Error; err != nil {
			return fmt.Errorf("load Cash Converters metadata for missing verification %s: %w", listingID, err)
		}

		verificationCount := meta.MissingVerificationCount + 1
		confirmed = verificationCount >= 2
		status := "suspected_missing"
		updates := map[string]interface{}{
			"missingVerificationCount": verificationCount,
			"updatedAt":                now,
		}
		if confirmed {
			status = "confirmed_missing"
			updates["missingVerifiedAt"] = now
		}
		updates["sourceStatus"] = status

		if err := tx.Model(&meta).Updates(updates).Error; err != nil {
			return fmt.Errorf("record Cash Converters missing verification for %s: %w", listingID, err)
		}
		if !confirmed {
			return nil
		}
		if err := tx.Model(&models.Listing{}).Where("id = ?", listingID).Updates(map[string]interface{}{
			"availability":  "unavailable",
			"unavailableAt": now,
		}).Error; err != nil {
			return fmt.Errorf("mark verified Cash Converters listing unavailable for %s: %w", listingID, err)
		}
		if err := tx.Model(&models.CashConvertersSearchDoc{}).Where("listingId = ?", listingID).Updates(map[string]interface{}{
			"sourceStatus": "confirmed_missing",
			"availability": "unavailable",
			"updatedAt":    now,
		}).Error; err != nil {
			return fmt.Errorf("mark verified Cash Converters search document unavailable for %s: %w", listingID, err)
		}
		return nil
	})
	return confirmed, err
}

func (q *Query) DeleteConfirmedCashConvertersListings(ctx context.Context, olderThan time.Time, limit int) (int64, error) {
	db := q.db.WithContext(ctx)

	var metas []models.CashConvertersListingMeta
	err := db.Where("sourceStatus = ? AND missingVerifiedAt <= ?", "confirmed_missing", olderThan).
		Limit(limit).
		Find(&metas).Error
	if err != nil || len(metas) == 0 {
		return 0, err
	}

	var count int64
	for _, meta := range metas {
		listingID := meta.ListingID

		var lst models.Listing
		if err := db.Where("id = ?", listingID).First(&lst).Error; err != nil {
			return count, fmt.Errorf("load confirmed missing listing %s for pruning: %w", listingID, err)
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			tombstone := models.CashConvertersDeletedListing{
				ListingID:      listingID,
				CanonicalURL:   lst.CanonicalUrl,
				ExternalItemID: meta.ExternalItemID,
				DeletedAt:      time.Now().UTC(),
				Reason:         "confirmed_missing_prune",
			}
			if err := tx.Create(&tombstone).Error; err != nil {
				return fmt.Errorf("create Cash Converters tombstone for %s: %w", listingID, err)
			}
			if err := tx.Where("listingId = ?", listingID).Delete(&models.ListingObservation{}).Error; err != nil {
				return fmt.Errorf("delete observations for %s: %w", listingID, err)
			}
			if err := tx.Where("listingId = ?", listingID).Delete(&models.QueryListingState{}).Error; err != nil {
				return fmt.Errorf("delete query states for %s: %w", listingID, err)
			}
			if err := tx.Exec("DELETE FROM CashConvertersSearchFts WHERE listingId = ?", listingID).Error; err != nil &&
				!strings.Contains(strings.ToLower(err.Error()), "no such table") {
				return fmt.Errorf("delete FTS document for %s: %w", listingID, err)
			}
			if err := tx.Where("listingId = ?", listingID).Delete(&models.CashConvertersSearchDoc{}).Error; err != nil {
				return fmt.Errorf("delete search document for %s: %w", listingID, err)
			}
			if err := tx.Where("listingId = ?", listingID).Delete(&models.CashConvertersDetailJob{}).Error; err != nil {
				return fmt.Errorf("delete detail jobs for %s: %w", listingID, err)
			}
			if err := tx.Where("listingId = ?", listingID).Delete(&models.CashConvertersListingMeta{}).Error; err != nil {
				return fmt.Errorf("delete metadata for %s: %w", listingID, err)
			}
			if err := tx.Where("id = ?", listingID).Delete(&models.Listing{}).Error; err != nil {
				return fmt.Errorf("delete listing %s: %w", listingID, err)
			}
			return nil
		})
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func updateFtsSearchDoc(tx *gorm.DB, doc *models.CashConvertersSearchDoc) {
	_ = tx.Exec("DELETE FROM CashConvertersSearchFts WHERE listingId = ?", doc.ListingID).Error
	_ = tx.Exec("INSERT INTO CashConvertersSearchFts (listingId, title, description, canonicalUrl) VALUES (?, ?, ?, ?)",
		doc.ListingID, doc.Title, doc.Description, doc.CanonicalURL).Error
}

func ccStringPtr(s string) *string {
	return &s
}

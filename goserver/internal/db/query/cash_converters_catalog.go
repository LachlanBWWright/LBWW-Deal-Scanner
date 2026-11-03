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

// Keep batches below SQLite's bind-parameter limit, including wider model rows.
const ccCatalogWriteBatchSize = 40

func (q *Query) UpsertCashConvertersSummaries(
	ctx context.Context,
	summaries []CashConvertersDiscoveredSummary,
	sweepID int64,
	page int,
	observedAt time.Time,
) (CashConvertersSummaryUpsertResult, error) {
	return q.upsertCashConvertersSummaries(ctx, summaries, sweepID, page, observedAt, nil)
}

// UpsertCashConvertersSummariesWithDetailJobs commits discovery and initial job
// scheduling together. A failed enqueue cannot leave a known listing without jobs.
func (q *Query) UpsertCashConvertersSummariesWithDetailJobs(
	ctx context.Context,
	summaries []CashConvertersDiscoveredSummary,
	sweepID int64,
	page int,
	observedAt time.Time,
	priority int,
) (CashConvertersSummaryUpsertResult, error) {
	return q.upsertCashConvertersSummaries(ctx, summaries, sweepID, page, observedAt, &priority)
}

func (q *Query) upsertCashConvertersSummaries(
	ctx context.Context,
	summaries []CashConvertersDiscoveredSummary,
	sweepID int64,
	page int,
	observedAt time.Time,
	detailPriority *int,
) (CashConvertersSummaryUpsertResult, error) {
	var result CashConvertersSummaryUpsertResult
	if len(summaries) == 0 {
		return result, nil
	}
	// Deduplicate before batching; retain the latest summary for each listing.
	unique := make([]CashConvertersDiscoveredSummary, 0, len(summaries))
	positions := make(map[string]int, len(summaries))
	for _, summary := range summaries {
		id := StableListingId("cashConverters", summary.CanonicalURL)
		if position, exists := positions[id]; exists {
			unique[position] = summary
		} else {
			positions[id] = len(unique)
			unique = append(unique, summary)
			result.ListingIDs = append(result.ListingIDs, id)
		}
	}
	err := q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for start := 0; start < len(unique); start += ccCatalogWriteBatchSize {
			end := min(start+ccCatalogWriteBatchSize, len(unique))
			if err := upsertCashConvertersSummaryBatch(tx, unique[start:end], result.ListingIDs[start:end], sweepID, page, observedAt, &result); err != nil {
				return err
			}
		}
		if detailPriority != nil {
			return enqueueCashConvertersDetailJobs(tx, result.NewListingIDs, observedAt, *detailPriority)
		}
		return nil
	})
	if err != nil {
		return CashConvertersSummaryUpsertResult{}, err
	}
	return result, nil
}

func upsertCashConvertersSummaryBatch(
	tx *gorm.DB,
	summaries []CashConvertersDiscoveredSummary,
	ids []string,
	sweepID int64,
	page int,
	observedAt time.Time,
	result *CashConvertersSummaryUpsertResult,
) error {
	var existingListings []models.Listing
	if err := tx.Where("id IN ?", ids).Find(&existingListings).Error; err != nil {
		return fmt.Errorf("load catalogue listings: %w", err)
	}
	var existingDocs []models.CashConvertersSearchDoc
	if err := tx.Where("listingId IN ?", ids).Find(&existingDocs).Error; err != nil {
		return fmt.Errorf("load catalogue search documents: %w", err)
	}
	listingMap := make(map[string]models.Listing, len(existingListings))
	for _, listing := range existingListings {
		listingMap[listing.ID] = listing
	}
	docMap := make(map[string]models.CashConvertersSearchDoc, len(existingDocs))
	for _, doc := range existingDocs {
		docMap[doc.ListingID] = doc
	}
	var listings []models.Listing
	var observations []models.ListingObservation
	var metadata []models.CashConvertersListingMeta
	var docs []models.CashConvertersSearchDoc
	var unchangedListings, unchangedDocs []string
	for index, summary := range summaries {
		id := ids[index]
		listing, exists := listingMap[id]
		availability := "available"
		if exists && listing.LastDetailAt != nil && listing.Availability != nil {
			availability = *listing.Availability
		}
		listingChanged := !exists || listing.Title != summary.Title ||
			listing.ImageUrl == nil || *listing.ImageUrl != summary.ImageURL ||
			listing.Availability == nil || *listing.Availability != availability ||
			(summary.ExternalItemID != nil && (listing.ExternalId == nil || *listing.ExternalId != *summary.ExternalItemID))
		if !exists {
			listing = models.Listing{ID: id, Source: "cashConverters", CanonicalUrl: summary.CanonicalURL, FirstSeenAt: observedAt}
			result.NewListingIDs = append(result.NewListingIDs, id)
			observations = append(observations, models.ListingObservation{
				ID: uuid.NewString(), ListingId: id, Source: "cashConverters", ObservedAt: observedAt,
				Price: &summary.Price, Shipping: &summary.Shipping, TotalPrice: &summary.TotalPrice,
				Title: summary.Title, ImageUrl: &summary.ImageURL, Availability: ccStringPtr(availability),
			})
		} else {
			result.HadPreviouslySeen = true
		}
		if summary.ExternalItemID != nil {
			listing.ExternalId = summary.ExternalItemID
		}
		listing.Title = summary.Title
		listing.ImageUrl = &summary.ImageURL
		listing.Availability = ccStringPtr(availability)
		listing.LastSeenAt = observedAt
		if listingChanged {
			listings = append(listings, listing)
		} else {
			unchangedListings = append(unchangedListings, id)
		}
		metadata = append(metadata, models.CashConvertersListingMeta{
			ListingID: id, ExternalItemID: listing.ExternalId, FirstSeenSweepID: sweepID,
			LastSeenSweepID: sweepID, LastSeenPage: page, LastListedAt: observedAt,
			SourceStatus: "available", CreatedAt: observedAt, UpdatedAt: observedAt,
		})
		doc, hasDoc := docMap[id]
		docChanged := !hasDoc || doc.Title != summary.Title || doc.CanonicalURL != summary.CanonicalURL ||
			doc.ImageURL == nil || *doc.ImageURL != summary.ImageURL ||
			doc.TotalPrice == nil || *doc.TotalPrice != summary.TotalPrice ||
			doc.Availability != availability || doc.SourceStatus != "available"
		if !hasDoc {
			doc = models.CashConvertersSearchDoc{ListingID: id, LastDetailAt: listing.LastDetailAt}
			if listing.Description != nil {
				doc.Description = *listing.Description
			}
		}
		doc.Title = summary.Title
		doc.CanonicalURL = summary.CanonicalURL
		doc.ImageURL = &summary.ImageURL
		doc.TotalPrice = &summary.TotalPrice
		doc.Availability = availability
		doc.SourceStatus = "available"
		doc.LastSeenAt = observedAt
		if docChanged {
			doc.UpdatedAt = observedAt
			docs = append(docs, doc)
		} else {
			unchangedDocs = append(unchangedDocs, id)
		}
		if !exists || docChanged {
			result.ChangedListingIDs = append(result.ChangedListingIDs, id)
		}
	}
	if len(listings) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "imageUrl", "availability", "externalId", "lastSeenAt"}),
		}).CreateInBatches(&listings, ccCatalogWriteBatchSize).Error; err != nil {
			return fmt.Errorf("upsert catalogue listings: %w", err)
		}
	}
	if len(unchangedListings) > 0 {
		if err := tx.Model(&models.Listing{}).Where("id IN ?", unchangedListings).Update("lastSeenAt", observedAt).Error; err != nil {
			return fmt.Errorf("refresh catalogue listing sightings: %w", err)
		}
	}
	if len(observations) > 0 {
		if err := tx.CreateInBatches(&observations, ccCatalogWriteBatchSize).Error; err != nil {
			return fmt.Errorf("create catalogue observations: %w", err)
		}
	}
	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "listingId"}},
		DoUpdates: append(clause.AssignmentColumns([]string{
			"lastSeenSweepId", "lastSeenPage", "lastListedAt", "sourceStatus",
			"missingSuspectedAt", "missingVerifiedAt", "deletedAt", "updatedAt",
		}), clause.Assignment{Column: clause.Column{Name: "externalItemId"}, Value: gorm.Expr("COALESCE(excluded.externalItemId, CashConvertersListingMeta.externalItemId)")}),
	}).CreateInBatches(&metadata, ccCatalogWriteBatchSize).Error; err != nil {
		return fmt.Errorf("upsert catalogue metadata: %w", err)
	}
	if len(docs) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "listingId"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "canonicalUrl", "imageUrl", "totalPrice", "availability", "sourceStatus", "lastSeenAt", "updatedAt"}),
		}).CreateInBatches(&docs, ccCatalogWriteBatchSize).Error; err != nil {
			return fmt.Errorf("upsert catalogue search documents: %w", err)
		}
	}
	if len(unchangedDocs) > 0 {
		if err := tx.Model(&models.CashConvertersSearchDoc{}).Where("listingId IN ?", unchangedDocs).UpdateColumn("lastSeenAt", observedAt).Error; err != nil {
			return fmt.Errorf("refresh catalogue search sightings: %w", err)
		}
	}
	return nil
}

func (q *Query) EnqueueCashConvertersDetailJobs(ctx context.Context, listingID string, firstSeenAt time.Time, priority int) error {
	return enqueueCashConvertersDetailJobs(q.db.WithContext(ctx), []string{listingID}, firstSeenAt, priority)
}

func enqueueCashConvertersDetailJobs(db *gorm.DB, listingIDs []string, firstSeenAt time.Time, priority int) error {
	stages := []struct {
		stage string
		delay time.Duration
	}{
		{stage: "initial"},
		{stage: "plus_5m", delay: 5 * time.Minute},
		{stage: "plus_10m", delay: 10 * time.Minute},
		{stage: "plus_60m", delay: 60 * time.Minute},
		{stage: "plus_300m", delay: 300 * time.Minute},
	}
	jobs := make([]models.CashConvertersDetailJob, 0, len(listingIDs)*len(stages))
	now := time.Now().UTC()
	for _, id := range listingIDs {
		for _, stage := range stages {
			jobs = append(jobs, models.CashConvertersDetailJob{
				ID: uuid.NewString(), ListingID: id, DueAt: firstSeenAt.Add(stage.delay),
				Stage: stage.stage, Priority: priority, CreatedAt: now, UpdatedAt: now,
			})
		}
	}
	if len(jobs) == 0 {
		return nil
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&jobs, ccCatalogWriteBatchSize).Error
}

func (q *Query) ListDueCashConvertersDetailJobs(ctx context.Context, limit int, now time.Time) ([]models.CashConvertersDetailJob, error) {
	if limit <= 0 {
		return []models.CashConvertersDetailJob{}, nil
	}
	load := func(stage string, remaining int) ([]models.CashConvertersDetailJob, error) {
		var jobs []models.CashConvertersDetailJob
		query := q.db.WithContext(ctx).
			Where("completedAt IS NULL AND dueAt <= ?", now)
		if stage == "initial" {
			query = query.Where("stage = ?", stage)
		} else {
			query = query.Where("stage <> ?", "initial")
		}
		err := query.
			Order("priority DESC").
			Order("dueAt ASC").
			Limit(remaining).
			Find(&jobs).Error
		return jobs, err
	}
	initial, err := load("initial", limit)
	if err != nil {
		return nil, err
	}
	if len(initial) == limit {
		return initial, nil
	}
	remaining, err := load("non-initial", limit-len(initial))
	if err != nil {
		return nil, err
	}
	return append(initial, remaining...), nil
}

func (q *Query) DeleteOrphanedCashConvertersChildren(ctx context.Context, listingID string) error {
	return q.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		children := []interface{}{
			&models.QueryListingState{},
			&models.CashConvertersSearchDoc{},
			&models.CashConvertersDetailJob{},
			&models.CashConvertersListingMeta{},
			&models.ListingObservation{},
		}
		for _, child := range children {
			if err := tx.Where("listingId = ?", listingID).Delete(child).Error; err != nil {
				return fmt.Errorf("delete orphaned Cash Converters children for %s: %w", listingID, err)
			}
		}
		return nil
	})
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

	ftsExpr := cashConvertersFTSPrefixExpression(words)

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
		Where("CashConvertersSearchFts MATCH ?", cashConvertersFTSPrefixExpression(words))
	query = applyCashConvertersCatalogFilters(query, input)

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func cashConvertersFTSPrefixExpression(words []string) string {
	expressions := make([]string, 0, len(words))
	for _, word := range words {
		expressions = append(expressions, `"`+word+`"*`)
	}
	return strings.Join(expressions, " AND ")
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

	type pruneCandidate struct {
		ListingID      string  `gorm:"column:listingId"`
		CanonicalURL   string  `gorm:"column:canonicalUrl"`
		ExternalItemID *string `gorm:"column:externalItemId"`
	}

	var candidates []pruneCandidate
	err := db.Table("CashConvertersListingMeta AS m").
		Select("m.listingId, l.canonicalUrl, m.externalItemId").
		Joins("JOIN Listing AS l ON l.id = m.listingId").
		Where("m.sourceStatus = ? AND m.missingVerifiedAt <= ?", "confirmed_missing", olderThan).
		Order("m.missingVerifiedAt ASC").
		Limit(limit).
		Find(&candidates).Error
	if err != nil || len(candidates) == 0 {
		return 0, err
	}

	listingIDs := make([]string, 0, len(candidates))
	tombstones := make([]models.CashConvertersDeletedListing, 0, len(candidates))
	deletedAt := time.Now().UTC()
	for _, candidate := range candidates {
		listingIDs = append(listingIDs, candidate.ListingID)
		tombstones = append(tombstones, models.CashConvertersDeletedListing{
			ListingID:      candidate.ListingID,
			CanonicalURL:   candidate.CanonicalURL,
			ExternalItemID: candidate.ExternalItemID,
			DeletedAt:      deletedAt,
			Reason:         "confirmed_missing_prune",
		})
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&tombstones).Error; err != nil {
			return fmt.Errorf("create Cash Converters prune tombstones: %w", err)
		}
		if err := tx.Where("listingId IN ?", listingIDs).Delete(&models.ListingObservation{}).Error; err != nil {
			return fmt.Errorf("delete Cash Converters observations: %w", err)
		}
		if err := tx.Where("listingId IN ?", listingIDs).Delete(&models.QueryListingState{}).Error; err != nil {
			return fmt.Errorf("delete Cash Converters query states: %w", err)
		}
		if err := tx.Where("listingId IN ?", listingIDs).Delete(&models.CashConvertersSearchDoc{}).Error; err != nil {
			return fmt.Errorf("delete Cash Converters search documents: %w", err)
		}
		if err := tx.Where("listingId IN ?", listingIDs).Delete(&models.CashConvertersDetailJob{}).Error; err != nil {
			return fmt.Errorf("delete Cash Converters detail jobs: %w", err)
		}
		if err := tx.Where("listingId IN ?", listingIDs).Delete(&models.CashConvertersListingMeta{}).Error; err != nil {
			return fmt.Errorf("delete Cash Converters metadata: %w", err)
		}
		if err := tx.Where("id IN ?", listingIDs).Delete(&models.Listing{}).Error; err != nil {
			return fmt.Errorf("delete Cash Converters listings: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int64(len(candidates)), nil
}

func ccStringPtr(s string) *string {
	return &s
}

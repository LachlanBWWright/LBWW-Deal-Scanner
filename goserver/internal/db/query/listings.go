package query

import (
	"context"
	"errors"
	"time"

	"dealscanner/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const listingFreshnessWriteInterval = time.Hour

type DiscoveredListing struct {
	Source       string
	ExternalId   *string
	CanonicalUrl string
	Title        string
	Price        *float64
	Shipping     *float64
	TotalPrice   *float64
	Currency     *string
	ImageUrl     *string
	Description  *string
	Availability *string
	LastDetailAt *time.Time
}

func StableListingId(source, canonicalUrl string) string {
	return source + ":" + canonicalUrl
}

func ListingIDsForDiscovered(discovered []DiscoveredListing) []string {
	listingIDs := make([]string, 0, len(discovered))
	seen := make(map[string]struct{}, len(discovered))
	for _, found := range discovered {
		listingID := StableListingId(found.Source, found.CanonicalUrl)
		if _, ok := seen[listingID]; ok {
			continue
		}
		seen[listingID] = struct{}{}
		listingIDs = append(listingIDs, listingID)
	}
	return listingIDs
}

type ListingEvaluationCache struct {
	ExistingListings   map[string]*models.Listing
	LatestObservations map[string]*models.ListingObservation
	ExistingStates     map[string]*models.QueryListingState
}

func (q *Query) PersistListingObservation(ctx context.Context, listing DiscoveredListing, observedAt time.Time) (string, error) {
	listingId := StableListingId(listing.Source, listing.CanonicalUrl)

	err := q.Transaction(func(tx *Query) error {
		existingListing, err := tx.Listing.WithContext(ctx).Where(tx.Listing.ID.Eq(listingId)).First()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			lst := models.Listing{
				ID:           listingId,
				Source:       listing.Source,
				ExternalId:   listing.ExternalId,
				CanonicalUrl: listing.CanonicalUrl,
				Title:        listing.Title,
				ImageUrl:     listing.ImageUrl,
				Description:  listing.Description,
				Availability: listing.Availability,
				FirstSeenAt:  observedAt,
				LastSeenAt:   observedAt,
				LastDetailAt: listing.LastDetailAt,
			}
			if listing.Availability != nil && *listing.Availability == "unavailable" {
				lst.UnavailableAt = &observedAt
			}
			if err := tx.Listing.WithContext(ctx).Create(&lst); err != nil {
				return err
			}
		} else if err == nil {
			if !shouldSkipListingUpdate(existingListing, listing, observedAt) {
				updates := map[string]interface{}{
					"externalId":   listing.ExternalId,
					"title":        listing.Title,
					"imageUrl":     listing.ImageUrl,
					"description":  listing.Description,
					"availability": listing.Availability,
					"lastSeenAt":   observedAt,
				}
				if listing.Availability != nil && *listing.Availability == "unavailable" {
					updates["unavailableAt"] = &observedAt
				}
				if listing.LastDetailAt != nil {
					updates["lastDetailAt"] = listing.LastDetailAt
				}
				if _, err := tx.Listing.WithContext(ctx).Where(tx.Listing.ID.Eq(listingId)).Updates(updates); err != nil {
					return err
				}
			}
		} else {
			return err
		}

		obs := models.ListingObservation{
			ID:           uuid.New().String(),
			ListingId:    listingId,
			Source:       listing.Source,
			ObservedAt:   observedAt,
			Price:        listing.Price,
			Shipping:     listing.Shipping,
			TotalPrice:   listing.TotalPrice,
			Currency:     listing.Currency,
			Title:        listing.Title,
			ImageUrl:     listing.ImageUrl,
			Description:  listing.Description,
			Availability: listing.Availability,
		}
		var latestObservation models.ListingObservation
		err = tx.UnderlyingDB().WithContext(ctx).
			Where("listingId = ?", listingId).
			Order("observedAt DESC").
			First(&latestObservation).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err == nil && shouldSkipListingObservation(&latestObservation, &obs, observedAt) {
			return nil
		}
		return tx.ListingObservation.WithContext(ctx).Create(&obs)
	})

	if err != nil {
		return "", err
	}

	return listingId, nil
}

func shouldSkipListingUpdate(existing *models.Listing, listing DiscoveredListing, observedAt time.Time) bool {
	if existing == nil {
		return false
	}

	metadataUnchanged :=
		stringPtrEqual(existing.ExternalId, listing.ExternalId) &&
			existing.Title == listing.Title &&
			stringPtrEqual(existing.ImageUrl, listing.ImageUrl) &&
			stringPtrEqual(existing.Description, listing.Description) &&
			stringPtrEqual(existing.Availability, listing.Availability) &&
			timePtrEqual(existing.LastDetailAt, listing.LastDetailAt)
	if !metadataUnchanged {
		return false
	}

	if listing.Availability != nil && *listing.Availability == "unavailable" && existing.UnavailableAt == nil {
		return false
	}

	return observedAt.Sub(existing.LastSeenAt) < listingFreshnessWriteInterval
}

func (q *Query) GetQueryListingState(ctx context.Context, queryId, listingId string) (*models.QueryListingState, error) {
	s, err := q.QueryListingState.WithContext(ctx).Where(q.QueryListingState.QueryId.Eq(queryId), q.QueryListingState.ListingId.Eq(listingId)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return s, nil
}

func (q *Query) LoadListingEvaluationCache(
	ctx context.Context,
	queryID string,
	discovered []DiscoveredListing,
) (ListingEvaluationCache, error) {
	listingIDs := ListingIDsForDiscovered(discovered)

	existingListings, err := q.LoadListings(ctx, listingIDs)
	if err != nil {
		return ListingEvaluationCache{}, err
	}

	latestObservations, err := q.LoadLatestListingObservations(ctx, listingIDs)
	if err != nil {
		return ListingEvaluationCache{}, err
	}

	existingStates, err := q.LoadQueryListingStates(ctx, queryID, listingIDs)
	if err != nil {
		return ListingEvaluationCache{}, err
	}

	return ListingEvaluationCache{
		ExistingListings:   existingListings,
		LatestObservations: latestObservations,
		ExistingStates:     existingStates,
	}, nil
}

func (q *Query) UpsertQueryListingState(ctx context.Context, state *models.QueryListingState) error {
	return q.QueryListingState.WithContext(ctx).Save(state)
}

func (q *Query) UpsertQueryListingStateIfChanged(
	ctx context.Context,
	previous *models.QueryListingState,
	state *models.QueryListingState,
	evaluatedAt time.Time,
) error {
	if previous != nil &&
		queryListingStateEqualIgnoringFreshness(previous, state) &&
		evaluatedAt.Sub(previous.LastEvaluatedAt) < queryListingStateFreshnessWriteInterval {
		return nil
	}
	return q.UpsertQueryListingState(ctx, state)
}

func (q *Query) LoadListings(ctx context.Context, listingIDs []string) (map[string]*models.Listing, error) {
	listingsByID := make(map[string]*models.Listing, len(listingIDs))
	if len(listingIDs) == 0 {
		return listingsByID, nil
	}

	var listings []*models.Listing
	if err := q.UnderlyingDB().WithContext(ctx).Where("id IN ?", listingIDs).Find(&listings).Error; err != nil {
		return nil, err
	}
	for _, listing := range listings {
		listingsByID[listing.ID] = listing
	}
	return listingsByID, nil
}

func (q *Query) LoadQueryListingStates(
	ctx context.Context,
	queryID string,
	listingIDs []string,
) (map[string]*models.QueryListingState, error) {
	statesByListingID := make(map[string]*models.QueryListingState, len(listingIDs))
	if len(listingIDs) == 0 {
		return statesByListingID, nil
	}

	var states []*models.QueryListingState
	if err := q.UnderlyingDB().WithContext(ctx).
		Where("queryId = ? AND listingId IN ?", queryID, listingIDs).
		Find(&states).Error; err != nil {
		return nil, err
	}
	for _, state := range states {
		statesByListingID[state.ListingId] = state
	}
	return statesByListingID, nil
}

func (q *Query) LoadLatestListingObservations(
	ctx context.Context,
	listingIDs []string,
) (map[string]*models.ListingObservation, error) {
	observationsByListingID := make(map[string]*models.ListingObservation, len(listingIDs))
	if len(listingIDs) == 0 {
		return observationsByListingID, nil
	}

	var observations []*models.ListingObservation
	err := q.UnderlyingDB().WithContext(ctx).
		Where("listingId IN ?", listingIDs).
		Where(`observedAt = (
			SELECT MAX(latest.observedAt)
			FROM ListingObservation AS latest
			WHERE latest.listingId = ListingObservation.listingId
		)`).
		Find(&observations).Error
	if err != nil {
		return nil, err
	}
	for _, observation := range observations {
		observationsByListingID[observation.ListingId] = observation
	}
	return observationsByListingID, nil
}

func (q *Query) PersistListingBatch(
	ctx context.Context,
	discovered []DiscoveredListing,
	existing map[string]*models.Listing,
	latestObservations map[string]*models.ListingObservation,
	existingStates map[string]*models.QueryListingState,
	states []*models.QueryListingState,
	observedAt time.Time,
) error {
	if len(discovered) == 0 {
		return nil
	}

	listings := make([]*models.Listing, 0, len(discovered))
	observations := make([]*models.ListingObservation, 0, len(discovered))
	for _, found := range discovered {
		listingID := StableListingId(found.Source, found.CanonicalUrl)
		if previous := existing[listingID]; previous == nil || !shouldSkipListingUpdate(previous, found, observedAt) {
			firstSeenAt := observedAt
			lastDetailAt := found.LastDetailAt
			var unavailableAt *time.Time
			if previous != nil {
				firstSeenAt = previous.FirstSeenAt
				unavailableAt = previous.UnavailableAt
				if lastDetailAt == nil {
					lastDetailAt = previous.LastDetailAt
				}
			}
			if found.Availability != nil && *found.Availability == "unavailable" {
				unavailableAt = &observedAt
			}

			listings = append(listings, &models.Listing{
				ID:            listingID,
				Source:        found.Source,
				ExternalId:    found.ExternalId,
				CanonicalUrl:  found.CanonicalUrl,
				Title:         found.Title,
				ImageUrl:      found.ImageUrl,
				Description:   found.Description,
				Availability:  found.Availability,
				FirstSeenAt:   firstSeenAt,
				LastSeenAt:    observedAt,
				LastDetailAt:  lastDetailAt,
				UnavailableAt: unavailableAt,
			})
		}
		observation := &models.ListingObservation{
			ID:           uuid.New().String(),
			ListingId:    listingID,
			Source:       found.Source,
			ObservedAt:   observedAt,
			Price:        found.Price,
			Shipping:     found.Shipping,
			TotalPrice:   found.TotalPrice,
			Currency:     found.Currency,
			Title:        found.Title,
			ImageUrl:     found.ImageUrl,
			Description:  found.Description,
			Availability: found.Availability,
		}
		if !shouldSkipListingObservation(latestObservations[listingID], observation, observedAt) {
			observations = append(observations, observation)
		}
	}

	statesToPersist := make([]*models.QueryListingState, 0, len(states))
	for _, state := range states {
		if state == nil {
			continue
		}
		previous := existingStates[state.ListingId]
		if previous != nil &&
			queryListingStateEqualIgnoringFreshness(previous, state) &&
			observedAt.Sub(previous.LastEvaluatedAt) < queryListingStateFreshnessWriteInterval {
			continue
		}
		statesToPersist = append(statesToPersist, state)
	}

	return q.UnderlyingDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if len(listings) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				UpdateAll: true,
			}).CreateInBatches(listings, 100).Error; err != nil {
				return err
			}
		}
		if len(observations) > 0 {
			if err := tx.CreateInBatches(observations, 100).Error; err != nil {
				return err
			}
		}
		if len(statesToPersist) == 0 {
			return nil
		}
		return tx.Omit("Query", "Listing").Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "queryId"}, {Name: "listingId"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"source",
				"status",
				"lastEvaluatedAt",
				"firstMatchedAt",
				"lastMatchedAt",
				"lastRejectedReason",
				"lastNotifiedAt",
				"lastNotifiedTotalPrice",
				"lowestObservedPrice",
			}),
		}).CreateInBatches(statesToPersist, 100).Error
	})
}

func shouldSkipListingObservation(previous *models.ListingObservation, next *models.ListingObservation, observedAt time.Time) bool {
	if previous == nil || next == nil {
		return false
	}
	if observedAt.Sub(previous.ObservedAt) >= listingFreshnessWriteInterval {
		return false
	}
	return previous.Source == next.Source &&
		floatPtrEqual(previous.Price, next.Price) &&
		floatPtrEqual(previous.Shipping, next.Shipping) &&
		floatPtrEqual(previous.TotalPrice, next.TotalPrice) &&
		stringPtrEqual(previous.Currency, next.Currency) &&
		previous.Title == next.Title &&
		stringPtrEqual(previous.ImageUrl, next.ImageUrl) &&
		stringPtrEqual(previous.Description, next.Description) &&
		stringPtrEqual(previous.Availability, next.Availability)
}

func stringPtrEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func floatPtrEqual(left, right *float64) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}

func timePtrEqual(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.Equal(*right)
}

func queryListingStateEqualIgnoringFreshness(left, right *models.QueryListingState) bool {
	if left == nil || right == nil {
		return left == right
	}
	return left.QueryId == right.QueryId &&
		left.ListingId == right.ListingId &&
		left.Source == right.Source &&
		left.Status == right.Status &&
		timePtrEqual(left.FirstMatchedAt, right.FirstMatchedAt) &&
		timePtrEqual(left.LastNotifiedAt, right.LastNotifiedAt) &&
		stringPtrEqual(left.LastRejectedReason, right.LastRejectedReason) &&
		floatPtrEqual(left.LastNotifiedTotalPrice, right.LastNotifiedTotalPrice) &&
		floatPtrEqual(left.LowestObservedPrice, right.LowestObservedPrice)
}

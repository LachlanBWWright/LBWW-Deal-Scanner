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

func (q *Query) PersistListingObservation(ctx context.Context, listing DiscoveredListing, observedAt time.Time) (string, error) {
	listingId := StableListingId(listing.Source, listing.CanonicalUrl)

	err := q.Transaction(func(tx *Query) error {
		_, err := tx.Listing.WithContext(ctx).Where(tx.Listing.ID.Eq(listingId)).First()
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
		} else {
			return err
		}

		// Insert observation
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
		return tx.ListingObservation.WithContext(ctx).Create(&obs)
	})

	if err != nil {
		return "", err
	}

	return listingId, nil
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

func (q *Query) UpsertQueryListingState(ctx context.Context, state *models.QueryListingState) error {
	return q.QueryListingState.WithContext(ctx).Save(state)
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

func (q *Query) PersistListingBatch(
	ctx context.Context,
	discovered []DiscoveredListing,
	existing map[string]*models.Listing,
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
		firstSeenAt := observedAt
		lastDetailAt := found.LastDetailAt
		var unavailableAt *time.Time
		if previous := existing[listingID]; previous != nil {
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
		observations = append(observations, &models.ListingObservation{
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
		})
	}

	return q.UnderlyingDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).CreateInBatches(listings, 100).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(observations, 100).Error; err != nil {
			return err
		}
		if len(states) == 0 {
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
		}).CreateInBatches(states, 100).Error
	})
}

package query

import (
	"context"
	"errors"
	"time"

	"dealscanner/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

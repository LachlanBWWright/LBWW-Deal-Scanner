package db

import (
	"context"
	"errors"
	"time"

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

func (db *DB) PersistListingObservation(ctx context.Context, listing DiscoveredListing, observedAt time.Time) (string, error) {
	listingId := StableListingId(listing.Source, listing.CanonicalUrl)

	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing Listing
		err := tx.Where("id = ?", listingId).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			lst := Listing{
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
			if err := tx.Create(&lst).Error; err != nil {
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
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
		} else {
			return err
		}

		// Insert observation
		obs := ListingObservation{
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
		return tx.Create(&obs).Error
	})

	if err != nil {
		return "", err
	}

	return listingId, nil
}

func (db *DB) GetQueryListingState(ctx context.Context, queryId, listingId string) (*QueryListingState, error) {
	var s QueryListingState
	if err := db.WithContext(ctx).Where("queryId = ? AND listingId = ?", queryId, listingId).First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (db *DB) UpsertQueryListingState(ctx context.Context, state *QueryListingState) error {
	// GORM Save will handle upsert correctly
	return db.WithContext(ctx).Save(state).Error
}

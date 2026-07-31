package scanners

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/notifications"
	"gorm.io/gorm"
)

type CashConvertersFilterMatcher struct {
	Required     []string
	RequiredMode string // all/any
	Excluded     []string
	ExcludedMode string // all/any
	MinPrice     *float64
	MaxPrice     *float64
}

func phraseMode(value string) (combine string, wholeWords bool) {
	switch value {
	case "any_words":
		return "any", true
	case "all_words":
		return "all", true
	case "any_substring":
		return "any", false
	case "all_substring":
		return "all", false
	case "any":
		return "any", true
	default:
		return "all", true
	}
}

func containsWholePhrase(searchable string, phrase string) bool {
	start := 0
	for {
		index := strings.Index(searchable[start:], phrase)
		if index < 0 {
			return false
		}
		index += start
		beforeOK := index == 0 || !isPhraseCharacter(rune(searchable[index-1]))
		after := index + len(phrase)
		afterOK := after == len(searchable) || !isPhraseCharacter(rune(searchable[after]))
		if beforeOK && afterOK {
			return true
		}
		start = index + 1
	}
}

func isPhraseCharacter(char rune) bool {
	return char >= 'a' && char <= 'z' || char >= '0' && char <= '9'
}

type MatchResult struct {
	Matched      bool
	RejectReason *string
}

func (m CashConvertersFilterMatcher) Match(title string, description string, totalPrice float64) MatchResult {
	searchable := strings.ToLower(title + " " + description)

	if len(m.Required) > 0 {
		if !phrasesMatch(searchable, m.Required, m.RequiredMode) {
			reason := "MissingRequiredPhrase"
			return MatchResult{Matched: false, RejectReason: &reason}
		}
	}

	if len(m.Excluded) > 0 {
		if phrasesMatch(searchable, m.Excluded, m.ExcludedMode) {
			reason := "ExcludedPhrase"
			return MatchResult{Matched: false, RejectReason: &reason}
		}
	}

	if m.MinPrice != nil && totalPrice < *m.MinPrice {
		reason := "BelowMinPrice"
		return MatchResult{Matched: false, RejectReason: &reason}
	}

	if m.MaxPrice != nil && totalPrice > *m.MaxPrice {
		reason := "AboveMaxPrice"
		return MatchResult{Matched: false, RejectReason: &reason}
	}

	return MatchResult{Matched: true, RejectReason: nil}
}

func (s *CashConvertersScanner) evaluateChangedListings(
	ctx context.Context,
	filters []qry.QueryItem,
	changedListingIDs []string,
	now time.Time,
) ([]notifications.AppNotification, error) {
	if len(changedListingIDs) == 0 || len(filters) == 0 {
		return nil, nil
	}

	db := s.dbClient.UnderlyingDB().WithContext(ctx)

	var searchDocs []models.CashConvertersSearchDoc
	if err := db.Where("listingId IN ?", changedListingIDs).Find(&searchDocs).Error; err != nil {
		return nil, fmt.Errorf("load search docs for matching: %w", err)
	}

	docMap := make(map[string]models.CashConvertersSearchDoc, len(searchDocs))
	for _, doc := range searchDocs {
		docMap[doc.ListingID] = doc
	}

	var notifs []notifications.AppNotification

	for _, listingID := range changedListingIDs {
		doc, exists := docMap[listingID]
		if !exists {
			continue
		}

		totalPrice := 0.0
		if doc.TotalPrice != nil {
			totalPrice = *doc.TotalPrice
		}

		// Don't notify if listing is unavailable or missing
		if doc.Availability != "available" || doc.SourceStatus != "available" {
			for _, filter := range filters {
				var prevState models.QueryListingState
				res := db.Where("queryId = ? AND listingId = ?", filter.QueryId, listingID).First(&prevState)
				if res.Error == nil {
					if err := db.Model(&prevState).Updates(map[string]interface{}{
						"status":          models.QueryListingStateStatusUnavailable,
						"lastEvaluatedAt": now,
					}).Error; err != nil {
						return nil, fmt.Errorf("mark Cash Converters query listing unavailable: %w", err)
					}
				} else if !errors.Is(res.Error, gorm.ErrRecordNotFound) {
					return nil, fmt.Errorf("load Cash Converters query listing state: %w", res.Error)
				}
			}
			continue
		}

		for _, filter := range filters {
			reqPhrases := parsePhrases(valueOr(filter.RequiredPhrases, ""))
			exclPhrases := parsePhrases(valueOr(filter.ExcludePhrases, ""))

			// Catalog summaries do not contain the full description. Evaluating
			// before the individual page is fetched can miss excluded phrases and
			// produce a notification that does not actually satisfy the filter.
			if doc.LastDetailAt == nil {
				continue
			}

			matcher := CashConvertersFilterMatcher{
				Required:     reqPhrases,
				RequiredMode: valueOr(filter.RequiredMatchMode, "all"),
				Excluded:     exclPhrases,
				ExcludedMode: valueOr(filter.ExcludeMatchMode, "any"),
				MinPrice:     filter.MinPrice,
				MaxPrice:     filter.MaxPrice,
			}

			matchRes := matcher.Match(doc.Title, doc.Description, totalPrice)

			var prevState models.QueryListingState
			res := db.Where("queryId = ? AND listingId = ?", filter.QueryId, listingID).First(&prevState)
			if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("load Cash Converters query listing state: %w", res.Error)
			}
			hasPrev := res.Error == nil

			shouldNotify := false
			if matchRes.Matched {
				if !hasPrev || prevState.Status == models.QueryListingStateStatusRejected ||
					prevState.Status == models.QueryListingStateStatusUnavailable ||
					prevState.Status == models.QueryListingStateStatusUnseen {
					shouldNotify = true
				}
			}

			var status models.QueryListingStateStatus
			if matchRes.Matched {
				if shouldNotify {
					status = models.QueryListingStateStatusNotified
				} else {
					status = models.QueryListingStateStatusMatched
				}
			} else {
				status = models.QueryListingStateStatusRejected
			}

			state := models.QueryListingState{
				QueryId:            filter.QueryId,
				ListingId:          listingID,
				Source:             "cashConverters",
				Status:             status,
				LastEvaluatedAt:    now,
				LastRejectedReason: matchRes.RejectReason,
			}

			if hasPrev {
				state.FirstMatchedAt = prevState.FirstMatchedAt
				state.LastMatchedAt = prevState.LastMatchedAt
				state.LastNotifiedAt = prevState.LastNotifiedAt
				state.LastNotifiedTotalPrice = prevState.LastNotifiedTotalPrice
				state.LowestObservedPrice = prevState.LowestObservedPrice

				if matchRes.Matched && state.FirstMatchedAt == nil {
					state.FirstMatchedAt = &now
				}
				if prevState.LowestObservedPrice == nil || totalPrice < *prevState.LowestObservedPrice {
					state.LowestObservedPrice = &totalPrice
				}
			} else {
				state.LowestObservedPrice = &totalPrice
				if matchRes.Matched {
					state.FirstMatchedAt = &now
				}
			}

			if matchRes.Matched {
				state.LastMatchedAt = &now
			}

			if shouldNotify {
				state.LastNotifiedAt = &now
				state.LastNotifiedTotalPrice = &totalPrice

				img := ""
				if doc.ImageURL != nil {
					img = *doc.ImageURL
				}

				titleMsg := fmt.Sprintf("a %s for $%.2f is available at %s", doc.Title, totalPrice, doc.CanonicalURL)
				notifs = append(notifs, notifications.AppNotification{
					Kind:     "deal",
					Source:   "cashConverters",
					Title:    titleMsg,
					Url:      doc.CanonicalURL,
					Price:    &totalPrice,
					ImageUrl: &img,
					Query: &notifications.NotificationQuery{
						Type: "cashConverters",
						Id:   filter.Id,
					},
				})
			}

			if err := db.Save(&state).Error; err != nil {
				return nil, fmt.Errorf("save Cash Converters query listing state: %w", err)
			}
		}
	}

	return notifs, nil
}

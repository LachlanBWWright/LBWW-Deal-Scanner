package query

import (
	"context"
	"strings"
	"time"

	"dealscanner/internal/models"
)

type MatchType string

const (
	MatchTypeMatched  MatchType = "Matched"
	MatchTypeRejected MatchType = "Rejected"
)

type MatchResult struct {
	Type   MatchType
	Reason string
}

type DecisionType string

const (
	DecisionTypeNotify      DecisionType = "Notify"
	DecisionTypeDoNotNotify DecisionType = "DoNotNotify"
)

type DecisionReason string

const (
	DecisionReasonFirstMatch            DecisionReason = "FirstMatch"
	DecisionReasonPriceDroppedIntoRange DecisionReason = "PriceDroppedIntoRange"
	DecisionReasonFurtherPriceDrop      DecisionReason = "FurtherPriceDrop"
	DecisionReasonRejected              DecisionReason = "Rejected"
	DecisionReasonAlreadyNotified       DecisionReason = "AlreadyNotified"
	DecisionReasonUnavailable           DecisionReason = "Unavailable"
	DecisionReasonMissingPrice          DecisionReason = "MissingPrice"
)

type NotificationDecision struct {
	Type   DecisionType
	Reason DecisionReason
}

func ParsePhraseList(phraseList string) []string {
	var results []string
	for _, p := range strings.Split(phraseList, ",") {
		clean := strings.TrimSpace(strings.ToLower(p))
		if clean != "" {
			results = append(results, clean)
		}
	}
	return results
}

func MatchPhrases(text string, requiredPhrases, excludePhrases string) MatchResult {
	lowercaseText := strings.ToLower(text)

	if requiredPhrases != "" {
		required := ParsePhraseList(requiredPhrases)
		for _, p := range required {
			if !strings.Contains(lowercaseText, p) {
				return MatchResult{Type: MatchTypeRejected, Reason: "MissingRequiredPhrase"}
			}
		}
	}

	if excludePhrases != "" {
		excluded := ParsePhraseList(excludePhrases)
		for _, p := range excluded {
			if strings.Contains(lowercaseText, p) {
				return MatchResult{Type: MatchTypeRejected, Reason: "ExcludedPhrase"}
			}
		}
	}

	return MatchResult{Type: MatchTypeMatched}
}

func MatchPriceRange(price float64, minPrice, maxPrice *float64) MatchResult {
	if minPrice != nil && price < *minPrice {
		return MatchResult{Type: MatchTypeRejected, Reason: "BelowMinPrice"}
	}
	if maxPrice != nil && price > *maxPrice {
		return MatchResult{Type: MatchTypeRejected, Reason: "AboveMaxPrice"}
	}
	return MatchResult{Type: MatchTypeMatched}
}

func CombineMatchResults(phraseResult, priceResult MatchResult) MatchResult {
	if phraseResult.Type == MatchTypeRejected {
		return phraseResult
	}
	if priceResult.Type == MatchTypeRejected {
		return priceResult
	}
	return MatchResult{Type: MatchTypeMatched}
}

func (q *Query) EvaluateListingForQuery(
	ctx context.Context,
	queryId string,
	listingId string,
	source string,
	totalPrice *float64,
	match MatchResult,
	evaluatedAt time.Time,
	furtherPriceDropRatio float64,
) (NotificationDecision, error) {
	previous, err := q.QueryListingState.WithContext(ctx).Where(q.QueryListingState.QueryId.Eq(queryId), q.QueryListingState.ListingId.Eq(listingId)).First()
	var previousStatus *string
	var lastNotifiedTotalPrice *float64
	var lowestObsPrice *float64

	if err == nil {
		previousStatus = &previous.Status
		lastNotifiedTotalPrice = previous.LastNotifiedTotalPrice
		lowestObsPrice = previous.LowestObservedPrice
	}

	if match.Type == MatchTypeRejected {
		status := "rejected"
		if match.Reason == "Unavailable" {
			status = "unavailable"
		}

		lowest := lowestPrice(lowestObsPrice, totalPrice)

		state := models.QueryListingState{
			QueryId:             queryId,
			ListingId:           listingId,
			Source:              source,
			Status:              status,
			LastEvaluatedAt:     evaluatedAt,
			LastRejectedReason:  &match.Reason,
			LowestObservedPrice: lowest,
		}

		err = q.UpsertQueryListingState(ctx, &state)
		if err != nil {
			return NotificationDecision{}, err
		}

		reason := DecisionReasonRejected
		if match.Reason == "Unavailable" {
			reason = DecisionReasonUnavailable
		}
		return NotificationDecision{Type: DecisionTypeDoNotNotify, Reason: reason}, nil
	}

	if totalPrice == nil {
		return NotificationDecision{Type: DecisionTypeDoNotNotify, Reason: DecisionReasonMissingPrice}, nil
	}

	decision := getMatchedDecision(previousStatus, lastNotifiedTotalPrice, *totalPrice, furtherPriceDropRatio)

	status := "matched"
	if decision.Type == DecisionTypeNotify {
		status = "notified"
	}

	lowest := lowestPrice(lowestObsPrice, totalPrice)

	state := models.QueryListingState{
		QueryId:             queryId,
		ListingId:           listingId,
		Source:              source,
		Status:              status,
		LastEvaluatedAt:     evaluatedAt,
		LastMatchedAt:       &evaluatedAt,
		LowestObservedPrice: lowest,
	}

	if previousStatus != nil {
		state.FirstMatchedAt = previous.FirstMatchedAt
		state.LastNotifiedAt = previous.LastNotifiedAt
		state.LastNotifiedTotalPrice = previous.LastNotifiedTotalPrice
	} else {
		state.FirstMatchedAt = &evaluatedAt
	}

	if decision.Type == DecisionTypeNotify {
		state.LastNotifiedAt = &evaluatedAt
		state.LastNotifiedTotalPrice = totalPrice
	}

	err = q.UpsertQueryListingState(ctx, &state)
	if err != nil {
		return NotificationDecision{}, err
	}

	return decision, nil
}

func getMatchedDecision(
	previousStatus *string,
	lastNotifiedTotalPrice *float64,
	totalPrice float64,
	furtherPriceDropRatio float64,
) NotificationDecision {
	if previousStatus == nil || *previousStatus == "unseen" {
		return NotificationDecision{Type: DecisionTypeNotify, Reason: DecisionReasonFirstMatch}
	}

	if *previousStatus == "rejected" {
		return NotificationDecision{Type: DecisionTypeNotify, Reason: DecisionReasonPriceDroppedIntoRange}
	}

	if lastNotifiedTotalPrice != nil && furtherPriceDropRatio > 0 && totalPrice < (*lastNotifiedTotalPrice)*(1-furtherPriceDropRatio) {
		return NotificationDecision{Type: DecisionTypeNotify, Reason: DecisionReasonFurtherPriceDrop}
	}

	return NotificationDecision{Type: DecisionTypeDoNotNotify, Reason: DecisionReasonAlreadyNotified}
}

func lowestPrice(prevPrice *float64, currPrice *float64) *float64 {
	if currPrice == nil {
		return prevPrice
	}
	if prevPrice == nil {
		return currPrice
	}
	if *currPrice < *prevPrice {
		return currPrice
	}
	return prevPrice
}

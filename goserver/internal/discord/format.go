package discord

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"dealscanner/internal/db/query"
)

func removeMarkdownCodeFences(message string) string {
	return strings.ReplaceAll(message, "```", "")
}

func formatQueryReference(value string) string {
	cleaned := strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(value))
	parsed, err := url.Parse(cleaned)
	if err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https") {
		return fmt.Sprintf("<%s>", cleaned)
	}

	return fmt.Sprintf("`%s`", strings.ReplaceAll(cleaned, "`", "ˋ"))
}

func formatOptionalPhraseFilter(value *string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return "`None`"
	}

	return formatQueryReference(*value)
}

func formatOptionalMode(value *string, fallback string) string {
	if value == nil || strings.TrimSpace(*value) == "" {
		return fallback
	}

	return strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(*value))
}

func formatOptionalPrice(value *float64) string {
	if value == nil {
		return "Any"
	}
	return fmt.Sprintf("$%.2f", *value)
}

func formatOptionalFloat(value *float64, fallback string) string {
	if value == nil {
		return fallback
	}
	return fmt.Sprintf("%.5f", *value)
}

func formatQueryInfo(q query.QueryItem) string {
	switch q.Type {
	case "ebay", "gumtree":
		return fmt.Sprintf("- URL: %s | Max Price: %s", formatQueryUrl(q), formatOptionalPrice(q.MaxPrice))
	case "cashConverters":
		return formatCashConvertersQueryInfo(q)
	case "salvos":
		return fmt.Sprintf("- Name: %s | Price Range: %s - %s", formatQueryName(q), formatOptionalPrice(q.MinPrice), formatOptionalPrice(q.MaxPrice))
	case "csMarket":
		displayUrl := ""
		if q.DisplayUrl != nil && *q.DisplayUrl != "" && (q.Url == nil || *q.DisplayUrl != *q.Url) {
			displayUrl = fmt.Sprintf(" | Display URL: %s", formatQueryReference(*q.DisplayUrl))
		}
		return fmt.Sprintf("- URL: %s%s | Max Price: %s | Max Float: %s", formatQueryUrl(q), displayUrl, formatOptionalPrice(q.MaxPrice), formatOptionalFloat(q.MaxFloat, "Any"))
	case "csTradeBot":
		return fmt.Sprintf("- Name: %s | Max Price: %s | Float Range: %s - %s", formatQueryName(q), formatOptionalPrice(q.MaxPrice), formatOptionalFloat(q.MinFloat, "Any"), formatOptionalFloat(q.MaxFloat, "Any"))
	case "steamMarket":
		marketLink := ""
		if q.DisplayUrl != nil && *q.DisplayUrl != "" {
			marketLink = fmt.Sprintf(" | Market Link: %s", formatQueryReference(*q.DisplayUrl))
		}
		return fmt.Sprintf("- Name: %s | Max Price: %s%s", formatQueryName(q), formatOptionalPrice(q.MaxPrice), marketLink)
	default:
		return fmt.Sprintf("- ID: %s", formatQueryReference(q.Id))
	}
}

func formatQueryUrl(q query.QueryItem) string {
	if q.Url != nil && *q.Url != "" {
		return formatQueryReference(*q.Url)
	}
	return formatQueryReference(q.Id)
}

func formatQueryName(q query.QueryItem) string {
	if q.Name != nil && *q.Name != "" {
		return formatQueryReference(*q.Name)
	}
	return formatQueryReference(q.Id)
}

func formatCashConvertersQueryInfo(q query.QueryItem) string {
	urlRef := "missing URL"
	if q.Url != nil && *q.Url != "" {
		urlRef = formatQueryReference(*q.Url)
	}

	return fmt.Sprintf(
		"- URL: %s | ID: `%s` | Price Range: %s - %s | Required: %s (%s) | Excluded: %s (%s) | Scan Mode: `%s`",
		urlRef,
		q.Id,
		formatOptionalPrice(q.MinPrice),
		formatOptionalPrice(q.MaxPrice),
		formatOptionalPhraseFilter(q.RequiredPhrases),
		formatOptionalMode(q.RequiredMatchMode, "all"),
		formatOptionalPhraseFilter(q.ExcludePhrases),
		formatOptionalMode(q.ExcludeMatchMode, "any"),
		formatOptionalMode(q.ScanMode, "searchUrl"),
	)
}

func (b *Bot) formatQueryDetails(ctx context.Context, queryType string, queryId string) string {
	queries, err := b.dbClient.ListSavedQueries(ctx, queryType)
	if err != nil {
		return fmt.Sprintf("❌ Failed to load query details: %v", err)
	}

	for _, q := range queries {
		if queryMatchesID(q, queryId) {
			return truncateDiscordMessage(fmt.Sprintf("**%s Query Details**\n%s", formatQueryTypeLabel(queryType), strings.TrimPrefix(formatQueryInfo(q), "- ")))
		}
	}

	return "❌ Query not found."
}

func (b *Bot) deleteSavedQueryWithDetails(ctx context.Context, queryType string, queryId string) (string, error) {
	queries, err := b.dbClient.ListSavedQueries(ctx, queryType)
	if err != nil {
		return "", fmt.Errorf("load query before deletion: %w", err)
	}

	var details string
	for _, q := range queries {
		if queryMatchesID(q, queryId) {
			details = strings.TrimPrefix(formatQueryInfo(q), "- ")
			break
		}
	}

	if err := b.dbClient.DeleteSavedQuery(ctx, queryType, queryId); err != nil {
		return "", err
	}

	if details == "" {
		details = fmt.Sprintf("ID: %s", formatQueryReference(queryId))
	}
	return details, nil
}

func queryMatchesID(q query.QueryItem, queryId string) bool {
	return q.Id == queryId ||
		q.QueryId == queryId ||
		(q.Url != nil && *q.Url == queryId) ||
		(q.Name != nil && *q.Name == queryId)
}

func formatQueryTypeLabel(queryType string) string {
	switch queryType {
	case "cashConverters":
		return "Cash Converters"
	case "ebay":
		return "eBay"
	case "gumtree":
		return "Gumtree"
	case "salvos":
		return "Salvos"
	case "csMarket":
		return "CS Market"
	case "csTradeBot":
		return "CS Trade Bot"
	case "steamMarket":
		return "Steam Market"
	default:
		return queryType
	}
}

package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/scanners"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

const cashSearchPageAction = "cash_search_page"

func cashSearchInputFromOptions(options map[string]*discordgo.ApplicationCommandInteractionDataOption) qry.CashConvertersCatalogSearchInput {
	text := ""
	if opt, ok := options["text"]; ok && opt != nil {
		text = opt.StringValue()
	}

	var minPrice *float64
	if opt, ok := options["minprice"]; ok && opt != nil {
		val := opt.FloatValue()
		minPrice = &val
	}

	var maxPrice *float64
	if opt, ok := options["maxprice"]; ok && opt != nil {
		val := opt.FloatValue()
		maxPrice = &val
	}

	availableOnly := true
	if opt, ok := options["availableonly"]; ok && opt != nil {
		availableOnly = opt.BoolValue()
	}

	limit := 5
	if opt, ok := options["limit"]; ok && opt != nil {
		limit = int(opt.IntValue())
	}

	sortOrder := "newest"
	if opt, ok := options["sort"]; ok && opt != nil {
		sortOrder = opt.StringValue()
	}

	return qry.CashConvertersCatalogSearchInput{
		Text:          text,
		MinPrice:      minPrice,
		MaxPrice:      maxPrice,
		AvailableOnly: availableOnly,
		Limit:         limit,
		Required:      getOptionalString(options, "required", ""),
		RequiredMode:  getOptionalString(options, "requiredmode", "all_words"),
		Excluded:      getOptionalString(options, "excluded", ""),
		ExcludedMode:  getOptionalString(options, "excludedmode", "any_words"),
		Sort:          sortOrder,
	}
}

func encodeCashSearchInput(input qry.CashConvertersCatalogSearchInput) (string, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("encode Cash Converters search: %w", err)
	}
	return string(encoded), nil
}

func decodeCashSearchInput(encoded string) (qry.CashConvertersCatalogSearchInput, error) {
	var input qry.CashConvertersCatalogSearchInput
	if err := json.Unmarshal([]byte(encoded), &input); err != nil {
		return qry.CashConvertersCatalogSearchInput{}, fmt.Errorf("decode Cash Converters search: %w", err)
	}
	return input, nil
}

func liveCashSearchSort(sort string) (string, bool) {
	switch sort {
	case "newest":
		return "newest", true
	case "price_asc":
		return "price", true
	case "relevance", "":
		return "relevance", true
	default:
		return "", false
	}
}

func liveCashSearchEligible(input qry.CashConvertersCatalogSearchInput) bool {
	_, sortSupported := liveCashSearchSort(input.Sort)
	return strings.TrimSpace(input.Text) != "" &&
		input.MinPrice == nil &&
		input.MaxPrice == nil &&
		input.AvailableOnly &&
		sortSupported
}

func cashSearchResultsFromSummaries(
	summaries []qry.CashConvertersDiscoveredSummary,
	now time.Time,
) []qry.CashConvertersCatalogSearchResult {
	results := make([]qry.CashConvertersCatalogSearchResult, 0, len(summaries))
	for _, summary := range summaries {
		totalPrice := summary.TotalPrice
		imageURL := summary.ImageURL
		results = append(results, qry.CashConvertersCatalogSearchResult{
			ListingID:    "cashConverters:" + summary.CanonicalURL,
			Title:        summary.Title,
			CanonicalURL: summary.CanonicalURL,
			ImageURL:     &imageURL,
			TotalPrice:   &totalPrice,
			Availability: "available",
			SourceStatus: "available",
			LastSeenAt:   now,
		})
	}
	return results
}

func fetchLiveCashSearchDisplayPage(
	ctx context.Context,
	input qry.CashConvertersCatalogSearchInput,
	displayPage int,
) ([]qry.CashConvertersCatalogSearchResult, int, error) {
	itemsPerPage := input.Limit
	if itemsPerPage < 1 {
		itemsPerPage = 5
	}
	if itemsPerPage > 5 {
		itemsPerPage = 5
	}
	if displayPage < 1 {
		displayPage = 1
	}

	sort, _ := liveCashSearchSort(input.Sort)
	firstOffset := (displayPage - 1) * itemsPerPage
	apiPage := firstOffset/24 + 1
	start := firstOffset % 24
	client := &http.Client{Timeout: 20 * time.Second}
	first, totalItems, err := scanners.FetchCashConvertersSearchPage(
		ctx,
		client,
		input.Text,
		sort,
		apiPage,
	)
	if err != nil {
		return nil, 0, err
	}

	combined := first
	if start+itemsPerPage > len(first) && apiPage*24 < totalItems {
		second, _, secondErr := scanners.FetchCashConvertersSearchPage(
			ctx,
			client,
			input.Text,
			sort,
			apiPage+1,
		)
		if secondErr != nil {
			return nil, 0, secondErr
		}
		combined = append(combined, second...)
	}

	if start > len(combined) {
		start = len(combined)
	}
	end := start + itemsPerPage
	if end > len(combined) {
		end = len(combined)
	}
	totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage
	return cashSearchResultsFromSummaries(combined[start:end], time.Now().UTC()), totalPages, nil
}

func formatCashSearchEntries(results []qry.CashConvertersCatalogSearchResult) []string {
	entries := make([]string, 0, len(results))
	for idx, res := range results {
		priceStr := "N/A"
		if res.TotalPrice != nil {
			priceStr = fmt.Sprintf("$%.2f", *res.TotalPrice)
		}

		lastSeen := res.LastSeenAt.Format("2006-01-02 15:04 UTC")
		statusStr := res.Availability
		if res.SourceStatus != "" && res.SourceStatus != "available" {
			statusStr = res.SourceStatus
		}

		var entry strings.Builder
		entry.WriteString(fmt.Sprintf("%d. [%s](<%s>) - %s\n", idx+1, res.Title, res.CanonicalURL, priceStr))
		entry.WriteString(fmt.Sprintf("   *Last seen: %s | Status: %s*\n", lastSeen, statusStr))

		if res.Description != "" {
			descriptionRunes := []rune(res.Description)
			if len(descriptionRunes) > 180 {
				descriptionRunes = append(descriptionRunes[:177], '.', '.', '.')
			}
			entry.WriteString(fmt.Sprintf("   \"%s\"\n", string(descriptionRunes)))
		}
		entry.WriteString("\n")
		entries = append(entries, entry.String())
	}
	return entries
}

func formatLiveCashSearchEntries(results []qry.CashConvertersCatalogSearchResult) string {
	var content strings.Builder
	for idx, result := range results {
		titleRunes := []rune(result.Title)
		if len(titleRunes) > 100 {
			titleRunes = append(titleRunes[:97], '.', '.', '.')
		}
		price := "N/A"
		if result.TotalPrice != nil {
			price = fmt.Sprintf("$%.2f", *result.TotalPrice)
		}
		content.WriteString(fmt.Sprintf(
			"%d. [%s](<%s>) — %s\n",
			idx+1,
			string(titleRunes),
			result.CanonicalURL,
			price,
		))
	}
	return content.String()
}

func (b *Bot) sendPaginatedCashSearch(
	ctx context.Context,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
	input qry.CashConvertersCatalogSearchInput,
	page int,
) {
	itemsPerPage := input.Limit
	if itemsPerPage < 1 || itemsPerPage > 5 {
		itemsPerPage = 5
	}
	if page < 1 {
		page = 1
	}

	totalResults, err := b.dbClient.CountCashConvertersCatalog(ctx, input)
	if err != nil {
		content := removeMarkdownCodeFences(fmt.Sprintf("❌ Error searching Cash Converters catalog: %v", err))
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}
	totalPages := int((totalResults + int64(itemsPerPage) - 1) / int64(itemsPerPage))
	if totalPages > 0 && page > totalPages {
		page = totalPages
	}

	pageInput := input
	pageInput.Limit = itemsPerPage
	pageInput.Offset = (page - 1) * itemsPerPage
	results, err := b.dbClient.SearchCashConvertersCatalog(ctx, pageInput)
	if err != nil {
		content := removeMarkdownCodeFences(fmt.Sprintf("❌ Error searching Cash Converters catalog: %v", err))
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	if len(results) == 0 {
		content := removeMarkdownCodeFences(fmt.Sprintf("🔍 No Cash Converters items matched **%s**.", input.Text))
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	resultLabel := fmt.Sprintf("%d results — page %d of %d", totalResults, page, totalPages)
	title := fmt.Sprintf("Cash Converters search: \"%s\" — %s", input.Text, resultLabel)
	content := removeMarkdownCodeFences(title + "\n\n" + strings.Join(formatCashSearchEntries(results), ""))
	components := b.cashSearchPaginationComponents(ctx, input, page, totalPages)
	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &components,
	}); err != nil {
		log.Printf("Failed to send paginated Cash Converters search: %v", err)
	}
}

func (b *Bot) cashSearchPaginationComponents(
	ctx context.Context,
	input qry.CashConvertersCatalogSearchInput,
	page int,
	totalPages int,
) []discordgo.MessageComponent {
	if totalPages <= 1 {
		return nil
	}

	payload, err := encodeCashSearchInput(input)
	if err != nil {
		log.Printf("Failed to encode Cash Converters pagination action: %v", err)
		return nil
	}

	backPage := page - 1
	if backPage < 1 {
		backPage = 1
	}
	forwardPage := page + 1
	if forwardPage > totalPages {
		forwardPage = totalPages
	}

	pageTargets := []int{1, backPage, forwardPage, totalPages}
	keys := make([]string, len(pageTargets))
	now := time.Now().UnixMilli()
	for idx, targetPage := range pageTargets {
		key := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		keys[idx] = key
		if err := b.dbClient.SetAction(ctx, key, models.ActionRegistry{
			ID:        key,
			Type:      cashSearchPageAction,
			QueryType: &payload,
			QueryId:   stringPtr(strconv.Itoa(targetPage)),
			Timestamp: now,
		}); err != nil {
			log.Printf("Failed to register Cash Converters page action: %v", err)
		}
	}

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "⏮️ First",
					Style:    discordgo.SecondaryButton,
					CustomID: keys[0],
					Disabled: page == 1,
				},
				discordgo.Button{
					Label:    "◀️ Back",
					Style:    discordgo.SecondaryButton,
					CustomID: keys[1],
					Disabled: page == 1,
				},
				discordgo.Button{
					Label:    "▶️ Forward",
					Style:    discordgo.SecondaryButton,
					CustomID: keys[2],
					Disabled: page == totalPages,
				},
				discordgo.Button{
					Label:    "⏭️ Last",
					Style:    discordgo.SecondaryButton,
					CustomID: keys[3],
					Disabled: page == totalPages,
				},
			},
		},
	}
}

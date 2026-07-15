package discord

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func getStringOption(opts map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	if opt, ok := opts[name]; ok && opt != nil {
		val := opt.StringValue()
		val = strings.ReplaceAll(val, "\r", "")
		val = strings.ReplaceAll(val, "\n", "")
		return strings.TrimSpace(val)
	}
	return ""
}

func getFloatOption(opts map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) float64 {
	if opt, ok := opts[name]; ok && opt != nil {
		return opt.FloatValue()
	}
	return 0
}

func getBoolOption(opts map[string]*discordgo.ApplicationCommandInteractionDataOption, name string) bool {
	if opt, ok := opts[name]; ok && opt != nil {
		return opt.BoolValue()
	}
	return false
}

func (b *Bot) formatQueriesWithFoundItems(ctx context.Context, queries []query.QueryItem, title string) string {
	queryIds := make([]string, 0, len(queries))
	for _, q := range queries {
		queryIds = append(queryIds, q.QueryId)
	}

	foundMap, err := b.dbClient.GetLastFoundItemsBatch(ctx, queryIds, 3)
	if err != nil {
		log.Printf("Failed to batch get last found items: %v", err)
		foundMap = make(map[string][]query.FoundItem)
	}

	entries := make([]string, 0, len(queries))
	for _, q := range queries {
		var sb strings.Builder
		sb.WriteString(formatQueryInfo(q))
		if q.DmOnly {
			sb.WriteString(" (DM Only)")
		}
		sb.WriteString("\n\n")

		found := foundMap[q.QueryId]
		if len(found) > 0 {
			sb.WriteString("  *Last Found:*\n")
			for _, item := range found {
				titleClean := strings.Join(strings.Fields(item.Title), " ")
				urlClean := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(item.Url, "\r", ""), "\n", ""))
				sb.WriteString(fmt.Sprintf("  • [%s](<%s>) - $%.2f\n", titleClean, urlClean, item.Price))
			}
		} else {
			sb.WriteString("  *Last Found:* None\n")
		}
		sb.WriteString("\n")
		entries = append(entries, sb.String())
	}

	return fmt.Sprintf("**%s**\n%s", title, strings.Join(entries, ""))
}

func (b *Bot) sendPaginatedQueries(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, queryType string, title string, page int) {
	queries, err := b.dbClient.ListSavedQueries(ctx, queryType)
	if err != nil {
		content := fmt.Sprintf("❌ Error: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}
	if queries == nil {
		queries = []query.QueryItem{}
	}

	total := len(queries)
	if total == 0 {
		var content string
		if queryType == "csTradeBot" {
			content = "No saved CS Trade Bot multisearch queries found."
		} else if queryType == "steamMarket" {
			content = "No saved Steam SCM queries found."
		} else {
			name := queryType
			if name == "cashConverters" {
				name = "Cash Converters"
			} else if name == "csMarket" {
				name = "CS Market"
			} else {
				name = strings.Title(name)
			}
			content = fmt.Sprintf("No saved %s queries found.", name)
		}
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	formattedQueries := make([]string, 0, total)
	for _, savedQuery := range queries {
		formattedQueries = append(formattedQueries, b.formatQueriesWithFoundItems(ctx, []query.QueryItem{savedQuery}, ""))
	}
	pages := paginateQueryEntries(title, formattedQueries)
	totalPages := len(pages)
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	content := pages[page-1]

	var components []discordgo.MessageComponent
	if totalPages > 1 {
		firstKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		backKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		forwardKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		lastKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]

		now := time.Now().UnixMilli()

		if err := b.dbClient.SetAction(ctx, firstKey, models.ActionRegistry{
			ID:        firstKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr("1"),
			Timestamp: now,
		}); err != nil {
			log.Printf("Failed to register Discord first page action: %v", err)
		}

		backPage := page - 1
		if backPage < 1 {
			backPage = 1
		}
		if err := b.dbClient.SetAction(ctx, backKey, models.ActionRegistry{
			ID:        backKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr(strconv.Itoa(backPage)),
			Timestamp: now,
		}); err != nil {
			log.Printf("Failed to register Discord back page action: %v", err)
		}

		forwardPage := page + 1
		if forwardPage > totalPages {
			forwardPage = totalPages
		}
		if err := b.dbClient.SetAction(ctx, forwardKey, models.ActionRegistry{
			ID:        forwardKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr(strconv.Itoa(forwardPage)),
			Timestamp: now,
		}); err != nil {
			log.Printf("Failed to register Discord forward page action: %v", err)
		}

		if err := b.dbClient.SetAction(ctx, lastKey, models.ActionRegistry{
			ID:        lastKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr(strconv.Itoa(totalPages)),
			Timestamp: now,
		}); err != nil {
			log.Printf("Failed to register Discord last page action: %v", err)
		}

		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "⏮️ First",
						Style:    discordgo.SecondaryButton,
						CustomID: firstKey,
						Disabled: page == 1,
					},
					discordgo.Button{
						Label:    "◀️ Back",
						Style:    discordgo.SecondaryButton,
						CustomID: backKey,
						Disabled: page == 1,
					},
					discordgo.Button{
						Label:    "▶️ Forward",
						Style:    discordgo.SecondaryButton,
						CustomID: forwardKey,
						Disabled: page == totalPages,
					},
					discordgo.Button{
						Label:    "⏭️ Last",
						Style:    discordgo.SecondaryButton,
						CustomID: lastKey,
						Disabled: page == totalPages,
					},
				},
			},
		}
	}

	if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &components,
	}); err != nil {
		log.Printf("Failed to send paginated %s queries: %v", queryType, err)
	}
}

func stringPtr(s string) *string {
	return &s
}

const discordMessageLimit = 2000

func paginateQueryEntries(title string, entries []string) []string {
	if len(entries) == 0 {
		return []string{fmt.Sprintf("**%s (Page 1/1)**\n", title)}
	}

	maxHeader := fmt.Sprintf("**%s (Page %d/%d)**\n", title, len(entries), len(entries))
	contentLimit := discordMessageLimit - len([]rune(maxHeader))
	rawPages := make([]string, 0)
	var page strings.Builder

	for _, entry := range entries {
		cleanEntry := strings.TrimPrefix(entry, "** **\n")
		entryRunes := []rune(cleanEntry)
		if page.Len() > 0 && len([]rune(page.String()))+len(entryRunes) > contentLimit {
			rawPages = append(rawPages, page.String())
			page.Reset()
		}
		if len(entryRunes) > contentLimit {
			cleanEntry = truncateToRuneLimit(cleanEntry, contentLimit)
		}
		page.WriteString(cleanEntry)
	}
	if page.Len() > 0 {
		rawPages = append(rawPages, page.String())
	}

	pages := make([]string, 0, len(rawPages))
	for index, rawPage := range rawPages {
		header := fmt.Sprintf("**%s (Page %d/%d)**\n", title, index+1, len(rawPages))
		pages = append(pages, header+rawPage)
	}
	return pages
}

func truncateDiscordMessage(content string) string {
	return truncateToRuneLimit(content, discordMessageLimit)
}

func truncateToRuneLimit(content string, limit int) string {
	runes := []rune(content)
	if len(runes) <= limit {
		return content
	}

	suffix := "\n\n… More query details were omitted from this page."
	suffixRunes := []rune(suffix)
	return string(runes[:limit-len(suffixRunes)]) + suffix
}

func combinePhraseFilters(first, second string) string {
	switch {
	case first == "":
		return second
	case second == "":
		return first
	default:
		return first + "," + second
	}
}

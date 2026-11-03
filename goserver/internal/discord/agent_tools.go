package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"github.com/bwmarrin/discordgo"
)

var agentPromptNames = []string{"find-opportunities", "review-query-prices", "improve-matching", "find-duplicates", "review-portfolio"}

func agentPrompt(name string) (string, error) {
	goals := map[string]string{
		"find-opportunities":  "Research opportunities in the category I specify and propose new searches for the supported sites.",
		"review-query-prices": "Review busy, quiet, and near-miss searches. Research comparable sold prices and propose justified threshold changes.",
		"improve-matching":    "Review query phrases and examples and propose better matching and exclusions.",
		"find-duplicates":     "Find overlapping searches and recommend consolidation; do not propose deletions.",
		"review-portfolio":    "Review the search portfolio and prioritize evidence-backed improvements.",
	}
	goal, ok := goals[name]
	if !ok {
		return "", fmt.Errorf("unknown prompt %q; use /agentprompts", name)
	}
	return fmt.Sprintf(`%s

Use the DealScanner query data supplied below or in an attached export. Research Australian prices where relevant. Distinguish sold prices from asking prices, identify dates, sources, sample size, condition, shipping and fees. Do not invent evidence. Explain uncertainty and recommend no change where evidence is insufficient. A high alert count does not prove a threshold is too high; a quiet search does not prove it is too low. Use scanner coverage and near-miss data only where available. Preserve queryId and revision for updates; do not invent IDs. Never propose deletions. Return one JSON object and no Markdown, using the requested changes format.`, goal), nil
}

type agentQuery struct {
	query.QueryItem
	Revision string `json:"revision"`
}
type queryActivity struct {
	QueryID       string     `json:"queryId"`
	Evaluated     int        `json:"distinctListingsLastEvaluatedInPeriod"`
	LastNotified  int        `json:"distinctListingsLastNotifiedInPeriod"`
	AboveMax      int        `json:"currentlyRejectedAboveMaxLastEvaluatedInPeriod"`
	LastEvaluated *time.Time `json:"lastEvaluatedAt,omitempty"`
}
type agentBundle struct {
	Format      string          `json:"format"`
	Version     int             `json:"version"`
	GeneratedAt time.Time       `json:"generatedAt"`
	PeriodDays  int             `json:"periodDays"`
	Limitations string          `json:"limitations"`
	Queries     []agentQuery    `json:"queries"`
	Activity    []queryActivity `json:"activity"`
}

func (b *Bot) makeAgentBundle(ctx context.Context, site string, days int) (agentBundle, error) {
	now := time.Now().UTC()
	bundle := agentBundle{Format: "dealscanner.agent-context", Version: 1, GeneratedAt: now, PeriodDays: days, Limitations: "Latest per-listing state only; counts are not notification delivery events or historical evaluation totals. Zero may mean no coverage. Near-miss counts and scan coverage are unavailable.", Queries: []agentQuery{}, Activity: []queryActivity{}}
	queries, err := b.dbClient.ListSavedQueries(ctx, site)
	if err != nil {
		return bundle, err
	}
	var states []models.QueryListingState
	if err := b.dbClient.UnderlyingDB().WithContext(ctx).Find(&states).Error; err != nil {
		return bundle, err
	}
	counts := make(map[string]*queryActivity, len(queries))
	for _, q := range queries {
		rev, err := queryRevision(q)
		if err != nil {
			return bundle, err
		}
		bundle.Queries = append(bundle.Queries, agentQuery{QueryItem: q, Revision: rev})
		counts[q.QueryId] = &queryActivity{QueryID: q.QueryId}
	}
	cutoff := now.Add(-time.Duration(days) * 24 * time.Hour)
	for _, state := range states {
		a := counts[state.QueryId]
		if a == nil {
			continue
		}
		if a.LastEvaluated == nil || state.LastEvaluatedAt.After(*a.LastEvaluated) {
			t := state.LastEvaluatedAt
			a.LastEvaluated = &t
		}
		if !state.LastEvaluatedAt.Before(cutoff) {
			a.Evaluated++
			if state.LastRejectedReason != nil && *state.LastRejectedReason == "AboveMaxPrice" {
				a.AboveMax++
			}
		}
		if state.LastNotifiedAt != nil && !state.LastNotifiedAt.Before(cutoff) {
			a.LastNotified++
		}
	}
	for _, q := range queries {
		bundle.Activity = append(bundle.Activity, *counts[q.QueryId])
	}
	sort.SliceStable(bundle.Activity, func(i, j int) bool { return bundle.Activity[i].LastNotified > bundle.Activity[j].LastNotified })
	return bundle, nil
}

func sendAgentFile(s *discordgo.Session, i *discordgo.InteractionCreate, content, name string, data []byte) {
	_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content, Files: []*discordgo.File{{Name: name, ContentType: "text/plain; charset=utf-8", Reader: strings.NewReader(string(data))}}})
}
func (b *Bot) handleAgentCommand(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, name string, options map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	site := getOptionalString(options, "site", "")
	days := 30
	if opt := options["days"]; opt != nil {
		days = int(opt.IntValue())
	}
	if days < 1 || days > 365 {
		content := "Days must be between 1 and 365."
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}
	switch name {
	case "agentprompts":
		content := "Available prompts: `" + strings.Join(agentPromptNames, "`, `") + "`. Use `/agentprompt` to prepare one with current data."
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	case "agentprompt", "agentexport", "querysummary":
		bundle, err := b.makeAgentBundle(ctx, site, days)
		if err != nil {
			content := fmt.Sprintf("Could not build query data: %v", err)
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
			return
		}
		data, err := json.MarshalIndent(bundle, "", "  ")
		if err != nil {
			content := fmt.Sprintf("Could not encode query data: %v", err)
			_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
			return
		}
		if name == "agentprompt" {
			prompt, err := agentPrompt(getOptionalString(options, "name", "review-query-prices"))
			if err != nil {
				content := err.Error()
				_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
				return
			}
			text := prompt + "\n\n" + agentChangesInstructions + "\n\nCURRENT DATA (JSON):\n" + string(data)
			sendAgentFile(s, i, "Copy the attached prompt into your chat service. Return the JSON using `/agentimport`.", "agent-prompt.txt", []byte(text))
			return
		}
		content := fmt.Sprintf("%d queries; %d-day snapshot. Counts describe latest listing states, not notification deliveries. Coverage and near misses are unavailable.", len(bundle.Queries), days)
		if name == "querysummary" {
			var lines []string
			for idx, a := range bundle.Activity {
				if idx >= 8 {
					break
				}
				lines = append(lines, fmt.Sprintf("`%s`: %d last notified, %d evaluated, %d above max", a.QueryID, a.LastNotified, a.Evaluated, a.AboveMax))
			}
			if len(lines) > 0 {
				content += "\n" + strings.Join(lines, "\n")
			}
			quiet := []string{}
			for _, a := range bundle.Activity {
				if a.LastNotified == 0 && a.Evaluated > 0 && len(quiet) < 5 {
					quiet = append(quiet, "`"+a.QueryID+"`")
				}
			}
			if len(quiet) > 0 {
				content += "\nQuiet with evaluated listings: " + strings.Join(quiet, ", ")
			}
		}
		sendAgentFile(s, i, content, "query-context.json", data)
	}
}

func agentContextOptions() []*discordgo.ApplicationCommandOption {
	choices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, site := range []string{"cashConverters", "ebay", "gumtree", "salvos", "csMarket", "steamMarket", "csTradeBot"} {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{Name: site, Value: site})
	}
	return []*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "site", Description: "Limit to a site", Choices: choices}, {Type: discordgo.ApplicationCommandOptionInteger, Name: "days", Description: "Lookback in days (1 to 365)"}}
}
func agentPromptOptions() []*discordgo.ApplicationCommandOption {
	choices := []*discordgo.ApplicationCommandOptionChoice{}
	for _, name := range agentPromptNames {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{Name: name, Value: name})
	}
	return append([]*discordgo.ApplicationCommandOption{{Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "Prompt template", Required: true, Choices: choices}}, agentContextOptions()...)
}

package discord

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"dealscanner/internal/db/query"
	"dealscanner/internal/models"
)

const agentChangesInstructions = `Return JSON in this format: {"format":"dealscanner.query-changes","version":1,"changes":[{"operation":"update","queryId":"EXPORTED_QUERY_ID","revision":"EXPORTED_REVISION","query":{"type":"ebay","url":"https://www.ebay.com.au/...","maxPrice":30,"dmOnly":false},"rationale":"..."}]}. For creates use operation create and omit queryId and revision. Provide complete query fields, not a patch. Supported types: cashConverters, ebay, gumtree, salvos, csMarket, steamMarket, csTradeBot. eBay, Gumtree and CS Market require url; Salvos, Steam and CS Trade Bot require name. maxPrice is required except for Cash Converters, where it can be null. Cash Converters uses scanMode (siteWide only), requiredPhrases and excludePhrases as comma-separated strings, requiredMatchMode and excludeMatchMode (all_words, any_words, all_substring, any_substring). Salvos needs minPrice; CS Market needs maxFloat; CS Trade Bot needs minFloat and maxFloat. Updates must include all fields, including null minPrice/maxPrice where unset and empty strings for unset phrase fields. Preserve all relevant fields on updates. Prices are AUD where appropriate. Do not add unrecognized fields. Return no deletions.`

type agentChange struct {
	Operation string          `json:"operation"`
	QueryID   string          `json:"queryId,omitempty"`
	Revision  string          `json:"revision,omitempty"`
	Query     query.QueryItem `json:"query"`
	Rationale string          `json:"rationale,omitempty"`
}
type agentChanges struct {
	Format  string        `json:"format"`
	Version int           `json:"version"`
	Changes []agentChange `json:"changes"`
}

func queryRevision(q query.QueryItem) (string, error) {
	data, err := json.Marshal(q)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
func parseAgentChanges(input string) (agentChanges, error) {
	var changes agentChanges
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "```") {
		input = strings.TrimSpace(strings.TrimPrefix(input, "```json"))
		input = strings.TrimSpace(strings.TrimPrefix(input, "```"))
		input = strings.TrimSpace(strings.TrimSuffix(input, "```"))
	}
	dec := json.NewDecoder(strings.NewReader(input))
	dec.DisallowUnknownFields()
	var shape struct {
		Changes []struct {
			Query map[string]json.RawMessage `json:"query"`
		} `json:"changes"`
	}
	if err := json.Unmarshal([]byte(input), &shape); err != nil {
		return changes, fmt.Errorf("invalid JSON: %w", err)
	}
	if err := dec.Decode(&changes); err != nil {
		return changes, fmt.Errorf("invalid JSON: %w", err)
	}
	var extra interface{}
	if err := dec.Decode(&extra); err != io.EOF {
		return changes, fmt.Errorf("expected one JSON object")
	}
	if changes.Format != "dealscanner.query-changes" || changes.Version != 1 {
		return changes, fmt.Errorf("expected dealscanner.query-changes version 1")
	}
	if len(changes.Changes) == 0 || len(changes.Changes) > 50 {
		return changes, fmt.Errorf("provide 1 to 50 changes")
	}
	seen := map[string]bool{}
	for n, c := range changes.Changes {
		if c.Operation != "create" && c.Operation != "update" {
			return changes, fmt.Errorf("change %d: operation must be create or update", n+1)
		}
		if c.Operation == "update" {
			if c.QueryID == "" || c.Revision == "" {
				return changes, fmt.Errorf("change %d: update requires queryId and revision", n+1)
			}
			if seen[c.QueryID] {
				return changes, fmt.Errorf("change %d: duplicate queryId", n+1)
			}
			seen[c.QueryID] = true
		} else if c.QueryID != "" || c.Revision != "" {
			return changes, fmt.Errorf("change %d: create must omit queryId and revision", n+1)
		}
		q := c.Query
		fields := shape.Changes[n].Query
		if c.Operation == "update" {
			required := map[string][]string{"cashConverters": {"scanMode", "requiredPhrases", "requiredMatchMode", "excludePhrases", "excludeMatchMode", "minPrice", "maxPrice"}, "csMarket": {"displayUrl"}, "steamMarket": {"displayUrl"}}
			for _, field := range required[q.Type] {
				if _, ok := fields[field]; !ok {
					return changes, fmt.Errorf("change %d: update must include %s to preserve existing settings", n+1, field)
				}
			}
		}
		if _, ok := shape.Changes[n].Query["dmOnly"]; !ok {
			return changes, fmt.Errorf("change %d: dmOnly is required", n+1)
		}
		switch q.Type {
		case "ebay", "gumtree", "csMarket":
			if q.Url == nil {
				return changes, fmt.Errorf("change %d: url is required", n+1)
			}
			u, err := url.Parse(*q.Url)
			if err != nil || u.Scheme != "https" || u.Hostname() == "" {
				return changes, fmt.Errorf("change %d: a valid HTTPS URL is required", n+1)
			}
			host := strings.ToLower(u.Hostname())
			allowed := map[string]string{"ebay": "ebay.com.au", "gumtree": "gumtree.com.au", "csMarket": "steamcommunity.com"}[q.Type]
			if host != allowed && !strings.HasSuffix(host, "."+allowed) {
				return changes, fmt.Errorf("change %d: URL must be on %s", n+1, allowed)
			}
			if u.User != nil || u.Port() != "" {
				return changes, fmt.Errorf("change %d: URL credentials and ports are not supported", n+1)
			}
		case "salvos", "steamMarket", "csTradeBot":
			if q.Name == nil || strings.TrimSpace(*q.Name) == "" {
				return changes, fmt.Errorf("change %d: name is required", n+1)
			}
		case "cashConverters":
			if q.RequiredPhrases == nil || strings.TrimSpace(*q.RequiredPhrases) == "" {
				return changes, fmt.Errorf("change %d: requiredPhrases is required", n+1)
			}
			if q.ScanMode != nil && *q.ScanMode != "siteWide" {
				return changes, fmt.Errorf("change %d: only siteWide Cash Converters imports are supported", n+1)
			}
		default:
			return changes, fmt.Errorf("change %d: unsupported type %q", n+1, q.Type)
		}
		if q.Type != "cashConverters" && q.MaxPrice == nil {
			return changes, fmt.Errorf("change %d: maxPrice is required", n+1)
		}
		if q.MaxPrice != nil && *q.MaxPrice < 0 || q.MinPrice != nil && *q.MinPrice < 0 || q.MaxPrice != nil && q.MinPrice != nil && *q.MinPrice > *q.MaxPrice {
			return changes, fmt.Errorf("change %d: invalid price range", n+1)
		}
		if q.Type == "salvos" && q.MinPrice == nil || q.Type == "csMarket" && q.MaxFloat == nil || q.Type == "csTradeBot" && (q.MinFloat == nil || q.MaxFloat == nil) {
			return changes, fmt.Errorf("change %d: missing required site field", n+1)
		}
		if q.MinFloat != nil && (*q.MinFloat < 0 || *q.MinFloat > 1) || q.MaxFloat != nil && (*q.MaxFloat < 0 || *q.MaxFloat > 1) || q.MinFloat != nil && q.MaxFloat != nil && *q.MinFloat > *q.MaxFloat {
			return changes, fmt.Errorf("change %d: invalid float range", n+1)
		}
		for _, mode := range []*string{q.RequiredMatchMode, q.ExcludeMatchMode} {
			if mode != nil && *mode != "all" && *mode != "any" && *mode != "all_words" && *mode != "any_words" && *mode != "all_substring" && *mode != "any_substring" {
				return changes, fmt.Errorf("change %d: invalid match mode", n+1)
			}
		}
	}
	return changes, nil
}
func agentFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
func agentString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
func (b *Bot) checkAgentChanges(ctx context.Context, changes agentChanges) ([]query.QueryItem, error) {
	current, err := b.dbClient.ListSavedQueries(ctx, "")
	if err != nil {
		return nil, err
	}
	existing := make(map[string]query.QueryItem, len(current))
	for _, q := range current {
		existing[q.QueryId] = q
	}
	targets := make([]query.QueryItem, len(changes.Changes))
	for n, c := range changes.Changes {
		if c.Operation != "update" {
			continue
		}
		q, ok := existing[c.QueryID]
		if !ok {
			return nil, fmt.Errorf("change %d: query not found", n+1)
		}
		rev, err := queryRevision(q)
		if err != nil {
			return nil, err
		}
		if rev != c.Revision || q.Type != c.Query.Type {
			return nil, fmt.Errorf("change %d: query changed since export or type differs", n+1)
		}
		targets[n] = q
	}
	return targets, nil
}
func (b *Bot) applyAgentChange(ctx context.Context, c agentChange, old query.QueryItem) error {
	q := c.Query
	update := c.Operation == "update"
	var err error
	switch q.Type {
	case "cashConverters":
		v := query.CashConvertersQueryInput{Url: agentString(q.Url), ScanMode: agentString(q.ScanMode), RequiredPhrases: agentString(q.RequiredPhrases), RequiredMatchMode: agentString(q.RequiredMatchMode), ExcludePhrases: agentString(q.ExcludePhrases), ExcludeMatchMode: agentString(q.ExcludeMatchMode), MinPrice: q.MinPrice, MaxPrice: q.MaxPrice}
		if update {
			_, err = b.dbClient.UpdateCashConvertersQuery(ctx, old.Id, q.DmOnly, v)
		} else {
			_, err = b.dbClient.CreateCashConvertersQuery(ctx, q.DmOnly, v)
		}
	case "ebay":
		v := models.Ebay{Url: agentString(q.Url), MaxPrice: agentFloat(q.MaxPrice)}
		if update {
			_, err = b.dbClient.UpdateEbayQuery(ctx, old.Id, q.DmOnly, &v)
		} else {
			_, err = b.dbClient.CreateEbayQuery(ctx, q.DmOnly, &v)
		}
	case "gumtree":
		v := models.Gumtree{Url: agentString(q.Url), MaxPrice: agentFloat(q.MaxPrice)}
		if update {
			_, err = b.dbClient.UpdateGumtreeQuery(ctx, old.Id, q.DmOnly, &v)
		} else {
			_, err = b.dbClient.CreateGumtreeQuery(ctx, q.DmOnly, &v)
		}
	case "salvos":
		v := models.Salvos{Name: agentString(q.Name), MinPrice: agentFloat(q.MinPrice), MaxPrice: agentFloat(q.MaxPrice)}
		if update {
			_, err = b.dbClient.UpdateSalvosQuery(ctx, old.Id, q.DmOnly, &v)
		} else {
			_, err = b.dbClient.CreateSalvosQuery(ctx, q.DmOnly, &v)
		}
	case "csMarket":
		v := models.CsMarket{Url: agentString(q.Url), DisplayUrl: agentString(q.DisplayUrl), MaxPrice: agentFloat(q.MaxPrice), MaxFloat: agentFloat(q.MaxFloat)}
		if update {
			_, err = b.dbClient.UpdateCsMarketQuery(ctx, old.Id, q.DmOnly, &v)
		} else {
			_, err = b.dbClient.CreateCsMarketQuery(ctx, q.DmOnly, &v)
		}
	case "steamMarket":
		v := models.SteamMarket{Name: agentString(q.Name), DisplayUrl: agentString(q.DisplayUrl), MaxPrice: agentFloat(q.MaxPrice)}
		if update {
			_, err = b.dbClient.UpdateSteamMarketQuery(ctx, old.Id, q.DmOnly, &v)
		} else {
			_, err = b.dbClient.CreateSteamMarketQuery(ctx, q.DmOnly, &v)
		}
	case "csTradeBot":
		v := models.CsTradeBot{Name: agentString(q.Name), MaxPrice: agentFloat(q.MaxPrice), MinFloat: agentFloat(q.MinFloat), MaxFloat: agentFloat(q.MaxFloat)}
		if update {
			_, err = b.dbClient.UpdateCsTradeBotQuery(ctx, old.Id, q.DmOnly, &v)
		} else {
			_, err = b.dbClient.CreateCsTradeBotQuery(ctx, q.DmOnly, &v)
		}
	}
	return err
}

type pendingAgentImport struct {
	UserID  string
	Changes agentChanges
	Expires time.Time
}

func (b *Bot) handleAgentImport(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, name string, opts map[string]*discordgo.ApplicationCommandInteractionDataOption) {
	userID := ""
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
	} else if i.User != nil {
		userID = i.User.ID
	}
	reply := func(message string) {
		_, _ = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &message})
	}
	if userID == "" {
		reply("Could not identify the user.")
		return
	}
	if name == "agentimport" {
		opt := opts["json"]
		if opt == nil {
			reply("Paste a JSON changeset.")
			return
		}
		changes, err := parseAgentChanges(opt.StringValue())
		if err != nil {
			reply(fmt.Sprintf("Validation failed: %v", err))
			return
		}
		old, err := b.checkAgentChanges(ctx, changes)
		if err != nil {
			reply(fmt.Sprintf("Validation failed: %v", err))
			return
		}
		token := uuid.NewString()
		b.agentMu.Lock()
		if b.agentPending == nil {
			b.agentPending = make(map[string]pendingAgentImport)
		}
		for key, p := range b.agentPending {
			if time.Now().After(p.Expires) {
				delete(b.agentPending, key)
			}
		}
		b.agentPending[token] = pendingAgentImport{UserID: userID, Changes: changes, Expires: time.Now().Add(15 * time.Minute)}
		b.agentMu.Unlock()
		lines := []string{}
		for n, c := range changes.Changes {
			before := ""
			if c.Operation == "update" {
				before = fmt.Sprintf(" (current max: %s)", formatAgentPrice(old[n].MaxPrice))
			}
			lines = append(lines, fmt.Sprintf("%d. %s %s%s", n+1, c.Operation, c.Query.Type, before))
		}
		report, _ := json.MarshalIndent(struct {
			Proposed agentChanges      `json:"proposed"`
			Current  []query.QueryItem `json:"currentBeforeChanges"`
		}{Proposed: changes, Current: old}, "", "  ")
		message := fmt.Sprintf("Preview: %d changes. Review the attached JSON. Use `/agentapply token:%s` within 15 minutes. Changes apply one at a time; a failure stops the batch.\n%s", len(changes.Changes), token, strings.Join(lines, "\n"))
		if len(message) > 1800 {
			message = fmt.Sprintf("Preview: %d changes. Review the attached JSON. Use `/agentapply token:%s` within 15 minutes. Changes apply one at a time; a failure stops the batch.", len(changes.Changes), token)
		}
		sendAgentFile(s, i, message, "proposed-changes.json", report)
		return
	}
	token := getStringOption(opts, "token")
	b.agentMu.Lock()
	p, ok := b.agentPending[token]
	if ok && p.UserID == userID {
		delete(b.agentPending, token)
	}
	b.agentMu.Unlock()
	if !ok || p.UserID != userID || time.Now().After(p.Expires) {
		reply("Preview expired or not found; run /agentimport again.")
		return
	}
	old, err := b.checkAgentChanges(ctx, p.Changes)
	if err != nil {
		reply(fmt.Sprintf("Changes not applied: %v", err))
		return
	}
	for n, c := range p.Changes.Changes {
		if err := b.applyAgentChange(ctx, c, old[n]); err != nil {
			reply(fmt.Sprintf("Applied %d of %d changes; change %d failed: %v. Review current queries before retrying.", n, len(p.Changes.Changes), n+1, err))
			return
		}
	}
	reply(fmt.Sprintf("Applied %d changes. Export queries to inspect the result.", len(p.Changes.Changes)))
}

func formatAgentPrice(v *float64) string {
	if v == nil {
		return "unset"
	}
	return fmt.Sprintf("%.2f", *v)
}

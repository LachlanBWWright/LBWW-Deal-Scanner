package discord

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dealscanner/internal/config"
	"dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"dealscanner/internal/runtime"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
}

func downloadFile(url string) (*discordgo.File, error) {
	resp, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: status code %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Extract filename from URL
	name := "image.jpg"
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastName := parts[len(parts)-1]
		if idx := strings.Index(lastName, "?"); idx != -1 {
			lastName = lastName[:idx]
		}
		if lastName != "" {
			name = lastName
		}
	}

	return &discordgo.File{
		Name:        name,
		ContentType: resp.Header.Get("Content-Type"),
		Reader:      bytes.NewReader(data),
	}, nil
}

type Bot struct {
	cfg          *config.Config
	dbClient     *query.Query
	stateManager *runtime.StateManager
	session      *discordgo.Session
}

func NewBot(cfg *config.Config, dbClient *query.Query, stateManager *runtime.StateManager) *Bot {
	return &Bot{
		cfg:          cfg,
		dbClient:     dbClient,
		stateManager: stateManager,
	}
}

func (b *Bot) Start(ctx context.Context) error {
	if b.cfg.DiscordToken == "" || b.cfg.BotClientId == "" || b.cfg.DiscordGuildId == "" {
		b.stateManager.SetBotStatus(runtime.BotStatusDisabled, nil)
		log.Println("Missing Discord configuration, skipping Discord bot startup.")
		return nil
	}

	session, err := discordgo.New("Bot " + b.cfg.DiscordToken)
	if err != nil {
		errStr := err.Error()
		b.stateManager.SetBotStatus(runtime.BotStatusError, &errStr)
		return fmt.Errorf("failed to create discord session: %w", err)
	}

	b.session = session
	b.session.AddHandler(b.handleInteraction)

	err = b.session.Open()
	if err != nil {
		errStr := err.Error()
		b.stateManager.SetBotStatus(runtime.BotStatusError, &errStr)
		return fmt.Errorf("failed to open discord connection: %w", err)
	}

	b.stateManager.SetBotStatus(runtime.BotStatusReady, nil)
	log.Println("Discord bot connected and running.")

	b.SetStatus("Starting up...")
	go b.registerCommands()

	return nil
}

func (b *Bot) Close() {
	if b.session != nil {
		b.session.Close()
	}
}

func (b *Bot) registerCommands() {
	commands := []*discordgo.ApplicationCommand{
		// eBay
		{Name: "createebayquery", Description: "Creates a saved query for eBay", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "The URL of the query.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter the maximum price (in AUD).", Required: true},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editedbayquery", Description: "Edit an existing eBay saved query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL/ID of the saved query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New query string", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter the new maximum price (in AUD).", Required: false},
		}},
		{Name: "deleteebayquery", Description: "Delete an existing eBay saved query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL/ID of the saved query to delete", Required: true},
		}},
		{Name: "viewebayqueries", Description: "List all saved eBay queries"},

		{Name: "createcashquery", Description: "Creates a saved query for Cash Converters", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "The URL of the query.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter the maximum price (in AUD).", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "requiredphrases", Description: "Comma-separated required phrases.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excludephrases", Description: "Comma-separated excluded phrases.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "requiredindescription", Description: "Required phrases in description.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excludeindescription", Description: "Excluded phrases in description.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "scanmode", Description: "How to scan this query", Required: false, Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Search URL", Value: "searchUrl"},
				{Name: "Site wide", Value: "siteWide"},
			}},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editcashquery", Description: "Edit an existing Cash Converters saved query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "ID of the saved query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New URL or query string to replace the old one", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Optional maximum total price for notifications", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "requiredphrases", Description: "Optional required phrases", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excludephrases", Description: "Optional excluded phrases", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "requiredindescription", Description: "Optional required phrases in description", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excludeindescription", Description: "Optional excluded phrases in description", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "scanmode", Description: "How to scan this query", Required: false, Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Search URL", Value: "searchUrl"},
				{Name: "Site wide", Value: "siteWide"},
			}},
		}},
		{Name: "deletecashquery", Description: "Delete an existing Cash Converters query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL/ID of the query to delete", Required: true},
		}},
		{Name: "viewcashqueries", Description: "List all saved Cash Converters queries"},

		// Gumtree
		{Name: "creategumtreequery", Description: "Creates a saved query for Gumtree", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "The URL of the query.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter the maximum price (in AUD).", Required: true},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editgumtreequery", Description: "Edit an existing Gumtree query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL/ID of the query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New query string", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter the new maximum price (in AUD).", Required: false},
		}},
		{Name: "deletegumtreequery", Description: "Delete an existing Gumtree query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL/ID of the query to delete", Required: true},
		}},
		{Name: "viewgumtreequeries", Description: "List all saved Gumtree queries"},

		// Salvos
		{Name: "createsalvosquery", Description: "Creates a saved query for Salvos", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "Item search term name", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "minprice", Description: "Minimum price limit.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Maximum price limit.", Required: true},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editsalvosquery", Description: "Edit an existing Salvos query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "Name of the query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New query string", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "minprice", Description: "New minimum price.", Required: false},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "New maximum price.", Required: false},
		}},
		{Name: "deletesalvosquery", Description: "Delete an existing Salvos query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "Name of the query to delete", Required: true},
		}},
		{Name: "viewsalvosqueries", Description: "List all saved Salvos queries"},

		// CS Market
		{Name: "createcsmarket", Description: "Creates a saved query for CS Market", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "CS item listing URL.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter maximum price.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxfloat", Description: "Enter maximum float value.", Required: true},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editcsmarket", Description: "Edit an existing CS Market query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL of the query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New query string", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "New maximum price.", Required: false},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxfloat", Description: "New maximum float value.", Required: false},
		}},
		{Name: "deletecsmarket", Description: "Delete an existing CS Market query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL of the query to delete", Required: true},
		}},
		{Name: "viewcsmarketqueries", Description: "List all CS Market queries"},

		// CS Trade Bot / MultiSearch
		{Name: "createmultisearch", Description: "Creates a saved CS Trade Bot query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "skinname", Description: "Exact item name.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "minfloat", Description: "Minimum float limit.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxfloat", Description: "Maximum float limit.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Maximum price limit.", Required: false},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editmultisearchquery", Description: "Edit an existing CS Trade Bot query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "Name of the query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New query string", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "minfloat", Description: "New minimum float.", Required: false},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxfloat", Description: "New maximum float.", Required: false},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "New maximum price.", Required: false},
		}},
		{Name: "deletemultisearchquery", Description: "Delete an existing CS Trade Bot query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "Name of the query to delete", Required: true},
		}},
		{Name: "viewmultisearchqueries", Description: "List all CS Trade Bot queries"},

		// Steam SCM
		{Name: "createscmquery", Description: "Creates a saved Steam SCM query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "SCM search render URL.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter maximum price limit.", Required: true},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editscmquery", Description: "Edit an existing Steam SCM query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL/Name of the SCM query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New query string", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "New maximum price.", Required: false},
		}},
		{Name: "deletescmquery", Description: "Delete an existing Steam SCM query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "URL/Name of SCM query to delete", Required: true},
		}},
		{Name: "viewscmqueries", Description: "List all Steam SCM queries"},
	}

	cmdNames := make([]string, 0, len(commands))
	for _, cmd := range commands {
		cmdNames = append(cmdNames, cmd.Name)
		_, err := b.session.ApplicationCommandCreate(b.cfg.BotClientId, b.cfg.DiscordGuildId, cmd)
		if err != nil {
			log.Printf("Failed to register command %s: %v", cmd.Name, err)
		}
	}
	b.stateManager.SetCommandNames(cmdNames)
	log.Printf("Registered %d Discord command definitions.", len(commands))
}

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleButton(s, i)
	}
}

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	ctx := context.Background()

	// 1. Verify Command Permission Role ID
	globals, err := b.dbClient.GetGlobals(ctx)
	if err == nil && globals != nil && globals.CommandPermissionRoleId != "" {
		hasRole := false
		if i.Member != nil {
			for _, roleID := range i.Member.Roles {
				if roleID == globals.CommandPermissionRoleId {
					hasRole = true
					break
				}
			}
		}
		if !hasRole {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ You do not have permission to execute command mutations.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
	}

	// Defer reply
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	var responseContent string

	options := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range data.Options {
		options[opt.Name] = opt
	}

	switch data.Name {
	// eBay
	case "createebayquery":
		query := getStringOption(options, "query")
		maxPrice := getFloatOption(options, "maxprice")
		dmOnly := getBoolOption(options, "dmonly")
		_, err := b.dbClient.CreateEbayQuery(ctx, dmOnly, &models.Ebay{Url: query, MaxPrice: maxPrice})
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created eBay query for `%s` (Max Price: $%.2f)", query, maxPrice)
		}
	case "editedbayquery":
		id := getStringOption(options, "id")
		eb, err := b.dbClient.Ebay.WithContext(ctx).Where(b.dbClient.Ebay.Url.Eq(id)).First()
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok && opt != nil {
				urlVal = opt.StringValue()
			}
			maxPrice := eb.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				maxPrice = opt.FloatValue()
			}
			_, err := b.dbClient.UpdateEbayQuery(ctx, id, false, &models.Ebay{Url: urlVal, MaxPrice: maxPrice})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated eBay query `%s` (New Max Price: $%.2f)", id, maxPrice)
			}
		}
	case "deleteebayquery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "ebay", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted eBay query: `%s`", id)
		}
	case "viewebayqueries":
		b.sendPaginatedQueries(ctx, s, i, "ebay", "Saved eBay Queries", 1)

	// Cash Converters
	case "createcashquery":
		query := getStringOption(options, "query")
		var maxPrice *float64
		if opt, ok := options["maxprice"]; ok && opt != nil {
			val := opt.FloatValue()
			maxPrice = &val
		}
		var requiredPhrases string
		if opt, ok := options["requiredphrases"]; ok && opt != nil {
			requiredPhrases = opt.StringValue()
		}
		var excludePhrases string
		if opt, ok := options["excludephrases"]; ok && opt != nil {
			excludePhrases = opt.StringValue()
		}
		var requiredInDesc string
		if opt, ok := options["requiredindescription"]; ok && opt != nil {
			requiredInDesc = opt.StringValue()
		}
		var excludeInDesc string
		if opt, ok := options["excludeindescription"]; ok && opt != nil {
			excludeInDesc = opt.StringValue()
		}
		var scanMode string
		if opt, ok := options["scanmode"]; ok && opt != nil {
			scanMode = opt.StringValue()
		}
		dmOnly := getBoolOption(options, "dmonly")
		_, err := b.dbClient.CreateCashConvertersQuery(ctx, dmOnly, &models.CashConverters{
			Url:                   query,
			MaxPrice:              maxPrice,
			RequiredPhrases:       requiredPhrases,
			ExcludePhrases:        excludePhrases,
			RequiredInDescription: requiredInDesc,
			ExcludeInDescription:  excludeInDesc,
			ScanMode:              scanMode,
		})
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Cash Converters query for `%s`", query)
		}
	case "editcashquery":
		id := getStringOption(options, "id")
		cc, err := b.dbClient.CashConverters.WithContext(ctx).Where(b.dbClient.CashConverters.Url.Eq(id)).First()
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok && opt != nil {
				urlVal = opt.StringValue()
			}
			maxPrice := cc.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				val := opt.FloatValue()
				maxPrice = &val
			}
			scanMode := cc.ScanMode
			if opt, ok := options["scanmode"]; ok && opt != nil {
				scanMode = opt.StringValue()
			}
			requiredPhrases := cc.RequiredPhrases
			if opt, ok := options["requiredphrases"]; ok && opt != nil {
				requiredPhrases = opt.StringValue()
			}
			excludePhrases := cc.ExcludePhrases
			if opt, ok := options["excludephrases"]; ok && opt != nil {
				excludePhrases = opt.StringValue()
			}
			requiredInDesc := cc.RequiredInDescription
			if opt, ok := options["requiredindescription"]; ok && opt != nil {
				requiredInDesc = opt.StringValue()
			}
			excludeInDesc := cc.ExcludeInDescription
			if opt, ok := options["excludeindescription"]; ok && opt != nil {
				excludeInDesc = opt.StringValue()
			}

			_, err = b.dbClient.UpdateCashConvertersQuery(ctx, id, false, &models.CashConverters{
				Url:                   urlVal,
				MaxPrice:              maxPrice,
				ScanMode:              scanMode,
				RequiredPhrases:       requiredPhrases,
				ExcludePhrases:        excludePhrases,
				RequiredInDescription: requiredInDesc,
				ExcludeInDescription:  excludeInDesc,
			})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Cash Converters query `%s`", id)
			}
		}
	case "deletecashquery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "cashConverters", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Cash Converters query: `%s`", id)
		}
	case "viewcashqueries":
		b.sendPaginatedQueries(ctx, s, i, "cashConverters", "Saved Cash Converters Queries", 1)

	// Gumtree
	case "creategumtreequery":
		query := getStringOption(options, "query")
		maxPrice := getFloatOption(options, "maxprice")
		dmOnly := getBoolOption(options, "dmonly")
		_, err := b.dbClient.CreateGumtreeQuery(ctx, dmOnly, &models.Gumtree{Url: query, MaxPrice: maxPrice})
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Gumtree query for `%s` (Max Price: $%.2f)", query, maxPrice)
		}
	case "editgumtreequery":
		id := getStringOption(options, "id")
		gt, err := b.dbClient.Gumtree.WithContext(ctx).Where(b.dbClient.Gumtree.Url.Eq(id)).First()
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok && opt != nil {
				urlVal = opt.StringValue()
			}
			maxPrice := gt.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				maxPrice = opt.FloatValue()
			}
			_, err := b.dbClient.UpdateGumtreeQuery(ctx, id, false, &models.Gumtree{Url: urlVal, MaxPrice: maxPrice})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Gumtree query `%s` (New Max Price: $%.2f)", id, maxPrice)
			}
		}
	case "deletegumtreequery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "gumtree", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Gumtree query: `%s`", id)
		}
	case "viewgumtreequeries":
		b.sendPaginatedQueries(ctx, s, i, "gumtree", "Saved Gumtree Queries", 1)

	// Salvos
	case "createsalvosquery":
		name := getStringOption(options, "name")
		minPrice := getFloatOption(options, "minprice")
		maxPrice := getFloatOption(options, "maxprice")
		dmOnly := getBoolOption(options, "dmonly")
		_, err := b.dbClient.CreateSalvosQuery(ctx, dmOnly, &models.Salvos{Name: name, MinPrice: minPrice, MaxPrice: maxPrice})
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Salvos query for `%s` (Price Range: $%.2f - $%.2f)", name, minPrice, maxPrice)
		}
	case "editsalvosquery":
		id := getStringOption(options, "id")
		sa, err := b.dbClient.Salvos.WithContext(ctx).Where(b.dbClient.Salvos.Name.Eq(id)).First()
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			nameVal := id
			if opt, ok := options["query"]; ok && opt != nil {
				nameVal = opt.StringValue()
			}
			minPrice := sa.MinPrice
			if opt, ok := options["minprice"]; ok && opt != nil {
				minPrice = opt.FloatValue()
			}
			maxPrice := sa.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				maxPrice = opt.FloatValue()
			}
			_, err = b.dbClient.UpdateSalvosQuery(ctx, id, false, &models.Salvos{Name: nameVal, MinPrice: minPrice, MaxPrice: maxPrice})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Salvos query `%s` (New range: $%.2f - $%.2f)", id, minPrice, maxPrice)
			}
		}
	case "deletesalvosquery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "salvos", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Salvos query: `%s`", id)
		}
	case "viewsalvosqueries":
		b.sendPaginatedQueries(ctx, s, i, "salvos", "Saved Salvos Queries", 1)

	case "createcsmarket":
		query := getStringOption(options, "query")
		maxPrice := getFloatOption(options, "maxprice")
		maxFloat := getFloatOption(options, "maxfloat")
		dmOnly := getBoolOption(options, "dmonly")
		_, err := b.dbClient.CreateCsMarketQuery(ctx, dmOnly, &models.CsMarket{Url: query, MaxPrice: maxPrice, MaxFloat: maxFloat})
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created CS Market query for `%s` (Max Price: $%.2f, Max Float: %.5f)", query, maxPrice, maxFloat)
		}
	case "editcsmarket":
		id := getStringOption(options, "id")
		cm, err := b.dbClient.CsMarket.WithContext(ctx).Where(b.dbClient.CsMarket.Url.Eq(id)).First()
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok && opt != nil {
				urlVal = opt.StringValue()
			}
			maxPrice := cm.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				maxPrice = opt.FloatValue()
			}
			maxFloat := cm.MaxFloat
			if opt, ok := options["maxfloat"]; ok && opt != nil {
				maxFloat = opt.FloatValue()
			}
			_, err := b.dbClient.UpdateCsMarketQuery(ctx, id, false, &models.CsMarket{Url: urlVal, MaxPrice: maxPrice, MaxFloat: maxFloat})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated CS Market query `%s` (Max Price: $%.2f, Max Float: %.5f)", id, maxPrice, maxFloat)
			}
		}
	case "deletecsmarket":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "csMarket", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted CS Market query: `%s`", id)
		}
	case "viewcsmarketqueries":
		b.sendPaginatedQueries(ctx, s, i, "csMarket", "Saved CS Market Queries", 1)

	// CS Trade Bot / MultiSearch
	case "createmultisearch":
		skinName := getStringOption(options, "skinname")
		minFloat := getFloatOption(options, "minfloat")
		maxFloat := getFloatOption(options, "maxfloat")
		maxPrice := -1.0
		if opt, ok := options["maxprice"]; ok && opt != nil {
			maxPrice = opt.FloatValue()
		}
		dmOnly := getBoolOption(options, "dmonly")

		if minFloat > maxFloat {
			responseContent = "❌ Error: The minimum float cannot be higher than the maximum float value."
		} else if maxFloat <= 0 || maxFloat >= 1 {
			responseContent = "❌ Error: The maximum float must be between 0 and 1."
		} else if minFloat < 0 || minFloat >= 1 {
			responseContent = "❌ Error: The minimum float must be between 0 and 1."
		} else if maxPrice <= 0 && maxPrice != -1.0 {
			responseContent = "❌ Error: The price must be positive, and greater than $0."
		} else {
			_, err := b.dbClient.CreateCsTradeBotQuery(ctx, dmOnly, &models.CsTradeBot{Name: skinName, MaxPrice: maxPrice, MinFloat: minFloat, MaxFloat: maxFloat})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Created multisearch query for `%s` (Max Price: $%.2f, Float: %.5f - %.5f)", skinName, maxPrice, minFloat, maxFloat)
			}
		}
	case "editmultisearchquery":
		id := getStringOption(options, "id")
		ct, err := b.dbClient.CsTradeBot.WithContext(ctx).Where(b.dbClient.CsTradeBot.Name.Eq(id)).First()
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			nameVal := id
			if opt, ok := options["query"]; ok && opt != nil {
				nameVal = opt.StringValue()
			}
			maxPrice := ct.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				maxPrice = opt.FloatValue()
			}
			minFloat := ct.MinFloat
			if opt, ok := options["minfloat"]; ok && opt != nil {
				minFloat = opt.FloatValue()
			}
			maxFloat := ct.MaxFloat
			if opt, ok := options["maxfloat"]; ok && opt != nil {
				maxFloat = opt.FloatValue()
			}

			_, err := b.dbClient.UpdateCsTradeBotQuery(ctx, id, false, &models.CsTradeBot{Name: nameVal, MaxPrice: maxPrice, MinFloat: minFloat, MaxFloat: maxFloat})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated multisearch query `%s` (Max Price: $%.2f, Float: %.5f - %.5f)", id, maxPrice, minFloat, maxFloat)
			}
		}
	case "deletemultisearchquery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "csTradeBot", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted multisearch query: `%s`", id)
		}
	case "viewmultisearchqueries":
		b.sendPaginatedQueries(ctx, s, i, "csTradeBot", "Saved CS Trade Bot Multisearches", 1)

	// Steam SCM
	case "createscmquery":
		name := getStringOption(options, "name")
		maxPrice := getFloatOption(options, "maxprice")
		dmOnly := getBoolOption(options, "dmonly")
		_, err := b.dbClient.CreateSteamMarketQuery(ctx, dmOnly, &models.SteamMarket{Name: name, MaxPrice: maxPrice})
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Steam SCM query for `%s` (Max Price: $%.2f)", name, maxPrice)
		}
	case "editscmquery":
		id := getStringOption(options, "id")
		sm, err := b.dbClient.SteamMarket.WithContext(ctx).Where(b.dbClient.SteamMarket.Name.Eq(id)).First()
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			nameVal := id
			if opt, ok := options["query"]; ok && opt != nil {
				nameVal = opt.StringValue()
			}
			maxPrice := sm.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				maxPrice = opt.FloatValue()
			}
			_, err := b.dbClient.UpdateSteamMarketQuery(ctx, id, false, &models.SteamMarket{Name: nameVal, MaxPrice: maxPrice})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Steam SCM query `%s` (New Max Price: $%.2f)", id, maxPrice)
			}
		}
	case "deletescmquery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "steamMarket", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Steam SCM query: `%s`", id)
		}
	case "viewscmqueries":
		b.sendPaginatedQueries(ctx, s, i, "steamMarket", "Saved Steam SCM Queries", 1)

	default:
		responseContent = fmt.Sprintf("Command %s not supported.", data.Name)
	}

	if responseContent != "" {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &responseContent,
		})
	}
}

func (b *Bot) handleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	customId := i.MessageComponentData().CustomID

	action, err := b.dbClient.GetAction(ctx, customId)
	if err != nil || action == nil {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Action expired or not found.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	if action.Type == "view_page" {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredMessageUpdate,
		})
	} else {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Flags: discordgo.MessageFlagsEphemeral,
			},
		})
	}

	var responseContent string
	var responseComponents []discordgo.MessageComponent

	switch action.Type {
	case "delete":
		confirmKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		cancelKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		now := time.Now().UnixMilli()

		b.dbClient.SetAction(ctx, confirmKey, models.ActionRegistry{
			ID:         confirmKey,
			Type:       "confirm_delete",
			QueryType:  action.QueryType,
			QueryId:    action.QueryId,
			UserId:     &i.Member.User.ID,
			Timestamp:  now,
			RelatedKey: &customId,
		})

		b.dbClient.SetAction(ctx, cancelKey, models.ActionRegistry{
			ID:         cancelKey,
			Type:       "cancel_delete",
			QueryType:  action.QueryType,
			QueryId:    action.QueryId,
			UserId:     &i.Member.User.ID,
			Timestamp:  now,
			RelatedKey: &customId,
		})

		responseContent = fmt.Sprintf("Are you sure you want to delete this %s query?\n`%s`\n\n**This action cannot be undone.**", *action.QueryType, *action.QueryId)
		responseComponents = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Yes, Delete Query",
						Style:    discordgo.DangerButton,
						CustomID: confirmKey,
					},
					discordgo.Button{
						Label:    "Cancel",
						Style:    discordgo.SecondaryButton,
						CustomID: cancelKey,
					},
				},
			},
		}

	case "confirm_delete":
		if action.QueryType != nil && action.QueryId != nil {
			err = b.dbClient.DeleteSavedQuery(ctx, *action.QueryType, *action.QueryId)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error deleting query: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Successfully deleted query of type `%s` (`%s`)", *action.QueryType, *action.QueryId)
				b.dbClient.DeleteAction(ctx, customId)
				if action.RelatedKey != nil {
					b.dbClient.DeleteAction(ctx, *action.RelatedKey)
				}
			}
		}

	case "cancel_delete":
		responseContent = "Deletion cancelled."
		b.dbClient.DeleteAction(ctx, customId)
		if action.RelatedKey != nil {
			b.dbClient.DeleteAction(ctx, *action.RelatedKey)
		}

	case "subscribe_dm":
		var memberID string
		if i.Member != nil {
			memberID = i.Member.User.ID
		} else if i.User != nil {
			memberID = i.User.ID
		}

		parentQueryId, err := b.getQueryIdByTypeAndKey(ctx, *action.QueryType, *action.QueryId)
		if err != nil || parentQueryId == "" {
			responseContent = "❌ Query not found."
		} else {
			count, err := b.dbClient.UserQuery.WithContext(ctx).Where(b.dbClient.UserQuery.UserId.Eq(memberID), b.dbClient.UserQuery.QueryId.Eq(parentQueryId)).Count()
			if err == nil && count > 0 {
				responseContent = "✅ You are already subscribed to DMs for this query."
			} else {
				uq := models.UserQuery{
					ID:        uuid.New().String(),
					UserId:    memberID,
					QueryId:   parentQueryId,
					QueryType: *action.QueryType,
					CreatedAt: time.Now().UTC(),
				}
				if err := b.dbClient.UserQuery.WithContext(ctx).Create(&uq); err != nil {
					responseContent = fmt.Sprintf("❌ Failed to subscribe: %v", err)
				} else {
					responseContent = fmt.Sprintf("✅ You have been subscribed to DM notifications for this %s query.", *action.QueryType)
					b.dbClient.DeleteAction(ctx, customId)
				}
			}
		}

	case "unsubscribe_dm":
		var memberID string
		if i.Member != nil {
			memberID = i.Member.User.ID
		} else if i.User != nil {
			memberID = i.User.ID
		}

		parentQueryId, err := b.getQueryIdByTypeAndKey(ctx, *action.QueryType, *action.QueryId)
		if err != nil || parentQueryId == "" {
			responseContent = "❌ Query not found."
		} else {
			res, err := b.dbClient.UserQuery.WithContext(ctx).Where(b.dbClient.UserQuery.UserId.Eq(memberID), b.dbClient.UserQuery.QueryId.Eq(parentQueryId)).Delete()
			if err != nil {
				responseContent = fmt.Sprintf("❌ Failed to unsubscribe: %v", err)
			} else if res.RowsAffected == 0 {
				responseContent = "❌ You are not subscribed to this query."
			} else {
				responseContent = "✅ You have been unsubscribed from DM notifications for this query."
				b.dbClient.DeleteAction(ctx, customId)
			}
		}

	case "view_page":
		if action.QueryType != nil && action.QueryId != nil {
			pageVal, err := strconv.Atoi(*action.QueryId)
			if err == nil {
				title := ""
				switch *action.QueryType {
				case "ebay":
					title = "Saved eBay Queries"
				case "cashConverters":
					title = "Saved Cash Converters Queries"
				case "gumtree":
					title = "Saved Gumtree Queries"
				case "salvos":
					title = "Saved Salvos Queries"
				case "csMarket":
					title = "Saved CS Market Queries"
				case "csTradeBot":
					title = "Saved CS Trade Bot Multisearches"
				case "steamMarket":
					title = "Saved Steam SCM Queries"
				}
				b.sendPaginatedQueries(ctx, s, i, *action.QueryType, title, pageVal)
				b.dbClient.DeleteAction(ctx, customId)
				return
			}
		}
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &responseContent,
		Components: &responseComponents,
	})
}

// SendDeal publishes a deal notification to the configured channel and subscribed users
func (b *Bot) SendDeal(ctx context.Context, channelId, message string, imageUrl *string, queryType, queryId string) error {
	if b.session == nil {
		return fmt.Errorf("bot session is not initialized")
	}

	// 1. Fetch parent queryId and send DMs to subscribed users
	parentQueryId, err := b.getQueryIdByTypeAndKey(ctx, queryType, queryId)
	if err == nil && parentQueryId != "" {
		userQueries, err := b.dbClient.UserQuery.WithContext(ctx).Where(b.dbClient.UserQuery.QueryId.Eq(parentQueryId)).Find()
		if err == nil {
			for _, uq := range userQueries {
				dmChan, err := b.session.UserChannelCreate(uq.UserId)
				if err == nil && dmChan != nil {
					unsubscribeActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
					b.dbClient.SetAction(ctx, unsubscribeActionKey, models.ActionRegistry{
						ID:        unsubscribeActionKey,
						Type:      "unsubscribe_dm",
						QueryType: &queryType,
						QueryId:   &queryId,
						Timestamp: time.Now().UnixMilli(),
					})

					dmMsg := &discordgo.MessageSend{
						Content: message,
						Components: []discordgo.MessageComponent{
							discordgo.ActionsRow{
								Components: []discordgo.MessageComponent{
									discordgo.Button{
										Label:    "Unsubscribe from DM",
										Style:    discordgo.SecondaryButton,
										CustomID: unsubscribeActionKey,
									},
								},
							},
						},
					}
					if imageUrl != nil && *imageUrl != "" {
						if file, err := downloadFile(*imageUrl); err == nil {
							dmMsg.Files = []*discordgo.File{file}
						} else {
							log.Printf("Failed to download image %s for DM: %v", *imageUrl, err)
						}
					}
					b.session.ChannelMessageSendComplex(dmChan.ID, dmMsg)
				}
			}
		}

		// If query is DM-only, do not notify public channel
		parentQuery, err := b.dbClient.SearchQuery.WithContext(ctx).Where(b.dbClient.SearchQuery.ID.Eq(parentQueryId)).First()
		if err == nil && parentQuery.DmOnly {
			return nil
		}
	}

	// 2. Setup public channel message components
	var buttons []discordgo.MessageComponent

	if queryType != "" && queryId != "" {
		deleteActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		b.dbClient.SetAction(ctx, deleteActionKey, models.ActionRegistry{
			ID:        deleteActionKey,
			Type:      "delete",
			QueryType: &queryType,
			QueryId:   &queryId,
			Timestamp: time.Now().UnixMilli(),
		})

		subscribeDMActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		b.dbClient.SetAction(ctx, subscribeDMActionKey, models.ActionRegistry{
			ID:        subscribeDMActionKey,
			Type:      "subscribe_dm",
			QueryType: &queryType,
			QueryId:   &queryId,
			Timestamp: time.Now().UnixMilli(),
		})

		buttons = append(buttons, discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Delete Query",
					Style:    discordgo.DangerButton,
					CustomID: deleteActionKey,
				},
				discordgo.Button{
					Label:    "Subscribe to DM",
					Style:    discordgo.PrimaryButton,
					CustomID: subscribeDMActionKey,
				},
			},
		})
	}

	msgData := &discordgo.MessageSend{
		Content:    message,
		Components: buttons,
	}

	if imageUrl != nil && *imageUrl != "" {
		if file, err := downloadFile(*imageUrl); err == nil {
			msgData.Files = []*discordgo.File{file}
		} else {
			log.Printf("Failed to download image %s for public channel: %v", *imageUrl, err)
		}
	}

	_, err = b.session.ChannelMessageSendComplex(channelId, msgData)
	return err
}

// SendError publishes an error message to the error channel
func (b *Bot) SendError(channelId, message string) error {
	if b.session == nil {
		return fmt.Errorf("bot session is not initialized")
	}
	_, err := b.session.ChannelMessageSend(channelId, message)
	return err
}

func (b *Bot) getQueryIdByTypeAndKey(ctx context.Context, qType, key string) (string, error) {
	var qId string
	var err error
	switch qType {
	case "salvos":
		var sa *models.Salvos
		sa, err = b.dbClient.Salvos.WithContext(ctx).Where(b.dbClient.Salvos.Name.Eq(key)).First()
		if err == nil {
			qId = sa.QueryId
		}
	case "ebay":
		var eb *models.Ebay
		eb, err = b.dbClient.Ebay.WithContext(ctx).Where(b.dbClient.Ebay.Url.Eq(key)).First()
		if err == nil {
			qId = eb.QueryId
		}
	case "gumtree":
		var gt *models.Gumtree
		gt, err = b.dbClient.Gumtree.WithContext(ctx).Where(b.dbClient.Gumtree.Url.Eq(key)).First()
		if err == nil {
			qId = gt.QueryId
		}
	case "cashConverters":
		var cc *models.CashConverters
		cc, err = b.dbClient.CashConverters.WithContext(ctx).Where(b.dbClient.CashConverters.Url.Eq(key)).First()
		if err == nil {
			qId = cc.QueryId
		}
	case "steamMarket":
		var sm *models.SteamMarket
		sm, err = b.dbClient.SteamMarket.WithContext(ctx).Where(b.dbClient.SteamMarket.Name.Eq(key)).First()
		if err == nil {
			qId = sm.QueryId
		}
	case "csTradeBot":
		var ct *models.CsTradeBot
		ct, err = b.dbClient.CsTradeBot.WithContext(ctx).Where(b.dbClient.CsTradeBot.Name.Eq(key)).First()
		if err == nil {
			qId = ct.QueryId
		}
	case "csMarket":
		var cm *models.CsMarket
		cm, err = b.dbClient.CsMarket.WithContext(ctx).Where(b.dbClient.CsMarket.Url.Eq(key)).First()
		if err == nil {
			qId = cm.QueryId
		}
	default:
		return "", fmt.Errorf("unknown query type: %s", qType)
	}

	if err != nil {
		return "", err
	}
	return qId, nil
}

func (b *Bot) SetStatus(statusText string) {
	if b.session == nil {
		return
	}
	usd := discordgo.UpdateStatusData{
		Status: "online",
		Activities: []*discordgo.Activity{
			{
				Name:  "Custom Status",
				Type:  discordgo.ActivityTypeCustom,
				State: statusText,
			},
		},
	}
	b.session.UpdateStatusComplex(usd)
}

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
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s**\n", title))

	queryIds := make([]string, 0, len(queries))
	for _, q := range queries {
		queryIds = append(queryIds, q.QueryId)
	}

	foundMap, err := b.dbClient.GetLastFoundItemsBatch(ctx, queryIds, 3)
	if err != nil {
		log.Printf("Failed to batch get last found items: %v", err)
		foundMap = make(map[string][]query.FoundItem)
	}

	for _, q := range queries {
		var info string
		switch q.Type {
		case "ebay", "gumtree", "cashConverters":
			priceStr := "Any"
			if q.MaxPrice != nil {
				priceStr = fmt.Sprintf("$%.2f", *q.MaxPrice)
			}
			info = fmt.Sprintf("- URL: <%s> | Max Price: %s", q.Id, priceStr)
		case "salvos":
			info = fmt.Sprintf("- Name: `%s` | Price Range: $%.2f - $%.2f", q.Id, *q.MinPrice, *q.MaxPrice)
		case "csMarket":
			info = fmt.Sprintf("- URL: <%s> | Max Price: $%.2f | Max Float: %.5f", q.Id, *q.MaxPrice, *q.MaxFloat)
		case "csTradeBot":
			info = fmt.Sprintf("- Name: `%s` | Max Price: $%.2f | Float Range: %.5f - %.5f", q.Id, *q.MaxPrice, *q.MinFloat, *q.MaxFloat)
		case "steamMarket":
			urlPart := ""
			if q.DisplayUrl != nil && *q.DisplayUrl != "" {
				urlPart = fmt.Sprintf(" | [Market Link](<%s>)", *q.DisplayUrl)
			}
			info = fmt.Sprintf("- Name: `%s` | Max Price: $%.2f%s", q.Id, *q.MaxPrice, urlPart)
		}

		sb.WriteString(info)
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
	}
	return sb.String()
}

func (b *Bot) sendPaginatedQueries(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, queryType string, title string, page int) {
	queries, err := b.dbClient.ListSavedQueries(ctx, queryType)
	if err != nil {
		content := fmt.Sprintf("❌ Error: %v", err)
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
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

	pageSize := 10
	totalPages := (total + pageSize - 1) / pageSize
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	offset := (page - 1) * pageSize
	end := offset + pageSize
	if end > total {
		end = total
	}

	slicedQueries := queries[offset:end]

	formattedTitle := fmt.Sprintf("%s (Page %d/%d)", title, page, totalPages)
	content := b.formatQueriesWithFoundItems(ctx, slicedQueries, formattedTitle)

	var components []discordgo.MessageComponent
	if totalPages > 1 {
		firstKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		backKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		forwardKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		lastKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]

		now := time.Now().UnixMilli()

		b.dbClient.SetAction(ctx, firstKey, models.ActionRegistry{
			ID:        firstKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr("1"),
			Timestamp: now,
		})

		backPage := page - 1
		if backPage < 1 {
			backPage = 1
		}
		b.dbClient.SetAction(ctx, backKey, models.ActionRegistry{
			ID:        backKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr(strconv.Itoa(backPage)),
			Timestamp: now,
		})

		forwardPage := page + 1
		if forwardPage > totalPages {
			forwardPage = totalPages
		}
		b.dbClient.SetAction(ctx, forwardKey, models.ActionRegistry{
			ID:        forwardKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr(strconv.Itoa(forwardPage)),
			Timestamp: now,
		})

		b.dbClient.SetAction(ctx, lastKey, models.ActionRegistry{
			ID:        lastKey,
			Type:      "view_page",
			QueryType: &queryType,
			QueryId:   stringPtr(strconv.Itoa(totalPages)),
			Timestamp: now,
		})

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

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &content,
		Components: &components,
	})
}

func stringPtr(s string) *string {
	return &s
}

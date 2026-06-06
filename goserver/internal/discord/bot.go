package discord

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"dealscanner/internal/config"
	"dealscanner/internal/db"
	"dealscanner/internal/runtime"
	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

type Bot struct {
	cfg          *config.Config
	dbClient     *db.DB
	stateManager *runtime.StateManager
	session      *discordgo.Session
}

func NewBot(cfg *config.Config, dbClient *db.DB, stateManager *runtime.StateManager) *Bot {
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

	b.registerCommands()

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

		// Cash Converters
		{Name: "createcashquery", Description: "Creates a saved query for Cash Converters", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "The URL of the query.", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter the maximum price (in AUD).", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "requiredphrases", Description: "Comma-separated required phrases.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excludephrases", Description: "Comma-separated excluded phrases.", Required: false},
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
		query := options["query"].StringValue()
		maxPrice := options["maxprice"].FloatValue()
		dmOnly := false
		if opt, ok := options["dmonly"]; ok {
			dmOnly = opt.BoolValue()
		}
		payload := map[string]interface{}{"url": query, "maxPrice": maxPrice}
		_, err := b.dbClient.CreateSavedQuery(ctx, "ebay", dmOnly, payload)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created eBay query for `%s` (Max Price: $%.2f)", query, maxPrice)
		}
	case "editedbayquery":
		id := options["id"].StringValue()
		var eb db.Ebay
		if err := b.dbClient.Where("url = ?", id).First(&eb).Error; err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok {
				urlVal = opt.StringValue()
			}
			maxPrice := eb.MaxPrice
			if opt, ok := options["maxprice"]; ok {
				maxPrice = opt.FloatValue()
			}
			payload := map[string]interface{}{"url": urlVal, "maxPrice": maxPrice}
			_, err := b.dbClient.UpdateSavedQuery(ctx, "ebay", id, false, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated eBay query `%s` (New Max Price: $%.2f)", id, maxPrice)
			}
		}
	case "deleteebayquery":
		id := options["id"].StringValue()
		err := b.dbClient.DeleteSavedQuery(ctx, "ebay", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted eBay query: `%s`", id)
		}
	case "viewebayqueries":
		queries, err := b.dbClient.ListSavedQueries(ctx, "ebay")
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else if len(queries) == 0 {
			responseContent = "No saved eBay queries found."
		} else {
			var sb strings.Builder
			sb.WriteString("**Saved eBay Queries:**\n")
			for _, q := range queries {
				priceStr := "Any"
				if q.MaxPrice != nil {
					priceStr = fmt.Sprintf("$%.2f", *q.MaxPrice)
				}
				sb.WriteString(fmt.Sprintf("- ID: `%s` | Max Price: %s\n", *q.Url, priceStr))
			}
			responseContent = sb.String()
		}

	// Cash Converters
	case "createcashquery":
		query := options["query"].StringValue()
		payload := map[string]interface{}{"url": query}
		if opt, ok := options["maxprice"]; ok {
			payload["maxPrice"] = opt.FloatValue()
		}
		if opt, ok := options["requiredphrases"]; ok {
			payload["requiredPhrases"] = opt.StringValue()
		}
		if opt, ok := options["excludephrases"]; ok {
			payload["excludePhrases"] = opt.StringValue()
		}
		if opt, ok := options["scanmode"]; ok {
			payload["scanMode"] = opt.StringValue()
		}
		dmOnly := false
		if opt, ok := options["dmonly"]; ok {
			dmOnly = opt.BoolValue()
		}
		_, err := b.dbClient.CreateSavedQuery(ctx, "cashConverters", dmOnly, payload)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Cash Converters query for `%s`", query)
		}
	case "editcashquery":
		id := options["id"].StringValue()
		var cc db.CashConverters
		if err := b.dbClient.Where("url = ?", id).First(&cc).Error; err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok {
				urlVal = opt.StringValue()
			}
			maxPrice := cc.MaxPrice
			if opt, ok := options["maxprice"]; ok {
				val := opt.FloatValue()
				maxPrice = &val
			}
			scanMode := cc.ScanMode
			if opt, ok := options["scanmode"]; ok {
				scanMode = opt.StringValue()
			}
			requiredPhrases := cc.RequiredPhrases
			if opt, ok := options["requiredphrases"]; ok {
				requiredPhrases = opt.StringValue()
			}
			excludePhrases := cc.ExcludePhrases
			if opt, ok := options["excludephrases"]; ok {
				excludePhrases = opt.StringValue()
			}

			payload := map[string]interface{}{
				"url":             urlVal,
				"scanMode":        scanMode,
				"requiredPhrases": requiredPhrases,
				"excludePhrases":  excludePhrases,
			}
			if maxPrice != nil {
				payload["maxPrice"] = *maxPrice
			}
			_, err = b.dbClient.UpdateSavedQuery(ctx, "cashConverters", id, false, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Cash Converters query `%s`", id)
			}
		}
	case "deletecashquery":
		id := options["id"].StringValue()
		err := b.dbClient.DeleteSavedQuery(ctx, "cashConverters", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Cash Converters query: `%s`", id)
		}
	case "viewcashqueries":
		queries, err := b.dbClient.ListSavedQueries(ctx, "cashConverters")
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else if len(queries) == 0 {
			responseContent = "No saved Cash Converters queries found."
		} else {
			var sb strings.Builder
			sb.WriteString("**Saved Cash Converters Queries:**\n")
			for _, q := range queries {
				priceStr := "Any"
				if q.MaxPrice != nil {
					priceStr = fmt.Sprintf("$%.2f", *q.MaxPrice)
				}
				sb.WriteString(fmt.Sprintf("- ID: `%s` | Max Price: %s\n", *q.Url, priceStr))
			}
			responseContent = sb.String()
		}

	// Gumtree
	case "creategumtreequery":
		query := options["query"].StringValue()
		maxPrice := options["maxprice"].FloatValue()
		dmOnly := false
		if opt, ok := options["dmonly"]; ok {
			dmOnly = opt.BoolValue()
		}
		payload := map[string]interface{}{"url": query, "maxPrice": maxPrice}
		_, err := b.dbClient.CreateSavedQuery(ctx, "gumtree", dmOnly, payload)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Gumtree query for `%s` (Max Price: $%.2f)", query, maxPrice)
		}
	case "editgumtreequery":
		id := options["id"].StringValue()
		var gt db.Gumtree
		if err := b.dbClient.Where("url = ?", id).First(&gt).Error; err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok {
				urlVal = opt.StringValue()
			}
			maxPrice := gt.MaxPrice
			if opt, ok := options["maxprice"]; ok {
				maxPrice = opt.FloatValue()
			}
			payload := map[string]interface{}{"url": urlVal, "maxPrice": maxPrice}
			_, err := b.dbClient.UpdateSavedQuery(ctx, "gumtree", id, false, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Gumtree query `%s` (New Max Price: $%.2f)", id, maxPrice)
			}
		}
	case "deletegumtreequery":
		id := options["id"].StringValue()
		err := b.dbClient.DeleteSavedQuery(ctx, "gumtree", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Gumtree query: `%s`", id)
		}
	case "viewgumtreequeries":
		queries, err := b.dbClient.ListSavedQueries(ctx, "gumtree")
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else if len(queries) == 0 {
			responseContent = "No saved Gumtree queries found."
		} else {
			var sb strings.Builder
			sb.WriteString("**Saved Gumtree Queries:**\n")
			for _, q := range queries {
				priceStr := "Any"
				if q.MaxPrice != nil {
					priceStr = fmt.Sprintf("$%.2f", *q.MaxPrice)
				}
				sb.WriteString(fmt.Sprintf("- ID: `%s` | Max Price: %s\n", *q.Url, priceStr))
			}
			responseContent = sb.String()
		}

	// Salvos
	case "createsalvosquery":
		name := options["name"].StringValue()
		minPrice := options["minprice"].FloatValue()
		maxPrice := options["maxprice"].FloatValue()
		dmOnly := false
		if opt, ok := options["dmonly"]; ok {
			dmOnly = opt.BoolValue()
		}
		payload := map[string]interface{}{"name": name, "minPrice": minPrice, "maxPrice": maxPrice}
		_, err := b.dbClient.CreateSavedQuery(ctx, "salvos", dmOnly, payload)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Salvos query for `%s` (Price Range: $%.2f - $%.2f)", name, minPrice, maxPrice)
		}
	case "editsalvosquery":
		id := options["id"].StringValue()
		var sa db.Salvos
		if err := b.dbClient.Where("name = ?", id).First(&sa).Error; err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			nameVal := id
			if opt, ok := options["query"]; ok {
				nameVal = opt.StringValue()
			}
			minPrice := sa.MinPrice
			if opt, ok := options["minprice"]; ok {
				minPrice = opt.FloatValue()
			}
			maxPrice := sa.MaxPrice
			if opt, ok := options["maxprice"]; ok {
				maxPrice = opt.FloatValue()
			}
			payload := map[string]interface{}{"name": nameVal, "minPrice": minPrice, "maxPrice": maxPrice}
			_, err = b.dbClient.UpdateSavedQuery(ctx, "salvos", id, false, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Salvos query `%s` (New range: $%.2f - $%.2f)", id, minPrice, maxPrice)
			}
		}
	case "deletesalvosquery":
		id := options["id"].StringValue()
		err := b.dbClient.DeleteSavedQuery(ctx, "salvos", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Salvos query: `%s`", id)
		}
	case "viewsalvosqueries":
		queries, err := b.dbClient.ListSavedQueries(ctx, "salvos")
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else if len(queries) == 0 {
			responseContent = "No saved Salvos queries found."
		} else {
			var sb strings.Builder
			sb.WriteString("**Saved Salvos Queries:**\n")
			for _, q := range queries {
				sb.WriteString(fmt.Sprintf("- Name: `%s` | Price Range: $%.2f - $%.2f\n", *q.Name, *q.MinPrice, *q.MaxPrice))
			}
			responseContent = sb.String()
		}

	// CS Market
	case "createcsmarket":
		query := options["query"].StringValue()
		maxPrice := options["maxprice"].FloatValue()
		maxFloat := options["maxfloat"].FloatValue()
		dmOnly := false
		if opt, ok := options["dmonly"]; ok {
			dmOnly = opt.BoolValue()
		}
		payload := map[string]interface{}{"url": query, "maxPrice": maxPrice, "maxFloat": maxFloat}
		_, err := b.dbClient.CreateSavedQuery(ctx, "csMarket", dmOnly, payload)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created CS Market query for `%s` (Max Price: $%.2f, Max Float: %.5f)", query, maxPrice, maxFloat)
		}
	case "editcsmarket":
		id := options["id"].StringValue()
		var cm db.CsMarket
		if err := b.dbClient.Where("url = ?", id).First(&cm).Error; err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := id
			if opt, ok := options["query"]; ok {
				urlVal = opt.StringValue()
			}
			maxPrice := cm.MaxPrice
			if opt, ok := options["maxprice"]; ok {
				maxPrice = opt.FloatValue()
			}
			maxFloat := cm.MaxFloat
			if opt, ok := options["maxfloat"]; ok {
				maxFloat = opt.FloatValue()
			}
			payload := map[string]interface{}{"url": urlVal, "maxPrice": maxPrice, "maxFloat": maxFloat}
			_, err := b.dbClient.UpdateSavedQuery(ctx, "csMarket", id, false, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated CS Market query `%s` (Max Price: $%.2f, Max Float: %.5f)", id, maxPrice, maxFloat)
			}
		}
	case "deletecsmarket":
		id := options["id"].StringValue()
		err := b.dbClient.DeleteSavedQuery(ctx, "csMarket", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted CS Market query: `%s`", id)
		}
	case "viewcsmarketqueries":
		queries, err := b.dbClient.ListSavedQueries(ctx, "csMarket")
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else if len(queries) == 0 {
			responseContent = "No saved CS Market queries found."
		} else {
			var sb strings.Builder
			sb.WriteString("**Saved CS Market Queries:**\n")
			for _, q := range queries {
				sb.WriteString(fmt.Sprintf("- URL: `%s` | Max Price: $%.2f | Max Float: %.5f\n", *q.Url, *q.MaxPrice, *q.MaxFloat))
			}
			responseContent = sb.String()
		}

	// CS Trade Bot / MultiSearch
	case "createmultisearch":
		skinName := options["skinname"].StringValue()
		minFloat := options["minfloat"].FloatValue()
		maxFloat := options["maxfloat"].FloatValue()
		maxPrice := -1.0
		if opt, ok := options["maxprice"]; ok {
			maxPrice = opt.FloatValue()
		}
		dmOnly := false
		if opt, ok := options["dmonly"]; ok {
			dmOnly = opt.BoolValue()
		}

		if minFloat > maxFloat {
			responseContent = "❌ Error: The minimum float cannot be higher than the maximum float value."
		} else if maxFloat <= 0 || maxFloat >= 1 {
			responseContent = "❌ Error: The maximum float must be between 0 and 1."
		} else if minFloat < 0 || minFloat >= 1 {
			responseContent = "❌ Error: The minimum float must be between 0 and 1."
		} else if maxPrice <= 0 && maxPrice != -1.0 {
			responseContent = "❌ Error: The price must be positive, and greater than $0."
		} else {
			payload := map[string]interface{}{"name": skinName, "maxPrice": maxPrice, "minFloat": minFloat, "maxFloat": maxFloat}
			_, err := b.dbClient.CreateSavedQuery(ctx, "csTradeBot", dmOnly, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Created multisearch query for `%s` (Max Price: $%.2f, Float: %.5f - %.5f)", skinName, maxPrice, minFloat, maxFloat)
			}
		}
	case "editmultisearchquery":
		id := options["id"].StringValue()
		var ct db.CsTradeBot
		if err := b.dbClient.Where("name = ?", id).First(&ct).Error; err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			nameVal := id
			if opt, ok := options["query"]; ok {
				nameVal = opt.StringValue()
			}
			maxPrice := ct.MaxPrice
			if opt, ok := options["maxprice"]; ok {
				maxPrice = opt.FloatValue()
			}
			minFloat := ct.MinFloat
			if opt, ok := options["minfloat"]; ok {
				minFloat = opt.FloatValue()
			}
			maxFloat := ct.MaxFloat
			if opt, ok := options["maxfloat"]; ok {
				maxFloat = opt.FloatValue()
			}

			payload := map[string]interface{}{"name": nameVal, "maxPrice": maxPrice, "minFloat": minFloat, "maxFloat": maxFloat}
			_, err := b.dbClient.UpdateSavedQuery(ctx, "csTradeBot", id, false, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated multisearch query `%s` (Max Price: $%.2f, Float: %.5f - %.5f)", id, maxPrice, minFloat, maxFloat)
			}
		}
	case "deletemultisearchquery":
		id := options["id"].StringValue()
		err := b.dbClient.DeleteSavedQuery(ctx, "csTradeBot", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted multisearch query: `%s`", id)
		}
	case "viewmultisearchqueries":
		queries, err := b.dbClient.ListSavedQueries(ctx, "csTradeBot")
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else if len(queries) == 0 {
			responseContent = "No saved CS Trade Bot multisearch queries found."
		} else {
			var sb strings.Builder
			sb.WriteString("**Saved CS Trade Bot Multisearches:**\n")
			for _, q := range queries {
				sb.WriteString(fmt.Sprintf("- Name: `%s` | Max Price: $%.2f | Float Range: %.5f - %.5f\n", *q.Name, *q.MaxPrice, *q.MinFloat, *q.MaxFloat))
			}
			responseContent = sb.String()
		}

	// Steam SCM
	case "createscmquery":
		name := options["name"].StringValue()
		maxPrice := options["maxprice"].FloatValue()
		dmOnly := false
		if opt, ok := options["dmonly"]; ok {
			dmOnly = opt.BoolValue()
		}
		payload := map[string]interface{}{"name": name, "maxPrice": maxPrice}
		_, err := b.dbClient.CreateSavedQuery(ctx, "steamMarket", dmOnly, payload)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Steam SCM query for `%s` (Max Price: $%.2f)", name, maxPrice)
		}
	case "editscmquery":
		id := options["id"].StringValue()
		var sm db.SteamMarket
		if err := b.dbClient.Where("name = ?", id).First(&sm).Error; err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			nameVal := id
			if opt, ok := options["query"]; ok {
				nameVal = opt.StringValue()
			}
			maxPrice := sm.MaxPrice
			if opt, ok := options["maxprice"]; ok {
				maxPrice = opt.FloatValue()
			}
			payload := map[string]interface{}{"name": nameVal, "maxPrice": maxPrice}
			_, err := b.dbClient.UpdateSavedQuery(ctx, "steamMarket", id, false, payload)
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Steam SCM query `%s` (New Max Price: $%.2f)", id, maxPrice)
			}
		}
	case "deletescmquery":
		id := options["id"].StringValue()
		err := b.dbClient.DeleteSavedQuery(ctx, "steamMarket", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Steam SCM query: `%s`", id)
		}
	case "viewscmqueries":
		queries, err := b.dbClient.ListSavedQueries(ctx, "steamMarket")
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else if len(queries) == 0 {
			responseContent = "No saved Steam SCM queries found."
		} else {
			var sb strings.Builder
			sb.WriteString("**Saved Steam SCM Queries:**\n")
			for _, q := range queries {
				sb.WriteString(fmt.Sprintf("- Name: `%s` | Max Price: $%.2f\n", *q.Name, *q.MaxPrice))
			}
			responseContent = sb.String()
		}

	default:
		responseContent = fmt.Sprintf("Command %s not supported.", data.Name)
	}

	s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &responseContent,
	})
}

func (b *Bot) handleButton(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.Background()
	customId := i.MessageComponentData().CustomID

	// Defer reply ephemerally (visible only to user)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	action, err := b.dbClient.GetAction(ctx, customId)
	if err != nil || action == nil {
		content := "❌ Action expired or not found."
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
		return
	}

	var responseContent string
	var responseComponents []discordgo.MessageComponent

	switch action.Type {
	case "delete":
		confirmKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		cancelKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		now := time.Now().UnixMilli()

		b.dbClient.SetAction(ctx, confirmKey, db.ActionRegistry{
			ID:         confirmKey,
			Type:       "confirm_delete",
			QueryType:  action.QueryType,
			QueryId:    action.QueryId,
			UserId:     &i.Member.User.ID,
			Timestamp:  now,
			RelatedKey: &customId,
		})

		b.dbClient.SetAction(ctx, cancelKey, db.ActionRegistry{
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
			var count int64
			b.dbClient.Table("UserQuery").Where("userId = ? AND queryId = ?", memberID, parentQueryId).Count(&count)
			if count > 0 {
				responseContent = "✅ You are already subscribed to DMs for this query."
			} else {
				uq := db.UserQuery{
					ID:        uuid.New().String(),
					UserId:    memberID,
					QueryId:   parentQueryId,
					QueryType: *action.QueryType,
					CreatedAt: time.Now().UTC(),
				}
				if err := b.dbClient.Create(&uq).Error; err != nil {
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
			res := b.dbClient.Where("userId = ? AND queryId = ?", memberID, parentQueryId).Delete(&db.UserQuery{})
			if res.Error != nil {
				responseContent = fmt.Sprintf("❌ Failed to unsubscribe: %v", res.Error)
			} else if res.RowsAffected == 0 {
				responseContent = "❌ You are not subscribed to this query."
			} else {
				responseContent = "✅ You have been unsubscribed from DM notifications for this query."
				b.dbClient.DeleteAction(ctx, customId)
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
		var userQueries []db.UserQuery
		b.dbClient.Where("queryId = ?", parentQueryId).Find(&userQueries)

		for _, uq := range userQueries {
			dmChan, err := b.session.UserChannelCreate(uq.UserId)
			if err == nil {
				unsubscribeActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
				b.dbClient.SetAction(ctx, unsubscribeActionKey, db.ActionRegistry{
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
					dmMsg.Embeds = []*discordgo.MessageEmbed{
						{
							Image: &discordgo.MessageEmbedImage{
								URL: *imageUrl,
							},
						},
					}
				}
				b.session.ChannelMessageSendComplex(dmChan.ID, dmMsg)
			}
		}

		// If query is DM-only, do not notify public channel
		var parentQuery db.Query
		if err := b.dbClient.Where("id = ?", parentQueryId).First(&parentQuery).Error; err == nil && parentQuery.DmOnly {
			return nil
		}
	}

	// 2. Setup public channel message components
	var buttons []discordgo.MessageComponent

	if queryType != "" && queryId != "" {
		deleteActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		b.dbClient.SetAction(ctx, deleteActionKey, db.ActionRegistry{
			ID:        deleteActionKey,
			Type:      "delete",
			QueryType: &queryType,
			QueryId:   &queryId,
			Timestamp: time.Now().UnixMilli(),
		})

		subscribeDMActionKey := strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
		b.dbClient.SetAction(ctx, subscribeDMActionKey, db.ActionRegistry{
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
		msgData.Embeds = []*discordgo.MessageEmbed{
			{
				Image: &discordgo.MessageEmbedImage{
					URL: *imageUrl,
				},
			},
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
		var sa db.Salvos
		err = b.dbClient.Where("name = ?", key).First(&sa).Error
		qId = sa.QueryId
	case "ebay":
		var eb db.Ebay
		err = b.dbClient.Where("url = ?", key).First(&eb).Error
		qId = eb.QueryId
	case "gumtree":
		var gt db.Gumtree
		err = b.dbClient.Where("url = ?", key).First(&gt).Error
		qId = gt.QueryId
	case "cashConverters":
		var cc db.CashConverters
		err = b.dbClient.Where("url = ?", key).First(&cc).Error
		qId = cc.QueryId
	case "steamMarket":
		var sm db.SteamMarket
		err = b.dbClient.Where("name = ?", key).First(&sm).Error
		qId = sm.QueryId
	case "csTradeBot":
		var ct db.CsTradeBot
		err = b.dbClient.Where("name = ?", key).First(&ct).Error
		qId = ct.QueryId
	case "csMarket":
		var cm db.CsMarket
		err = b.dbClient.Where("url = ?", key).First(&cm).Error
		qId = cm.QueryId
	default:
		return "", fmt.Errorf("unknown query type: %s", qType)
	}

	if err != nil {
		return "", err
	}
	return qId, nil
}

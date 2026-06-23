package discord

import (
	"context"
	"fmt"
	"log"

	qry "dealscanner/internal/db/query"
	"dealscanner/internal/models"
	"github.com/bwmarrin/discordgo"
)

func valueOrString(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}

func matchModeChoices() []*discordgo.ApplicationCommandOptionChoice {
	return []*discordgo.ApplicationCommandOptionChoice{
		{Name: "All phrases", Value: "all"},
		{Name: "Any phrase", Value: "any"},
	}
}

func getOptionalString(options map[string]*discordgo.ApplicationCommandInteractionDataOption, name string, fallback string) string {
	if option, ok := options[name]; ok && option != nil {
		return option.StringValue()
	}
	return fallback
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
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "minprice", Description: "Enter the minimum price (in AUD).", Required: false},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Enter the maximum price (in AUD).", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "required", Description: "Comma-separated phrases required in the title or description.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "requiredmode", Description: "Require any or all phrases.", Required: false, Choices: matchModeChoices()},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excluded", Description: "Comma-separated phrases excluded from the title and description.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excludedmode", Description: "Exclude on any or all phrases.", Required: false, Choices: matchModeChoices()},
			{Type: discordgo.ApplicationCommandOptionString, Name: "scanmode", Description: "How to scan this query", Required: false, Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Search URL", Value: "searchUrl"},
				{Name: "Site wide", Value: "siteWide"},
			}},
			{Type: discordgo.ApplicationCommandOptionBoolean, Name: "dmonly", Description: "Make this query DM-only.", Required: false},
		}},
		{Name: "editcashquery", Description: "Edit an existing Cash Converters saved query", Options: []*discordgo.ApplicationCommandOption{
			{Type: discordgo.ApplicationCommandOptionString, Name: "id", Description: "ID of the saved query to edit", Required: true},
			{Type: discordgo.ApplicationCommandOptionString, Name: "query", Description: "New URL or query string to replace the old one", Required: true},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "minprice", Description: "Optional minimum total price for notifications", Required: false},
			{Type: discordgo.ApplicationCommandOptionNumber, Name: "maxprice", Description: "Optional maximum total price for notifications", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "required", Description: "Comma-separated phrases required in the title or description.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "requiredmode", Description: "Require any or all phrases.", Required: false, Choices: matchModeChoices()},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excluded", Description: "Comma-separated phrases excluded from the title and description.", Required: false},
			{Type: discordgo.ApplicationCommandOptionString, Name: "excludedmode", Description: "Exclude on any or all phrases.", Required: false, Choices: matchModeChoices()},
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
			responseContent = fmt.Sprintf("✅ Created eBay query for %s (Max Price: $%.2f)", formatQueryReference(query), maxPrice)
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
				responseContent = fmt.Sprintf("✅ Updated eBay query %s (New Max Price: $%.2f)", formatQueryReference(id), maxPrice)
			}
		}
	case "deleteebayquery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "ebay", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted eBay query: %s", formatQueryReference(id))
		}
	case "viewebayqueries":
		b.sendPaginatedQueries(ctx, s, i, "ebay", "Saved eBay Queries", 1)

	// Cash Converters
	case "createcashquery":
		query := getStringOption(options, "query")
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
		var required string
		if opt, ok := options["required"]; ok && opt != nil {
			required = opt.StringValue()
		}
		var excluded string
		if opt, ok := options["excluded"]; ok && opt != nil {
			excluded = opt.StringValue()
		}
		requiredMode := getOptionalString(options, "requiredmode", "all")
		excludedMode := getOptionalString(options, "excludedmode", "any")
		var scanMode string
		if opt, ok := options["scanmode"]; ok && opt != nil {
			scanMode = opt.StringValue()
		}
		dmOnly := getBoolOption(options, "dmonly")
		_, err := b.dbClient.CreateCashConvertersQuery(ctx, dmOnly, qry.CashConvertersQueryInput{
			Url: query, MinPrice: minPrice, MaxPrice: maxPrice,
			RequiredPhrases: required, RequiredMatchMode: requiredMode,
			ExcludePhrases: excluded, ExcludeMatchMode: excludedMode, ScanMode: scanMode,
		})
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Created Cash Converters query for %s", formatQueryReference(query))
		}
	case "editcashquery":
		id := getStringOption(options, "id")
		items, err := b.dbClient.ListSavedQueries(ctx, "cashConverters")
		cc := qry.QueryItem{}
		found := false
		for index := range items {
			if items[index].Id == id {
				cc = items[index]
				found = true
				break
			}
		}
		if err == nil && !found {
			err = fmt.Errorf("query not found")
		}
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: Query not found: %s", id)
		} else {
			urlVal := ""
			if cc.Url != nil {
				urlVal = *cc.Url
			}
			if opt, ok := options["query"]; ok && opt != nil {
				urlVal = opt.StringValue()
			}
			maxPrice := cc.MaxPrice
			if opt, ok := options["maxprice"]; ok && opt != nil {
				val := opt.FloatValue()
				maxPrice = &val
			}
			minPrice := cc.MinPrice
			if opt, ok := options["minprice"]; ok && opt != nil {
				val := opt.FloatValue()
				minPrice = &val
			}
			scanMode := valueOrString(cc.ScanMode, "searchUrl")
			if opt, ok := options["scanmode"]; ok && opt != nil {
				scanMode = opt.StringValue()
			}
			required := valueOrString(cc.RequiredPhrases, "")
			if opt, ok := options["required"]; ok && opt != nil {
				required = opt.StringValue()
			}
			excluded := valueOrString(cc.ExcludePhrases, "")
			if opt, ok := options["excluded"]; ok && opt != nil {
				excluded = opt.StringValue()
			}
			requiredMode := valueOrString(cc.RequiredMatchMode, "all")
			if opt, ok := options["requiredmode"]; ok && opt != nil {
				requiredMode = opt.StringValue()
			}
			excludedMode := valueOrString(cc.ExcludeMatchMode, "any")
			if opt, ok := options["excludedmode"]; ok && opt != nil {
				excludedMode = opt.StringValue()
			}

			_, err = b.dbClient.UpdateCashConvertersQuery(ctx, id, cc.DmOnly, qry.CashConvertersQueryInput{
				Url: urlVal, MinPrice: minPrice, MaxPrice: maxPrice, ScanMode: scanMode,
				RequiredPhrases: required, RequiredMatchMode: requiredMode,
				ExcludePhrases: excluded, ExcludeMatchMode: excludedMode,
			})
			if err != nil {
				responseContent = fmt.Sprintf("❌ Error: %v", err)
			} else {
				responseContent = fmt.Sprintf("✅ Updated Cash Converters query %s", formatQueryReference(id))
			}
		}
	case "deletecashquery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "cashConverters", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Cash Converters query: %s", formatQueryReference(id))
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
			responseContent = fmt.Sprintf("✅ Created Gumtree query for %s (Max Price: $%.2f)", formatQueryReference(query), maxPrice)
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
				responseContent = fmt.Sprintf("✅ Updated Gumtree query %s (New Max Price: $%.2f)", formatQueryReference(id), maxPrice)
			}
		}
	case "deletegumtreequery":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "gumtree", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted Gumtree query: %s", formatQueryReference(id))
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
			responseContent = fmt.Sprintf("✅ Created CS Market query for %s (Max Price: $%.2f, Max Float: %.5f)", formatQueryReference(query), maxPrice, maxFloat)
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
				responseContent = fmt.Sprintf("✅ Updated CS Market query %s (Max Price: $%.2f, Max Float: %.5f)", formatQueryReference(id), maxPrice, maxFloat)
			}
		}
	case "deletecsmarket":
		id := getStringOption(options, "id")
		err := b.dbClient.DeleteSavedQuery(ctx, "csMarket", id)
		if err != nil {
			responseContent = fmt.Sprintf("❌ Error: %v", err)
		} else {
			responseContent = fmt.Sprintf("✅ Deleted CS Market query: %s", formatQueryReference(id))
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

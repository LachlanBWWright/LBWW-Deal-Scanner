package discord

import (
	"context"
	"fmt"
	"log"

	"dealscanner/internal/config"
	"dealscanner/internal/db/query"
	"dealscanner/internal/runtime"
	"github.com/bwmarrin/discordgo"
)

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

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleButton(s, i)
	}
}

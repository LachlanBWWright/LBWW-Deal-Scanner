package discord

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"dealscanner/internal/config"
	"dealscanner/internal/db/query"
	"dealscanner/internal/runtime"
	"github.com/bwmarrin/discordgo"
)

const (
	discordGatewayAPIVersion = "10"
	discordInitialRetryLimit = 5
)

type Bot struct {
	cfg          *config.Config
	dbClient     *query.Query
	stateManager *runtime.StateManager
	session      *discordgo.Session
	agentMu      sync.Mutex
	agentPending map[string]pendingAgentImport
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

	// discordgo v0.29 defaults to Gateway API v9. Discord's current Gateway
	// version is v10; set this before Open builds the websocket URL.
	discordgo.APIVersion = discordGatewayAPIVersion

	session, err := discordgo.New("Bot " + b.cfg.DiscordToken)
	if err != nil {
		errStr := err.Error()
		b.stateManager.SetBotStatus(runtime.BotStatusError, &errStr)
		return fmt.Errorf("failed to create discord session: %w", err)
	}

	b.session = session
	b.session.ShouldReconnectOnError = true
	b.session.AddHandler(b.handleInteraction)
	b.session.AddHandler(b.handleDiscordConnect)
	b.session.AddHandler(b.handleDiscordDisconnect)

	err = b.openWithRetry(ctx)
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

func (b *Bot) openWithRetry(ctx context.Context) error {
	var err error
	wait := time.Second
	for attempt := 1; attempt <= discordInitialRetryLimit; attempt++ {
		err = b.session.Open()
		if err == nil {
			return nil
		}
		if attempt == discordInitialRetryLimit {
			break
		}
		log.Printf("Discord gateway connection attempt %d/%d failed: %v; retrying in %s", attempt, discordInitialRetryLimit, err, wait)
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
		wait *= 2
	}
	return err
}

func (b *Bot) handleDiscordConnect(_ *discordgo.Session, _ *discordgo.Connect) {
	b.stateManager.SetBotStatus(runtime.BotStatusReady, nil)
	log.Println("Discord bot gateway connected.")
}

func (b *Bot) handleDiscordDisconnect(_ *discordgo.Session, _ *discordgo.Disconnect) {
	errStr := "Discord gateway disconnected; reconnecting."
	b.stateManager.SetBotStatus(runtime.BotStatusError, &errStr)
	log.Println(errStr)
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

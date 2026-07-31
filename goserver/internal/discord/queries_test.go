package discord

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestGetBoolOptionDefault(t *testing.T) {
	options := map[string]*discordgo.ApplicationCommandInteractionDataOption{}
	if !getBoolOptionDefault(options, "ephemeral", true) {
		t.Fatal("missing ephemeral option did not use the true default")
	}

	options["ephemeral"] = &discordgo.ApplicationCommandInteractionDataOption{
		Name:  "ephemeral",
		Type:  discordgo.ApplicationCommandOptionBoolean,
		Value: false,
	}
	if getBoolOptionDefault(options, "ephemeral", true) {
		t.Fatal("explicit false ephemeral option was ignored")
	}
}

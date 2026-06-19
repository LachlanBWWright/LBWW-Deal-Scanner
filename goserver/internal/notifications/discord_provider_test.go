package notifications

import (
	"context"
	"strings"
	"testing"

	"dealscanner/internal/config"
)

type capturedDeal struct {
	channelMessage string
	dmMessage      string
}

type capturingBotSender struct {
	deal capturedDeal
}

func (s *capturingBotSender) SendDeal(
	_ context.Context,
	_ string,
	channelMessage string,
	dmMessage string,
	_ *string,
	_ string,
	_ string,
) error {
	s.deal = capturedDeal{
		channelMessage: channelMessage,
		dmMessage:      dmMessage,
	}
	return nil
}

func (s *capturingBotSender) SendError(_, _ string) error {
	return nil
}

func TestDiscordProviderRoleMentionIsOnlyIncludedInChannelMessage(t *testing.T) {
	const roleID = "123456789"
	sender := &capturingBotSender{}
	provider := NewDiscordProvider(&config.Config{
		EbayChannelId: "channel-id",
		EbayRoleId:    roleID,
	}, sender)

	err := provider.Send(context.Background(), AppNotification{
		Kind:   "deal",
		Source: "ebay",
		Title:  "Test deal",
	})
	if err != nil {
		t.Fatalf("Send returned an error: %v", err)
	}

	roleMention := "<@&" + roleID + "> "
	if !strings.HasPrefix(sender.deal.channelMessage, roleMention) {
		t.Fatalf("channel message %q does not start with role mention %q", sender.deal.channelMessage, roleMention)
	}
	if strings.Contains(sender.deal.dmMessage, roleMention) {
		t.Fatalf("DM message %q contains role mention %q", sender.deal.dmMessage, roleMention)
	}
	if strings.TrimPrefix(sender.deal.channelMessage, roleMention) != sender.deal.dmMessage {
		t.Fatalf("channel and DM message bodies differ: channel=%q DM=%q", sender.deal.channelMessage, sender.deal.dmMessage)
	}
}

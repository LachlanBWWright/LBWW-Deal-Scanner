package discord

import (
	"strings"
	"testing"

	"dealscanner/internal/config"
)

func TestBuildStartupMessageIncludesConfiguredChannelsAndRoles(t *testing.T) {
	cfg := &config.Config{
		ErrorChannelId:          "100",
		CashConvertersChannelId: "200",
		CashConvertersRoleId:    "300",
	}
	message := BuildStartupMessage(cfg)
	for _, expected := range []string{"<#100>", "<#200>", "<@&300>"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("startup message %q does not contain %q", message, expected)
		}
	}
}

func TestIsErrorLogLine(t *testing.T) {
	if !isErrorLogLine("scanner failed") {
		t.Fatal("expected failure log to be forwarded")
	}
	if isErrorLogLine("scan completed") {
		t.Fatal("expected informational log to be ignored")
	}
}

package discord

import (
	"fmt"
	"net/url"
	"strings"
)

func formatQueryReference(value string) string {
	cleaned := strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(value))
	parsed, err := url.Parse(cleaned)
	if err == nil && parsed.Host != "" && (parsed.Scheme == "http" || parsed.Scheme == "https") {
		return fmt.Sprintf("<%s>", cleaned)
	}

	return fmt.Sprintf("`%s`", strings.ReplaceAll(cleaned, "`", "ˋ"))
}

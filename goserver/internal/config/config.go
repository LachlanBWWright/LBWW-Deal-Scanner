package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ApiHost             string
	ApiPort             int
	ApiSecret           string
	DatabaseUrl         string
	TursoDatabaseUrl    string
	TursoAuthToken      string
	ScheduledMode       bool
	ScheduledDurationMs int
	EnableTestingApi    bool

	// Discord / Globals defaults
	BotClientId    string
	DiscordGuildId string
	DiscordToken   string

	// Feature flags / Channels
	CashConverters          bool
	CashConvertersChannelId string
	CashConvertersRoleId    string
	CommandPermissionRoleId string
	CsChannelId             string
	CsItems                 bool
	CsMarketChannelId       string
	CsMarketRoleId          string
	CsRoleId                string
	Ebay                    bool
	EbayChannelId           string
	EbayRoleId              string
	ErrorChannelId          string
	Gumtree                 bool
	GumtreeChannelId        string
	GumtreeRoleId           string
	Salvos                  bool
	SalvosChannelId         string
	SalvosRoleId            string
	SteamQuery              bool
	SteamQueryChannelId     string
	SteamQueryRoleId        string
}

// LoadEnv loads environment variables from a .env file relative to the working directory.
func LoadEnv(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		return // Ignore if file doesn't exist
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Strip quotes if any
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	val = strings.ToLower(strings.TrimSpace(val))
	return val == "true" || val == "1" || val == "yes" || val == "on"
}

func getEnvInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(val))
	if err != nil {
		return defaultValue
	}
	return parsed
}

func LoadConfig() *Config {
	// Load the single project-level environment file.
	LoadEnv(".env")

	return &Config{
		ApiHost:             getEnv("API_HOST", "0.0.0.0"),
		ApiPort:             getEnvInt("API_PORT", getEnvInt("PORT", 3000)),
		ApiSecret:           os.Getenv("API_SECRET"),
		DatabaseUrl:         os.Getenv("TURSO_DATABASE_URL"),
		TursoDatabaseUrl:    os.Getenv("TURSO_DATABASE_URL"),
		TursoAuthToken:      os.Getenv("TURSO_AUTH_TOKEN"),
		ScheduledMode:       getEnvBool("SCHEDULED_SCANNER_MODE", false),
		ScheduledDurationMs: getEnvInt("SCHEDULED_SCANNER_DURATION_MS", 300000),
		EnableTestingApi:    getEnvBool("ENABLE_TESTING_API", true),

		BotClientId:    os.Getenv("BOT_CLIENT_ID"),
		DiscordGuildId: os.Getenv("DISCORD_GUILD_ID"),
		DiscordToken:   os.Getenv("DISCORD_TOKEN"),

		CashConverters:          getEnvBool("CASH_CONVERTERS", false),
		CashConvertersChannelId: os.Getenv("CASH_CONVERTERS_CHANNEL_ID"),
		CashConvertersRoleId:    os.Getenv("CASH_CONVERTERS_ROLE_ID"),
		CommandPermissionRoleId: os.Getenv("COMMAND_PERMISSION_ROLE_ID"),
		CsChannelId:             os.Getenv("CS_CHANNEL_ID"),
		CsItems:                 getEnvBool("CS_ITEMS", false),
		CsMarketChannelId:       os.Getenv("CS_MARKET_CHANNEL_ID"),
		CsMarketRoleId:          os.Getenv("CS_MARKET_ROLE_ID"),
		CsRoleId:                os.Getenv("CS_ROLE_ID"),
		Ebay:                    getEnvBool("EBAY", false),
		EbayChannelId:           os.Getenv("EBAY_CHANNEL_ID"),
		EbayRoleId:              os.Getenv("EBAY_ROLE_ID"),
		ErrorChannelId:          os.Getenv("ERROR_CHANNEL_ID"),
		Gumtree:                 getEnvBool("GUMTREE", false),
		GumtreeChannelId:        os.Getenv("GUMTREE_CHANNEL_ID"),
		GumtreeRoleId:           os.Getenv("GUMTREE_ROLE_ID"),
		Salvos:                  getEnvBool("SALVOS", false),
		SalvosChannelId:         os.Getenv("SALVOS_CHANNEL_ID"),
		SalvosRoleId:            os.Getenv("SALVOS_ROLE_ID"),
		SteamQuery:              getEnvBool("STEAM_QUERY", false),
		SteamQueryChannelId:     os.Getenv("STEAM_QUERY_CHANNEL_ID"),
		SteamQueryRoleId:        os.Getenv("STEAM_QUERY_ROLE_ID"),
	}
}

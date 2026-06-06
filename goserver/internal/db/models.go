package db

import "time"

type Query struct {
	ID        string    `gorm:"primaryKey;column:id"`
	DmOnly    bool      `gorm:"column:dmOnly"`
	CreatedAt time.Time `gorm:"column:createdAt"`
}

func (Query) TableName() string {
	return "Query"
}

type UserQuery struct {
	ID        string    `gorm:"primaryKey;column:id"`
	UserId    string    `gorm:"column:userId"`
	QueryId   string    `gorm:"column:queryId"`
	QueryType string    `gorm:"column:queryType"`
	CreatedAt time.Time `gorm:"column:createdAt"`
}

func (UserQuery) TableName() string {
	return "UserQuery"
}

type CashConverters struct {
	Url             string   `gorm:"primaryKey;column:url"`
	RequiredPhrases string   `gorm:"column:requiredPhrases"`
	ExcludePhrases  string   `gorm:"column:excludePhrases"`
	MaxPrice        *float64 `gorm:"column:maxPrice"`
	ScanMode        string   `gorm:"column:scanMode"`
	QueryId         string   `gorm:"column:queryId"`
}

func (CashConverters) TableName() string {
	return "CashConverters"
}

type Ebay struct {
	Url      string  `gorm:"primaryKey;column:url"`
	MaxPrice float64 `gorm:"column:maxPrice"`
	QueryId  string  `gorm:"column:queryId"`
}

func (Ebay) TableName() string {
	return "Ebay"
}

type Gumtree struct {
	Url      string  `gorm:"primaryKey;column:url"`
	MaxPrice float64 `gorm:"column:maxPrice"`
	QueryId  string  `gorm:"column:queryId"`
}

func (Gumtree) TableName() string {
	return "Gumtree"
}

type Salvos struct {
	Name     string  `gorm:"primaryKey;column:name"`
	MinPrice float64 `gorm:"column:minPrice"`
	MaxPrice float64 `gorm:"column:maxPrice"`
	QueryId  string  `gorm:"column:queryId"`
}

func (Salvos) TableName() string {
	return "Salvos"
}

type CsMarket struct {
	Url        string  `gorm:"primaryKey;column:url"`
	DisplayUrl string  `gorm:"column:displayUrl"`
	MaxPrice   float64 `gorm:"column:maxPrice"`
	MaxFloat   float64 `gorm:"column:maxFloat"`
	LastPrice  float64 `gorm:"column:lastPrice"`
	QueryId    string  `gorm:"column:queryId"`
}

func (CsMarket) TableName() string {
	return "CsMarket"
}

type SteamMarket struct {
	Name       string  `gorm:"primaryKey;column:name"`
	DisplayUrl string  `gorm:"column:displayUrl"`
	MaxPrice   float64 `gorm:"column:maxPrice"`
	LastPrice  float64 `gorm:"column:lastPrice"`
	QueryId    string  `gorm:"column:queryId"`
}

func (SteamMarket) TableName() string {
	return "SteamMarket"
}

type CsTradeBot struct {
	Name     string  `gorm:"primaryKey;column:name"`
	MaxPrice float64 `gorm:"column:maxPrice"`
	MinFloat float64 `gorm:"column:minFloat"`
	MaxFloat float64 `gorm:"column:maxFloat"`
	QueryId  string  `gorm:"column:queryId"`
}

func (CsTradeBot) TableName() string {
	return "CsTradeBot"
}

type Globals struct {
	ID                      string  `gorm:"primaryKey;column:id"`
	BotClientId             string  `gorm:"column:BOT_CLIENT_ID"`
	CashConverters          bool    `gorm:"column:CASH_CONVERTERS"`
	CashConvertersChannelId string  `gorm:"column:CASH_CONVERTERS_CHANNEL_ID"`
	CashConvertersRoleId    string  `gorm:"column:CASH_CONVERTERS_ROLE_ID"`
	CommandPermissionRoleId string  `gorm:"column:COMMAND_PERMISSION_ROLE_ID"`
	CsChannelId             string  `gorm:"column:CS_CHANNEL_ID"`
	CsItems                 bool    `gorm:"column:CS_ITEMS"`
	CsMarketChannelId       string  `gorm:"column:CS_MARKET_CHANNEL_ID"`
	CsMarketRoleId          string  `gorm:"column:CS_MARKET_ROLE_ID"`
	CsRoleId                string  `gorm:"column:CS_ROLE_ID"`
	DiscordGuildId          string  `gorm:"column:DISCORD_GUILD_ID"`
	DiscordToken            string  `gorm:"column:DISCORD_TOKEN"`
	Ebay                    bool    `gorm:"column:EBAY"`
	EbayChannelId           string  `gorm:"column:EBAY_CHANNEL_ID"`
	EbayRoleId              string  `gorm:"column:EBAY_ROLE_ID"`
	ErrorChannelId          string  `gorm:"column:ERROR_CHANNEL_ID"`
	Gumtree                 bool    `gorm:"column:GUMTREE"`
	GumtreeChannelId        string  `gorm:"column:GUMTREE_CHANNEL_ID"`
	GumtreeRoleId           string  `gorm:"column:GUMTREE_ROLE_ID"`
	Salvos                  bool    `gorm:"column:SALVOS"`
	SalvosChannelId         string  `gorm:"column:SALVOS_CHANNEL_ID"`
	SalvosRoleId            string  `gorm:"column:SALVOS_ROLE_ID"`
	SteamQuery              bool    `gorm:"column:STEAM_QUERY"`
	SteamQueryChannelId     string  `gorm:"column:STEAM_QUERY_CHANNEL_ID"`
	SteamQueryRoleId        string  `gorm:"column:STEAM_QUERY_ROLE_ID"`
	CsTradeDollarRatio      float64 `gorm:"column:CS_TRADE_DOLLAR_RATIO"`
	LootFarmDollarRatio     float64 `gorm:"column:LOOT_FARM_DOLLAR_RATIO"`
	CsDealsDollarRatio      float64 `gorm:"column:CS_DEALS_DOLLAR_RATIO"`
	TradeitGgDollarRatio    float64 `gorm:"column:TRADEIT_GG_DOLLAR_RATIO"`
}

func (Globals) TableName() string {
	return "Globals"
}

type Listing struct {
	ID            string     `gorm:"primaryKey;column:id"`
	Source        string     `gorm:"column:source"`
	ExternalId    *string    `gorm:"column:externalId"`
	CanonicalUrl  string     `gorm:"column:canonicalUrl"`
	Title         string     `gorm:"column:title"`
	ImageUrl      *string    `gorm:"column:imageUrl"`
	Description   *string    `gorm:"column:description"`
	Availability  *string    `gorm:"column:availability"`
	FirstSeenAt   time.Time  `gorm:"column:firstSeenAt"`
	LastSeenAt    time.Time  `gorm:"column:lastSeenAt"`
	LastDetailAt  *time.Time `gorm:"column:lastDetailAt"`
	UnavailableAt *time.Time `gorm:"column:unavailableAt"`
}

func (Listing) TableName() string {
	return "Listing"
}

type ListingObservation struct {
	ID           string    `gorm:"primaryKey;column:id"`
	ListingId    string    `gorm:"column:listingId"`
	Source       string    `gorm:"column:source"`
	ObservedAt   time.Time `gorm:"column:observedAt"`
	Price        *float64  `gorm:"column:price"`
	Shipping     *float64  `gorm:"column:shipping"`
	TotalPrice   *float64  `gorm:"column:totalPrice"`
	Currency     *string   `gorm:"column:currency"`
	Title        string    `gorm:"column:title"`
	ImageUrl     *string   `gorm:"column:imageUrl"`
	Description  *string   `gorm:"column:description"`
	Availability *string   `gorm:"column:availability"`
}

func (ListingObservation) TableName() string {
	return "ListingObservation"
}

type QueryListingState struct {
	QueryId                string     `gorm:"primaryKey;column:queryId"`
	ListingId              string     `gorm:"primaryKey;column:listingId"`
	Source                 string     `gorm:"column:source"`
	Status                 string     `gorm:"column:status"`
	LastEvaluatedAt        time.Time  `gorm:"column:lastEvaluatedAt"`
	FirstMatchedAt         *time.Time `gorm:"column:firstMatchedAt"`
	LastMatchedAt          *time.Time `gorm:"column:lastMatchedAt"`
	LastRejectedReason     *string    `gorm:"column:lastRejectedReason"`
	LastNotifiedAt         *time.Time `gorm:"column:lastNotifiedAt"`
	LastNotifiedTotalPrice *float64   `gorm:"column:lastNotifiedTotalPrice"`
	LowestObservedPrice    *float64   `gorm:"column:lowestObservedPrice"`
}

func (QueryListingState) TableName() string {
	return "QueryListingState"
}

type ScannerRuntimeState struct {
	ID                      string     `gorm:"primaryKey;column:id"`
	SteamScanCount          int        `gorm:"column:steamScanCount"`
	CsTradeScanCount        int        `gorm:"column:csTradeScanCount"`
	ScheduledRuns           int        `gorm:"column:scheduledRuns"`
	TotalScheduledRuntimeMs int64      `gorm:"column:totalScheduledRuntimeMs"`
	LastStartedAt           *time.Time `gorm:"column:lastStartedAt"`
	LastStoppedAt           *time.Time `gorm:"column:lastStoppedAt"`
	UpdatedAt               time.Time  `gorm:"column:updatedAt"`
}

func (ScannerRuntimeState) TableName() string {
	return "ScannerRuntimeState"
}

type ActionRegistry struct {
	ID         string    `gorm:"primaryKey;column:id"`
	Type       string    `gorm:"column:type"`
	QueryType  *string   `gorm:"column:queryType"`
	QueryId    *string   `gorm:"column:queryId"`
	UserId     *string   `gorm:"column:userId"`
	Timestamp  int64     `gorm:"column:timestamp"`
	RelatedKey *string   `gorm:"column:relatedKey"`
	CreatedAt  time.Time `gorm:"column:createdAt"`
}

func (ActionRegistry) TableName() string {
	return "ActionRegistry"
}

package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type QueryListingStateStatus string

const (
	QueryListingStateStatusMatched     QueryListingStateStatus = "matched"
	QueryListingStateStatusNotified    QueryListingStateStatus = "notified"
	QueryListingStateStatusRejected    QueryListingStateStatus = "rejected"
	QueryListingStateStatusUnavailable QueryListingStateStatus = "unavailable"
	QueryListingStateStatusUnseen      QueryListingStateStatus = "unseen"
)

func (s QueryListingStateStatus) String() string {
	return string(s)
}

func (s QueryListingStateStatus) Valid() bool {
	switch s {
	case QueryListingStateStatusMatched,
		QueryListingStateStatusNotified,
		QueryListingStateStatusRejected,
		QueryListingStateStatusUnavailable,
		QueryListingStateStatusUnseen:
		return true
	default:
		return false
	}
}

func ParseQueryListingStateStatus(value string) (QueryListingStateStatus, error) {
	status := QueryListingStateStatus(value)
	if !status.Valid() {
		return "", fmt.Errorf("invalid query listing state status %q", value)
	}
	return status, nil
}

func (s QueryListingStateStatus) Value() (driver.Value, error) {
	if !s.Valid() {
		return nil, fmt.Errorf("invalid query listing state status %q", s)
	}
	return string(s), nil
}

func (s *QueryListingStateStatus) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		status, err := ParseQueryListingStateStatus(v)
		if err != nil {
			return err
		}
		*s = status
		return nil
	case []byte:
		status, err := ParseQueryListingStateStatus(string(v))
		if err != nil {
			return err
		}
		*s = status
		return nil
	default:
		return fmt.Errorf("unsupported query listing state status value %T", value)
	}
}

type SearchQuery struct {
	ID        string    `gorm:"primaryKey;column:id"`
	DmOnly    bool      `gorm:"column:dmOnly"`
	CreatedAt time.Time `gorm:"column:createdAt"`
}

func (SearchQuery) TableName() string {
	return "Query"
}

type UserQuery struct {
	ID        string      `gorm:"primaryKey;column:id"`
	UserId    string      `gorm:"column:userId"`
	QueryId   string      `gorm:"column:queryId"`
	Query     SearchQuery `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	QueryType string      `gorm:"column:queryType"`
	CreatedAt time.Time   `gorm:"column:createdAt"`
}

func (UserQuery) TableName() string {
	return "UserQuery"
}

type CashConverters struct {
	Url      string `gorm:"primaryKey;column:url" json:"url"`
	ScanMode string `gorm:"column:scanMode" json:"scanMode"`
}

func (CashConverters) TableName() string {
	return "CashConverters"
}

type CashConvertersFilter struct {
	ID                string         `gorm:"primaryKey;column:id" json:"id"`
	CashConvertersUrl string         `gorm:"column:cashConvertersUrl;index" json:"cashConvertersUrl"`
	CashConverters    CashConverters `gorm:"foreignKey:CashConvertersUrl;references:Url;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
	QueryId           string         `gorm:"column:queryId;uniqueIndex" json:"queryId"`
	Query             SearchQuery    `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"query,omitempty"`
	RequiredPhrases   string         `gorm:"column:requiredPhrases" json:"requiredPhrases"`
	RequiredMatchMode string         `gorm:"column:requiredMatchMode" json:"requiredMatchMode"`
	ExcludePhrases    string         `gorm:"column:excludePhrases" json:"excludePhrases"`
	ExcludeMatchMode  string         `gorm:"column:excludeMatchMode" json:"excludeMatchMode"`
	MinPrice          *float64       `gorm:"column:minPrice" json:"minPrice"`
	MaxPrice          *float64       `gorm:"column:maxPrice" json:"maxPrice"`
	CreatedAt         time.Time      `gorm:"column:createdAt" json:"createdAt"`
}

func (CashConvertersFilter) TableName() string {
	return "CashConvertersFilter"
}

type Ebay struct {
	Url      string      `gorm:"primaryKey;column:url" json:"url"`
	MaxPrice float64     `gorm:"column:maxPrice" json:"maxPrice"`
	QueryId  string      `gorm:"column:queryId" json:"queryId"`
	Query    SearchQuery `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"query,omitempty"`
}

func (Ebay) TableName() string {
	return "Ebay"
}

type Gumtree struct {
	Url      string      `gorm:"primaryKey;column:url" json:"url"`
	MaxPrice float64     `gorm:"column:maxPrice" json:"maxPrice"`
	QueryId  string      `gorm:"column:queryId" json:"queryId"`
	Query    SearchQuery `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"query,omitempty"`
}

func (Gumtree) TableName() string {
	return "Gumtree"
}

type Salvos struct {
	Name     string      `gorm:"primaryKey;column:name" json:"name"`
	MinPrice float64     `gorm:"column:minPrice" json:"minPrice"`
	MaxPrice float64     `gorm:"column:maxPrice" json:"maxPrice"`
	QueryId  string      `gorm:"column:queryId" json:"queryId"`
	Query    SearchQuery `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"query,omitempty"`
}

func (Salvos) TableName() string {
	return "Salvos"
}

type CsMarket struct {
	Url        string      `gorm:"primaryKey;column:url" json:"url"`
	DisplayUrl string      `gorm:"column:displayUrl" json:"displayUrl"`
	MaxPrice   float64     `gorm:"column:maxPrice" json:"maxPrice"`
	MaxFloat   float64     `gorm:"column:maxFloat" json:"maxFloat"`
	LastPrice  float64     `gorm:"column:lastPrice" json:"lastPrice"`
	QueryId    string      `gorm:"column:queryId" json:"queryId"`
	Query      SearchQuery `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"query,omitempty"`
}

func (CsMarket) TableName() string {
	return "CsMarket"
}

type SteamMarket struct {
	Name       string      `gorm:"primaryKey;column:name" json:"name"`
	DisplayUrl string      `gorm:"column:displayUrl" json:"displayUrl"`
	MaxPrice   float64     `gorm:"column:maxPrice" json:"maxPrice"`
	LastPrice  float64     `gorm:"column:lastPrice" json:"lastPrice"`
	QueryId    string      `gorm:"column:queryId" json:"queryId"`
	Query      SearchQuery `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"query,omitempty"`
}

func (SteamMarket) TableName() string {
	return "SteamMarket"
}

type CsTradeBot struct {
	Name     string      `gorm:"primaryKey;column:name" json:"name"`
	MaxPrice float64     `gorm:"column:maxPrice" json:"maxPrice"`
	MinFloat float64     `gorm:"column:minFloat" json:"minFloat"`
	MaxFloat float64     `gorm:"column:maxFloat" json:"maxFloat"`
	QueryId  string      `gorm:"column:queryId" json:"queryId"`
	Query    SearchQuery `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"query,omitempty"`
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
	QueryId                string                  `gorm:"primaryKey;column:queryId"`
	Query                  SearchQuery             `gorm:"foreignKey:QueryId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	ListingId              string                  `gorm:"primaryKey;column:listingId"`
	Listing                Listing                 `gorm:"foreignKey:ListingId;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Source                 string                  `gorm:"column:source"`
	Status                 QueryListingStateStatus `gorm:"column:status"`
	LastEvaluatedAt        time.Time               `gorm:"column:lastEvaluatedAt"`
	FirstMatchedAt         *time.Time              `gorm:"column:firstMatchedAt"`
	LastMatchedAt          *time.Time              `gorm:"column:lastMatchedAt"`
	LastRejectedReason     *string                 `gorm:"column:lastRejectedReason"`
	LastNotifiedAt         *time.Time              `gorm:"column:lastNotifiedAt"`
	LastNotifiedTotalPrice *float64                `gorm:"column:lastNotifiedTotalPrice"`
	LowestObservedPrice    *float64                `gorm:"column:lowestObservedPrice"`
}

func (QueryListingState) TableName() string {
	return "QueryListingState"
}

func (s *QueryListingState) BeforeSave(tx *gorm.DB) error {
	if !s.Status.Valid() {
		return fmt.Errorf("invalid query listing state status %q", s.Status)
	}
	return nil
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

type TtlItem struct {
	ItemId      string    `gorm:"primaryKey;column:itemId" json:"itemId"`
	Scanner     int       `gorm:"primaryKey;column:scanner" json:"scanner"`
	LastUpdated time.Time `gorm:"column:lastUpdated" json:"lastUpdated"`
}

func (TtlItem) TableName() string {
	return "TtlItem"
}

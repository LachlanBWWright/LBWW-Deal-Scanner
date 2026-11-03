// Package valuation provides evidence-based, advisory item valuations.
// Prices are gross amounts in currency minor units, not realizable net profit.
package valuation

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type Status string

const (
	Normal   Status = "normal"
	StatTrak Status = "stattrak"
	Souvenir Status = "souvenir"
)

type EvidenceKind string

const (
	Sale EvidenceKind = "sale"
	Ask  EvidenceKind = "ask"
)

type Balance string

const (
	Cash        Balance = "cash"
	SteamWallet Balance = "steam_wallet"
	SiteCredit  Balance = "site_credit"
)

type Wear string

const (
	FactoryNew    Wear = "factory_new"
	MinimalWear   Wear = "minimal_wear"
	FieldTested   Wear = "field_tested"
	WellWorn      Wear = "well_worn"
	BattleScarred Wear = "battle_scarred"
)

type Skin struct {
	Name     string  `json:"name"`
	MinFloat float64 `json:"min_float"`
	MaxFloat float64 `json:"max_float"`
}

// Observation must describe a plain skin: exclude patterned/stickered/charmed
// premiums before importing evidence into this float-only model.
type Observation struct {
	ID         string       `json:"id"`
	AssetID    string       `json:"asset_id"`
	Skin       string       `json:"skin"`
	Status     Status       `json:"status"`
	Float      *float64     `json:"float"`
	PriceMinor int64        `json:"price_minor"`
	Currency   string       `json:"currency"`
	Balance    Balance      `json:"balance"`
	Venue      string       `json:"venue"`
	Kind       EvidenceKind `json:"kind"`
	ObservedAt time.Time    `json:"observed_at"`
	URL        string       `json:"url,omitempty"`
}

type Config struct {
	MaxAgeHours    int     `json:"max_age_hours"`
	FloatRadius    float64 `json:"float_radius"`
	MinComparables int     `json:"min_comparables"`
	MaxComparables int     `json:"max_comparables"`
}

type Dataset struct {
	AsOf         time.Time     `json:"as_of"`
	Config       Config        `json:"config"`
	Skins        []Skin        `json:"skins"`
	Observations []Observation `json:"observations"`
	Candidates   []Observation `json:"candidates"`
}

type Error struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Field) }

// Summary reports dispersion in observed prices, not a confidence interval.
type Summary struct {
	MedianMinor  int64    `json:"median_minor"`
	MinimumMinor int64    `json:"minimum_minor"`
	MaximumMinor int64    `json:"maximum_minor"`
	EvidenceIDs  []string `json:"evidence_ids"`
}

type BetterTier struct {
	Wear               Wear    `json:"wear"`
	DistanceToBoundary float64 `json:"distance_to_boundary"`
	Prices             Summary `json:"prices"`
}

type Estimate struct {
	Basis                EvidenceKind `json:"basis"`
	ComparablePrices     Summary      `json:"comparable_prices"`
	CategoryBaseline     Summary      `json:"category_baseline"`
	FloatPremiumMinor    int64        `json:"float_premium_minor"`
	GrossDiscountMinor   int64        `json:"gross_discount_minor"`
	GrossDiscountPercent float64      `json:"gross_discount_percent"`
	BetterTier           *BetterTier  `json:"better_tier,omitempty"`
}

type Report struct {
	Candidate Observation `json:"candidate"`
	Wear      Wear        `json:"wear"`
	Model     string      `json:"model"`
	State     string      `json:"state"`
	Reason    string      `json:"reason,omitempty"`
	Estimate  *Estimate   `json:"estimate,omitempty"`
}

type band struct {
	wear      Wear
	low, high float64
}

var bands = []band{
	{FactoryNew, 0, .07}, {MinimalWear, .07, .15},
	{FieldTested, .15, .38}, {WellWorn, .38, .45}, {BattleScarred, .45, 1},
}

func wearBand(value float64) band {
	for _, b := range bands {
		if value < b.high {
			return b
		}
	}
	return bands[len(bands)-1]
}

func invalid(field, message string) error {
	return &Error{Code: "invalid_input", Field: field, Message: message}
}

func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }

func validate(d Dataset) (map[string]Skin, error) {
	c := d.Config
	if d.AsOf.IsZero() {
		return nil, invalid("as_of", "an explicit valuation timestamp is required")
	}
	if c.MaxAgeHours < 1 || c.MaxAgeHours > 24*365 {
		return nil, invalid("config.max_age_hours", "must be between 1 and 8760")
	}
	if !finite(c.FloatRadius) || c.FloatRadius <= 0 || c.FloatRadius > .1 {
		return nil, invalid("config.float_radius", "must be greater than zero and at most 0.1")
	}
	if c.MinComparables < 3 || c.MaxComparables < c.MinComparables || c.MaxComparables > 1000 {
		return nil, invalid("config", "require 3 <= min_comparables <= max_comparables <= 1000")
	}
	skins := make(map[string]Skin)
	for _, skin := range d.Skins {
		if strings.TrimSpace(skin.Name) == "" || !finite(skin.MinFloat) || !finite(skin.MaxFloat) ||
			skin.MinFloat < 0 || skin.MaxFloat > 1 || skin.MinFloat >= skin.MaxFloat {
			return nil, invalid("skins", "each skin requires a name and a valid attainable float range")
		}
		if _, exists := skins[skin.Name]; exists {
			return nil, invalid("skins", "duplicate skin name")
		}
		skins[skin.Name] = skin
	}
	ids := make(map[string]bool)
	validateObservation := func(o Observation) error {
		if o.ID == "" || o.AssetID == "" || strings.TrimSpace(o.Venue) == "" {
			return invalid("observation", "id, asset_id and venue are required")
		}
		if ids[o.ID] {
			return invalid(o.ID, "observation IDs must be unique, including candidates")
		}
		ids[o.ID] = true
		skin, exists := skins[o.Skin]
		if !exists {
			return invalid(o.ID, "unknown skin")
		}
		if o.Float == nil {
			return invalid(o.ID, "float is required; missing is not zero")
		}
		if !finite(*o.Float) || *o.Float < skin.MinFloat || *o.Float > skin.MaxFloat {
			return invalid(o.ID, "float is outside the skin's attainable range")
		}
		if o.Status != Normal && o.Status != StatTrak && o.Status != Souvenir {
			return invalid(o.ID, "unknown item status")
		}
		if o.Kind != Sale && o.Kind != Ask {
			return invalid(o.ID, "kind must be sale or ask")
		}
		if o.Balance != Cash && o.Balance != SteamWallet && o.Balance != SiteCredit {
			return invalid(o.ID, "unknown balance type")
		}
		if len(o.Currency) != 3 || strings.ToUpper(o.Currency) != o.Currency ||
			strings.IndexFunc(o.Currency, func(r rune) bool { return r < 'A' || r > 'Z' }) >= 0 ||
			o.PriceMinor <= 0 || o.PriceMinor > 1_000_000_000_000 {
			return invalid(o.ID, "require a three-letter uppercase currency and a positive bounded price in minor units")
		}
		if o.ObservedAt.IsZero() || o.ObservedAt.After(d.AsOf) {
			return invalid(o.ID, "timestamp must be present and no later than as_of")
		}
		return nil
	}
	for _, o := range d.Observations {
		if err := validateObservation(o); err != nil {
			return nil, err
		}
	}
	for _, o := range d.Candidates {
		if err := validateObservation(o); err != nil {
			return nil, err
		}
		if o.Kind != Ask {
			return nil, invalid(o.ID, "candidate must be an asking-price observation")
		}
	}
	return skins, nil
}

func sameMarket(a, b Observation) bool {
	return a.Skin == b.Skin && a.Status == b.Status && a.Venue == b.Venue &&
		a.Currency == b.Currency && a.Balance == b.Balance
}

func summarize(items []Observation) Summary {
	prices := make([]int64, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, o := range items {
		prices = append(prices, o.PriceMinor)
		ids = append(ids, o.ID)
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i] < prices[j] })
	sort.Strings(ids)
	mid := len(prices) / 2
	median := prices[mid]
	if len(prices)%2 == 0 {
		median = prices[mid-1] + (prices[mid]-prices[mid-1])/2
	}
	return Summary{median, prices[0], prices[len(prices)-1], ids}
}

func estimate(candidate Observation, pool []Observation, skin Skin, c Config, kind EvidenceKind) *Estimate {
	if candidate.Float == nil {
		return nil
	}
	value := *candidate.Float
	b := wearBand(value)
	low, high := math.Max(b.low, skin.MinFloat), math.Min(b.high, skin.MaxFloat)
	// Use the middle half of the attainable tier as the ordinary-float baseline.
	// This avoids treating low-float or extreme high-float examples as ordinary.
	baseLow, baseHigh := low+(high-low)*.25, low+(high-low)*.75
	local := make([]Observation, 0)
	baseline := make([]Observation, 0)
	better := make([]Observation, 0)
	var previous band
	for i, tier := range bands {
		if tier.wear == b.wear && i > 0 {
			previous = bands[i-1]
		}
	}
	for _, o := range pool {
		if o.Float == nil || o.Kind != kind || !sameMarket(candidate, o) {
			continue
		}
		f := *o.Float
		if wearBand(f).wear == b.wear {
			if math.Abs(f-value) <= c.FloatRadius {
				local = append(local, o)
			}
			if f >= baseLow && f <= baseHigh {
				baseline = append(baseline, o)
			}
		}
		if previous.wear != "" && wearBand(f).wear == previous.wear {
			better = append(better, o)
		}
	}
	if len(local) < c.MinComparables || len(baseline) < c.MinComparables {
		return nil
	}
	sort.Slice(local, func(i, j int) bool {
		if local[i].Float == nil || local[j].Float == nil {
			return local[i].ID < local[j].ID
		}
		a, z := math.Abs(*local[i].Float-value), math.Abs(*local[j].Float-value)
		if a == z {
			return local[i].ID < local[j].ID
		}
		return a < z
	})
	if len(local) > c.MaxComparables {
		local = local[:c.MaxComparables]
	}
	comp, base := summarize(local), summarize(baseline)
	discount := comp.MedianMinor - candidate.PriceMinor
	result := &Estimate{
		Basis: kind, ComparablePrices: comp, CategoryBaseline: base,
		FloatPremiumMinor:    comp.MedianMinor - base.MedianMinor,
		GrossDiscountMinor:   discount,
		GrossDiscountPercent: float64(discount) / float64(comp.MedianMinor) * 100,
	}
	if len(better) >= c.MinComparables {
		result.BetterTier = &BetterTier{previous.wear, value - b.low, summarize(better)}
	}
	return result
}

// Evaluate estimates float-associated premiums within exact skin/status/wear/
// venue/currency/balance groups. Sales take precedence when both baseline and
// local evidence are sufficient; otherwise asks are used without mixing them.
// Expected validation failures are returned as typed *Error values.
func Evaluate(d Dataset) ([]Report, error) {
	skins, err := validate(d)
	if err != nil {
		return nil, err
	}
	cutoff := d.AsOf.Add(-time.Duration(d.Config.MaxAgeHours) * time.Hour)
	reports := make([]Report, 0, len(d.Candidates))
	for _, candidate := range d.Candidates {
		if candidate.Float == nil {
			return nil, invalid(candidate.ID, "float is required")
		}
		report := Report{Candidate: candidate, Wear: wearBand(*candidate.Float).wear,
			Model: "cs2-float-comparables-v1", State: "insufficient_evidence"}
		if candidate.ObservedAt.Before(cutoff) {
			report.State, report.Reason = "stale_candidate", "candidate predates the configured evidence window"
			reports = append(reports, report)
			continue
		}
		// One vote per physical asset, excluding the candidate itself. Deduplicate
		// before filtering categories so an older snapshot cannot survive a newer one.
		latest := make(map[string]Observation)
		for _, o := range d.Observations {
			if o.AssetID == candidate.AssetID || o.ObservedAt.Before(cutoff) {
				continue
			}
			old, exists := latest[o.AssetID]
			if !exists || o.ObservedAt.After(old.ObservedAt) ||
				(o.ObservedAt.Equal(old.ObservedAt) && (o.Kind == Sale && old.Kind != Sale || o.Kind == old.Kind && o.ID < old.ID)) {
				latest[o.AssetID] = o
			}
		}
		pool := make([]Observation, 0, len(latest))
		for _, o := range latest {
			pool = append(pool, o)
		}
		for _, kind := range []EvidenceKind{Sale, Ask} {
			report.Estimate = estimate(candidate, pool, skins[candidate.Skin], d.Config, kind)
			if report.Estimate != nil {
				report.State = "estimated"
				break
			}
		}
		if report.Estimate == nil {
			report.Reason = "need sufficient independent nearby-float and ordinary-float examples of the same evidence kind and exact market category"
		}
		reports = append(reports, report)
	}
	return reports, nil
}

package valuation

import (
	"errors"
	"math"
	"reflect"
	"testing"
	"time"
)

func floatPointer(value float64) *float64 { return &value }

func fixture() Dataset {
	now := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC)
	observation := func(id string, f float64, price int64, kind EvidenceKind) Observation {
		return Observation{ID: id, AssetID: id, Skin: "Example Rifle | Example Finish", Status: Normal,
			Float: floatPointer(f), PriceMinor: price, Currency: "USD", Balance: Cash,
			Venue: "example", Kind: kind, ObservedAt: now.Add(-time.Hour)}
	}
	return Dataset{
		AsOf: now, Config: Config{MaxAgeHours: 72, FloatRadius: .01, MinComparables: 3, MaxComparables: 8},
		Skins: []Skin{{Name: "Example Rifle | Example Finish", MinFloat: .06, MaxFloat: .8}},
		Observations: []Observation{
			observation("low-1", .1505, 1400, Sale), observation("low-2", .1515, 1500, Sale),
			observation("low-3", .154, 1600, Sale),
			observation("ordinary-1", .25, 900, Sale), observation("ordinary-2", .26, 1000, Sale),
			observation("ordinary-3", .27, 1100, Sale),
			observation("mw-1", .08, 2100, Sale), observation("mw-2", .10, 2200, Sale),
			observation("mw-3", .13, 2300, Sale),
		},
		Candidates: []Observation{observation("candidate", .1501, 1050, Ask)},
	}
}

func requireEstimate(t *testing.T, d Dataset) Estimate {
	t.Helper()
	reports, err := Evaluate(d)
	if err != nil {
		t.Fatal(err)
	}
	if len(reports) != 1 || reports[0].State != "estimated" || reports[0].Estimate == nil {
		t.Fatalf("expected estimate, got %+v", reports)
	}
	return *reports[0].Estimate
}

func TestFloatPremiumUsesSameTierEvidence(t *testing.T) {
	d := fixture()
	got := requireEstimate(t, d)
	if got.Basis != Sale || got.ComparablePrices.MedianMinor != 1500 || got.CategoryBaseline.MedianMinor != 1000 ||
		got.FloatPremiumMinor != 500 || got.GrossDiscountMinor != 450 {
		t.Fatalf("unexpected estimate: %+v", got)
	}
	if math.Abs(got.GrossDiscountPercent-450.0/1500*100) > .0001 {
		t.Fatal("wrong discount percentage")
	}
	if got.BetterTier == nil || got.BetterTier.Wear != MinimalWear || got.BetterTier.Prices.MedianMinor != 2200 {
		t.Fatalf("missing better-tier context: %+v", got.BetterTier)
	}
	if got.ComparablePrices.MedianMinor == got.BetterTier.Prices.MedianMinor {
		t.Fatal("borrowed better-tier price")
	}
	if !reflect.DeepEqual(got.ComparablePrices.EvidenceIDs, []string{"low-1", "low-2", "low-3"}) {
		t.Fatal("missing comparable provenance")
	}
}

func TestNoPremiumInOrdinaryFloatBand(t *testing.T) {
	d := fixture()
	d.Candidates[0].Float = floatPointer(.26)
	d.Config.FloatRadius = .011
	got := requireEstimate(t, d)
	if got.FloatPremiumMinor != 0 || got.GrossDiscountMinor != -50 {
		t.Fatalf("unexpected estimate: %+v", got)
	}
}

func TestAskFallbackDoesNotMixSalesAndAsks(t *testing.T) {
	d := fixture()
	for _, o := range d.Observations {
		o.ID += "-ask"
		o.AssetID = o.ID
		o.Kind = Ask
		o.PriceMinor += 200
		d.Observations = append(d.Observations, o)
	}
	if got := requireEstimate(t, d); got.Basis != Sale || got.ComparablePrices.MedianMinor != 1500 {
		t.Fatalf("should prefer sales: %+v", got)
	}
	d.Observations = d.Observations[1:]
	got := requireEstimate(t, d)
	if got.Basis != Ask || got.ComparablePrices.MedianMinor != 1700 || got.CategoryBaseline.MedianMinor != 1200 {
		t.Fatalf("should fall back to independent ask estimate: %+v", got)
	}
}

func TestCategoryAndFreshnessIsolation(t *testing.T) {
	for _, name := range []string{"status", "skin", "venue", "currency", "balance", "wear", "stale", "same_asset", "distant_float", "no_baseline"} {
		t.Run(name, func(t *testing.T) {
			d := fixture()
			switch name {
			case "status":
				d.Candidates[0].Status = StatTrak
			case "skin":
				d.Skins = append(d.Skins, Skin{"Other Skin", 0, 1})
				d.Candidates[0].Skin = "Other Skin"
			case "venue":
				d.Candidates[0].Venue = "elsewhere"
			case "currency":
				d.Candidates[0].Currency = "AUD"
			case "balance":
				d.Candidates[0].Balance = SteamWallet
			case "wear":
				d.Candidates[0].Float = floatPointer(.07)
			case "stale":
				d.Observations[0].ObservedAt = d.AsOf.Add(-73 * time.Hour)
			case "same_asset":
				d.Observations[0].AssetID = d.Candidates[0].AssetID
			case "distant_float":
				d.Candidates[0].Float = floatPointer(.18)
			case "no_baseline":
				d.Observations = d.Observations[:3]
			}
			reports, err := Evaluate(d)
			if err != nil {
				t.Fatal(err)
			}
			if reports[0].State != "insufficient_evidence" || reports[0].Estimate != nil {
				t.Fatalf("must not invent value: %+v", reports[0])
			}
		})
	}
}

func TestDeduplicatesAssetsBeforeCountingEvidence(t *testing.T) {
	d := fixture()
	d.Observations[1].AssetID = d.Observations[0].AssetID
	d.Observations[1].ObservedAt = d.Observations[0].ObservedAt.Add(time.Minute)
	reports, err := Evaluate(d)
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].Estimate != nil {
		t.Fatal("counted repeated snapshots as independent evidence")
	}
}

func TestLatestSnapshotReplacesOlderPrice(t *testing.T) {
	d := fixture()
	old := d.Observations[0]
	old.ID = "old-overpriced-snapshot"
	old.PriceMinor = 99999
	old.ObservedAt = old.ObservedAt.Add(-time.Minute)
	d.Observations = append(d.Observations, old)
	got := requireEstimate(t, d)
	if got.ComparablePrices.MaximumMinor != 1600 {
		t.Fatalf("used obsolete snapshot: %+v", got)
	}
}

func TestAttainableRangeChangesBaseline(t *testing.T) {
	d := fixture()
	d.Skins[0].MinFloat = .145
	d.Skins[0].MaxFloat = .2
	d.Observations = d.Observations[:3]
	for i, f := range []float64{.17, .175, .18} {
		o := d.Observations[i]
		o.ID += "-baseline"
		o.AssetID = o.ID
		o.Float = floatPointer(f)
		o.PriceMinor = 1000
		d.Observations = append(d.Observations, o)
	}
	if got := requireEstimate(t, d); got.CategoryBaseline.MedianMinor != 1000 {
		t.Fatalf("did not use attainable range: %+v", got)
	}
}

func TestStaleCandidate(t *testing.T) {
	d := fixture()
	d.Candidates[0].ObservedAt = d.AsOf.Add(-73 * time.Hour)
	reports, err := Evaluate(d)
	if err != nil {
		t.Fatal(err)
	}
	if reports[0].State != "stale_candidate" {
		t.Fatalf("unexpected state: %+v", reports[0])
	}
}

func TestWearBoundaries(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  Wear
	}{
		{0, FactoryNew}, {.069999, FactoryNew}, {.07, MinimalWear}, {.149999, MinimalWear},
		{.15, FieldTested}, {.38, WellWorn}, {.45, BattleScarred}, {1, BattleScarred},
	} {
		if got := wearBand(tc.value).wear; got != tc.want {
			t.Errorf("%g: got %s, want %s", tc.value, got, tc.want)
		}
	}
}

func TestRejectsInvalidInputWithTypedError(t *testing.T) {
	for _, name := range []string{"missing_float", "nan_float", "range", "status", "price", "currency", "balance", "future", "missing_time", "config", "radius", "duplicate", "unknown_skin", "candidate_sale", "skin_range", "duplicate_skin", "missing_as_of"} {
		t.Run(name, func(t *testing.T) {
			d := fixture()
			switch name {
			case "missing_float":
				d.Candidates[0].Float = nil
			case "nan_float":
				d.Candidates[0].Float = floatPointer(math.NaN())
			case "range":
				d.Candidates[0].Float = floatPointer(.9)
			case "status":
				d.Candidates[0].Status = "unknown"
			case "price":
				d.Candidates[0].PriceMinor = 0
			case "currency":
				d.Candidates[0].Currency = "usd"
			case "balance":
				d.Candidates[0].Balance = "unknown"
			case "future":
				d.Observations[0].ObservedAt = d.AsOf.Add(time.Second)
			case "missing_time":
				d.Observations[0].ObservedAt = time.Time{}
			case "config":
				d.Config.MinComparables = 1
			case "radius":
				d.Config.FloatRadius = math.Inf(1)
			case "duplicate":
				d.Observations[0].ID = d.Candidates[0].ID
			case "unknown_skin":
				d.Candidates[0].Skin = "missing"
			case "candidate_sale":
				d.Candidates[0].Kind = Sale
			case "skin_range":
				d.Skins[0].MinFloat = .9
			case "duplicate_skin":
				d.Skins = append(d.Skins, d.Skins[0])
			case "missing_as_of":
				d.AsOf = time.Time{}
			}
			_, err := Evaluate(d)
			var typed *Error
			if !errors.As(err, &typed) || typed == nil || typed.Code != "invalid_input" {
				t.Fatalf("expected typed validation error, got %v", err)
			}
		})
	}
}

func TestDeterministicAndDoesNotMutateInput(t *testing.T) {
	d := fixture()
	before := append([]Observation(nil), d.Observations...)
	first, err := Evaluate(d)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		got, err := Evaluate(d)
		if err != nil || !reflect.DeepEqual(first, got) {
			t.Fatal("non-deterministic output", err)
		}
	}
	if !reflect.DeepEqual(before, d.Observations) {
		t.Fatal("mutated input")
	}
}

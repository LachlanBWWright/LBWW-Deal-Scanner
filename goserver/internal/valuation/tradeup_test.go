package valuation

import (
	"errors"
	"fmt"
	"math"
	"testing"
)

func tradeFloat(value float64) *float64 { return &value }

func tradeFixture() TradeStudy {
	skin := Skin{Name: "Synthetic input", MinFloat: 0, MaxFloat: 1}
	s := TradeStudy{
		Model: NormalizedAverage, Currency: "USD", Balance: Cash, OutputSafetyMargin: .00001,
		Candidate:   TradeInput{AssetID: "candidate", Skin: skin, Collection: "A", Rarity: 4, Status: Normal, Float: tradeFloat(.01), CostMinor: 300},
		Alternative: TradeInput{AssetID: "alternative", Skin: skin, Collection: "A", Rarity: 4, Status: Normal, Float: tradeFloat(.1), CostMinor: 100},
		Outputs: []TradeOutput{
			{Skin: Skin{Name: "Synthetic high output", MinFloat: 0, MaxFloat: 1}, Collection: "A", NetByWear: map[Wear]int64{FactoryNew: 4000, MinimalWear: 2000}},
			{Skin: Skin{Name: "Synthetic low output", MinFloat: .06, MaxFloat: .8}, Collection: "A", NetByWear: map[Wear]int64{MinimalWear: 500}},
		},
	}
	for i := 0; i < 9; i++ {
		s.Fixed = append(s.Fixed, TradeInput{AssetID: fmt.Sprintf("fixed-%d", i), Skin: skin, Collection: "A", Rarity: 4, Status: Normal, Float: tradeFloat(.075), CostMinor: 100})
	}
	return s
}

func closeTrade(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-8 {
		t.Fatalf("got %.12f, want %.12f", got, want)
	}
}

func TestTradeMarginalValueAndStatistics(t *testing.T) {
	report, err := EvaluateTradeStudy(tradeFixture())
	if err != nil {
		t.Fatal(err)
	}
	closeTrade(t, report.Candidate.Outcomes[0].Float, .0685)
	closeTrade(t, report.Alternative.Outcomes[0].Float, .0775)
	closeTrade(t, report.Candidate.ExpectedNetProceedsMinor, 2250)
	closeTrade(t, report.Candidate.ExpectedProfitMinor, 1050)
	closeTrade(t, report.Alternative.ExpectedProfitMinor, 250)
	closeTrade(t, report.MarginalExpectedProceedsMinor, 1000)
	closeTrade(t, report.MarginalExpectedProfitMinor, 800)
	closeTrade(t, report.RelativeCandidatePriceCeilingMinor, 1100)
	closeTrade(t, report.ContractBreakEvenCandidatePriceMinor, 1350)
	closeTrade(t, report.Candidate.ProfitProbability, .5)
	closeTrade(t, report.Candidate.LossProbability, .5)
	closeTrade(t, report.Candidate.ProfitStdDevMinor, 1750)
	if report.Candidate.WorstProfitMinor != -700 || report.Candidate.BestProfitMinor != 2800 {
		t.Fatal("incorrect profit range")
	}
	threshold := report.Thresholds[0]
	closeTrade(t, threshold.InputFloatExclusive, .0249)
	if !threshold.CandidateQualifies || threshold.AlternativeQualifies || threshold.TargetWear != FactoryNew {
		t.Fatalf("unexpected threshold: %+v", threshold)
	}
}

func TestTradeCollectionWeightedOdds(t *testing.T) {
	s := tradeFixture()
	s.Fixed[0].Collection = "B"
	s.Fixed[0].Skin.Name = "B input"
	s.Outputs = append(s.Outputs, TradeOutput{Skin: Skin{Name: "B output", MaxFloat: 1}, Collection: "B", NetByWear: map[Wear]int64{FactoryNew: 1000, MinimalWear: 1000}})
	r, err := EvaluateTradeStudy(s)
	if err != nil {
		t.Fatal(err)
	}
	closeTrade(t, r.Candidate.Outcomes[0].Probability, .45)
	closeTrade(t, r.Candidate.Outcomes[1].Probability, .45)
	closeTrade(t, r.Candidate.Outcomes[2].Probability, .1)
}

func TestTradeNormalizationAndThresholds(t *testing.T) {
	s := tradeFixture()
	s.Candidate.Skin = Skin{Name: "Capped candidate", MinFloat: .06, MaxFloat: .8}
	s.Candidate.Float = tradeFloat(.0674) // normalized .01, despite raw .0674
	r, err := EvaluateTradeStudy(s)
	if err != nil {
		t.Fatal(err)
	}
	closeTrade(t, r.Candidate.Outcomes[0].Float, .0685)
	closeTrade(t, r.Thresholds[0].InputFloatExclusive, .06+.0249*.74)
	if !r.Thresholds[0].CandidateQualifies || r.Thresholds[0].AlternativeQualifies {
		t.Fatal("alternative must use its own range")
	}
	s.Model = RawAverage
	r, err = EvaluateTradeStudy(s)
	if err != nil {
		t.Fatal(err)
	}
	closeTrade(t, r.Candidate.Outcomes[0].Float, .07424)
	if r.Thresholds[0].CandidateQualifies || r.Candidate.Outcomes[0].Wear != MinimalWear {
		t.Fatal("raw-average model not respected")
	}
}

func TestTradeBoundaryGuardAndUnreachableFloats(t *testing.T) {
	s := tradeFixture()
	s.Candidate.Float = tradeFloat(.025)
	r, err := EvaluateTradeStudy(s)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Candidate.Outcomes[0].NearBoundary || r.Thresholds[0].CandidateQualifies {
		t.Fatal("near-boundary result must not pass safety margin")
	}
	s.Candidate.Skin = Skin{Name: "Capped", MinFloat: .06, MaxFloat: .8}
	s.Candidate.Float = tradeFloat(.06)
	s.Model = RawAverage
	r, err = EvaluateTradeStudy(s)
	if err != nil {
		t.Fatal(err)
	}
	if r.Thresholds[0].Feasibility != "impossible" {
		t.Fatal("threshold outside range must be impossible")
	}
	if r.Thresholds[1].Feasibility != "all_floats" {
		t.Fatal("loose threshold should accept all attainable inputs")
	}
}

func TestTradeNoThresholdCrossingMeansNoTierValue(t *testing.T) {
	s := tradeFixture()
	s.Candidate.Float = tradeFloat(.09)
	r, err := EvaluateTradeStudy(s)
	if err != nil {
		t.Fatal(err)
	}
	closeTrade(t, r.MarginalExpectedProceedsMinor, 0)
	closeTrade(t, r.MarginalExpectedProfitMinor, -200)
	closeTrade(t, r.RelativeCandidatePriceCeilingMinor, 100)
}

func TestTradeRejectsInvalidInputs(t *testing.T) {
	tests := map[string]func(*TradeStudy){
		"missing model":            func(s *TradeStudy) { s.Model = "" },
		"wrong count":              func(s *TradeStudy) { s.Fixed = s.Fixed[:8] },
		"missing float":            func(s *TradeStudy) { s.Candidate.Float = nil },
		"nan":                      func(s *TradeStudy) { s.Candidate.Float = tradeFloat(math.NaN()) },
		"out of range":             func(s *TradeStudy) { s.Candidate.Float = tradeFloat(2) },
		"duplicate asset":          func(s *TradeStudy) { s.Fixed[0].AssetID = s.Candidate.AssetID },
		"conflicting caps":         func(s *TradeStudy) { s.Fixed[0].Skin.MaxFloat = .9 },
		"different collection":     func(s *TradeStudy) { s.Alternative.Collection = "B" },
		"different rarity":         func(s *TradeStudy) { s.Fixed[0].Rarity = 3 },
		"different status":         func(s *TradeStudy) { s.Fixed[0].Status = StatTrak },
		"covert":                   func(s *TradeStudy) { s.Candidate.Rarity = 6 },
		"souvenir":                 func(s *TradeStudy) { s.Candidate.Status = Souvenir },
		"missing pool":             func(s *TradeStudy) { s.Outputs = nil },
		"duplicate outcome":        func(s *TradeStudy) { s.Outputs = append(s.Outputs, s.Outputs[0]) },
		"unrepresented collection": func(s *TradeStudy) { s.Outputs[0].Collection = "B" },
		"missing price":            func(s *TradeStudy) { delete(s.Outputs[0].NetByWear, FactoryNew) },
		"negative proceeds":        func(s *TradeStudy) { s.Outputs[0].NetByWear[FactoryNew] = -1 },
		"bad wear":                 func(s *TradeStudy) { s.Outputs[0].NetByWear[Wear("bad")] = 1 },
		"invalid range":            func(s *TradeStudy) { s.Candidate.Skin.MaxFloat = 0 },
		"negative cost":            func(s *TradeStudy) { s.Candidate.CostMinor = -1 },
		"unsafe margin":            func(s *TradeStudy) { s.OutputSafetyMargin = 0 },
		"bad balance":              func(s *TradeStudy) { s.Balance = "bad" },
		"bad currency":             func(s *TradeStudy) { s.Currency = "usd" },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			s := tradeFixture()
			mutate(&s)
			_, err := EvaluateTradeStudy(s)
			var domain *Error
			if !errors.As(err, &domain) || domain.Code != "invalid_input" {
				t.Fatalf("expected typed validation error, got %v", err)
			}
		})
	}
}

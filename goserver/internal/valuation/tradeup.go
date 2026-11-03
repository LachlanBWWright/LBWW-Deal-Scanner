package valuation

import (
	"fmt"
	"math"
)

// FloatModel is explicit because published calculators disagree. Neither model
// has been verified here against game-produced outcomes or engine rounding.
type FloatModel string

const (
	RawAverage        FloatModel = "raw_average"
	NormalizedAverage FloatModel = "normalized_average"
)

type TradeInput struct {
	AssetID    string   `json:"asset_id"`
	Skin       Skin     `json:"skin"`
	Collection string   `json:"collection"`
	Rarity     int      `json:"rarity"` // 1=Consumer ... 5=Classified; Covert recipes excluded.
	Status     Status   `json:"status"`
	Float      *float64 `json:"float"`
	CostMinor  int64    `json:"cost_minor"` // All-in purchase cost, including buyer fees.
}

type TradeOutput struct {
	Skin       Skin   `json:"skin"`
	Collection string `json:"collection"`
	// Caller supplies achievable seller proceeds, after all selling costs, in
	// the contract's single currency/balance. Missing required wear is an error.
	NetByWear map[Wear]int64 `json:"net_by_wear"`
}

// TradeStudy compares two alternatives for the tenth slot of a standard
// contract. Same collection/status/rarity keeps the outcome odds unchanged.
type TradeStudy struct {
	Model              FloatModel   `json:"float_model"`
	Currency           string       `json:"currency"`
	Balance            Balance      `json:"balance"`
	OutputSafetyMargin float64      `json:"output_safety_margin"`
	Fixed              []TradeInput `json:"fixed_inputs"`
	Candidate          TradeInput   `json:"candidate"`
	Alternative        TradeInput   `json:"alternative"`
	// Must be the complete next-rarity pool for every represented collection.
	// Completeness and actual in-game eligibility require an external catalog.
	Outputs []TradeOutput `json:"outputs"`
}

type TradeOutcome struct {
	Skin             string  `json:"skin"`
	Probability      float64 `json:"probability"`
	Float            float64 `json:"float"`
	Wear             Wear    `json:"wear"`
	NearBoundary     bool    `json:"near_boundary"`
	NetProceedsMinor int64   `json:"net_proceeds_minor"`
	ProfitMinor      int64   `json:"profit_minor"`
}

type TradeStatistics struct {
	InputCostMinor           int64          `json:"input_cost_minor"`
	ExpectedNetProceedsMinor float64        `json:"expected_net_proceeds_minor"`
	ExpectedProfitMinor      float64        `json:"expected_profit_minor"`
	ProfitProbability        float64        `json:"profit_probability"`
	LossProbability          float64        `json:"loss_probability"`
	ProfitStdDevMinor        float64        `json:"profit_std_dev_minor"`
	WorstProfitMinor         int64          `json:"worst_profit_minor"`
	BestProfitMinor          int64          `json:"best_profit_minor"`
	Outcomes                 []TradeOutcome `json:"outcomes"`
}

type TradeThreshold struct {
	Skin                    string  `json:"skin"`
	Probability             float64 `json:"probability"`
	TargetWear              Wear    `json:"target_wear"`
	OutputBoundaryExclusive float64 `json:"output_boundary_exclusive"`
	// Maximum tenth-slot raw float, exclusive. May lie outside its attainable
	// range; feasibility explains this. Safety margin is already subtracted.
	InputFloatExclusive  float64 `json:"input_float_exclusive"`
	Feasibility          string  `json:"feasibility"` // impossible, some_floats, all_floats
	CandidateQualifies   bool    `json:"candidate_qualifies"`
	AlternativeQualifies bool    `json:"alternative_qualifies"`
}

type TradeReport struct {
	Model                         FloatModel      `json:"float_model"`
	Currency                      string          `json:"currency"`
	Balance                       Balance         `json:"balance"`
	Candidate                     TradeStatistics `json:"candidate"`
	Alternative                   TradeStatistics `json:"alternative"`
	MarginalExpectedProceedsMinor float64         `json:"marginal_expected_proceeds_minor"`
	ExtraInputCostMinor           int64           `json:"extra_input_cost_minor"`
	MarginalExpectedProfitMinor   float64         `json:"marginal_expected_profit_minor"`
	// Risk-neutral ceiling relative to the alternative, NOT an item appraisal
	// or an absolute profitable purchase ceiling.
	RelativeCandidatePriceCeilingMinor   float64          `json:"relative_candidate_price_ceiling_minor"`
	ContractBreakEvenCandidatePriceMinor float64          `json:"contract_break_even_candidate_price_minor"`
	Thresholds                           []TradeThreshold `json:"thresholds"`
}

func validTradeSkin(s Skin) bool {
	return s.Name != "" && finite(s.MinFloat) && finite(s.MaxFloat) &&
		s.MinFloat >= 0 && s.MaxFloat <= 1 && s.MinFloat < s.MaxFloat
}

func validateTradeStudy(s TradeStudy) error {
	if s.Model != RawAverage && s.Model != NormalizedAverage {
		return invalid("float_model", "select raw_average or normalized_average explicitly")
	}
	if len(s.Currency) != 3 || s.Currency[0] < 'A' || s.Currency[0] > 'Z' ||
		s.Currency[1] < 'A' || s.Currency[1] > 'Z' || s.Currency[2] < 'A' || s.Currency[2] > 'Z' {
		return invalid("currency", "expected three uppercase letters")
	}
	if s.Balance != Cash && s.Balance != SteamWallet && s.Balance != SiteCredit {
		return invalid("balance", "unknown balance type")
	}
	if !finite(s.OutputSafetyMargin) || s.OutputSafetyMargin <= 0 || s.OutputSafetyMargin > .001 {
		return invalid("output_safety_margin", "must be positive and at most 0.001")
	}
	if len(s.Fixed) != 9 {
		return invalid("fixed_inputs", "only nine-fixed-plus-one standard contracts are supported")
	}
	inputs := append(append([]TradeInput{}, s.Fixed...), s.Candidate, s.Alternative)
	assets := make(map[string]bool)
	skins := make(map[string]TradeInput)
	collections := make(map[string]bool)
	for i, input := range inputs {
		field := fmt.Sprintf("inputs[%d]", i)
		if input.AssetID == "" || assets[input.AssetID] || !validTradeSkin(input.Skin) || input.Collection == "" {
			return invalid(field, "unique asset ID, valid skin range and collection required")
		}
		assets[input.AssetID] = true
		if old, exists := skins[input.Skin.Name]; exists && (old.Skin != input.Skin || old.Collection != input.Collection) {
			return invalid(field, "conflicting float ranges or collections for one skin")
		}
		skins[input.Skin.Name] = input
		if input.Float == nil || !finite(*input.Float) || *input.Float < input.Skin.MinFloat || *input.Float > input.Skin.MaxFloat {
			return invalid(field, "float must be present and attainable")
		}
		if input.Rarity < 1 || input.Rarity > 5 || input.Rarity != s.Candidate.Rarity ||
			(input.Status != Normal && input.Status != StatTrak) || input.Status != s.Candidate.Status {
			return invalid(field, "requires matching regular or StatTrak inputs at rarity 1 through 5")
		}
		if input.CostMinor < 0 || input.CostMinor > 1e12 {
			return invalid(field, "cost must be between zero and 1e12 minor units")
		}
		collections[input.Collection] = true
	}
	if s.Candidate.Collection != s.Alternative.Collection {
		return invalid("alternative.collection", "replacement must preserve collection odds")
	}
	seenOutputs := make(map[string]bool)
	pools := make(map[string]int)
	for i, output := range s.Outputs {
		field := fmt.Sprintf("outputs[%d]", i)
		if !validTradeSkin(output.Skin) || seenOutputs[output.Skin.Name] || !collections[output.Collection] {
			return invalid(field, "requires a unique valid output from an input collection")
		}
		seenOutputs[output.Skin.Name] = true
		pools[output.Collection]++
		for wear, price := range output.NetByWear {
			known := false
			for _, b := range bands {
				if b.wear == wear {
					known = true
				}
			}
			if !known || price < 0 || price > 1e12 {
				return invalid(field, "invalid wear or proceeds outside zero to 1e12")
			}
		}
	}
	for collection := range collections {
		if pools[collection] == 0 {
			return invalid("outputs", "missing next-rarity collection pool: "+collection)
		}
	}
	return nil
}

func tradePosition(input TradeInput, model FloatModel) float64 {
	if input.Float == nil {
		return 0
	} // Validated before evaluation.
	if model == RawAverage {
		return *input.Float
	}
	return (*input.Float - input.Skin.MinFloat) / (input.Skin.MaxFloat - input.Skin.MinFloat)
}

func tradeStatistics(s TradeStudy, slot TradeInput, fixedSum float64, counts, pools map[string]int) (TradeStatistics, error) {
	stats := TradeStatistics{InputCostMinor: slot.CostMinor}
	for _, input := range s.Fixed {
		stats.InputCostMinor += input.CostMinor
	}
	average := (fixedSum + tradePosition(slot, s.Model)) / 10
	for i, output := range s.Outputs {
		value := output.Skin.MinFloat + average*(output.Skin.MaxFloat-output.Skin.MinFloat)
		wear := wearBand(value).wear
		price, exists := output.NetByWear[wear]
		if !exists {
			return TradeStatistics{}, invalid(fmt.Sprintf("outputs[%d].net_by_wear", i), "missing proceeds for predicted wear: "+string(wear))
		}
		probability := float64(counts[output.Collection]) / 10 / float64(pools[output.Collection])
		profit := price - stats.InputCostMinor
		near := false
		for _, b := range bands[1:] {
			if math.Abs(value-b.low) <= s.OutputSafetyMargin {
				near = true
			}
		}
		stats.Outcomes = append(stats.Outcomes, TradeOutcome{output.Skin.Name, probability, value, wear, near, price, profit})
		stats.ExpectedNetProceedsMinor += probability * float64(price)
		if profit > 0 {
			stats.ProfitProbability += probability
		}
		if profit < 0 {
			stats.LossProbability += probability
		}
		if i == 0 || profit < stats.WorstProfitMinor {
			stats.WorstProfitMinor = profit
		}
		if i == 0 || profit > stats.BestProfitMinor {
			stats.BestProfitMinor = profit
		}
	}
	stats.ExpectedProfitMinor = stats.ExpectedNetProceedsMinor - float64(stats.InputCostMinor)
	variance := 0.0
	for _, outcome := range stats.Outcomes {
		delta := float64(outcome.ProfitMinor) - stats.ExpectedProfitMinor
		variance += outcome.Probability * delta * delta
	}
	stats.ProfitStdDevMinor = math.Sqrt(variance)
	return stats, nil
}

// EvaluateTradeStudy computes conditional contract statistics, not empirical
// market estimates. Prices and the complete eligible outcome pool are supplied.
func EvaluateTradeStudy(s TradeStudy) (TradeReport, error) {
	if err := validateTradeStudy(s); err != nil {
		return TradeReport{}, err
	}
	fixedSum := 0.0
	counts, pools := make(map[string]int), make(map[string]int)
	for _, input := range s.Fixed {
		fixedSum += tradePosition(input, s.Model)
		counts[input.Collection]++
	}
	counts[s.Candidate.Collection]++
	for _, output := range s.Outputs {
		pools[output.Collection]++
	}
	candidate, err := tradeStatistics(s, s.Candidate, fixedSum, counts, pools)
	if err != nil {
		return TradeReport{}, err
	}
	alternative, err := tradeStatistics(s, s.Alternative, fixedSum, counts, pools)
	if err != nil {
		return TradeReport{}, err
	}
	delta := candidate.ExpectedNetProceedsMinor - alternative.ExpectedNetProceedsMinor
	extra := s.Candidate.CostMinor - s.Alternative.CostMinor
	report := TradeReport{
		Model: s.Model, Currency: s.Currency, Balance: s.Balance, Candidate: candidate, Alternative: alternative,
		MarginalExpectedProceedsMinor: delta, ExtraInputCostMinor: extra, MarginalExpectedProfitMinor: delta - float64(extra),
		RelativeCandidatePriceCeilingMinor:   float64(s.Alternative.CostMinor) + delta,
		ContractBreakEvenCandidatePriceMinor: candidate.ExpectedNetProceedsMinor - float64(candidate.InputCostMinor-s.Candidate.CostMinor),
	}
	for i, output := range s.Outputs {
		for _, b := range bands[:len(bands)-1] {
			if b.high <= output.Skin.MinFloat || b.high > output.Skin.MaxFloat {
				continue
			}
			limit := 10*(b.high-s.OutputSafetyMargin-output.Skin.MinFloat)/(output.Skin.MaxFloat-output.Skin.MinFloat) - fixedSum
			candidateLimit, alternativeLimit := limit, limit
			if s.Model == NormalizedAverage {
				candidateLimit = s.Candidate.Skin.MinFloat + limit*(s.Candidate.Skin.MaxFloat-s.Candidate.Skin.MinFloat)
				alternativeLimit = s.Alternative.Skin.MinFloat + limit*(s.Alternative.Skin.MaxFloat-s.Alternative.Skin.MinFloat)
			}
			feasibility := "some_floats"
			if candidateLimit <= s.Candidate.Skin.MinFloat {
				feasibility = "impossible"
			}
			if candidateLimit > s.Candidate.Skin.MaxFloat {
				feasibility = "all_floats"
			}
			qualifies, alternativeQualifies := false, false
			if s.Candidate.Float != nil {
				qualifies = *s.Candidate.Float < candidateLimit
			}
			if s.Alternative.Float != nil {
				alternativeQualifies = *s.Alternative.Float < alternativeLimit
			}
			report.Thresholds = append(report.Thresholds, TradeThreshold{
				output.Skin.Name, candidate.Outcomes[i].Probability, b.wear, b.high, candidateLimit,
				feasibility, qualifies, alternativeQualifies,
			})
		}
	}
	return report, nil
}

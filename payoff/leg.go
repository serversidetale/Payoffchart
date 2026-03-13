package payoff

// OptionType is call or put.
type OptionType string

const (
	Call OptionType = "call"
	Put  OptionType = "put"
)

// Side is long or short (position direction).
type Side string

const (
	Long  Side = "long"
	Short Side = "short"
)

// Leg represents one option leg in a strategy.
type Leg struct {
	Type       OptionType // "call" or "put"
	Side       Side       // "long" or "short"
	Strike     float64    // strike price
	Premium    float64    // option premium per share (paid/received)
	Contracts  int        // number of contracts (default 1)
	Multiplier int        // contract multiplier (default 100 for US equity)
}

// DefaultMultiplier is the typical US equity options multiplier.
const DefaultMultiplier = 100

// PayoffAt returns the P&L per share at expiration for this leg at spot S.
func (l Leg) PayoffAt(S float64) float64 {
	mult := l.Multiplier
	if mult <= 0 {
		mult = DefaultMultiplier
	}
	n := l.Contracts
	if n <= 0 {
		n = 1
	}
	perShare := l.payoffPerShare(S)
	return perShare * float64(n*mult)
}

func (l Leg) payoffPerShare(S float64) float64 {
	var intrinsic float64
	switch l.Type {
	case Call:
		if S > l.Strike {
			intrinsic = S - l.Strike
		}
	case Put:
		if S < l.Strike {
			intrinsic = l.Strike - S
		}
	}
	switch l.Side {
	case Long:
		return intrinsic - l.Premium
	case Short:
		return l.Premium - intrinsic
	}
	return 0
}

// Strategy is a slice of legs; total payoff is the sum of leg payoffs.
type Strategy []Leg

// PayoffAt returns total P&L at expiration at spot S.
func (s Strategy) PayoffAt(S float64) float64 {
	var total float64
	for _, leg := range s {
		total += leg.PayoffAt(S)
	}
	return total
}

// PayoffSeries computes payoff over a range of spot prices.
// spotMin, spotMax: range of underlying price
// numPoints: number of points to sample (e.g. 100)
func (s Strategy) PayoffSeries(spotMin, spotMax float64, numPoints int) (prices []float64, payoffs []float64) {
	if numPoints < 2 {
		numPoints = 2
	}
	step := (spotMax - spotMin) / float64(numPoints-1)
	prices = make([]float64, numPoints)
	payoffs = make([]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		S := spotMin + float64(i)*step
		prices[i] = S
		payoffs[i] = s.PayoffAt(S)
	}
	return prices, payoffs
}

// PayoffStats holds summary stats for the strategy at expiry over the given spot range.
type PayoffStats struct {
	MaxProfit  float64   // max payoff in range
	MaxLoss    float64   // min payoff in range
	RewardRisk float64   // MaxProfit / |MaxLoss| when MaxLoss < 0, else 0 or Inf
	Breakevens []float64 // spot prices where payoff crosses zero (linear interpolation)
}

// Stats computes PayoffStats over [spotMin, spotMax] using numPoints samples.
func (s Strategy) Stats(spotMin, spotMax float64, numPoints int) PayoffStats {
	prices, payoffs := s.PayoffSeries(spotMin, spotMax, numPoints)
	var maxP, minP float64
	if len(payoffs) > 0 {
		maxP, minP = payoffs[0], payoffs[0]
	}
	for _, p := range payoffs {
		if p > maxP {
			maxP = p
		}
		if p < minP {
			minP = p
		}
	}
	rewardRisk := 0.0
	if minP < 0 && maxP > 0 {
		rewardRisk = maxP / (-minP)
	} else if minP >= 0 {
		rewardRisk = 0 // no risk
	}
	// breakevens: crossings of zero (linear interpolation between points)
	var breakevens []float64
	for i := 1; i < len(payoffs); i++ {
		p0, p1 := payoffs[i-1], payoffs[i]
		if (p0 < 0 && p1 >= 0) || (p0 >= 0 && p1 < 0) {
			s0, s1 := prices[i-1], prices[i]
			if p1 != p0 {
				t := (0 - p0) / (p1 - p0)
				breakevens = append(breakevens, s0+t*(s1-s0))
			}
		}
	}
	return PayoffStats{MaxProfit: maxP, MaxLoss: minP, RewardRisk: rewardRisk, Breakevens: breakevens}
}

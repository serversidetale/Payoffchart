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
	Type      OptionType // "call" or "put"
	Side      Side       // "long" or "short"
	Strike    float64    // strike price
	Premium   float64    // option premium per share (paid/received)
	Contracts int        // number of contracts (default 1)
	Multiplier int       // contract multiplier (default 100 for US equity)
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

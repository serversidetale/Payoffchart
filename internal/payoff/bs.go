package payoff

import (
	"math"
)

// normCDF is the standard normal cumulative distribution function.
// N(x) = (1 + erf(x/sqrt(2))) / 2
func normCDF(x float64) float64 {
	return 0.5 * (1 + math.Erf(x/math.Sqrt(2)))
}

// BSCall returns Black-Scholes European call value (per share).
// S=spot, K=strike, T=time to expiry in years, sigma=volatility (e.g. 0.2 for 20%), r=risk-free rate (e.g. 0.05).
func BSCall(S, K, T, sigma, r float64) float64 {
	if T <= 0 {
		return math.Max(S-K, 0)
	}
	if sigma <= 0 {
		return math.Max(S-K*math.Exp(-r*T), 0)
	}
	sqrtT := math.Sqrt(T)
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * sqrtT)
	d2 := d1 - sigma*sqrtT
	return S*normCDF(d1) - K*math.Exp(-r*T)*normCDF(d2)
}

// BSPut returns Black-Scholes European put value (per share).
func BSPut(S, K, T, sigma, r float64) float64 {
	if T <= 0 {
		return math.Max(K-S, 0)
	}
	if sigma <= 0 {
		return math.Max(K*math.Exp(-r*T)-S, 0)
	}
	sqrtT := math.Sqrt(T)
	d1 := (math.Log(S/K) + (r+0.5*sigma*sigma)*T) / (sigma * sqrtT)
	d2 := d1 - sigma*sqrtT
	return K*math.Exp(-r*T)*normCDF(-d2) - S*normCDF(-d1)
}

// ValueAt returns the P&L for this leg at spot S with time T (years) to expiry,
// using Black-Scholes with the given volatility and risk-free rate.
// For T <= 0, uses expiration payoff (intrinsic only).
func (l Leg) ValueAt(S, T, sigma, r float64) float64 {
	mult := l.Multiplier
	if mult <= 0 {
		mult = DefaultMultiplier
	}
	n := l.Contracts
	if n <= 0 {
		n = 1
	}
	var value float64
	if T <= 0 {
		value = l.payoffPerShare(S)
	} else {
		var optValue float64
		switch l.Type {
		case Call:
			optValue = BSCall(S, l.Strike, T, sigma, r)
		case Put:
			optValue = BSPut(S, l.Strike, T, sigma, r)
		default:
			optValue = 0
		}
		switch l.Side {
		case Long:
			value = optValue - l.Premium
		case Short:
			value = l.Premium - optValue
		default:
			value = 0
		}
	}
	return value * float64(n*mult)
}

// PayoffAtBeforeExpiry returns total P&L at spot S with T years to expiry (BS valuation).
func (s Strategy) PayoffAtBeforeExpiry(S, T, sigma, r float64) float64 {
	var total float64
	for _, leg := range s {
		total += leg.ValueAt(S, T, sigma, r)
	}
	return total
}

// PayoffSeriesBeforeExpiry computes P&L over spot range using Black-Scholes.
// T in years (e.g. 30/365), sigma as decimal (e.g. 0.2), r as decimal (e.g. 0.05).
func (s Strategy) PayoffSeriesBeforeExpiry(spotMin, spotMax float64, T, sigma, r float64, numPoints int) (prices []float64, payoffs []float64) {
	if numPoints < 2 {
		numPoints = 2
	}
	step := (spotMax - spotMin) / float64(numPoints-1)
	prices = make([]float64, numPoints)
	payoffs = make([]float64, numPoints)
	for i := 0; i < numPoints; i++ {
		S := spotMin + float64(i)*step
		prices[i] = S
		payoffs[i] = s.PayoffAtBeforeExpiry(S, T, sigma, r)
	}
	return prices, payoffs
}

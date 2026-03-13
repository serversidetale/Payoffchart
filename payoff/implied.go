package payoff

import (
	"errors"
	"math"
)

// BSCallDiv returns Black-Scholes European call value with dividend yield q (per share).
// q = dividend yield as decimal (e.g. 0.02 for 2%).
func BSCallDiv(S, K, T, sigma, r, q float64) float64 {
	if T <= 0 {
		return math.Max(S*math.Exp(-q*T)-K, 0)
	}
	if sigma <= 0 {
		return math.Max(S*math.Exp(-q*T)-K*math.Exp(-r*T), 0)
	}
	sqrtT := math.Sqrt(T)
	d1 := (math.Log(S/K) + (r-q+0.5*sigma*sigma)*T) / (sigma * sqrtT)
	d2 := d1 - sigma*sqrtT
	return S*math.Exp(-q*T)*normCDF(d1) - K*math.Exp(-r*T)*normCDF(d2)
}

// BSPutDiv returns Black-Scholes European put value with dividend yield q (per share).
func BSPutDiv(S, K, T, sigma, r, q float64) float64 {
	if T <= 0 {
		return math.Max(K-S, 0)
	}
	if sigma <= 0 {
		return math.Max(K*math.Exp(-r*T)-S*math.Exp(-q*T), 0)
	}
	sqrtT := math.Sqrt(T)
	d1 := (math.Log(S/K) + (r-q+0.5*sigma*sigma)*T) / (sigma * sqrtT)
	d2 := d1 - sigma*sqrtT
	return K*math.Exp(-r*T)*normCDF(-d2) - S*math.Exp(-q*T)*normCDF(-d1)
}

// RFromParity derives risk-free rate r from put-call parity using call and put premiums.
// C - P = S*exp(-q*T) - K*exp(-r*T). Solves for r.
// T in years, q = dividend yield as decimal. Returns r as decimal (e.g. 0.05 for 5%).
func RFromParity(S, K, T, q, callPremium, putPremium float64) (r float64, err error) {
	if T <= 0 {
		return 0, errors.New("time to expiry must be positive")
	}
	if K <= 0 {
		return 0, errors.New("strike must be positive")
	}
	discSpot := S * math.Exp(-q*T)
	diff := discSpot - (callPremium - putPremium)
	if diff <= 0 {
		return 0, errors.New("put-call parity: S*exp(-q*T) - (C-P) must be positive (check premiums)")
	}
	ratio := diff / K
	if ratio <= 0 {
		return 0, errors.New("put-call parity: invalid ratio")
	}
	r = -math.Log(ratio) / T
	return r, nil
}

// ImpliedVolCall finds sigma such that BSCallDiv(S, K, T, sigma, r, q) = callPremium.
// Uses bisection. Returns sigma as decimal (e.g. 0.2 for 20%).
func ImpliedVolCall(S, K, T, r, q, callPremium float64) (sigma float64, err error) {
	if T <= 0 {
		return 0, errors.New("time to expiry must be positive")
	}
	intrinsic := math.Max(S*math.Exp(-q*T)-K*math.Exp(-r*T), 0)
	if callPremium < intrinsic {
		return 0, errors.New("call premium below intrinsic value")
	}
	if callPremium <= 0 {
		return 0, errors.New("call premium must be positive")
	}
	low, high := 1e-6, 10.0
	const maxIter = 100
	for i := 0; i < maxIter; i++ {
		mid := (low + high) / 2
		price := BSCallDiv(S, K, T, mid, r, q)
		if math.Abs(price-callPremium) < 1e-10 {
			return mid, nil
		}
		if price < callPremium {
			low = mid
		} else {
			high = mid
		}
	}
	return (low + high) / 2, nil
}

// DeriveRAndSigma computes implied r from put-call parity and implied σ from the call, given call & put premiums.
// S, K = spot and strike; days = days to expiry; q = dividend yield (decimal); callPremium, putPremium = per share.
// Returns r and sigma as decimals (e.g. 0.05, 0.2).
func DeriveRAndSigma(S, K float64, days int, q, callPremium, putPremium float64) (r, sigma float64, err error) {
	if days <= 0 {
		return 0, 0, errors.New("days to expiry must be positive")
	}
	T := float64(days) / 365.0
	r, err = RFromParity(S, K, T, q, callPremium, putPremium)
	if err != nil {
		return 0, 0, err
	}
	sigma, err = ImpliedVolCall(S, K, T, r, q, callPremium)
	if err != nil {
		return r, 0, err
	}
	return r, sigma, nil
}

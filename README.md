# Options Payoff Chart

Generate **options payoff charts** (P&L at expiration vs underlying price) using [go-echarts](https://github.com/go-echarts/go-echarts). Output is a single HTML file you can open in a browser.

## Quick start

**CLI** (writes a static HTML file):

```bash
go run ./cmd/payoff
```

Opens `payoff.html` in a browser. Default example: bull call spread.

**Server** (web form → payoff chart):

```bash
go run ./cmd/payoffserver
```

Then open **http://localhost:8080**. Enter chart title, spot range (min–max), **days to expiry** (0 = at-expiry only; &gt;0 adds a “Before expiry” curve via Black-Scholes), **volatility %** and **risk-free rate %** (used when days to expiry &gt; 0). Use **+ Add leg** / **Remove** to add or remove legs (no fixed limit). Click **Generate Payoff Chart** to get the chart.

**Derive σ and r:** Open **http://localhost:8080/derive**. Enter Spot, Strike, Days to expiry, Call premium, Put premium, and Dividend yield % (same strike). Submit to get implied **r** and **σ** from put-call parity and Black-Scholes; use these in the payoff chart for Volatility % and Risk-free rate %.

## Before-expiry P&L

If you set **days to expiry** &gt; 0 (and volatility + rate), the chart shows two curves:

- **At expiry** — intrinsic value only (same as before).
- **Before expiry (today)** — Black-Scholes value at that time to expiry; P&L = (option value − premium paid) × position.

Volatility and risk-free rate are used only for the before-expiry curve.

## Inputs (per leg)

Each option leg is defined by:

| Input        | Type   | Meaning                                      |
|-------------|--------|----------------------------------------------|
| **Type**    | string | `"call"` or `"put"`                          |
| **Side**    | string | `"long"` or `"short"`                        |
| **Strike**  | float  | Strike price \(K\)                           |
| **Premium** | float  | Option premium **per share** (paid/received) |
| **Contracts** | int  | Number of contracts (default 1)              |
| **Multiplier** / **Contract size** | int | Shares per contract (default 100 for US equity; configurable per leg in the server form)   |

**Chart range:** you also choose the underlying price range for the X-axis (`spotMin`, `spotMax`).

Payoff is **at expiration** only (intrinsic value). No Greeks or time value.

## Defining a strategy in code

Edit `cmd/payoff/main.go` and build a `payoff.Strategy` (slice of legs):

```go
strategy := payoff.Strategy{
    // Long 1 contract of 100-strike call, paid 3 per share
    {Type: payoff.Call, Side: payoff.Long, Strike: 100, Premium: 3.0, Contracts: 1, Multiplier: 100},
    // Short 1 contract of 110-strike call, received 1 per share
    {Type: payoff.Call, Side: payoff.Short, Strike: 110, Premium: 1.0, Contracts: 1, Multiplier: 100},
}

spotMin, spotMax := 90.0, 120.0
title := "Bull Call Spread"
// ... then call chart.RenderPayoff(f, strategy, spotMin, spotMax, title)
```

**Single options:**

- **Long call:** `{Type: payoff.Call, Side: payoff.Long, Strike: 100, Premium: 5.0}`
- **Short put:** `{Type: payoff.Put, Side: payoff.Short, Strike: 95, Premium: 2.0}`

**Spreads / multi-leg:** add more legs to the same `Strategy` slice.

## Project layout

```
PayOffChart/
├── cmd/
│   ├── payoff/main.go         # CLI: builds strategy, writes payoff.html
│   └── payoffserver/main.go   # HTTP server: form at /, chart at POST /chart
├── internal/
│   ├── payoff/leg.go          # Leg types and payoff math
│   └── chart/chart.go         # go-echarts line chart from payoff data
├── go.mod
└── README.md
```

## Payoff formulas (per share, at expiration)

- **Long call:**  \(\max(S - K, 0) - \text{premium}\)
- **Short call:** \(-\max(S - K, 0) + \text{premium}\)
- **Long put:**   \(\max(K - S, 0) - \text{premium}\)
- **Short put:**  \(-\max(K - S, 0) + \text{premium}\)

Total P&L at each price \(S\) is the sum over all legs, × contracts × multiplier.

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

## API (server endpoints)

The payoff server exposes **4 routes** (HTML form endpoints + one JSON API). All responses include **CORS headers** (`Access-Control-Allow-Origin: *`, `Access-Control-Allow-Methods`, `Access-Control-Allow-Headers`) so the API can be called from browser apps on other origins (e.g. `abc.dingalola.in` calling `xyz.dingalola.in`).

| Route       | Method | Purpose |
|------------|--------|--------|
| **`/`**    | GET    | Serves the main payoff chart form (strategy, legs, spot range, vol, rate, days to expiry). |
| **`/chart`** | POST | Accepts form-encoded data; returns the payoff chart HTML (with stats and optional “change days” form). GET redirects to `/`. |
| **`/derive`** | GET  | Serves the “derive σ and r” form (spot, strike, days, call/put premiums, dividend). |
| **`/derive`** | POST | Submits the form; returns the same page with implied **r** and **σ** (or an error). |
| **`/api/chart`** | POST | **JSON API.** Accepts a JSON body with legs and chart params; optionally include `derive` to compute vol_pct and rate_pct from call/put premiums. Returns the same chart HTML. For browser cross-origin use. |

### Using the API with cURL

Assume the server is running at `http://localhost:8080`.

**Get the payoff chart form (HTML):**
```bash
curl -s http://localhost:8080/
```

**Generate a payoff chart (POST form; returns HTML with chart):**

One leg: long call, strike 100, premium 3, 1 contract, contract size 100. Spot range 90–120, at-expiry only.
```bash
curl -s -X POST http://localhost:8080/chart \
  -d "title=Bull+Call" \
  -d "spot_min=90" \
  -d "spot_max=120" \
  -d "days_to_expiry=0" \
  -d "vol_pct=20" \
  -d "rate_pct=5" \
  -d "num_legs=1" \
  -d "leg_type_0=call" \
  -d "leg_side_0=long" \
  -d "leg_strike_0=100" \
  -d "leg_premium_0=3" \
  -d "leg_contracts_0=1" \
  -d "leg_multiplier_0=100" \
  -o payoff_chart.html
```
Open `payoff_chart.html` in a browser to view the chart.

**Two legs (e.g. bull call spread):** set `num_legs=2` and add `leg_type_1`, `leg_side_1`, `leg_strike_1`, `leg_premium_1`, `leg_contracts_1`, `leg_multiplier_1`. For “before expiry” curve, set `days_to_expiry` &gt; 0 (e.g. `days_to_expiry=30`).

**Get the derive σ and r form (HTML):**
```bash
curl -s http://localhost:8080/derive
```

**Derive implied r and σ from call and put premiums (POST form; returns HTML with results):**
```bash
curl -s -X POST http://localhost:8080/derive \
  -d "spot=1950" \
  -d "strike=1950" \
  -d "days=8" \
  -d "call_premium=130" \
  -d "put_premium=80" \
  -d "div_pct=0"
```
The response HTML includes the computed **Implied risk-free rate r** and **Implied volatility σ** (or an error message).

### JSON API: POST /api/chart

For browser or server clients. Send **`Content-Type: application/json`** and a JSON body.

**Request body:**

| Field | Type | Description |
|-------|------|-------------|
| `title` | string | Chart title (optional) |
| `spot_min`, `spot_max` | number | Underlying price range for the X-axis |
| `days_to_expiry` | number | 0 = at-expiry only; &gt;0 adds “before expiry” curve |
| `vol_pct`, `rate_pct` | number | Volatility % and risk-free rate % (used when `days_to_expiry` &gt; 0) |
| `legs` | array | Each element: `type` ("call"/"put"), `side` ("long"/"short"), `strike`, `premium`, `contracts` (default 1), `multiplier` (default 100) |
| `derive` | object (optional) | If present and `vol_pct`/`rate_pct` are not set, the server derives them from call/put premiums. Fields: `spot`, `strike`, `days`, `call_premium`, `put_premium`, `div_pct` |

**Response:** Same chart HTML as `POST /chart` (with stats block). On validation error (e.g. no valid legs), returns `400` with JSON `{"error": "..."}`.

**cURL example (explicit vol/rate):**
```bash
curl -s -X POST http://localhost:8080/api/chart \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Bull Call",
    "spot_min": 90,
    "spot_max": 120,
    "days_to_expiry": 0,
    "vol_pct": 20,
    "rate_pct": 5,
    "legs": [
      { "type": "call", "side": "long", "strike": 100, "premium": 3, "contracts": 1, "multiplier": 100 }
    ]
  }' -o chart.html
```

**cURL example (derive vol and rate from call/put):**
```bash
curl -s -X POST http://localhost:8080/api/chart \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Strategy",
    "spot_min": 1900,
    "spot_max": 2100,
    "days_to_expiry": 8,
    "legs": [
      { "type": "call", "side": "short", "strike": 1950, "premium": 130, "contracts": 1, "multiplier": 100 }
    ],
    "derive": { "spot": 1950, "strike": 1950, "days": 8, "call_premium": 130, "put_premium": 80, "div_pct": 0 }
  }' -o chart.html
```
Then open `chart.html` in a browser. The server will compute implied **r** and **σ** from the call/put premiums and use them for the before-expiry curve.

**CORS:** All endpoints send `Access-Control-Allow-Origin: *` and support `OPTIONS` preflight, so a front-end on another origin (e.g. `https://abc.dingalola.in`) can call `https://xyz.dingalola.in/api/chart` from the browser.

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

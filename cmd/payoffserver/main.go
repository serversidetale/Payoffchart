package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/serversidetale/payoffchart/chart"
	"github.com/serversidetale/payoffchart/payoff"
)

const formHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Options Payoff Chart</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 720px; margin: 2rem auto; padding: 0 1rem; }
    h1 { font-size: 1.5rem; }
    label { display: inline-block; min-width: 5rem; margin-bottom: 0.25rem; }
    input, select { padding: 0.35rem; margin-bottom: 0.5rem; }
    .row { display: flex; gap: 0.5rem; flex-wrap: wrap; align-items: flex-end; margin-bottom: 0.5rem; }
    .row label { min-width: 0; }
    .legs { margin: 1rem 0; }
    .leg { background: #f5f5f5; padding: 0.75rem; margin-bottom: 0.5rem; border-radius: 6px; }
    .leg h3 { margin: 0 0 0.5rem 0; font-size: 0.95rem; display: flex; justify-content: space-between; align-items: center; }
    .leg-actions { margin-top: 0.5rem; }
    button { padding: 0.5rem 1rem; background: #333; color: #fff; border: none; border-radius: 6px; cursor: pointer; }
    button:hover { background: #555; }
    button.remove-leg { background: #c00; font-size: 0.85rem; padding: 0.25rem 0.5rem; }
    button.remove-leg:hover { background: #e00; }
    .add-leg { background: #066; margin-top: 0.5rem; }
    .add-leg:hover { background: #088; }
    .hint { font-size: 0.85rem; color: #666; margin-top: 0.25rem; }
    .params { display: grid; gap: 0.5rem; margin: 1rem 0; }
  </style>
</head>
<body>
  <h1>Options Payoff Chart</h1>
  <p><a href="/derive">Derive σ and r</a> from call &amp; put premiums (same strike).</p>
  <p>Add/remove legs as needed. Premium is per share. Optionally show P&L before expiry (Black-Scholes).</p>
  <form action="/chart" method="POST" id="payoffForm">
    <div>
      <label>Chart title</label><br>
      <input type="text" name="title" value="Options Payoff" size="40" placeholder="e.g. Bull Call Spread">
    </div>
    <div class="params">
      <div>
        <label>Spot range</label><br>
        <input type="number" name="spot_min" step="any" value="90" placeholder="Min"> to
        <input type="number" name="spot_max" step="any" value="120" placeholder="Max">
      </div>
      <div>
        <label>Days to expiry</label><br>
        <input type="number" name="days_to_expiry" value="0" min="0" placeholder="0 = at expiry only">
        <span class="hint">If &gt; 0, adds a &quot;Before expiry&quot; curve (Black-Scholes).</span>
      </div>
      <div>
        <label>Volatility %</label><br>
        <input type="number" name="vol_pct" step="any" value="20" min="0" placeholder="e.g. 20 for 20%">
      </div>
      <div>
        <label>Risk-free rate %</label><br>
        <input type="number" name="rate_pct" step="any" value="5" min="0" placeholder="e.g. 5 for 5%">
      </div>
    </div>
    <div class="legs">
      <h2 style="font-size: 1.1rem;">Legs</h2>
      <div id="legsContainer"></div>
      <input type="hidden" name="num_legs" id="num_legs" value="1">
      <button type="button" class="add-leg" id="addLeg">+ Add leg</button>
    </div>
    <button type="submit">Generate Payoff Chart</button>
  </form>

  <template id="legTemplate">
    <div class="leg" data-leg-index="__IDX__">
      <h3>Leg <span class="leg-num"></span> <button type="button" class="remove-leg" aria-label="Remove leg">Remove</button></h3>
      <div class="row">
        <label>Type</label>
        <select name="leg_type___IDX__">
          <option value="call">Call</option>
          <option value="put">Put</option>
        </select>
        <label>Side</label>
        <select name="leg_side___IDX__">
          <option value="long">Long</option>
          <option value="short">Short</option>
        </select>
        <label>Strike</label>
        <input type="number" name="leg_strike___IDX__" step="any" placeholder="e.g. 100">
        <label>Premium</label>
        <input type="number" name="leg_premium___IDX__" step="any" placeholder="per share">
        <label>Contracts</label>
        <input type="number" name="leg_contracts___IDX__" value="1" min="1">
        <label>Contract size</label>
        <input type="number" name="leg_multiplier___IDX__" value="100" min="1" title="Shares per contract (e.g. 100 US equity)">
      </div>
    </div>
  </template>

  <script>
    (function() {
      var legCount = 1;
      var container = document.getElementById('legsContainer');
      var tpl = document.getElementById('legTemplate');
      var numLegsInput = document.getElementById('num_legs');

      function addLeg(index) {
        var html = tpl.innerHTML
          .replace(/__IDX__/g, String(index));
        var div = document.createElement('div');
        div.innerHTML = html.trim();
        var legEl = div.firstChild;
        legEl.setAttribute('data-leg-index', index);
        legEl.querySelector('.leg-num').textContent = index + 1;
        legEl.querySelector('.remove-leg').addEventListener('click', function() {
          if (legCount <= 1) return;
          legEl.remove();
          legCount--;
          numLegsInput.value = legCount;
          renumberLegs();
        });
        container.appendChild(legEl);
      }

      function renumberLegs() {
        var legs = container.querySelectorAll('.leg');
        legs.forEach(function(leg, i) {
          leg.setAttribute('data-leg-index', i);
          leg.querySelector('.leg-num').textContent = i + 1;
          leg.querySelectorAll('[name]').forEach(function(input) {
            input.name = input.name.replace(/_\d+$/, '_' + i);
          });
        });
      }

      document.getElementById('addLeg').addEventListener('click', function() {
        addLeg(legCount);
        legCount++;
        numLegsInput.value = legCount;
      });

      addLeg(0);
    })();
  </script>
</body>
</html>
`

const deriveFormHTML = `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>Derive σ and r</title>
  <style>
    body { font-family: system-ui, sans-serif; max-width: 520px; margin: 2rem auto; padding: 0 1rem; }
    h1 { font-size: 1.35rem; }
    label { display: inline-block; min-width: 10rem; margin-bottom: 0.25rem; }
    input { padding: 0.35rem; margin-bottom: 0.5rem; }
    .row { margin-bottom: 0.75rem; }
    button { padding: 0.5rem 1rem; background: #333; color: #fff; border: none; border-radius: 6px; cursor: pointer; }
    button:hover { background: #555; }
    .hint { font-size: 0.85rem; color: #666; margin-top: 0.2rem; }
    .result { background: #e8f5e9; padding: 1rem; border-radius: 6px; margin: 1rem 0; }
    .error { background: #ffebee; padding: 1rem; border-radius: 6px; margin: 1rem 0; }
    a { color: #066; }
  </style>
</head>
<body>
  <h1>Derive σ and r from call &amp; put</h1>
  <p>Enter spot, strike, days to expiry, call premium, put premium, and dividend yield. Same S, K, T for both options.</p>
  <p><a href="/">← Payoff chart</a></p>
  <form action="/derive" method="POST">
    <div class="row">
      <label>Spot (S)</label><br>
      <input type="number" name="spot" step="any" required placeholder="e.g. 1950">
    </div>
    <div class="row">
      <label>Strike (K)</label><br>
      <input type="number" name="strike" step="any" required placeholder="e.g. 1950">
    </div>
    <div class="row">
      <label>Days to expiry</label><br>
      <input type="number" name="days" min="1" required placeholder="e.g. 8">
    </div>
    <div class="row">
      <label>Call premium (per share)</label><br>
      <input type="number" name="call_premium" step="any" required placeholder="e.g. 130">
    </div>
    <div class="row">
      <label>Put premium (per share)</label><br>
      <input type="number" name="put_premium" step="any" required placeholder="e.g. 80">
    </div>
    <div class="row">
      <label>Dividend yield %</label><br>
      <input type="number" name="div_pct" step="any" value="0" min="0" placeholder="0">
      <span class="hint">e.g. 0 or 1.5 for 1.5%</span>
    </div>
    <button type="submit">Derive σ and r</button>
  </form>
</body>
</html>
`

// cors wraps a handler to add CORS headers and handle OPTIONS preflight (for browser cross-origin calls).
func cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h(w, r)
	}
}

func main() {
	http.HandleFunc("/", cors(handleForm))
	http.HandleFunc("/chart", cors(handleChart))
	http.HandleFunc("/derive", cors(handleDerive))
	http.HandleFunc("/api/chart", cors(handleChartJSON))
	addr := ":8080"
	log.Printf("Payoff chart server at http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleForm(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	t := template.Must(template.New("form").Parse(formHTML))
	if err := t.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func handleDerive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method != http.MethodPost {
		_, _ = w.Write([]byte(deriveFormHTML))
		return
	}
	if err := r.ParseForm(); err != nil {
		_, _ = w.Write([]byte(deriveFormHTML))
		fmt.Fprintf(w, `<div class="error">Invalid form.</div>`)
		return
	}
	S, _ := strconv.ParseFloat(r.FormValue("spot"), 64)
	K, _ := strconv.ParseFloat(r.FormValue("strike"), 64)
	days, _ := strconv.Atoi(r.FormValue("days"))
	callPrem, _ := strconv.ParseFloat(r.FormValue("call_premium"), 64)
	putPrem, _ := strconv.ParseFloat(r.FormValue("put_premium"), 64)
	divPct, _ := strconv.ParseFloat(r.FormValue("div_pct"), 64)
	q := divPct / 100

	_, _ = w.Write([]byte(deriveFormHTML))
	rPct, sigmaPct, err := payoff.DeriveRAndSigma(S, K, days, q, callPrem, putPrem)
	if err != nil {
		fmt.Fprintf(w, `<div class="error">%s</div>`, template.HTMLEscapeString(err.Error()))
		return
	}
	fmt.Fprintf(w, `<div class="result"><strong>Implied risk-free rate r</strong> = %.4f%% &nbsp; <strong>Implied volatility σ</strong> = %.4f%%<br><span class="hint">Use these in the <a href="/">payoff chart</a> for Volatility %% and Risk-free rate %%.</span></div>`, rPct*100, sigmaPct*100)
}

// JSON API request types for POST /api/chart
type legInput struct {
	Type       string  `json:"type"`       // "call" or "put"
	Side       string  `json:"side"`       // "long" or "short"
	Strike     float64 `json:"strike"`
	Premium    float64 `json:"premium"`
	Contracts  int     `json:"contracts"`
	Multiplier int     `json:"multiplier"`
}

type deriveInput struct {
	Spot         float64 `json:"spot"`
	Strike       float64 `json:"strike"`
	Days         int     `json:"days"`
	CallPremium  float64 `json:"call_premium"`
	PutPremium   float64 `json:"put_premium"`
	DivPct       float64 `json:"div_pct"`
}

type chartJSONRequest struct {
	Title         string       `json:"title"`
	SpotMin       float64      `json:"spot_min"`
	SpotMax       float64      `json:"spot_max"`
	DaysToExpiry  int          `json:"days_to_expiry"`
	VolPct        float64      `json:"vol_pct"`
	RatePct       float64      `json:"rate_pct"`
	Legs          []legInput   `json:"legs"`
	Derive        *deriveInput `json:"derive,omitempty"` // optional: use to derive vol_pct and rate_pct
}

func handleChartJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req chartJSONRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	title := req.Title
	if title == "" {
		title = "Options Payoff"
	}
	spotMin, spotMax := req.SpotMin, req.SpotMax
	if spotMin >= spotMax {
		spotMin, spotMax = 90, 120
	}
	daysToExpiry := req.DaysToExpiry
	volPct, ratePct := req.VolPct, req.RatePct
	// If vol/rate not set and derive is provided, derive them from call/put premiums
	if req.Derive != nil && req.Derive.Days > 0 {
		rDec, sigmaDec, err := payoff.DeriveRAndSigma(
			req.Derive.Spot, req.Derive.Strike, req.Derive.Days,
			req.Derive.DivPct/100, req.Derive.CallPremium, req.Derive.PutPremium,
		)
		if err == nil {
			volPct = sigmaDec * 100
			ratePct = rDec * 100
		}
	}
	vol := volPct / 100
	rate := ratePct / 100

	var strategy payoff.Strategy
	for _, l := range req.Legs {
		if l.Strike <= 0 {
			continue
		}
		contracts := l.Contracts
		if contracts <= 0 {
			contracts = 1
		}
		mult := l.Multiplier
		if mult <= 0 {
			mult = payoff.DefaultMultiplier
		}
		leg := payoff.Leg{
			Strike:     l.Strike,
			Premium:    l.Premium,
			Contracts:  contracts,
			Multiplier: mult,
		}
		if l.Type == "put" {
			leg.Type = payoff.Put
		} else {
			leg.Type = payoff.Call
		}
		if l.Side == "short" {
			leg.Side = payoff.Short
		} else {
			leg.Side = payoff.Long
		}
		strategy = append(strategy, leg)
	}
	if len(strategy) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "no valid legs (need at least one leg with strike > 0)"})
		return
	}

	renderOpts := (*chart.RenderOpts)(nil)
	if daysToExpiry > 0 {
		renderOpts = &chart.RenderOpts{
			DaysToExpiry: daysToExpiry,
			Volatility:   vol,
			RiskFreeRate: rate,
		}
	}
	var chartBuf bytes.Buffer
	if err := chart.RenderPayoff(&chartBuf, strategy, spotMin, spotMax, title, renderOpts); err != nil {
		log.Printf("render: %v", err)
		http.Error(w, "failed to render chart", http.StatusInternalServerError)
		return
	}
	chartHTML := chartBuf.String()
	stats := strategy.Stats(spotMin, spotMax, 500)
	statsHTML := buildStatsHTML(stats)
	if idx := strings.LastIndex(chartHTML, "</body>"); idx != -1 {
		chartHTML = chartHTML[:idx] + statsHTML + "\n" + chartHTML[idx:]
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(chartHTML))
}

func handleChart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	if title == "" {
		title = "Options Payoff"
	}
	spotMin, _ := strconv.ParseFloat(r.FormValue("spot_min"), 64)
	spotMax, _ := strconv.ParseFloat(r.FormValue("spot_max"), 64)
	if spotMin >= spotMax {
		spotMin, spotMax = 90, 120
	}

	daysToExpiry, _ := strconv.Atoi(r.FormValue("days_to_expiry"))
	volPct, _ := strconv.ParseFloat(r.FormValue("vol_pct"), 64)
	ratePct, _ := strconv.ParseFloat(r.FormValue("rate_pct"), 64)
	vol := volPct / 100
	rate := ratePct / 100

	numLegs, _ := strconv.Atoi(r.FormValue("num_legs"))
	if numLegs <= 0 {
		numLegs = 1
	}

	// formValue gets a form value, trying numeric index first, then literal INDEX/index (some clients send unchanged placeholder)
	formValue := func(key string, i int) string {
		if s := r.FormValue(fmt.Sprintf("%s_%d", key, i)); s != "" {
			return s
		}
		if i == 0 {
			if s := r.FormValue(key + "_INDEX"); s != "" {
				return s
			}
			return r.FormValue(key + "_index")
		}
		return ""
	}

	var strategy payoff.Strategy
	for i := 0; i < numLegs; i++ {
		strikeStr := formValue("leg_strike", i)
		if strikeStr == "" {
			continue
		}
		strike, err := strconv.ParseFloat(strikeStr, 64)
		if err != nil || strike <= 0 {
			continue
		}
		premium, _ := strconv.ParseFloat(formValue("leg_premium", i), 64)
		contracts, _ := strconv.Atoi(formValue("leg_contracts", i))
		if contracts <= 0 {
			contracts = 1
		}
		typeStr := formValue("leg_type", i)
		sideStr := formValue("leg_side", i)
		multiplier, _ := strconv.Atoi(formValue("leg_multiplier", i))
		if multiplier <= 0 {
			multiplier = payoff.DefaultMultiplier
		}
		leg := payoff.Leg{
			Strike:     strike,
			Premium:    premium,
			Contracts:  contracts,
			Multiplier: multiplier,
		}
		if typeStr == "put" {
			leg.Type = payoff.Put
		} else {
			leg.Type = payoff.Call
		}
		if sideStr == "short" {
			leg.Side = payoff.Short
		} else {
			leg.Side = payoff.Long
		}
		strategy = append(strategy, leg)
	}

	if len(strategy) == 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `<p>No legs with a valid strike. <a href="/">Go back</a> and fill at least one leg (e.g. Strike 100, Premium 3).</p>`)
		return
	}

	renderOpts := (*chart.RenderOpts)(nil)
	if daysToExpiry > 0 {
		renderOpts = &chart.RenderOpts{
			DaysToExpiry: daysToExpiry,
			Volatility:   vol,
			RiskFreeRate: rate,
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var chartBuf bytes.Buffer
	if err := chart.RenderPayoff(&chartBuf, strategy, spotMin, spotMax, title, renderOpts); err != nil {
		log.Printf("render: %v", err)
		http.Error(w, "failed to render chart", http.StatusInternalServerError)
		return
	}
	chartHTML := chartBuf.String()

	// Payoff stats (at expiry)
	stats := strategy.Stats(spotMin, spotMax, 500)
	statsHTML := buildStatsHTML(stats)

	// Optional: form to change days to expiry when before-expiry curve is shown
	if daysToExpiry > 0 {
		var formBuf strings.Builder
		formBuf.WriteString(`<div style="margin:1rem 20px; padding:0.75rem; background:#f0f4f8; border-radius:6px; font-family:system-ui,sans-serif;"><form method="POST" action="/chart">`)
		formBuf.WriteString(`<label>Days to expiry: <input type="number" name="days_to_expiry" value="` + strconv.Itoa(daysToExpiry) + `" min="1" style="width:4rem; padding:0.25rem;">`)
		formBuf.WriteString(`</label> <button type="submit">Update before-expiry curve</button>`)
		for key, vals := range r.Form {
			if key == "days_to_expiry" {
				continue
			}
			for _, v := range vals {
				formBuf.WriteString(`<input type="hidden" name="` + html.EscapeString(key) + `" value="` + html.EscapeString(v) + `">`)
			}
		}
		formBuf.WriteString(`</form></div>`)
		statsHTML = statsHTML + "\n" + formBuf.String()
	}

	if idx := strings.LastIndex(chartHTML, "</body>"); idx != -1 {
		chartHTML = chartHTML[:idx] + statsHTML + "\n" + chartHTML[idx:]
	}
	if _, err := w.Write([]byte(chartHTML)); err != nil {
		log.Printf("write: %v", err)
	}
}

func buildStatsHTML(stats payoff.PayoffStats) string {
	formatNum := func(x float64) string {
		if x >= 1000 || (x > -1000 && x < 1000 && x != 0) {
			return strconv.FormatFloat(x, 'f', 2, 64)
		}
		return strconv.FormatFloat(x, 'f', 0, 64)
	}
	var b strings.Builder
	b.WriteString(`<div style="margin:1rem 20px; padding:0.75rem; background:#f5f5f5; border-radius:6px; font-family:system-ui,sans-serif; font-size:14px;"><strong>At expiry (over range):</strong><br>`)
	b.WriteString(`Max Profit: ` + formatNum(stats.MaxProfit) + ` &nbsp;|&nbsp; Max Loss: ` + formatNum(stats.MaxLoss))
	rr := "—"
	if stats.MaxLoss < 0 && stats.MaxProfit > 0 {
		rr = strconv.FormatFloat(stats.RewardRisk, 'f', 2, 64)
	}
	b.WriteString(` &nbsp;|&nbsp; Reward/Risk: ` + rr + `<br>`)
	b.WriteString(`Breakeven: `)
	if len(stats.Breakevens) > 0 {
		for i, be := range stats.Breakevens {
			if i > 0 {
				b.WriteString(`, `)
			}
			b.WriteString(formatNum(be))
		}
	} else {
		b.WriteString(`—`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

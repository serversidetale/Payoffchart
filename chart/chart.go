package chart

import (
	"fmt"
	"io"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/serversidetale/payoffchart/payoff"
)

// RenderOpts optionally adds a "before expiry" curve. Zero values mean at-expiry only.
type RenderOpts struct {
	DaysToExpiry int     // if > 0, add a second series for P&L today (Black-Scholes)
	Volatility   float64 // as decimal, e.g. 0.2 for 20% (used when DaysToExpiry > 0)
	RiskFreeRate float64 // as decimal, e.g. 0.05 for 5%
}

// RenderPayoff writes an HTML file with a payoff chart for the given strategy.
// If renderOpts is non-nil and DaysToExpiry > 0, adds a second line for P&L before expiry (Black-Scholes).
func RenderPayoff(w io.Writer, strategy payoff.Strategy, spotMin, spotMax float64, title string, renderOpts *RenderOpts) error {
	const numPoints = 200
	prices, payoffsAtExpiry := strategy.PayoffSeries(spotMin, spotMax, numPoints)

	xLabels := make([]string, len(prices))
	for i, p := range prices {
		xLabels[i] = formatPrice(p)
	}

	lineDataExpiry := make([]opts.LineData, len(payoffsAtExpiry))
	for i, p := range payoffsAtExpiry {
		lineDataExpiry[i] = opts.LineData{Value: roundToTwo(p)}
	}

	subtitle := "P&L at expiration (per strategy)"
	if renderOpts != nil && renderOpts.DaysToExpiry > 0 {
		subtitle = fmt.Sprintf("At expiry vs before expiry (%d days, vol %.0f%%, r %.1f%%)", renderOpts.DaysToExpiry, renderOpts.Volatility*100, renderOpts.RiskFreeRate*100)
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title:    title,
			Subtitle: subtitle,
			Top:      "10px",
		}),
		charts.WithGridOpts(opts.Grid{
			Top:          "18%",
			Left:         "12%",
			Right:        "8%",
			Bottom:       "15%",
			ContainLabel: true,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Underlying price",
			Type: "category",
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: "P&L",
			SplitLine: &opts.SplitLine{
				Show: true,
			},
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    true,
			Trigger: "axis",
		}),
	)

	line.SetXAxis(xLabels).AddSeries("At expiry", lineDataExpiry,
		charts.WithLineChartOpts(opts.LineChart{
			Smooth:     true,
			ShowSymbol: false,
		}),
		charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: 0.2,
		}),
	)

	if renderOpts != nil && renderOpts.DaysToExpiry > 0 {
		T := float64(renderOpts.DaysToExpiry) / 365.0
		sigma := renderOpts.Volatility
		if sigma <= 0 {
			sigma = 0.2
		}
		r := renderOpts.RiskFreeRate
		_, payoffsBefore := strategy.PayoffSeriesBeforeExpiry(spotMin, spotMax, T, sigma, r, numPoints)
		lineDataBefore := make([]opts.LineData, len(payoffsBefore))
		for i, p := range payoffsBefore {
			lineDataBefore[i] = opts.LineData{Value: roundToTwo(p)}
		}
		line.AddSeries("Before expiry (today)", lineDataBefore,
			charts.WithLineChartOpts(opts.LineChart{
				Smooth:     true,
				ShowSymbol: false,
			}),
		)
	}

	return line.Render(w)
}

func formatPrice(p float64) string {
	if p >= 1000 || (p < 0.01 && p > -0.01) {
		return fmt.Sprintf("%.0f", p)
	}
	return fmt.Sprintf("%.2f", p)
}

func roundToTwo(x float64) float64 {
	return float64(int(x*100+0.5)) / 100
}

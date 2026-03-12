// Command payoff generates an options payoff chart and writes it to payoff.html.
package main

import (
	"fmt"
	"os"

	"github.com/payoffchart/internal/chart"
	"github.com/payoffchart/internal/payoff"
)

func main() {
	// Example: bull call spread (long call 100, short call 110)
	// Spot range 90–120, premiums per share
	strategy := payoff.Strategy{
		{Type: payoff.Call, Side: payoff.Long, Strike: 100, Premium: 3.0, Contracts: 1, Multiplier: 100},
		{Type: payoff.Call, Side: payoff.Short, Strike: 110, Premium: 1.0, Contracts: 1, Multiplier: 100},
	}

	spotMin, spotMax := 90.0, 120.0
	title := "Bull Call Spread (Long 100C @ 3, Short 110C @ 1)"

	outPath := "payoff.html"
	f, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	if err := chart.RenderPayoff(f, strategy, spotMin, spotMax, title, nil); err != nil {
		fmt.Fprintf(os.Stderr, "render: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s — open in a browser to view the payoff chart.\n", outPath)
}

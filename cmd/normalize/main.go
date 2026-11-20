// Command normalize reads a JSON spec (values as strings) and prints each
// resource converted into the ledger's base unit with exact arithmetic.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/neko233-com/game-ops/normalize"
	"github.com/neko233-com/game-ops/num"
)

type spec struct {
	Base  string `json:"base"`
	Units []struct {
		ID     string `json:"id"`
		Symbol string `json:"symbol"`
		Desc   string `json:"desc"`
	} `json:"units"`
	Rates []struct {
		From  string `json:"from"`
		To    string `json:"to"`
		Value string `json:"value"`
	} `json:"rates"`
	Conversions []struct {
		From   string `json:"from"`
		Amount string `json:"amount"`
	} `json:"conversions"`
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: normalize <spec.json>")
		os.Exit(2)
	}
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "read spec:", err)
		os.Exit(1)
	}
	var s spec
	if err := json.Unmarshal(data, &s); err != nil {
		fmt.Fprintln(os.Stderr, "parse spec:", err)
		os.Exit(1)
	}

	units := make([]normalize.Unit, 0, len(s.Units))
	for _, u := range s.Units {
		units = append(units, normalize.Unit{ID: u.ID, Symbol: u.Symbol, Desc: u.Desc})
	}
	rates := make([]normalize.Rate, 0, len(s.Rates))
	for _, r := range s.Rates {
		rates = append(rates, normalize.Rate{From: r.From, To: r.To, Value: num.MustFromString(r.Value)})
	}
	ledger, err := normalize.NewLedger(s.Base, units, rates)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ledger:", err)
		os.Exit(1)
	}

	fmt.Printf("Base unit: %s\n", ledger.Base())
	for _, c := range s.Conversions {
		amount := num.MustFromString(c.Amount)
		got, err := ledger.Normalize(c.From, amount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "convert %s: %v\n", c.From, err)
			os.Exit(1)
		}
		fmt.Printf("  %s %s = %s %s\n", c.Amount, c.From, got.DecimalString(8), ledger.Base())
	}
}

// CLI for exact rational formula evaluation.
// Usage:
//
//	go run ./cmd/formula '2^3^2'
//	go run ./cmd/formula 'a+b*c' -vars 'a=1,b=2,c=3'
//	go run ./cmd/formula -expr 'arpu * pay_rate' -vars 'arpu=1.2,pay_rate=0.06'
package formulacmd

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/neko233-com/game-ops/formula"
	"github.com/neko233-com/game-ops/num"
)

func Main() {
	var (
		expr     string
		varsList string
		prec     uint
	)
	flag.StringVar(&expr, "expr", "", "expression string")
	flag.StringVar(&varsList, "vars", "", "comma-separated k=v bindings")
	flag.UintVar(&prec, "prec", 6, "decimal precision")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "formula — exact rational expression evaluator\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  go run ./cmd/formula '2^3^2'\n  go run ./cmd/formula 'a+b*c' -vars 'a=1,b=2,c=3'\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Support both `-expr X` and positional arg.
	if expr == "" && flag.NArg() > 0 {
		expr = flag.Arg(0)
	}
	if expr == "" {
		flag.Usage()
		os.Exit(1)
	}

	vars := map[string]num.Num{}
	for _, pair := range strings.Split(varsList, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		eq := strings.IndexByte(pair, '=')
		if eq < 0 {
			fmt.Fprintf(os.Stderr, "invalid -vars entry: %q (need k=v)\n", pair)
			os.Exit(2)
		}
		k := strings.TrimSpace(pair[:eq])
		v := strings.TrimSpace(pair[eq+1:])
		n, err := num.FromString(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid value for %s: %s\n", k, err)
			os.Exit(2)
		}
		vars[k] = n
	}

	result, err := formula.Eval(expr, vars)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("expr: %s\n", expr)
	if len(vars) > 0 {
		fmt.Println("vars:")
		keys := make([]string, 0, len(vars))
		for k := range vars {
			keys = append(keys, k)
		}
		for _, k := range keys {
			fmt.Printf("  %s = %s\n", k, vars[k].String())
		}
	}
	fmt.Printf("result: %s (≈ %s)\n", result.String(), result.DecimalString(prec))
}

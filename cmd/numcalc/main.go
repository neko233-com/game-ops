// CLI for pure exact-rational calculations.
//
// This is the atomic arithmetic/statistics/probability tool used by both
// humans and the agent. All results are computed with arbitrary-precision
// big.Rat — never float64.
//
// Usage examples:
//
//	# exact arithmetic
//	go run ./cmd/numcalc '1/3 + 2/7'           // → 13/21
//	go run ./cmd/numcalc '100 * 1.15'          // → 115
//
//	# combinations / gacha
//	go run ./cmd/numcalc comb 10 3             // → 120
//	go run ./cmd/numcalc perm 10 3             // → 720
//	go run ./cmd/numcalc binom 10 3 0.06       // C(10,3)·0.06³·0.94⁷
//	go run ./cmd/numcalc geom 5 0.06           // P(first success on trial 5)
//
//	# statistics
//	go run ./cmd/numcalc mean 10 20 30         // → 20
//	go run ./cmd/numcalc median 1 2 3 4 5      // → 3
//	go run ./cmd/numcalc p50 1 2 3 4 5 6 7 8   // → 4.5
//	go run ./cmd/numcalc var 1 2 3 4 5         // population variance
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/neko233-com/game-ops/formula"
	"github.com/neko233-com/game-ops/num"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "numcalc — exact rational calculator (big.Rat, no floats)\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  go run ./cmd/numcalc <expr>         # evaluate arithmetic expression\n")
		fmt.Fprintf(os.Stderr, "  go run ./cmd/numcalc <subcmd> ...   # comb|perm|binom|geom|mean|median|pXX|var\n")
	}
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	first := flag.Arg(0)
	switch first {
	case "comb", "C":
		if flag.NArg() != 3 {
			fmt.Fprintf(os.Stderr, "usage: numcalc comb n k\n")
			os.Exit(1)
		}
		n, _ := strconv.Atoi(flag.Arg(1))
		k, _ := strconv.Atoi(flag.Arg(2))
		out := num.Comb(n, k)
		fmt.Printf("C(%d,%d) = %s (≈ %s)\n", n, k, out.String(), out.DecimalString(6))
	case "perm", "P":
		if flag.NArg() != 3 {
			fmt.Fprintf(os.Stderr, "usage: numcalc perm n k\n")
			os.Exit(1)
		}
		n, _ := strconv.Atoi(flag.Arg(1))
		k, _ := strconv.Atoi(flag.Arg(2))
		out := num.Perm(n, k)
		fmt.Printf("P(%d,%d) = %s (≈ %s)\n", n, k, out.String(), out.DecimalString(6))
	case "binom":
		if flag.NArg() != 4 {
			fmt.Fprintf(os.Stderr, "usage: numcalc binom n k p\n")
			os.Exit(1)
		}
		n, _ := strconv.Atoi(flag.Arg(1))
		k, _ := strconv.Atoi(flag.Arg(2))
		p, _ := num.FromString(flag.Arg(3))
		out := num.BinomialP(n, k, p)
		fmt.Printf("binom(n=%d,k=%d,p=%s) = %s (≈ %s)\n", n, k, p.String(), out.String(), out.DecimalString(6))
	case "geom":
		if flag.NArg() != 3 {
			fmt.Fprintf(os.Stderr, "usage: numcalc geom k p\n")
			os.Exit(1)
		}
		k, _ := strconv.Atoi(flag.Arg(1))
		p, _ := num.FromString(flag.Arg(2))
		out := num.GeometricP(k, p)
		fmt.Printf("geom(k=%d,p=%s) = %s (≈ %s)\n", k, p.String(), out.String(), out.DecimalString(6))
	case "mean":
		vals, err := parseNums(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		out := num.Mean(vals...)
		fmt.Printf("mean = %s (≈ %s)\n", out.String(), out.DecimalString(6))
	case "median":
		vals, err := parseNums(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		out := num.Median(vals...)
		fmt.Printf("median = %s (≈ %s)\n", out.String(), out.DecimalString(6))
	case "var":
		vals, err := parseNums(flag.Args()[1:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		out := num.Variance(vals...)
		fmt.Printf("variance = %s (≈ %s)\n", out.String(), out.DecimalString(6))
	default:
		// Check for pXX percentile shorthand.
		if strings.HasPrefix(first, "p") && len(first) > 1 {
			pct, err := strconv.Atoi(first[1:])
			if err == nil && pct >= 0 && pct <= 100 {
				vals, err := parseNums(flag.Args()[1:])
				if err != nil {
					fmt.Fprintf(os.Stderr, "error: %v\n", err)
					os.Exit(2)
				}
				p := num.MustFromString(strconv.Itoa(pct) + "/100")
				out := num.Percentile(p, vals...)
				fmt.Printf("%s = %s (≈ %s)\n", strings.ToUpper(first), out.String(), out.DecimalString(6))
				return
			}
		}
		// Fall back to expression evaluator — no command prefix.
		result, err := formula.Eval(first, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("%s = %s (≈ %s)\n", first, result.String(), result.DecimalString(6))
	}
}

func parseNums(args []string) ([]num.Num, error) {
	out := make([]num.Num, 0, len(args))
	for _, a := range args {
		n, err := num.FromString(a)
		if err != nil {
			return nil, fmt.Errorf("%q: %w", a, err)
		}
		out = append(out, n)
	}
	return out, nil
}

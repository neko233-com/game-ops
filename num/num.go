// Package num provides exact, arbitrary-precision rational arithmetic.
//
// Values are backed by math/big.Rat, so there is no float64/int64 ceiling and
// no rounding: 0.1 + 0.2 is exactly 3/10, and numbers of any magnitude are
// representable. This is the numeric foundation for "任意数值" support.
package num

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
)

// ErrDivZero is returned by Div and Inv when the divisor is zero.
var ErrDivZero = errors.New("num: division by zero")

// Num is an exact rational number. The zero value is a valid zero.
type Num struct {
	r *big.Rat
}

func (n Num) rat() *big.Rat {
	if n.r == nil {
		return new(big.Rat)
	}
	return new(big.Rat).Set(n.r)
}

func fromRat(r *big.Rat) Num { return Num{r: r} }

// Zero returns 0.
func Zero() Num { return Num{} }

// One returns 1.
func One() Num { return Num{r: big.NewRat(1, 1)} }

// FromInt64 converts an int64 into an exact Num.
func FromInt64(i int64) Num { return Num{r: big.NewRat(i, 1)} }

// FromString parses a fraction ("a/b") or a decimal possibly with an exponent
// ("0.1", "1.5e3", "-2e-2"). All parsing is exact.
func FromString(s string) (Num, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Num{}, errors.New("num: empty string")
	}
	r, ok := new(big.Rat).SetString(s)
	if !ok {
		return Num{}, errors.New("num: invalid number: " + s)
	}
	return fromRat(r), nil
}

// MustFromString is FromString but panics on invalid input. For config/constants.
func MustFromString(s string) Num {
	n, err := FromString(s)
	if err != nil {
		panic(err)
	}
	return n
}

// Add returns n + o.
func (n Num) Add(o Num) Num { return fromRat(new(big.Rat).Add(n.rat(), o.rat())) }

// Sub returns n - o.
func (n Num) Sub(o Num) Num { return fromRat(new(big.Rat).Sub(n.rat(), o.rat())) }

// Mul returns n * o.
func (n Num) Mul(o Num) Num { return fromRat(new(big.Rat).Mul(n.rat(), o.rat())) }

// Div returns n / o, or ErrDivZero when o is zero.
func (n Num) Div(o Num) (Num, error) {
	or := o.rat()
	if or.Sign() == 0 {
		return Num{}, ErrDivZero
	}
	return fromRat(new(big.Rat).Quo(n.rat(), or)), nil
}

// Neg returns -n.
func (n Num) Neg() Num { return fromRat(new(big.Rat).Neg(n.rat())) }

// Abs returns |n|.
func (n Num) Abs() Num { return fromRat(new(big.Rat).Abs(n.rat())) }

// Inv returns 1/n, or ErrDivZero when n is zero.
func (n Num) Inv() (Num, error) {
	r := n.rat()
	if r.Sign() == 0 {
		return Num{}, ErrDivZero
	}
	return fromRat(new(big.Rat).Inv(r)), nil
}

// Cmp compares n and o: -1 if n<o, 0 if equal, +1 if n>o.
func (n Num) Cmp(o Num) int { return n.rat().Cmp(o.rat()) }

// Equal reports whether n == o.
func (n Num) Equal(o Num) bool { return n.Cmp(o) == 0 }

// Sign is -1, 0 or +1 for negative, zero or positive.
func (n Num) Sign() int { return n.rat().Sign() }

// IsZero reports whether n == 0.
func (n Num) IsZero() bool { return n.Sign() == 0 }

// Min returns the smaller of n and o.
func (n Num) Min(o Num) Num {
	if n.Cmp(o) <= 0 {
		return n
	}
	return o
}

// Max returns the larger of n and o.
func (n Num) Max(o Num) Num {
	if n.Cmp(o) >= 0 {
		return n
	}
	return o
}

// Clamp returns n constrained to [lo, hi].
func (n Num) Clamp(lo, hi Num) Num { return n.Max(lo).Min(hi) }

// Pow returns n^exp using exponentiation by squaring (exact for any exponent).
func (n Num) Pow(exp uint) Num {
	result := One()
	base := n
	for exp > 0 {
		if exp&1 == 1 {
			result = result.Mul(base)
		}
		base = base.Mul(base)
		exp >>= 1
	}
	return result
}

// Float64 returns an approximate float64 (informational only; arithmetic is exact).
func (n Num) Float64() float64 {
	f, _ := n.rat().Float64()
	return f
}

// Rat returns a copy of the underlying *big.Rat.
func (n Num) Rat() *big.Rat { return n.rat() }

// String renders the exact canonical form: an integer for whole values,
// otherwise a reduced fraction "a/b".
func (n Num) String() string {
	return n.rat().RatString()
}

// DecimalString renders a decimal with up to prec digits after the point,
// trimming trailing zeros. Terminating decimals render exactly.
func (n Num) DecimalString(prec uint) string {
	r := n.rat()
	neg := r.Sign() < 0
	num := new(big.Int).Abs(r.Num())
	den := r.Denom()

	intPart, rem := new(big.Int).QuoRem(num, den, new(big.Int))
	var frac strings.Builder
	for i := uint(0); i < prec && rem.Sign() != 0; i++ {
		rem.Mul(rem, big.NewInt(10))
		digit, next := new(big.Int).QuoRem(rem, den, new(big.Int))
		frac.WriteByte(byte('0' + digit.Int64()))
		rem = next
	}
	digits := strings.TrimRight(frac.String(), "0")
	var out strings.Builder
	if neg {
		out.WriteByte('-')
	}
	out.WriteString(intPart.String())
	if digits != "" {
		out.WriteByte('.')
		out.WriteString(digits)
	}
	return out.String()
}

// MarshalText lets Num be used as a JSON map key and in text encodings.
func (n Num) MarshalText() ([]byte, error) { return []byte(n.String()), nil }

// UnmarshalText parses the exact canonical or decimal form.
func (n *Num) UnmarshalText(b []byte) error {
	parsed, err := FromString(string(b))
	if err != nil {
		return err
	}
	*n = parsed
	return nil
}

// MarshalJSON encodes the number as a JSON string (preserving arbitrary precision).
func (n Num) MarshalJSON() ([]byte, error) { return json.Marshal(n.String()) }

// UnmarshalJSON decodes a JSON string produced by MarshalJSON.
func (n *Num) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	return n.UnmarshalText([]byte(s))
}

// Sum adds all operands, returning 0 for no operands.
func Sum(values ...Num) Num {
	total := Zero()
	for _, v := range values {
		total = total.Add(v)
	}
	return total
}

// Product multiplies all operands, returning 1 for no operands.
func Product(values ...Num) Num {
	total := One()
	for _, v := range values {
		total = total.Mul(v)
	}
	return total
}

// Mean returns the arithmetic mean. Returns 0 for empty input.
func Mean(values ...Num) Num {
	if len(values) == 0 {
		return Zero()
	}
	total := Sum(values...)
	count := FromInt64(int64(len(values)))
	out, _ := total.Div(count)
	return out
}

// Variance returns the population variance (Σ(v-mean)² / n). Empty → 0.
func Variance(values ...Num) Num {
	if len(values) == 0 {
		return Zero()
	}
	m := Mean(values...)
	count := FromInt64(int64(len(values)))
	sumSq := Zero()
	for _, v := range values {
		d := v.Sub(m)
		sumSq = sumSq.Add(d.Mul(d))
	}
	out, _ := sumSq.Div(count)
	return out
}

// Factorial returns n! exactly. n<0 returns 0.
func Factorial(n int) Num {
	if n < 0 {
		return Zero()
	}
	result := One()
	for i := int64(2); i <= int64(n); i++ {
		result = result.Mul(FromInt64(i))
	}
	return result
}

// Comb returns the binomial coefficient C(n, k) exactly.
// C(n,k) = n! / (k! (n-k)!). Returns 0 for out-of-range k.
func Comb(n, k int) Num {
	if k < 0 || k > n || n < 0 {
		return Zero()
	}
	if k > n-k {
		k = n - k
	}
	// Iterative: result = Π_{i=1}^{k} (n-k+i)/i
	result := One()
	for i := 1; i <= k; i++ {
		mul, _ := FromInt64(int64(n - k + i)).Div(FromInt64(int64(i)))
		result = result.Mul(mul)
	}
	return result
}

// Perm returns n!/(n-k)! exactly. Returns 0 for out-of-range k.
func Perm(n, k int) Num {
	if k < 0 || k > n || n < 0 {
		return Zero()
	}
	result := One()
	for i := 0; i < k; i++ {
		result = result.Mul(FromInt64(int64(n - i)))
	}
	return result
}

// BinomialP returns P(X=k) for n Bernoulli trials with success probability p.
// This is the exact form C(n,k) · p^k · (1-p)^{n-k}, used for gacha drop rates.
func BinomialP(n, k int, p Num) Num {
	if k < 0 || k > n || n < 0 {
		return Zero()
	}
	c := Comb(n, k)
	q := One().Sub(p)
	return c.Mul(p.Pow(uint(k))).Mul(q.Pow(uint(n - k)))
}

// GeometricP returns P(X = k) — first success on trial k — with success rate p.
// P(X=k) = (1-p)^{k-1} · p for k=1,2,3...
func GeometricP(k int, p Num) Num {
	if k < 1 || p.Sign() <= 0 || p.Cmp(One()) >= 0 {
		return Zero()
	}
	q := One().Sub(p)
	return q.Pow(uint(k - 1)).Mul(p)
}

// Median returns the middle value. Sorts a copy; does not mutate input.
func Median(values ...Num) Num {
	if len(values) == 0 {
		return Zero()
	}
	sorted := make([]Num, len(values))
	copy(sorted, values)
	quickSort(sorted, 0, len(sorted)-1)
	mid := len(sorted) / 2
	if len(sorted)%2 == 1 {
		return sorted[mid]
	}
	a, _ := sorted[mid-1].Add(sorted[mid]).Div(Two)
	return a
}

// Percentile returns the p-th percentile (0 ≤ p ≤ 1) using linear
// interpolation between the two nearest ranks. Empty → 0.
func Percentile(p Num, values ...Num) Num {
	if len(values) == 0 {
		return Zero()
	}
	sorted := make([]Num, len(values))
	copy(sorted, values)
	quickSort(sorted, 0, len(sorted)-1)
	if p.Cmp(Zero()) <= 0 {
		return sorted[0]
	}
	if p.Cmp(One()) >= 0 {
		return sorted[len(sorted)-1]
	}
	// rank = p * (n - 1), interpolate.
	idx := p.Mul(FromInt64(int64(len(sorted) - 1))).Float64()
	lo := int(idx)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	f, _ := FromString(fmt.Sprintf("%g", frac))
	return sorted[lo].Add(sorted[hi].Sub(sorted[lo]).Mul(f))
}

// Two is 2 as an exact Num. Kept here for convenience (vs. MustFromString).
var Two = numTwo()

func numTwo() Num { return FromInt64(2) }

// quickSort is an in-place sort of []Num by exact comparison.
func quickSort(a []Num, lo, hi int) {
	if lo >= hi {
		return
	}
	pivot := a[hi]
	i := lo - 1
	for j := lo; j < hi; j++ {
		if a[j].Cmp(pivot) <= 0 {
			i++
			a[i], a[j] = a[j], a[i]
		}
	}
	a[i+1], a[hi] = a[hi], a[i+1]
	p := i + 1
	quickSort(a, lo, p-1)
	quickSort(a, p+1, hi)
}

package got

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
)

// Assertions helpers
type Assertions struct {
	Testable

	must bool

	desc string

	d  func(v interface{}) string    // Options.Dump
	k  func(string) string           // Options.Keyword
	df func(a, b interface{}) string // Options.Diff
}

// Desc returns a clone with the description for failure enabled
func (as Assertions) Desc(format string, args ...interface{}) Assertions {
	n := as
	n.desc = fmt.Sprintf(format, args...)
	return n
}

// Must returns a clone with the FailNow enabled
func (as Assertions) Must() Assertions {
	n := as
	n.must = true
	return n
}

// Eq asserts that x equals y when converted to the same type, such as compare float 1.0 and integer 1 .
// For strict value and type comparison use Assertions.Equal .
func (as Assertions) Eq(x, y interface{}) {
	as.Helper()
	if compare(x, y) == 0 {
		return
	}
	if sameType(x, y) {
		as.err("%s%s%s%s", as.d(x), as.k("not =="), as.d(y), as.diff(x, y))
		return
	}
	as.err("%s%s%s%s%s", as.d(x), as.k("not =="), as.d(y), as.k("even when converted to the same type"), as.diff(x, y))
}

// Neq asserts that x not equals y even when converted to the same type.
func (as Assertions) Neq(x, y interface{}) {
	as.Helper()
	if compare(x, y) != 0 {
		return
	}

	if sameType(x, y) {
		as.err("%s%s%s", as.d(x), as.k("=="), as.d(y))
		return
	}
	as.err("%s%s%s%s", as.d(x), as.k("=="), as.d(y), as.k("when converted to the same type"))
}

// Equal asserts that x equals y.
// For loose type comparison use Assertions.Eq, such as compare float 1.0 and integer 1 .
func (as Assertions) Equal(x, y interface{}) {
	as.Helper()
	if x == y {
		return
	}
	as.err("%s%s%s%s", as.d(x), as.k("not =="), as.d(y), as.diff(x, y))
}

// Gt asserts that x is greater than y.
func (as Assertions) Gt(x, y interface{}) {
	as.Helper()
	if compare(x, y) > 0 {
		return
	}
	as.err("%s%s%s", as.d(x), as.k("not >"), as.d(y))
}

// Gte asserts that x is greater than or equal to y.
func (as Assertions) Gte(x, y interface{}) {
	as.Helper()
	if compare(x, y) >= 0 {
		return
	}
	as.err("%s%s%s", as.d(x), as.k("not ≥"), as.d(y))
}

// Lt asserts that x is less than y.
func (as Assertions) Lt(x, y interface{}) {
	as.Helper()
	if compare(x, y) < 0 {
		return
	}
	as.err("%s%s%s", as.d(x), as.k("not <"), as.d(y))
}

// Lte asserts that x is less than or equal to b.
func (as Assertions) Lte(x, y interface{}) {
	as.Helper()
	if compare(x, y) <= 0 {
		return
	}
	as.err("%s%s%s", as.d(x), as.k("not ≤"), as.d(y))
}

// True asserts that x is true.
func (as Assertions) True(x bool) {
	as.Helper()
	if x {
		return
	}
	as.err("%s%s", as.k("should be"), as.d(true))
}

// False asserts that x is false.
func (as Assertions) False(x bool) {
	as.Helper()
	if !x {
		return
	}
	as.err("%s%s", as.k("should be"), as.d(false))
}

// Nil asserts that the last item in args is nil
func (as Assertions) Nil(args ...interface{}) {
	as.Helper()
	if len(args) == 0 {
		as.err("%s", as.k("no args received"))
		return
	}
	last := args[len(args)-1]
	if isNil(last) {
		return
	}
	as.err("%s%s%s%s", as.k("last value"), as.d(last), as.k("should be"), as.d(nil))
}

// NotNil asserts that the last item in args is not nil
func (as Assertions) NotNil(args ...interface{}) {
	as.Helper()
	if len(args) == 0 {
		as.err("%s", as.k("no args received"))
		return
	}
	last := args[len(args)-1]
	if !isNil(last) {
		return
	}
	if last == nil {
		as.err("%s%s", as.k("last value shouldn't be"), as.d(nil))
		return
	}
	as.err("<%s>%s%s", reflect.TypeOf(last), as.k("shouldn't be"), as.d(nil))
}

// Zero asserts x is zero value for its type.
func (as Assertions) Zero(x interface{}) {
	as.Helper()
	if reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface()) {
		return
	}
	as.err("%s%s", as.d(x), as.k("should be zero value for its type"))
}

// NotZero asserts that x is not zero value for its type.
func (as Assertions) NotZero(x interface{}) {
	as.Helper()
	if reflect.DeepEqual(x, reflect.Zero(reflect.TypeOf(x)).Interface()) {
		as.err("%s%s", as.d(x), as.k("should not be zero value for its type"))
	}
}

// Regex asserts that str matches the regex pattern
func (as Assertions) Regex(pattern, str string) {
	as.Helper()
	if regexp.MustCompile(pattern).MatchString(str) {
		return
	}
	as.err("%s%s%s", pattern, as.k("should match"), str)
}

// Has asserts that container contains str
func (as Assertions) Has(container, str string) {
	as.Helper()
	if strings.Contains(container, str) {
		return
	}
	as.err("%s%s%s", container, as.k("should has"), str)
}

// Len asserts that the length of list equals l
func (as Assertions) Len(list interface{}, l int) {
	as.Helper()
	actual := reflect.ValueOf(list).Len()
	if actual == l {
		return
	}
	as.err("%s%d%s%d", as.k("expect len"), actual, as.k("to be"), l)
}

// Err asserts that the last item in args is error
func (as Assertions) Err(args ...interface{}) {
	as.Helper()
	if len(args) == 0 {
		as.err("%s", as.k("no args received"))
		return
	}
	last := args[len(args)-1]
	if err, _ := last.(error); err != nil {
		return
	}
	as.err("%s%s%s", as.k("last value"), as.d(last), as.k("should be <error>"))
}

// E is a shortcut for Must().Nil(args...)
func (as Assertions) E(args ...interface{}) {
	as.Helper()
	as.Must().Nil(args...)
}

// Panic executes fn and asserts that fn panics
func (as Assertions) Panic(fn func()) {
	as.Helper()

	defer func() {
		as.Helper()

		val := recover()
		if val == nil {
			as.err("%s", as.k("should panic"))
		}
	}()

	fn()
}

// Is asserts that x is kind of y, it uses reflect.Kind to compare.
// If x and y are both error type, it will use errors.Is to compare.
func (as Assertions) Is(x, y interface{}) {
	as.Helper()

	if x == nil && y == nil {
		return
	}

	if ae, ok := x.(error); ok {
		if be, ok := y.(error); ok {
			if ae == be {
				return
			}

			if errors.Is(ae, be) {
				return
			}
			as.err("%s%s%s", as.d(x), as.k("should in chain of"), as.d(y))
			return
		}
	}

	at := reflect.TypeOf(x)
	bt := reflect.TypeOf(y)
	if x != nil && y != nil && at.Kind() == bt.Kind() {
		return
	}
	as.err("%s%s%s", as.d(x), as.k("should be kind of"), as.d(y))
}

// Count asserts that the returned function will be called n times
func (as Assertions) Count(n int) func() {
	as.Helper()
	var count int64

	as.Cleanup(func() {
		if int(atomic.LoadInt64(&count)) != n {
			as.Helper()
			as.Logf("Should count %d times, but got %d", n, count)
			as.Fail()
		}
	})

	return func() {
		atomic.AddInt64(&count, 1)
	}
}

func (as Assertions) err(format string, args ...interface{}) {
	as.Helper()

	if as.desc != "" {
		as.Logf("%s", as.desc)
	}
	as.Logf(format, args...)

	if as.must {
		as.FailNow()
		return
	}

	as.Fail()
}

func castType(x, y interface{}) interface{} {
	ta := reflect.ValueOf(x)
	tb := reflect.ValueOf(y)

	if (x == nil || y == nil) && (x != y) {
		return x
	}

	if ta.Type().ConvertibleTo(tb.Type()) {
		return ta.Convert(tb.Type()).Interface()
	}
	return x
}

func compare(x, y interface{}) float64 {
	if reflect.DeepEqual(x, y) {
		return 0
	}

	if na, ok := castType(x, 0.0).(float64); ok {
		if nb, ok := castType(y, 0.0).(float64); ok {
			return na - nb
		}
	}

	sa := fmt.Sprintf("%#v", x)
	sb := fmt.Sprintf("%#v", y)

	return float64(strings.Compare(sa, sb))
}

func sameType(x, y interface{}) bool {
	if x == nil || y == nil {
		return x == y
	}

	return reflect.TypeOf(x).Kind() == reflect.TypeOf(y).Kind()
}

func isNil(x interface{}) (yes bool) {
	if x == nil {
		return true
	}

	try(func() { yes = reflect.ValueOf(x).IsNil() })
	return
}

func (as Assertions) diff(x, y interface{}) string {
	if as.df != nil {
		return as.df(x, y)
	}
	return ""
}

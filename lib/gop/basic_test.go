package gop_test

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/ysmood/got"
	"github.com/ysmood/got/lib/gop"
)

type T struct {
	got.G
}

func Test(t *testing.T) {
	got.Each(t, T{})
}

func (t T) Tokenize() {
	ref := "test"
	timeStamp, _ := time.Parse(time.RFC3339Nano, "2021-08-28T08:36:36.807908+08:00")

	v := []interface{}{
		nil,
		[]interface{}{true, false, uintptr(0x17), float32(100.121111133)},
		true, 10, int8(2), int32(100),
		float64(100.121111133),
		complex64(1 + 2i), complex128(1 + 2i),
		[3]int{1, 2},
		make(chan int),
		make(chan string, 3),
		func(string) int { return 10 },
		map[interface{}]interface{}{
			"test": 10,
			"a":    1,
		},
		unsafe.Pointer(&ref),
		struct {
			Int int
			str string
			M   map[int]int
		}{10, "ok", map[int]int{1: 0x20}},
		[]byte("aa\xe2"),
		[]byte("bytes\n\tbytes"),
		byte('a'),
		byte(1),
		'天',
		"\ntest",
		&ref,
		(*struct{ Int int })(nil),
		&struct{ Int int }{},
		&map[int]int{1: 2, 3: 4},
		&[]int{1, 2},
		&[2]int{1, 2},
		&[]byte{1, 2},
		timeStamp,
		time.Hour,
	}

	out := gop.StripColor(gop.F(v))

	t.Eq(out, `[]interface {}{
    nil,
    []interface {}{
        true,
        false,
        uintptr(23),
        float32(100.12111),
    },
    true,
    10,
    int8(2),
    'd',
    float64(100.121111133),
    complex64(1+2i),
    1+2i,
    [3]int{
        1,
        2,
        0,
    },
    make(chan int),
    make(chan string, 3),
    (func(string) int)(nil),
    map[interface {}]interface {}{
        "a": 1,
        "test": 10,
    },
    unsafe.Pointer(uintptr(`+fmt.Sprintf("%v", &ref)+`)),
    struct { Int int; str string; M map[int]int }{
        Int: 10,
        str: "ok",
        M: map[int]int{
            1: 32,
        },
    },
    gop.Base64("YWHi"),
    []byte("" +
        "bytes\n" +
        "\tbytes"),
    byte('a'),
    byte(0x1),
    '天',
    "" +
        "\n" +
        "test",
    gop.ToPtr("test").(*string),
    (*struct { Int int })(nil),
    &struct { Int int }{
        Int: 0,
    },
    &map[int]int{
        1: 2,
        3: 4,
    },
    &[]int{
        1,
        2,
    },
    &[2]int{
        1,
        2,
    },
    gop.ToPtr([]byte("\x01\x02")).(*[]uint8),
    gop.Time("`+timeStamp.Format(time.RFC3339Nano)+`"),
    gop.Duration("1h0m0s"),
}`)
}

type A struct {
	Int int
	B   *B
}

type B struct {
	s string
	a *A
}

func (t T) CyclicRef() {
	a := A{Int: 10}
	b := B{"test", &a}
	a.B = &b

	ts := gop.Tokenize(a)

	t.Eq(gop.Format(ts, gop.NoTheme), ""+
		"gop_test.A{\n"+
		"    Int: 10,\n"+
		"    B: &gop_test.B{\n"+
		"        s: \"test\",\n"+
		"        a: &gop_test.A{\n"+
		"            Int: 10,\n"+
		"            B: gop.Cyclic(\"B\").(*gop_test.B),\n"+
		"        },\n"+
		"    },\n"+
		"}")
}

func (t T) CyclicMap() {
	a := map[int]interface{}{}
	a[0] = a

	ts := gop.Tokenize(a)

	t.Eq(gop.Format(ts, gop.NoTheme), ""+
		"map[int]interface {}{\n"+
		"    0: gop.Cyclic().(map[int]interface {}),\n"+
		"}")
}

func (t T) CyclicSlice() {
	a := []interface{}{nil}
	a[0] = a

	ts := gop.Tokenize(a)

	t.Eq(gop.Format(ts, gop.NoTheme), ""+
		"[]interface {}{\n"+
		"    gop.Cyclic().([]interface {}),\n"+
		"}")
}

func (t T) Plain() {
	t.Eq(gop.Plain(10), "10")
}

func (t T) P() {
	gop.Stdout = io.Discard
	_, _ = gop.P("test")
	gop.Stdout = os.Stdout
}

func (t T) Others() {
	gop.ToPtr(nil)
	_ = gop.Cyclic("")
	_ = gop.Base64("")
	_ = gop.Time("")
	_ = gop.Duration("")
}

func (t T) GetPrivateFieldErr() {
	t.Panic(func() {
		gop.GetPrivateField(reflect.ValueOf(1), 0)
	})
}

func (t T) Lab() {

}
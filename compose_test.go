package compose

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestCompose(t *testing.T) {
	fn := Must(intToStr, strToInt).(func(int) int)
	assert.Equal(t, 10, fn(10))
}

func TestVariadic(t *testing.T) {
	minus := Must(argVardMinusVard, argVard).(func(string) (string, int))
	s, i := minus("foo")
	assert.Equal(t, "foo minus argVard", s)
	assert.Equal(t, 0, i)

	lit := Must(argVardLit, argVard).(func(string, int) (string, int))
	s, i = lit("foo", 7)
	assert.Equal(t, "foo lit argVard", s)
	assert.Equal(t, 7, i)

	plus := Must(argVardPlus, argVard).(func(string, int, int) (string, int))
	s, i = plus("foo", 8, 9)
	assert.Equal(t, "foo plus argVard", s)
	assert.Equal(t, 17, i)

	slice := Must(argVardSlice, argVard).(func(string, []int) (string, int))
	s, i = slice("foo", []int{10, 11, 12})
	assert.Equal(t, "foo slice argVard", s)
	assert.Equal(t, 33, i)

	emptyToJustVard := Must(empty, justVard).(func() []int)
	assert.Equal(t, []int{}, emptyToJustVard())
}

func intToStr(x int) string {
	return strconv.Itoa(x)
}

func strToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

func argVard(a string, b ...int) (string, int) {
	var s int
	for _, i := range b {
		s += i
	}
	return a + " argVard", s
}

func argVardMinusVard(a string) string {
	return a + " minus"
}

func argVardLit(a string, b int) (string, int) {
	return a + " lit", b
}

func argVardPlus(a string, b, c int) (string, int, int) {
	return a + " plus", b, c
}

func argVardSlice(a string, b []int) (string, []int) {
	return a + " slice", b
}

func justVard(a ...int) []int {
	return a
}

func empty() {}

// I initially had very simple math operations (either returning the value given
// or doubling it), but I think those were getting in-lined, which was a real
// worse case scenario. Casting to a string and back to an int is a pretty
// small overhead, but enough to keep it from inlining.

func BenchmarkCompose(b *testing.B) {
	fn := Must(intToStr, strToInt).(func(int) int)
	for n := 0; n < b.N; n++ {
		fn(n)
	}
}

func BenchmarkCompile(b *testing.B) {
	fn := func(x int) int {
		return strToInt(intToStr(x))
	}
	for n := 0; n < b.N; n++ {
		fn(n)
	}
}

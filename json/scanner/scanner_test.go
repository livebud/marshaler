package scanner

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/matryer/is"
)

// Ensures that a positive number can be scanned.
func TestScanPositiveNumber(t *testing.T) {
	is := is.New(t)
	tok, b, err := NewScanner(strings.NewReader("100")).Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)
	is.Equal(string(b), "100")
}

// Ensures that a negative number can be scanned.
func TestScanNegativeNumber(t *testing.T) {
	is := is.New(t)
	tok, b, err := NewScanner(strings.NewReader("-1")).Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)
	is.Equal(string(b), "-1")
}

// Ensures that a fractional number can be scanned.
func TestScanFloat(t *testing.T) {
	is := is.New(t)
	tok, b, err := NewScanner(strings.NewReader("120.12931")).Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)
	is.Equal(string(b), "120.12931")
}

// Ensures that a fractional number in scientific notation can be scanned.
func TestScanFloatScientific(t *testing.T) {
	is := is.New(t)
	tok, b, err := NewScanner(strings.NewReader("10.1e01")).Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)
	is.Equal(string(b), "10.1e01")

	tok, b, err = NewScanner(strings.NewReader("10.1e-01")).Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)
	is.Equal(string(b), "10.1e-01")

	tok, b, err = NewScanner(strings.NewReader("10.1e+01")).Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)
	is.Equal(string(b), "10.1e01")
	f, _ := strconv.ParseFloat(string(b), 64)
	is.Equal(10.1e+01, f)

	tok, b, err = NewScanner(strings.NewReader("-1e1")).Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)
	is.Equal(string(b), "-1e1")
}

// Ensures that a quoted string can be scanned.
func TestScanString(t *testing.T) {
	is := is.New(t)
	tok, b, err := NewScanner(strings.NewReader(`"hello world"`)).Scan()
	is.NoErr(err)
	is.Equal(tok, TSTRING)
	is.Equal(string(b), "hello world")
}

// Ensures that a quoted string with escaped characters can be scanned.
func TestScanEscapedString(t *testing.T) {
	is := is.New(t)
	tok, b, err := NewScanner(strings.NewReader(`"\"\\\/\b\f\n\r\t"`)).Scan()
	is.NoErr(err)
	is.Equal(tok, TSTRING)
	is.Equal(string(b), "\"\\/\b\f\n\r\t")
}

// Ensures that escaped unicode sequences can be decoded.
func TestScanEscapedUnicode(t *testing.T) {
	is := is.New(t)
	tok, b, err := NewScanner(strings.NewReader(`"\u0026 \u0424 \u03B4 \u03b4"`)).Scan()
	is.NoErr(err)
	is.Equal(tok, TSTRING)
	is.Equal(string(b), "& Ф δ δ")
}

// Ensures that a true value can be scanned.
func TestScanTrue(t *testing.T) {
	is := is.New(t)
	tok, _, err := NewScanner(strings.NewReader(`true`)).Scan()
	is.NoErr(err)
	is.Equal(tok, TTRUE)
}

// Ensures that a false value can be scanned.
func TestScanFalse(t *testing.T) {
	is := is.New(t)
	tok, _, err := NewScanner(strings.NewReader(`false`)).Scan()
	is.NoErr(err)
	is.Equal(tok, TFALSE)
}

// Ensures that a null value can be scanned.
func TestScanNull(t *testing.T) {
	is := is.New(t)
	tok, _, err := NewScanner(strings.NewReader(`null`)).Scan()
	is.NoErr(err)
	is.Equal(tok, TNULL)
}

// Ensures that an EOF gets returned.
func TestScanEOF(t *testing.T) {
	is := is.New(t)
	_, _, err := NewScanner(strings.NewReader(``)).Scan()
	is.Equal(err, io.EOF)
}

// Ensures that a string can be read into a field.
func TestReadString(t *testing.T) {
	is := is.New(t)
	var v string
	err := NewScanner(strings.NewReader(`"foo"`)).ReadString(&v)
	is.NoErr(err)
	is.Equal(v, "foo")
}

// Ensures that strings largers than allocated buffer can be read.
func TestReadHugeString(t *testing.T) {
	is := is.New(t)
	var v string
	huge := strings.Repeat("s", bufSize*3)
	err := NewScanner(strings.NewReader(`"` + huge + `"`)).ReadString(&v)
	is.NoErr(err)
	is.Equal(v, huge)
}

// Ensures that a non-string value is read into a string field as blank.
func TestReadNonStringAsString(t *testing.T) {
	is := is.New(t)
	var v string
	err := NewScanner(strings.NewReader(`12`)).ReadString(&v)
	is.NoErr(err)
	is.Equal(v, "")
}

// Ensures that a non-value returns a read error.
func TestReadNonValueAsString(t *testing.T) {
	is := is.New(t)
	var v string
	err := NewScanner(strings.NewReader(`{`)).ReadString(&v)
	is.True(err != nil)
	// TODO: test error
}

// Ensures that an int can be read into a field.
func TestReadInt(t *testing.T) {
	is := is.New(t)
	var v int
	err := NewScanner(strings.NewReader(`100`)).ReadInt(&v)
	is.NoErr(err)
	is.Equal(v, 100)
}

// Ensures that a non-number value is read into an int field as zero.
func TestReadNonNumberAsInt(t *testing.T) {
	is := is.New(t)
	var v int
	err := NewScanner(strings.NewReader(`"foo"`)).ReadInt(&v)
	is.NoErr(err)
	is.Equal(v, 0)
}

// Ensures that an int64 can be read into a field.
func TestReadInt64(t *testing.T) {
	is := is.New(t)
	var v int64
	err := NewScanner(strings.NewReader(`-100`)).ReadInt64(&v)
	is.NoErr(err)
	is.Equal(v, int64(-100))
}

// Ensures that a uint can be read into a field.
func TestReadUint(t *testing.T) {
	is := is.New(t)
	var v uint
	err := NewScanner(strings.NewReader(`100`)).ReadUint(&v)
	is.NoErr(err)
	is.Equal(v, uint(100))
}

// Ensures that an uint64 can be read into a field.
func TestReadUint64(t *testing.T) {
	is := is.New(t)
	var v uint64
	err := NewScanner(strings.NewReader(`1024`)).ReadUint64(&v)
	is.NoErr(err)
	is.Equal(v, uint64(1024))
}

// Ensures that a float32 can be read into a field.
func TestReadFloat32(t *testing.T) {
	is := is.New(t)
	var v float32
	err := NewScanner(strings.NewReader(`1293.123`)).ReadFloat32(&v)
	is.NoErr(err)
	is.Equal(v, float32(1293.123))
}

// Ensures that a float64 can be read into a field.
func TestReadFloat64(t *testing.T) {
	is := is.New(t)
	var v float64
	err := NewScanner(strings.NewReader(`9871293.414123`)).ReadFloat64(&v)
	is.NoErr(err)
	is.Equal(v, 9871293.414123)
}

// Ensures that a boolean can be read into a field.
func TestReadBoolTrue(t *testing.T) {
	is := is.New(t)
	var v bool
	err := NewScanner(strings.NewReader(`true`)).ReadBool(&v)
	is.NoErr(err)
	is.Equal(v, true)
}

// Ensures whitespace between tokens are ignored.
func TestScanIgnoreWhitespace(t *testing.T) {
	is := is.New(t)
	s := NewScanner(strings.NewReader(" 100 true false "))

	tok, _, err := s.Scan()
	is.NoErr(err)
	is.Equal(tok, TNUMBER)

	tok, _, err = s.Scan()
	is.NoErr(err)
	is.Equal(tok, TTRUE)

	tok, _, err = s.Scan()
	is.NoErr(err)
	is.Equal(tok, TFALSE)

	tok, _, err = s.Scan()
	is.Equal(err, io.EOF)
	is.Equal(tok, 0)
}

// Ensures that a map can be read into a field.
func TestReadMap(t *testing.T) {
	is := is.New(t)
	var v map[string]interface{}
	err := NewScanner(strings.NewReader(`{"foo":"bar", "bat":1293,"truex":true,"falsex":false,"nullx":null,"nested":{"xxx":"yyy"}}`)).ReadMap(&v)
	is.NoErr(err)
	is.Equal(v["foo"], "bar")
	is.Equal(v["bat"], float64(1293))
	is.Equal(v["truex"], true)
	is.Equal(v["falsex"], false)
	_, exists := v["nullx"]
	is.Equal(v["nullx"], nil)
	is.True(exists)
	is.True(v["nested"] != nil)
	nested := v["nested"].(map[string]interface{})
	is.Equal(nested["xxx"], "yyy")
}

// Ensures that a map with arrays can be read into a field.
func TestReadMapWithArray(t *testing.T) {
	is := is.New(t)
	var v map[string]interface{}
	err := NewScanner(strings.NewReader(`{"foo":["bar", 42]}`)).ReadMap(&v)
	is.NoErr(err)
	arr := v["foo"].([]interface{})
	t.Logf("got=%#v", v)
	is.Equal("bar", arr[0].(string))
	is.Equal(42.0, arr[1].(float64))
}

func BenchmarkScanNumber(b *testing.B) {
	withBuffer(b, "100", func(buf []byte) {
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if _, _, err := s.Scan(); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkScanString(b *testing.B) {
	withBuffer(b, `"01234567"`, func(buf []byte) {
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if _, _, err := s.Scan(); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkScanLongString(b *testing.B) {
	withBuffer(b, `"foo foo foo foo foo foo foo foo foo foo foo foo foo foo"`, func(buf []byte) {
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if _, _, err := s.Scan(); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkScanEscapedString(b *testing.B) {
	withBuffer(b, `"\"\\\/\b\f\n\r\t"`, func(buf []byte) {
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if _, _, err := s.Scan(); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkReadString(b *testing.B) {
	withBuffer(b, `"01234567"`, func(buf []byte) {
		var v string
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if err := s.ReadString(&v); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkReadLongString(b *testing.B) {
	withBuffer(b, `"foo foo foo foo foo foo foo foo foo foo foo foo foo foo"`, func(buf []byte) {
		var v string
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if err := s.ReadString(&v); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkReadInt(b *testing.B) {
	withBuffer(b, `"100"`, func(buf []byte) {
		var v int
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if err := s.ReadInt(&v); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkReadFloat64(b *testing.B) {
	withBuffer(b, `"9871293.414123"`, func(buf []byte) {
		var v float64
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if err := s.ReadFloat64(&v); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func BenchmarkReadBool(b *testing.B) {
	withBuffer(b, `true`, func(buf []byte) {
		var v bool
		s := NewScanner(bytes.NewBuffer(buf))
		for i := 0; i < b.N; i++ {
			if err := s.ReadBool(&v); err == io.EOF {
				s = NewScanner(bytes.NewBuffer(buf))
			} else if err != nil {
				b.Fatal("scan error:", err)
			}
		}
	})
}

func withBuffer(b *testing.B, value string, fn func([]byte)) {
	b.StopTimer()
	var str string
	for i := 0; i < 1000; i++ {
		str += value + " "
	}
	b.StartTimer()

	fn([]byte(str))

	b.SetBytes(int64(len(value)))
}

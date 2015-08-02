// +build amd64

package msgp

import (
	"math"
	"reflect"
	"time"
	"unsafe"
)

// ensure 'sz' extra bytes in 'b' btw len(b) and cap(b)
func ensure(b []byte, sz int) ([]byte, int) {
	l := len(b)
	c := cap(b)
	if c-l < sz {
		o := make([]byte, (2*c)+sz) // exponential *aligned* growth
		n := copy(o, b)
		return o[:n+sz], n
	}
	return b[:l+sz], l
}

// AppendMapHeader appends a map header with the
// given size to the slice
func AppendMapHeader(b []byte, sz uint32) []byte {
	o, n := ensure(b, 8)
	return o[:n+putMapHdr(unsafe.Pointer(&o[n]), int(sz))]
}

// AppendArrayHeader appends an array header with
// the given size to the slice
func AppendArrayHeader(b []byte, sz uint32) []byte {
	o, n := ensure(b, 8)
	return o[:len(b)+putArrayHdr(unsafe.Pointer(&o[n]), int(sz))]
}

// AppendNil appends a 'nil' byte to the slice
func AppendNil(b []byte) []byte { return append(b, mnil) }

// AppendFloat64 appends a float64 to the slice
func AppendFloat64(b []byte, f float64) []byte {
	o, n := ensure(b, Float64Size)
	prefixu64(o[n:], math.Float64bits(f), mfloat64)
	return o
}

// AppendFloat32 appends a float32 to the slice
func AppendFloat32(b []byte, f float32) []byte {
	o, n := ensure(b, Float32Size)
	prefixu32(o[n:], math.Float32bits(f), mfloat32)
	return o
}

// AppendInt64 appends an int64 to the slice
func AppendInt64(b []byte, i int64) []byte {
	if i >= 0 {
		switch {
		case i <= math.MaxInt8:
			return append(b, wfixint(uint8(i)))
		case i <= math.MaxInt16:
			o, n := ensure(b, 3)
			putMint16(o[n:], int16(i))
			return o
		case i <= math.MaxInt32:
			o, n := ensure(b, 5)
			putMint32(o[n:], int32(i))
			return o
		default:
			o, n := ensure(b, 9)
			putMint64(o[n:], i)
			return o
		}
	}
	switch {
	case i >= -31:
		return append(b, wnfixint(int8(i)))
	case i >= math.MinInt8:
		o, n := ensure(b, 2)
		putMint8(o[n:], int8(i))
		return o
	case i >= math.MinInt16:
		o, n := ensure(b, 3)
		putMint16(o[n:], int16(i))
		return o
	case i >= math.MinInt32:
		o, n := ensure(b, 5)
		putMint32(o[n:], int32(i))
		return o
	default:
		o, n := ensure(b, 9)
		putMint64(o[n:], i)
		return o
	}
}

// AppendInt appends an int to the slice
func AppendInt(b []byte, i int) []byte { return AppendInt64(b, int64(i)) }

// AppendInt8 appends an int8 to the slice
func AppendInt8(b []byte, i int8) []byte { return AppendInt64(b, int64(i)) }

// AppendInt16 appends an int16 to the slice
func AppendInt16(b []byte, i int16) []byte { return AppendInt64(b, int64(i)) }

// AppendInt32 appends an int32 to the slice
func AppendInt32(b []byte, i int32) []byte { return AppendInt64(b, int64(i)) }

// AppendUint64 appends a uint64 to the slice
func AppendUint64(b []byte, u uint64) []byte {
	switch {
	case u <= (1<<7)-1:
		return append(b, wfixint(uint8(u)))
	case u <= math.MaxUint8:
		o, n := ensure(b, 2)
		putMuint8(o[n:], uint8(u))
		return o
	case u <= math.MaxUint16:
		o, n := ensure(b, 3)
		putMuint16(o[n:], uint16(u))
		return o
	case u <= math.MaxUint32:
		o, n := ensure(b, 5)
		putMuint32(o[n:], uint32(u))
		return o
	default:
		o, n := ensure(b, 9)
		putMuint64(o[n:], u)
		return o

	}
}

// AppendUint appends a uint to the slice
func AppendUint(b []byte, u uint) []byte { return AppendUint64(b, uint64(u)) }

// AppendUint8 appends a uint8 to the slice
func AppendUint8(b []byte, u uint8) []byte { return AppendUint64(b, uint64(u)) }

// AppendByte is analagous to AppendUint8
func AppendByte(b []byte, u byte) []byte { return AppendUint8(b, uint8(u)) }

// AppendUint16 appends a uint16 to the slice
func AppendUint16(b []byte, u uint16) []byte { return AppendUint64(b, uint64(u)) }

// AppendUint32 appends a uint32 to the slice
func AppendUint32(b []byte, u uint32) []byte { return AppendUint64(b, uint64(u)) }

// AppendBytes appends bytes to the slice as MessagePack 'bin' data
func AppendBytes(b []byte, bts []byte) []byte {
	sz := len(bts)

	// NOTE: we're playing a nasty trick here.
	// We need 8 free bytes of space in order
	// to store the prefix+uint32 combo, but
	// that only occurs if sz is already more
	// than 65535.
	o, n := ensure(b, 5+sz)
	hdr := n + putBinHdr(unsafe.Pointer(&o[n]), sz)
	return o[:hdr+copy(o[hdr:], bts)]
}

// AppendBool appends a bool to the slice
func AppendBool(b []byte, t bool) []byte {
	if t {
		return append(b, mtrue)
	}
	return append(b, mfalse)
}

// AppendString appends a string as a MessagePack 'str' to the slice
func AppendString(b []byte, s string) []byte {
	sz := len(s)

	// NOTE: we're playing a nasty trick here.
	// We need 8 free bytes of space in order
	// to store the prefix+uint32 combo, but
	// that only occurs if sz is already more
	// than 65535.
	o, n := ensure(b, 5+sz)
	hdr := n + putStrHdr(unsafe.Pointer(&o[n]), sz)
	return o[:hdr+copy(o[hdr:], s)]
}

// AppendStringFromBytes appends a []byte
// as a MessagePack 'str' to the slice 'b.'
func AppendStringFromBytes(b []byte, s []byte) []byte {
	sz := len(s)
	o, n := ensure(b, 5+sz)
	hdr := n + putStrHdr(unsafe.Pointer(&o[n]), sz)
	return o[:hdr+copy(o[hdr:], s)]
}

// AppendComplex64 appends a complex64 to the slice as a MessagePack extension
func AppendComplex64(b []byte, c complex64) []byte {
	o, n := ensure(b, Complex64Size)
	o[n] = mfixext8
	o[n+1] = Complex64Extension
	put232(o[n+2:], real(c), imag(c))
	return o
}

// AppendComplex128 appends a complex128 to the slice as a MessagePack extension
func AppendComplex128(b []byte, c complex128) []byte {
	o, n := ensure(b, Complex128Size)
	o[n] = mfixext16
	o[n+1] = Complex128Extension
	put264(o[n+2:], real(c), imag(c))
	return o
}

// AppendTime appends a time.Time to the slice as a MessagePack extension
func AppendTime(b []byte, t time.Time) []byte {
	o, n := ensure(b, TimeSize)
	t = t.UTC()
	o[n] = mext8
	o[n+1] = 12
	o[n+2] = TimeExtension
	putUnix(o[n+3:], t.Unix(), int32(t.Nanosecond()))
	return o
}

// AppendMapStrStr appends a map[string]string to the slice
// as a MessagePack map with 'str'-type keys and values
func AppendMapStrStr(b []byte, m map[string]string) []byte {
	sz := uint32(len(m))
	b = AppendMapHeader(b, sz)
	for key, val := range m {
		b = AppendString(b, key)
		b = AppendString(b, val)
	}
	return b
}

// AppendMapStrIntf appends a map[string]interface{} to the slice
// as a MessagePack map with 'str'-type keys.
func AppendMapStrIntf(b []byte, m map[string]interface{}) ([]byte, error) {
	sz := uint32(len(m))
	b = AppendMapHeader(b, sz)
	var err error
	for key, val := range m {
		b = AppendString(b, key)
		b, err = AppendIntf(b, val)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}

// AppendIntf appends the concrete type of 'i' to the
// provided []byte. 'i' must be one of the following:
//  - 'nil'
//  - A bool, float, string, []byte, int, uint, or complex
//  - A map[string]interface{} or map[string]string
//  - A []T, where T is another supported type
//  - A *T, where T is another supported type
//  - A type that satisfieds the msgp.Marshaler interface
//  - A type that satisfies the msgp.Extension interface
func AppendIntf(b []byte, i interface{}) ([]byte, error) {
	if i == nil {
		return AppendNil(b), nil
	}

	// all the concrete types
	// for which we have methods
	switch i := i.(type) {
	case Marshaler:
		return i.MarshalMsg(b)
	case Extension:
		return AppendExtension(b, i)
	case bool:
		return AppendBool(b, i), nil
	case float32:
		return AppendFloat32(b, i), nil
	case float64:
		return AppendFloat64(b, i), nil
	case complex64:
		return AppendComplex64(b, i), nil
	case complex128:
		return AppendComplex128(b, i), nil
	case string:
		return AppendString(b, i), nil
	case []byte:
		return AppendBytes(b, i), nil
	case int8:
		return AppendInt8(b, i), nil
	case int16:
		return AppendInt16(b, i), nil
	case int32:
		return AppendInt32(b, i), nil
	case int64:
		return AppendInt64(b, i), nil
	case int:
		return AppendInt64(b, int64(i)), nil
	case uint:
		return AppendUint64(b, uint64(i)), nil
	case uint8:
		return AppendUint8(b, i), nil
	case uint16:
		return AppendUint16(b, i), nil
	case uint32:
		return AppendUint32(b, i), nil
	case uint64:
		return AppendUint64(b, i), nil
	case time.Time:
		return AppendTime(b, i), nil
	case map[string]interface{}:
		return AppendMapStrIntf(b, i)
	case map[string]string:
		return AppendMapStrStr(b, i), nil
	case []interface{}:
		b = AppendArrayHeader(b, uint32(len(i)))
		var err error
		for _, k := range i {
			b, err = AppendIntf(b, k)
			if err != nil {
				return b, err
			}
		}
		return b, nil
	}

	var err error
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		l := v.Len()
		b = AppendArrayHeader(b, uint32(l))
		for i := 0; i < l; i++ {
			b, err = AppendIntf(b, v.Index(i).Interface())
			if err != nil {
				return b, err
			}
		}
		return b, nil

	default:
		return b, &ErrUnsupportedType{T: v.Type()}
	}
}
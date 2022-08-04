package goload

import (
	"encoding"
	"encoding/base64"
	"encoding/csv"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/samber/lo"
)

var ( // Some type variables for comparison.
	typeOfTextUnmarshaler = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	typeOfByteSlice       = reflect.TypeOf([]byte{})
)

// isSupportedType returns whether the type t is supported by goconfig for parsing.
func isSupportedType(t reflect.Type) error {
	if t.Implements(typeOfTextUnmarshaler) {
		return nil
	}

	if t == typeOfByteSlice {
		return nil
	}

	switch t.Kind() {
	case reflect.Bool:
	case reflect.String:
	case reflect.Float32, reflect.Float64:
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:

	case reflect.Struct:
		for i := 0; i < t.NumField(); i++ {
			if err := isSupportedType(t.Field(i).Type); err != nil {
				return fmt.Errorf("struct with unsupported type: %v", err)
			}
		}

	case reflect.Slice:
		// All but the fixed-bitsize types.
		if err := isSupportedType(t.Elem()); err != nil {
			return fmt.Errorf("slice of unsupported type: %v", err)
		}

	case reflect.Ptr:
		if err := isSupportedType(t.Elem()); err != nil {
			return fmt.Errorf("pointer to unsupported type: %v", err)
		}

	case reflect.Map:
		// if t.Key().Kind() != reflect.String || t.Elem().Kind() != reflect.Interface {
		kinds := []reflect.Kind{reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint64,
			reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64}
		if tm := t.Key().Kind(); !lo.Contains(kinds, tm) || isSupportedType(t.Elem()) != nil {
			return errors.New("only maps of type map[string]interface{} are supported")
		}

	default:
		return errors.New("type not supported")
	}

	return nil
}

// readAsCSV parses a CSV encoded list in its elements.
func readAsCSV(val string) ([]string, error) {
	if val == "" {
		return []string{}, nil
	}
	stringReader := strings.NewReader(val)
	csvReader := csv.NewReader(stringReader)
	return csvReader.Read()
}

// parseSimpleValue parses values other than structs, slices (except []byte),
// and encoding.TextUnmarshaler and stores them in v.
func parseSimpleValue(v reflect.Value, s string) error {
	t := v.Type()

	if t.Implements(typeOfTextUnmarshaler) {
		// Is a reference, we must create element first.
		v.Set(reflect.New(t.Elem()))
		unmarshaler := v.Interface().(encoding.TextUnmarshaler)
		if err := unmarshaler.UnmarshalText([]byte(s)); err != nil {
			return fmt.Errorf("failed to unmarshal '%v' into type %v: %v",
				s, t, err)
		}
		return nil
	}

	if t == typeOfByteSlice {
		decoded, err := base64.StdEncoding.DecodeString(s)
		if err != nil {
			return parseError(s, t, err)
		}
		v.Set(reflect.ValueOf(decoded))
		return nil
	}

	switch t.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Struct:
		// fmt.Printf("struct:%+v\n", t)
		// tv := reflect.New(t)
		// if err := p.setDefaultStruct(tv.Elem(), nil); err != nil {
		// 	panic(err)
		// }
		v.Set(zeroType(t))
		// v.Set(reflect.Indirect(tv))
	case reflect.Pointer:
		fmt.Printf("point:%+v\n", t)
		// tv := reflect.New(t)
		// if err := p.setDefaultStruct(tv.Elem(), nil); err != nil {
		// 	panic(err)
		// }
		tv := zeroType(t)
		v.Set(tv)

	case reflect.Interface:
		// We fill interfaces with the simple string value.
		v.Set(reflect.ValueOf(s))

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return parseError(s, t, err)
		}
		v.SetBool(b)

	case reflect.Float32, reflect.Float64:
		if err := parseFloat(v, s); err != nil {
			return err
		}

	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if err := parseInt(v, s); err != nil {
			return err
		}

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		if err := parseUint(v, s); err != nil {
			return err
		}

	default:
		panic("not simple value")
	}

	return nil
}

// parseInt parses s to any int type and stores it in v.
func parseInt(v reflect.Value, s string) error {
	var bitSize int
	switch v.Kind() {
	case reflect.Int:
		bitSize = 0
	case reflect.Int64:
		bitSize = 64
	case reflect.Int32:
		bitSize = 32
	case reflect.Int16:
		bitSize = 16
	case reflect.Int8:
		bitSize = 8
	default:
		panic("not an int")
	}
	p, err := strconv.ParseInt(s, 10, bitSize)
	if err != nil {
		return parseError(s, v.Type(), err)
	}
	v.SetInt(p)
	return nil
}

// parseUint parses s to any uint type and stores it in v.
func parseUint(v reflect.Value, s string) error {
	var bitSize int
	switch v.Kind() {
	case reflect.Uint:
		bitSize = 0
	case reflect.Uint64:
		bitSize = 64
	case reflect.Uint32:
		bitSize = 32
	case reflect.Uint16:
		bitSize = 16
	case reflect.Uint8:
		bitSize = 8
	default:
		panic("not a uint")
	}
	p, err := strconv.ParseUint(s, 10, bitSize)
	if err != nil {
		return parseError(s, v.Type(), err)
	}
	v.SetUint(p)
	return nil
}

// parseFloat parses s to any float type and stores it in v.
func parseFloat(v reflect.Value, s string) error {
	var bitSize int
	switch v.Kind() {
	case reflect.Float32:
		bitSize = 32
	case reflect.Float64:
		bitSize = 64
	default:
		panic("not a float")
	}
	p, err := strconv.ParseFloat(s, bitSize)
	if err != nil {
		return parseError(s, v.Type(), err)
	}
	v.SetFloat(p)
	return nil
}

func parseError(v string, t reflect.Type, err error) error {
	msg := fmt.Sprintf("failed to parse '%v' into type %v", v, t)
	if err != nil {
		msg += ": " + err.Error()
	}
	return errors.New(msg)
}

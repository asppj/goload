package parse

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

// parseError returns a nicely formatted error indicating that we failed to
// parse v into type t.
func parseError(v string, t reflect.Type, err error) error {
	msg := fmt.Sprintf("failed to parse '%v' into type %v", v, t)
	if err != nil {
		msg += ": " + err.Error()
	}
	return errors.New(msg)
}

// convertibleError returns a nicely formatted error indicating that the value
// v is not convertible to type t.
func convertibleError(v reflect.Value, t reflect.Type) error {
	return fmt.Errorf(
		"incompatible type: %v not convertible to %v", v.Type(), t)
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

// parseSimpleValue parses values other than structs, slices (except []byte),
// and encoding.TextUnmarshaler and stores them in v.
func (p *parser) parseSimpleValue(v reflect.Value, s string) error {
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
		v.Set(p.zeroType(nil, t))
		// v.Set(reflect.Indirect(tv))
	case reflect.Pointer:
		fmt.Printf("point:%+v\n", t)
		// tv := reflect.New(t)
		// if err := p.setDefaultStruct(tv.Elem(), nil); err != nil {
		// 	panic(err)
		// }
		tv := p.zeroType(nil, t)
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

// parseSlice parses s to a slice and stores the slice in v.
func (p *parser) parseSlice(v reflect.Value, s string) error {
	vals, err := readAsCSV(s)
	if err != nil {
		return fmt.Errorf("error parsing comma separated value '%v': %v", s, err)
	}

	slice := reflect.MakeSlice(v.Type(), len(vals), len(vals))
	for i := 0; i < len(vals); i++ {
		if err := p.parseSimpleValue(slice.Index(i), vals[i]); err != nil {
			return err
		}
	}

	v.Set(slice)
	return nil
}

func (p *parser) zeroPointValue(parent *parseField, tv reflect.Value) reflect.Value {
	ty := tv.Type()
	subTy := ty.Elem()
	if subTy.Kind() == reflect.Pointer {
		ele := p.zeroType(parent, subTy)
		// fmt.Println(1, ele.Type().String())
		// fmt.Println(2, tv.Type().String())
		// if ele.Elem().Type().Kind() == reflect.Struct {
		// 	if err := p.setDefaultStruct(ele.Elem(), parent); err != nil {
		// 		panic(fmt.Errorf("set sub struct defualt value error: %w", err))
		// 	}
		// }
		tv.Elem().Set(ele)
	}
	if tv.Elem().Type().Kind() == reflect.Struct {
		if err := p.setDefaultStruct(tv.Elem(), parent); err != nil {
			panic(fmt.Errorf("set sub struct defualt value error: %w", err))
		}
	}
	return tv
}

// init a type value
func (p *parser) zeroType(parent *parseField, ty reflect.Type) (tv reflect.Value) {
	tv = reflect.Zero(ty) // 根值
	if ty.Kind() == reflect.Pointer {
		tv = reflect.New(ty.Elem())
		p.zeroPointValue(parent, tv) // 设置值
	} else if ty.Kind() == reflect.Struct {
		tp := reflect.New(ty)
		if err := p.setDefaultStruct(tp.Elem(), parent); err != nil {
			panic(fmt.Errorf("set sub struct defualt value error: %w", err))
		}
		tv = reflect.Indirect(tp)
	}
	// if tv.Type().Kind() == reflect.Struct {
	// 	if err := p.setDefaultStruct(tv.Elem(), parent); err != nil {
	// 		panic(fmt.Errorf("set sub struct defualt value error: %w", err))
	// 	}
	// }
	// if tv.IsNil() {
	// 	zeroPointValue(tv) // 设置值
	// }
	return tv
}

// parseSlice parses s to a slice and stores the slice in v.
func (p *parser) parseMap(parent *parseField, v reflect.Value, s string) error {
	vals, err := readAsCSV(s)
	if err != nil {
		return fmt.Errorf("error parsing comma separated value '%v': %v", s, err)
	}
	m := reflect.MakeMapWithSize(v.Type(), len(vals))
	// key type,value type
	kt, vt := v.Type().Key(), v.Type().Elem()
	for i := 0; i < len(vals); i++ {
		ele := p.zeroType(parent, vt)
		m.SetMapIndex(reflect.ValueOf(vals[i]).Convert(kt), ele)
		// if err := parseSimpleValue(m.Index(i), vals[i]); err != nil {
		// 	return err
		// }
	}
	v.Set(m)
	return nil
}

// parseMapToStruct converts a map[string]interface{} into a struct value.
func (p *parser) parseMapToStruct(from, to reflect.Value, tagOption *TagOption) error {
	opts, _, err := inspectField(to, nil, tagOption)
	if err != nil {
		// Here we panic since there is a problem in the config structure.
		panic(fmt.Sprintf("error in config structure: "+
			"invalid struct inside slice: %v", err))
	}

keys:
	for _, key := range from.MapKeys() {
		for _, opt := range opts {
			if opt.fullID() == key.String() {
				fromVal := from.MapIndex(key)
				// All values should be interfaces (we have
				// map[string]interface{}), so first uninterface the element.
				if fromVal.Kind() == reflect.Interface {
					fromVal = fromVal.Elem()
				}

				if err := p.setValue(opt.value, fromVal, tagOption); err != nil {
					return fmt.Errorf("failed to set value in nested struct "+
						"slice option '%v': %v", opt.fullID(), err)
				}
				continue keys
			}
		}
		// No option found for the key.
		return fmt.Errorf("found no option with id '%v' in nested struct slice",
			key.String())
	}

	return nil
}

// isKindOrPtrTo returns true is the given type is of the given kind or if it is
// a pointer to the given kind.
func isKindOrPtrTo(t reflect.Type, k reflect.Kind) bool {
	if t.Kind() == k {
		return true
	}
	return t.Kind() == reflect.Ptr && t.Elem().Kind() == k
}

// convertSlice converts the slice from into the slice to by converting all the
// individual elements.
func (p *parser) convertSlice(from, to reflect.Value, tagOption *TagOption) error {
	subType := to.Type().Elem()
	converted := reflect.MakeSlice(to.Type(), from.Len(), from.Len())
	for i := 0; i < from.Len(); i++ {
		elem := from.Index(i)
		if elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}

		// When coming from a file decoder, sices of structs are slices of maps.
		// So when we find a map and the target value is a struct (or a pointer
		// to one), we convert the map into the struct.
		if elem.Kind() == reflect.Map && isKindOrPtrTo(subType, reflect.Struct) {
			inVal := converted.Index(i)
			if subType.Kind() == reflect.Ptr {
				ptr := reflect.New(subType.Elem())
				inVal.Set(ptr)
				inVal = ptr.Elem()
			}
			if err := p.parseMapToStruct(elem, inVal, tagOption); err != nil {
				return fmt.Errorf("failed to convert to struct: %v", err)
			}

			continue
		}

		if !elem.Type().ConvertibleTo(subType) {
			return convertibleError(elem, subType)
		}

		converted.Index(i).Set(elem.Convert(subType))
	}

	to.Set(converted)
	return nil
}

// cleanUpYAML replaces all the map[interface{}]interface{} values into
// map[string]interface{} values.
func cleanUpYAML(v interface{}) interface{} {
	switch v := v.(type) {

	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, v := range v {
			result[k] = cleanUpYAML(v)
		}
		return result

	case map[interface{}]interface{}:
		result := make(map[string]interface{})
		for k, v := range v {
			result[fmt.Sprintf("%v", k)] = cleanUpYAML(v)
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, v := range v {
			result[i] = cleanUpYAML(v)
		}
		return result

	default:
		return v
	}
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
func isSlice(v reflect.Value) bool {
	return v.Kind() == reflect.Slice && v.Type() != typeOfByteSlice
}
func isMap(v reflect.Value) bool {
	return v.Kind() == reflect.Map
}

// isZero checks if the value is the zero value for its type.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Map:
		return v.Len() == 0
	case reflect.Func, reflect.Slice:
		return v.IsNil()
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	case reflect.Ptr:
		return isZero(reflect.Indirect(v))
	}
	// Compare other types directly:
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

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

// setValueByString sets the value by parsing the string.
func (p *parser) setValueByString(v reflect.Value, s string) error {
	if isSlice(v) {
		if err := p.parseSlice(v, s); err != nil {
			return fmt.Errorf("failed to parse slice value: %v", err)
		}
	} else {
		if err := p.parseSimpleValue(v, s); err != nil {
			return fmt.Errorf("failed to parse value: %v", err)
		}
	}

	return nil
}

// setValue sets the option value to the given value.
// If the tye of the value is assignable or convertible to the type of the
// option value, it is directly set after optional conversion.
// If not, but the value is a string, it is passed to setValueByString.
// If not, and both v and the option's value are is a slice, we try converting
// the slice elements to the right elemens of the options slice.
func (p *parser) setValue(toSet, v reflect.Value, tagOption *TagOption) error {
	t := toSet.Type()
	if v.Type().AssignableTo(t) {
		toSet.Set(v)
		return nil
	}

	if v.Type().ConvertibleTo(t) && toSet.Type() != typeOfByteSlice {
		toSet.Set(v.Convert(t))
		return nil
	}

	if v.Type().Kind() == reflect.String {
		return p.setValueByString(toSet, v.String())
	}

	if isSlice(toSet) && v.Type().Kind() == reflect.Slice {
		return p.convertSlice(v, toSet, tagOption)
	}

	return convertibleError(v, toSet.Type())
}

// setSimpleMapValue trues to add the key and value to the map.
func (p *parser) setSimpleMapValue(mapValue reflect.Value, key, value string) error {
	v := reflect.New(mapValue.Type().Elem()).Elem()
	if err := p.parseSimpleValue(v, value); err != nil {
		return err
	}
	mapValue.SetMapIndex(reflect.ValueOf(key), v)
	return nil
}

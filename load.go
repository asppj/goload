package goload

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/samber/lo"
)

type TagOption struct {
	fullTag  reflect.StructTag
	TagName  string
	TagValue *string
}

func (t *TagOption) getDefault() []string {
	if t == nil || t.TagValue == nil {
		return []string{}
	}
	vals, err := readAsCSV(*t.TagValue)
	if err != nil {
		panic(err)
	}
	return vals
}

func (t *TagOption) parseFromField(field reflect.StructField) *TagOption {
	value, ok := field.Tag.Lookup(t.TagName)
	if ok {
		return &TagOption{
			fullTag:  field.Tag,
			TagName:  t.TagName,
			TagValue: &value,
		}
	}
	return &TagOption{
		fullTag:  field.Tag,
		TagName:  t.TagName,
		TagValue: nil,
	}

}
func LoadStruct(c any, tagName string) error {
	rv := reflect.ValueOf(c)
	rt := reflect.TypeOf(c)
	if rt.Kind() != reflect.Ptr {
		return fmt.Errorf("variable must be a pointer to a struct")
	}
	if rv.IsNil() {
		return fmt.Errorf("valiable catnot is nil pointer")
	}
	if t := rt.Elem().Kind(); t != reflect.Struct {
		return errors.New("config variable must be a pointer to a struct")
	}
	tagOption := &TagOption{
		TagName:  tagName,
		TagValue: nil,
	}
	return parseValue(rv.Elem(), tagOption)
}

func parseStruct(v reflect.Value, option *TagOption) error {
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		field := v.Type().Field(i)
		if err := isSupportedType(field.Type); err != nil {
			return fmt.Errorf(
				"type of field %+v (%+v) is not supported: %w",
				field.Name, field.Type, err)
		}
		if !fieldValue.CanSet() {
			continue
		}
		tagOption := option.parseFromField(field)
		// if fieldValue.IsZero() {
		// if isZero(fieldValue) {
		// pField := reflect.New(fieldValue.Type())
		// fieldValue.SetPointer(unsafe.Pointer(pField.Addr().UnsafeAddr()))
		if fieldValue.Type().Kind() == reflect.Struct {
			fmt.Printf("")
		}
		if err := parseValue(fieldValue, tagOption); err != nil {
			return err
			// }
		}
	}
	return nil
}

func parseSlice(v reflect.Value, option *TagOption) error {
	// continue parse sub element
	vals := option.getDefault()
	for i := 0; i < v.Len(); i++ {
		indexValue := v.Index(i)
		if option != nil && option.TagValue != nil {
			var tagValue *string
			if len(vals) > i {
				tagValue = &vals[i]
			}
			opt := &TagOption{
				TagName:  option.TagName,
				fullTag:  option.fullTag,
				TagValue: tagValue,
			}
			if err := parseValue(indexValue, opt); err != nil {
				return err
			}
		}
	}
	return nil
}

func parseMap(v reflect.Value, option *TagOption) error {
	if v.Len() > 0 {
		iter := v.MapRange()
		for iter.Next() {
			indexValue := iter.Value()
			tmp := reflect.New(indexValue.Type())
			tmp.Elem().Set(indexValue)
			if err := parseValue(tmp.Elem(), option); err != nil {
				return err
			}
			v.SetMapIndex(iter.Key(), tmp.Elem())
		}
	}
	return nil
}

func parsePointer(v reflect.Value, option *TagOption) error {
	value := reflect.Zero(v.Type().Elem())
	v.Elem().Set(value)
	if err := parseValue(v.Elem(), option); err != nil {
		return err
	}
	return nil
}

func parseSample(v reflect.Value, option *TagOption) error {
	if v.CanSet() && v.IsZero() && option != nil && option.TagValue != nil {
		strV := *option.TagValue
		if strV == "-" {
			return nil
		}
		if err := parseSimpleValue(v, strV); err != nil {
			return err
		}
	}
	return nil
}

// parse reflect.Value set default value
func parseValue(v reflect.Value, option *TagOption) error {
	setZeroType(v, option)
	switch v.Type().Kind() {
	case reflect.Struct:
		return parseStruct(v, option)
	case reflect.Pointer:
		return parsePointer(v, option)
	case reflect.Map:
		return parseMap(v, option)
	case reflect.Slice, reflect.Array:
		return parseSlice(v, option)
	case reflect.Chan, reflect.Interface, reflect.Invalid, reflect.Complex64,
		reflect.Complex128, reflect.Func, reflect.UnsafePointer, reflect.Uintptr: // no supported
		return nil
	default:
		return parseSample(v, option) // bool,int,string,float...
	}
}

func setZeroType(v reflect.Value, option *TagOption) {
	if !v.CanSet() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsZero() {
			return
		}
		zero := reflect.New(v.Type().Elem())
		v.Set(zero)
	case reflect.Struct:
		if !v.IsZero() {
			return
		}
		zero := reflect.New(v.Type())
		v.Set(reflect.Indirect(zero))
	case reflect.Array, reflect.Slice:
		if !v.IsZero() {
			return
		}
		vals := option.getDefault()
		slice := reflect.MakeSlice(v.Type(), len(vals), len(vals))
		// for i := 0; i < len(vals); i++ {
		// 	if err := parseSimpleValue(slice.Index(i), vals[i]); err != nil {
		// 		panic(err)
		// 	}
		// }
		v.Set(slice)
	case reflect.Map:
		if !v.IsZero() {
			return
		}
		vals := option.getDefault()
		m := reflect.MakeMapWithSize(v.Type(), len(vals))
		// key type,value type
		kt, vt := v.Type().Key(), v.Type().Elem()
		for i := 0; i < len(vals); i++ {
			ele := zeroType(vt)
			if err := parseValue(ele, option); err != nil {
				panic(err)
			}
			if lo.Contains([]reflect.Kind{reflect.Int, reflect.Uint, reflect.Int8, reflect.Int16, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64}, kt.Kind()) {
				if kValue, err := strconv.Atoi(vals[i]); err == nil {
					m.SetMapIndex(reflect.ValueOf(kValue).Convert(kt), ele)
				}
			} else { // string
				m.SetMapIndex(reflect.ValueOf(vals[i]).Convert(kt), ele)
			}
		}
		v.Set(m)
	default:
		if !v.IsZero() {
			return
		}
		zero := reflect.Zero(v.Type())
		v.Set(zero)
	}
}

func zeroPointValue(tv reflect.Value) reflect.Value {
	ty := tv.Type()
	subTy := ty.Elem()
	if subTy.Kind() == reflect.Pointer {
		ele := zeroType(subTy)
		tv.Elem().Set(ele)
	}
	if tv.Elem().Type().Kind() == reflect.Struct {
	}
	return tv
}

// init a type value
func zeroType(ty reflect.Type) (tv reflect.Value) {
	tv = reflect.Zero(ty) // 根值
	if ty.Kind() == reflect.Pointer {
		tv = reflect.New(ty.Elem())
		zeroPointValue(tv) // 设置值
	} else if ty.Kind() == reflect.Struct {
		tp := reflect.New(ty)
		tv = reflect.Indirect(tp)
	}

	return tv
}

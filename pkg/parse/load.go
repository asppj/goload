package parse

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"reflect"
	"strconv"

	"github.com/samber/lo"
)

func (p *parser) Load(content []byte) error {
	if err := p.decoder(content, p.source); err != nil {
		return err
	}
	return p.InspectStruct(p.source)
}

func (p *parser) ImportFile(fileName string) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	return p.decoder(content, p.source)
}

func (p *parser) ExportFile(filePath string) error {
	content, err := p.encoder(p.source)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filePath, content, fs.FileMode(0600))
}

func (p *parser) LoadEnv() error {
	// TODO implement me
	panic("implement me")
}

func (p *parser) LoadCmd() error {
	// TODO implement me
	panic("implement me")
}

func LoadStruct(c any, option *TagOption) error {
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
	return parseValue(rv.Elem(), option)
}

// parse reflect.Value set default value
func parseValue(v reflect.Value, option *TagOption) error {

	setZeroType(v, option)
	switch v.Type().Kind() {
	case reflect.Struct:
		fmt.Printf("caseSet:%v,value:%v\n", v.CanSet(), v.Interface())
		return parseStruct(v, option)
	case reflect.Pointer:
		fmt.Printf("caseSet:%v,value:%v\n", v.CanSet(), v.Interface())
		return parsePointer(v, option)
	case reflect.Map:
		fmt.Printf("caseSet:%v,value:%v\n", v.CanSet(), v.Interface())
		return parseMap(v, option)
	case reflect.Slice, reflect.Array:
		fmt.Printf("caseSet:%v,value:%v\n", v.CanSet(), v.Interface())
		return parseSlice(v, option)
	case reflect.Chan, reflect.Interface, reflect.Invalid, reflect.Complex64,
		reflect.Complex128, reflect.Func, reflect.UnsafePointer, reflect.Uintptr: // no supported
		return nil
	default:
		return parseSample(v, option) // bool,int,string,float...
	}
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
		if option != nil && option.parseField != nil {
			opt := option.clone()
			opt.parseField = option.parseField
			opt.parseField.tagValue.Default = vals[i]
		}
		if !indexValue.CanSet() {
			fmt.Printf("indexValue")
		}
		if err := parseValue(indexValue, option); err != nil {
			return err
		}
	}
	return nil
}

func parseMap(v reflect.Value, option *TagOption) error {
	// continue parse sub element
	// for _, value := range v.MapKeys() {
	// 	indexValue := v.MapIndex(value)
	// 	reflect.ValueOf(indexValue)
	// 	if !indexValue.CanAddr() {
	// 		fmt.Printf("indexValue")
	// 		// indexValue = indexValue.Elem()
	// 	} else {
	// 		fmt.Printf("2")
	// 	}
	// 	// nv := reflect.NewAt(indexValue.Type(), unsafe.Pointer(indexValue.UnsafeAddr()))
	// 	if err := parseValue(indexValue, option); err != nil {
	// 		return err
	// 	}
	// }
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
	if v.CanSet() && v.IsZero() && option != nil && option.parseField != nil {
		strV := option.parseField.tagValue.Default
		if strV == "-" {
			return nil
		}
		if err := newParserWithOption(option).parseSimpleValue(v, strV); err != nil {
			return err
		}
	} else {
		fmt.Printf("%+v;opt:%+v\n", v, option)
	}
	return nil
}

func setZeroType(v reflect.Value, option *TagOption) {
	fmt.Println(v.IsValid(), v.IsZero(), v.CanSet())
	if !v.CanSet() {
		return
	}
	switch v.Kind() {
	case reflect.Pointer:
		if !v.IsZero() {
			return
		}
		zero := reflect.New(v.Type().Elem())
		fmt.Printf("v.Type:%s,zero.Type:%s\n", v.Type(), zero.Type())
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
		fmt.Printf(" type:%s,canSet:%v,canInterface:%v,isZero:%v \n", v.Type(), v.CanSet(), v.CanInterface(), v.IsZero())
		zero := reflect.Zero(v.Type())
		v.Set(zero)
		fmt.Printf("isZero:%v \n", v.IsZero())
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

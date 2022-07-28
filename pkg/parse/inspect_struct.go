package parse

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// InspectStruct 解析结构体
func (p *parser) InspectStruct(c any) error {
	p.source = c // 保存元数据
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
	// // struct
	// fields, allFields, err := inspectField(rv.Elem(), nil, p.tagOpt)
	// if err != nil {
	// 	return err
	// }
	// p.fields = fields
	// p.allFields = allFields
	// p.setDefaults()
	return p.setDefaultStruct(rv.Elem(), nil)
}

func (p *parser) setDefaultStruct(v reflect.Value, parent *parseField) error {
	// struct
	fields, allFields, err := inspectField(v, parent, p.tagOpt)
	if err != nil {
		return err
	}
	p.fields = fields
	p.allFields = allFields
	return p.setDefaults(allFields)
}

func inspectField(v reflect.Value, parentField *parseField, tagOpt *TagOption) (fields []*parseField, allFields []*parseField, err error) {
	for i := 0; i < v.NumField(); i++ {
		fieldValue := v.Field(i)
		field := v.Type().Field(i)
		if err = isSupportedType(field.Type); err != nil {
			return nil, nil, fmt.Errorf(
				"type of field %v (%v) is not supported: %v",
				field.Name, field.Type, err)
		}

		// reflect.Value
		fieldParse := parseFromField(field, parentField, tagOpt)
		fieldParse.value = fieldValue
		if fieldValue.CanSet() {
			// true: exported field.
			fieldParse.canSet = true
		}
		var (
			t = field.Type
			k = field.Type.Kind()
		)

		// If it is a pointer, it might be nil. Let's fill it with something.
		if k == reflect.Ptr && fieldParse.value.IsNil() {
			fieldParse.value.Set(reflect.New(t.Elem()))
		}
		if k == reflect.Map && fieldParse.value.IsNil() {
			fieldParse.value.Set(reflect.MakeMap(t))
		}

		var anonymousFields []*parseField
		if t.Implements(typeOfTextUnmarshaler) {
			// TextUnmarshaler is a normal type, should not do more.
		} else if k == reflect.Map {
			fieldParse.isMap = true
		} else if k == reflect.Struct {
			fieldParse.isParent = true
			fieldParse.subFields, anonymousFields, err = inspectField(fieldParse.value, fieldParse, tagOpt)
			if err != nil {
				return nil, nil, err
			}
		} else if k == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
			fieldParse.isParent = true
			fieldParse.subFields, anonymousFields, err = inspectField(fieldParse.value.Elem(), fieldParse, tagOpt)
			if err != nil {
				return nil, nil, err
			}
		}
		fields = append(fields, fieldParse)
		allFields = append(allFields, append(anonymousFields, fieldParse)...)
		/**
		// Check for duplicate values for IDs inside the same struct.
		for i := range opts {
			for j := range opts {
				if i != j {
					if opts[i].id == opts[j].id {
						return nil, nil, errors.New(
							"duplicate config variable: \"" + opts[i].id + "\"")
					}
				}
			}
		}
		*/
	}
	return
}

//
func parseFromField(field reflect.StructField, parentField *parseField, tagOpt *TagOption) *parseField {
	resultField := &parseField{}
	ident := field.Tag.Get(tagOpt.IdentTag)
	if len(ident) == 0 {
		ident = strings.ToLower(field.Name)
	}
	resultField.tagValue.Ident = ident

	if parentField == nil {
		resultField.fullIDParts = []string{ident}
	} else {
		resultField.fullIDParts = append(resultField.fullIDParts, parentField.fullIDParts...)
		resultField.fullIDParts = append(resultField.fullIDParts, ident)
	}

	resultField.tagValue.Describe = field.Tag.Get(tagOpt.DescTag)
	resultField.tagValue.Option = field.Tag.Get(tagOpt.OptionTag)
	resultField.tagValue.Valid = field.Tag.Get(tagOpt.ValidTag)
	resultField.tagValue.Default, resultField.tagValue.DefaultSet = field.Tag.Lookup(tagOpt.DefaultTag)
	return resultField
}

func (p *parser) setDefaults(allFields []*parseField) error {
	for _, opt := range allFields {
		if !opt.tagValue.DefaultSet {
			continue
		}

		if opt.isParent {
			// Default values should not be set for nested options.
			return fmt.Errorf("default value specified for nested value '%v'",
				opt.fullID())
		}
		if !opt.value.CanInterface() {
			continue // 不能修改值
		}

		if !isZero(opt.value) {
			// The value has already set before calling goconfig.  In this case,
			// we don't touch it aymore.
			continue
		}

		opt.defaultValue = reflect.New(opt.value.Type()).Elem()
		if isSlice(opt.value) {
			if err := p.parseSlice(opt.defaultValue, opt.tagValue.Default); err != nil {
				return fmt.Errorf(
					"error parsing default value for %v: %v", opt.fullID(), err)
			}
		} else if isMap(opt.value) {
			// TODO: set map default value
			if err := p.parseMap(opt, opt.defaultValue, opt.tagValue.Default); err != nil {
				return fmt.Errorf(
					"error parsing default value for %v: %v", opt.fullID(), err)
			}
		} else {
			if err := p.parseSimpleValue(opt.defaultValue, opt.tagValue.Default); err != nil {
				return fmt.Errorf(
					"error parsing default value for %v: %v", opt.fullID(), err)
			}
		}

		if err := p.setValue(opt.value, opt.defaultValue, p.tagOpt); err != nil {
			return fmt.Errorf("error setting default value for option "+
				"'%v' to '%v': %v", opt.tagValue.Ident, opt.defaultValue, err)
		}
	}

	return nil
}

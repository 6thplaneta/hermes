package hermes

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
)

//It receives value of a tag and a key with key:value format
//It returns the value of key:value
//Example1: GetValueOfTagByKey("searchable,editable,type:int","type") returns "int"
//Example2: GetValueOfTagByKey("searchable,editable,type:int","editable") returns ""
func GetValueOfTagByKey(tag, key string) string {

	if strings.Contains(tag, key) {
		arr1 := strings.Split(tag, ",")
		for _, val := range arr1 {
			if strings.Contains(val, key+":") {
				return strings.Split(val, key+":")[1]
			}
		}
	}
	return ""
}
func GetTypeName(instance interface{}) string {
	typ := reflect.TypeOf(instance)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	return typ.Name()
}
func GetFieldValuesByTag(value interface{}, tg string) map[string]interface{} {
	b := map[string]interface{}{}

	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		if strings.Contains(typeField.Tag.Get("hermes"), tg) {

			value2 := valueField.Interface()
			val := reflect.ValueOf(value2)

			b[typeField.Name] = val.Interface()
		}
	}

	return b
}

//It is a function that receives an object and returns the name of fields that has certain tag value
//Example1: GetFieldsByTag(struct,"hermes","editable") returns all fields of struct that are editable
func GetFieldsByTag(val interface{}, tag, key string) []string {
	arrFields := []string{}
	rval := reflect.ValueOf(val)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	for i := 0; i < rval.NumField(); i++ {
		typeField := rval.Type().Field(i)
		if strings.Contains(typeField.Tag.Get(tag), key) {
			arrFields = append(arrFields, typeField.Name)
		}
	}

	return arrFields
}

/*
* This is a function that returns if the feild exists or not and field type
* It iterates in inner objects to get type of inner feilds (example User.Email for Person Struct)
* @param 	interface{}		instance of struct
* @param 	string 			name of field
* @return	bool 			determines if feild exists or not.
* @return	string 			type of feild
 */
func GetFieldExistanceAndType(val interface{}, name string) (bool, string) {
	tval := reflect.TypeOf(val)
	if tval.Kind() == reflect.Ptr {
		tval = tval.Elem()
	}

	var fieldExists = false
	var typeField reflect.StructField

	arrFields := strings.Split(name, ".")
	for i := 0; i < len(arrFields); i++ {
		if i == 0 {

			typeField, fieldExists = tval.FieldByName(arrFields[i])

		} else {
			if typeField.Type.Kind() == reflect.Slice {
				typeField, fieldExists = typeField.Type.Elem().FieldByName(arrFields[i])
			} else {
				typeField, fieldExists = typeField.Type.FieldByName(arrFields[i])
			}
		}
		if !fieldExists {
			return false, ""
		}
	}

	strHermes := typeField.Tag.Get("hermes")
	hermesType := GetValueOfTagByKey(strHermes, "type")
	kind := typeField.Type.Kind().String()

	var typeOfField string
	//if the kind is struct ,get type of feild from hermes tags
	if kind != "struct" {
		typeOfField = kind
	} else {
		if hermesType != "" {
			typeOfField = hermesType
		}
	}

	return true, typeOfField
}

//It receives array of struct and  and return string splited with spliter
//Example: ConcatArrayToString([1,2,4],",") --> "1,2,3"
func ConcatArrayToString(instance interface{}, feild, spliter string) string {
	col := reflect.ValueOf(instance)
	var res string
	s := col
	for i := 0; i < s.Len(); i++ {
		val := s.Index(i)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		for j := 0; j < val.NumField(); j++ {
			typeField := val.Type().Field(j)
			if strings.EqualFold(typeField.Name, feild) {
				valueField := val.Field(j)
				f := valueField.Interface()
				obj := f.(int)
				res = res + spliter + strconv.Itoa(obj)
			}
		}
	}
	return strings.Replace(res, spliter, "", 1)
}

/*
* This is a function that returns field type
* It iterates in inner objects to get type of inner feilds (example User.Email for Person Struct)
* @param 	interface{}		instance of struct
* @param 	string 			name of field
* @return	reflect.Type 			type of feild
 */
func GetFieldType(instance interface{}, name string) (reflect.Type, error) {

	rval := reflect.ValueOf(instance)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}

	obj := rval
	arrObjects := strings.Split(name, ".")

	for i := 0; i < len(arrObjects); i++ {

		obj = obj.FieldByName(arrObjects[i])
		if !obj.IsValid() {
			return nil, errors.New("Property does not exists: " + arrObjects[i])
		}
	}

	return obj.Type(), nil

}

/*
* This is a function that returns tag value of a field
* @param 	reflect.StructField		instance of struct
* @param 	string 			tag name
* @param 	string 			key
* @return	string 			value of key
 */
func GetTagValueByFeild(field reflect.StructField, tag, key string) (string, bool) {
	strHermes := field.Tag.Get(tag)

	if strings.Contains(strHermes, key) {
		tagResult := GetValueOfTagByKey(strHermes, key)
		return tagResult, true
	}
	return "", false
}

/*
* This is a function that returns tag value of the entered field
* @param 	interface{}		instance of struct
* @param 	string 			field name to get tag value
* @param 	string 			tag name
* @param 	string 			key
* @return	string 			value of key
 */
func GetTagValue(instance interface{}, field, tag, key string) (string, bool) {
	tp := reflect.TypeOf(instance)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	return GetTagValueByType(tp, field, tag, key)
}

func GetTagValueByType(typ reflect.Type, field, tag, key string) (string, bool) {
	if typ.Kind() != reflect.Struct {
		return "", false
	}

	typeField, isexist := typ.FieldByName(field)
	if isexist {
		return GetTagValueByFeild(typeField, tag, key)
	}
	return "", false
}

/*
* This is a function that searches a value in array of struct objects
* @param 	reflect.Value	reflect value of array object
* @param 	string 			It searches value in this field
* @param 	string 			value to search
* @return	bool 			the value exists in array or not
 */
func HasValue(col reflect.Value, key string, value interface{}) bool {
	for i := 0; i < col.Len(); i++ {
		val := col.Index(i)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		valx := val.FieldByName(key).Interface()
		if valx == value {
			return true
		}

	}
	return false
}

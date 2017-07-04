package hermes

import (
	"fmt"
	"github.com/jmoiron/sqlx"

	"errors"
	"reflect"
	"strings"
)

//a function that returns index of closed bracket
//for example closeBracketIndex("(((m)))",0) returns 6
func closeBracketIndex(strval string, openIndex int) (int, error) {
	//
	if strval[openIndex:openIndex+1] != "(" {
		return 0, errors.New("expected ( value")
	}
	var arrOpen []int
	var arrClose []int
	for i := openIndex; i < len(strval); i++ {
		if strval[i:i+1] == "(" {

			arrOpen = append(arrOpen, i)

		} else if strval[i:i+1] == ")" {

			arrClose = append(arrClose, i)
			if len(arrClose) == len(arrOpen) {
				break
			}
		}
	}
	if len(arrClose) != len(arrOpen) {
		return 0, errors.New("not matched paranthesis")
	}
	return arrClose[len(arrClose)-1], nil
}

func strPopulateToArr(strval string) map[string]string {
	var arrPopulate map[string]string
	arrPopulate = make(map[string]string)
	// remove spaces
	str := strings.Replace(strval, " ", "", -1)
	if str[len(str)-1:len(str)] != "," {
		str = str + ","
	}
	for str != "" {
		bindex := strings.Index(str, "(")
		cindex := strings.Index(str, ",")
		if bindex < cindex && bindex != -1 || (cindex == -1 && bindex > -1) {

			k := str[0:bindex]
			b2index, _ := closeBracketIndex(str, bindex)
			v := str[bindex+1 : b2index]
			arrPopulate[k] = v
			str = strings.Replace(str, k+"("+v+"),", "", 1)
			//it has ,

		} else /*it has ,*/ if (cindex < bindex && cindex != -1) || (bindex == -1 && cindex > -1) {
			k := str[0:cindex]
			arrPopulate[k] = ""
			str = strings.Replace(str, k+",", "", 1)
		}
	}
	return arrPopulate
}

// this function receives an object and fills foreign values from foreign resources
// example: city struct has countryـid(one2one relation) this function fetchs country information from db and sets the country object
/*
* This function receives an object and filles foreign values from foreign resources
* Example: city struct has countryـid(one2one relation) this function fetchs country information from db and sets the country object
* @param 	*sqlx.DB	database
* @param 	interface{} 	struct value
* @param	string 			field to populate (example: cities,provinces,student(user))
* @return	error 			error
 */
func PopulateStruct(token string, dbInstance *sqlx.DB, value interface{}, populate string) error {
	// get array of fields for population
	arrPopulate := strPopulateToArr(populate)
	//reflected value of struct
	rval := reflect.ValueOf(value)
	if rval.Kind() == reflect.Ptr {
		//get element if the value is a pointer
		rval = rval.Elem()
	}

	structName := strings.ToLower(rval.Type().Name())
	for strPopulate, innerPopulate := range arrPopulate {

		strPopulate = GetFieldJsonByInst(value, strPopulate)
		// innerPopulate = GetTitle(innerPopulate)

		fieldType, err := GetFieldType(value, strPopulate)

		if err != nil {
			return err
		}

		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}

		//find table name in database
		dbName, _ := GetTagValueByType(fieldType, "Id", "hermes", "dbspace")

		if dbName == "" || fieldType == nil {
			return errors.New("Error: Populate condition wrong, ref dbspace: " + dbName + "filedType: " + fieldType.String())
		}

		//kind of feild
		kind := rval.FieldByName(strPopulate).Kind()
		if kind == reflect.Struct {
			// one 2 one relationship

			refField, _ := GetTagValue(value, strPopulate, "hermes", "key")
			if refField == "" {
				refField = strPopulate + "_Id"
			}

			id := getFeildValue(rval, refField).(int)

			if id != 0 {
				obj, err := CollectionsMap[fieldType].Get(token, id, innerPopulate)

				if err != nil {
					return err
				}
				reflectObj := reflect.ValueOf(obj)
				oo := reflect.Indirect(reflectObj)

				rval.FieldByName(strPopulate).Set(oo)
			} /*else {
				return errors.New("Relation does not have any entry!: " + strPopulate)
			}*/

		} else if kind == reflect.Slice {
			//one2many and many2many relationships
			objId := getFeildValue(rval, "id")
			many2many, _ := GetTagValue(value, strPopulate, "hermes", "many2many")
			fkey, _ := GetTagValue(value, strPopulate, "hermes", "fkey")
			if fkey == "" {
				fkey = fieldType.Name() + "_Id"
			}
			err := embedCollection(token, dbInstance, rval, fieldType, dbName, structName, strPopulate, innerPopulate, objId.(int), many2many, fkey)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

/*
* This function receives a list and fills foreign values from foreign resources
* Example: list of cities have countryـid(one2one relation) this function fetchs country information from db and sets the country object
* @param 	*sqlx.DB		database
* @param 	interface{} 	collection
* @param	string 			field to populate (example: cities,provinces,student(user))
* @return	error 			error
 */
func PopulateCollection(token string, dbInstance *sqlx.DB, collection interface{}, instance interface{}, populate string) error {

	arrPopulate := strPopulateToArr(populate)
	colv := reflect.ValueOf(collection)
	if colv.Kind() == reflect.Ptr {
		colv = colv.Elem()
	}

	colvLen := colv.Len()
	if colvLen == 0 {
		return nil
	}

	for strPopulate, innerPopulate := range arrPopulate {

		strPopulate = GetFieldJsonByInst(instance, strPopulate)

		// innerPopulate = GetTitle(innerPopulate)

		struct_ := colv.Index(0)
		if struct_.Kind() == reflect.Ptr {
			struct_ = struct_.Elem()
		}

		fieldType, err := GetFieldType(instance, strPopulate)
		if err != nil {
			return err
		}
		if fieldType.Kind() == reflect.Slice {
			fieldType = fieldType.Elem()
		}
		ref, _ := GetTagValueByType(fieldType, "Id", "hermes", "dbspace")

		if ref == "" || fieldType == nil {
			return errors.New("Error: Populate condition wrong, ref dbspace: " + ref + "filedType: " + fieldType.String())
		}
		kind := struct_.FieldByName(strPopulate).Kind()

		if kind == reflect.Struct {

			refField, _ := GetTagValue(instance, strPopulate, "hermes", "key")
			if refField == "" {
				refField = strPopulate + "_Id"
			}
			ids := getFieldIdValues(colv, refField)
			params := NewParams(instance)
			params.List["id"] = Filter{Type: "array", Value: ids, FieldType: "int"}

			pg := &Paging{-1, 0, "", ""}
			result, err := CollectionsMap[fieldType].List(SystemToken, params, pg, innerPopulate, "")

			popx := reflect.ValueOf(result)
			if popx.Kind() == reflect.Ptr {
				popx = popx.Elem()
			}
			if err != nil {
				return err
			}

			for k := 0; k < colvLen; k++ {
				for m := 0; m < popx.Len(); m++ {
					colk := colv.Index(k)
					if colk.Kind() == reflect.Ptr {
						colk = colk.Elem()
					}

					cc := colk.FieldByName(refField).Interface().(int)
					row := popx.Index(m)
					if row.Kind() == reflect.Ptr {
						row = row.Elem()
					}
					dd := row.FieldByName("Id").Interface().(int)
					if cc == dd {
						colk.FieldByName(strPopulate).Set(row)
						break
					}

				}
			}
		} else if kind == reflect.Slice {
			for k := 0; k < colvLen; k++ {
				colk := colv.Index(k)
				rval := reflect.Indirect(colk)
				objId := getFeildValue(rval, "id")
				structName := strings.ToLower(rval.Type().Name())

				many2many, _ := GetTagValue(instance, strPopulate, "hermes", "many2many")
				fkey, _ := GetTagValue(instance, strPopulate, "hermes", "fkey")
				if fkey == "" {
					fkey = fieldType.Name() + "_Id"
				}
				err := embedCollection(token, dbInstance, rval, fieldType, ref, structName, strPopulate, innerPopulate, objId.(int), many2many, fkey)
				if err != nil {
					return err
				}
			}

		}

	}

	return nil
}

//this function receives a reflect.value which is a struct and then returns value of specific field
func getFeildValue(rval reflect.Value, field string) interface{} {
	var feildValue interface{}
	for j := 0; j < rval.NumField(); j++ {
		typeField := rval.Type().Field(j)
		if strings.EqualFold(typeField.Name, field) {
			valueField := rval.Field(j)
			feildValue = valueField.Interface()
			break
		}
	}
	return feildValue
}

func embedCollection(token string, dbInstance *sqlx.DB, obj reflect.Value, fieldType reflect.Type, dbspace, structName, populate, innerPopulate string, id int, manyToMany string, fkey string) error {
	slice := reflect.MakeSlice(reflect.SliceOf(fieldType), 0, 0)

	col := reflect.New(slice.Type())
	col.Elem().Set(slice)

	query := ""
	destCollection := CollectionsMap[fieldType]
	destCollectionInst := destCollection.GetInstance()
	if manyToMany == "" {
		//one 2 many
		params := NewParams(destCollectionInst)

		params.List[structName+"_id"] = Filter{Type: "exact", Value: id, FieldType: "int"}

		pg := &Paging{-1, 0, "", ""}
		result, err := destCollection.List(token, params, pg, innerPopulate, "")
		if err != nil {
			return err
		}
		rval := reflect.ValueOf(result)
		if rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}
		collection := reflect.Indirect(rval)

		obj.FieldByName(populate).Set(collection)

	} else {
		_struct := StructsMap[manyToMany]
		middleTableName, _ := GetTagValue(_struct, "Id", "hermes", "dbspace")

		params := NewParams(destCollectionInst)

		query = fmt.Sprintf("select %s from %s where %s_id=%d", fkey, middleTableName, structName, id)
		var arrId []int
		err := dbInstance.Select(&arrId, query)
		if err != nil {
			return err
		}
		if len(arrId) == 0 {
			return nil
		}
		params.List["id"] = Filter{Type: "array", Value: arrId, FieldType: "int"}

		pg := &Paging{-1, 0, "", ""}
		result, err := CollectionsMap[fieldType].List(token, params, pg, innerPopulate, "")

		if err != nil {
			return err
		}
		rval := reflect.ValueOf(result)
		if rval.Kind() == reflect.Ptr {
			rval = rval.Elem()
		}
		//get value of collection Pointer and set obj
		collection := reflect.Indirect(rval)
		obj.FieldByName(populate).Set(collection)
	}

	return nil
}

//it receives a reflect.value which is array of struct and returns array of specified feilds
func getFieldIdValues(col reflect.Value, idString string) []int {
	var res []int
	var newid int
	for i := 0; i < col.Len(); i++ {
		val := col.Index(i)

		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		newid = val.FieldByName(idString).Interface().(int)
		res = append(res, newid)
	}
	return res
}

package hermes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
)

/*
* This function insert parent values(one 2 one) in db before insert a value in databse
* @param 	string 			token
* @param 	*sql.Tx
* @param 	interface{} 	struct value
* @return	error 			error
 */
func PreUnPopulate(token string, trans *sql.Tx, value interface{}) error {
	//reflected value of struct
	rval := reflect.ValueOf(value)
	// ival := rval.Interface()
	if rval.Kind() == reflect.Ptr {
		//get element if the value is a pointer
		rval = rval.Elem()
	}

	tstruct := reflect.TypeOf(value)
	if tstruct.Kind() == reflect.Ptr {
		tstruct = tstruct.Elem()
	}
	for k := 0; k < rval.NumField(); k++ {
		if rval.Field(k).Kind() == reflect.Struct {
			structfield := tstruct.Field(k)
			fieldname := structfield.Name

			_, isOne2one := GetTagValueByFeild(structfield, "hermes", "one2one")
			_, ismany2one := GetTagValueByFeild(structfield, "hermes", "many2one")
			if isOne2one || ismany2one {
				obj := rval.Field(k)
				ival := obj.Addr().Interface()
				emptyObj := reflect.New(obj.Type())
				iemptyObj := emptyObj.Interface()

				outmain, _ := json.Marshal(ival)
				outempty, _ := json.Marshal(iemptyObj)
				if string(outmain) != string(outempty) {
					result, err := CollectionsMap[obj.Type()].Create(token, trans, ival)
					rresult := reflect.ValueOf(result)
					rresult = reflect.Indirect(rresult)
					if err != nil {
						return err
					}
					obj.FieldByName("Id").SetInt(rresult.FieldByName("Id").Int())

					rval.FieldByName(fieldname + "_Id").SetInt(obj.FieldByName("Id").Int())

				}

			}
		}
	}

	return nil
}

/*
* This function insert child values(many 2 many,one 2 many) in db after insert a value in databse
* @param 	string 			token
* @param 	*sql.Tx
* @param 	interface{} 	struct value
* @return	error 			error
 */
func UnPopulate(token string, trans *sql.Tx, value interface{}) error {
	//reflected value of struct
	rval := reflect.ValueOf(value)
	// ival := rval.Interface()
	if rval.Kind() == reflect.Ptr {
		//get element if the value is a pointer
		rval = rval.Elem()
	}

	tstruct := reflect.TypeOf(value)

	if tstruct.Kind() == reflect.Ptr {
		tstruct = tstruct.Elem()

	}

	for k := 0; k < rval.NumField(); k++ {

		if rval.Field(k).Kind() == reflect.Slice {

			obj := rval.Field(k)
			structfield := tstruct.Field(k)
			many2many, _ := GetTagValueByFeild(structfield, "hermes", "many2many")
			_, isOne2many := GetTagValueByFeild(structfield, "hermes", "one2many")
			fkey, _ := GetTagValueByFeild(structfield, "hermes", "fkey")

			if many2many != "" {
				mkey, _ := GetTagValueByFeild(structfield, "hermes", "mkey")
				if mkey == "" {
					mkey = rval.Type().Name() + "_Id"
				}
				//iterate on array for insert into database
				if obj.Len() >= 1 {
					obj0 := obj.Index(0)
					if fkey == "" {
						fkey = obj0.Type().Name() + "_Id"
					}
				}

				for i := 0; i < obj.Len(); i++ {
					mainobj := obj.Index(i)

					if mainobj.FieldByName("Id").Int() == 0 {
						// save main object

						ival1 := mainobj.Addr().Interface()
						_, err := CollectionsMap[mainobj.Type()].Create(token, trans, ival1)

						if err != nil {
							return err
						}
					}

					//SAVE middle table
					id := mainobj.FieldByName("Id").Int()

					rid := rval.FieldByName("Id").Int()
					middleTableName, _ := GetTagValue(StructsMap[many2many], "Id", "hermes", "dbspace")
					insertQ := fmt.Sprintf("insert into %s(%s,%s) VALUES(%d,%d)", middleTableName, mkey, fkey, rid, id)

					_, err := trans.Exec(insertQ)

					if err != nil {
						return err
					}
				}
			} else if isOne2many {
				if fkey == "" {
					fkey = rval.Type().Name() + "_Id"
				}

				for i := 0; i < obj.Len(); i++ {
					mainobj := obj.Index(i)
					mid := int(mainobj.FieldByName("Id").Int())
					ival := mainobj.Addr().Interface()
					if mid == 0 {
						// create object
						mainobj.FieldByName(fkey).Set(rval.FieldByName("Id"))
						_, err := CollectionsMap[mainobj.Type()].Create(token, trans, ival)
						if err != nil {
							return err
						}
					} else {
						// update object to reflect relation
						tableName, tbnameSet := GetTagValue(ival, "Id", "hermes", "dbspace")
						if !tbnameSet {
							return NewError("Invalid", "Object table name unknown"+tableName)
						}
						rid := int(rval.FieldByName("Id").Int())
						updateQ := fmt.Sprintf("update %s set %s=%d where id=%d;", tableName, fkey, rid, mid)
						// fmt.Println("updating relation...", updateQ)
						_, err := trans.Exec(updateQ)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}

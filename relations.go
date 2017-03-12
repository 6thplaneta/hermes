package hermes

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func GetRelation(col Collectionist, id int, rel string) (interface{}, error) {

	rval := reflect.ValueOf(col.GetInstance())
	if rval.Kind() == reflect.Ptr {
		//get element if the value is a pointer
		rval = rval.Elem()
	}
	arrRel := strings.Split(rel, ".")

	// structName := strings.ToLower(rval.Type().Name())
	strField := GetFieldJson(col, arrRel[0])
	// to be tested
	innerField := strings.Join(arrRel[1:], ".")

	//kind of feild
	kind := rval.FieldByName(strField).Kind()
	// if !ok {
	// 	return nil, errors.New("No such a field for get relation: " + strField)
	// }
	// kind := fl.Kind()
	fieldType, err := GetFieldType(col.GetInstance(), strField)
	if err != nil {
		return nil, err
	}

	if fieldType.Kind() == reflect.Slice {
		fieldType = fieldType.Elem()
	}

	//find table name in database
	dbName, _ := GetTagValueByType(fieldType, "Id", "hermes", "dbspace")
	baseDbName, _ := GetTagValueByType(col.GetInstanceType(), "Id", "hermes", "dbspace")

	if dbName == "" || fieldType == nil {
		return nil, errors.New("Error: Populate condition wrong, ref dbspace: " + dbName + "filedType: " + fieldType.String())
	}

	if kind == reflect.Struct {
		// one 2 one relationship
		// fmt.Println("in get relation, kinf is strct", innerField, dbName)

		refField := strField + "_id"

		//get id of relation
		var innerid int
		err := col.GetDataSrc().DB.Get(&innerid, fmt.Sprintf("select %s from %s where id=%d", refField, baseDbName, id))
		if err != nil {
			return nil, err
		}
		var obj interface{}
		var errN error
		if innerid != 0 {
			if innerField == "" {
				obj, errN = CollectionsMap[fieldType].Get(SystemToken, innerid, "")
			} else {
				obj, errN = GetRelation(CollectionsMap[fieldType], innerid, innerField)
			}
			if errN != nil {
				return nil, errN
			}
			reflectObj := reflect.ValueOf(obj)
			oo := reflect.Indirect(reflectObj)

			return oo.Interface(), nil
		} else {
			return nil, errors.New("Relation does not have any entry!: " + strField)
		}

	} else if kind == reflect.Slice {
		return nil, errors.New("Get relation can not return slice of objects")
	}

	return nil, nil
}

func GetOwner(col Collectionist, id int) (interface{}, error) {
	onwerRel, hasOwner := GetTagValue(col.GetInstance(), "Id", "hermes", "owner")
	if !hasOwner {
		return nil, errors.New("owner is not specified!" + col.GetInstanceType().Name())
	}
	return GetRelation(col, id, onwerRel)
}

func GetOwnerAgent(col Collectionist, id int) (*Agent, error) {
	obj, err := GetOwner(col, id)
	if err != nil {
		return nil, err
	}
	objtype := reflect.TypeOf(obj)
	// fmt.Println("owner type is ", objtype, reflect.TypeOf(Agent{}), objtype == reflect.TypeOf(Agent{}))
	if objtype == reflect.TypeOf(Agent{}) {
		ag := obj.(Agent)
		return &ag, nil
	} else {
		newColl, ok := CollectionsMap[objtype]
		if !ok {
			return nil, errors.New("collection is not found for type: " + objtype.Name())
		}
		nid := getFeildValue(reflect.ValueOf(obj), "id").(int)
		return GetOwnerAgent(newColl, nid)
	}
}

package hermes

import (
	"reflect"
	"strings"
)

//searches in CollectionsJsonMap for the json title of a field
//GetFieldJson by collection
func GetFieldJson(col Collectionist, field string) string {
	return CollectionsJsonMap[col.GetInstanceType()][field]

}

//searches in CollectionsJsonMap for the json title of a field
//GetFieldJson by struct
func GetFieldJsonByInst(instance interface{}, field string) string {
	tp := reflect.TypeOf(instance)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}

	return CollectionsJsonMap[tp][field]
}

//searches in CollectionsJsonMap for the json title of a nested fields
//example instance=Agent_Token feild= Agent.Is_Active returns is_active
//GetFieldJson by struct
func GetFieldJsonIndirectByInst(instance interface{}, field string) string {
	tp := reflect.TypeOf(instance)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	kkl := strings.Split(field, ".")
	key := CollectionsJsonMap[tp][kkl[0]]
	if len(kkl) > 1 {
		njson := strings.Join(kkl[1:], ".")
		fl, _ := tp.FieldByName(key)
		fltype := fl.Type
		if fltype.Kind() == reflect.Slice {
			fltype = fltype.Elem()
		}
		ninst := CollectionsMap[fltype].GetInstance()
		return key + "." + GetFieldJsonIndirectByInst(ninst, njson)
	} else {
		return key
	}
}

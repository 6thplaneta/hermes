package hermes

import (
	// "fmt"
	"reflect"
	"strings"
)

func GetFieldJson(col Collectionist, json string) string {
	return CollectionsJsonMap[col.GetInstanceType()][json]

}

func GetFieldJsonByInst(instance interface{}, json string) string {
	tp := reflect.TypeOf(instance)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	// fmt.Println("field***********", CollectionsJsonMap[tp])

	return CollectionsJsonMap[tp][json]
}

func GetFieldJsonIndirectByInst(instance interface{}, json string) string {
	tp := reflect.TypeOf(instance)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	kkl := strings.Split(json, ".")
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

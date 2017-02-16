package hermes

import (
	// "fmt"
	"net/url"
	"strings"
)

func stripKey(inp string) string {
	var result string
	result = strings.TrimSuffix(inp, "$from")
	result = strings.TrimSuffix(result, "$to")
	return result
}

func ReadHttpParams(vals url.Values, instance interface{}) *Params {
	// instance, _ := cont.Coll.GetInstance()
	params := NewParams(instance)
	for k, strParam := range vals {
		strValue := strParam[0]
		if strValue == "" {
			continue
		}
		if k == "$search" {
			params.AddFilter("$$search", Filter{Type: "search", Value: EscapeChars(strValue)})
			// params["$$search"] =
			continue
		}
		if k == "$random" {
			params.AddFilter("$$random", Filter{Type: "random", Value: strValue})
			continue
		}
		kk := strings.Split(k, "$")[0]
		key := GetFieldJsonIndirectByInst(instance, kk)
		// fmt.Println("checking key in params", k, "computed:", key)
		//all filters should exist in struct
		isExist, typeOfField := GetFieldExistanceAndType(instance, key)
		if !isExist {
			continue
		}
		castedVal := CastStrToVal(strValue, typeOfField)
		// all cases and kind of filters apply here
		if strings.Contains(strValue, ",") {
			castedValArr := CastStrToArr(strValue, typeOfField)
			params.List[key] = Filter{Type: "array", Value: castedValArr, FieldType: typeOfField}

		} else if strings.HasSuffix(k, "$from") || strings.HasSuffix(k, "$to") {
			if (params.List[key] == Filter{}) {
				var rangeFilter RangeFilter
				if strings.HasSuffix(k, "$from") {
					rangeFilter = RangeFilter{From: castedVal}
				} else {
					rangeFilter = RangeFilter{To: castedVal}
				}

				params.List[key] = Filter{Type: "range", Value: rangeFilter, FieldType: typeOfField}

			} else {
				rangeFilter := params.List[key].Value.(RangeFilter)
				if strings.HasSuffix(k, "$from") {
					rangeFilter.From = castedVal
				} else {
					rangeFilter.To = castedVal
				}

				params.List[key] = Filter{Type: "range", Value: rangeFilter, FieldType: typeOfField}
			}
		} else {
			params.List[key] = Filter{Type: "exact", Value: castedVal, FieldType: typeOfField}

		}
	}
	return params
}

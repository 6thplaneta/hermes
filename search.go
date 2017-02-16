package hermes

import (
	"encoding/json"
	"fmt"
	"gopkg.in/olivere/elastic.v3"
	"reflect"
	"strconv"
)

func GenerateSearchSQL(searchInstance *SearchClient, instance interface{}, search, baseTable string) (string, error) {
	sqlQuery := ""
	tp := GetTypeName(instance)
	// fmt.Println("Type Name in search ", tp)
	q := elastic.NewQueryStringQuery(search)
	searchResult, err := searchInstance.Elastic.Search().
		Index(searchInstance.IndexName).
		Query(q).
		Type(tp).
		// Sort("user", true).
		// From(0).Size(10).
		Do()
	if err != nil {
		return "", err
	}
	if searchResult.Hits.TotalHits > 0 {
		sqlQuery += baseTable + ".id in (" + GetSearchResultIds(searchResult) + ")"
	}

	return sqlQuery, nil
}

func DoIndexDocument(searchInstance *SearchClient, obj interface{}) error {
	tp := GetTypeName(obj)
	arrSearchables := GetFieldsByTag(obj, "hermes", "searchable")
	if len(arrSearchables) == 0 {
		return nil
	}
	mp := make(map[string]interface{})
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for _, fld := range arrSearchables {
		mp[fld] = val.FieldByName(fld).Interface()
	}
	jsondata, errEncode := json.Marshal(mp)
	if errEncode != nil {
		return errEncode
	}
	fmt.Println("search client", searchInstance)
	id := strconv.Itoa(val.FieldByName("Id").Interface().(int))
	_, errIndex := searchInstance.Elastic.Index().
		Index(searchInstance.IndexName).
		Type(tp).
		Id(id).
		BodyString(string(jsondata)).
		Refresh(true).
		Do()

	fmt.Println("errIndex", errIndex)

	if errIndex != nil {
		return errIndex
	}
	return nil

}

func GetSearchResultIds(sr *elastic.SearchResult) string {
	str := ""
	for _, ht := range sr.Hits.Hits {
		str += ht.Id + ","
	}
	return str[:len(str)-1]
}

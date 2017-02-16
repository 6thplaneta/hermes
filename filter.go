package hermes

import (
	"fmt"
)

type RangeFilter struct {
	From interface{}
	To   interface{}
}

type Filter struct {
	Type      string
	FieldType string
	Value     interface{}
}

type Params struct {
	List map[string]Filter
	Inst interface{}
}

func NewParams(instance interface{}) *Params {
	return &Params{List: make(map[string]Filter), Inst: instance}
}

// key should be struct property not json in all adds

func (prm *Params) Add(key string, value interface{}) *Params {
	isExist, typeOfField := GetFieldExistanceAndType(prm.Inst, key)
	if !isExist {
		fmt.Println("no such a property", key)
		return prm
	}
	prm.List[key] = Filter{Type: "exact", Value: value, FieldType: typeOfField}

	return prm
}

func (prm *Params) AddFilter(key string, f Filter) *Params {
	prm.List[key] = f
	return prm
}

func (prm *Params) AddRange(key string, val1, val2 interface{}) *Params {
	isExist, typeOfField := GetFieldExistanceAndType(prm.Inst, key)
	if !isExist {
		return prm
	}
	rangeFilter := RangeFilter{From: val1, To: val2}
	prm.List[key] = Filter{Type: "range", Value: rangeFilter, FieldType: typeOfField}
	return prm
}

func (prm *Params) AddFrom(key string, val interface{}) *Params {
	isExist, typeOfField := GetFieldExistanceAndType(prm.Inst, key)
	if !isExist {
		return prm
	}
	rangeFilter := RangeFilter{From: val}
	prm.List[key] = Filter{Type: "range", Value: rangeFilter, FieldType: typeOfField}
	return prm
}
func (prm *Params) AddTo(key string, val interface{}) *Params {
	isExist, typeOfField := GetFieldExistanceAndType(prm.Inst, key)
	if !isExist {
		return prm
	}
	rangeFilter := RangeFilter{To: val}
	prm.List[key] = Filter{Type: "range", Value: rangeFilter, FieldType: typeOfField}
	return prm
}

func (prm *Params) AddArray(key string, vals []interface{}) *Params {
	isExist, typeOfField := GetFieldExistanceAndType(prm.Inst, key)
	if !isExist {
		return prm
	}
	prm.List[key] = Filter{Type: "array", Value: vals, FieldType: typeOfField}
	return prm
}

func (prm *Params) Digest(query string) *Params {
	return prm
}

package hermes

import (
	"database/sql"
	"errors"
	"fmt"
	// "github.com/jmoiron/sqlx"
	"net/url"
	"reflect"
	"strings"
)

var trans *sql.Tx

type AuthorizationSetting struct {
	Create string
	Read   string
	List   string
	Update string
	Delete string
	Relate string
}

type CollectionConfig struct {
	Authorizations AuthorizationSetting
	DuplicateError bool
	CacheExpire    int //in seconds
	CheckAccess    func(string, int, string) (bool, error)
}

func (conf *CollectionConfig) SetAuth(create, read, list, update, del, rel string) {
	conf.Authorizations.Create = create
	conf.Authorizations.Read = read
	conf.Authorizations.List = list
	conf.Authorizations.Update = update
	conf.Authorizations.Delete = del
	conf.Authorizations.Relate = rel
	RegisterPermissions([]string{create, read, list, update, del, rel})
}

type Collection struct {
	Instance interface{}
	DataSrc  *DataSrc
	Dbspace  string
	Typ      reflect.Type
	Config   CollectionConfig
}

var CollectionsMap map[reflect.Type]Collectionist = make(map[reflect.Type]Collectionist)
var CollectionsJsonMap map[reflect.Type]map[string]string = make(map[reflect.Type]map[string]string)

func RegisterCollection(col Collectionist) {
	typ := col.GetInstanceType()
	CollectionsMap[typ] = col
}

type Collectionist interface {
	GetInstance() interface{}
	GetInstanceType() reflect.Type
	Delete(string, int) error
	GetDataSrc() *DataSrc
	Create(string, *sql.Tx, interface{}) (interface{}, error)
	Update(string, int, interface{}) error
	List(string, *Params, *Paging, string, string) (interface{}, error)
	ListQuery(string, string) (interface{}, error)
	Get(string, int, string) (interface{}, error)
	Str2Params(string) (*Params, error)
	Report(string, int, int, *Params, string, string, string, string, map[string][]string) (interface{}, error)
	Conf() *CollectionConfig
	Meta() []FieldMeta
	Rel(string, int, string, []int) error
	UnRel(string, int, string, []int) error
	UpdateRel(string, int, string, []int) error
	GetRel(string, int, string) ([]int64, error)
	IndexDocument(interface{})
	// SetConf(string, string, string, string)
}

func (col *Collection) Conf() *CollectionConfig {
	return &col.Config
}

func (col *Collection) Report(token string, page, pageSize int, params *Params, search, sortBy, sortOrder, populate string, aggregation map[string][]string) (interface{}, error) {
	result, err := Report(col.DataSrc, col.Instance, nil, page, pageSize, params, search, sortBy, sortOrder, populate, aggregation)
	return result, err
}

type FieldMeta struct {
	Name         string   `json:"name"`
	Json         string   `json:"json"`
	Editable     bool     `json:"editable"`
	Searchable   bool     `json:"searchable"`
	Type         string   `json:"type"`
	UI_HTML      string   `json:"ui_html"`
	Enums        []string `json:"enums"`
	Show         bool     `json:"show"`
	IsRelation   bool     `json:"is_relation"`
	RelationPath string   `json:"relation_path"`
	List         bool     `json:"list"`
}

func (col *Collection) Meta() []FieldMeta {
	tp := reflect.TypeOf(col.Instance)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	meta := make([]FieldMeta, tp.NumField())
	for indd := 0; indd < tp.NumField(); indd++ {
		tpField := tp.Field(indd)
		fm := FieldMeta{}
		fm.Name = tpField.Name
		fm.Type = tpField.Type.Kind().String()
		fm.Json = tpField.Tag.Get("json")
		if strings.Contains(fm.Json, ",") {
			fm.Json = strings.Split(fm.Json, ",")[0]
		}
		fm.Show = true
		hermesStr := tpField.Tag.Get("hermes")
		fm.Editable = strings.Contains(hermesStr, "editable")
		fm.Searchable = strings.Contains(hermesStr, "searchable")
		if fm.Type == "slice" {
			fm.List = true
			fm.IsRelation = true
		}
		if fm.Type == "struct" {
			hermesType := GetValueOfTagByKey(hermesStr, "type")
			if hermesType == "" {
				fm.IsRelation = true
			} else {
				fm.Type = hermesType
			}
		}
		fm.UI_HTML = GetValueOfTagByKey(hermesStr, "ui-html")
		if fm.UI_HTML == "-" || fm.IsRelation {
			fm.Show = false
		}
		// fill enums
		if strings.Contains(fm.UI_HTML, "enum") {
			fs := strings.Split(fm.UI_HTML, "enum(")[1]
			ss := strings.Split(fs, ")")[0]
			sl := strings.Split(ss, " ")
			for _, en := range sl {
				if en != "" {
					fm.Enums = append(fm.Enums, en)
				}
			}
		}
		if fm.IsRelation {
			ftp := tpField.Type
			if ftp.Kind() == reflect.Slice {
				ftp = ftp.Elem()
			}
			cnt, ok := ControllerMap[ftp]
			if !ok {
				fmt.Println("controller not found for type: ", ftp)
			} else {
				fm.RelationPath = cnt.GetBase()
			}

		}
		meta[indd] = fm

	}
	return meta

}

func FillJsonMap(instance interface{}) {
	tp := reflect.TypeOf(instance)
	if tp.Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	CollectionsJsonMap[tp] = make(map[string]string)
	for i := 0; i < tp.NumField(); i++ {
		fl := tp.Field(i)
		jsontag := fl.Tag.Get("json")
		if jsontag != "" {
			tagspl := strings.Split(jsontag, ",")
			CollectionsJsonMap[tp][tagspl[0]] = fl.Name
		} else {
			CollectionsJsonMap[tp][fl.Name] = fl.Name
		}

	}

}

func NewCollection(instance interface{}, datasrc *DataSrc) (*Collection, error) {
	col := &Collection{}
	col.Instance = instance
	col.DataSrc = datasrc
	dbspace, _ := GetTagValue(col.Instance, "Id", "hermes", "dbspace")

	if dbspace == "" {
		return col, NewError("NotValid", "No db Space defined in Id field!")
	}

	col.Dbspace = dbspace
	col.Config.DuplicateError = true
	col.Config.CacheExpire = 3 //
	typ := reflect.TypeOf(instance)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	col.Typ = typ

	// add collection to map
	CollectionsMap[typ] = col

	// add struct map for midtable
	_, midtable := GetTagValue(col.Instance, "Id", "hermes", "midtable")
	if midtable {
		AddStructMap(typ.Name(), instance)
	}
	FillJsonMap(instance)
	return col, nil

}

func NewDBCollection(instance interface{}, datasrc *DataSrc) (*Collection, error) {
	var err error
	dbInstance := datasrc.DB
	err = AddTable(dbInstance, instance)
	if err != nil {
		return nil, err
	}
	err = SyncSchema(dbInstance, instance)
	if err != nil {
		return nil, err
	}

	err = AddIndexs(dbInstance, instance)
	if err != nil {
		return nil, err
	}
	return NewCollection(instance, datasrc)
}

func (col *Collection) GetDataSrc() *DataSrc {
	return col.DataSrc
}
func (col *Collection) GetInstance() interface{} {
	return col.Instance
}
func (col *Collection) GetInstanceType() reflect.Type {
	return col.Typ
}

// delete object by id
func (col *Collection) Delete(token string, id int) error {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Delete, id, "DELETE", cnf.CheckAccess) {
		return ErrForbidden
	}
	r, err := col.DataSrc.DB.Exec(fmt.Sprintf("delete from %s where Id= %d", col.Dbspace, id))
	if err != nil {
		return err
	}
	rowCount, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rowCount == 0 {
		return ErrNotFound
	}
	return err
}

// insert object
type Rels map[string][]int

func (col *Collection) Create(token string, trans *sql.Tx, obj interface{}) (interface{}, error) {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Create, 0, "CREATE", cnf.CheckAccess) {
		return nil, ErrForbidden
	}
	var err error

	validationError := ValidateStruct(obj)
	if validationError != nil {
		return obj, validationError
	}
	err = PreUnPopulate(token, trans, obj)
	if err != nil {
		return obj, err
	}

	var id int64
	insertQ := getInsertQuery(obj)

	if trans == nil {
		err = col.DataSrc.DB.QueryRow(insertQ).Scan(&id)
	} else {
		err = trans.QueryRow(insertQ).Scan(&id)
	}

	if err != nil {
		if !strings.Contains(err.Error(), Messages["DuplicateIndex"]) || col.Config.DuplicateError {
			return obj, err
		}
	}

	rval := reflect.ValueOf(obj)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	rval.FieldByName("Id").SetInt(id)

	err = UnPopulate(token, trans, obj)
	if err != nil {
		return obj, err
	}
	col.IndexDocument(obj)

	return obj, nil
}

func (col *Collection) IndexDocument(obj interface{}) {
	s := col.DataSrc.Search
	if s.Engine == "elastic" {
		s.ToIndex <- obj
	}
}

// single object update
func (col *Collection) Update(token string, id int, obj interface{}) error {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Update, id, "UPDATE", cnf.CheckAccess) {
		return ErrForbidden
	}

	vl := reflect.ValueOf(obj).Elem()
	if vl.Kind() == reflect.Ptr {
		vl = vl.Elem()
	}

	validationError := ValidateStruct(obj)

	if validationError != nil {
		return validationError
	}

	//property fields
	editables := GetFieldsByTag(obj, "hermes", "editable")
	if len(editables) == 0 {
		return nil
	}
	_, err := col.DataSrc.DB.Exec(getUpdateQuery(editables, obj, id))
	if err != nil {
		return err
	}
	// count, _ := result.RowsAffected()
	// fmt.Println("******************count********** ", count)

	// if count == 0 {
	// 	return ErrNotFound
	// }
	return nil
}

// getCollectionn

func (col *Collection) List(token string, params *Params, pg *Paging, populate, project string) (interface{}, error) {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.List, 0, "LIST", cnf.CheckAccess) {

		return nil, ErrForbidden
	}

	var projectArray []string
	if project != "" {
		projectArray = strings.Split(project, ",")
	}

	result, err := GetCollection(token, col.DataSrc, col.Instance, params, pg, populate, projectArray)
	return result, err
}

//single Object get by id
func (col *Collection) Get(token string, id int, populate string) (interface{}, error) {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Read, id, "GET", cnf.CheckAccess) {
		return nil, ErrForbidden
	}
	inst := reflect.New(col.Typ)
	res := inst.Interface()
	// err := col.DataSrc.DB.C(col.Dbspace).Find(bson.M{"_id": bson.ObjectIdHex(id)}).Select(nil).One(res)

	err := col.DataSrc.DB.Get(res, fmt.Sprintf("select * from %s where Id= %d ", col.Dbspace, id))
	if err != nil && err.Error() == Messages["DbNotFoundError"] {
		return res, ErrNotFound
	}
	if populate != "" {
		err = PopulateStruct(SystemToken, col.DataSrc.DB, res, populate)
	}
	return res, err
}

func (col *Collection) ListQuery(query, populate string) (interface{}, error) {
	var err error
	prms, err := col.Str2Params(query)
	if err != nil {
		return nil, err
	}

	pg := &Paging{1, 1000, "", ""}
	return col.List(SystemToken, prms, pg, populate, "")

}

func (col *Collection) GetRel(token string, origin_id int, field string) ([]int64, error) {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Read, origin_id, "GET", cnf.CheckAccess) {
		return []int64{0}, ErrForbidden
	}

	field = GetFieldJson(col, field)
	ival := reflect.ValueOf(col.Instance)
	if ival.Kind() == reflect.Ptr {
		ival = ival.Elem()
	}

	dval := ival.FieldByName(field)
	if !dval.IsValid() {
		return []int64{0}, errors.New("Field/relation does not exsits: " + field)
	}
	dtyp := dval.Type()
	if dtyp.Kind() == reflect.Slice {
		dtyp = dtyp.Elem()
	}
	origin := ival.Type().Name() + "_Id"
	dest := dtyp.Name() + "_Id"
	many2many, _ := GetTagValue(col.Instance, field, "hermes", "many2many")
	_, isOne2Many := GetTagValueByType(ival.Type(), field, "hermes", "one2many")
	_, isOne2One := GetTagValueByType(ival.Type(), field, "hermes", "one2one")
	_, isMany2One := GetTagValueByType(ival.Type(), field, "hermes", "many2one")
	var arrIds []int64

	if isOne2One || isMany2One {

		dbspace, _ := GetTagValueByType(ival.Type(), "Id", "hermes", "dbspace")
		q := fmt.Sprintf("select %s from %s where id=%d;", dest, dbspace, origin_id)
		err := col.DataSrc.DB.Select(&arrIds, q)

		return arrIds, err
	}

	if many2many != "" {
		dbspace, _ := GetTagValue(StructsMap[many2many], "Id", "hermes", "dbspace")
		q := fmt.Sprintf(" select %s from %s where %s=%d;", dest, dbspace, origin, origin_id)
		err := col.DataSrc.DB.Select(&arrIds, q)

		return arrIds, err
	}
	if isOne2Many {
		dbspace, _ := GetTagValueByType(dtyp, "Id", "hermes", "dbspace")
		q := fmt.Sprintf("select Id from %s where %s=%d;", dbspace, origin, origin_id)
		err := col.DataSrc.DB.Select(&arrIds, q)

		return arrIds, err
	}

	return arrIds, nil
}

func (col *Collection) UpdateRel(token string, origin_id int, field string, arr_dest []int) error {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Relate, origin_id, "RELATE", cnf.CheckAccess) {
		return ErrForbidden
	}

	ival := reflect.ValueOf(col.Instance)
	if ival.Kind() == reflect.Ptr {
		ival = ival.Elem()
	}
	Pfield := GetFieldJson(col, field)
	dval := ival.FieldByName(Pfield)
	if !dval.IsValid() {
		return errors.New("Field/relation does not exsits: " + field)
	}
	dtyp := dval.Type()
	if dtyp.Kind() == reflect.Slice {
		dtyp = dtyp.Elem()
	}
	origin := ival.Type().Name() + "_Id"
	dest := dtyp.Name() + "_Id"

	many2many, _ := GetTagValue(col.Instance, Pfield, "hermes", "many2many")
	_, isOne2Many := GetTagValueByType(ival.Type(), Pfield, "hermes", "one2many")
	_, isOne2One := GetTagValueByType(ival.Type(), Pfield, "hermes", "one2one")
	_, isMany2One := GetTagValueByType(ival.Type(), Pfield, "hermes", "many2one")

	if isOne2One || isMany2One {

		if len(arr_dest) > 1 {
			return errors.New(Messages["ManyItems"])
		}
		dbspace, _ := GetTagValueByType(ival.Type(), "Id", "hermes", "dbspace")
		q := fmt.Sprintf("update %s set %s=0 where id=%d;", dbspace, dest, origin_id)

		_, err := col.DataSrc.DB.Exec(q)
		if err != nil {
			return err
		}
	}
	if many2many != "" {

		dbspace, _ := GetTagValue(StructsMap[many2many], "Id", "hermes", "dbspace")
		q := fmt.Sprintf("delete from  %s where  %s=%d ;", dbspace, origin, origin_id)
		_, err := col.DataSrc.DB.Exec(q)
		if err != nil {
			return err
		}
	}
	if isOne2Many {
		dbspace, _ := GetTagValueByType(dtyp, "Id", "hermes", "dbspace")
		q := fmt.Sprintf("update %s set %s=0 where %s=%d;", dbspace, origin, origin, origin_id)

		_, err := col.DataSrc.DB.Exec(q)
		if err != nil {
			return err
		}
	}

	err := col.Rel(token, origin_id, field, arr_dest)
	if err != nil {
		return err
	}
	return nil
}

func (col *Collection) Rel(token string, origin_id int, field string, arr_dest []int) error {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Relate, origin_id, "RELATE", cnf.CheckAccess) {

		return ErrForbidden

	}

	field = GetFieldJson(col, field)
	ival := reflect.ValueOf(col.Instance)
	if ival.Kind() == reflect.Ptr {
		ival = ival.Elem()
	}

	dval := ival.FieldByName(field)
	if !dval.IsValid() {
		return errors.New("Field/relation does not exsits: " + field)
	}
	dtyp := dval.Type()
	if dtyp.Kind() == reflect.Slice {
		dtyp = dtyp.Elem()
	}
	origin := ival.Type().Name() + "_Id"
	dest := dtyp.Name() + "_Id"
	many2many, _ := GetTagValue(col.Instance, field, "hermes", "many2many")
	_, isOne2Many := GetTagValueByType(ival.Type(), field, "hermes", "one2many")
	_, isOne2One := GetTagValueByType(ival.Type(), field, "hermes", "one2one")
	_, isMany2One := GetTagValueByType(ival.Type(), field, "hermes", "many2one")
	if isOne2One || isMany2One {
		if len(arr_dest) > 1 {
			return errors.New(Messages["ManyItems"])
		}
		dbspace, _ := GetTagValueByType(ival.Type(), "Id", "hermes", "dbspace")
		q := fmt.Sprintf("update %s set %s=%d where id=%d;", dbspace, dest, arr_dest[0], origin_id)
		_, err := col.DataSrc.DB.Exec(q)
		return err
	}

	if many2many != "" {
		dbspace, _ := GetTagValue(StructsMap[many2many], "Id", "hermes", "dbspace")
		q := ""
		for i := 0; i < len(arr_dest); i++ {
			q += fmt.Sprintf(" insert into %s(%s,%s) values(%d,%d); ", dbspace, dest, origin, arr_dest[i], origin_id)
		}
		_, err := col.DataSrc.DB.Exec(q)
		return err
	}
	if isOne2Many {
		dbspace, _ := GetTagValueByType(dtyp, "Id", "hermes", "dbspace")
		q := ""
		for i := 0; i < len(arr_dest); i++ {
			q += fmt.Sprintf("update %s set %s=%d where id=%d;", dbspace, origin, origin_id, arr_dest[i])
		}
		_, err := col.DataSrc.DB.Exec(q)
		return err
	}

	return nil
}

func (col *Collection) UnRel(token string, origin_id int, field string, arr_dest []int) error {
	cnf := col.Conf()
	if !Authorize(token, cnf.Authorizations.Relate, origin_id, "RELATE", cnf.CheckAccess) {
		return ErrForbidden
	}
	ival := reflect.ValueOf(col.Instance)
	if ival.Kind() == reflect.Ptr {
		ival = ival.Elem()
	}
	pfield := GetFieldJson(col, field)
	dval := ival.FieldByName(pfield)
	if !dval.IsValid() {
		return errors.New("Field/relation does not exsits: " + field)
	}
	dtyp := dval.Type()
	if dtyp.Kind() == reflect.Slice {
		dtyp = dtyp.Elem()
	}
	origin := ival.Type().Name() + "_Id"
	dest := dtyp.Name() + "_Id"

	many2many, _ := GetTagValue(col.Instance, pfield, "hermes", "many2many")
	_, isOne2Many := GetTagValueByType(ival.Type(), pfield, "hermes", "one2many")
	_, isOne2One := GetTagValueByType(ival.Type(), pfield, "hermes", "one2one")
	_, isMany2One := GetTagValueByType(ival.Type(), pfield, "hermes", "many2one")

	if isOne2One || isMany2One {
		if len(arr_dest) > 1 {
			return errors.New(Messages["ManyItems"])
		}
		dbspace, _ := GetTagValueByType(ival.Type(), "Id", "hermes", "dbspace")
		q := fmt.Sprintf("update %s set %s=0 where id=%d;", dbspace, dest, origin_id)
		_, err := col.DataSrc.DB.Exec(q)
		return err
	}
	if many2many != "" {

		dbspace, _ := GetTagValue(StructsMap[many2many], "Id", "hermes", "dbspace")
		q := ""
		for i := 0; i < len(arr_dest); i++ {
			q += fmt.Sprintf("delete from  %s where  %s=%d and %s=%d;", dbspace, dest, arr_dest[i], origin, origin_id)
		}
		_, err := col.DataSrc.DB.Exec(q)
		return err
	}
	if isOne2Many {
		dbspace, _ := GetTagValueByType(dtyp, "Id", "hermes", "dbspace")
		q := ""
		for i := 0; i < len(arr_dest); i++ {
			q += fmt.Sprintf("update %s set %s=0 where id=%d;", dbspace, origin, arr_dest[i])
		}
		_, err := col.DataSrc.DB.Exec(q)
		return err
	}

	return nil
}

func (col *Collection) Str2Params(query string) (*Params, error) {
	vals, err := url.ParseQuery(query)
	if err != nil {
		return nil, err
	}
	ins := col.GetInstance()
	return ReadHttpParams(vals, ins), nil

}

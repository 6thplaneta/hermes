package hermes

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ChangeGoTypeToPostgres(typeOfField, dbtype string) string {

	var typ string
	if typeOfField == "string" {
		typ = "text"
	} else if typeOfField == "int" {
		typ = "integer"
	} else if typeOfField == "int64" {
		typ = "bigint"
	} else if typeOfField == "int32" || typeOfField == "rune" {
		typ = "integer"
	} else if typeOfField == "int16" {
		typ = "smallint"
	} else if typeOfField == "bool" {
		typ = "boolean"
	} else if typeOfField == "time" {

		if dbtype == "date" {
			typ = "date"

		} else if dbtype == "time" {
			typ = "time without time zone"

		} else {
			typ = "timestamp with time zone"
		}

	} else if typeOfField == "float32" || typeOfField == "float64" {
		if dbtype != "" {
			if dbtype == "double" {
				typ = "double precision"
			}
		} else {
			typ = "real"
		}
	}

	return typ
}

func getUpdateQuery(obj interface{}, id int) (string, error) {
	editables := GetFieldsByTag(obj, "hermes", "editable")
	if len(editables) == 0 {
		return "", errors.New("No editable field!")
	}

	rval := reflect.ValueOf(obj)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	var strQ, strTypes, strVals, strKeys, updateQuery string
	counter := 1
	for i := 0; i < len(editables); i++ {
		typeField, _ := rval.Type().FieldByName(editables[i])
		dbtag := typeField.Tag.Get("db")

		if dbtag != "-" {
			_, typeOfField := GetFieldExistanceAndType(obj, typeField.Name)
			dbtype, _ := GetTagValue(obj, typeField.Name, "hermes", "dbtype")

			if typeField.Name != "Id" {
				updateQuery = updateQuery + typeField.Name + "=" + "$" + strconv.Itoa(counter) + ","
				counter++

			}

			strKeys = strKeys + typeField.Name + ","
			val := prepareValue(rval.FieldByName(editables[i]).Interface(), typeOfField, dbtype)
			strVals = strVals + val + ","
			strTypes = strTypes + ChangeGoTypeToPostgres(typeOfField, dbtype) + ","

		}

	}
	rndStatement := RandStringRunes(50)
	strQ = " PREPARE " + rndStatement + "(" + strTypes + "int" + ") as "

	dbspace, _ := GetTagValue(obj, "Id", "hermes", "dbspace")
	strQ += " update " + dbspace + " set " + TrimSuffix(updateQuery, ",")
	strQ += " where id=" + "$" + strconv.Itoa(counter) + ";"
	strQ += " EXECUTE " + rndStatement + "(" + strVals + strconv.Itoa(id) + ");"
	// fmt.Println("update query is...", strQ)
	strQ += ` DEALLOCATE all;`
	return strQ, nil
}

func getInsertQuery(obj interface{}) string {
	rval := reflect.ValueOf(obj)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	var strQ, strVals, strKeys, strTypes, strParams string

	counter := 1

	for i := 0; i < rval.NumField(); i++ {
		typeField := rval.Type().Field(i)
		dbtag := typeField.Tag.Get("db")

		if dbtag != "-" {
			_, typeOfField := GetFieldExistanceAndType(obj, typeField.Name)

			if typeField.Name != "Id" {
				defaultv, _ := GetTagValue(obj, typeField.Name, "hermes", "default")
				dbtype, _ := GetTagValue(obj, typeField.Name, "hermes", "dbtype")

				fname := rval.FieldByName(typeField.Name)

				if defaultv != "" {
					if typeOfField == "string" {
						strv := fname.Interface().(string)
						if strv == "" {
							fname.Set(reflect.ValueOf(defaultv))
						}
					} else if typeOfField == "int" {
						intv := fname.Interface().(int)
						if intv == 0 {
							intv, _ := strconv.Atoi(defaultv)
							fname.Set(reflect.ValueOf(intv))
						}
					} else if typeOfField == "int64" {
						intv := fname.Interface().(int64)
						if intv == 0 {
							intv, _ := strconv.ParseInt(defaultv, 10, 64)
							fname.Set(reflect.ValueOf(intv))
						}
					} else if typeOfField == "int32" {
						intv := fname.Interface().(int32)
						if intv == 0 {
							intv, _ := strconv.ParseInt(defaultv, 10, 32)
							fname.Set(reflect.ValueOf(intv))
						}
					} else if typeOfField == "rune" {
						intv := fname.Interface().(rune)
						if intv == 0 {
							intv, _ := strconv.ParseInt(defaultv, 10, 32)
							fname.Set(reflect.ValueOf(intv))
						}
					} else if typeOfField == "int16" {
						intv := fname.Interface().(int16)
						if intv == 0 {
							intv, _ := strconv.ParseInt(defaultv, 10, 16)
							fname.Set(reflect.ValueOf(intv))
						}
					} else if typeOfField == "bool" {
						boolv := fname.Interface().(bool)
						if boolv == false {
							boolv, _ := strconv.ParseBool(defaultv)
							fname.Set(reflect.ValueOf(boolv))
						}
					} else if typeOfField == "float32" || typeOfField == "float64" {
						floatv := fname.Interface().(float64)
						if floatv == 0 {
							floatv, _ := strconv.ParseFloat(defaultv, 64)
							fname.Set(reflect.ValueOf(floatv))
						}
					} else if typeOfField == "time" {
						tim := fname.Interface().(time.Time)

						if defaultv == "$now" && tim.String() == "0001-01-01 00:00:00 +0000 UTC" {
							fname.Set(reflect.ValueOf(time.Now()))
						}
					}
				}

				strKeys = strKeys + typeField.Name + ","
				val := prepareValue(rval.Field(i).Interface(), typeOfField, dbtype)
				strVals = strVals + val + ","
				strTypes = strTypes + ChangeGoTypeToPostgres(typeOfField, dbtype) + ","
				strParams = strParams + "$" + strconv.Itoa(counter) + ","
				counter++
			}
		}

	}

	rndStatement := RandStringRunes(50)

	strQ = " PREPARE " + rndStatement + "(" + TrimSuffix(strTypes, ",") + ") as "

	dbspace, _ := GetTagValue(obj, "Id", "hermes", "dbspace")
	strQ += " insert into " + dbspace + "(" + TrimSuffix(strKeys, ",") + ") values(" + TrimSuffix(strParams, ",") + ") RETURNING id;"

	strQ += " EXECUTE " + rndStatement + "(" + TrimSuffix(strVals, ",") + ");"
	strQ += ` DEALLOCATE all;`
	return strQ
}

func AddTable(db *sqlx.DB, instance interface{}) error {
	dbspace, _ := GetTagValue(instance, "Id", "hermes", "dbspace")

	rval := reflect.ValueOf(instance)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	var strQ string
	strQ = " CREATE TABLE IF NOT EXISTS " + dbspace + " ( "

	for i := 0; i < rval.NumField(); i++ {
		typeField := rval.Type().Field(i)
		dbtag := typeField.Tag.Get("db")
		dbtype, _ := GetTagValue(instance, typeField.Name, "hermes", "dbtype")

		_, typeOfField := GetFieldExistanceAndType(instance, typeField.Name)
		//change golang types to postgres data types

		typ := ChangeGoTypeToPostgres(typeOfField, dbtype)

		if dbtag != "-" {

			if typeField.Name == "Id" {
				strQ = strQ + " " + typeField.Name + " SERIAL PRIMARY KEY " + ","
			} else {
				strQ = strQ + " " + typeField.Name + " " + typ + ","
			}
		}

	}

	strQ = TrimSuffix(strQ, ",") + " ) "

	_, err := db.Exec(strQ)
	return err
}

/*
* connects to database and checks existance of column in table
* @param 	*sql.Db		database
* @param 	interface{} 	struct
* @param	string 			field to search
* @return	bool 			it exists or not
* @return	error 			error
 */
func CheckPostgresColumn(db *sqlx.DB, table interface{}, field string) (bool, error) {
	tableName, _ := GetTagValue(table, "Id", "hermes", "dbspace")

	var result int
	err := db.Get(&result, "SELECT count(*) FROM information_schema.columns WHERE table_name='"+tableName+"' and column_name='"+field+"'")
	if err != nil && err.Error() != Messages["DbNotFoundError"] {
		return false, err
	}

	if result == 0 {
		return false, nil
	}

	return true, nil
}

/*
* connects to database and adds a column in table
* @param 	*sql.DB		database
* @param 	interface{} 	struct
* @param	string 			field to add
* @param	string 			postgres database types (integer,text,real,...)
* @return	error 			error
 */
func AddPostgresColumn(db *sqlx.DB, table interface{}, field, dbtype string) error {
	tableName, _ := GetTagValue(table, "Id", "hermes", "dbspace")

	exists, err := CheckPostgresColumn(db, table, field)

	if err != nil {
		return err
	}

	//set default value
	var defaultVal string
	if dbtype == "text" {
		defaultVal = "''"
	} else if dbtype == "integer" || dbtype == "real" || dbtype == "double precision" {
		defaultVal = "0"
	} else if dbtype == "boolean" {
		defaultVal = "false"
	} else if dbtype == "timestamp with time zone" {
		defaultVal = "'0001-01-01 03:25:44+03:25:44'"
	}

	if !exists {

		strQ := "alter table " + tableName + " add column " + field + " " + dbtype + " default " + defaultVal
		_, err = db.Exec(strQ)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
* connects to database and drops a column in table
* @param 	*sqlx.DB			database
* @param 	interface{} 	struct
* @param	string 			field to remove
* @return	error 			error
 */
func DropPostgresColumn(db *sqlx.DB, table interface{}, field string) error {

	tableName, _ := GetTagValue(table, "Id", "hermes", "dbspace")

	_, err := db.Exec("alter table " + tableName + " drop column " + field)

	if err != nil {
		return err
	}

	return nil
}

/*
* connects to database, adds columns that exists in struct but does not exist in database
* and also removes columns removed from struct
* @param 	*sqlx.DB			database
* @param 	interface{} 	struct
* @return	error 			error
 */
func SyncSchema(db *sqlx.DB, table interface{}) error {
	ival := reflect.ValueOf(table)
	if ival.Kind() == reflect.Ptr {
		ival = ival.Elem()
	}
	tableName, _ := GetTagValue(table, "Id", "hermes", "dbspace")

	structFields := make(map[string]bool)
	for j := 0; j < ival.NumField(); j++ {
		typeField := ival.Type().Field(j)
		field := typeField.Name
		lowerField := strings.ToLower(field)
		//check if the field exists in db or not
		exists, _ := CheckPostgresColumn(db, table, lowerField)
		dbtag := typeField.Tag.Get("db")

		if dbtag != "-" {
			structFields[lowerField] = true
		}
		//if it is not in db and not excluded from struct add field to database
		if !exists && dbtag != "-" {

			typ := ""
			_, typeOfField := GetFieldExistanceAndType(table, field)
			dbtype, _ := GetTagValue(table, typeField.Name, "hermes", "dbtype")
			//change golang types to postgres data types
			if typeOfField == "string" {
				typ = "text"
			} else if typeOfField == "int" {
				typ = "integer"
			} else if typeOfField == "int64" {
				typ = "bigint"
			} else if typeOfField == "int32" {
				typ = "integer"
			} else if typeOfField == "rune" {
				typ = "integer"
			} else if typeOfField == "int16" {
				typ = "smallint"
			} else if typeOfField == "bool" {
				typ = "boolean"
			} else if typeOfField == "time" {
				if dbtype == "time" {
					typ = "time without time zone "

				} else if dbtype == "date" {
					typ = "date"

				} else {
					typ = "timestamp with time zone"

				}
			} else if typeOfField == "float32" || typeOfField == "float64" {
				if dbtype != "" {
					if dbtype == "double" {
						typ = "double precision"
					}
				} else {
					typ = "real"

				}
			}

			AddPostgresColumn(db, table, field, typ)
		}
	}

	var dbcols []string
	//get list of table columns
	err := db.Select(&dbcols, "SELECT Column_Name FROM information_schema.columns WHERE table_name='"+tableName+"'")
	if err != nil {
		return err
	}

	for j := 0; j < len(dbcols); j++ {
		//drop database column if don't exist in struct
		_, exists := structFields[dbcols[j]]
		if !exists {
			DropPostgresColumn(db, table, dbcols[j])
		}
	}

	return nil
}

func AddIndexs(db *sqlx.DB, table interface{}) error {
	ival := reflect.ValueOf(table)
	if ival.Kind() == reflect.Ptr {
		ival = ival.Elem()
	}
	// tableName, _ := GetTagValue(table, "Id", "hermes", "dbspace")

	for j := 0; j < ival.NumField(); j++ {
		typeField := ival.Type().Field(j)
		field := typeField.Name
		_, hasIndex := GetTagValueByFeild(typeField, "hermes", "index")
		if hasIndex {

			err := AddPostgresIndex(db, table, field)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/*
* connects to database and adds index on field
* @param 	*sqlx.DB		database
* @param 	interface{} 	struct
* @param 	string 			colum name
* @return	error 			error
 */
func AddPostgresIndex(db *sqlx.DB, table interface{}, field string) error {

	tableName, _ := GetTagValue(table, "Id", "hermes", "dbspace")
	indexName := "hermes_index_" + tableName + "_" + field

	_, err := db.Exec(" CREATE INDEX " + indexName + " ON " + tableName + " (" + field + ")")
	if err != nil {
		if !strings.Contains(err.Error(), Messages["IndexError"]) {
			return err
		}
	}
	return nil
}

/*
* connects to database and adds index on field
* @param 	*sqlx.DB		database
* @param 	interface{} 	struct
* @param 	string 			colum name
* @return	error 			error
 */
func AddPostgresUIndex(db *sqlx.DB, table interface{}, field string) error {

	var result int
	tableName, _ := GetTagValue(table, "Id", "hermes", "dbspace")
	indexName := tableName + "_index_" + field

	//get indexes count on the feild
	err := db.Get(&result, "select 1 from pg_class t, pg_class i,    pg_index ix,    pg_attribute a where t.oid = ix.indrelid and i.oid = ix.indexrelid"+
		" and a.attrelid = t.oid"+
		" and a.attnum = ANY(ix.indkey)"+
		" and t.relkind = 'r'"+
		" and t.relname ='"+tableName+"'"+
		" and a.attname ='"+field+"'")

	if err != nil && err.Error() != Messages["DbNotFoundError"] {
		return err
	}
	//add index if there is not index on field
	if err != nil && err.Error() == Messages["DbNotFoundError"] {
		_, err = db.Exec("CREATE UNIQUE INDEX " + indexName + " ON " + tableName + " (" + field + ")")
		if err != nil {
			return err
		}
	}
	return nil
}

/*
* connects to database and adds a foreign key to table
* @param 	*sqlx.DB		database
* @param 	interface{} 	struct
* @param 	string 			foreign key name in table
* @param 	interface{} 	foreign table
* @param 	string 			column name in foreign table(primary key)
* @param 	bool 			cascade delete
* @return	error 			error
 */
func AddPostgresForeignKey(db *sqlx.DB, local interface{}, localFeild string, foreign interface{}, foreignFeild string, cascadeDelete bool) error {
	var result int
	localName, _ := GetTagValue(local, "Id", "hermes", "dbspace")

	foreginName, _ := GetTagValue(foreign, "Id", "hermes", "dbspace")
	err := db.Get(&result, "select 1 FROM pg_constraint WHERE conname = '"+localName+"_"+localFeild+"_fkey' ")

	if err != nil && !strings.Contains(err.Error(), Messages["DbNotFoundError"]) {
		return err
	}
	//add foreign key if does not exist
	if err != nil && strings.Contains(err.Error(), Messages["DbNotFoundError"]) {
		strQ := "ALTER TABLE " + localName + " ADD foreign key (" + localFeild + ") REFERENCES " + foreginName + "(" + foreignFeild + ")"
		if cascadeDelete {
			strQ = strQ + " on delete cascade"
		}

		_, err = db.Exec(strQ)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
* connects to database and gets list of ollection
* @param 	string			token for authorization
* @param 	*sqlx.DB		database
* @param 	interface{} 	this is a new instance of the collection struct(Person{},User{})
* @param 	[]string 		feilds to select splited with , ('*' and '' selects all)
* @param 	int 			page number
* @param 	int 			page size
* @param 	Params 			filter values map[string][interface{}]
* @param 	string 			search value searched in searchable fields
* @param 	string 			sort by
* @param 	string 			sort order (asc ,desc)
* @param 	string 			it s a feild which refers to struct ,its data will be filled by its own table
* @param 	string 			add a condition to created condition
* @return	interface{} 	list of collection
* @return	PageInfo		pageinfo includes page number, page size, total pages and total rows
* @return	error 			error
 */

func GetCollection(token string, datasrc *DataSrc, instance interface{}, params *Params, pg *Paging, populate string, fields []string) (interface{}, error) {
	// fmt.Println("getcollection called with params and instance type:", params, reflect.TypeOf(instance))
	db := datasrc.DB
	myType := reflect.TypeOf(instance)
	if myType.Kind() == reflect.Ptr {
		myType = myType.Elem()
	}
	page := pg.Num
	pageSize := pg.Size
	sortBy := pg.Sort
	sortOrder := pg.Order

	var search string
	var strRandom string
	var random bool

	paramsList := params.List
	if paramsList["$$search"].Value != nil {
		search = paramsList["$$search"].Value.(string)
	}

	if paramsList["$$random"].Value != nil {
		strRandom = paramsList["$$random"].Value.(string)

	}

	random, _ = strconv.ParseBool(strRandom)
	slice := reflect.MakeSlice(reflect.SliceOf(myType), 0, 0)
	x := reflect.New(slice.Type())
	x.Elem().Set(slice)

	baseTable, _ := GetTagValue(instance, "Id", "hermes", "dbspace")

	// generate select query

	sqlQuery := generateQuery(datasrc, instance, baseTable, fields, params, nil, page, pageSize, search, sortBy, sortOrder, random)

	//if page number is 0 return pages information includes page number, page size, total pages and total rows
	//if page number is greater than 0 fetch data from database
	err := db.Select(x.Interface(), sqlQuery)

	ress := x.Interface()
	if populate != "" {

		err = PopulateCollection(SystemToken, db, ress, populate)
	}
	return ress, err
}

func Report(datasrc *DataSrc, instance interface{}, fields []string, page, pageSize int, params *Params, search, sortBy, sortOrder, populate string, aggregation map[string][]string) (interface{}, error) {
	db := datasrc.DB
	baseTable, _ := GetTagValue(instance, "Id", "hermes", "dbspace")

	sqlQuery := generateQuery(datasrc, instance, baseTable, fields, params, aggregation, page, pageSize, search, sortBy, sortOrder, false)

	// var result interface{}
	var err error
	keys := make([]string, 0)
	for k, _ := range aggregation {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	mmp := make(map[int]string, 0)
	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}

	results := make([]interface{}, 0)
	cols, _ := rows.Columns()
	queryValuesPtr := make([]interface{}, len(cols))

	for rows.Next() {
		var ind int = 0
		myFields := make(map[string]interface{})
		for _, key := range keys {
			fields := aggregation[key]
			if key == "group_by" {
				groupByField := GetFieldJsonByInst(instance, fields[0])
				ex, tp := GetFieldExistanceAndType(instance, groupByField)
				if !ex {
					return nil, errors.New("group_by field does not exists")
				}

				if tp == "string" {
					myFields[groupByField] = ""
					queryValuesPtr[ind] = new(string)
				} else if tp == "int" {
					myFields[groupByField] = 0
					queryValuesPtr[ind] = new(int)
				} else if tp == "int64" {
					myFields[groupByField] = 0
					queryValuesPtr[ind] = new(int64)
				} else if tp == "int32" {
					myFields[groupByField] = 0
					queryValuesPtr[ind] = new(int32)
				} else if tp == "rune" {
					myFields[groupByField] = 0
					queryValuesPtr[ind] = new(rune)
				} else if tp == "int16" {
					myFields[groupByField] = 0
					queryValuesPtr[ind] = new(int16)
				} else if tp == "bool" {
					myFields[groupByField] = false
					queryValuesPtr[ind] = new(bool)
				} else if tp == "float32" {
					myFields[groupByField] = 0.0
					queryValuesPtr[ind] = new(float32)
				} else if tp == "float64" {
					myFields[groupByField] = 0.0
					queryValuesPtr[ind] = new(float64)
				}
				mmp[ind] = groupByField
				ind += 1

			} else {

				for _, field := range fields {
					field_key := strings.ToLower(key + "_" + field)
					myFields[field_key] = 0
					mmp[ind] = field_key
					queryValuesPtr[ind] = new(float32)
					ind += 1
				}
			}

		}

		rows.Scan(queryValuesPtr...)

		for in, fk := range mmp {
			myFields[fk] = queryValuesPtr[in]
		}

		results = append(results, myFields)

	}

	return &results, nil
}

func prepareValue(val interface{}, tp, dbtype string) string {
	var cval string

	cval = CastToStr(val, tp, dbtype)
	if tp == "string" {
		cval = EscapeChars(cval)
		cval = "'" + cval + "'"

	} else if tp == "time" {
		cval = "'" + cval + "'"

	}

	return cval
}
func prepareValueArr(val interface{}, tp, dbtype string) string {
	var cval string
	cval = CastArrToStr(val, tp, dbtype)
	return cval
}

func hasJoined(strKey string, madeJoins []string) bool {
	var reversed string
	arr := strings.Split(strKey, ",")
	if len(arr) == 2 {
		reversed = arr[1] + "," + arr[0]
	}

	for _, val := range madeJoins {
		if val == strKey {
			return true
		}
		if val == reversed {
			return true
		}
	}
	return false

}

/* use make joins and wheres from query string */
func makeJoin(instance interface{}, strKey string, params Params, madeJoins *[]string) (string, string) {
	//reflected value of struct
	// rval := reflect.ValueOf(instance)
	// // ival := rval.Interface()
	// if rval.Kind() == reflect.Ptr {
	// 	//get element if the value is a pointer
	// 	rval = rval.Elem()
	// }

	baseTable, _ := GetTagValue(instance, "Id", "hermes", "dbspace")
	parts := strings.Split(baseTable+"."+strKey, ".")

	length := len(parts)
	query := ""
	clauses := ""
	// obj := rval
	var field reflect.StructField

	// var field reflect.StructField
	typ := reflect.TypeOf(instance)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	for i, _ := range parts {

		if i < length-2 {

			//current part

			if typ.Kind() == reflect.Slice {
				typ = typ.Elem()
			}
			curTable, _ := GetTagValueByType(typ, "Id", "hermes", "dbspace")

			//next part
			nextPart := parts[i+1]
			field, _ = typ.FieldByName(nextPart)

			nextType := field.Type

			if nextType.Kind() == reflect.Struct {
				nextTable, _ := GetTagValueByType(nextType, "Id", "hermes", "dbspace")

				if !hasJoined(nextTable+","+curTable, *madeJoins) {
					key, _ := GetTagValueByType(typ, nextPart, "hermes", "key")
					if key == "" {
						key = nextPart + "_id"
					}
					query = query + " JOIN " + nextTable + " on " + curTable + "." + key + " = " + nextTable + ".Id"
					*madeJoins = append(*madeJoins, nextTable+","+curTable)
				}

			} else if nextType.Kind() == reflect.Slice {

				many2many, _ := GetTagValueByType(typ, nextPart, "hermes", "many2many")
				_, isOne2Many := GetTagValueByType(typ, nextPart, "hermes", "one2many")

				if many2many != "" {

					midTable, _ := GetTagValue(StructsMap[many2many], "Id", "hermes", "dbspace")
					nextTable, _ := GetTagValueByType(nextType.Elem(), "Id", "hermes", "dbspace")

					if !hasJoined(midTable+","+curTable, *madeJoins) {
						query = query + " JOIN " + midTable + " on " + curTable + "." + "Id" + " = " + midTable + "." + typ.Name() + "_id"
						*madeJoins = append(*madeJoins, midTable+","+curTable)
					}

					if !hasJoined(nextTable+","+midTable, *madeJoins) {
						query = query + " JOIN " + nextTable + " on " + nextTable + "." + "Id" + " = " + midTable + "." + nextType.Elem().Name() + "_id"
						*madeJoins = append(*madeJoins, nextTable+","+midTable)
					}
					fieldName := getConditionKey(instance, strKey, baseTable)
					sqlQuery := getCluaseVal(fieldName, params.List[strKey])
					if sqlQuery != "" {
						clauses = clauses + " And " + midTable + "." + nextType.Elem().Name() + "_id  in  (select id from " + nextTable + " where 1=1  And " + sqlQuery + ") "

					}
				} else if isOne2Many {
					nextTable, _ := GetTagValueByType(nextType.Elem(), "Id", "hermes", "dbspace")
					if !hasJoined(nextTable+","+curTable, *madeJoins) {
						fkey, _ := GetTagValueByType(typ, nextPart, "hermes", "fkey")
						if fkey == "" {
							fkey = typ.Name() + "_id"
						}
						query = query + " JOIN " + nextTable + " on " + curTable + "." + "Id" + " = " + nextTable + "." + fkey
						*madeJoins = append(*madeJoins, nextTable+","+curTable)
					}
				}

			}
			typ = nextType

		}
	}
	return query, clauses
}

func getCluaseVal(fieldName string, param Filter) string {
	con := ""
	switch param.Type {
	case "array":
		strVal := prepareValueArr(param.Value, param.FieldType, "")
		if strVal != "" {
			con = fieldName + " IN (" + prepareValueArr(param.Value, param.FieldType, "") + ")"
		}
	case "range":
		rangeFilter, _ := param.Value.(RangeFilter)
		if rangeFilter.From != nil {
			fromval := prepareValue(rangeFilter.From, param.FieldType, "")
			con = fieldName + " >= " + fromval
		}
		if rangeFilter.To != nil {
			toval := prepareValue(rangeFilter.To, param.FieldType, "")
			con = fieldName + " < " + toval
		}
		if rangeFilter.From != nil && rangeFilter.To != nil {
			fromval := prepareValue(rangeFilter.From, param.FieldType, "")
			toval := prepareValue(rangeFilter.To, param.FieldType, "")
			con = fieldName + " >= " + fromval + " and " + fieldName + " < " + toval

		}
	case "exact":
		con = fieldName + "=" + prepareValue(param.Value, param.FieldType, "")
	}
	return con
}

func getCluaseValues(fieldName string, param Filter) (string, string) {
	con := ""
	switch param.Type {
	case "array":

		con = prepareValueArr(param.Value, param.FieldType, "")
		return "array[" + con + "]", ""
	case "range":

		rangeFilter, _ := param.Value.(RangeFilter)
		if rangeFilter.From != nil && rangeFilter.To != nil {

			fromval := prepareValue(rangeFilter.From, param.FieldType, "")
			toval := prepareValue(rangeFilter.To, param.FieldType, "")
			con = fromval
			con1 := toval
			return con, con1

		}
		if rangeFilter.From != nil {
			fromval := prepareValue(rangeFilter.From, param.FieldType, "")
			con = fromval
			return con, ""
		}
		if rangeFilter.To != nil {
			toval := prepareValue(rangeFilter.To, param.FieldType, "")
			con = toval
			return con, ""

		}

	case "exact":
		con = prepareValue(param.Value, param.FieldType, "")
		return con, ""

	}
	return "", ""
}

func getCluaseValParams(fieldName string, param Filter, counter *int) string {
	con := ""
	switch param.Type {
	case "array":
		strVal := prepareValueArr(param.Value, param.FieldType, "")
		if strVal != "" {
			con = fieldName + " = Any ($" + strconv.Itoa(*counter) + ")"
		}
	case "range":
		rangeFilter, _ := param.Value.(RangeFilter)
		if rangeFilter.From != nil {
			con = fieldName + " >= " + "$" + strconv.Itoa(*counter)
		}
		if rangeFilter.To != nil {
			con = fieldName + " < " + "$" + strconv.Itoa(*counter)
		}
		if rangeFilter.From != nil && rangeFilter.To != nil {
			con = fieldName + " >= " + "$" + strconv.Itoa(*counter)
			*counter++
			con = con + " and " + fieldName + " < " + "$" + strconv.Itoa(*counter)

		}
	case "exact":
		con = fieldName + "=" + "$" + strconv.Itoa(*counter)
	}
	return con
}
func getConditionKey(instance interface{}, key string, baseTable string) string {

	result := ""

	if !strings.Contains(key, ".") {
		return baseTable + "." + key
	} else {

		var field reflect.StructField
		typ := reflect.TypeOf(instance)
		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
		parts := strings.Split(key, ".")
		length := len(parts)
		for i := 0; i < length-1; i++ {
			if typ.Kind() == reflect.Slice {
				typ = typ.Elem()
			}
			field, _ = typ.FieldByName(parts[i])
			typ = field.Type

			if i == length-2 {
				if typ.Kind() == reflect.Slice {
					typ = typ.Elem()
				}
				dbspace, _ := GetTagValueByType(typ, "Id", "hermes", "dbspace")
				result = dbspace + "." + parts[length-1]
				return result

			}

		}
	}

	return ""
}

func generateQuery(datasrc *DataSrc, instance interface{}, baseTable string, fields []string, params *Params, aggregation map[string][]string, page, pageSize int, search, sortBy, sortOrder string, random bool) string {
	// fmt.Println("params for list filter...", params)
	search = EscapeChars(search)
	sortBy = EscapeChars(sortBy)
	sortOrder = EscapeChars(sortOrder)

	rval := reflect.ValueOf(instance)
	// ival := rval.Interface()
	if rval.Kind() == reflect.Ptr {
		//get element if the value is a pointer
		rval = rval.Elem()
	}
	/* MY OWN SQL QUERY */

	sqlQuery := "SELECT "
	groupBy := false
	groupByField := ""

	/* Sort values */
	keys := make([]string, 0)
	midleCons := ""
	for k, _ := range aggregation {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	/* End of sort */
	if aggregation != nil {
		// first := true
		for _, key := range keys {
			agg_fields := aggregation[key]
			// length := len(agg_fields)
			if key == "group_by" {
				groupBy = true
				groupByField = GetFieldJsonByInst(instance, agg_fields[0])
				fields = append(fields, groupByField+" AS "+groupByField)

			} else {

				for _, field := range agg_fields {
					str := key + "("
					if !strings.Contains(field, ".") {
						str += baseTable + "."
					}
					str += field + ") AS " + key + "_" + field + " "
					fields = append(fields, str)
					// myFields[strings.ToLower(key+"_"+field)] = ""

				}
			}
			// first = false
		}
		sqlQuery += " distinct "
	}

	if fields == nil {
		fields = []string{"*"}
	}

	for i, field := range fields {
		if i > 0 {
			sqlQuery += ", "
		}
		if !strings.Contains(field, ".") {
			sqlQuery += baseTable + "."
		}
		sqlQuery += field + " "
	}

	madeJoins := make([]string, 0)
	sqlQuery += " FROM " + baseTable
	for key, _ := range params.List {
		if strings.Contains(key, ".") { // The key refers to another table

			join, midleCon := makeJoin(instance, key, *params, &madeJoins)

			midleCons += midleCon
			sqlQuery += join
		}
	}

	counter := 1
	var strTypes, strVals string
	customWhere := ""
	sqlQuery += " WHERE 1=1 "
	paramsList := params.List
	for key, v := range paramsList {
		if strings.Contains(key, "$$") {
			if key == "$$custom" {
				customWhere = v.Value.(string)
			}
			continue
		}
		// fmt.Println("key to getConditionKey... ", key)
		//fieldname like customers.agent.Id tables and at end property
		fieldName := getConditionKey(instance, key, baseTable)
		wCluase := getCluaseValParams(fieldName, v, &counter)
		// fmt.Println("fieldname, wclause, gotopost", fieldName, wCluase, ChangeGoTypeToPostgres(v.FieldType))
		counter++
		if wCluase != "" {
			sqlQuery += " and " + wCluase

		}

		rangeFilter, _ := v.Value.(RangeFilter)

		if v.Type == "array" {

			strTypes = strTypes + ChangeGoTypeToPostgres(v.FieldType, v.DBType) + "[],"

		} else if v.Type == "range" && rangeFilter.From != nil && rangeFilter.To != nil {

			//add two types for prepared statement parameters
			strTypes = strTypes + ChangeGoTypeToPostgres(v.FieldType, v.DBType) + ","
			strTypes = strTypes + ChangeGoTypeToPostgres(v.FieldType, v.DBType) + ","

		} else {
			strTypes = strTypes + ChangeGoTypeToPostgres(v.FieldType, v.DBType) + ","

		}
		val1, val2 := getCluaseValues(fieldName, v)

		if val1 != "" {
			strVals = strVals + val1 + ","

		}
		if val2 != "" {
			strVals = strVals + val2 + ","

		}

	}
	// apply search
	if search != "" {

		if datasrc.Search.Engine == "sql" {
			sqlQuery += " AND ( "
			for i, str := range getSearchableValues(instance, search, counter) {
				if i > 0 {
					sqlQuery += " OR "
				}
				sqlQuery += str
				strTypes = strTypes + "text,"
				strVals = strVals + "'" + search + "%',"
				counter++
			}
			sqlQuery += " ) "
		} else if datasrc.Search.Engine == "elastic" {
			searchsql, errSearchSQL := GenerateSearchSQL(datasrc.Search, instance, search, baseTable)
			// fmt.Println("search sql is ", searchsql)
			if errSearchSQL == nil {
				sqlQuery += " AND  " + searchsql
			} else {
				fmt.Println("Error in search generation: ", errSearchSQL)
			}
		}
	}

	if midleCons != "" {
		sqlQuery += midleCons
	}

	if customWhere != "" {
		sqlQuery += " And " + customWhere
	}

	if sortBy == "" {
		sortBy = baseTable + ".id "
	}

	if sortOrder == "" {
		sortOrder = " ASC "
	}

	if aggregation == nil {
		if random == true {
			sqlQuery += " ORDER BY  random() "

		} else {
			sqlQuery += " ORDER BY " + sortBy + " " + sortOrder
		}
	}

	if groupBy {
		sqlQuery += " GROUP BY " + baseTable + "." + groupByField
	}

	if page > 0 {
		offset := (page - 1) * pageSize
		sqlQuery += " LIMIT " + strconv.Itoa(pageSize) + " OFFSET " + strconv.Itoa(offset)
	}
	/* END OF MY OWN SQL QUERY */

	var prepared string
	rndStatement := RandStringRunes(100)
	if strTypes != "" {
		prepared += " PREPARE " + rndStatement + "(" + TrimSuffix(strTypes, ",") + ") as "
	}
	prepared += sqlQuery
	if strTypes != "" {

		prepared += " ; EXECUTE " + rndStatement + "(" + TrimSuffix(strVals, ",") + ");"
		prepared += ` DEALLOCATE all;`

	}
	// prepared = prepared + `DEALLOCATE "` + rndStatement + `";`

	return prepared
}

func getSearchableValues(value interface{}, search string, counter int) []string {
	var searchs []string
	if search != "" {
		arrSearchables := GetFieldsByTag(value, "hermes", "searchable")
		for i := 0; i < len(arrSearchables); i++ {
			searchs = append(searchs, " lower("+arrSearchables[i]+") LIKE lower($"+strconv.Itoa(counter)+") ")
		}
	}
	return searchs
}

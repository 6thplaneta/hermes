package hermes

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCheckPostgresColumn(t *testing.T) {
	assert.NoError(t, addTempTables())

	var dbcols []string
	e = DBTest().Select(&dbcols, "SELECT Column_Name FROM information_schema.columns WHERE table_name='persons'")
	assert.NoError(t, e)

	assert.Contains(t, dbcols, "id")
	assert.Contains(t, dbcols, "name")
	assert.NotContains(t, dbcols, "middleـname")
	assert.NotContains(t, dbcols, "notexist")

	assert.NoError(t, rmTempTables())

}

func TestAddPostgresColumn(t *testing.T) {
	assert.NoError(t, addTempTables())

	e = AddPostgresColumn(DBTest(), Person{}, "age", "integer")
	assert.NoError(t, e)

	var dbcols []string
	e = DBTest().Select(&dbcols, "SELECT Column_Name FROM information_schema.columns WHERE table_name='persons'")
	assert.NoError(t, e)
	assert.Contains(t, dbcols, "id")
	assert.Contains(t, dbcols, "name")
	assert.Contains(t, dbcols, "age")
	assert.NotContains(t, dbcols, "notexist")
	assert.NotContains(t, dbcols, "middle_name")

	assert.NoError(t, rmTempTables())

}

func TestDropPostgresColumn(t *testing.T) {
	assert.NoError(t, addTempTables())

	e = DropPostgresColumn(DBTest(), Person{}, "name")
	assert.NoError(t, e)

	var dbcols []string
	e = DBTest().Select(&dbcols, "SELECT Column_Name FROM information_schema.columns WHERE table_name='persons'")
	assert.NoError(t, e)
	assert.Contains(t, dbcols, "id")
	assert.NotContains(t, dbcols, "name")
	assert.NotContains(t, dbcols, "middle_name")
	assert.NoError(t, rmTempTables())

}

func TestSyncSchema(t *testing.T) {
	assert.NoError(t, addTempTables())

	type Person1 struct {
		Id int ` hermes:"dbspace:persons"`
		//delete name column
		// Name string `json:"name"`
		MiddleـName string `db:"-"`
		//add new columns
		Family      string
		Age         int
		Female      bool
		Birth_Date  time.Time `hermes:"type:time"`
		Average     float64
		Not_Include int `db:"-"`
	}
	e = SyncSchema(DBTest(), Person1{})
	assert.NoError(t, e)

	type Column struct {
		Column_Name string
		Data_Type   string
	}

	var dbcols []string
	e = DBTest().Select(&dbcols, "SELECT Column_Name FROM information_schema.columns WHERE table_name='persons'")
	assert.NoError(t, e)

	assert.Contains(t, dbcols, "family")
	assert.Contains(t, dbcols, "age")
	assert.Contains(t, dbcols, "female")
	assert.Contains(t, dbcols, "birth_date")
	assert.Contains(t, dbcols, "average")

	assert.NotContains(t, dbcols, "name")
	assert.NotContains(t, dbcols, "middleـname")
	assert.NotContains(t, dbcols, "not_include")

	var Columns []Column
	e = DBTest().Select(&Columns, "SELECT Column_Name,Data_Type FROM information_schema.columns WHERE table_name='persons'")
	assert.NoError(t, e)

	for i := 0; i < len(Columns); i++ {
		if Columns[i].Column_Name == "family" {
			assert.Equal(t, "text", Columns[i].Data_Type)
		} else if Columns[i].Column_Name == "age" {
			assert.Equal(t, "integer", Columns[i].Data_Type)
		} else if Columns[i].Column_Name == "female" {
			assert.Equal(t, "boolean", Columns[i].Data_Type)
		} else if Columns[i].Column_Name == "birth_date" {
			assert.Equal(t, "timestamp with time zone", Columns[i].Data_Type)
		} else if Columns[i].Column_Name == "average" {
			assert.Equal(t, "real", Columns[i].Data_Type)

		}

	}
	assert.NoError(t, rmTempTables())

}

func TestAddPostgresIndex(t *testing.T) {
	assert.NoError(t, addTempTables())

	e = AddPostgresUIndex(DBTest(), Person{}, "name")
	assert.NoError(t, e)

	var result int
	e = DBTest().Get(&result, "select 1 from pg_class t, pg_class i,    pg_index ix,    pg_attribute a where t.oid = ix.indrelid and i.oid = ix.indexrelid"+
		" and a.attrelid = t.oid"+
		" and a.attnum = ANY(ix.indkey)"+
		" and t.relkind = 'r'"+
		" and t.relname ='persons'"+
		" and a.attname ='name'")

	assert.NoError(t, e)
	assert.Equal(t, 1, result)
	assert.NoError(t, rmTempTables())

}

func TestAddPostgresForeignKey1(t *testing.T) {
	assert.NoError(t, addTempTables())

	//cascade delete

	//test creating foreign key successfully start

	e = AddPostgresForeignKey(DBTest(), Person{}, "gender_id", Gender{}, "id", true)
	assert.NoError(t, e)
	var result int
	e = DBTest().Get(&result, "select 1 FROM pg_constraint WHERE conname = 'persons_gender_id_fkey'")
	assert.NoError(t, e)
	assert.Equal(t, 1, result)
	//test creating foreign key successfully end

	//
	gender := Gender{Title: "female"}
	genderColl.Create(SystemToken, nil, &gender)
	assert.NoError(t, e)

	_, e = DBTest().Exec("insert into persons(id,name,gender_id) values(1,'mahsa',1);")
	assert.NoError(t, e)

	//remove gender for test cascade delete
	_, e = DBTest().Exec("delete from gender where id=1")
	assert.NoError(t, e)

	result = 0
	e = DBTest().Get(&result, "select count(1) FROM persons")
	assert.NoError(t, e)
	//the person should be deleted because the key is cascade
	assert.Equal(t, 0, result)
	assert.NoError(t, rmTempTables())

}

func TestAddPostgresForeignKey2(t *testing.T) {
	assert.NoError(t, addTempTables())

	//not cascade delete

	//test creating foreign key successfully start
	e = AddPostgresForeignKey(DBTest(), Person{}, "gender_id", Gender{}, "id", false)
	assert.NoError(t, e)
	var result int
	e = DBTest().Get(&result, "select 1 FROM pg_constraint WHERE conname = 'persons_gender_id_fkey'")
	assert.NoError(t, e)
	assert.Equal(t, 1, result)
	//test creating foreign key successfully end

	gender := Gender{Title: "female"}
	genderColl.Create(SystemToken, nil, &gender)
	assert.NoError(t, e)

	_, e = DBTest().Exec("insert into persons(id,name,gender_id) values(1,'mahsa',1);")
	assert.NoError(t, e)

	//remove gender for test cascade delete
	_, e = DBTest().Exec("delete from gender where id=1")
	assert.Error(t, e, "pq: update or delete on table \"gender\" violates foreign key constraint \"persons_gender_id_fkey\" on table \"persons\"")

	result = 0
	e = DBTest().Get(&result, "select count(1) FROM persons")
	assert.NoError(t, e)
	//the person should not be deleted because the gender_id  key is not cascade
	assert.Equal(t, 1, result)
	assert.NoError(t, rmTempTables())

}

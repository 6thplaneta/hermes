package hermes

import (
	// "fmt"
	"github.com/jmoiron/sqlx"
	"os"
	"reflect"
	"testing"
	"time"
)

var instance *DataSrc

var sclassColl, sexColl, genderColl, classColl, studentColl, supervisorColl, personColl *Collection
var e error

// type DataSrc struct {
// 	DB     *sqlx.DB
// 	Cache  *CacheClient
// 	Search *SearchClient
// }

func DBTest() *DataSrc {

	if instance == nil {
		instance = &DataSrc{}
		db, _ := sqlx.Connect("postgres", "user=postgres password=123456 dbname=test_hermes sslmode=disable")
		search := SearchClient{}
		instance.Search = &search
		instance.Search.Engine = "sql"
		instance.DB = db
	}

	return instance
}

type Person struct {
	Id            int       `json:"id" hermes:"dbspace:persons"`
	Name          string    `json:"name"`
	Family        string    `json:"family"`
	Email         string    `json:"email"`
	Register_Date time.Time `hermes:"type:time" json:"register_date"`
	Age           int       `json:"age"`
	Child_Count   int       `json:"child_count"`
	Male          bool      `json:"male"`

	Middle_Name string  `json:"middle_name" db:"-"`
	Gender_Id   int     `json:"gender_id"`
	Gender      Gender  `json:"gender" db:"-" hermes:"one2one"`
	Student_Id  int     `json:"student_id"`
	Student     Student `json:"student" db:"-" hermes:"one2one"`
}

type Gender struct {
	Id    int    `json:"id" hermes:"dbspace:gender"`
	Title string `json:"title"`
}

type Sex struct {
	Id    int    `json:"id" hermes:"dbspace:sex"`
	Title string `json:"title"`
}

type Class struct {
	Id       int       `json:"id" hermes:"dbspace:classes"`
	Title    string    `json:"title"`
	Students []Student `db:"-" json:"students" hermes:"many2many:Student_Class"`
}

type Supervisor struct {
	Id        int       `json:"id" hermes:"dbspace:supervisors"`
	Name      string    `json:"name"`
	Gender_Id int       `json:"gender_id"`
	Gender    Gender    `db:"-" json:"gender" hermes:"one2one"`
	Students  []Student `db:"-" json:"students" hermes:"one2many"`
}

type Student struct {
	Id            int        `json:"id" hermes:"dbspace:students"`
	Title         string     `json:"title" hermes:"editable,searchable"`
	Sex_Id        int        `json:"sex_id"`
	Sex           Sex        `json:"sex" db:"-" hermes:"one2one"`
	Gender_Id     int        `json:"gender_id"`
	Gender        Gender     `json:"gender" db:"-" hermes:"many2one"`
	Supervisor_Id int        `json:"supervisor_id"`
	Supervisor    Supervisor `json:"supervisor" db:"-" hermes:"one2one"`
	Age           int        `json:"age"`
	Login_Date    time.Time  `json:"login_date" hermes:"type:time"`
	Classes       []Class    `json:"classes" db:"-" hermes:"many2many:Student_Class"`
}
type Student_Class struct {
	Id         int `json:"id" hermes:"dbspace:student_class"`
	Class_Id   int `json:"class_id"`
	Student_Id int `json:"student_id"`
}

func addTempTables() error {
	err := AddTable(DBTest().DB, AgentToken{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Agent{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Device{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Class{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Student_Class{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Sex{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Gender{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Student{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Supervisor{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest().DB, Person{})
	if err != nil {
		return err
	}

	return nil
}

func rmTempTables() error {

	//remove table
	_, e = DBTest().DB.Exec("drop table persons;drop table students;drop table supervisors;drop table gender;drop table sex;drop table classes;drop table student_class;drop table agents;drop table agent_tokens;drop table devices;")
	return e

}

func addTempCollections() error {

	AgentColl, e = NewAgentCollection(DBTest())
	if e != nil {
		return e
	}
	typ := reflect.TypeOf(Agent{})
	CollectionsMap[typ] = AgentColl

	DeviceColl, e = NewDBCollection(&Device{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Device{})
	CollectionsMap[typ] = DeviceColl

	sclassColl, e = NewCollection(&Student_Class{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Student_Class{})
	CollectionsMap[typ] = sclassColl

	classColl, e = NewCollection(&Class{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Class{})
	CollectionsMap[typ] = classColl

	sexColl, e = NewCollection(&Sex{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Sex{})
	CollectionsMap[typ] = sexColl

	genderColl, e = NewCollection(&Gender{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Gender{})
	CollectionsMap[typ] = genderColl

	studentColl, e = NewCollection(Student{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Student{})
	CollectionsMap[typ] = studentColl

	supervisorColl, e = NewCollection(Supervisor{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Supervisor{})
	CollectionsMap[typ] = supervisorColl

	personColl, e = NewCollection(Person{}, DBTest())
	if e != nil {
		return e
	}
	typ = reflect.TypeOf(Person{})
	CollectionsMap[typ] = personColl
	return nil
}

func DBTestDeallocate() {
	for {

		time.Sleep(time.Second)
		DBTest().DB.Exec("deallocate all;")
	}

}
func TestMain(m *testing.M) {
	// person := Person{}
	// FillJsonMap(person)

	//start
	application = NewApp("conf.yml")
	application.InitLogs("")
	application.Mount(AuthorizationModule, "/auth")

	InitMessages()
	StructsMap["Student_Class"] = Student_Class{}
	// dbInstance = DBTest().DB
	DBTest().Cache = &CacheClient{}
	addTempCollections()

	retCode := m.Run()

	//end
	os.Exit(retCode)

}

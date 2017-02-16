package hermes

import (
	"github.com/jmoiron/sqlx"
	"os"
	"reflect"
	"testing"
	"time"
)

var instance *sqlx.DB
var sclassColl, sexColl, genderColl, classColl, studentColl, supervisorColl *Collection
var e error

func DBTest() *sqlx.DB {
	if instance == nil {
		db, _ := sqlx.Connect("postgres", "user=postgres password=123456 dbname=test_hermes sslmode=disable")

		instance = db
	}
	return instance
}

type Person struct {
	Id          int `hermes:"dbspace:persons"`
	Name        string
	Middle_Name string `db:"-"`
	Gender_Id   int
	Gender      Gender `db:"-" hermes:"one2one:Person"`
	Student_Id  int
	Student     Student `db:"-" hermes:"one2one:Person"`
}

type Gender struct {
	Id    int `hermes:"dbspace:gender"`
	Title string
}

type Sex struct {
	Id    int `hermes:"dbspace:sex"`
	Title string
}

type Class struct {
	Id       int `hermes:"dbspace:classes"`
	Title    string
	Students []Student `db:"-" hermes:"many2many:Student_Class"`
}

type Supervisor struct {
	Id        int `hermes:"dbspace:supervisors"`
	Name      string
	Gender_Id int
	Gender    Gender    `db:"-" hermes:"one2one:Supervisor"`
	Students  []Student `db:"-" hermes:"one2many:Supervisor"`
}

type Student struct {
	Id            int    `hermes:"dbspace:students"`
	Title         string `hermes:"editable,searchable"`
	Sex_Id        int
	Sex           Sex `db:"-" hermes:"one2one:Student"`
	Gender_Id     int
	Gender        Gender `db:"-" hermes:"many2one"`
	Supervisor_Id int
	Supervisor    Supervisor `db:"-" hermes:"one2one:Student"`
	Age           int
	Login_Date    time.Time `hermes:"type:time"`
	Classes       []Class   `db:"-" hermes:"many2many:Student_Class"`
}
type Student_Class struct {
	Id         int `hermes:"dbspace:student_class"`
	Class_Id   int
	Student_Id int
}

func addTempTables() error {
	err := AddTable(DBTest(), AgentToken{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Agent{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Device{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Class{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Student_Class{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Sex{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Gender{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Student{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Supervisor{})
	if err != nil {
		return err
	}
	err = AddTable(DBTest(), Person{})
	if err != nil {
		return err
	}

	return nil
}

func rmTempTables() error {
	//remove table
	_, e = DBTest().Exec("drop table persons;drop table students;drop table supervisors;drop table gender;drop table sex;drop table classes;drop table student_class;drop table agents;drop table agent_tokens;drop table devices;")
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
	return nil
}

func TestMain(m *testing.M) {
	//start
	application = NewApp("conf.yml")
	application.InitLogs("")
	application.Mount(AuthorizationModule, "/auth")

	InitMessages()
	StructsMap["Student_Class"] = Student_Class{}
	dbInstance = DBTest()
	application.Cache = &CacheClient{}
	addTempCollections()

	retCode := m.Run()

	//end
	os.Exit(retCode)

}

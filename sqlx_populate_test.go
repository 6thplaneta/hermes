package hermes

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var testvals = []struct {
	str                   string
	openIndex, closeIndex int
}{
	{"()", 0, 1},
	{"(())", 0, 3},
	{"((()))", 0, 5},
	{"(())", 1, 2},
	{"((()))", 1, 4},
	{"((()))", 2, 3},
	{"(),(),()", 0, 1},
	{"(),(),()", 3, 4},
	{"(),(),()", 6, 7},
	{"(m),((y)),(y)", 4, 8},
	{"(m),((y)),(y)", 5, 7},
}

func TestCloseBracketIndex(t *testing.T) {
	for i := 0; i < len(testvals); i++ {
		clos, _ := closeBracketIndex(testvals[i].str, testvals[i].openIndex)
		assert.Equal(t, testvals[i].closeIndex, clos)
	}
}

func TestStrPopulateToArr(t *testing.T) {
	assert.NoError(t, addTempTables())

	val := strPopulateToArr("student(sex,gender,supervisor(gender,country(cities))),gender,country")
	assert.Equal(t, "sex,gender,supervisor(gender,country(cities))", val["student"])
	assert.Equal(t, "", val["gender"])
	assert.Equal(t, "", val["country"])

	val = strPopulateToArr("sex,gender,supervisor(gender,country(cities))")
	assert.Equal(t, "", val["sex"])
	assert.Equal(t, "", val["gender"])
	assert.Equal(t, "gender,country(cities)", val["supervisor"])

	val = strPopulateToArr("gender,country(cities)")
	assert.Equal(t, "", val["gender"])
	assert.Equal(t, "cities", val["country"])

	val = strPopulateToArr(",gender , country( cities , regions), , ,")
	assert.Equal(t, "", val["gender"])
	assert.Equal(t, "cities,regions", val["country"])
	assert.NoError(t, rmTempTables())

}

func TestPopulateOne2One(t *testing.T) {
	assert.NoError(t, addTempTables())

	//test one 2 one one level relationship start

	_, e = DBTest().Exec("insert into sex(id,title)values(1,'woman');")
	assert.NoError(t, e)

	gender := Gender{Title: "female"}
	genderColl.Create(SystemToken, nil, &gender)
	assert.NoError(t, e)

	superv := Supervisor{Name: "dr jessica", Gender_Id: 1}
	supervisorColl.Create(SystemToken, nil, &superv)
	assert.NoError(t, e)

	stu := Student{Title: "mahsa", Age: 31, Gender_Id: 1, Sex_Id: 1, Supervisor_Id: 1, Login_Date: time.Now()}
	_, e := studentColl.Create(SystemToken, nil, &stu)
	assert.NoError(t, e)

	//populate struct
	person := Person{}
	person.Gender_Id = 1
	person.Student_Id = 1

	e = PopulateStruct(SystemToken, DBTest(), &person, "student(sex,gender,supervisor(gender)),gender,")
	assert.NoError(t, e)
	assert.Equal(t, "female", person.Gender.Title)
	assert.Equal(t, 1, person.Gender.Id)
	assert.Equal(t, 1, person.Student.Id)
	assert.Equal(t, "mahsa", person.Student.Title)
	assert.Equal(t, "woman", person.Student.Sex.Title)
	assert.Equal(t, "female", person.Student.Gender.Title)
	assert.Equal(t, "dr jessica", person.Student.Supervisor.Name)
	assert.Equal(t, "female", person.Student.Supervisor.Gender.Title)

	var arr [1]Person
	arr[0] = person
	e = PopulateCollection(SystemToken, DBTest(), &arr, &Person{}, "student(sex,gender,supervisor(gender)),gender,")
	assert.NoError(t, e)
	assert.Equal(t, "female", arr[0].Gender.Title)
	assert.Equal(t, 1, arr[0].Gender.Id)
	assert.Equal(t, 1, arr[0].Student.Id)
	assert.Equal(t, "mahsa", arr[0].Student.Title)
	assert.Equal(t, "woman", arr[0].Student.Sex.Title)
	assert.Equal(t, "female", arr[0].Student.Gender.Title)
	assert.Equal(t, "dr jessica", arr[0].Student.Supervisor.Name)
	assert.Equal(t, "female", arr[0].Student.Supervisor.Gender.Title)

	//test one 2 one one level relationship end
	assert.NoError(t, rmTempTables())

}

func TestPopulateOne2Many(t *testing.T) {
	assert.NoError(t, addTempTables())

	//test one 2 many start

	_, e = DBTest().Exec("insert into sex(id,title)values(1,'woman');")
	assert.NoError(t, e)

	superv := Supervisor{Name: "dr jessica", Gender_Id: 1}
	supervisorColl.Create(SystemToken, nil, &superv)
	assert.NoError(t, e)

	stu := Student{Title: "mahsa ghoreishi", Age: 27, Gender_Id: 1, Sex_Id: 1, Supervisor_Id: 1, Login_Date: time.Now()}
	_, e := studentColl.Create(SystemToken, nil, &stu)
	assert.NoError(t, e)

	stu = Student{Title: "sara test", Age: 22, Gender_Id: 1, Sex_Id: 1, Supervisor_Id: 1, Login_Date: time.Now()}
	_, e = studentColl.Create(SystemToken, nil, &stu)
	assert.NoError(t, e)

	//populate struct
	supervisor := Supervisor{}
	supervisor.Id = 1

	e = PopulateStruct(SystemToken, DBTest(), &supervisor, "students(sex)")
	assert.NoError(t, e)

	assert.Equal(t, 2, len(supervisor.Students))
	assert.Equal(t, "woman", supervisor.Students[0].Sex.Title)
	assert.Equal(t, "woman", supervisor.Students[1].Sex.Title)

	var arr [1]Supervisor
	arr[0] = supervisor
	e = PopulateCollection(SystemToken, DBTest(), &arr, &supervisor, "students(sex)")
	assert.NoError(t, e)
	assert.Equal(t, 2, len(arr[0].Students))
	assert.Equal(t, "woman", supervisor.Students[0].Sex.Title)
	assert.Equal(t, "woman", supervisor.Students[1].Sex.Title)

	//test one 2 many end
	assert.NoError(t, rmTempTables())

}

func TestPopulateMany2Many(t *testing.T) {
	assert.NoError(t, addTempTables())

	//test many 2 many start

	_, e = DBTest().Exec("insert into sex(id,title)values(1,'woman');")
	assert.NoError(t, e)

	_, e = DBTest().Exec("insert into classes(id,title)values(1,'artificial intelligence');")
	assert.NoError(t, e)

	stu := Student{Title: "mahsa ghoreishi", Age: 26, Gender_Id: 1, Sex_Id: 1, Supervisor_Id: 1, Login_Date: time.Now()}
	_, e := studentColl.Create(SystemToken, nil, &stu)
	assert.NoError(t, e)

	stu = Student{Title: "sara test", Age: 32, Gender_Id: 1, Sex_Id: 1, Supervisor_Id: 1, Login_Date: time.Now()}
	_, e = studentColl.Create(SystemToken, nil, &stu)
	assert.NoError(t, e)

	_, e = DBTest().Exec("insert into student_class(id,class_id,student_id)values(1,1,1);insert into student_class(id,class_id,student_id)values(2,1,2);")
	assert.NoError(t, e)

	//populate struct
	class := Class{}
	class.Id = 1

	e = PopulateStruct(SystemToken, DBTest(), &class, "students(sex)")
	assert.NoError(t, e)

	assert.Equal(t, 2, len(class.Students))
	assert.Equal(t, "woman", class.Students[0].Sex.Title)
	assert.Equal(t, "woman", class.Students[1].Sex.Title)

	//populate collection
	var arr [1]Class
	arr[0] = class
	e = PopulateCollection(SystemToken, DBTest(), &arr, &class, "students(sex)")
	assert.NoError(t, e)
	assert.Equal(t, 2, len(arr[0].Students))
	assert.Equal(t, "woman", class.Students[0].Sex.Title)
	assert.Equal(t, "woman", class.Students[1].Sex.Title)

	//test one 2 many end

	assert.NoError(t, rmTempTables())
}

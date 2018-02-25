package hermes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Collection_Start(t *testing.T) {
	assert.NoError(t, addTempTables())

}
func Test_Create(t *testing.T) {

	ti, _ := time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T11:58:15+04:30")
	stu := Student{Title: "mahsa gh", Age: 27, Gender_Id: 1, Sex_Id: 1, Supervisor_Id: 1, Login_Date: ti}
	result, e := studentColl.Create(SystemToken, nil, &stu)
	assert.NoError(t, e)
	r := result.(*Student)
	assert.Equal(t, 1, r.Id)
	assert.Equal(t, "mahsa gh", r.Title)
	assert.Equal(t, 1, r.Sex_Id)
	assert.Equal(t, 1, r.Supervisor_Id)

	ti, _ = time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T13:58:15+04:30")

	stu = Student{Title: "ali", Age: 40, Gender_Id: 2, Sex_Id: 1, Supervisor_Id: 1, Login_Date: ti}
	_, e = studentColl.Create(SystemToken, nil, &stu)
	assert.NoError(t, e)
}

func Test_Update(t *testing.T) {

	// stu := Student{Id: 1, Title: "mahsa ghoreishi", Age: 27, Gender_Id: 1, Sex_Id: 1, Supervisor_Id: 1}

	// e := studentColl.Update(SystemToken, 1, &stu)
	// assert.NoError(t, e)

	// result, e := studentColl.Get(SystemToken, 1, "")
	// assert.NoError(t, e)
	// r := result.(*Student)
	// assert.Equal(t, 1, r.Id)
	// assert.Equal(t, "mahsa ghoreishi", r.Title)
	// assert.Equal(t, 1, r.Sex_Id)
	// assert.Equal(t, 1, r.Supervisor_Id)

}

func Test_Get(t *testing.T) {
	result, e := studentColl.Get(SystemToken, 1, "")
	assert.NoError(t, e)

	rstu := result.(*Student)
	assert.Equal(t, 1, rstu.Id)
	assert.Equal(t, "mahsa ghoreishi", rstu.Title)
}

func Test_List(t *testing.T) {
	pg := Paging{Size: 10, Num: 1}
	params := NewParams(Student{})

	//get top 10 records
	result, e := studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr := *result.(*[]Student)
	stud := arr[0]
	assert.Equal(t, 2, len(arr))
	assert.Equal(t, 1, stud.Id)
	assert.Equal(t, "mahsa ghoreishi", stud.Title)
	assert.Equal(t, 1, stud.Sex_Id)

	//search not exist value
	params.AddFilter("$$search", Filter{Type: "search", Value: "sara"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)
	arr = *result.(*[]Student)
	assert.Equal(t, 0, len(arr))

	//search like test
	params.AddFilter("$$search", Filter{Type: "search", Value: "mah"})

	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)
	arr = *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	//filter exact not exist
	params = NewParams(Student{})
	params.AddFilter("age", Filter{Type: "exact", Value: 22, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 0, len(arr))

	//filter exact exist
	params = NewParams(Student{})

	params.AddFilter("age", Filter{Type: "exact", Value: 27, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	//filter range test (from)
	params = NewParams(Student{})

	params.AddFilter("age", Filter{Type: "range", Value: RangeFilter{From: 42}, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 0, len(arr))

	//filter range test (from)
	params = NewParams(Student{})

	params.AddFilter("age", Filter{Type: "range", Value: RangeFilter{From: 20}, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 2, len(arr))

	//filter range test (to)
	params = NewParams(Student{})

	params.AddFilter("age", Filter{Type: "range", Value: RangeFilter{To: 40}, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	//filter range test (to)
	params = NewParams(Student{})

	params.AddFilter("age", Filter{Type: "range", Value: RangeFilter{To: 41}, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 2, len(arr))

	//filter range test (from-to) not exist
	params = NewParams(Student{})

	params.AddFilter("age", Filter{Type: "range", Value: RangeFilter{From: 28, To: 40}, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 0, len(arr))

	//filter range test (from-to) exist
	params = NewParams(Student{})

	params.AddFilter("age", Filter{Type: "range", Value: RangeFilter{From: 20, To: 45}, FieldType: "int"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 2, len(arr))

	//filter array test (in)

	params = NewParams(Student{})
	params.AddFilter("age", Filter{Type: "array", Value: []int{27}, FieldType: "int"})
	result, e = studentColl.List("system", params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	params = NewParams(Student{})
	params.AddFilter("age", Filter{Type: "array", Value: []int{35}, FieldType: "int"})
	result, e = studentColl.List("system", params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 0, len(arr))

}
func Test_List_Date_Filter(t *testing.T) {
	pg := Paging{Size: 10, Num: 1}

	//exists
	params := NewParams(Student{})

	ti, _ := time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T11:58:15+04:30")
	params.AddFilter("login_date", Filter{Type: "exact", Value: ti, FieldType: "time"})
	result, e := studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr := *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	// not existing time with different timezone
	params = NewParams(Student{})

	ti, _ = time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T11:58:15+05:30")
	params.AddFilter("login_date", Filter{Type: "exact", Value: ti, FieldType: "time"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 0, len(arr))

	// existing time with different timezone
	params = NewParams(Student{})

	ti, _ = time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T05:58:15-01:30")
	params.AddFilter("login_date", Filter{Type: "exact", Value: ti, FieldType: "time"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	//date range filter(from)
	params = NewParams(Student{})

	ti, _ = time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T11:58:15+04:30")
	params.AddFilter("login_date", Filter{Type: "exact", Value: ti, FieldType: "time"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	//date range filter(from)
	ti, _ = time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T11:59:15+04:30")
	params = NewParams(Student{})
	params.AddFilter("login_date", Filter{Type: "range", Value: RangeFilter{From: ti}, FieldType: "time"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 1, len(arr))

	//date range filter(to)
	ti, _ = time.Parse("2006-01-02T15:04:05-07:00", "2016-04-11T11:58:15+04:30")
	params = NewParams(Student{})

	params.AddFilter("login_date", Filter{Type: "range", Value: RangeFilter{To: ti}, FieldType: "time"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr = *result.(*[]Student)
	assert.Equal(t, 0, len(arr))
}
func Test_Indirect_Filter(t *testing.T) {
	pg := Paging{Size: 10, Num: 1}

	gender := Gender{Title: "female"}
	genderColl.Create(SystemToken, nil, &gender)
	assert.NoError(t, e)

	gender = Gender{Title: "male"}
	genderColl.Create(SystemToken, nil, &gender)
	assert.NoError(t, e)

	superv := Supervisor{Name: "dr jessica", Gender_Id: 1}
	supervisorColl.Create(SystemToken, nil, &superv)
	assert.NoError(t, e)

	//one 2 many filter by supervisor
	params := NewParams(Student{})

	params.AddFilter("Supervisor.Gender.Title", Filter{Type: "exact", Value: "female", FieldType: "string"})
	result, e := studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arr := *result.(*[]Student)
	assert.Equal(t, 2, len(arr))

	//one 2 many filter by
	params = NewParams(Student{})

	params.AddFilter("Students.Title", Filter{Type: "exact", Value: "mahsa ghoreishi", FieldType: "string"})
	result, e = supervisorColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)
	arrs := *result.(*[]Supervisor)
	assert.Equal(t, 1, len(arrs))

	//one 2 many filter
	params = NewParams(Student{})

	params.AddFilter("Students.Gender.Title", Filter{Type: "exact", Value: "male", FieldType: "string"})
	result, e = supervisorColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)
	arrs = *result.(*[]Supervisor)
	assert.Equal(t, 1, len(arrs))

	//many 2 many filter
	class := Class{Title: "Computer Networks"}
	classColl.Create(SystemToken, nil, &class)
	assert.NoError(t, e)

	class = Class{Title: "Programming"}
	classColl.Create(SystemToken, nil, &class)
	assert.NoError(t, e)

	_, e = DBTest().DB.Exec("insert into student_class(class_id,student_id)values(1,1);insert into student_class(class_id,student_id)values(2,2);")
	assert.NoError(t, e)

	params = NewParams(Student{})

	params.AddFilter("Classes.Title", Filter{Type: "exact", Value: "Computer Networks", FieldType: "string"})
	result, e = studentColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arrst := *result.(*[]Student)
	assert.Equal(t, 1, len(arrst))

	assert.Equal(t, 1, arrst[0].Id)
	assert.Equal(t, "mahsa ghoreishi", arrst[0].Title)

	//
	params = NewParams(Student{})

	params.AddFilter("Students.Classes.Title", Filter{Type: "exact", Value: "Programming", FieldType: "string"})

	result, e = supervisorColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arrs = *result.(*[]Supervisor)
	assert.Equal(t, 1, len(arrs))

	assert.Equal(t, 1, arrs[0].Id)
	assert.Equal(t, "dr jessica", arrs[0].Name)

	//not exists
	params = NewParams(Student{})

	params.AddFilter("Students.Classes.Title", Filter{Type: "exact", Value: "notExists", FieldType: "string"})
	result, e = supervisorColl.List(SystemToken, params, &pg, "", "")
	assert.NoError(t, e)

	arrs = *result.(*[]Supervisor)
	assert.Equal(t, 0, len(arrs))

}

func Test_Sort(t *testing.T) {
	//sort desc

	params := NewParams(Student{})

	result, e := studentColl.List(SystemToken, params, &Paging{Size: 10, Num: 1, Sort: "Title", Order: "asc"}, "", "")
	assert.NoError(t, e)
	arr := *result.(*[]Student)

	assert.Equal(t, "ali", arr[0].Title)
	assert.Equal(t, "mahsa ghoreishi", arr[1].Title)

	//sort asc
	result, e = studentColl.List(SystemToken, params, &Paging{Size: 10, Num: 1, Sort: "Title", Order: "desc"}, "", "")
	assert.NoError(t, e)
	arr = *result.(*[]Student)

	assert.Equal(t, "mahsa ghoreishi", arr[0].Title)
	assert.Equal(t, "ali", arr[1].Title)

}

func Test_Rel(t *testing.T) {
	//many 2 one test
	//change student gender from 1 to 2(female to male)
	result, e := studentColl.Get(SystemToken, 1, "")
	assert.NoError(t, e)
	stu := result.(*Student)
	assert.Equal(t, 1, stu.Gender_Id)

	e = studentColl.Rel(SystemToken, 1, "gender", []int{2})
	assert.NoError(t, e)
	result, e = studentColl.Get(SystemToken, 1, "")
	assert.NoError(t, e)
	stu = result.(*Student)
	assert.Equal(t, 2, stu.Gender_Id)

	//one 2 many change student supervisor_id from 1 to 2
	e = supervisorColl.Rel(SystemToken, 2, "students", []int{1})
	assert.NoError(t, e)
	result, e = studentColl.Get(SystemToken, 1, "")
	assert.NoError(t, e)
	stu = result.(*Student)
	assert.Equal(t, 2, stu.Supervisor_Id)

	//many 2 many
	e = studentColl.Rel(SystemToken, 1, "classes", []int{3})
	assert.NoError(t, e)
	var arr []Student_Class
	e = DBTest().DB.Select(&arr, "select * from student_class where student_id=1 and class_id=3")
	assert.NoError(t, e)
	assert.Equal(t, 1, len(arr))

}

func Test_UnRel(t *testing.T) {
	//many 2 one test
	//delete student gender
	result, e := studentColl.Get(SystemToken, 1, "")
	assert.NoError(t, e)
	stu := result.(*Student)
	assert.Equal(t, 2, stu.Gender_Id)

	e = studentColl.UnRel(SystemToken, 1, "gender", []int{0})
	assert.NoError(t, e)
	result, e = studentColl.Get(SystemToken, 1, "")
	assert.NoError(t, e)
	stu = result.(*Student)
	assert.Equal(t, 0, stu.Gender_Id)

	//one 2 many
	e = supervisorColl.UnRel(SystemToken, 0, "students", []int{1})
	assert.NoError(t, e)
	result, e = studentColl.Get(SystemToken, 1, "")
	assert.NoError(t, e)
	stu = result.(*Student)
	assert.Equal(t, 0, stu.Supervisor_Id)

	//many 2 many
	e = studentColl.UnRel(SystemToken, 1, "classes", []int{1})
	assert.NoError(t, e)
	var arr []Student_Class
	e = DBTest().DB.Select(&arr, "select * from student_class where student_id=1 and class_id=1")
	assert.NoError(t, e)
	assert.Equal(t, 0, len(arr))
}

func Test_Delete(t *testing.T) {
	// e = studentColl.Delete(SystemToken, 1)
	// assert.NoError(t, e)

	// _, e := studentColl.Get(SystemToken, 1, "")
	// assert.Error(t, e, "sql: no rows in result set")
}

func Test_Collection_End(t *testing.T) {
	assert.NoError(t, rmTempTables())
	DBTest().DB.Exec("deallocate all;")

}

package hermes

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPreUnPopulate(t *testing.T) {
	assert.NoError(t, addTempTables())

	//test one 2 one start

	//unpopulate struct
	supervisor := Supervisor{}
	supervisor.Gender.Title = "female"

	trans, _ := DBTest().Begin()
	e = PreUnPopulate(SystemToken, trans, &supervisor)
	if e != nil {
		trans.Rollback()
	}
	trans.Commit()
	assert.NoError(t, e)
	var result []Gender
	DBTest().Select(&result, "select * from gender ")

	assert.Equal(t, 1, len(result))
	assert.Equal(t, 1, result[0].Id)
	assert.Equal(t, "female", result[0].Title)

	//test one 2 one end
	assert.NoError(t, rmTempTables())

}

func TestUnPopulateOne2Many(t *testing.T) {
	assert.NoError(t, addTempTables())

	//test one 2 many start

	//unpopulate struct
	supervisor := Supervisor{}
	supervisor.Id = 1
	var students []Student
	students = append(students, Student{Id: 1, Title: "test1", Sex_Id: 1, Gender_Id: 1, Supervisor_Id: 1})
	students = append(students, Student{Id: 2, Title: "test2", Sex_Id: 1, Gender_Id: 1, Supervisor_Id: 1})

	supervisor.Students = students

	trans, _ := DBTest().Begin()
	e = UnPopulate(SystemToken, trans, &supervisor)
	if e != nil {
		trans.Rollback()
	}
	trans.Commit()
	assert.NoError(t, e)
	var result []Student
	DBTest().Select(&result, "select * from students ")

	assert.Equal(t, 2, len(result))
	assert.Equal(t, 1, result[0].Id)
	assert.Equal(t, "test1", result[0].Title)
	assert.Equal(t, 1, result[0].Sex_Id)
	assert.Equal(t, 1, result[0].Gender_Id)

	assert.Equal(t, 2, result[1].Id)
	assert.Equal(t, "test2", result[1].Title)
	assert.Equal(t, 1, result[1].Sex_Id)
	assert.Equal(t, 1, result[1].Gender_Id)

	//test one 2 many end
	assert.NoError(t, rmTempTables())

}

func TestUnPopulateMany2Many(t *testing.T) {
	assert.NoError(t, addTempTables())

	//test many 2 many start

	//unpopulate struct
	class := Class{}
	class.Id = 1
	var students []Student
	students = append(students, Student{Title: "test1", Sex_Id: 1, Gender_Id: 1, Supervisor_Id: 1})
	students = append(students, Student{Title: "test2", Sex_Id: 1, Gender_Id: 1, Supervisor_Id: 1})

	class.Students = students

	trans, _ := DBTest().Begin()
	e = UnPopulate(SystemToken, trans, &class)
	if e != nil {
		trans.Rollback()
	}
	trans.Commit()
	assert.NoError(t, e)
	var result []Student
	DBTest().Select(&result, "select * from students ")
	assert.Equal(t, 2, len(result))
	assert.Equal(t, 2, result[1].Id)
	assert.Equal(t, "test2", result[1].Title)
	assert.Equal(t, 1, result[1].Sex_Id)
	assert.Equal(t, 1, result[1].Gender_Id)

	var scresult []Student_Class
	DBTest().Select(&scresult, "select * from student_class ")
	assert.Equal(t, 2, len(scresult))
	assert.Equal(t, 1, scresult[0].Id)
	assert.Equal(t, 1, scresult[0].Class_Id)
	assert.Equal(t, 1, scresult[0].Student_Id)
	assert.Equal(t, 2, scresult[1].Id)
	assert.Equal(t, 1, scresult[1].Class_Id)
	assert.Equal(t, 2, scresult[1].Student_Id)

	//test add duplicate data
	e = UnPopulate(SystemToken, trans, &class)
	var result1 []Student

	DBTest().Select(&result1, "select * from students ")
	assert.Equal(t, 2, len(result1))
	assert.Equal(t, "test1", result1[0].Title)
	assert.Equal(t, "test2", result1[1].Title)

	var scresult1 []Student_Class

	DBTest().Select(&scresult1, "select * from student_class ")
	assert.Equal(t, 2, len(scresult1))
	assert.Equal(t, 1, scresult1[0].Id)
	assert.Equal(t, 1, scresult1[0].Class_Id)
	assert.Equal(t, 1, scresult1[0].Student_Id)
	assert.Equal(t, 2, scresult1[1].Id)
	assert.Equal(t, 1, scresult1[1].Class_Id)
	assert.Equal(t, 2, scresult1[1].Student_Id)
	//test many 2 many end
	assert.NoError(t, rmTempTables())

}

package hermes

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestStrInArr(t *testing.T) {
	arrNames := []string{"mahsa", "sara", "hanieh", "elham"}
	assert.Contains(t, arrNames, "mahsa")
	assert.Contains(t, arrNames, "sara")
	assert.Contains(t, arrNames, "hanieh")
	assert.Contains(t, arrNames, "elham")
	assert.NotContains(t, arrNames, "elha")
	assert.NotContains(t, arrNames, "kdjkjk")

}

func TestIsNil(t *testing.T) {
	//integer test start
	var intv int
	intv = 1
	isnil, err := IsNil(intv)
	//this function only works with pointers
	assert.Error(t, err)
	assert.Equal(t, false, isnil)

	isnil, err = IsNil(&intv)
	assert.NoError(t, err)
	assert.Equal(t, false, isnil)

	var intp *int
	isnil, err = IsNil(intp)
	assert.NoError(t, err)
	assert.Equal(t, true, isnil)
	//integer test end

	//string test start
	var strv string
	strv = "m"
	isnil, err = IsNil(strv)
	//this function only works with pointers
	assert.Error(t, err)
	assert.Equal(t, false, isnil)

	isnil, err = IsNil(&strv)
	assert.NoError(t, err)
	assert.Equal(t, false, isnil)

	var strp *string
	isnil, err = IsNil(strp)
	assert.NoError(t, err)
	assert.Equal(t, true, isnil)
	//string test end

	//date test start
	var datev time.Time
	datev = time.Now()
	isnil, err = IsNil(datev)
	//this function only works with pointers
	assert.Error(t, err)
	assert.Equal(t, false, isnil)

	isnil, err = IsNil(&datev)
	assert.NoError(t, err)
	assert.Equal(t, false, isnil)

	var datep *time.Time
	isnil, err = IsNil(datep)
	assert.NoError(t, err)
	assert.Equal(t, true, isnil)
	//date test end
}

func TestCastToStr(t *testing.T) {
	assert.Equal(t, "true", CastToStr(true, "bool", ""))
	assert.Equal(t, "false", CastToStr(false, "bool", ""))
	assert.Equal(t, "1", CastToStr(1, "int", ""))
	assert.Equal(t, "-2.5", CastToStr(-2.5, "float64", ""))
	assert.Equal(t, "2.5", CastToStr(2.5, "float64", ""))
	assert.Equal(t, "2.5", CastToStr(+2.5, "float64", ""))
	assert.Equal(t, "'hello!'", CastToStr("hello!", "string", ""))
	assert.Equal(t, "'1988-05-19T02:23:21+00:00'", CastToStr(time.Date(1988, 05, 19, 02, 23, 21, 0, time.UTC), "time", ""))
}

func TestCastArrToStr(t *testing.T) {
	assert.Equal(t, "true,false", CastArrToStr([]bool{true, false}, "bool", ""))
	assert.Equal(t, "1,2,5", CastArrToStr([]int{1, 2, 5}, "int", ""))
	assert.Equal(t, "1,2.5,5", CastArrToStr([]float64{1, 2.5, 5}, "float64", ""))
	assert.Equal(t, "1,2.5,5", CastArrToStr([]float64{1, 2.5, 5}, "float64", ""))
	assert.Equal(t, "'t','n','h'", CastArrToStr([]string{"t", "n", "h"}, "string", ""))
	assert.Equal(t, "'1988-05-19T02:23:21+00:00','2016-05-19T02:23:21+00:00'", CastArrToStr([]time.Time{time.Date(1988, 05, 19, 02, 23, 21, 0, time.UTC), time.Date(2016, 05, 19, 02, 23, 21, 0, time.UTC)}, "time", ""))
}

func TestCastStrToVal(t *testing.T) {
	assert.Equal(t, "mahsa", CastStrToVal("mahsa", "string"))
	assert.Equal(t, false, CastStrToVal("false", "bool"))
	assert.Equal(t, true, CastStrToVal("true", "bool"))
	assert.Equal(t, 1, CastStrToVal("1", "int"))
	assert.Equal(t, 1.5, CastStrToVal("1.5", "float64"))
	assert.Equal(t, "1988-05-19+02:23:21++0000+UTC", CastStrToVal("1988-05-19 02:23:21 +0000 UTC", "time"))
}

func TestCastStrToArr(t *testing.T) {
	assert.Equal(t, []string{"mahsa", "sara", "elham"}, CastStrToArr("mahsa,sara,elham", "string"))
	assert.NotEqual(t, []string{"mahsa", "elham", "sara"}, CastStrToArr("mahsa,sara,elham", "string"))

	assert.Equal(t, []int{1, 2, 3}, CastStrToArr("1,2,3", "int"))
	assert.Equal(t, []bool{true, false}, CastStrToArr("true,false", "bool"))

	assert.Equal(t, []float64{1.5, 2.5, 3}, CastStrToArr("1.5,2.5,3", "float64"))

}

// func TestGetTitle(t *testing.T) {
// 	assert.Equal(t, "Mahsa", GetTitle("mahsa"))
// 	assert.Equal(t, "Mahsa_Ghoreishi", GetTitle("mahsa_ghoreishi"))
// 	assert.Equal(t, "Mahsa_Ghoreishi_27_Years_Old", GetTitle("mahsa_ghoreishi_27_years_old"))

// 	assert.Equal(t, "MAhSa", GetTitle("mAhSa"))
// 	assert.Equal(t, "OnSale", GetTitle("OnSale"))

// }

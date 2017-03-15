package hermes

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"github.com/antonholmquist/jason"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

/*
* This is a function that searches a string value in array
* @param 	string			string value to search in array
* @param 	[]string 		array
* @return	bool 			determines if value exists or not.
 */
func strInArr(st string, arr []string) bool {
	for _, s := range arr {
		if s == st {
			return true
		}
	}
	return false
}

func GenerateHash(input, secret string) string {
	h := sha1.New()
	h.Write([]byte(secret))
	h.Write([]byte(input))
	bs := h.Sum(nil)
	bs2 := fmt.Sprintf("%x", bs)
	return string(bs2)
}

/*
* This function determines if a interface value is nil or not
* @param 	interface		the value should be pointer
* @return	bool 			determines if value is nil or not
 */
func IsNil(value interface{}) (bool, error) {
	switch value.(type) {
	case *string:
		a := value.(*string)
		return a == nil, nil
	case string:
		return false, errors.New(Messages["ExpectedPointer"])
	case *int:
		a := value.(*int)
		return a == nil, nil
	case int:
		return false, errors.New(Messages["ExpectedPointer"])
	case *time.Time:
		a := value.(*time.Time)
		return a == nil, nil
	case time.Time:
		return false, errors.New(Messages["ExpectedPointer"])
	default:
		return true, nil
		// t is some other type that we didn't name.
	}
}

/*
* This is a function that casts an interface value to string
* @param 	interface{}		value
* @param 	string 			type of field
* @return	string 			string value
 */
func CastToStr(vl interface{}, typ, dbtype string) string {
	var castedVal string
	switch typ {
	case "time":
		val, _ := vl.(time.Time)

		if dbtype == "date" {
			castedVal = val.Format(Messages["ShortForm"])

		} else if dbtype == "time" {
			castedVal = val.Format(Messages["TimeForm"])

		} else {
			castedVal = val.Format(Messages["LongForm"])

		}
		break
	case "bool":
		castedVal = fmt.Sprintf("%t", vl)
		break
	case "int":
		val, _ := vl.(int)
		castedVal = strconv.Itoa(val)
		break
	case "float64":
		val, _ := vl.(float64)
		castedVal = strconv.FormatFloat(val, 'g', -1, 64)
		break
	case "float32":
		val, _ := vl.(float32)
		castedVal = strconv.FormatFloat(float64(val), 'g', -1, 32)
		break
	case "string":
		castedVal = vl.(string)
		break
	}

	return castedVal
}

/*
* This is a function that casts an array to string(splict with , )
* @param 	interface{}		array to cast
* @param 	string 			type of array
* @return	string 			casted value (example 'hi','salam')
 */
func CastArrToStr(vl interface{}, typ, dbtype string) string {
	var cs string
	rs := ""
	if typ == "int" {
		arr := vl.([]int)
		if len(arr) == 0 {
			return ""
		}
		for _, v := range arr {
			cs = CastToStr(v, "int", dbtype)

			rs += cs + ","
		}
	} else if typ == "float64" {
		arr := vl.([]float64)
		if len(arr) == 0 {
			return ""
		}
		for _, v := range arr {
			cs = CastToStr(v, "float64", dbtype)
			rs += cs + ","
		}
	} else if typ == "string" {
		arr := vl.([]string)
		if len(arr) == 0 {
			return ""
		}
		for _, v := range arr {
			cs = CastToStr(v, "string", dbtype)
			cs = "'" + cs + "'"
			rs += cs + ","
		}
	} else if typ == "bool" {
		arr := vl.([]bool)
		if len(arr) == 0 {
			return ""
		}
		for _, v := range arr {
			cs = CastToStr(v, "bool", dbtype)
			rs += cs + ","
		}
	} else if typ == "time" {
		arr := vl.([]time.Time)
		if len(arr) == 0 {
			return ""
		}
		for _, v := range arr {
			cs = CastToStr(v, "time", dbtype)
			cs = "'" + cs + "'"
			rs += cs + ","
		}
	}

	rs = rs[:len(rs)-1]
	return rs

}

/*
* This is a function that casts a string value to specified type
* @param 	string			string for cast to interface{}
* @param 	string 			type of value
* @return	interface{} 	casted value
 */
func CastStrToVal(strValue, typ string) interface{} {
	var castedVal interface{}

	castedVal = strValue
	switch typ {
	case "time":
		// t, _ := time.Parse(time.RFC3339, strings.Replace(strValue, " ", "+", -1))
		// castedVal = t
		castedVal = strings.Replace(strValue, " ", "+", -1)
		break
	case "bool":
		t, _ := strconv.ParseBool(strValue)
		castedVal = t
		break
	case "int":
		inted, _ := strconv.Atoi(strValue)
		castedVal = inted
		break
	case "float64":
		f, _ := strconv.ParseFloat(strValue, 64)
		castedVal = f
		break
	case "float32":
		f, _ := strconv.ParseFloat(strValue, 32)
		castedVal = float32(f)
		break
	case "string":
		castedVal = strValue
		break
	}
	return castedVal
}

/*
* This is a function that casts a string value to array
* @param 	string			string for cast to array (split with ,)
* @param 	string 			type of array
* @return	interface{} 	array of values
 */
func CastStrToArr(strValue, typeOfField string) interface{} {
	vals := strings.Split(strValue, ",")

	switch typeOfField {
	case "bool":
		var castedVal []bool
		for _, v := range vals {
			castedVal = append(castedVal, CastStrToVal(v, typeOfField).(bool))
		}
		return castedVal
	case "int":
		var castedVal []int
		for _, v := range vals {
			castedVal = append(castedVal, CastStrToVal(v, typeOfField).(int))
		}
		return castedVal
	case "float64":
		var castedVal []float64
		for _, v := range vals {
			castedVal = append(castedVal, CastStrToVal(v, typeOfField).(float64))
		}
		return castedVal
	case "string":
		var castedVal []string
		for _, v := range vals {
			castedVal = append(castedVal, CastStrToVal(v, typeOfField).(string))
		}
		return castedVal
	}

	return nil
}

func GetTitle(val string) string {
	strTitle := ""
	arr := strings.Split(val, "_")
	for i := 0; i < len(arr); i++ {
		strTitle = strTitle + "_" + strings.Title(arr[i])
	}
	return strings.Replace(strTitle, "_", "", 1)
}
func ReadHttpBody(response *http.Response) string {

	bodyBuffer := make([]byte, 5000)
	var str string

	count, err := response.Body.Read(bodyBuffer)

	for ; count > 0; count, err = response.Body.Read(bodyBuffer) {

		if err != nil {

		}

		str += string(bodyBuffer[:count])
	}

	return str

}

func GetFbUser(accessToken string) (*jason.Object, error) {
	response, err := http.Get("https://graph.facebook.com/me?access_token=" + accessToken + "&fields=id,email,name,gender,first_name,last_name")

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(Messages["NotFound"])
	}
	str := ReadHttpBody(response)
	user, _ := jason.NewObjectFromBytes([]byte(str))
	return user, nil
}

func EscapeChars(str string) string {
	return strings.Replace(str, "'", "''", -1)

}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func DeallocateStatements() {
	for {

		time.Sleep(time.Second)
		application.DataSrc.DB.Exec("deallocate all;")
	}
}

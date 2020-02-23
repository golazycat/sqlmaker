package sqlmaker

import (
	"reflect"
	"strconv"
	"time"
)

const datetimeFormat = "2006-01-02 15:04:05"

// 要想使用sqlmaker生成某个结构体的SQL语句，则该结构体必须实现该接口
// 另外，每个字段需要使用标签"field"来指定其在数据表中的字段名称
type Entity interface {
	// 返回结构体在数据库中对应的表名
	TableName() string

	// 返回表中id字段名和对应的结构体中的值
	// 如果想使用SQLMaker的ById()功能，则该函数必须返回可用的值
	// 如果不适用ById()功能，则该函数返回值将不会被用到
	GetId() (string, interface{})
}

// 字段结构体
// Name: 字段在结构体中的名称
// TableFieldName: 字段在数据表中的名称，需要通过字段标签"field"指定，如果不指定，则和Name一致
// val: 字段的具体值
type Field struct {
	Name           string
	TableFieldName string
	val            string
}

// 将一个Entity的所有字段解析出来，返回一个field列表
func decodeEntity(o interface{}, selects []string) []Field {

	fields := make([]Field, 0)

	vs := reflect.ValueOf(o)

	for i := 0; i < vs.NumField(); i++ {

		field := vs.Type().Field(i)
		tag := field.Tag.Get("field")

		if !contains(field.Name, tag, selects) {
			continue
		}

		originVal := vs.Field(i).Interface()
		if tag == "" || tag == "-" {
			tag = field.Name
		}

		val := ""
		switch field.Type.Name() {
		case "string":
			val = stringValue(originVal.(string))
		case "int":
			val = intToString(originVal.(int))
		case "Time":
			val = dateToString(originVal.(time.Time))
		}

		if val != "" {
			fieldObj := Field{
				Name:           field.Name,
				TableFieldName: tag,
				val:            val,
			}
			fields = append(fields, fieldObj)
		}
	}

	return fields
}

func intToString(v int) string {
	return strconv.Itoa(v)
}

func dateToString(v time.Time) string {
	return stringValue(v.Format(datetimeFormat))
}

func stringValue(s string) string {
	return "'" + s + "'"
}

func stringName(s string) string {
	return "`" + s + "`"
}

func contains(s1 string, s2 string, ss []string) bool {

	if ss == nil {
		return true
	}

	for _, ts := range ss {
		if s1 == ts || s2 == ts {
			return true
		}
	}
	return false
}

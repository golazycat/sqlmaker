package sqlmaker

import (
	"reflect"
	"strconv"
	"time"
)

const datetimeFormat = "2006-01-02 15:04:05"

type Entity interface {
	TableName() string
	GetId() (string, interface{})
}

type Field struct {
	Name           string
	TableFieldName string
	val            string
}

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

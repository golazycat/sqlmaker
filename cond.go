package sqlmaker

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	_EQ     = "%s=%s {}"
	_NOTEQ  = "%s!=%s {}"
	_AND    = "AND {}"
	_ANDALL = "AND ({})"
	_OR     = "OR {}"
	_ORALL  = "OR ({})"
	_LT     = "%s>%s {}"
	_ST     = "%s<%s {}"
	_LTEQ   = "%s>=%s {}"
	_STEQ   = "%s<=%s {}"
	_IN     = "%s IN (%s) {}"
	_NOTIN  = "%s NOT IN (%s) {}"
)

type Cond struct {
	ops []string
}

func NewCond() *Cond {
	return &Cond{ops: make([]string, 0)}
}

func (cond *Cond) Eq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_EQ, k, getVal(v)))
	return cond
}

func (cond *Cond) NotEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_NOTEQ, k, getVal(v)))
	return cond
}

func (cond *Cond) And() *Cond {
	cond.ops = append(cond.ops, _AND)
	return cond
}

func (cond *Cond) AndAll() *Cond {
	cond.ops = append(cond.ops, _ANDALL)
	return cond
}

func (cond *Cond) Or() *Cond {
	cond.ops = append(cond.ops, _OR)
	return cond
}

func (cond *Cond) OrAll() *Cond {
	cond.ops = append(cond.ops, _ORALL)
	return cond
}

func (cond *Cond) Lt(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_LT, k, getVal(v)))
	return cond
}

func (cond *Cond) St(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_ST, k, getVal(v)))
	return cond
}

func (cond *Cond) LtEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_LTEQ, k, getVal(v)))
	return cond
}

func (cond *Cond) StEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_STEQ, k, getVal(v)))
	return cond
}

func (cond *Cond) In(k string, vs []interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_IN, k, getManyVal(vs)))
	return cond
}

func (cond *Cond) NotIn(k string, vs []interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_NOTIN, k, getManyVal(vs)))
	return cond
}

func (cond *Cond) Make() string {
	if len(cond.ops) == 0 {
		return ""
	}

	cond.ops = append(cond.ops, "")

	res := "{}"
	for _, op := range cond.ops {
		res = strings.ReplaceAll(res, "{}", op)
	}

	return strings.Trim(res, " ")
}

func getManyVal(vs []interface{}) string {
	vals := make([]string, len(vs))
	for _, v := range vs {
		vals = append(vals, getVal(v))
	}
	return strings.Join(vals, ",")
}

func getVal(v interface{}) string {

	switch v.(type) {
	case string:
		return stringValue(v.(string))
	case int:
		return strconv.Itoa(v.(int))
	case time.Time:
		return dateToString(v.(time.Time))
	default:
		return stringValue(fmt.Sprintf("%v", v))
	}

}

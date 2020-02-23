package sqlmaker

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 所有的条件表达式格式，%s会被替换为具体的值
// {}表示下一个表达式出现的位置
const (
	_EQ     = "%s=%s {}"
	_NOTEQ  = "%s!=%s {}"
	_AND    = "AND {}"
	_ANDALL = "AND ( {})"
	_OR     = "OR {}"
	_ORALL  = "OR ( {})"
	_LT     = "%s>%s {}"
	_ST     = "%s<%s {}"
	_LTEQ   = "%s>=%s {}"
	_STEQ   = "%s<=%s {}"
	_IN     = "%s IN (%s) {}"
	_NOTIN  = "%s NOT IN (%s) {}"
	_ENDALL = "endall"
)

// 用于构建条件表达式
type Cond struct {
	// 所有的条件集合
	ops []string
}

// 新建一个空的条件表达式
func NewCond() *Cond {
	return &Cond{ops: make([]string, 0)}
}

// 新增一个相等条件
func (cond *Cond) Eq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_EQ, k, getVal(v)))
	return cond
}

// 新增一个不相等条件
func (cond *Cond) NotEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_NOTEQ, k, getVal(v)))
	return cond
}

// 新增一个AND条件连接符
func (cond *Cond) And() *Cond {
	cond.ops = append(cond.ops, _AND)
	return cond
}

// 新增一个AND条件连接符和一对括号，接下来的条件都会在括号中
func (cond *Cond) AndAll() *Cond {
	cond.ops = append(cond.ops, _ANDALL)
	return cond
}

// 新增一个OR条件连接符
func (cond *Cond) Or() *Cond {
	cond.ops = append(cond.ops, _OR)
	return cond
}

// 新增一个OR条件连接符和一对括号，接下来的条件都会在括号中
func (cond *Cond) OrAll() *Cond {
	cond.ops = append(cond.ops, _ORALL)
	return cond
}

// 跳出当前括号，接下来的条件会从括号后面开始
func (cond *Cond) EndAll() *Cond {
	cond.ops = append(cond.ops, _ENDALL)
	return cond
}

// 新增一个大于条件
func (cond *Cond) Lt(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_LT, k, getVal(v)))
	return cond
}

// 新增一个小于条件
func (cond *Cond) St(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_ST, k, getVal(v)))
	return cond
}

// 新增一个大于等于条件
func (cond *Cond) LtEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_LTEQ, k, getVal(v)))
	return cond
}

// 新增一个小于等于条件
func (cond *Cond) StEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_STEQ, k, getVal(v)))
	return cond
}

// 新增一个IN条件
func (cond *Cond) In(k string, vs []interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_IN, k, getManyVal(vs)))
	return cond
}

// 新增一个NOT IN条件
func (cond *Cond) NotIn(k string, vs []interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_NOTIN, k, getManyVal(vs)))
	return cond
}

// 根据所有增加的条件生成条件表达式
func (cond *Cond) Make() string {
	if len(cond.ops) == 0 {
		return ""
	}

	cond.ops = append(cond.ops, "")

	res := "{}"
	for _, op := range cond.ops {
		if op == _ENDALL {
			res = strings.ReplaceAll(res, "{})", ") {}")
		} else {
			res = strings.ReplaceAll(res, "{}", op)
		}
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

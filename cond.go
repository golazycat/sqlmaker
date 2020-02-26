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
	_LIKE   = "%s LIKE %s"
	_ENDALL = "endall"
)

// 用于构建条件表达式
// 表达式由很多条件和条件连接符组成，通过调用该结构体的函数可以
// 顺序组织这些条件和条件连接符，最终生成SQL识别的条件表达式
type Cond struct {

	// 所有的条件集合
	ops []string

	// 表达式的所有值，在需要prepare的时候使用
	values []interface{}
}

// 新建一个空的条件表达式(不再建议使用)
// 这样构建出来的表达式的value会直接赋值，非常不安全(可能产生SQL注入攻击)
// Deprecated: 请使用NewPrepareCond函数作为替代
func NewCond() *Cond {
	return &Cond{
		ops:    make([]string, 0),
		values: nil,
	}
}

// 新建一个空的Prepare表达式
// 这样生成的表达式的value将会为?，可以通过Cond.Values取得对应的值
func NewPrepareCond() *Cond {
	return &Cond{
		ops:    make([]string, 0),
		values: make([]interface{}, 0),
	}
}

// 新增一个相等条件
func (cond *Cond) Eq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_EQ, k, cond.getVal(v)))
	return cond
}

// 新增一个不相等条件
func (cond *Cond) NotEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_NOTEQ, k, cond.getVal(v)))
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

// 新增一个LIKE条件
func (cond *Cond) Like(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_LIKE, k, cond.getVal(v)))
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
	cond.ops = append(cond.ops, fmt.Sprintf(_LT, k, cond.getVal(v)))
	return cond
}

// 新增一个小于条件
func (cond *Cond) St(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_ST, k, cond.getVal(v)))
	return cond
}

// 新增一个大于等于条件
func (cond *Cond) LtEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_LTEQ, k, cond.getVal(v)))
	return cond
}

// 新增一个小于等于条件
func (cond *Cond) StEq(k string, v interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_STEQ, k, cond.getVal(v)))
	return cond
}

// 新增一个IN条件
func (cond *Cond) In(k string, vs []interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_IN, k, cond.getManyVal(vs)))
	return cond
}

// 新增一个NOT IN条件
func (cond *Cond) NotIn(k string, vs []interface{}) *Cond {
	cond.ops = append(cond.ops, fmt.Sprintf(_NOTIN, k, cond.getManyVal(vs)))
	return cond
}

// 如果使用的是prepare表达式，返回所有"?"替换符对应的值
func (cond *Cond) Values() []interface{} {
	return cond.values
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

func (cond *Cond) getManyVal(vs []interface{}) string {
	vals := make([]string, len(vs))
	for _, v := range vs {
		vals = append(vals, cond.getVal(v))
	}
	return strings.Join(vals, ",")
}

func (cond *Cond) getVal(v interface{}) string {

	if cond.values != nil {
		cond.values = append(cond.values, v)
		return "?"
	}

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

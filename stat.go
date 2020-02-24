package sqlmaker

import (
	"fmt"
	"strconv"
	"strings"
)

// SQL子句格式，maker将会对%s进行替换
const (
	_FROM   = "FROM %s"
	_INSERT = "INSERT INTO %s(%s)"
	_VALUES = "VALUES(%s)"
	_SELECT = "SELECT %s"
	_WHERE  = "WHERE %s"
	_UPDATE = "UPDATE %s"
	_SET    = "SET %s"
	_DELETE = "DELETE FROM %s"
	_LIMIT  = "LIMIT %s"
)

// SQL子句生成器，用于根据Entity生成所有已知的SQL子句
type StatMaker struct {
	fields    []Field
	entity    Entity
	filter    []string
	built     bool
	tableName string
	prepare   bool
}

// 创建一个SQL子句生成器，需要传入entity表示这个生成器是针对哪个实体的
// 生成的SQL子句替换串将会从这个对象中获取
func NewStatMaker(entity Entity) StatMaker {
	return StatMaker{
		entity:    entity,
		built:     false,
		tableName: entity.TableName(),
		prepare:   true,
	}
}

// 构建生成器，在调用这个函数之前，生成器并不会解析entity，但是当调用这个函数
// 之后，生成器就会实际的解析entity。在调用Make函数之前，必须调用这个函数
func (maker *StatMaker) Build() {
	if !maker.built {
		maker.fields = decodeEntity(maker.entity, maker.filter)
		maker.built = true
	}
}

// 设置过滤字段名称。如果希望输出的SQL子句只包含entity的部分字段，需要在调用
// Make前调用该函数，传入希望输出的字段名称
func (maker *StatMaker) Filter(filter []string) {
	maker.filter = filter
}

func (maker *StatMaker) Prepare(prepare bool) {
	maker.prepare = prepare
}

// 生成FROM子句，需要用到表名
func (maker *StatMaker) MakeFrom() string {
	return fmt.Sprintf(_FROM, maker.tableName)
}

// 生成INSERT子句，需要用到表名和字段名
func (maker *StatMaker) MakeInsert() string {
	return fmt.Sprintf(_INSERT, maker.tableName, maker.makeNames())
}

// 生成VALUES子句，需要用到字段值
func (maker *StatMaker) MakeValues() string {
	return fmt.Sprintf(_VALUES, maker.makeValues())
}

// 生成SELECT子句，有两种子句，取决于count是否为true
// 如果为true，则生成默认的统计子句"SELECT COUNT(1)"
// 如果为false，则使用字段名生成
func (maker *StatMaker) MakeSelect(count bool) string {
	var stat string
	if count {
		stat = "COUNT(1)"
	} else {
		stat = maker.makeNames()
	}
	return fmt.Sprintf(_SELECT, stat)
}

// 生成WHERE子句，需要用到条件对象，关于如何构建条件表达式，见Cond
func (maker *StatMaker) MakeWhere(cond *Cond) string {
	return fmt.Sprintf(_WHERE, cond.Make())
}

// 生成UPDATE子句，需要用到表名
func (maker *StatMaker) MakeUpdate() string {
	return fmt.Sprintf(_UPDATE, maker.tableName)
}

// 生成SET子句，需要用到所有字段的"fieldName=value"格式的等式
func (maker *StatMaker) MakeSet() string {
	return fmt.Sprintf(_SET, maker.makeEquals())
}

// 生成DELETE子句，需要用到表名
func (maker *StatMaker) MakeDelete() string {
	return fmt.Sprintf(_DELETE, maker.tableName)
}

// 生成LIMIT子句
func (maker *StatMaker) MakeLimit(limit, offset int) string {
	if offset == -1 {
		return fmt.Sprintf(_LIMIT, strconv.Itoa(limit))
	} else {
		return fmt.Sprintf(_LIMIT, fmt.Sprintf(
			"%d,%d", limit, offset))
	}
}

func (maker *StatMaker) GetValues() []interface{} {
	ret := make([]interface{}, 0)
	for _, field := range maker.fields {
		ret = append(ret, field.originVal)
	}
	return ret
}

func (maker *StatMaker) GetNames() []string {
	ret := make([]string, 0)
	for _, field := range maker.fields {
		ret = append(ret, field.Name)
	}
	return ret
}

// 生成字段的"fieldName=value"表达式
func (maker *StatMaker) makeEquals() string {
	var genFunc genStatFunc
	if !maker.prepare {
		genFunc = func(field Field) string {
			return fmt.Sprintf("`%s`=%s",
				field.TableFieldName, field.val)
		}
	} else {
		genFunc = func(field Field) string {
			return fmt.Sprintf("`%s`=?",
				field.TableFieldName)
		}
	}
	return maker.makeStat(genFunc)
}

// 生成所有字段的值
func (maker *StatMaker) makeValues() string {
	var genFunc genStatFunc
	if !maker.prepare {
		genFunc = func(field Field) string {
			return field.val
		}
	} else {
		genFunc = func(Field) string {
			return "?"
		}
	}
	return maker.makeStat(genFunc)
}

// 生成所有字段的名称
func (maker *StatMaker) makeNames() string {
	return maker.makeStat(func(field Field) string {
		return "`" + field.TableFieldName + "`"
	})
}

type genStatFunc func(Field) string

func (maker *StatMaker) makeStat(statFunc genStatFunc) string {

	stat := make([]string, 0)
	for _, field := range maker.fields {
		stat = append(stat, statFunc(field))
	}

	return strings.Join(stat, ",")

}

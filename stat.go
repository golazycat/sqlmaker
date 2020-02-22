package sqlmaker

import (
	"fmt"
	"strings"
)

const (
	_FROM   = "FROM %s"
	_INSERT = "INSERT INTO %s(%s)"
	_VALUES = "VALUES(%s)"
	_SELECT = "SELECT %s"
	_WHERE  = "WHERE %s"
	_UPDATE = "UPDATE FROM %s"
	_SET    = "SET %s"
	_DELETE = "DELETE FROM %s"
)

type StatMaker struct {
	fields    []Field
	entity    Entity
	filter    []string
	built     bool
	tableName string
}

func NewStatMaker(entity Entity) StatMaker {
	return StatMaker{
		entity:    entity,
		built:     false,
		tableName: entity.TableName(),
	}
}

func (maker *StatMaker) Build() {
	if !maker.built {
		maker.fields = decodeEntity(maker.entity, maker.filter)
		maker.built = true
	}
}

func (maker *StatMaker) Filter(filter []string) {
	maker.filter = filter
}

func (maker *StatMaker) MakeFrom() string {
	return fmt.Sprintf(_FROM, maker.tableName)
}

func (maker *StatMaker) MakeInsert() string {
	return fmt.Sprintf(_INSERT, maker.tableName, maker.makeNames())
}

func (maker *StatMaker) MakeValues() string {
	return fmt.Sprintf(_VALUES, maker.makeValues())
}

func (maker *StatMaker) MakeSelect(count bool) string {
	var stat string
	if count {
		stat = "COUNT(1)"
	} else {
		stat = maker.makeNames()
	}
	return fmt.Sprintf(_SELECT, stat)
}

func (maker *StatMaker) MakeWhere(cond *Cond) string {
	return fmt.Sprintf(_WHERE, cond.Make())
}

func (maker *StatMaker) MakeUpdate() string {
	return fmt.Sprintf(_UPDATE, maker.tableName)
}

func (maker *StatMaker) MakeSet() string {
	return fmt.Sprintf(_SET, maker.makeEquals())
}

func (maker *StatMaker) MakeDelete() string {
	return fmt.Sprintf(_DELETE, maker.tableName)
}

func (maker *StatMaker) makeEquals() string {
	return maker.makeStat(func(field Field) string {
		return fmt.Sprintf("`%s`=%s",
			field.TableFieldName, field.val)
	})
}

func (maker *StatMaker) makeValues() string {
	return maker.makeStat(func(field Field) string {
		return field.val
	})
}

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

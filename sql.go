package sqlmaker

import (
	"errors"
	"strings"
)

var MakerNotBuildError = errors.New("maker not build")

type SqlMaker struct {
	maker     StatMaker
	split     string
	statOrder []string
	cond      *Cond
	built     bool
	isCount   bool
	idName    string
	idValue   interface{}
}

func (maker *SqlMaker) Filter(fs ...string) *SqlMaker {
	maker.maker.Filter(fs)
	return maker
}

func (maker *SqlMaker) Cond(cond *Cond) *SqlMaker {
	maker.cond = cond
	return maker
}

func (maker *SqlMaker) Split(s string) *SqlMaker {
	maker.split = s
	return maker
}

func (maker *SqlMaker) Beauty() *SqlMaker {
	return maker.Split("\n")
}

func (maker *SqlMaker) BuildMake() string {
	maker.Build()
	return maker.MustMake()
}

func (maker *SqlMaker) Build() *SqlMaker {
	maker.maker.Build()
	maker.built = true
	return maker
}

func (maker *SqlMaker) ByID() *SqlMaker {
	maker.cond = NewCond().Eq(maker.idName, maker.idValue)
	return maker
}

func (maker *SqlMaker) Count() *SqlMaker {
	maker.isCount = true
	return maker
}

func (maker *SqlMaker) Make() (string, error) {

	if !maker.built {
		return "", MakerNotBuildError
	}

	sql := make([]string, 0)
	for _, stat := range maker.statOrder {
		switch stat {
		case "from":
			sql = append(sql, maker.maker.MakeFrom())
		case "insert":
			sql = append(sql, maker.maker.MakeInsert())
		case "values":
			sql = append(sql, maker.maker.MakeValues())
		case "update":
			sql = append(sql, maker.maker.MakeUpdate())
		case "set":
			sql = append(sql, maker.maker.MakeSet())
		case "where":
			if maker.cond == nil {
				continue
			}
			sql = append(sql, maker.maker.MakeWhere(maker.cond))
		case "delete":
			sql = append(sql, maker.maker.MakeDelete())

		}
	}

	return strings.Join(sql, maker.split), nil
}

func (maker *SqlMaker) MustMake() string {
	s, err := maker.Make()
	if err != nil {
		panic(err)
	}
	return s
}

func NewInsertMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"insert", "values"})
}

func NewUpdateMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"update", "set", "where"})
}

func NewDeleteMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"delete", "where"})
}

func NewSearchMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"select", "from", "where"})
}

func newSqlMaker(e Entity, statOrder []string) *SqlMaker {
	idName, idValue := e.GetId()
	return &SqlMaker{
		maker:     NewStatMaker(e),
		split:     " ",
		statOrder: statOrder,
		cond:      nil,
		built:     false,
		idName:    stringName(idName),
		idValue:   idValue,
	}

}

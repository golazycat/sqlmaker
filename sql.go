package sqlmaker

import (
	"errors"
	"strings"
)

// 在使用SqlMaker的时候，如果在Make前没有Build，会返回这个错误
var MakerNotBuildError = errors.New("maker not build")

// SqlMaker结构体，所有SQL语句都通过该结构体的函数生成
// 结构体有很多配置函数，可以配置生成SQL
type SqlMaker struct {
	maker     StatMaker
	split     string
	statOrder []string
	cond      *Cond
	built     bool
	isCount   bool
	idName    string
	idValue   interface{}
	limit     int
	offset    int
}

// 设置过滤字段名称。如果希望输出的SQL子句只包含entity的部分字段，需要在调用
// Make前调用该函数，传入希望输出的字段名称
func (maker *SqlMaker) Filter(fs ...string) *SqlMaker {
	maker.maker.Filter(fs)
	return maker
}

// 设置条件表达式，如果生成的SQL语句有WHERE条件子句，需要调用这个函数设置条件
// 如果不调用，则不会生成WHERE子句
func (maker *SqlMaker) Cond(cond *Cond) *SqlMaker {
	maker.cond = cond
	return maker
}

// 设置不同子句的间隔符，默认是" "
func (maker *SqlMaker) Split(s string) *SqlMaker {
	maker.split = s
	return maker
}

// 将子句的间隔符设置为换行，这样可以使生成的SQL更易读
func (maker *SqlMaker) Beauty() *SqlMaker {
	return maker.Split("\n")
}

// Build()之后调用MustMake()，一次性生成SQL语句返回
func (maker *SqlMaker) BuildMake() string {
	maker.Build()
	return maker.MustMake()
}

// 构建SQL语句，但是不生成，这会解析entity对象
// 调用该函数之后就可以调用Make()生成SQL语句了
// 当entity对象改变时，注意Make()仍然返回之前那个entity的SQL语句
// 如果需要生成新的，需要重新调用该函数
func (maker *SqlMaker) Build() *SqlMaker {
	maker.maker.Build()
	maker.built = true
	return maker
}

// 直接将条件设置为根据ID查询。这需要entity通过getId()函数返回id字段名和值
// 这样可以直接将WHERE子句设置为idName=idValue
func (maker *SqlMaker) ByID() *SqlMaker {
	maker.cond = NewCond().Eq(maker.idName, maker.idValue)
	return maker
}

// 查询结果以统计的方式返回。这将SELECT子句设置为"SELECT COUNT(1)"
// 会返回查询到的数量而不是数据
func (maker *SqlMaker) Count() *SqlMaker {
	maker.isCount = true
	return maker
}

// 设置查询的Limit参数
func (maker *SqlMaker) Limit(limit, offset int) *SqlMaker {
	maker.limit = limit
	maker.offset = offset
	return maker
}

// 分页
func (maker *SqlMaker) Page(curPage, pageSize int) *SqlMaker {
	return maker.Limit((curPage-1)*pageSize, pageSize)
}

// 生成SQL语句
// 这会根据配置和解析的entity生成SQL语句，注意调用该函数前必须调用Build()函数
// 否则会返回MakerNotBuildError错误
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
		case "select":
			sql = append(sql, maker.maker.MakeSelect(maker.isCount))
		case "limit":
			if maker.limit == -1 {
				continue
			}
			sql = append(sql, maker.maker.MakeLimit(
				maker.limit, maker.offset))

		}
	}

	return strings.Join(sql, maker.split), nil
}

// 和Make()一样，但是如果没有Build()，会直接panic
func (maker *SqlMaker) MustMake() string {
	s, err := maker.Make()
	if err != nil {
		panic(err)
	}
	return s
}

// 新建一个新建SQL语句生成器
func NewInsertMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"insert", "values"})
}

// 新建一个更新SQL语句生成器
func NewUpdateMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"update", "set", "where"})
}

// 新建一个删除SQL语句生成器
func NewDeleteMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"delete", "where"})
}

// 新建一个查询语句生成器
func NewSearchMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"select", "from", "where", "limit"})
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
		limit:     -1,
		offset:    -1,
	}

}

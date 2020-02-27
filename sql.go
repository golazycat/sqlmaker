package sqlmaker

import "C"
import (
	"database/sql"
	"errors"
	"strings"
)

// 在使用SqlMaker的时候，如果在Make前没有Build，会返回这个错误
var (
	MakerNotBuildError         = errors.New("maker not build")
	defaultDB          *sql.DB = nil
)

func SetDefaultDB(db *sql.DB) {
	defaultDB = db
}

// SqlMaker结构体，所有SQL语句都通过该结构体的函数生成
// 另外在exec中实现了一键执行SQL，但是在调用exec的函数前
// 必须调用SetDb()函数
type SqlMaker struct {

	// SQL子句生成器，SQL语句由很多子句构成，因此需要子句生成器
	// 来生成具体的子句
	maker StatMaker

	// 不同SQL子句之间的分隔符
	split string

	// SQL子句的顺序。不同SQL语句的子句类型和顺序都不同
	// 该属性保存给SQL语句的子句的名称和顺序
	statOrder []string

	// SQL条件，如果该SQL语句有WHERE子句，并且需要为WHERE
	// 设定条件，则需要Cond来生成条件
	cond *Cond

	// Maker是否已经被构建
	built bool

	// 该SQL是否是统计语句，如果是，则SELECT子句为COUNT(1)
	isCount bool

	// entity的id值，通过Entity接口函数GetId()获取
	idName string

	// entity的value值，通过Entity接口函数GetId()获取
	idValue interface{}

	// LIMIT分页参数
	limit  int
	offset int

	// 如果需要执行SQL语句，必须为db赋值
	db *sql.DB
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

// 设定生成的SQL是否为prepare语句
// 为了安全起见，默认情况下生成的SQL均是prepare语句
// 该调用并不一定会影响exec的时候是否真正按照prepare执行，详见IsPrepare说明
func (maker *SqlMaker) Prepare(prepare bool) *SqlMaker {
	maker.maker.Prepare(prepare)
	return maker
}

// 判断SQL是否为prepare语句
// 注意，如果当前SQL即不存在可能产生prepare的子句，
// 如SET、VALUES子句，也没有WHERE查询条件，则直接认为该SQL不是prepare语句
// 即使之前通过Prepare调用将SqlMaker设为prepare的
// 这是为了节约性能，例如语句："SELECT COUNT(1) FROM user"
// 该语句没有需要替换的地方，没必要将其作为prepare语句使用(也不会产生危险)
// 因此即使之前显示调用过Prepare(true)，该SQL对应的Maker的IsPrepare()函数仍然会返回false
// 这个函数非常重要，在exec执行的时候依据这个函数判断是否使用PrepareStmt
func (maker *SqlMaker) IsPrepare() bool {

	if !maker.hasPrepareStat() && maker.cond == nil {
		return false
	}

	return maker.maker.prepare ||
		(maker.cond != nil && maker.cond.values != nil)
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
	if maker.IsPrepare() {
		maker.cond = NewPrepareCond().Eq(maker.idName, maker.idValue)
	} else {
		maker.cond = NewPrepareCond().Eq(maker.idName, maker.idValue)
	}
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

// 分页，这会自动根据curPage和pageSize来计算LIMIT参数
func (maker *SqlMaker) Page(curPage, pageSize int) *SqlMaker {
	return maker.Limit((curPage-1)*pageSize, pageSize)
}

// 获取entity的所有属性名(golang中的，而不是表字段名)
// 这个函数会受到Filter()的影响，一般在查询后反射时调用
func (maker *SqlMaker) Names() []string {
	return maker.maker.GetNames()
}

// 返回entity中prepare好的values。如果SQL语句是prepare的，则
// 生成的SQL不会包含value，而是"?"占位符。在执行的时候需要传入真正的
// value。这个函数就会通过prepare的具体情况，来返回prepare SQL中占位符对应的值。
// 这个函数的返回值可以直接传给`sql.Stmt`结构体的Exec()或Query()函数
func (maker *SqlMaker) Values() []interface{} {
	if !maker.hasPrepareStat() {
		maker.maker.prepare = false
	}
	if maker.cond != nil && maker.cond.values != nil {
		if maker.maker.prepare {
			return append(maker.maker.GetValues(), maker.cond.values...)
		}
		return maker.cond.values
	}
	if maker.maker.prepare {
		return maker.maker.GetValues()
	} else {
		return make([]interface{}, 0)
	}
}

// 判断该SQL是否包含潜在的需要prepare的子句
func (maker *SqlMaker) hasPrepareStat() bool {
	for _, statType := range maker.statOrder {
		switch statType {
		case "set":
			return true
		case "values":
			return true
		}
	}
	return false
}

// 生成SQL语句
// 这会根据配置和解析的entity生成SQL语句，注意调用该函数前必须调用Build()函数
// 否则会返回MakerNotBuildError错误
func (maker *SqlMaker) Make() (string, error) {

	if !maker.built {
		return "", MakerNotBuildError
	}

	_sql := make([]string, 0)
	for _, stat := range maker.statOrder {
		switch stat {
		case "from":
			_sql = append(_sql, maker.maker.MakeFrom())
		case "insert":
			_sql = append(_sql, maker.maker.MakeInsert())
		case "replace":
			_sql = append(_sql, maker.maker.MakeReplace())
		case "values":
			_sql = append(_sql, maker.maker.MakeValues())
		case "update":
			_sql = append(_sql, maker.maker.MakeUpdate())
		case "set":
			_sql = append(_sql, maker.maker.MakeSet())
		case "where":
			if maker.cond == nil {
				continue
			}
			_sql = append(_sql, maker.maker.MakeWhere(maker.cond))
		case "delete":
			_sql = append(_sql, maker.maker.MakeDelete())
		case "select":
			_sql = append(_sql, maker.maker.MakeSelect(maker.isCount))
		case "limit":
			if maker.limit == -1 {
				continue
			}
			_sql = append(_sql, maker.maker.MakeLimit(
				maker.limit, maker.offset))

		}
	}

	return strings.Join(_sql, maker.split), nil
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

// 新建一个新建SQL语句生成器
func NewReplaceMaker(e Entity) *SqlMaker {
	return newSqlMaker(e, []string{"replace", "values"})
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
func NewQueryMaker(e Entity) *SqlMaker {
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

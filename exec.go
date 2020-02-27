package sqlmaker

import (
	"database/sql"
	"errors"
	"io"
	"log"
)

var DBNotSetError = errors.New("db is not set")

// 查询的返回结果
type QueryResult struct {
	// 这实际上相当于查询到的Columns名称
	// 但是这是golang结构体中的属性名称
	names []string

	// 每一行的具体值
	valuesTable [][]interface{}
}

// 结果是否还有剩余数据，一般用于迭代，如果返回false，表示已经没有剩余数据了
// 调用QueryResult.Decode会消耗剩余数据
func (result *QueryResult) Next() bool {
	return len(result.valuesTable) > 0
}

// 将当前行返回数据解码为结构体，随后前进到下一行。该函数一般和QueryResult.Next配合使用
// 当Next()返回false时，该函数就不可以被继续调用了
// 注意参数o必须是一个指针，这样在调用后它指向的结构体就会被设置为该行对应的数据了
// 请确保o指向的结构体和Maker对应的entity是一致的，否则在reflect的时候会产生panic级别的错误
func (result *QueryResult) Decode(o interface{}) {
	values := result.valuesTable[0]
	setValues(o, result.names, values)
	result.valuesTable = result.valuesTable[1:]
}

// 为Maker设置db对象，该函数是为调用SQL执行函数做准备的
// 如果在调用各种SQL执行函数前没有调用该函数，则会返回DBNotSetError
func (maker *SqlMaker) SetDB(db *sql.DB) *SqlMaker {
	maker.db = db
	return maker
}

func (maker *SqlMaker) checkDB() bool {
	if maker.db == nil {
		if defaultDB == nil {
			return false
		}
		maker.db = defaultDB
	}
	return true
}

// 执行SQL语句，返回执行影响的数据行数
func (maker *SqlMaker) Exec() (int64, error) {

	if !maker.checkDB() {
		return 0, DBNotSetError
	}

	_sql := maker.BuildMake()
	var (
		result sql.Result
		err    error
	)
	wLog("### Exec SQL: %s", _sql)

	// prepare和non-prepare逻辑不同
	if maker.IsPrepare() {
		_, result, err = maker.execPrepare(_sql, false)
	} else {
		result, err = maker.db.Exec(_sql)
	}

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// 执行查询多个数据SQL，返回的QueryResult对象可以迭代，通过迭代QueryResult
// 来将查询结果转换为具体的entity。
func (maker *SqlMaker) ExecQueryMany() (*QueryResult, error) {
	return maker.execQuery(true, false, nil, nil)
}

// 执行查询单个数据，确认SQL只会返回一个数据时调用该函数
// 单个数据通过传入指针的方式赋值，确保o是一个指针
func (maker *SqlMaker) ExecQueryOne(o interface{}) error {

	_, err := maker.execQuery(false, false, o, nil)
	return err
}

// 执行统计数据，如果SQL是统计的数据，返回的结果是一个整数，则可以调用该函数
func (maker *SqlMaker) ExecCount() (int, error) {

	if !maker.isCount {
		maker.Count()
	}

	var cnt int
	_, err := maker.execQuery(false, true, nil, &cnt)
	return cnt, err
}

// 通过search函数，囊括了上述三种查询
func (maker *SqlMaker) execQuery(many, count bool, o interface{}, i *int) (*QueryResult, error) {

	if !maker.checkDB() {
		return nil, DBNotSetError
	}

	_sql := maker.BuildMake()
	var (
		rows *sql.Rows
		err  error
	)

	wLog("### Exec query sql: %s", _sql)

	if maker.IsPrepare() {
		rows, _, err = maker.execPrepare(_sql, true)
	} else {
		rows, err = maker.db.Query(_sql)
	}

	if err != nil {
		return nil, err
	}

	names := maker.Names()
	valuesTable := make([][]interface{}, 0)

	for rows.Next() {

		// 统计，直接将结果赋值为int后返回
		if count {
			err = rows.Scan(i)
			return nil, err
		}

		// 保存当前row返回的值
		values := make([]interface{}, len(names))

		// 指向values所有元素的指针，用于给values赋值
		valuePts := make([]interface{}, len(names))
		for i := range valuePts {
			valuePts[i] = &values[i]
		}

		// 通过valuePts间接向values赋值
		err = rows.Scan(valuePts...)
		if err != nil {
			return nil, err
		}
		if !many {
			// 只用解析第一个数据即可返回
			setValues(o, names, values)
			return nil, nil
		}

		valuesTable = append(valuesTable, values)
	}

	return &QueryResult{
		names:       names,
		valuesTable: valuesTable,
	}, nil
}

// 指向PrepareSQL语句，这会创建一个SQL stmt，随后调用maker的Values()函数获取具体的
// 值传给stmt执行，详情见SqlMaker.Values文档
func (maker *SqlMaker) execPrepare(_sql string, isQuery bool) (*sql.Rows, sql.Result, error) {
	stmt, err := maker.db.Prepare(_sql)
	if err != nil {
		return nil, nil, err
	}
	defer safeClose(stmt)

	wLog("### prepare values: %s", printValues(maker.Values()))

	if isQuery {
		rows, err := stmt.Query(maker.Values()...)
		if err != nil {
			return nil, nil, err
		}
		return rows, nil, nil
	} else {
		result, err := stmt.Exec(maker.Values()...)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}

}

func safeClose(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Printf("error close: %s", err)
	}
}

func setValues(o interface{}, names []string, values []interface{}) {
	for i, name := range names {
		setValue(o, name, values[i])
	}
}

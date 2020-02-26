package sqlmaker

import (
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("mysql", "root:19971008@tcp(127.0.0.1:3306)/test?charset=utf8")
	db.SetMaxOpenConns(1000)
	err := db.Ping()
	if err != nil {
		fmt.Println("Failed to connect to mysql, err:" + err.Error())
		os.Exit(1)
	}
}

type User struct {
	Id         int       `field:"id"`
	Name       string    `field:"name"`
	Age        int       `field:"age"`
	Phone      string    `field:"phone"`
	CreateDate time.Time `field:"create_date"`
	Status     int       `field:"status"`
}

func (t User) GetId() (string, interface{}) {
	return "id", t.Id
}

func (t User) TableName() string {
	return "user"
}

var user = User{
	Id:         3,
	Name:       "Mike",
	Age:        18,
	Phone:      "78231234",
	CreateDate: time.Now(),
	Status:     2,
}

func TestInsert(t *testing.T) {

	DebugMode()

	user.Id = 5
	maker := NewInsertMaker(user).SetDB(db)

	fmt.Println(maker.BuildMake())
	affect, err := maker.Exec()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(affect)
}

func TestUpdate(t *testing.T) {

	user.Name = "Tang"
	maker := NewUpdateMaker(user).ByID().SetDB(db).Filter("name", "age")

	affect, err := maker.Exec()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(affect)

}

func TestQueryMany(t *testing.T) {

	DebugMode()

	// 查询全部
	cond := NewPrepareCond().Like("name", "%T%")
	maker := NewQueryMaker(user).SetDB(db).Cond(cond).Page(1, 10)
	result, err := maker.ExecQueryMany()

	if err != nil {
		panic(err)
	}

	for result.Next() {
		u := User{}
		result.Decode(&u)
		fmt.Println(u)
	}

	// 按照条件查询
	cond = NewPrepareCond().Eq("name", "Tang").And().Eq("age", 18)
	maker = NewQueryMaker(user).SetDB(db).Cond(cond)
	result, err = maker.ExecQueryMany()
	if err != nil {
		panic(err)
	}
	for result.Next() {
		u := User{}
		result.Decode(&u)
		fmt.Println(u)
	}

	// 查询单个数据
	maker = NewQueryMaker(user).ByID().SetDB(db)
	u := User{}
	_ = maker.ExecQueryOne(&u)
	fmt.Println("Query One: ", u)

	// 统计个数
	cond = NewPrepareCond().St("age", 23)
	maker = NewQueryMaker(user).SetDB(db).Cond(cond)
	cnt, _ := maker.ExecCount()
	fmt.Println("count: ", cnt)

}

func TestOther(t *testing.T) {

	i := 1
	changeVal(&i)
	fmt.Println(i)

}

func changeVal(o *int) {
	*o = 12
}

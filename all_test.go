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

func TestSet(t *testing.T) {
	u := User{}
	setValue(&u, "CreateDate", "2020-01-02 10:10:23")
	fmt.Println(u)
}

func TestInsert(t *testing.T) {

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

}

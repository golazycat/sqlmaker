package sqlmaker

import (
	"fmt"
	"testing"
	"time"
)

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
	Id:         1,
	Name:       "John",
	Age:        23,
	Phone:      "7891234",
	CreateDate: time.Now(),
	Status:     2,
}

func TestInsert(t *testing.T) {

	sql := NewInsertMaker(user).BuildMake()
	fmt.Println(sql)

	sql = NewInsertMaker(user).Beauty().BuildMake()
	fmt.Println(sql)

	sql = NewInsertMaker(user).Beauty().Filter("id", "name", "age").BuildMake()
	fmt.Println(sql)

	fmt.Println(NewInsertMaker(user).Beauty().BuildMake())
	fmt.Println(NewInsertMaker(user).BuildMake())
}

func TestUpdate(t *testing.T) {

	user.Status = 3
	cond := NewCond().Lt("age", 80).And().NotEq("status", 3)
	sql := NewUpdateMaker(user).Cond(cond).Filter("status").Beauty().BuildMake()
	fmt.Println(sql)

	sql = NewUpdateMaker(user).ByID().Filter("name").Beauty().BuildMake()
	fmt.Println(sql)

	cond = NewCond().Eq("name", "mike").AndAll().
		StEq("age", 20).Or().Eq("phone", 110)
	sql = NewDeleteMaker(user).Cond(cond).Beauty().BuildMake()
	fmt.Println(sql)

	cond = NewCond().Lt("age", 50)
	sql = NewSearchMaker(user).Cond(cond).Count().Beauty().BuildMake()
	fmt.Println(sql)

	sql = NewSearchMaker(user).Filter("name").Beauty().BuildMake()
	fmt.Println(sql)
}

func TestCond(t *testing.T) {
	cond := NewCond().Eq("age", 12).AndAll().Eq("c", "d").Or().Eq("name", "nihao")
	fmt.Println(cond.Make())

	cond = NewCond().Lt("age", 23).OrAll().Eq("name", "asd").And().Eq("age", 14)
	fmt.Println(cond.Make())

	cond = NewCond().Lt("age", 23).OrAll().Eq("name", "asd").And()
	fmt.Println(cond.Make())

	cond = NewCond().
		Eq("name", "Tang").AndAll().
		Eq("status", 1).Or().Eq("age", 23).EndAll().Or().
		Lt("age", 56)
	fmt.Println(cond.Make())
}

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
}

func TestInsert(t *testing.T) {

	fmt.Println(NewInsertMaker(user).Beauty().BuildMake())
	fmt.Println(NewInsertMaker(user).Filter("phone", "Name", "age").BuildMake())
}

func TestCond(t *testing.T) {
	cond := NewCond().Eq("age", 12).AndAll().Eq("c", "d").Or().Eq("name", "nihao")
	fmt.Println(cond.Make())

	cond = NewCond().Lt("age", 23).OrAll().Eq("name", "asd").And().Eq("age", 14)
	fmt.Println(cond.Make())

	cond = NewCond().Lt("age", 23).OrAll().Eq("name", "asd").And()
	fmt.Println(cond.Make())
}

func TestUpdate(t *testing.T) {

	// 根据年龄更新
	cond := NewCond().Eq("age", 20)
	fmt.Println(NewUpdateMaker(user).Cond(cond).Beauty().BuildMake())

	// 直接根据ID更新
	fmt.Println(NewUpdateMaker(user).ByID().Beauty().Filter("name", "phone", "age").BuildMake())
}

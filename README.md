# sqlmaker

sqlmaker是一个简单的SQL语句生成器。使用它可以根据结构体生成CRUD SQL语句。免去了大量无聊又繁琐的SQL语句拼接工作。

## 安装

推荐使用go mod安装，执行：

```text
$ go get github.com/golazycat/sqlmaker
```

## 使用

需要生成SQL的结构体需要实现两个函数：`TableName() string`和`GetId() (string interface{})`。

并且每个字段都需要增加`field`标签来标明该字段对应的数据表中的字段名称。

以一个简单的用户表为例子：

```go
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
```

新建一个对象，下面的所有例子依据这个对象进行：

```go
var user = User{
	Id:         1,
	Name:       "John",
	Age:        23,
	Phone:      "7891234",
	CreateDate: time.Now(),
	Status:     2,
}
```

### Insert

直接根据所有字段生成INSERT语句：

```go
sql := NewInsertMaker(user).BuildMake()
fmt.Println(sql)
```

输出：

```sql
INSERT INTO user(`id`,`name`,`age`,`phone`,`create_date`,`status`) VALUES(1,'John',23,'7891234','2020-02-23 16:02:00',2)
```

使用`Beauty()`可以让输出的SQL更可读：

```go
sql := NewInsertMaker(user).Beauty().BuildMake()
fmt.Println(sql)
```

输出：

```sql
INSERT INTO user(`id`,`name`,`age`,`phone`,`create_date`,`status`)
VALUES(1,'John',23,'7891234','2020-02-23 16:02:00',2)
```

使用`Filter()`可以过滤指定属性：

```go
sql = NewInsertMaker(user).Beauty().Filter("id", "name", "age").BuildMake()
fmt.Println(sql)
```

输出：

```sql
INSERT INTO user(`id`,`name`,`age`)
VALUES(1,'John',23)
```

### Update

UPDATE语句一般需要构建条件，使用`Cond`对象可以构建条件。例如，更新那些年龄大于80岁且status不为3的用户，将status设为3：

```go
user.Status = 3
cond := NewCond().Lt("age", 80).And().NotEq("status", 3)
sql := NewUpdateMaker(user).Cond(cond).Filter("status").Beauty().BuildMake()
fmt.Println(sql)
```

输出：

```sql
UPDATE FROM user
SET `status`=3
WHERE age>80 AND status!=3
```

可以直接根据ID进行更新，例如，根据ID更新姓名：

```go
sql := NewUpdateMaker(user).ByID().Filter("name").Beauty().BuildMake()
fmt.Println(sql)
```

输出：

```sql
UPDATE FROM user
SET `name`='John'
WHERE `id`=1
```

### Delete

DELETE语句也需要构建条件，例如，删除那些name为"Mike"，并且age小于等于20或phone等于119的用户：

```go
cond := NewCond().Eq("name", "mike").AndAll().
    StEq("age", 20).Or().Eq("phone", 110)
sql := NewDeleteMaker(user).Cond(cond).Beauty().BuildMake()
fmt.Println(sql)
```

输出：

```sql
DELETE FROM user
WHERE name='mike' AND ( age<=20 OR phone=110 )
```

### Search

Search条件的使用方法和上述一样了，但是Search可以统计个数，例如，统计年龄大于50岁的用户个数：

```go
cond := NewCond().Lt("age", 50)
sql := NewSearchMaker(user).Cond(cond).Count().Beauty().BuildMake()
fmt.Println(sql)
```

输出：

```sql
SELECT COUNT(1)
FROM user
WHERE age>50
```

Search可以查询指定字段，例如，查询所有的name：

```go
sql = NewSearchMaker(user).Filter("name").Beauty().BuildMake()
fmt.Println(sql)
```

输出：

```sql
SELECT `name`
FROM user
```

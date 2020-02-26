# sqlmaker

sqlmaker是一个简单的SQL语句生成器。使用它可以根据结构体生成CRUD SQL语句。免去了大量无聊又繁琐的SQL语句拼接工作。

在v1.1中加入了执行SQL的相关函数。

## 安装

推荐使用go mod安装，执行：

```text
$ go get github.com/golazycat/sqlmaker
```

---

## 使用

使用`sqlmaker.DebugMode()`，可以在执行SQL的时候显示日志。

sqlmaker需要根据结构体生成SQL语句，这个结构体必须实现`sqlmaker.Entity`接口：

- `GetId()`函数用于取得`entity`对应表的id字段名和值。如果不调用maker的`ByID()`函数，则该函数的返回值将不会被用到。
- `Tablename()`用于取得`entity`对应的表名。

另外，每个字段都必须用`"field"`标签指定该字段在数据库中对应的字段名称。

SQL的生成和执行全部通过`sqlmaker.SqlMaker`完成。每个SQL语句需要用到不同种类的SQL。下面是各种SQL语句的生成简单示例（更多用法参见`SqlMaker`的函数列表）

为了简单，下面的所有例子都是以下面这个结构体演示的：

```golang
type User struct {
	Id         int       `field:"id"`
	Name       string    `field:"name"`
	Age        int       `field:"age"`
	Phone      string    `field:"phone"`
	CreateDate time.Time `field:"create_date"`
	Status     int       `field:"status"`
}
```

插入、修改需要一个实体，我创建了一个全局实体作为演示：

```golang
var user = User{
	Id:         3,
	Name:       "Mike",
	Age:        18,
	Phone:      "78231234",
	CreateDate: time.Now(),
	Status:     2,
}
```

在`MySQL`中，该结构体对应的表结构为：

```sql
CREATE TABLE `user` (
  `id` int(11) NOT NULL,
  `name` varchar(255) DEFAULT NULL,
  `age` int(11) DEFAULT NULL,
  `phone` varchar(255) DEFAULT NULL,
  `create_date` datetime DEFAULT NULL,
  `status` int(11) DEFAULT NULL,
  PRIMARY KEY (`id`)
)
```

### Insert语句

使用`sqlmaker.NewInsertMaker`，可以生成插入语句的maker，如果需要执行语句，还需要调用maker的`SetDB`函数，需要传入创建好连接的`*sql.DB`。

下面的例子假设连接对象已经初始化好，为`db`。

插入一条数据的例子为：

```golang
maker := NewInsertMaker(user).SetDB(db)

affect, err := maker.Exec()
```

如果插入执行成功，`affect`将会为`1`。失败`affect`为`0`，并且`err`会返回对应的错误。

### Update语句

使用`sqlmaker.NewUpdateMaker`，可以生成更新语句的maker。

假设根据ID将`user`的`name`更新为`"John"`：

```golang
user.Name = "John"
maker := NewUpdateMaker(user).ByID().SetDB(db).Filter("name")

affect, err := maker.Exec()
```

调用`ByID()`表示根据ID进行更新，调用`Filter()`设定需要更新哪些属性

可以构造一个条件`sqlmaker.Cond`，实现根据条件进行更新。

将`age`小于`15`的`status`更新为`3`，`phone`更新为`"null"`：

```golang
cond := NewPrepareCond().St("age", 20)
user.Status = 3
user.Phone = "null"
maker := NewUpdateMaker(user).Cond(cond).SetDB(db).Filter("status", "phone")

affect, err := maker.Exec()
```

### Query查询

查询涉及三种方式：

- 查询多个数据
- 查询一个数据
- 统计数据

查询多个数据，可以使用分页功能来缩减返回数据的大小，进行分页，假设`currentPage=1`，`pageSize=10`，并且增加一个模糊查询，对`name`搜索，关键字为`"Mike"`：

```golang
cond := NewPrepareCond().Like("name", "%Mike%")
maker := NewQueryMaker(user).SetDB(db).Cond(cond).Page(1, 10)
result, err := maker.ExecQueryMany()
```

返回的`result`是一个`sqlmaker.QueryResult`指针，通过`Next()`和`Decode()`函数可以将返回结果解码为`entity`结构体：

```golang
users := make([]User, 0)
for result.Next() {
	u := User{}
	result.Decode(&u)
	users = append(users, u)
}
```

查询一个数据只需要调用`SqlMaker.ExecQueryOne`即可，需要把要赋值的结构体指针传入，下面是根据ID进行查询：

```golang
maker := NewQueryMaker(user).ByID().SetDB(db)
u := User{}
err := maker.ExecQueryOne(&u)
```

如果正常，`u`就会在执行后保存查询到的结果。

统计数据会直接将统计到的`int`返回出来，调用`SqlMaker.ExecCount`即可，例如统计`age`小于`23`的：

```golang
cond := NewPrepareCond().St("age", 23)
maker := NewQueryMaker(user).SetDB(db).Cond(cond)
cnt, err := maker.ExecCount()
```

### Delete语句

Delete用法和Update差别不大，需要传入删除条件，例如，删除那些`name="Mike"`的数据：

```golang
cond := NewPrepareCond().Eq("name", "Mike")
maker := NewDeleteMaker(user).SetDB(db).Cond(cond)

affect, err := maker.Exec()
```

---

`sqlmaker`还有很多功能，关于`sqlmaker`的更多用法，请见`go doc`文档。


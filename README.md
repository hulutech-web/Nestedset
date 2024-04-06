# 关于 nestedset
nestedset 是基于goravel框架的三方扩展，一个用于在嵌套集合中进行插入、删除、更新和查询操作的 goravel 扩展包。它提供了一种简单而高效的方法来处理具有层级关系的数据。
提供了批量建树，添加数，添加子树，删除节点，更新节点，查询子树等功能，默认模型包含lft,rgt,pid字段，该扩展任在积极的加强中。
## 安装
go get -u  github.com/hulu-web/nestedset.git

## 集成
框架集成：config/app.go 中的"providers"加入: []foundation.ServiceProvider{&nestedset.ServiceProvider{}}，确保初始化时自动加载。
## 使用
nestedset 需要先定义一个模型，该模型添加nestedset字段，同时添加Children字段，定义Tablename方法，如下：
#### 模型定义
```go
type Productcategory struct {
    orm.Model
    nestedset.Nestedset
    Name     string            `gorm:"column:name;type:varchar(255);not null;" form:"name" json:"name"`
    Title    string            `gorm:"column:title;type:varchar(255);not null;" form:"title" json:"title"`
    Children []Productcategory `gorm:"-" json:"children" form:"children"`
    }
```
#### 指定数据库名
```go
func (c Productcategory) TableName() string {
return "productcategories"
}
```
### 方法使用
#### 批量建树，控制器中
```err := category.CreateTree(&category)```
#### 查询树形
```go
    category := models.Productcategory{}
	id := ctx.Request().Query("id")
	facades.Orm().Query().Where("id = ?", id).First(&category)
	categories, err := category.GetTree(&category)
```
#### 添加子树
```go
if err := category.AppendChild(&category, parentID); err != nil {
	return err
}
```
#### 删除树，将自动删除rmCategory和其下所有的子树
```go
category.RemoveNode(&rmCategory)
```
### 特别说明，扩展默认需要模型中存在name和title字段，如想添加其他字段，请另行添加
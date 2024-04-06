package nestedset

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"gorm.io/gorm"
	"reflect"
)

type Nestedset struct {
	Lft int `gorm:"column:lft;type:int(11);not null;default:0" json:"lft" form:"lft"`
	Rgt int `gorm:"column:rgt;type:int(11);not null;default:0" json:"rgt" form:"rgt"`
	Pid int `gorm:"column:pid;type:int(11);null;default:0" json:"pid" form:"pid"`
}

func (n *Nestedset) NewInstance() *Nestedset {
	return &Nestedset{}
}

func (n *Nestedset) ToTree(model interface{}) (interface{}, error) {
	//model必须为一个指针
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		panic("model must be a pointer")
	}
	result, err := n.GetTree(model)
	if err != nil {
		return nil, err
	}
	return result, nil
}

//接收参数为一个指针

func (n *Nestedset) CreateTree(modelPrt interface{}) error {

	//判断modelPrt是否为指针,如果不是指针报错
	if reflect.TypeOf(modelPrt).Kind() != reflect.Ptr {
		panic("modelPrt must be a pointer")
	}
	//获取modelPrt的类型
	model := reflect.TypeOf(modelPrt).Elem()

	v := reflect.ValueOf(modelPrt).Elem()

	//获取modelPrt结构体中所有的字段
	fieldNum := v.NumField()
	fieldNames := make([]string, 0)
	//遍历字段
	for i := 0; i < fieldNum; i++ {
		fieldName := v.Field(i).Type().Name()
		if fieldName == "Model" ||
			fieldName == "Nestedset" ||
			fieldName == "" {
			continue
		}
		fieldNames = append(fieldNames, v.Type().Field(i).Name)
	}

	Pid := v.FieldByName("Pid").Int()

	Lft := v.FieldByName("Lft").Int()

	if Lft == 0 {
		Lft = 1
	}

	moveIndex := Lft
	tableName := v.MethodByName("TableName").Call(nil)

	tableNameStr := tableName[0].String()
	if tableNameStr == "" {
		return errors.New("请申明TableName方法")
	}
	topModel := reflect.New(model).Interface()
	//循环设置
	for _, fieldName := range fieldNames {
		//获取字段的值
		fieldValue := v.FieldByName(fieldName).Interface()
		//设置字段的值
		reflect.ValueOf(topModel).Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue))
	}

	if Pid != 0 {
		reflect.ValueOf(topModel).Elem().FieldByName("Pid").SetInt(Pid)
	}
	reflect.ValueOf(topModel).Elem().FieldByName("Lft").SetInt(Lft)
	//判断有没children
	children := v.FieldByName("Children").Interface()
	//将children转换为model类型的切片类型，并为topModel设置Children属性，值为children
	childrenSlice := reflect.ValueOf(children)
	reflect.ValueOf(topModel).Elem().FieldByName("Children").Set(childrenSlice)
	err := facades.Orm().Query().Table(tableNameStr).Create(topModel)
	if err != nil {
		return err
	}
	n.createTree(tableNameStr, topModel, moveIndex)
	return nil
}
func (n *Nestedset) createTree(table string, model interface{}, moveIndex int64) int64 {
	curIndex := moveIndex
	//moveIndex始终为节点的左值
	if reflect.TypeOf(model).Kind() != reflect.Ptr {
		panic("model must be a pointer")
	}
	//v := reflect.ValueOf(model).Elem()

	//获取modelPrt结构体中所有的字段
	//fieldNum := v.NumField()

	children := reflect.ValueOf(model).Elem().FieldByName("Children").Interface()
	parent_id := reflect.ValueOf(model).Elem().FieldByName("ID").Uint()
	childrenSlice := reflect.ValueOf(children)
	childrenLen := childrenSlice.Len()

	if childrenLen > 0 {
		//循环
		for i := 0; i < childrenLen; i++ {
			curIndex++
			child := childrenSlice.Index(i).Interface()
			newChild := reflect.New(reflect.TypeOf(child)).Interface()

			//region====================固定的字段====================

			//动态的判断，先获取child中的
			//遍历字段
			name := reflect.ValueOf(children).Index(i).FieldByName("Name").String()
			title := reflect.ValueOf(children).Index(i).FieldByName("Title").String()
			//设置pid
			reflect.ValueOf(newChild).Elem().FieldByName("Name").SetString(name)
			reflect.ValueOf(newChild).Elem().FieldByName("Title").SetString(title)

			reflect.ValueOf(newChild).Elem().FieldByName("Pid").SetInt(int64(parent_id))
			reflect.ValueOf(newChild).Elem().FieldByName("Lft").SetInt(curIndex)

			//是否存在其他的字段，除了name,title
			//判断
			//获取所有字段的名字，放入fieldNames中
			//endregion====================固定的字段====================

			//region====================其他字段====================

			fieldNames := make([]string, 0)
			//遍历字段
			fieldChild := reflect.ValueOf(children).Index(i)
			//反射获取字段
			for j := 0; j < fieldChild.NumField(); j++ {
				//判断类型,如果是struct类型，跳过
				if fieldChild.Type().Field(j).Type.Kind() == reflect.Struct || fieldChild.Type().Field(j).Type.Kind() == reflect.Slice {
					continue
				}
				fieldNames = append(fieldNames, fieldChild.Type().Field(j).Name)
			}

			//遍历fieldNames
			for _, fieldName := range fieldNames {
				//如果字段名为name或者title，跳过
				if fieldName == "Name" || fieldName == "Title" || fieldName == "Pid" || fieldName == "Lft" || fieldName == "Rgt" {
					continue
				}
				//获取字段的类型
				fieldType := reflect.ValueOf(children).Index(i).FieldByName(fieldName).Type()
				//获取字段的值
				fieldValue := reflect.ValueOf(children).Index(i).FieldByName(fieldName).Interface()
				//根据类型的不同设置不同类型的值
				switch fieldType.Name() {
				case "string":
					reflect.ValueOf(newChild).Elem().FieldByName(fieldName).SetString(fieldValue.(string))
				case "int":
					reflect.ValueOf(newChild).Elem().FieldByName(fieldName).SetInt(int64(fieldValue.(int)))
				case "int64":
					reflect.ValueOf(newChild).Elem().FieldByName(fieldName).SetInt(fieldValue.(int64))
				case "float64":
					reflect.ValueOf(newChild).Elem().FieldByName(fieldName).SetFloat(fieldValue.(float64))
				case "bool":
					reflect.ValueOf(newChild).Elem().FieldByName(fieldName).SetBool(fieldValue.(bool))
				case "time.Time":
					reflect.ValueOf(newChild).Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue.(carbon.DateTime)))
				default:
					reflect.ValueOf(newChild).Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue))
				}
			}

			//endregion====================其他字段====================

			//newChild中还有没有Children
			childChildren := reflect.ValueOf(children).Index(i).FieldByName("Children").Interface()
			//如果childChildren不为空
			if childChildren != nil {
				reflect.ValueOf(newChild).Elem().FieldByName("Children").Set(reflect.ValueOf(childChildren))
			}

			//创建
			facades.Orm().Query().Table(table).Create(newChild)
			curIndex = n.createTree(table, newChild, curIndex)
		}
		//更新右值
		tid := n.getId(model)
		curIndex++
		facades.Orm().Query().Table(table).Where("id=?", tid).Update("rgt", curIndex)
	} else {
		//更新右值
		lft := reflect.ValueOf(model).Elem().FieldByName("Lft").Int()
		tid := n.getId(model)
		facades.Orm().Query().Table(table).Where("id=?", tid).Update("rgt", lft+1)
		curIndex = lft + 1
	}
	return curIndex
}

func (n *Nestedset) GetTree(modelPrt interface{}) (interface{}, error) {
	//根据modelPrt的id作为子节点的pid，以树形结构展示，递归
	resultMap := make(map[string]interface{})
	//判断modelPrt是否为指针,如果不是指针报错
	if reflect.TypeOf(modelPrt).Kind() != reflect.Ptr {
		panic("modelPrt must be a pointer")
	}
	//获取modelPrt的值
	v := reflect.ValueOf(modelPrt).Elem()

	fieldNames := make([]string, 0)

	//反射获取字段
	for j := 0; j < v.NumField(); j++ {
		//判断类型,如果是struct类型，跳过
		if v.Type().Field(j).Type.Kind() == reflect.Struct || v.Type().Field(j).Type.Kind() == reflect.Slice {
			continue
		}
		fieldNames = append(fieldNames, v.Type().Field(j).Name)
	}

	//固定字段将结果放入map中
	resultMap["id"] = v.FieldByName("ID").Uint()
	resultMap["name"] = v.FieldByName("Name").String()
	resultMap["title"] = v.FieldByName("Title").String()
	resultMap["pid"] = v.FieldByName("Pid").Int()
	resultMap["lft"] = v.FieldByName("Lft").Int()
	resultMap["rgt"] = v.FieldByName("Rgt").Int()
	//先断言为carbon.Carbon类型，再转换为string类型
	created_at := v.FieldByName("CreatedAt").Interface().(carbon.DateTime).String()
	updated_at := v.FieldByName("UpdatedAt").Interface().(carbon.DateTime).String()
	resultMap["created_at"] = created_at
	resultMap["updated_at"] = updated_at
	resultMap["children"] = []interface{}{}
	//其他字段，循环
	for i, fieldName := range fieldNames {
		//如果字段名为name或者title，跳过
		if fieldName == "Name" || fieldName == "Title" || fieldName == "Pid" || fieldName == "Lft" || fieldName == "Rgt" {
			continue
		}
		//获取字段的类型
		fieldType := v.FieldByName(fieldName).Type()
		//获取字段的值
		fieldValue := v.FieldByName(fieldName).Interface()
		//获取字段的fieldName对应的tag中的json值,v.type().Field(x)中的x是从1开始，而for range中的Key是从0开始,所以i+1
		jsonName := v.Type().Field(i + 1).Tag.Get("json")
		formName := v.Type().Field(i + 1).Tag.Get("form")
		if jsonName != "" {
			fieldName = jsonName
		} else if formName != "" {
			fieldName = formName
		}

		//根据类型的不同设置不同类型的值
		switch fieldType.Name() {
		case "string":
			resultMap[fieldName] = fieldValue.(string)
		case "int":
			resultMap[fieldName] = int(fieldValue.(int))
		case "int64":
			resultMap[fieldName] = fieldValue.(int64)
		case "float64":
			resultMap[fieldName] = fieldValue.(float64)
		case "bool":
			resultMap[fieldName] = fieldValue.(bool)
		case "time.Time":
			resultMap[fieldName] = fieldValue.(carbon.DateTime).String()
		default:
			resultMap[fieldName] = fieldValue
		}
	}
	var childrenLen int64
	tableName := v.MethodByName("TableName").Call(nil)
	//子节点长度放入childrenLen中
	facades.Orm().Query().Table(tableName[0].String()).Where("pid=?", resultMap["id"]).Count(&childrenLen)
	if childrenLen > 0 {
		//获取子节点
		children := reflect.New(reflect.SliceOf(reflect.TypeOf(modelPrt).Elem())).Interface()
		facades.Orm().Query().Table(tableName[0].String()).Where("pid=?", resultMap["id"]).Find(children)
		for i := 0; i < int(childrenLen); i++ {
			//将children的第i个元素传入GetTree方法中，获取其子节点
			childrenMap, err := n.GetTree(reflect.ValueOf(children).Elem().Index(i).Addr().Interface())
			//childrenMap转换为map类型
			childrenMap = childrenMap.(map[string]interface{})
			if err != nil {
				return nil, err
			}
			//将子节点加入到resultMap中
			resultMap["children"] = append(resultMap["children"].([]interface{}), childrenMap)
		}
	}
	return resultMap, nil
}

func (n *Nestedset) PrettyPrint(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		fmt.Println(v)
		return
	}

	var out bytes.Buffer
	err = json.Indent(&out, b, "", "  ")
	if err != nil {
		fmt.Println(v)
		return
	}

	fmt.Println(out.String())
}
func (n *Nestedset) getId(newModel interface{}) int {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.Encode(newModel)
	var data map[string]interface{}
	json.Unmarshal(buf.Bytes(), &data)
	id := data["ID"]
	//将id转换为int类型
	idInt := int(id.(float64))
	return idInt
}

func (n *Nestedset) AppendChild(modelPrt interface{}, id int) error {

	//判断modelPrt是否为指针,如果不是指针报错
	if reflect.TypeOf(modelPrt).Kind() != reflect.Ptr {
		panic("modelPrt must be a pointer")
	}
	//获取modelPrt的类型
	model := reflect.TypeOf(modelPrt).Elem()

	v := reflect.ValueOf(modelPrt).Elem()
	tableName := v.MethodByName("TableName").Call(nil)

	name := v.FieldByName("Name").String()
	title := v.FieldByName("Title").String()
	//region设置其他值====================

	fieldNames := make([]string, 0)
	//遍历字段(通过modelPrt）
	//反射获取字段
	for j := 0; j < v.NumField(); j++ {
		//判断类型,如果是struct类型或者Slice，跳过
		if v.Type().Field(j).Type.Kind() == reflect.Struct || v.Type().Field(j).Type.Kind() == reflect.Slice {
			continue
		}

		fieldNames = append(fieldNames, v.Type().Field(j).Name)
	}

	//endregion====================其他字段====================
	tableNameStr := tableName[0].String()

	//获取到pid为id的最后一个，按id降序排列
	//创建一个modelPrt类型的变量lastModel
	lastModel := reflect.New(model).Interface()
	facades.Orm().Query().Table(tableNameStr).Where("pid=?", id).Order("id desc").First(&lastModel)
	//创建新节点
	newModel := reflect.New(model).Interface()
	reflect.ValueOf(newModel).Elem().FieldByName("Name").SetString(name)
	reflect.ValueOf(newModel).Elem().FieldByName("Title").SetString(title)
	//设置pid
	reflect.ValueOf(newModel).Elem().FieldByName("Pid").SetInt(int64(id))
	//设置左值
	reflect.ValueOf(newModel).Elem().FieldByName("Lft").SetInt(reflect.ValueOf(lastModel).Elem().FieldByName("Rgt").Int() + 1)
	//设置右值
	reflect.ValueOf(newModel).Elem().FieldByName("Rgt").SetInt(reflect.ValueOf(lastModel).Elem().FieldByName("Rgt").Int() + 2)

	//遍历fieldNames
	for _, fieldName := range fieldNames {
		//如果字段名为name或者title，跳过
		if fieldName == "Name" || fieldName == "Title" || fieldName == "Pid" || fieldName == "Lft" || fieldName == "Rgt" {
			continue
		}
		//获取model字段的类型
		fieldType := reflect.ValueOf(modelPrt).Elem().FieldByName(fieldName).Type()
		//获取字段的值
		fieldValue := reflect.ValueOf(modelPrt).Elem().FieldByName(fieldName).Interface()
		//根据类型的不同设置不同类型的值
		switch fieldType.Name() {
		case "string":
			reflect.ValueOf(newModel).Elem().FieldByName(fieldName).SetString(fieldValue.(string))
		case "int":
			reflect.ValueOf(newModel).Elem().FieldByName(fieldName).SetInt(int64(fieldValue.(int)))
		case "int64":
			reflect.ValueOf(newModel).Elem().FieldByName(fieldName).SetInt(fieldValue.(int64))
		case "float64":
			reflect.ValueOf(newModel).Elem().FieldByName(fieldName).SetFloat(fieldValue.(float64))
		case "bool":
			reflect.ValueOf(newModel).Elem().FieldByName(fieldName).SetBool(fieldValue.(bool))
		case "time.Time":
			reflect.ValueOf(newModel).Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue.(carbon.DateTime)))
		default:
			reflect.ValueOf(newModel).Elem().FieldByName(fieldName).Set(reflect.ValueOf(fieldValue))
		}
	}
	tx, _ := facades.Orm().Query().Begin()
	if err := tx.Table(tableNameStr).Create(newModel); err != nil {
		//	批量更新其他所有的节点，节点的右值都大于lastmodel的右值，这些节点左值和右值分别加上lastmodel的右值+2
		if _, err2 := tx.Table(tableNameStr).Where("rgt>?", reflect.ValueOf(lastModel).Elem().FieldByName("Rgt").Int()).
			Update(map[string]interface{}{"lft": gorm.Expr("lft+2"), "rgt": gorm.Expr("rgt+2")}); err2 != nil {
			return err2
		}
		err1 := tx.Rollback()
		return err1
	} else {
		err1 := tx.Commit()
		return err1
	}
	//创建

	return nil
}

func (n *Nestedset) RemoveNode(modelPrt interface{}) error {
	v := reflect.ValueOf(modelPrt).Elem()

	tableName := v.MethodByName("TableName").Call(nil)
	tableNameStr := tableName[0].String()
	//获取modelPrt的左值
	lft := v.FieldByName("Lft").Int()
	//获取modelPrt的右值
	rgt := v.FieldByName("Rgt").Int()
	//获取modelPrt的id
	id := v.FieldByName("ID").Uint()
	//1、删除左右值在modelPrt的左右值之间的节点
	//2、删除modelPrt节点
	//3、更新左右值大于modelPrt的右值的节点，左值减去modelPrt的右值-左值+1，右值减去modelPrt的右值-左值+1
	tx, _ := facades.Orm().Query().Begin()
	ids := []int{}
	//查找ids
	tx.Table(tableNameStr).Where("lft>? and rgt<?", lft, rgt).Pluck("id", &ids)
	//执行删除
	_, err := tx.Table(tableNameStr).Delete(&modelPrt, ids)
	if err != nil {
		return err
	}
	//UPDATE nested_category SET rgt = rgt - @myWidth WHERE rgt > @myRight;
	//UPDATE nested_category SET lft = lft - @myWidth WHERE lft > @myRight;

	if _, err1 := tx.Table(tableNameStr).Where("rgt>?", rgt).Update("rgt", gorm.Expr("rgt-?", rgt-lft+1)); err1 != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
	}

	if _, err2 := tx.Table(tableNameStr).Where("lft>?", rgt).Update("lft", gorm.Expr("lft-?", rgt-lft+1)); err2 != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
	}

	//删除自身
	if _, err3 := tx.Table(tableNameStr).Delete(&modelPrt, id); err3 != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
	}
	err4 := tx.Commit()
	return err4
}

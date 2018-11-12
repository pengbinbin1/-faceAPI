package server

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Intelligentvision/faceAPI/conn"
	"github.com/jinzhu/gorm"
)

type FaceAPIServer struct{}

var TIMEOUT = 10

type Tables struct {
	FaceLibrary *FaceLibrary
}

func init() {
	CheckTables()
}

//检查表
func CheckTables() {
	gormDb, err := conn.GetGorm()
	if err != nil {
		fmt.Println("get gormDb error.", err)
		panic(err)
	}

	var o interface{}
	o = Tables{}
	t := reflect.TypeOf(o)         //反射使用 TypeOf 和 ValueOf 函数从接口中获取目标对象信息
	fmt.Println("Type:", t.Name()) //调用t.Name方法来获取这个类型的名称

	v := reflect.ValueOf(o) //打印出所包含的字段
	fmt.Println("Fields:")
	for i := 0; i < t.NumField(); i++ { //通过索引来取得它的所有字段，这里通过t.NumField来获取它多拥有的字段数量，同时来决定循环的次数
		f := t.Field(i)               //通过这个i作为它的索引，从0开始来取得它的字段
		val := v.Field(i).Interface() //通过interface方法来取出这个字段所对应的值
		checkTable(gormDb, val, f.Name)
		//fmt.Printf("%6s:%v =%v\n", f.Name, f.Type, val)
	}
	for i := 0; i < t.NumMethod(); i++ { //这里同样通过t.NumMethod来获取它拥有的方法的数量，来决定循环的次数
		m := t.Method(i)
		fmt.Printf("%6s:%v\n", m.Name, m.Type)

	}

}

//检查表是否存在，不存在时创建
func checkTable(gormDb *gorm.DB, tb interface{}, name string) {
	name = strings.ToLower(name) + "s"
	//判断数据库中是否有该表，没有则创建该表
	if st := gormDb.HasTable(tb); st == false {
		//数据表不存在，创建数据表
		if err := gormDb.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8").CreateTable(tb).Error; err != nil {
		} else {
			fmt.Println(" not find table " + name + ",but create it success. ")
		}
	} else {
		fmt.Println("table:  " + name + " exist and can be access ")
	}

	//根据clusterID创建索引

	err := gormDb.Model(tb).AddIndex("idx_cluster_id", "cluster_id").Error
	if err != nil {
		fmt.Println("create index failed", err)
	} else {
		fmt.Println("create index success")
	}
}

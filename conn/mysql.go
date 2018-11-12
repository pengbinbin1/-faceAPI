package conn

import (
	"bytes"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var GormDb *gorm.DB

func init() {
	newGorm()
}

/*
* DESC: 返回当前系统的唯一mysql连接句柄
* PRE: 不需要传入任何参数
* POST: 新建了一个句柄
 */
func newGorm() (*gorm.DB, error) {
	args := []string{MYSQL_USER, ":", MYSQL_PASSWORD, "@", "tcp(", MYSQL_HOST, ")/", MYSQL_DB, "?charset=utf8&parseTime=True&loc=Local"}
	argBuf := bytes.Buffer{}
	for _, arg := range args {
		argBuf.WriteString(arg)
	}
	argsStr := argBuf.String()
	if GormDb == nil {
		lock.Lock()
		defer lock.Unlock()
		if GormDb == nil {
			GormDb, err = gorm.Open("mysql", argsStr)
			if err != nil {
				GormDb = nil
				log.Println("mysql数据库连接失败:", err)
			}
			fmt.Println("gormdb get success")
		}
	}

	return GormDb, err

}

/*
* DESC: 返回当前系统的唯一mysql连接句柄
* PRE: 不需要传入任何参数
* POST: 当句柄不为nil时没有做任何修改，当为nil时新建了一个句柄
 */
func GetGorm() (*gorm.DB, error) {
	return newGorm()
}

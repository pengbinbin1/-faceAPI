package conn

import (
	"sync"

	"github.com/Intelligentvision/faceAPI/config"
)

var (
	//xgface xgindex 相关参数
	XGFACE_IPP  = config.Config.Services.Web.Xgfaceaddr
	XGINDEX_IPP = config.Config.Services.Web.Xgindexaddr
	TIMEOUT     = config.Config.Services.Web.Deadline

	//mysql相关
	MYSQL_HOST     = config.Config.Services.Mysql.Host
	MYSQL_USER     = config.Config.Services.Mysql.User
	MYSQL_PASSWORD = config.Config.Services.Mysql.Password
	MYSQL_DB       = config.Config.Services.Mysql.Database

	//图像存储路径
	IMG_PATH = config.Config.Services.Image.Imgbasepath
)

var lock *sync.Mutex = &sync.Mutex{}

var FORMAT = ".jpg"

var err error

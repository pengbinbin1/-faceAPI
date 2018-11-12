package common

const (

	//xgface相关
	ERR_OK             = 0      //正确
	ERR_INVALID_PARAM  = -10001 //参数错误
	ERR_CONNECT_FAILED = -10002 //连接错误
	ERR_ALG_FAILED     = -10003 //算法返回出错
	ERR_DETECT_FAILED  = -10004 //人脸检测失败
	ERR_DETECT_NO_FACE = -10005 //没有检测到人脸
	ERR_MULTI_FACE     = -10006 //检测到多张人脸

	//人脸库(mysql)相关
	ERR_CONN_MYSQL_FAILED  = -11001 //连接数据库失败
	ERR_EXIST              = -11002 //该人脸已存在
	ERR_QUERY_FAILED       = -11003 //检索出错
	ERR_GET_PATH_FAILED    = -11004 //获取路径时出错
	ERR_CREATE_PATH_FAILED = -11005 //创建路径时出错
	ERR_CREATE_FILE_FAILED = -11006 //创建文件失败
	ERR_INSERT_FAILED      = -11007 //插入记录失败
	ERR_ENCODE_FAILED      = -11008 //base64解码失败
	ERR_DEL_FILE_FAILED    = -11009 //删除文件失败
	ERR_DEL_RECORD_FAILED  = -11010 //删除记录失败

	//xgindex相关
	ERR_CONN_INDEX_FAILED   = -12001 //连接xgindex失败
	ERR_SIMIARITY_FAILED    = -12002 //获取相似度失败
	ERR_INSERT_INDEX_FAILED = -12003 //插入index获取clusterID
	ERR_DEL_CLUSTER_FAILED  = -12004 //删除cluster失败
)

var ErrMsg = map[int32]string{

	//xgface相关
	ERR_OK:             "正确",
	ERR_INVALID_PARAM:  "非法参数",
	ERR_CONNECT_FAILED: "连接xgface失败",
	ERR_ALG_FAILED:     "算法返回错误",
	ERR_DETECT_FAILED:  "人脸检测失败",
	ERR_DETECT_NO_FACE: "没有检测到人脸",
	ERR_MULTI_FACE:     "检测到多张人脸",

	//人脸库相关
	ERR_CONN_MYSQL_FAILED:  "连接数据库失败",
	ERR_EXIST:              "人脸已存在",
	ERR_QUERY_FAILED:       "检索失败",
	ERR_GET_PATH_FAILED:    "获取路径失败",
	ERR_CREATE_PATH_FAILED: "创建路径失败",
	ERR_CREATE_FILE_FAILED: "创建文件失败",
	ERR_INSERT_FAILED:      "插入失败",
	ERR_ENCODE_FAILED:      "base64解码失败",
	ERR_DEL_FILE_FAILED:    "删除文件失败",
	ERR_DEL_RECORD_FAILED:  "删除记录失败",

	//xgindex相关
	ERR_CONN_INDEX_FAILED:   "连接xgindex失败",
	ERR_SIMIARITY_FAILED:    "获取相似度失败",
	ERR_INSERT_INDEX_FAILED: "插入index失败",
	ERR_DEL_CLUSTER_FAILED:  "删除cluster失败",
}

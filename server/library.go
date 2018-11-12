package server

import (
	"bufio"
	"context"
	"errors"
	//	"fmt"
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"time"

	common "github.com/Intelligentvision/faceAPI/common"
	"github.com/Intelligentvision/faceAPI/conn"
	apiService "github.com/Intelligentvision/faceAPI/proto/faceAPI"
	//pbface "github.com/Intelligentvision/faceAPI/proto/xgface"
	"github.com/jinzhu/gorm"
)

type FaceLibrary struct {
	ID        int64  `gorm:"AUTO_INCREMENT"`
	ClusterId int64  `gorm:"default:0"`
	ImgToken  string `gorm:"default:''"` //图像数据唯一标记
	FaceToken string `gorm:"default:''"` //人脸数据唯一标记
	//Path       string  `gorm:"default:''"` //人脸图像存储路径
	BoxScore   float32 `gorm:"default:0"`  //人脸位置置信度
	BoxWidth   float32 `gorm:"default:0"`  //位置框宽度
	BoxHeight  float32 `gorm:"default:0"`  //位置框高度
	BoxCenterx float32 `gorm:"default:0"`  //位置框中心点x坐标
	BoxCentery float32 `gorm:"default:0"`  //位置框中心点y坐标
	UserID     string  `gorm:"default:''"` //用户ID
	UserInfo   string  `gorm:"default:''"` //用户信息
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

//在人脸库中添加一张人脸：首先根据图像计算ImgToken，到库中搜索，若存在，说明已经入库。如查找不存在，调用xgface做人脸检测，将位置信息等存入
//库中并返回位置
func (s *FaceAPIServer) AddFace(ctx context.Context, in *apiService.AddFaceReq) (*apiService.AddFaceResp, error) {

	var resp apiService.AddFaceResp
	//检测参数是否合法
	if len(in.Data) == 0 {
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.Location = nil
		return &resp, nil
	}

	//计算ImgToken,判断库中是否存在
	imgToken := common.GetImgToken(in.Data)
	faceToken := common.GetImgToken(in.Data + strconv.Itoa(0))
	//创建数据库连接
	gormdb, err := conn.GetGorm()
	if err != nil {
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		resp.Location = nil
		resp.FaceToken = ""
		return &resp, nil
	}

	//查找数据库是否存在
	var faceInfo FaceLibrary
	err = gormdb.Debug().Where("img_token = ? ", imgToken).First(&faceInfo).Error
	if err == nil {
		//说明库中已经存在，无需再次存入
		resp.ErrCode = common.ERR_EXIST
		resp.ErrMsg = common.ErrMsg[common.ERR_EXIST]

		//将查询到的信息赋值返回
		var loca apiService.Box
		var rect apiService.Rect
		var point apiService.Point
		point.CoordinateX = faceInfo.BoxCenterx
		point.CoordinateY = faceInfo.BoxCentery
		rect.Width = faceInfo.BoxWidth
		rect.Height = faceInfo.BoxHeight
		loca.Center = &point
		loca.Rect = &rect
		resp.Location = &loca
		resp.FaceToken = faceInfo.FaceToken
		resp.ImgToken = faceInfo.ImgToken
		return &resp, nil
	} else {
		//查询无记录直接跳过，进行入库操作
		if err == gorm.ErrRecordNotFound {
			//do noting
			log.Println("not found in library,will insert...")
		} else {
			//查询出错
			resp.ErrCode = common.ERR_QUERY_FAILED
			resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
			resp.Location = nil
			resp.FaceToken = ""
			resp.ImgToken = ""
			return &resp, err
		}
	}

	//提取人脸信息
	var apiReq apiService.DetectReq
	var faceField apiService.FaceField
	faceField.Attributes = 1
	faceField.Quality = 1
	faceField.Feature = 1
	faceField.Landmark = 1
	apiReq.Facefield = &faceField
	apiReq.Images = append(apiReq.Images, in.Data)

	detectinfo, errcode := getDetectInfo(&apiReq)
	if errcode != 0 {
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		resp.Location = nil
		resp.FaceToken = ""
		resp.ImgToken = ""
		return &resp, nil
	}
	if len(detectinfo.Faces) <= 0 {
		log.Println("detect no face")
		resp.ErrCode = common.ERR_DETECT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DETECT_FAILED]
		resp.Location = nil
		resp.FaceToken = ""
		return &resp, nil
	}

	//单张图像多张人脸不允许入库
	if len(detectinfo.Faces) > 1 {
		log.Println("detect multi faces:", len(detectinfo.Faces))
		resp.ErrCode = common.ERR_MULTI_FACE
		resp.ErrMsg = common.ErrMsg[common.ERR_MULTI_FACE]
		return &resp, nil
	}

	//获取box
	box := generateBox(detectinfo.Faces[0])

	//从index获得clusterID
	feature := generateFeature(detectinfo.Faces[0])
	cid, errcode := insertIntoIndex(feature.Values)
	if errcode != 0 {
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		resp.Location = nil
		resp.FaceToken = ""
		return &resp, nil
	}

	/*
		//存储图像至本地,取得文件名
		errcode, filename := SaveImg(in.Data)
		if errcode != 0 {
			resp.ErrCode = errcode
			resp.ErrMsg = common.ErrMsg[errcode]
			resp.Location = nil
			resp.FaceToken = ""
			return &resp, nil
		}*/

	//赋值后存入数据库
	faceL := getStruct(box, cid)
	faceL.ImgToken = imgToken
	faceL.FaceToken = faceToken
	faceL.UserID = in.UserID
	faceL.UserInfo = in.UserInfo
	err = gormdb.Debug().Create(&faceL).Error
	if err != nil {

		resp.ErrCode = common.ERR_INSERT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_INSERT_FAILED]
		resp.Location = nil
		resp.FaceToken = ""
		resp.ImgToken = ""
		return &resp, nil
	}

	resp.Location = &box
	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]
	resp.FaceToken = faceToken
	resp.ImgToken = imgToken

	return &resp, nil
}

//查找人脸库中最相似的一张人脸
func (s *FaceAPIServer) Scan(ctx context.Context, in *apiService.Image) (*apiService.ScanResp, error) {

	//先做人脸检测，提取人脸特征值，调用index获取比分

	//验证参数是否合法
	var resp apiService.ScanResp
	if len(in.Data) == 0 {
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.Score = -1
		resp.ImgToken = ""
		resp.FaceToken = ""
		return &resp, errors.New("invalid params")
	}

	//人脸检测
	var apiReq apiService.DetectReq
	var faceField apiService.FaceField
	faceField.Attributes = 0
	faceField.Quality = 0
	faceField.Feature = 1
	faceField.Landmark = 1
	apiReq.Facefield = &faceField
	apiReq.Images = append(apiReq.Images, in.Data)

	//获取facetoken
	faceToken := common.GetImgToken(in.Data + strconv.Itoa(0))

	detectinfo, errcode := getDetectInfo(&apiReq)
	if errcode != 0 {
		log.Println("get detectinfo failed")
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		resp.Score = -1
		resp.ImgToken = ""
		resp.FaceToken = ""
		return &resp, nil
	}

	//没检测到人脸
	if len(detectinfo.Faces) <= 0 {
		log.Println("detect face failed,facenum = ", len(detectinfo.Faces))
		resp.ErrCode = common.ERR_DETECT_NO_FACE
		resp.ErrMsg = common.ErrMsg[common.ERR_DETECT_NO_FACE]
		resp.Score = -1
		resp.ImgToken = ""
		resp.FaceToken = ""
		return &resp, nil
	}
	//检测到多张人脸
	if len(detectinfo.Faces) > 1 {
		log.Println("detect multi face  ,facenum = ", len(detectinfo.Faces))
		resp.ErrCode = common.ERR_MULTI_FACE
		resp.ErrMsg = common.ErrMsg[common.ERR_MULTI_FACE]
		resp.Score = -1
	}

	//提取特征
	feature := generateFeature(detectinfo.Faces[0])
	//从index处获得最相似人脸clusterID
	score, cID, errcode := getMostSimilarity(feature.Values)
	if errcode != 0 {
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		resp.Score = -1
		resp.ImgToken = ""
		resp.FaceToken = ""
		return &resp, nil
	}

	//根据clusterID去表中查询图片地址
	gormdb, err := conn.GetGorm()
	if err != nil {
		log.Println("connect to mysql failed")
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		resp.FaceToken = ""
		resp.ImgToken = ""
		resp.Score = -1
		return &resp, nil
	}

	var face FaceLibrary
	err = gormdb.Debug().Where("cluster_id = ?", cID).First(&face).Error
	if err != nil {
		log.Println("query failed, clusterID = ", cID)
		resp.ErrCode = common.ERR_QUERY_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
		resp.Score = -1
		resp.ImgToken = ""
		resp.FaceToken = ""
		return &resp, nil
	}

	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]
	resp.ImgToken = face.ImgToken
	resp.Score = score
	resp.FaceToken = faceToken
	resp.UserID = face.UserID
	resp.UserInfo = face.UserInfo
	return &resp, nil
}

//给定一张图中有多张人脸，输出每一张人脸最相似的图像
func (s *FaceAPIServer) ScanEx(ctx context.Context, in *apiService.Image) (*apiService.ScanRespEx, error) {

	//先做人脸检测，提取人脸特征值，调用index获取比分

	//验证参数是否合法
	var resp apiService.ScanRespEx

	if len(in.Data) == 0 {
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.ScanResult = nil
		return &resp, errors.New("invalid params")
	}

	//人脸检测
	var apiReq apiService.DetectReq
	var faceField apiService.FaceField
	faceField.Attributes = 0
	faceField.Quality = 0
	faceField.Feature = 1
	faceField.Landmark = 1
	apiReq.Facefield = &faceField
	apiReq.Images = append(apiReq.Images, in.Data)

	//获取人脸信息
	detectinfo, errcode := getDetectInfo(&apiReq)
	if errcode != 0 {
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		resp.ScanResult = nil
		return &resp, nil
	}

	//没有检测到人脸或检测失败
	if len(detectinfo.Faces) <= 0 {
		resp.ErrCode = common.ERR_DETECT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DETECT_FAILED]
		resp.ScanResult = nil
		return &resp, nil
	}

	var tempResult apiService.Result

	for i := 0; i < len(detectinfo.Faces); i++ {
		//首先提取人脸特征
		feature := generateFeature(detectinfo.Faces[i])

		//通过人脸特征在index获得clusterID
		score, cID, errcode := getMostSimilarity(feature.Values)
		if errcode != 0 {
			resp.ErrCode = errcode
			resp.ErrMsg = common.ErrMsg[errcode]
			resp.ScanResult = nil
			return &resp, nil
		}

		//根据ClusterID查询数据表获得path
		gormdb, err := conn.GetGorm()
		if err != nil {
			log.Println("connect to mysql failed")
			resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
			resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
			resp.ScanResult = nil
			return &resp, nil
		}

		var face FaceLibrary
		err = gormdb.Debug().Where("cluster_id = ?", cID).First(&face).Error
		if err != nil {
			log.Println("query failed, clusterID = ", cID)
			resp.ErrCode = common.ERR_QUERY_FAILED
			resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
			resp.ScanResult = nil
			return &resp, nil
		}

		//赋值图像路径和分数
		var tempScanFace apiService.ScanFace
		var box apiService.Box
		var point apiService.Point
		var rect apiService.Rect
		point.CoordinateX = face.BoxCenterx
		point.CoordinateY = face.BoxCentery
		rect.Height = face.BoxHeight
		rect.Width = face.BoxWidth
		box.Center = &point
		box.Rect = &rect
		box.Score = face.BoxScore

		tempScanFace.Location = &box
		tempScanFace.UserID = face.UserID
		tempScanFace.UserInfo = face.UserInfo
		tempScanFace.Score = score
		tempScanFace.ImgToken = face.ImgToken
		//获取faceToken并赋值
		faceToken := common.GetImgToken(in.Data + strconv.Itoa(i))
		tempScanFace.FaceToken = faceToken
		tempResult.FaceList = append(tempResult.FaceList, &tempScanFace)
	}

	//将各项赋值给resp并返回
	tempResult.FaceNum = (int32)(len(detectinfo.Faces))
	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]
	resp.ScanResult = &tempResult

	return &resp, nil
}

//获取人脸库中的人脸列表
func (s *FaceAPIServer) ListFace(ctx context.Context, in *apiService.ListFaceReq) (*apiService.ListFaceResp, error) {

	/*
		//判断参数合法性,起始时间小于0的时候置为0，结束时间小于0时报错
		var startTime, endTime interface{}
		if in.StartTime <= 0 {
			log.Println("start time is invalied:", startTime, ",will set 0 to it")
			startTime = time.Unix(0, 0)
		} else {
			startTime = time.Unix(in.StartTime+28800, 0) //加时区
		}

		var resp apiService.ListFaceResp
		if in.EndTime < 0 {
			resp.ErrCode = common.ERR_INVALID_PARAM
			resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
			resp.FaceRecord = nil
			return &resp, nil
		} else {
			endTime = time.Unix(in.EndTime+28800, 0) //加时区
		}

		//判断pageSize与currentPage是否合法
		var pageSize, currentPage int32
		if in.CurrentPage < 1 {
			log.Println("current page is invalid:", in.CurrentPage)
			currentPage = 1
		} else {
			currentPage = in.CurrentPage
		}

		if in.PageSize <= 0 {
			log.Println("pageSize is invalid:", in.PageSize)
			pageSize = 10 //默认为10
		} else {
			pageSize = in.PageSize
		}*/

	var resp apiService.ListFaceResp
	//检查参数是否合法
	if len(in.UserID) <= 0 || in.UserID == "" {
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.FaceRecord = nil
	}
	//到数据库查找
	gormdb, err := conn.GetGorm()
	if err != nil {
		log.Println("connect to mysql failed")
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		resp.FaceRecord = nil
		return &resp, nil
	}

	var face []FaceLibrary
	err = gormdb.Debug().Where("user_id = ?", in.UserID).Find(&face).Error
	if err != nil {
		resp.ErrCode = common.ERR_QUERY_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
		resp.FaceRecord = nil
		return &resp, nil
	}

	//拼接结果并返回
	for i := 0; i < len(face); i++ {
		var tempRecord apiService.SingleRecord
		tempRecord.CreateTime = face[i].CreatedAt.Unix()
		tempRecord.UpdateTime = face[i].UpdatedAt.Unix()
		tempRecord.FaceToken = face[i].FaceToken
		tempRecord.ImgToken = face[i].ImgToken
		resp.FaceRecord = append(resp.FaceRecord, &tempRecord)
	}

	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]

	return &resp, nil
}

//更新人脸信息
func (s *FaceAPIServer) UpdateFace(ctx context.Context, in *apiService.AddFaceReq) (*apiService.AddFaceResp, error) {
	//功能与添加人脸基本相同，只是在添加前先删除原有图像
	var resp apiService.AddFaceResp
	//检测参数是否合法
	if len(in.Data) == 0 {
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.Location = nil
		return &resp, nil
	}

	imgToken := common.GetImgToken(in.Data)
	faceToken := common.GetImgToken(in.Data + strconv.Itoa(0))

	//创建数据库连接
	gormdb, err := conn.GetGorm()
	if err != nil {
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		resp.Location = nil
		resp.FaceToken = ""
		return &resp, nil
	}

	//删除原数据
	var face []FaceLibrary

	//先删除index
	err = gormdb.Debug().Where("user_id = ?", in.UserID).Unscoped().Find(&face).Error
	if err != nil {
		log.Println("query failed")
		resp.ErrCode = common.ERR_QUERY_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
		return &resp, nil
	}
	for i := 0; i < len(face); i++ {
		errcode := DelCluster((int32)(face[i].ClusterId))
		if errcode != 0 {
			log.Println("delet cluster failed:", face[i].ClusterId)
			resp.ErrCode = errcode
			resp.ErrMsg = common.ErrMsg[errcode]
			return &resp, nil
		}
	}

	//删除数据库
	err = gormdb.Debug().Where(" user_id = ?", in.UserID).Unscoped().Delete(&face).Error
	if err != nil {
		log.Println("delete failed")
		gormdb.Rollback()
		resp.ErrCode = common.ERR_DEL_RECORD_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DEL_RECORD_FAILED]
		return &resp, nil
	}

	//提取人脸信息
	var apiReq apiService.DetectReq
	var faceField apiService.FaceField
	faceField.Attributes = 1
	faceField.Quality = 1
	faceField.Feature = 1
	faceField.Landmark = 1
	apiReq.Facefield = &faceField
	apiReq.Images = append(apiReq.Images, in.Data)

	detectinfo, errcode := getDetectInfo(&apiReq)
	if errcode != 0 {
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		return &resp, nil
	}
	if len(detectinfo.Faces) <= 0 {
		log.Println("detect no face")
		resp.ErrCode = common.ERR_DETECT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DETECT_FAILED]

		return &resp, nil
	}

	//单张图像多张人脸不允许入库
	if len(detectinfo.Faces) > 1 {
		log.Println("detect multi faces:", len(detectinfo.Faces))
		resp.ErrCode = common.ERR_MULTI_FACE
		resp.ErrMsg = common.ErrMsg[common.ERR_MULTI_FACE]
		return &resp, nil
	}

	//获取box
	box := generateBox(detectinfo.Faces[0])

	//从index获得clusterID
	feature := generateFeature(detectinfo.Faces[0])
	cid, errcode := insertIntoIndex(feature.Values)
	log.Println("insert into index cid:=", cid)
	if errcode != 0 {
		log.Println("insertinto Index failed")
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		return &resp, nil
	}

	faceL := getStruct(box, cid)
	faceL.ImgToken = imgToken
	faceL.FaceToken = faceToken
	faceL.UserID = in.UserID
	faceL.UserInfo = in.UserInfo
	err = gormdb.Debug().Create(faceL).Error
	if err != nil {

		resp.ErrCode = common.ERR_INSERT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_INSERT_FAILED]
		resp.Location = nil
		resp.FaceToken = ""
		resp.ImgToken = ""
		return &resp, nil
	}

	resp.Location = &box
	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]
	resp.FaceToken = faceToken
	resp.ImgToken = imgToken

	return &resp, nil
}

//获取用户列表
func (s *FaceAPIServer) ListUser(ctx context.Context, in *apiService.Empty) (*apiService.ListUserResp, error) {
	//直接到mysql查询
	var resp apiService.ListUserResp

	var face []FaceLibrary
	gormdb, err := conn.GetGorm()
	if err != nil {
		log.Println("connect to mysql failed")
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		return &resp, nil
	}

	err = gormdb.Debug().Select("distinct user_id").Find(&face).Error
	if err != nil {
		resp.ErrCode = common.ERR_QUERY_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
		return &resp, nil
	}

	for i := 0; i < len(face); i++ {
		resp.UserList = append(resp.UserList, face[i].UserID)
	}

	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]

	return &resp, nil
}

//获取指定用户信息
func (s *FaceAPIServer) GetUserInfo(ctx context.Context, in *apiService.GetUserInfoReq) (*apiService.GetUserInfoResp, error) {
	//检查参数
	var resp apiService.GetUserInfoResp
	if in.UserID == "" || len(in.UserID) == 0 {
		log.Print("param is invalid")
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		return &resp, nil
	}

	var face []FaceLibrary
	gormdb, err := conn.GetGorm()
	if err != nil {
		log.Println("connect to mysql failed")
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		return &resp, nil
	}

	err = gormdb.Debug().Where(" user_id = ?", in.UserID).Unscoped().Find(&face).Error
	if err != nil {
		log.Println("query failed")
		gormdb.Rollback()
		resp.ErrCode = common.ERR_DEL_RECORD_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DEL_RECORD_FAILED]
		return &resp, nil
	}

	for i := 0; i < len(face); i++ {
		resp.UserInfo = append(resp.UserInfo, face[i].UserInfo)
	}

	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]
	return &resp, nil
}

//删除用户
func (s *FaceAPIServer) DelUser(ctx context.Context, in *apiService.DelUserReq) (*apiService.DelUserResp, error) {
	//判断参数是否合法
	var resp apiService.DelUserResp
	if len(in.UserID) <= 0 || in.UserID == "" {
		log.Println("invalid params")
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		return &resp, nil
	}

	var face []FaceLibrary
	gormdb, err := conn.GetGorm()
	if err != nil {
		log.Println("connect to mysql failed")
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		return &resp, nil
	}

	err = gormdb.Debug().Where("user_id = ?", in.UserID).Unscoped().Find(&face).Error
	if err != nil {
		log.Println("query failed")
		resp.ErrCode = common.ERR_QUERY_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
		return &resp, nil
	}
	//先删除index
	for i := 0; i < len(face); i++ {
		errcode := DelCluster((int32)(face[i].ClusterId))
		if errcode != 0 {
			log.Println("delet cluster failed:", face[i].ClusterId)
			resp.ErrCode = errcode
			resp.ErrMsg = common.ErrMsg[errcode]
			return &resp, nil
		}
	}

	//删除mysql
	err = gormdb.Debug().Where(" user_id = ?", in.UserID).Unscoped().Delete(&face).Error
	if err != nil {
		log.Println("delete failed")
		gormdb.Rollback()
		resp.ErrCode = common.ERR_DEL_RECORD_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DEL_RECORD_FAILED]
		return &resp, nil
	}

	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]
	return &resp, nil

}

//删除指定人脸
func (s *FaceAPIServer) DelFace(ctx context.Context, in *apiService.DelFaceReq) (*apiService.DelFaceResp, error) {
	//验证输入参数
	var resp apiService.DelFaceResp
	if in.FaceToken == "" || len(in.FaceToken) <= 0 || in.UserID == "" || len(in.UserID) <= 0 {
		log.Println("invalid params")
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		return &resp, nil
	}

	//首先删除index
	var face FaceLibrary
	gormdb, err := conn.GetGorm()
	if err != nil {
		log.Println("connect to mysql failed")
		resp.ErrCode = common.ERR_CONN_MYSQL_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONN_MYSQL_FAILED]
		return &resp, nil
	}
	err = gormdb.Debug().Where("face_token = ? AND user_id = ?", in.FaceToken, in.UserID).Unscoped().Find(&face).Error
	if err != nil {
		log.Println("query failed")
		resp.ErrCode = common.ERR_QUERY_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_QUERY_FAILED]
		return &resp, nil
	}

	errcode := DelCluster((int32)(face.ClusterId))
	if errcode != 0 {
		log.Println("delet cluster failed:", face.ClusterId)
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		return &resp, nil
	}

	//在数据库中删除
	err = gormdb.Debug().Where("face_token = ? AND user_id = ?", in.FaceToken, in.UserID).Unscoped().Delete(&face).Error
	if err != nil {
		log.Println("delete failed")
		gormdb.Rollback()
		resp.ErrCode = common.ERR_DEL_RECORD_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DEL_RECORD_FAILED]
		return &resp, nil
	}

	/*
		//删除文件
		errcode := removeFile(face.Path)
		if errcode != 0 {
			log.Println("delete file failed:", face.Path)
			resp.ErrCode = errcode
			resp.ErrMsg = common.ErrMsg[errcode]
			return &resp, nil
		}
	*/

	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]
	return &resp, nil
}

//存储图像至本地,返回错误码和存储文件的全路径
func SaveImg(img string) (int32, string) {
	//判断路径是否存在
	exist, err := PathExists(conn.IMG_PATH)
	if err != nil {
		return common.ERR_GET_PATH_FAILED, ""
	}

	if exist {
		log.Println("path exist")
	} else {
		//路径不存在，新建路径
		erro := os.MkdirAll(conn.IMG_PATH, 0777)
		if erro != nil {
			return common.ERR_CREATE_PATH_FAILED, ""
		}
	}

	filename := conn.IMG_PATH + strconv.Itoa(int(time.Now().UnixNano())) + conn.FORMAT
	file, err := os.Create(filename)
	if err != nil {
		log.Println("CreateFile1 file err:", err)
		return common.ERR_CREATE_FILE_FAILED, ""
	}
	defer file.Close()

	//base64解码
	decodeImg, err := base64.StdEncoding.DecodeString(img)
	if err != nil {
		log.Println("DecodeString err:", err)
		return common.ERR_ENCODE_FAILED, ""
	}

	content := []byte(decodeImg)
	w := bufio.NewWriter(file)
	_, err = w.Write(content)
	if err != nil {
		log.Println("CreateFile2 file err:", err)
		return common.ERR_CREATE_FILE_FAILED, ""
	}
	return common.ERR_OK, filename
}

//删除指定文件
func removeFile(filename string) int32 {
	err := os.Remove(filename)
	if err != nil {
		log.Println("remove file failed,filename:", filename)
		return common.ERR_DEL_FILE_FAILED
	}
	return 0
}

//判断路径是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		//路径存在
		return true, nil
	}
	if os.IsNotExist(err) {
		//路径不存在
		return false, nil
	}
	return false, err
}

//根据输入变量拼接需要插入数据库的数据结构
func getStruct(box apiService.Box, clusterID int64) *FaceLibrary {
	var faceLibrary FaceLibrary
	faceLibrary.BoxCenterx = box.Center.CoordinateX
	faceLibrary.BoxCentery = box.Center.CoordinateY
	faceLibrary.BoxScore = box.Score
	faceLibrary.BoxWidth = box.Rect.Width
	faceLibrary.BoxHeight = box.Rect.Height

	faceLibrary.ClusterId = clusterID
	return &faceLibrary
}

/*
func (s *FaceAPIServer) AddFace(ctx context.Context, in *apiService.Image) (*apiService.AddFaceResp, error) {
	return nil, nil
}*/

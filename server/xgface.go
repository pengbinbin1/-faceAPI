package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	common "github.com/Intelligentvision/faceAPI/common"
	"github.com/Intelligentvision/faceAPI/conn"
	apiService "github.com/Intelligentvision/faceAPI/proto/faceAPI"
	pbface "github.com/Intelligentvision/faceAPI/proto/xgface"
)

//调用xgface server ，实现人脸检测，并返回相应信息
func (s *FaceAPIServer) GetDetectInfo(ctx context.Context, in *apiService.DetectReq) (*apiService.DetectResp, error) {

	var resp apiService.DetectResp

	//检测传入参数，若传入参数的图像数据为空，返回非法参数
	if in.Images == nil || len(in.Images[0]) == 0 {
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.DetectInfo = nil
		return &resp, nil
		//return &apiService.DetectResp{ErrCode: common.ERR_INVALID_PARAM, ErrMsg: "", nil}, nil
	}

	//创建xgface连接,并返回人脸检测结果
	detectinfo, errcode := getDetectInfo(in)
	if errcode == common.ERR_OK && len(detectinfo.Faces) > 0 {
		resp.ErrCode = common.ERR_OK
		resp.ErrMsg = common.ErrMsg[common.ERR_OK]
		detectFaces := detectinfo.GetFaces()
		//fmt.Println(detectFaces)
		var detect apiService.DetectInfo

		//赋值人脸数量
		faceNum := len(detectFaces)
		detect.FaceNum = (int32)(faceNum)

		for i := 0; i < faceNum; i++ {
			var face apiService.Face

			//生成facetoken，全局唯一
			face_str := in.Images[0] + strconv.Itoa(i)
			face_token := common.GetFaceToken(face_str)
			face.FaceToken = face_token

			//生成box
			box := generateBox(detectFaces[i])
			face.Box = &box

			//landmark置位时生成landmark
			if in.Facefield.Landmark == 1 {
				landmark := generateLandmark(detectFaces[i])
				face.Landmark = &landmark
			}

			//quality置位时，赋值quality
			if in.Facefield.Quality == 1 {
				face.Yaw = detectFaces[i].GetYaw()
				face.Roll = detectFaces[i].GetRoll()
				face.GlobalIsFace = detectFaces[i].GetGlobalIsFace()
				face.GlobalFrontFace = detectFaces[i].GetGlobalFrontFace()
				face.LocalIsFace = detectFaces[i].GetLocalIsFace()
			}

			//attribute置位时生成attribute
			if in.Facefield.Attributes == 1 {
				attr := generateAttributes(detectFaces[i])
				face.Attributes = attr
			}

			//feture置位时生成feature,feture暂不对外开放
			/*
				if in.Facefield.Feature == 1 {
					feature := generateFeature(detectFaces[i])
					face.Feature = &feature
				}*/

			detect.Faces = append(detect.Faces, &face)
		}

		resp.DetectInfo = &detect
		return &resp, nil
	} else if len(detectinfo.Faces) <= 0 {
		resp.ErrCode = common.ERR_DETECT_NO_FACE
		resp.ErrMsg = common.ErrMsg[common.ERR_DETECT_NO_FACE]
		resp.DetectInfo.Faces = nil
		resp.DetectInfo.FaceNum = 0
		return &resp, nil
	} else {
		//人脸检测出错
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		resp.DetectInfo = nil
		return &resp, nil
	}

}

//调用xgface server,实现两张人脸比对

func (s *FaceAPIServer) MatchFaces(ctx context.Context, in *apiService.Images) (*apiService.ScoreResp, error) {
	//检测传入参数
	var resp apiService.ScoreResp
	if len(in.Image1) == 0 || len(in.Image2) == 0 {
		log.Println("image1 or image2 is null")
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.Score = -1
		return &resp, errors.New("invalid params")
	}

	//创建连接
	connection, err := conn.GetXgfaceConn()
	if err != nil {
		log.Println("get connection to xgface failed")
		resp.ErrCode = common.ERR_CONNECT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_CONNECT_FAILED]
		resp.Score = -1
		return &resp, err
	} else {
		c := pbface.NewXgfaceServiceClient(connection)
		contx, cancle := context.WithTimeout(context.Background(), time.Duration(TIMEOUT)*time.Second)
		defer cancle()

		//构造xgface的输入参数
		var imgs pbface.Images
		imgs.Image1 = in.Image1
		imgs.Image2 = in.Image2

		score, err := c.MatchFaces(contx, &imgs)
		if err != nil {
			log.Println("xgface matchface failed")
			resp.ErrCode = common.ERR_ALG_FAILED
			resp.ErrMsg = common.ErrMsg[common.ERR_ALG_FAILED]
			resp.Score = -1
			return &resp, err
		} else {
			resp.ErrCode = common.ERR_OK
			resp.ErrMsg = common.ErrMsg[common.ERR_OK]
			resp.Score = score.Score
			return &resp, nil
		}
	}
}

//apiServer 对比两张人脸
func (s *FaceAPIServer) MatchFacesEx(ctx context.Context, in *apiService.Images) (*apiService.ScoreResp, error) {
	//检测传入参数
	var resp apiService.ScoreResp
	if len(in.Image1) == 0 || len(in.Image2) == 0 {
		log.Println("image1 or image2 is null")
		resp.ErrCode = common.ERR_INVALID_PARAM
		resp.ErrMsg = common.ErrMsg[common.ERR_INVALID_PARAM]
		resp.Score = -1
		return &resp, nil
	}

	//提取第一张人脸的特征
	var apiReq1 apiService.DetectReq
	var faceField apiService.FaceField
	faceField.Attributes = 0
	faceField.Quality = 0
	faceField.Feature = 1
	faceField.Landmark = 1
	apiReq1.Facefield = &faceField
	apiReq1.Images = append(apiReq1.Images, in.Image1)

	detectinfo1, errcode := getDetectInfo(&apiReq1)
	if errcode != 0 {
		resp.ErrCode = errcode
		resp.ErrMsg = common.ErrMsg[errcode]
		return &resp, nil
	}
	if len(detectinfo1.Faces) <= 0 {
		log.Println("img1 detect no face")
		resp.ErrCode = common.ERR_DETECT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DETECT_FAILED]
		return &resp, nil
	}

	//单张图像多张人脸报错
	if len(detectinfo1.Faces) > 1 {
		log.Println("img1 detect multi faces:", len(detectinfo1.Faces))
		resp.ErrCode = common.ERR_MULTI_FACE
		resp.ErrMsg = common.ErrMsg[common.ERR_MULTI_FACE]
		return &resp, nil
	}
	feature1 := generateFeature(detectinfo1.Faces[0])

	//提取第二张图像的信息
	var apiReq2 apiService.DetectReq
	apiReq2.Facefield = &faceField
	apiReq2.Images = append(apiReq2.Images, in.Image2)
	detectinfo2, errcode2 := getDetectInfo(&apiReq2)
	if errcode != 0 {
		resp.ErrCode = errcode2
		resp.ErrMsg = common.ErrMsg[errcode2]
		return &resp, nil
	}
	if len(detectinfo2.Faces) <= 0 {
		log.Println("img1 detect no face")
		resp.ErrCode = common.ERR_DETECT_FAILED
		resp.ErrMsg = common.ErrMsg[common.ERR_DETECT_FAILED]

		return &resp, nil
	}

	//单张图像多张人脸报错
	if len(detectinfo2.Faces) > 1 {
		log.Println("img1 detect multi faces:", len(detectinfo2.Faces))
		resp.ErrCode = common.ERR_MULTI_FACE
		resp.ErrMsg = common.ErrMsg[common.ERR_MULTI_FACE]
		return &resp, nil
	}

	feature2 := generateFeature(detectinfo2.Faces[0])

	//算比分
	score := generateScore(feature1, feature2)
	resp.Score = (float32)(score)
	resp.ErrCode = common.ERR_OK
	resp.ErrMsg = common.ErrMsg[common.ERR_OK]

	return &resp, nil
}
func getDetectInfo(in *apiService.DetectReq) (*pbface.DetectInfo, int32) {

	connect, err := conn.GetXgfaceConn()
	if err != nil {
		log.Println("get xgfaceConn faield")
		return nil, common.ERR_CONNECT_FAILED
	} else {
		c := pbface.NewXgfaceServiceClient(connect)
		contx, cancel := context.WithTimeout(context.Background(), time.Duration(TIMEOUT)*time.Second)
		defer cancel()

		var req pbface.Request
		//将请求参数逐项付给xgface server，目前只支持单张图片

		req.Images = append(req.Images, in.Images[0])
		var facefield pbface.FaceField
		facefield.Attributes = in.Facefield.Attributes
		facefield.Landmark = in.Facefield.Landmark
		facefield.Quality = in.Facefield.Quality
		facefield.Feature = in.Facefield.Feature //人脸特征不对用户开放
		req.Facefield = &facefield

		detectinfo, err := c.GetDetectInfo(contx, &req)
		if err != nil {
			fmt.Println("xface get detectinfo failed, err:", err)
			return nil, common.ERR_ALG_FAILED
		}
		return detectinfo, common.ERR_OK
	}
}

func generateBox(detectface *pbface.Face) apiService.Box {

	var box apiService.Box

	//获取中心点
	var point apiService.Point
	point.CoordinateX = detectface.Box.Center.CoordinateX
	point.CoordinateY = detectface.Box.Center.CoordinateY
	box.Center = &point

	//获取矩形
	var rect apiService.Rect
	rect.Width = detectface.Box.Rect.Width
	rect.Height = detectface.Box.Rect.Height
	box.Rect = &rect

	//获取分数
	box.Score = detectface.Box.Score
	return box
}

func generateLandmark(detectface *pbface.Face) apiService.Landmark {
	var landmark apiService.Landmark

	pointLen := len(detectface.Landmark.Points)
	for i := 0; i < pointLen; i++ {
		var point apiService.Point
		point.CoordinateX = detectface.Landmark.Points[i].CoordinateX
		point.CoordinateY = detectface.Landmark.Points[i].CoordinateY
		landmark.Points = append(landmark.Points, &point)
	}
	landmark.Score = detectface.Landmark.Score
	return landmark
}

func generateAttributes(detectface *pbface.Face) map[string]int32 {
	attr := make(map[string]int32)
	for k, v := range detectface.Attributes {
		attr[k] = v
	}
	return attr
}

func generateFeature(detectface *pbface.Face) apiService.Feature {
	var feature apiService.Feature
	var values []float32
	dim := len(detectface.Feature.GetValues())
	for i := 0; i < dim; i++ {
		values = append(values, detectface.Feature.Values[i])
	}
	feature.Values = values
	return feature
}
func generateScore(feature1 apiService.Feature, feature2 apiService.Feature) float64 {
	//检查参数是否合法

	if len(feature1.Values) != 192 || len(feature2.Values) != 192 {
		log.Println("feat1 or feat2 is not 192 dim")
		log.Println("feature1 is:", len(feature1.Values), ",feateure2 is:", len(feature2.Values))
		return -1
	}

	var temp2, temp3 float64
	var temp1 float64
	for i := 0; i < 192; i++ {
		temp1 += (float64)(feature1.Values[i]) * (float64)(feature2.Values[i])
		temp2 += math.Pow((float64)(feature1.Values[i]), 2)
		temp3 += math.Pow((float64)(feature2.Values[i]), 2)
	}

	tempScore := (float64)(temp1) / (math.Sqrt(temp2) * math.Sqrt(temp3))

	score := (tempScore + 1) / 2
	return score
}

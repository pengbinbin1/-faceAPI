package server

import (
	"context"

	"log"
	"time"

	common "github.com/Intelligentvision/faceAPI/common"
	"github.com/Intelligentvision/faceAPI/conn"
	xgindex "github.com/Intelligentvision/faceAPI/proto/xgindex"
)

//在index中获取最相似的人脸clusterID 及比分
func getMostSimilarity(Feature []float32) (score float32, clusterID int64, errcode int32) {

	//创建xgindex连接
	connect, err := conn.GetXgindexConn()
	if err != nil {
		return -1, -1, common.ERR_CONN_INDEX_FAILED
	}

	//创建客户端
	c := xgindex.NewIndexServiceClient(connect)
	contx, cancel := context.WithTimeout(context.Background(), time.Duration(TIMEOUT)*time.Second)
	defer cancel()

	var req xgindex.Request

	req.Threshhold = 0.6
	var feature xgindex.Feature
	feature.Value = Feature
	req.Feature = &feature

	resp, err := c.FindWithThreshhold(contx, &req)
	if (err != nil) || (len(resp.Similarity) == 0) {
		log.Println("get simiarity failed")
		return -1, -1, common.ERR_SIMIARITY_FAILED
	}

	for i := 0; i < len(resp.Cids); i++ {
		log.Println("cluster = ", resp.Cids[i], "similarity=", resp.Similarity[i])
	}

	return resp.Similarity[0], (int64)(resp.Cids[0]), common.ERR_OK
}

//将人脸特征插入index
func insertIntoIndex(Feature []float32) (clusterid int64, errcode int32) {
	//创建xgindex连接

	connect, err := conn.GetXgindexConn()
	if err != nil {
		log.Println("connect index failed")
		return -1, common.ERR_CONN_INDEX_FAILED
	}

	//创建客户端
	c := xgindex.NewIndexServiceClient(connect)
	contx, cancel := context.WithTimeout(context.Background(), time.Duration(TIMEOUT)*time.Second)
	defer cancel()
	var feature xgindex.Feature
	feature.Value = Feature
	resp, err := c.InsertCluster(contx, &feature)
	//resp, err := c.InsertPoint(contx, &feature)
	if err != nil {
		log.Println("insert index failed")
		return -1, common.ERR_INSERT_INDEX_FAILED
	}
	return (int64)(resp.Id), common.ERR_OK
}

//根据clusterID 删除cluster
func DelCluster(clusterID int32) (errcode int32) {
	connect, err := conn.GetXgindexConn()
	if err != nil {
		return common.ERR_CONN_INDEX_FAILED
	}

	//创建客户端
	c := xgindex.NewIndexServiceClient(connect)
	contx, cancel := context.WithTimeout(context.Background(), time.Duration(TIMEOUT)*time.Second)
	defer cancel()

	var cluster xgindex.Cluster
	cluster.Id = clusterID
	_, err = c.DeletebyCId(contx, &cluster)
	if err != nil {
		log.Println("index delete cluster by clusterID failed")
		return common.ERR_DEL_CLUSTER_FAILED
	}
	return 0

}

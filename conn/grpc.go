package conn

import (
	"fmt"
	"log"

	"google.golang.org/grpc"
)

var XgfaceConn, XgindexConn *grpc.ClientConn

/*
* DESC: 创建xgface连接句柄
* PARM: 无
* RETURN: 连接句柄，错误码
 */
func newXgfaceConn() (*grpc.ClientConn, error) {
	if nil == XgfaceConn {
		lock.Lock()
		defer lock.Unlock()
		if nil == XgfaceConn {
			var opts []grpc.DialOption
			opts = append(opts, grpc.WithInsecure())
			XgfaceConn, err = grpc.Dial(XGFACE_IPP, opts...)
			if err != nil {
				log.Println("xgface grpc connect failed,err =", err)
				XgfaceConn = nil
				return XgfaceConn, err
			}
			fmt.Println("xgface ipp :", XGFACE_IPP)
		}
	}
	return XgfaceConn, err
}

/*
* DESC: 创建xgindex连接句柄
* PARM: 无
* RETURN: 连接句柄，错误码
 */
func newXgindexConn() (*grpc.ClientConn, error) {
	if nil == XgindexConn {
		lock.Lock()
		defer lock.Unlock()
		if nil == XgindexConn {
			var opts []grpc.DialOption
			opts = append(opts, grpc.WithInsecure())
			XgindexConn, err = grpc.Dial(XGINDEX_IPP, opts...)
			if err != nil {
				log.Println("xgface grpc connect failed,err =", err)
				XgindexConn = nil
				return XgindexConn, err
			}
			fmt.Println("xgindex ipp :", XGINDEX_IPP)
		}
	}
	return XgindexConn, err
}

func GetXgindexConn() (*grpc.ClientConn, error) {
	return newXgindexConn()
}

func GetXgfaceConn() (*grpc.ClientConn, error) {
	return newXgfaceConn()
}

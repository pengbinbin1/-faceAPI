package common

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

func GetFaceToken(in string) string {

	data := []byte(in)
	md5Ctx := md5.New()
	md5Ctx.Write(data)
	temp := md5Ctx.Sum(nil)

	return hex.EncodeToString(temp)

}

func GetImgToken(in string) string {

	data := []byte(in)
	h := sha256.New()
	h.Write(data)
	sha256str := hex.EncodeToString(h.Sum(nil))
	return sha256str
}

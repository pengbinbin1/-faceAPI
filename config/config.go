//读取当前路径下的conf.yaml 文件并解析各项配置参数

package config

import (
	//"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

func init() {
	loadConfig()
}

type cweb struct {
	Port        string `yaml:"port"`
	Baseip      string `yaml:"baseip"`
	Nsqdtcpaddr string `yaml:"nsqdtcpaddr"`
	Xgfaceaddr  string `yaml:"xgfaceaddr"`
	Xgindexaddr string `yaml:"xgindexaddr"`
	Deadline    int    `yaml:"deadline"`
	FaceAPIaddr string `yaml:"faceapiaddr"`
}
type cimage struct {
	Imgbasepath    string `yaml:"imgbasepath"`
	Faceimgpath    string `yaml:"faceimgpath"`
	Errimgpath     string `yaml:"errimgpath"`
	Mainimgpath    string `yaml:"mainimgpath"`
	Facefilepath   string `yaml:"facefilepath"`
	Peopleimgpath  string `yaml:"peopleimgpath"`
	Accessbasepath string `yaml:"accessbasepath"`
}
type cmysql struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
}
type credis struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
}
type Services struct {
	Web   cweb   `yaml:"web"`
	Image cimage `yaml:"image"`
	Mysql cmysql `yaml:"mysql"`
	Redis credis `yaml:"redis"`
}
type Conf struct {
	Version  string   `yaml:"version"`
	Services Services `yaml:"services"`
}

var Config Conf

func loadConfig() {
	yamlFile, err := ioutil.ReadFile("./conf.yml")
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &Config)
	if err != nil {
		log.Fatalf("yamlFile.Unmarshal: %v", err)
	}
	//fmt.Printf("%+v\n", Config)
	//
	//f, err := yaml.Marshal(Config)
	//ioutil.WriteFile("./t3.yml", f, 0666)
}

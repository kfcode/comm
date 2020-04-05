
package config

import (
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	RpcxClientSvrInfo map[string]*CliConfig
	Mutex sync.Mutex
)

type CliSvrInfo struct {
	SvrId   int32
	Ip      string
	Port    int32
	Weight  int32
	Valid   bool
}

type CliConfig struct {
	LastCheckTime    int64
	SvrNum           int32
	ConnTimeOut      int32
	ReadWriteTimeOut int32
	SvrInfo          map[int]*CliSvrInfo
}

func init() {

	RpcxClientSvrInfo = make(map[string]*CliConfig)
}


func GetClientConfig(psm string) (*CliConfig, error) {
	var err error
	if len(psm) == 0 {
		return nil, errors.WithMessage(err, "psm invalid len")
	}
	if _, ok := RpcxClientSvrInfo[psm]; !ok {  //reload
		return loadPsmFileConfig(psm)
	}else if time.Now().Unix() - RpcxClientSvrInfo[psm].LastCheckTime > 30 {   //check时间是否过期，过期//reload一次
		return loadPsmFileConfig(psm)
	}
	return RpcxClientSvrInfo[psm], nil
}


func getClientFIleName(psm string) string {
	return "./" + psm + "_cli.conf"
}

func loadPsmFileConfig(psm string) (*CliConfig, error) {

	fileName := getClientFIleName(psm)
	cfg, err := ini.Load(fileName)
	if err != nil {
		return nil, errors.WithMessage(err, "ini load file failed fileName: " + fileName)
	}

	general, err := cfg.GetSection("General")
	if err != nil || general == nil {
		return nil, errors.WithMessage(err, "config file not find section General: " + fileName)
	}

	var svrCount, connTimeOut, readWriteTimeOut int

	//读写svr数量
	strSvrCount, err := general.GetKey("SvrCount")
	if strSvrCount == nil || err != nil {
		return nil, errors.WithMessage(err, "config file not find  SvrCount in General: " + fileName)
	}
	svrCount, err = strconv.Atoi(strSvrCount.Value())
	if err != nil {
		return nil, errors.WithMessage(err, "config file invalid SvrCount")
	}

	//读取链接超时
	strConnTimeOut, err := general.GetKey("ConnTimeOut")
	if strConnTimeOut == nil || err != nil {
		connTimeOut = 100
	}else {
		connTimeOut, err = strconv.Atoi(strConnTimeOut.Value())
		if err != nil {
			connTimeOut = 100
		}
	}

	//解析读写超时
	strRadWriteTimeOut, err := general.GetKey("ReadWriteTimeOut")
	if strRadWriteTimeOut == nil || err != nil {
		readWriteTimeOut = 3000
	}else {
		connTimeOut, err = strconv.Atoi(strRadWriteTimeOut.Value())
		if err != nil {
			connTimeOut = 3000
		}
	}

	svrInfoList, err := loadSvrInfo(cfg, svrCount)
	if err != nil {
		return nil, errors.WithMessage(err, "loadSvrInfo failed psm:" + psm)
	}

	cliInfo := CliConfig{}
	cliInfo.LastCheckTime = time.Now().Unix()
	cliInfo.ConnTimeOut = int32(connTimeOut)
	cliInfo.ReadWriteTimeOut = int32(readWriteTimeOut)
	cliInfo.SvrNum = int32(svrCount)
	cliInfo.SvrInfo = svrInfoList

	Mutex.Lock()
	defer Mutex.Unlock()

	if _, ok := RpcxClientSvrInfo[psm]; ok {  //找到配置了
		if RpcxClientSvrInfo[psm].LastCheckTime <  int64(30) {  ////上次更新时间小于 30 不更新，直接返回
			return RpcxClientSvrInfo[psm], nil
		}
		if checkFileChange(psm, RpcxClientSvrInfo[psm].LastCheckTime) {  //大于 30s看看是否文件有变化，有更新，没有返回
			RpcxClientSvrInfo[psm] = &cliInfo
			return &cliInfo, nil
		}
		return RpcxClientSvrInfo[psm], nil
	}
	RpcxClientSvrInfo[psm] = &cliInfo
	return &cliInfo, nil
}

func loadSvrInfo(file *ini.File, svrCount int) ( map[int]*CliSvrInfo, error) {

	svrInfoList := make(map[int]*CliSvrInfo, 0)
	var err error
	for i := 0; i< svrCount; i++ {
		sectionName := "Svr" + strconv.Itoa(i)
		svrName, err := file.GetSection(sectionName)
		if err != nil || svrName == nil {
			return nil, errors.WithMessage(err, "config file not find Svr: "+ sectionName)
		}

		strSvrIp := svrName.Key("Ip")
		if strSvrIp == nil {
			return nil, errors.WithMessage(err, "config file invalid not find Ip in Svr")
		}

		strSvrPort := svrName.Key("Port")
		if strSvrPort == nil {
			return nil, errors.WithMessage(err, "config file invalid not find Port in Svr")
		}
		port, err := strconv.Atoi(strSvrPort.Value())
		if err != nil || port <= 0 {
			return nil, errors.WithMessage(err, "config file invalid not find Port in Svr or port invalid")
		}

		weight := 1000
		strWeight  := svrName.Key("Weight")
		if strWeight == nil {
			weight = 1000
		}else {
			weight, err = strconv.Atoi(strWeight.Value())
			if err != nil || port <= 0 {
				weight = 1000
			}
		}

		valid := true
		strValid  := svrName.Key("Valid")
		if strWeight == nil {
			valid = true
		}else {

			if strValid.Value() == "false" {
				valid = false
			}
		}
		svrInfo := &CliSvrInfo{}
		svrInfo.Ip = strSvrIp.Value()
		svrInfo.Port = int32(port)
		svrInfo.Weight = int32(weight)
		svrInfo.Valid = valid
		if valid == true {
			svrInfoList[i+1] = svrInfo
		}
	}
	return svrInfoList, err
}

func checkFileChange(psm string, lastTime int64) bool {

	fileName := getClientFIleName(psm)
	info, err := os.Stat(fileName)
	if err != nil {
		return true
	}
	if lastTime <= info.ModTime().Unix() {
		return true
	}
	return false
}

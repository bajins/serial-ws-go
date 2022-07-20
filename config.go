package main

import (
	"gopkg.in/ini.v1"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

var cfg *ini.File

func initConfig() {
	// 判断配置文件是否存在
	str, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	if str == "" && len(str) <= 0 {
		log.Fatal("当前目录获取失败")
	}
	if !IsFileExist(filepath.Join(str, configFileName)) { // 文件不存在释放到本地
		ReleaseFile(conf, configFileName)
	}
	cfg, err = ini.Load(configFileName)
	if err != nil {
		log.Fatalf("Fail to read file: %v", err)
	}
	server := cfg.Section("server")
	if server == nil {
		server, err = cfg.NewSection("server")
		if err != nil {
			log.Fatal(err)
		}
	}
	host := server.Key("host")
	if host == nil {
		host, err = server.NewKey("host", "127.0.0.1")
		if err != nil {
			log.Fatal(err)
		}
	}
	if host.Value() == "" || len(host.Value()) <= 0 { // 命令行
		host.SetValue(flagVar.Server.Host)
	}
	port := server.Key("port")
	if port == nil {
		port, err = server.NewKey("port", "8080")
		if err != nil {
			log.Fatal(err)
		}
	}
	if port.Value() == "" || len(port.Value()) <= 0 { // 命令行
		host.SetValue(strconv.Itoa(flagVar.Server.Port))
	}

	serial_ := cfg.Section("serial")
	if serial_ == nil {
		serial_, err = cfg.NewSection("serial")
		if err != nil {
			log.Fatal(err)
		}
	}
	serialName := serial_.Key("name")
	if serialName == nil {
		serialName, err = serial_.NewKey("name", "")
		if err != nil {
			log.Fatal(err)
		}
	}
	if serialName.Value() == "" || len(serialName.Value()) <= 0 {
		if flagVar.Serial.Name != "" && len(flagVar.Serial.Name) > 0 { // 命令行
			serialName.SetValue(flagVar.Serial.Name)
		} else {
			serialName.SetValue(scanSerialName())
		}
	}
	// 判断配置文件中的串口名是否存在设备中
	if _, ok := portSet[serialName.Value()]; !ok { // 不存在
		serialName.SetValue(scanSerialName())
	}
	baud := serial_.Key("baud")
	if baud == nil {
		baud, err = serial_.NewKey("baud", "")
		if err != nil {
			log.Fatal(err)
		}
	}
	if baud.Value() == "" || len(baud.Value()) <= 0 {
		if flagVar.Serial.Baud != "" && len(flagVar.Serial.Baud) <= 0 {
			baud.SetValue(flagVar.Serial.Baud)
		} else {
			baud.SetValue(scanSerialBaud())
		}
	}
	flagVar.Serial.Name = serialName.Value()
	flagVar.Serial.Baud = baud.Value()

	err = cfg.SaveTo(configFileName)
	if err != nil {
		log.Fatal(err)
	}
}

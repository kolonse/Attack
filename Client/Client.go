// Server project Server.go
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

var url = flag.String("url", "http://localhost:10080", "服务器地址URL")
var cache = flag.String("cache", "attack/cache", "缓存目录")

type Package struct {
	Name    string
	Url     string
	Md5     string
	Version string
	Cmd     [][]string
}

type PackageFile struct {
	Version  string
	Packages []Package
}

func Load(str string) (*PackageFile, error) {
	value := &PackageFile{}
	err := json.Unmarshal([]byte(str), value)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func getNotSame(localFile string, remoteObj *PackageFile) []Package {
	file, err := os.Open(localFile)
	if err != nil {
		println("open " + localFile + " Error:" + err.Error())
		return remoteObj.Packages
	}
	defer file.Close()
	buff, err := ioutil.ReadAll(file)
	if err != nil {
		println("ReadAll " + localFile + " Error:" + err.Error())
		return remoteObj.Packages
	}
	value, err := Load(string(buff))
	if err != nil {
		println("Load " + localFile + " Error:" + err.Error())
		return remoteObj.Packages
	}

	ret := []Package{}
	local := make(map[string]Package)
	for _, v := range value.Packages {
		local[v.Name] = v
	}

	for _, v := range remoteObj.Packages {
		o, ok := local[v.Name]
		if ok {
			if o.Version != v.Version {
				ret = append(ret, o)
			}
		}
	}
	return ret
}

func process(pck Package) error {
	// 下载包
	resp, err := http.Get(pck.Url)
	if err != nil {
		println("Get "+pck.Url+" Error:", err.Error())
		return err
	}
	defer resp.Body.Close()
	buff := make([]byte, 1024*1024*10)
	rd := bufio.NewReader(resp.Body)
	dir := filepath.Join(os.Getenv("TEMP"), *cache)
	file := filepath.Join(dir, pck.Name+".att")
	fd, err := os.Create(file)
	if err != nil {
		println("Open "+file+" Error:", err.Error())
		return err
	}

	for {
		n, err := rd.Read(buff)
		if err != nil {
			if err.Error() != "EOF" {
				fd.Close()
				os.Remove(file)
				println("Read "+file+" Error:", err.Error())
				return err
			}
			break
		}
		fd.Write(buff[:n])
	}
	fd.Close()
	// 进行重命名
	err = os.Rename(file, file[0:len(file)-len(".att")])
	if err != nil {
		os.Remove(file)
		println("Rename "+file+" Error:", err.Error())
		return err
	}
	// 运行安装命令
	for _, cmd := range pck.Cmd {
		if len(cmd) == 0 {
			continue
		} else if len(cmd) > 1 {
			c := exec.Command(cmd[0], cmd[1:]...)
			c.Stdout = os.Stdout
			c.Dir = dir
			err = c.Run()
			if err != nil {
				println("Run "+cmd[0]+" Error:", err.Error())
				return err
			}
		} else {
			c := exec.Command(cmd[0])
			c.Stdout = os.Stdout
			c.Dir = dir
			err = c.Run()
			if err != nil {
				println("Run "+cmd[0]+" Error:", err.Error())
				return err
			}
		}
	}

	println("Process " + pck.Name + " Success")
	return nil
}

func update(list []Package) []Package {
	updateSuccessList := make([]Package, 0)
	for _, p := range list {
		err := process(p)
		if err != nil {
			println("Process " + p.Name + " Fail")
			continue
		}
		updateSuccessList = append(updateSuccessList, p)
	}

	return updateSuccessList
}

func Run() {
	// 下载版本文件
	versionUrl := *url + "/version.json"
	println("versionUrl:" + versionUrl)
	resp, err := http.Get(versionUrl)
	if err != nil {
		println("GET " + versionUrl + " Error:" + err.Error())
		return
	}
	defer resp.Body.Close()
	buff, _ := ioutil.ReadAll(resp.Body)
	println("GET " + versionUrl + " Success:" + string(buff))
	value, err := Load(string(buff))
	if err != nil {
		println("Unmarshal " + versionUrl + " Error:" + err.Error())
		return
	}
	println("Unmarshal " + versionUrl + " Success")
	cacheDir := filepath.Join(os.Getenv("TEMP"), *cache)
	err = os.MkdirAll(cacheDir, 777)
	if err != nil {
		println("MkdirAll " + cacheDir + " Error:" + err.Error())
		return
	}
	println("MkdirAll " + cacheDir + " Success")
	// 与本地文件进行比较
	versionLocal := filepath.Join(cacheDir, "version.json")
	updatePackage := getNotSame(versionLocal, value)
	if len(updatePackage) == 0 {
		println("Local Same With Remote,Version:" + value.Version)
		return
	}
	// 对要更新的内容进行格式化
	buff, _ = json.Marshal(updatePackage)
	println("Update List:" + string(buff))
	value.Packages = update(updatePackage)
	buff, _ = json.Marshal(value)
	err = ioutil.WriteFile(versionLocal, buff, 666)
	if err != nil {
		println("Save File " + versionLocal + " Error:" + err.Error())
		return
	}
	println("Update " + versionLocal + " Completed")
}

func main() {
	Run()
}

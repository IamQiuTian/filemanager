package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	download  *bool   = flag.Bool("download", false, "download file")
	upload    *bool   = flag.Bool("upload", false, "upload file")
	directory *string = flag.String("d", "nil", "Directory path")
	file      *string = flag.String("f", "nil", "file path")
	port      *string = flag.String("p", "8888", "Listening port")
	pwd       *string = flag.String("pwd", "nil", "password")
)

var (
	randomstr   = randomStr(28)
	randomstrup = randomStr(28)
	ipList      = getIP()
)

var filename string

func main() {
	flag.Parse()
	// 参数为空或参数都不为空
	switch {
	case *download == false && *upload == false || *download != false && *upload != false:
		flag.Usage()
		return
	case *directory == "nil" && *file == "nil" || *directory != "nil" && *file != "nil":
		flag.Usage()
		return
	case *upload == true && *directory == "nil":
		flag.Usage()
		return
	default:
		func() {
			if *download {
				fileDown()
			} else {
				uploadFile()
			}
		}()
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}

// 文件上传控制
func uploadFile() {
	ok, _ := pathExist(*directory)
	if !ok {
		log.Fatal("Directory does not exist")
	}

	http.HandleFunc(fmt.Sprintf("/%s", randomstrup), func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"%s  %s  %s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
		)

		fileup, fileupinfo, err := r.FormFile("file")
		if err != nil {
			w.Write([]byte("upload error!"))
			return
		}

		fileupname := fileupinfo.Filename
		filewn, err := os.OpenFile(fmt.Sprintf("%s/%s", *directory, fileupname), os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			w.Write([]byte("upload error!"))
			return
		}

		defer func() {
			if err := fileup.Close(); err != nil {
				log.Fatal("Close: ", err.Error())
				return
			}
			if err := filewn.Close(); err != nil {
				log.Fatal("Close: ", err.Error())
				return
			}
		}()

		_, err = io.Copy(filewn, fileup)
		if err != nil {
			w.Write([]byte("file upload error"))
		}
		w.Write([]byte("file upload success"))

	})

	if *pwd != "nil" {
		http.HandleFunc("/", authIndex)
		http.HandleFunc(fmt.Sprintf("/%s", randomStr), uploadIndex)
		http.HandleFunc("/check_auth", auth_Check)

	} else {
		http.HandleFunc(fmt.Sprintf("/"), uploadIndex)
	}
	for _, ip := range ipList {
		fmt.Printf("upload link: http://%s:%s/\n\n", ip, *port)
	}
}

/// 文件上传页面展示
func uploadIndex(w http.ResponseWriter, r *http.Request) {
	tpl, err := ioutil.ReadFile("./static/upload.html")
	if err != nil {
		w.Write([]byte("error!"))
	}
	log.Printf(
		"%s  %s  %s",
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
	)
	t, _ := template.New("fileupload").Parse(string(tpl))
	t.Execute(w, randomstrup)
}

// 文件下载
func fileDown() {
	// 如果未设置密码就直接生成下载链接
	if *pwd == "nil" {
		if *file != "nil" {
			file_Server()
		} else {
			directory_Server()
		}

		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
	}

	if *file != "nil" {
		file_Server()
	} else {
		directory_Server()
	}

	http.HandleFunc("/", authIndex)
	http.HandleFunc("/check_auth", auth_Check)
	fmt.Printf("Please access %s at the browser!\n\n", "0.0.0.0:"+*port)
}

// 密码验证页面
func authIndex(w http.ResponseWriter, r *http.Request) {
	tpl, err := ioutil.ReadFile("./static/password.html")
	if err != nil {
		w.Write([]byte("error!"))
	}

	log.Printf(
		"%s  %s  %s",
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
	)

	t, _ := template.New("filedown").Parse(string(tpl))
	t.Execute(w, nil)
}

// 校验密码
func auth_Check(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	if password != *pwd {
		w.Write([]byte("password error!"))
		return
	}

	if *upload != false {
		http.Redirect(w, r, fmt.Sprintf("/%s", randomstr), 302)
	}

	if *file != "false" {
		http.Redirect(w, r, fmt.Sprintf("/%s/%s", randomstr, filename), 302)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/%s/", randomstr), 302)
	}

}

// 生成随机字符串
func randomStr(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// 文件下载
func file_Server() {
	ok, _ := pathExist(*file)
	if !ok {
		log.Fatal("File does not exist")
	}

	filename = filepath.Base(*file)

	if *pwd == "nil" {
		randomstr = "/"
	} else {
		randomstr = "/" + randomstr + "/"
	}

	for _, ip := range ipList {
		fmt.Printf("Download link: http://%s:%s%s%s\n\n", ip, *port, randomstr, filename)
	}

	// 打印访问日志
	http.HandleFunc(fmt.Sprintf("%s%s", randomstr, filename), func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"%s  %s  %s",
			r.RemoteAddr,
			r.Method,
			r.RequestURI,
		)
		http.ServeFile(w, r, *file)
	})
}

// 目录游览
func directory_Server() {
	ok, _ := pathExist(*directory)
	if !ok {
		log.Fatal("Directory does not exist")
	}

	if *pwd == "nil" {
		randomstr = "/"
	} else {
		randomstr = "/" + randomStr(7) + "/"
	}

	for _, ip := range ipList {
		fmt.Printf("Access link: http://%s:%s%s\n\n", ip, *port, randomstr)
	}

	http.Handle(fmt.Sprintf("%s", randomstr), func(prefix string, h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf(
				"%s  %s  %s",
				r.RemoteAddr,
				r.Method,
				r.RequestURI,
			)
			if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
				r2 := new(http.Request)
				*r2 = *r
				r2.URL = new(url.URL)
				*r2.URL = *r.URL
				r2.URL.Path = p
				h.ServeHTTP(w, r2)
			} else {
				http.NotFound(w, r)
			}
		})
	}(fmt.Sprintf("%s", randomstr), http.FileServer(http.Dir(*directory))))
}

// 判断文件或目录是否存在
func pathExist(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 获取IP列表
func getIP() (ipList []string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		goto getPublic
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ipList = append(ipList, ipnet.IP.String())
			}
		}
	}

getPublic:
	client := http.Client{
		Timeout: time.Duration(2 * time.Second),
	}
	resp, err := client.Get("http://members.3322.org/dyndns/getip")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	ipList = append(ipList, strings.Replace(string(b), "\n", "", -1))
	return
}

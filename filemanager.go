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
	directory *string = flag.String("d", "", "Directory path")
	file      *string = flag.String("f", "", "file path")
	port      *string = flag.String("p", "8888", "Listening port")
	pwd       *string = flag.String("pwd", "", "password")
)

var (
	filename  string
	randomstr = randomStr()
	ipList    = getIP()
)

// Password validation page template
var passwdtpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>password</title>
    <style type="text/css">
        body{
            margin:0px;
            background-color: white;
            font-family: 'PT Sans', Helvetica, Arial, sans-serif;
            text-align: center;
            color: #A6A6A6;
        }
        input{
            width: 100%;
            height: 50px;
            border:none;
            padding-left:3px;
            font-size: 18px;
        }
        input:focus {
            outline: none;
        }
        img{
            width: 40px;
            height: 25px;
            position: absolute;
            right: 0px;
            margin: 15px;
        }
        button{
            width: 200px;
            height: 50px;
            margin-top: 25px;
            background: #1E90FF;
            border-radius: 10px;
            border:none;
            font-size: 18px;
            font-weight: 700;
            color: #fff;
        }
        button:hover {
            background: #79A84B;
            outline: 0;
        }
        .input_block {
            border-bottom: 1px solid rgba(0,0,0,.1);
        }
        /*container*/
        #page_container{
            margin: 50px;
        }
    </style>
</head>
<body>
<form  action="/authcheck" method='post'>
    <div id="page_container">
        <div class="input_block" id="psw_invisible">
            <input type="password" id="input_invisible" placeholder="Password" name="password"/>
        </div>
        <button onclick="">Enter</button>
    </div>
</form>
</body>
</html>
`

var uploadtpl = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>upload</title>
</head>
<style>
    .upload {
        position: relative;
        display: inline-block;
        background: #D0EEFF;
        border: 1px solid #99D3F5;
        border-radius: 4px;
        padding: 4px 12px;
        overflow: hidden;
        color: #1E88C7;
        text-decoration: none;
        text-indent: 0;
        line-height: 20px;
    }
    .upload input {
        position: absolute;
        font-size: 100px;
        right: 0;
        top: 0;
        opacity: 0;
    }
    .upload:hover {
        background: #AADFFD;
        border-color: #78C3F3;
        color: #004974;
        text-decoration: none;
    }
</style>
<body>
<form id="uploadForm" method="POST" enctype="multipart/form-data" action="/{{ . }}">
    <input type="FILE" id="file" name="file" class="upload"/>
    <input type="SUBMIT" value="upload"  class="upload">
</form>
</body>
</html>
`

func main() {
	flag.Parse()
	switch {
	case *download == false && *upload == false || *download != false && *upload != false:
		flag.Usage()
		return
	case *directory == "" && *file == "" || *directory != "" && *file != "":
		flag.Usage()
		return
	case *upload == true && *directory == "":
		flag.Usage()
		return
	case *pwd == "":
		notPwd()
	case *pwd != "":
		aPwd()
	default:
		flag.Usage()
		return
	}

}

// No password verification is required
func notPwd() {
	switch {
	case *file != "" && *download:
		if ok, _ := pathExist(*file); !ok {
			log.Fatal("file does not exist")
		}

		filename = filepath.Base(*file)
		http.HandleFunc(fmt.Sprintf("/%s", filename), fileServer)
		for _, ip := range ipList {
			fmt.Printf("link: http://%s:%s/%s\n\n", ip, *port, filename)
		}

	case *directory != "" && *download:
		if ok, _ := pathExist(*directory); !ok {
			log.Fatal("Directory does not exist")
		}

		http.Handle("/", dirServer("/", http.FileServer(http.Dir(*directory))))
		for _, ip := range ipList {
			fmt.Printf("link: http://%s:%s/\n\n", ip, *port)
		}
	case *directory != "" && *upload:
		http.HandleFunc(fmt.Sprintf("/"), uploadIndex)
		http.HandleFunc(fmt.Sprintf("/%s", randomstr), uploadFile)
		for _, ip := range ipList {
			fmt.Printf("link: http://%s:%s/\n\n", ip, *port)
		}
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}

// You need to verify the password
func aPwd() {
	switch {
	case *file != "" && *download:
		if ok, _ := pathExist(*file); !ok {
			log.Fatal("file does not exist")
		}

		filename = filepath.Base(*file)
		http.HandleFunc(fmt.Sprintf("/%s/%s", randomstr, filename), fileServer)
		for _, ip := range ipList {
			fmt.Printf("link: http://%s:%s/\n\n", ip, *port)
		}

	case *directory != "" && *download:
		if ok, _ := pathExist(*directory); !ok {
			log.Fatal("Directory does not exist")
		}

		http.Handle(fmt.Sprintf("/%s", randomstr), dirServer(fmt.Sprintf("/%s", randomstr), http.FileServer(http.Dir(*directory))))
		for _, ip := range ipList {
			fmt.Printf("link: http://%s:%s/\n\n", ip, *port)
		}
	case *directory != "" && *upload:
		http.HandleFunc(fmt.Sprintf("/%s/%s", randomstr, "upload"), uploadIndex)
		http.HandleFunc(fmt.Sprintf("/%s", randomstr), uploadFile)
		for _, ip := range ipList {
			fmt.Printf("link: http://%s:%s/\n\n", ip, *port)
		}
	}
	http.HandleFunc("/", authIndex)
	http.HandleFunc("/authcheck", authCheck)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", *port), nil))
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s  %s  %s",
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
	)

	file, fileinfo, err := r.FormFile("file")
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	filename := fileinfo.Filename
	tofile, err := os.OpenFile(fmt.Sprintf("%s/%s", *directory, filename), os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		w.Write([]byte(err.Error()))
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal("Close: ", err.Error())
			return
		}
		if err := tofile.Close(); err != nil {
			log.Fatal("Close: ", err.Error())
			return
		}
	}()

	_, err = io.Copy(tofile, file)
	if err != nil {
		log.Println(err)
	}
	w.Write([]byte("file upload success"))

}

func uploadIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s  %s  %s",
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
	)
	t, _ := template.New("").Parse(uploadtpl)
	t.Execute(w, randomstr)
}

func authIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s  %s  %s",
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
	)

	t, _ := template.New("").Parse(passwdtpl)
	t.Execute(w, nil)
}

func authCheck(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	if password != *pwd {
		w.Write([]byte("password error!"))
		return
	}

	if *upload != false {
		http.Redirect(w, r, fmt.Sprintf("/%s/%s", randomstr, "upload"), 302)
	}

	http.Redirect(w, r, fmt.Sprintf("/%s/%s", randomstr, filename), 302)
}

func fileServer(w http.ResponseWriter, r *http.Request) {
	log.Printf(
		"%s  %s  %s",
		r.RemoteAddr,
		r.Method,
		r.RequestURI,
	)
	http.ServeFile(w, r, *file)
}

func dirServer(prefix string, h http.Handler) http.Handler {
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
}

func randomStr() string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 16; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

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

func removeDuplicateElement(addrs []string) []string {
	result := make([]string, 0, len(addrs))
	temp := map[string]struct{}{}
	for _, item := range addrs {
		if _, ok := temp[item]; !ok {
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

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
	resp, err := client.Get("http://ifconfig.me")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	ipList = append(ipList, strings.Replace(string(b), "\n", "", -1))
	return removeDuplicateElement(ipList)
}

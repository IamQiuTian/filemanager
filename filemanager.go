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
	filename  = ""
	randomstr = randomStr()
	ipList    = getIP()
)

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
		noPwd()
	case *pwd != "":
		isPwd()
	default:
		flag.Usage()
		return
	}

}

func noPwd() {
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

func isPwd() {
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
	t, _ := template.New("").Parse(string(tpl))
	t.Execute(w, randomstr)
}

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

	t, _ := template.New("").Parse(string(tpl))
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

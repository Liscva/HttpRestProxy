package main

import (
	"encoding/json"
	"fmt"
	io "io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

type Server struct {
	Port string `json:port`
}
type Porxys struct {
	Porxys []Porxy
}

type Porxy struct {
	Ip        string `json:ip`
	Url       string `json:url`
	ProxyAdd  string `json:proxyAdd`
	AllowPath string `json:allowPath`
}

func (proxys Porxys) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//1.循环代理配置文件，和访问的请求进行匹配，目前有IP和url匹配2种模式
	for _, v := range proxys.Porxys {
		//IP匹配模式：根据IP去代理
		if v.Ip != "" {
			var ipArr = strings.Split(v.Ip, ",")
			var remoteIp = strings.Split(r.RemoteAddr, ":")[0]
			flag := false
			for i := range ipArr {
				if strings.Contains(remoteIp, ipArr[i]) {
					flag = true
				}
			}
			if flag {
				//配置了多个IP地址进行随机轮询代理
				rand.Seed(time.Now().UnixNano())
				var proxyArr = strings.Split(v.ProxyAdd, ",")
				var proxyIp = proxyArr[rand.Intn(len(proxyArr))]
				fmt.Println("请求地址：" + r.RemoteAddr + ",代理地址：" + proxyIp)

				remote, _ := url.Parse(proxyIp)
				proxy := myReverseProxy(remote, w, v)
				proxy.ServeHTTP(w, r)
			}
		} else if v.Url != "" {
			//url匹配模式：根据url去代理
			if strings.Index(r.RequestURI, v.Url) == 0 {
				remote, _ := url.Parse(v.ProxyAdd)
				// proxy := httputil.NewSingleHostReverseProxy(remote) // 代理的核心方法 内部实现了请求改写功能
				proxy := myReverseProxy(remote, w, v) // 使用自定义ReverseProxy
				proxy.ServeHTTP(w, r)
			}
		}
	}
}

// 使用自定义 ReverseProxy
func myReverseProxy(target *url.URL, w http.ResponseWriter, proxypz Porxy) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
		// 自定义header处理 删除指定header
		//if _, ok := req.Header["Test"]; ok {
		//	req.Header.Del("Test")
		//}

		//过滤白名单
		if proxypz.AllowPath != "" {
			var allowArr = strings.Split(proxypz.AllowPath, ",")
			flag := false
			for i := range allowArr {
				if strings.Contains(req.URL.Path, allowArr[i]) {
					flag = true
				}
			}
			if !flag {
				w.WriteHeader(404)
			}
		}
		// 读取并修改 request.Body
		// 在反向代理过程中 修改请求体内容
		//if req.Body != nil && req.URL.Path == "/internal/security/login" {
		//	bodyBytes, _ := ioutil.ReadAll(req.Body)
		//	fmt.Println("读取原始请求体--->", string(bodyBytes))
		//	// 替换请求体内容
		//	bodyBytes = []byte(`{"username":"admin", "password":"xxxxxx"}`)
		//	req.ContentLength = int64(len(bodyBytes)) // 替换后 ContentLength 会变化 需要同步修改
		//	// 生成 req.Body
		//	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		//}
	}

	errorHandler := func(res http.ResponseWriter, req *http.Request, err error) {
		res.Write([]byte(err.Error()))
	}
	return &httputil.ReverseProxy{Director: director, ErrorHandler: errorHandler}
}

// 使用自定义 ReverseProxy 工具函数
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// 使用自定义 ReverseProxy 工具函数
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}

/*************************************************
Function: LoadConfig
Description: read config file to config struct
@parameter filename: config file
Return: Config,bool
*************************************************/

func LoadServerConfig(filename string) (Server, bool) {
	var conf Server
	b, err := io.ReadFile(filename) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	configJson := string(b) // convert content to a 'string'
	fmt.Println(configJson) // print the content as a 'string'
	datajson := []byte(configJson)
	err = json.Unmarshal(datajson, &conf)
	if err != nil {
		fmt.Println("unmarshal json file error")
		return conf, false
	}
	return conf, true
}

func LoadPorxyConfig(filename string) []Porxy {
	b, err := io.ReadFile(filename) // just pass the file name
	if err != nil {
		fmt.Print(err)
	}
	configJson := string(b) // convert content to a 'string'
	fmt.Println(configJson) // print the content as a 'string'
	jsonAsBytes := []byte(configJson)
	configs := make([]Porxy, 0)
	err = json.Unmarshal(jsonAsBytes, &configs)
	if err != nil {
		panic(err)
		os.Exit(1)
	}
	return configs
}

func startServer(server Server, porxys []Porxy) {
	var proxyList Porxys
	proxyList.Porxys = porxys
	mux := http.NewServeMux()
	mux.Handle("/", &proxyList)
	fmt.Println("Success Start HttpRestProxy in 0.0.0.0:" + server.Port)
	fmt.Println("The HttpRestProxy Created By Liscva!")
	// 注册被代理的服务器 (host,port)
	err := http.ListenAndServe(":"+server.Port, mux)
	// http.ListenAndServe 第二个参数是 Handler interface 所以 service 需要实现该接口 提供 ServeHTTP 方法
	// type Handler interface {
	// 	ServeHTTP(ResponseWriter, *Request)
	// }
	if err != nil {
		log.Fatalln("ListenAndServe: ", err)
	}
}

func main() {
	server, flag := LoadServerConfig("./config.json")
	if !flag {
		fmt.Println("InitConfig failed")
		os.Exit(1)
	}
	proxys := LoadPorxyConfig("./proxy.json")
	startServer(server, proxys)
}

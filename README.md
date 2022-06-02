
<h1 align="center" style="margin: 30px 0 30px; font-weight: bold;">HttpRestpRroxy v1.0.0</h1>
<h4 align="center">一个轻量级请求代理框架，让代理变得简单、优雅！</h4>


---




## HttpRestpRroxy 介绍

**HttpRestpRroxy** 是一个轻量级请求代理框架，简单配置文件即可快速开始代理

配置config.json,port为代理监听端口
``` json
{
  "port": "6088"
}
```

配置proxy.json
``` json
[
  {
    "ip": "10.165.3.15",
    "url": "",
    "proxyAdd": "https://10.165.3.230",
    "allowPath": ""
  },
  {
    "ip": "",
    "url": "/arcgis",
    "proxyAdd": "http://127.0.0.1:8055,http://127.0.0.1:8056,http://127.0.0.1:8057",
    "allowPath": "MapServer,Static"
  }
]
```
有2种代理方式，第一种根据访问的IP进行代理，第二种根据url匹配进行代理
如果有多个proxyAdd，会随机请求

## HttpRestpRroxy 打包
程序用go写的，熟悉的人自行打包即可，觉得打包麻烦的可以直接下载build中的对应系统的程序即可，linux的程序需要给权限


## 有问题欢迎请联系我
QQ：3214444445

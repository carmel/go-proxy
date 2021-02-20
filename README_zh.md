# go-proxy
golang实现的内网穿透，基于[IntranetPenetration](https://github.com/alongL/IntranetPenetration)有所改进。


高性能http代理服务器，主要应用与内网穿透。支持多站点配置、客户端与服务端连接中断自动重连，多路传输，大大的提高请求处理速度，go语言编写，无第三方依赖，经过测试内存占用小，普通场景下，仅占用10m内存。


## 特点

- [X] 支持多站点配置
- [X] 断线自动重连
- [X] 支持多路传输,提高并发
- [X] 跨站自动匹配替换


## 编译安装
```sh
git clone github.com/carmel/go-proxy.git
cd go-proxy && make
```

## 使用说明 
- 服务端 
  ```
  ./server -token DKibZF5TXvic1g3kY -log proxy.log -tcpport=8284 -httpport=8024
  ```

  参数 | 含义
  ---|---
  log | 日志文件路径
  token | 验证密钥
  tcpport | 服务端与客户端通信端口
  httpport | 代理的http端口

- 客户端
  ```sh
  ./client -conf conf.yml
  ```
  参数 | 含义
  ---|---
  conf | yaml配置文件路径

  - yaml配置示例
    ```yml
    server-addr: 23.23.23.23:8284
    token: DKibZF5TXvic1g3kY
    max-conn: 1
    domain-as-proto: false
    tunnel:
      - domain: 23.23.23.23
        proto: 127.0.0.1:80
    ```
    参数 | 含义
    ---|---
    server-addr | 服务端公网ip及端口
    token | 验证密钥
    max-conn | 服务端与客户端通信最大连接数
    domain-as-proto | 使用绑定过`server-addr`的域名来访问
    tunnel | 本地解析的通道列表
    domain | 绑定过`server-addr`的域名
    proto | 内网代理的地址及端口


## 操作系统支持  
支持Windows、Linux、MacOSX等。

## 参考
https://github.com/cnlh/nps

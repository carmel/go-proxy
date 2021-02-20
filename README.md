# go-proxy
[中文版](./README_zh.md)

The intranet penetration implemented by golang is improved based on [IntranetPenetration](https://github.com/alongL/IntranetPenetration).


High-performance http proxy server, the main application and intranet penetration. Support multi-site configuration, automatic reconnection of client and server connection interruption, multiple transmission, greatly improve request processing speed, written in go language, no third-party dependency, tested memory footprint, only 10m memory in normal scenarios .


## Features

- [X] Support multi-site configuration
- [X] Automatically reconnect when disconnected
- [X] Support multiple transmission, improve concurrency
- [X] Cross-site automatic matching and replacement


## Compile and install
```sh
git clone github.com/carmel/go-proxy.git
cd go-proxy && make
```

## Instructions for use  
- Server
  ```
  ./server -token DKibZF5TXvic1g3kY -log proxy.log -tcpport=8284 -httpport=8024
  ```

  Parameters | Meaning
  ---|---
  log | log file path
  token | verification key
  tcpport | server-client communication port
  httpport | proxy http port

- Client
  ```sh
  ./client -conf conf.yml
  ```
  Parameters | Meaning
  ---|---
  conf | yaml configuration file path

  - yaml configuration example
    ```yml
    server-addr: 23.23.23.23:8284
    token: DKibZF5TXvic1g3kY
    max-conn: 1
    domain-as-proto: false
    tunnel:
      -domain: 23.23.23.23
        proto: 127.0.0.1:80
    ```
    Parameters | Meaning
    ---|---
    server-addr | Server public network ip and port
    token | verification key
    max-conn | The maximum number of connections between the server and the client
    domain-as-proto | Use the domain name bound with `server-addr` to access
    tunnel | Locally resolved channel list
    domain | The domain name bound to `server-addr`
    proto | Intranet proxy address and port


## Operating system support
Supports Windows, Linux, MacOSX, etc.

## Reference
https://github.com/cnlh/nps
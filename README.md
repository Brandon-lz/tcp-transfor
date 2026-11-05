# tcp-transfor
tcp transfor by golang
transfor local port to remote 

## warning
Just an open-source software, please use it legally.

## 查看在线客户端（server clients）

server 启动时会额外监听一个只读状态接口 `127.0.0.1:38081`，用 `clients` 子命令可以查看当前在线的客户端和端口映射情况。

```bash
./server clients          # 表格输出
./server clients -json    # 原始 JSON，方便 jq 处理
./server clients -addr 127.0.0.1:38081   # 指定状态接口地址
```

输出示例：

```
服务端口: 38080
在线客户端: 1

CLIENT     REMOTE           LISTEN  ->  LOCAL-TARGET
ubuntu223  1.2.3.4:51234    :9091   ->  127.0.0.1:9090
                            :30022  ->  10.0.0.5:22
```

- `CLIENT` / `REMOTE`：客户端名称（config.toml 里的 `name`，全局唯一）和它的主连接来源地址
- `LISTEN`：服务端为该客户端监听的端口（hello 消息里的 `server-port`）
- `LOCAL-TARGET`：客户端本地目标地址（`local-host:local-port`，`local-host` 留空即 127.0.0.1）
- 一个客户端有多条 `[[map]]` 时，第一个映射行显示名称，其余行缩进对齐

只显示已经完成 hello 握手、且还没被心跳判死的客户端；客户端掉线后会在心跳超时（`pongTimeout`，15s 量级）内从这里消失。

状态接口只绑定回环地址，不需要在云安全组放通。想在本地机器上看远程 server 的情况，用 SSH 转发即可：

```bash
ssh -L 38081:127.0.0.1:38081 user@server
# 另开一个终端，-addr 指定本地转发端口（给了 -addr 就不读 config.toml）
./server clients -addr 127.0.0.1:38081
```

接口端口可以在 `server/config.toml` 里改：

```toml
ui-port = 38081
```


# pushMessage
Go消息推送

### 实现功能
> 
- 对点消息推送
- 全覆盖消息推送
- 客户端离线保存消息
- 客户端重连后消息不丢失
- 服务端心跳
- 自定义启动端口

### api
|路由|请求方式|body|参数|说明|
|---:|---:|---:|---:|---:|
|/message|post| json(消息)|-|推送消息给全部客户端，包括离线未销毁客户端|
|/message/:id|post| json(消息)|:id 为client标记|推送消息给指定客户端, 包括离线未销毁客户端|
|/client/:id|websocket| -|:id 为client标记|websocket客户端连接|
|/client/:id|delete| -|:id 为client标记|销毁已经连接过的客户端|

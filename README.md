# quickim
一款超轻量级网页版即时聊天工具，包含服务端和客户端。跨平台，加密通讯，一键部署，开箱即用。

快速启动： 
export GOPROXY="https://goproxy.cn"
go mod tidy
go run ./
或者在release中下载可执行文件，直接运行即可

浏览器打开地址 http://localhost:5534

通讯加密原理：
1. 建立websocket连接
2. 服务端下发RSA公钥
3. 客户端随机生成对称加密密钥，并通过服务端下发的RSA将生成的密钥传给服务端
4. 服务端发出SHAKESUCC指令，表示协商密钥成功，后续进入加密通讯模式，使用DES ECB加密通讯

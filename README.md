# hzau校园网ip检测

这是一个基于 Go 开发的 HZAU 校园网 IP 检测后端服务，可用于判断访问用户是否处于 HZAU 校园网环境，并查询返回对应的用户 ID（通常为学号）。它可以作为校内资源分享网站的一种轻量级身份验证方案。

部分资源分享网站（如学科资料共享平台等）出于信任与安全考虑，不希望校外用户访问，仅允许连接校园网的用户使用，例如宿舍区网络及免费校园 WiFi 用户。

## 工作原理

服务通过运行一个 HTTP 接口，在用户访问时：

1. 解析请求信息，获取 X-Forwarded-For 中的客户端 IP；
2. 查询 QQWry IP 数据库，判断 IP 是否属于校园网地址段；
3. 调用校园网相关接口查询用户 ID；
4. 最终判断用户是否处于校园网环境，并返回对应身份信息。

受限：

- 服务必须部署在宿舍区，且拨号获得教育网ip
- 学校屏蔽校外入站，校外访问会超时

## 使用方法

1. git本项目
2. 下载纯真ip数据库

    运行下载脚本：`bash download_qqwry.sh` 下载最新ip数据库文件。
    
    或者手动下载 [https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat](https://github.com/metowolf/qqwry.dat/releases/latest/download/qqwry.dat)，复制到项目根目录。
    
3. 运行

    终端里运行 `go run hzau-ip-locator.go` ，检测是否正常输出日志。

4. 访问

    前端访问 http://yourip/api/ipcheck, 可直接查看返回的json。

gitflow工具流服务端

这是个git仓库服务端，提供必要的权限控制。


##注意
程序使用了Git的git-http-backend 后台服务 并且没有配置客户端认证。
 
在push的时候遇到这个错误: The requested URL returned error: 403..,查看apache后台提示错误：Service not enabled: 'receive-pack'。 

解决方法如下：

在git目录下执行下面的命令，以打开匿名情况下的http.receivepack服务。

git config --file config http.receivepack true


##依赖库
go get github.com/go-sql-driver/mysql

go get github.com/jteeuwen/go-bindata/go-bindata

go-bindata -debug res/
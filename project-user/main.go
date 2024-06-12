package main

import (
	"github.com/gin-gonic/gin"
	_ "test.com/project-api/api"
	srv "test.com/project-common"
	"test.com/project-user/config"
	"test.com/project-user/router"
)

func main() {
	r := gin.Default()

	config.C.InitZapLog()
	//路由
	router.InitRouter(r)
	//grpc服务注册
	gc := router.RegisterGrpc()
	//grpc注册到etcd
	router.RegisterEtcdServer()
	stop := func() {
		gc.Stop()
	}

	srv.Run(r, config.C.SC.Name, config.C.SC.Addr, stop)
}

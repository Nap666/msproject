package router

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// 定义一个接口
type Router interface {
	Route(r *gin.Engine)
}

// 写一个注册结构体
type RegisterRouter struct {
}

func (*RegisterRouter) Route(router Router, r *gin.Engine) {
	router.Route(r)
}

var routers []Router

func InitRouter(r *gin.Engine) {
	for _, ro := range routers {
		ro.Route(r)
	}
}

func Register(ro ...Router) {
	routers = append(routers, ro...)
}

type gRPCConfig struct {
	Addr         string
	RegisterFunc func(*grpc.Server)
}

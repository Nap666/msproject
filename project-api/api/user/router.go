package user

import (
	"github.com/gin-gonic/gin"
	"log"
	"test.com/project-api/api/rpc"
	"test.com/project-api/router"
)

// 实现Router接口（两个包中的都实现了）
type RouterUser struct {
}

func init() {
	log.Println("package_user:init user router")
	ru := &RouterUser{}
	router.Register(ru)
}
func (*RouterUser) Route(r *gin.Engine) {
	//初始化grpc客户端连接
	log.Println("初始化grpc客户端连接")
	rpc.InitGrpcUserClient()
	h := New()
	r.POST("/project/login/getCaptcha", h.getCaptcha)
	r.POST("/project/login/register", h.register)
	r.POST("/project/login", h.Login)
}

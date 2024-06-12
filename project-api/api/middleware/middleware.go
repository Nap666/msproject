package middleware

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"test.com/project-api/api/rpc"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/user/login"

	"time"
)

func TokenVeify() func(*gin.Context) {
	return func(c *gin.Context) {
		result := &common.Result{}
		//1 从header获取token
		token := c.GetHeader("Authorization")
		ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancelFunc()
		//2 grpc调用user中的认证方法
		response, err := rpc.LoginServiceClient.TokenVerify(ctx, &login.LoginMessage{Token: token})
		if err != nil {
			//解析错误
			code, msg := errs.ParseGrpcError(err)
			c.JSON(http.StatusOK, result.Fail(code, msg))
			c.Abort() //？？
			return
		}
		//3 认证通过-》处理结果 将信息放入gin的上下文 失败返回未登录
		//验证token时，把解析到的用户id放在上下文中，方便后面拿取
		c.Set("memberId", response.Member.Id)
		c.Next()
	}
}

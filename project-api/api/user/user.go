package user

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"test.com/project-api/api/rpc"
	"test.com/project-api/pkg/model/user"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/user/login"
	"time"
)

type HandlerUser struct {
}

func New() *HandlerUser {
	return &HandlerUser{}
}
func (*HandlerUser) getCaptcha(c *gin.Context) {
	result := &common.Result{}
	mobile := c.PostForm("mobile")
	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()
	rsp, err := rpc.LoginServiceClient.GetCaptcha(ctx, &login.CaptchaMessage{Mobile: mobile})

	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	c.JSON(http.StatusOK, result.Success(rsp.Code))
}

func (l *HandlerUser) register(c *gin.Context) {
	//*gin.Context.JSON()
	//1接收参数
	result := &common.Result{}
	var req user.RegisterReq
	//用于将请求中的数据绑定到指定的结构体实例 req 上
	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}

	//2 校验参数
	if err := req.Verify(); err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, err.Error()))
		return
	}

	//调用user grpc 服务
	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()
	msg := &login.RegisterMessage{}
	err = copier.Copy(msg, req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "拷贝login.msg有误"))
		return
	}
	_, err = rpc.LoginServiceClient.Register(ctx, msg)

	if err != nil {
		code, msg := errs.ParseGrpcError(err)

		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	//返回结果
	c.JSON(http.StatusOK, result.Success(""))

}

func (l *HandlerUser) Login(c *gin.Context) {
	//接收参数
	result := &common.Result{}
	var req user.LoginReq

	err := c.ShouldBind(&req)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "参数格式有误"))
		return
	}

	//调用user grpc服务
	ctx, cancelFunc := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancelFunc()
	var msg = &login.LoginMessage{}
	err = copier.Copy(msg, req)

	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "拷贝login.msg有误"))
		return
	}
	loginRsp, err := rpc.LoginServiceClient.Login(ctx, msg)

	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
		return
	}
	rsp := &user.LoginRsp{}
	err = copier.Copy(rsp, loginRsp)
	if err != nil {
		c.JSON(http.StatusOK, result.Fail(http.StatusBadRequest, "copy有误"))
		return
	}
	//返回结果
	c.JSON(http.StatusOK, result.Success(rsp))

}

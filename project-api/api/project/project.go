package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"net/http"
	"test.com/project-api/pkg/model"
	"test.com/project-api/pkg/model/pro"
	common "test.com/project-common"
	"test.com/project-common/errs"
	"test.com/project-grpc/project"
	"time"
)

type HandlerProject struct {
}

func (p *HandlerProject) index(c *gin.Context) {
	//设置参数
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	msg := &project.IndexMessage{}
	//grpc调用project-project模块的方法
	indexResponse, err := ProjectServiceClient.Index(ctx, msg)
	if err != nil {
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	//返回前端，第二个参数的响应码才是给前端的code,msg,data中的code
	c.JSON(http.StatusOK, result.Success(indexResponse.Menus))
}

func (p *HandlerProject) myProjectList(c *gin.Context) {
	result := &common.Result{}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	//这里msg是不是直接有值？不是的，要从上下文中获取
	msg := &project.ProjectRpcMessage{}
	//从上下文获取memberId,为什么是memberId? 答：在解析token时将用户id存在上下文中
	memberIdStr, _ := c.Get("memberId")
	msg.MemberId = memberIdStr.(int64)
	page := model.Page{}
	page.Bind(c)

	msg.Page = page.Page
	msg.PageSize = page.PageSize

	//grpc远程调用,调用后会返回rsp，
	myProjectResponse, err := ProjectServiceClient.FindProjectByMemId(ctx, msg)
	if err != nil {
		//解析err，返回一个前端接收到的code和前端接收到的msg
		code, msg := errs.ParseGrpcError(err)
		c.JSON(http.StatusOK, result.Fail(code, msg))
	}
	//搞不懂为什么要有这里？
	if myProjectResponse.Pm == nil {
		myProjectResponse.Pm = []*project.ProjectMessage{}
	}
	var pms []*pro.ProjectAndMember
	copier.Copy(&pms, myProjectResponse.Pm)
	c.JSON(http.StatusOK, result.Success(gin.H{
		"list":  pms,
		"total": myProjectResponse.Total,
	}))

}

func New() *HandlerProject {
	return &HandlerProject{}
}

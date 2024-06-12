package model

import (
	"test.com/project-common/errs"
)

var (
	RedisError       = errs.NewError(999, "Redis错误")      //
	DBError          = errs.NewError(998, "DB错误")         //
	NoLogin          = errs.NewError(997, "未登录")          //
	NoLegalMobile    = errs.NewError(10102001, "手机号码不合法") //手机号码不合法
	CaptchaError     = errs.NewError(10102002, "验证码错误")   //
	CaptchaNotExist  = errs.NewError(10102003, "验证码不存在")  //
	EmailExit        = errs.NewError(10102004, "邮箱已经存在")  //
	AccountExit      = errs.NewError(10102005, "账号已经存在")  //
	MobileExit       = errs.NewError(10102006, "手机号已经存在") //
	AccountAndPwdErr = errs.NewError(10102007, "账号密码不正确") //

)

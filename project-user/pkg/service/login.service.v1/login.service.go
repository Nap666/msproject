package login_service_v1

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/jinzhu/copier"
	"go.uber.org/zap"
	"strconv"
	"strings"
	common "test.com/project-common"
	"test.com/project-common/encrypts"
	"test.com/project-common/errs"
	"test.com/project-common/jwts"
	"test.com/project-grpc/user/login"
	"test.com/project-user/config"
	"test.com/project-user/internal/dao"
	"test.com/project-user/internal/data/member"
	"test.com/project-user/internal/data/organization"
	"test.com/project-user/internal/database"
	"test.com/project-user/internal/database/tran"

	"test.com/project-user/internal/repo"
	"test.com/project-user/pkg/model"
	"time"
)

type LoginService struct {
	login.UnimplementedLoginServiceServer
	Cache            repo.Cache
	MemberRepo       repo.MemberRepo
	OrganizationRepo repo.OrganizationRepo
	Transaction      tran.Transaction
}

func New() *LoginService {
	return &LoginService{
		Cache:            dao.Rc,
		MemberRepo:       dao.NewMemberDao(),
		OrganizationRepo: dao.NewOrganizationDao(),
		Transaction:      dao.NewTransaction(),
	}
}
func (ls *LoginService) GetCaptcha(ctx context.Context, msg *login.CaptchaMessage) (*login.CaptchaResponse, error) {
	//1.获取参数
	mobile := msg.Mobile
	//2.校验参数
	if !common.VerifyMobile(mobile) {
		return nil, errs.GrpcError(model.NoLegalMobile)
	}
	//3.生成验证码（随机4位1000-9999或者6位100000-999999）
	code := "123456"
	//4.调用短信平台（三方 放入go协程中执行 接口可以快速响应）
	go func() {
		time.Sleep(2 * time.Second)
		zap.L().Info("短信平台调用成功，发送短信")
		//redis 假设后续缓存可能存在mysql当中，也可能存在mongo当中 也可能存在memcache当中
		//5.存储验证码 redis当中 过期时间15分钟
		c, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := ls.Cache.Put(c, model.RegisterRedisKey+mobile, code, 15*time.Minute)
		if err != nil {
			zap.L().Info(fmt.Sprintf("验证码存入redis出错,cause by: %v \n", err))
		}
	}()
	return &login.CaptchaResponse{Code: code}, nil
}

func (ls *LoginService) Register(ctx context.Context, msg *login.RegisterMessage) (*login.RegisterResponse, error) {
	c := context.Background()
	//可以校验参数
	//校验验证码
	redisCode, err := ls.Cache.Get(c, model.RegisterRedisKey+msg.Mobile)
	if err == redis.Nil {
		return nil, errs.GrpcError(model.CaptchaNotExist)
	}
	if err != nil {
		zap.L().Error("Register redis get error", zap.Error(err))
		return nil, errs.GrpcError(model.RedisError)
	}
	if redisCode != msg.Captcha {
		return nil, errs.GrpcError(model.CaptchaError)
	}

	//3校验业务逻辑（邮箱、账号、手机号是否被注册）
	exit, err := ls.MemberRepo.GetMemberByEmail(c, msg.Email)
	if err != nil {
		zap.L().Error("Register DB get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)

	}
	if exit {
		zap.L().Error("Register Email  has exit", zap.Error(err))
		return nil, errs.GrpcError(model.EmailExit)
	}

	exit, err = ls.MemberRepo.GetMemberByAccount(c, msg.Name)
	if err != nil {
		zap.L().Error("Register DB get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)

	}
	if exit {
		zap.L().Error("Register Name has exit", zap.Error(err))
		return nil, errs.GrpcError(model.AccountExit)
	}

	exit, err = ls.MemberRepo.GetMemberByMobile(c, msg.Mobile)
	if err != nil {
		zap.L().Error("Register DB get error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)

	}
	if exit {
		zap.L().Error("Register Mobile  has exit", zap.Error(err))
		return nil, errs.GrpcError(model.MobileExit)
	}
	//4 执行业务、数据存入member表，生成数据存入组织表
	//前端已经加密一次了。。。
	pwd := encrypts.Md5(msg.Password)
	mem := &member.Member{
		Account:       msg.Name,
		Password:      pwd,
		Name:          msg.Name,
		Mobile:        msg.Mobile,
		Email:         msg.Email,
		CreateTime:    time.Now().UnixMilli(),
		LastLoginTime: time.Now().UnixMilli(),
		Status:        model.Normal,
	}
	//优雅的事务，保证member表和organization表插入的原子性
	err = ls.Transaction.Action(func(conn database.DbConn) error {
		//存入member表
		err = ls.MemberRepo.SaveMember(conn, c, mem)
		if err != nil {
			zap.L().Error("register db SaveMember  err", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		//存入orgnization
		org := &organization.Organization{
			Name:       mem.Name + "个人组织",
			MemberId:   mem.Id,
			CreateTime: time.Now().UnixMilli(),
			Personal:   model.Personal,
			Avatar:     "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Fc-ssl.dtstatic.com%2Fuploads%2Fblog%2F202103%2F31%2F20210331160001_9a852.thumb.1000_0.jpg&refer=http%3A%2F%2Fc-ssl.dtstatic.com&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1673017724&t=ced22fc74624e6940fd6a89a21d30cc5",
		}
		err = ls.OrganizationRepo.SaveOrganization(c, org)
		if err != nil {
			zap.L().Error("register SaveOrganization db err", zap.Error(err))
			return errs.GrpcError(model.DBError)
		}
		return nil
	})

	//5 返回
	return &login.RegisterResponse{}, err
}

func (ls *LoginService) Login(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	c := context.Background()
	//1.去数据库查询 账号密码是否正确

	pwd := encrypts.Md5(msg.Password)
	mem, err := ls.MemberRepo.FindMember(c, msg.Account, pwd)
	if err != nil {
		zap.L().Error("Login db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	if mem == nil {
		return nil, errs.GrpcError(model.AccountAndPwdErr)
	}
	memMsg := &login.MemberMessage{}
	err = copier.Copy(memMsg, mem)
	//AES加密处理
	memMsg.Code, _ = encrypts.EncryptInt64(mem.Id, model.AEXKEY)
	//2.根据用户id查组织
	orgs, err := ls.OrganizationRepo.FindOrganizationByMemId(c, mem.Id)
	if err != nil {
		zap.L().Error("Login db FindMember error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	var orgsMessage []*login.OrganizationMessage
	err = copier.Copy(&orgsMessage, orgs)
	for _, v := range orgsMessage {
		v.Code, _ = encrypts.EncryptInt64(mem.Id, model.AEXKEY)
	}
	//3.用jwt生成token
	memIdStr := strconv.Itoa(int(mem.Id))
	//以分钟为单位
	exp := time.Duration(config.C.JwtConfig.AccessExp*3600*24) * time.Second
	rExp := time.Duration(config.C.JwtConfig.RefreshExp*3600*24) * time.Second
	token := jwts.CreateToken(memIdStr, exp, config.C.JwtConfig.AccessSecret, rExp, config.C.JwtConfig.RefreshSecret)
	tokenList := &login.TokenMessage{
		AccessToken:    token.AccessToken,
		RefreshToken:   token.RefreshToken,
		AccessTokenExp: token.AccessExp,
		TokenType:      "bearer",
	}
	return &login.LoginResponse{
		Member:           memMsg,
		OrganizationList: orgsMessage,
		TokenList:        tokenList,
	}, nil
}

func (ls *LoginService) TokenVerify(ctx context.Context, msg *login.LoginMessage) (*login.LoginResponse, error) {
	token := msg.Token
	if strings.Contains(token, "bearer") {
		token = strings.ReplaceAll(token, "bearer ", "")
	}
	//解析后返回用户id（string类型）
	parseToken, err := jwts.ParseToken(token, config.C.JwtConfig.AccessSecret)
	if err != nil {
		zap.L().Error("Login TokenVerify err", zap.Error(err))
		return nil, errs.GrpcError(model.NoLogin)
	}
	//todo (可优化)数据库查询 优化点 登录之后 应该把用户信息缓存起来
	id, _ := strconv.ParseInt(parseToken, 10, 64)
	memberById, err := ls.MemberRepo.FindMemberById(context.Background(), id)
	if err != nil {
		zap.L().Error("TokenVerify db FindMemberById error", zap.Error(err))
		return nil, errs.GrpcError(model.DBError)
	}
	memMsg := &login.MemberMessage{}
	copier.Copy(memMsg, memberById)
	memMsg.Code, _ = encrypts.EncryptInt64(memberById.Id, model.AEXKEY)
	return &login.LoginResponse{Member: memMsg}, nil

}

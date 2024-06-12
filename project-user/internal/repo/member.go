package repo

import (
	"context"
	"test.com/project-user/internal/data/member"
	"test.com/project-user/internal/database"
)

type MemberRepo interface {
	GetMemberByEmail(ctx context.Context, email string) (bool, error)
	GetMemberByAccount(ctx context.Context, Name string) (bool, error)
	GetMemberByMobile(ctx context.Context, Mobile string) (bool, error)
	SaveMember(conn database.DbConn, ctx context.Context, member *member.Member) error
	FindMember(ctx context.Context, account string, pwd string) (member *member.Member, err error)
	FindMemberById(background context.Context, id int64) (mem *member.Member, err error)
}

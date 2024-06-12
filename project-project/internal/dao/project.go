package dao

import (
	"context"
	"test.com/project-project/internal/data/pro"
	"test.com/project-project/internal/database/gorms"
)

type ProjectDao struct {
	conn *gorms.GormConn
}

// 传来用户id page页号 size页面大小
// limt 0,10  从第0行开始返回10行数据（0-9行）
func (p ProjectDao) FindProjectByMemId(ctx context.Context, memId int64, page int64, size int64) ([]*pro.ProjectAndMember, int64, error) {
	var pms []*pro.ProjectAndMember
	session := p.conn.Session(ctx)
	index := (page - 1) * size
	raw := session.Raw("select * from ms_project a, ms_project_member b where a.id = b.project_code and b.member_code=? limit ?,?", memId, index, size)
	raw.Scan(&pms)
	var total int64
	err := session.Model(&pro.ProjectMember{}).Where("member_code=?", memId).Count(&total).Error
	return pms, total, err
}

func NewProjectDao() *ProjectDao {
	return &ProjectDao{
		conn: gorms.New(),
	}
}

package dao

import (
	"context"
	"test.com/project-user/internal/data/organization"

	"test.com/project-user/internal/database/gorms"
)

type OrganizationDao struct {
	conn *gorms.GormConn
}

func (o *OrganizationDao) FindOrganizationByMemId(ctx context.Context, memId int64) ([]*organization.Organization, error) {
	orgs := []*organization.Organization{}
	err := o.conn.Session(ctx).Where("member_id = ?  ", memId).Find(&orgs).Error
	return orgs, err
}

func NewOrganizationDao() *OrganizationDao {
	return &OrganizationDao{
		conn: gorms.New(),
	}
}

func (o *OrganizationDao) SaveOrganization(ctx context.Context, org *organization.Organization) error {
	err := o.conn.Session(ctx).Create(org).Error
	return err
}

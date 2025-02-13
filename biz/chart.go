package biz

import (
	"context"

	"github.com/cuigh/auxo/net/web"
	"github.com/cuigh/swirl/dao"
)

type ChartBiz interface {
	Search(args *dao.ChartSearchArgs) (charts []*dao.Chart, total int, err error)
	Delete(id, title string, user web.User) (err error)
	Find(id string) (chart *dao.Chart, err error)
	Batch(ids ...string) (charts []*dao.Chart, err error)
	Create(chart *dao.Chart, user web.User) (err error)
	Update(chart *dao.Chart, user web.User) (err error)
}

func NewChart(d dao.Interface, mb MetricBiz, eb EventBiz) ChartBiz {
	return &chartBiz{
		d:  d,
		mb: mb,
		eb: eb,
	}
}

type chartBiz struct {
	d  dao.Interface
	mb MetricBiz
	eb EventBiz
}

func (b *chartBiz) Search(args *dao.ChartSearchArgs) (charts []*dao.Chart, total int, err error) {
	return b.d.ChartSearch(context.TODO(), args)
}

func (b *chartBiz) Create(chart *dao.Chart, user web.User) (err error) {
	chart.ID = createId()
	chart.CreatedAt = now()
	chart.CreatedBy = newOperator(user)
	chart.UpdatedAt = chart.CreatedAt
	chart.UpdatedBy = chart.CreatedBy
	err = b.d.ChartCreate(context.TODO(), chart)
	if err == nil {
		b.eb.CreateChart(EventActionCreate, chart.ID, chart.Title, user)
	}
	return
}

func (b *chartBiz) Delete(id, title string, user web.User) (err error) {
	err = b.d.ChartDelete(context.TODO(), id)
	if err == nil {
		b.eb.CreateChart(EventActionDelete, id, title, user)
	}
	return
}

func (b *chartBiz) Find(id string) (chart *dao.Chart, err error) {
	return b.d.ChartGet(context.TODO(), id)
}

func (b *chartBiz) Batch(ids ...string) (charts []*dao.Chart, err error) {
	charts, err = b.d.ChartGetBatch(context.TODO(), ids...)
	return
}

func (b *chartBiz) Update(chart *dao.Chart, user web.User) (err error) {
	chart.UpdatedAt = now()
	chart.UpdatedBy = newOperator(user)
	err = b.d.ChartUpdate(context.TODO(), chart)
	if err == nil {
		b.eb.CreateChart(EventActionUpdate, chart.ID, chart.Title, user)
	}
	return
}

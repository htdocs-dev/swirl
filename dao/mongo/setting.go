package mongo

import (
	"context"

	"github.com/cuigh/swirl/model"
	"go.mongodb.org/mongo-driver/bson"
)

const Setting = "setting"

func (d *Dao) SettingGetAll(ctx context.Context) (settings []*model.Setting, err error) {
	settings = []*model.Setting{}
	err = d.fetch(ctx, Setting, bson.M{}, &settings)
	return
}

func (d *Dao) SettingGet(ctx context.Context, id string) (setting *model.Setting, err error) {
	setting = &model.Setting{}
	found, err := d.find(ctx, Setting, id, setting)
	if !found {
		return nil, err
	}
	return
}

func (d *Dao) SettingUpdate(ctx context.Context, setting *model.Setting) (err error) {
	update := bson.M{
		"$set": bson.M{
			"options":    setting.Options,
			"updated_at": setting.UpdatedAt,
			"updated_by": setting.UpdatedBy,
		},
	}
	return d.upsert(ctx, Setting, setting.ID, update)
}

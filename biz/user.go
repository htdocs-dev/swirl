package biz

import (
	"context"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/net/web"
	"github.com/cuigh/auxo/security/passwd"
	"github.com/cuigh/swirl/dao"
	"github.com/cuigh/swirl/misc"
)

const (
	// UserTypeInternal is internal user of swirl
	UserTypeInternal = "internal"
	// UserTypeLDAP is external user of LDAP
	UserTypeLDAP = "ldap"
)

const (
	// UserStatusBlocked is the status which user is blocked
	UserStatusBlocked = 0
	// UserStatusActive is the normal status
	UserStatusActive = 1
)

type UserBiz interface {
	Search(name, loginName, filter string, pageIndex, pageSize int) (users []*dao.User, total int, err error)
	Create(user *dao.User, ctxUser web.User) (id string, err error)
	Update(user *dao.User, ctxUser web.User) (err error)
	FindByID(id string) (user *dao.User, err error)
	FindByName(loginName string) (user *dao.User, err error)
	FindByToken(token string) (user *dao.User, err error)
	FindPrivacy(loginName string) (privacy *UserPrivacy, err error)
	Count() (count int, err error)
	Delete(id, name string, user web.User) (err error)
	SetStatus(id string, status int32, user web.User) (err error)
	ModifyPassword(oldPwd, newPwd string, user web.User) (err error)
	ModifyProfile(user *dao.User, ctxUser web.User) (err error)
}

func NewUser(d dao.Interface, eb EventBiz) UserBiz {
	return &userBiz{d: d, eb: eb}
}

type userBiz struct {
	d  dao.Interface
	eb EventBiz
}

func (b *userBiz) Search(name, loginName, filter string, pageIndex, pageSize int) (users []*dao.User, total int, err error) {
	var args = &dao.UserSearchArgs{
		Name:      name,
		LoginName: loginName,
		Status:    -1,
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}

	switch filter {
	case "admins":
		args.Admin = true
	case "active":
		args.Status = UserStatusActive
	case "blocked":
		args.Status = UserStatusBlocked
	}

	return b.d.UserSearch(context.TODO(), args)
}

func (b *userBiz) FindByID(id string) (user *dao.User, err error) {
	return b.d.UserGet(context.TODO(), id)
}

func (b *userBiz) FindByName(loginName string) (user *dao.User, err error) {
	return b.d.UserGetByName(context.TODO(), loginName)
}

func (b *userBiz) FindByToken(token string) (user *dao.User, err error) {
	return b.d.UserGetByToken(context.TODO(), token)
}

func (b *userBiz) FindPrivacy(loginName string) (privacy *UserPrivacy, err error) {
	var u *dao.User
	u, err = b.d.UserGetByName(context.TODO(), loginName)
	if u != nil {
		privacy = &UserPrivacy{
			ID:       u.ID,
			Name:     u.Name,
			Password: u.Password,
			Salt:     u.Salt,
			Type:     u.Type,
			Status:   u.Status,
		}
	}
	return
}

func (b *userBiz) Create(user *dao.User, ctxUser web.User) (id string, err error) {
	user.Tokens = data.Options{data.Option{Name: "test", Value: "abc123"}}
	user.ID = createId()
	user.Status = UserStatusActive
	user.CreatedAt = now()
	if ctxUser != nil {
		user.CreatedBy = newOperator(ctxUser)
	}
	user.UpdatedAt = user.CreatedAt
	user.UpdatedBy = user.CreatedBy
	if user.Type == UserTypeInternal {
		user.Password, user.Salt, err = passwd.Generate(user.Password)
		if err != nil {
			return
		}
	}

	if err = b.d.UserCreate(context.TODO(), user); err == nil && ctxUser != nil {
		b.eb.CreateUser(EventActionCreate, user.LoginName, user.Name, ctxUser)
	}
	id = user.ID
	return
}

func (b *userBiz) Update(user *dao.User, ctxUser web.User) (err error) {
	user.UpdatedAt = now()
	user.UpdatedBy = newOperator(ctxUser)
	if err = b.d.UserUpdate(context.TODO(), user); err == nil {
		b.eb.CreateUser(EventActionUpdate, user.LoginName, user.Name, ctxUser)
	}
	return
}

func (b *userBiz) SetStatus(id string, status int32, user web.User) (err error) {
	u := &dao.User{
		ID:        id,
		Status:    status,
		UpdatedAt: now(),
		UpdatedBy: newOperator(user),
	}
	return b.d.UserUpdateStatus(context.TODO(), u)
}

func (b *userBiz) Delete(id, name string, user web.User) (err error) {
	err = b.d.UserDelete(context.TODO(), id)
	if err == nil {
		b.eb.CreateUser(EventActionDelete, id, name, user)
	}
	return
}

func (b *userBiz) ModifyPassword(oldPwd, newPwd string, user web.User) (err error) {
	var u *dao.User
	u, err = b.d.UserGet(context.TODO(), user.ID())
	if err != nil {
		return err
	} else if u == nil {
		return errors.Format("user not found: %s", user.ID())
	}

	if !passwd.Validate(oldPwd, u.Password, u.Salt) {
		return errors.Coded(misc.ErrOldPasswordIncorrect, "current password is incorrect")
	}

	if u.Password, u.Salt, err = passwd.Generate(newPwd); err != nil {
		return
	}

	u.UpdatedAt = now()
	u.UpdatedBy = newOperator(user)
	return b.d.UserUpdatePassword(context.TODO(), u)
}

func (b *userBiz) ModifyProfile(u *dao.User, user web.User) (err error) {
	u.ID = user.ID()
	u.UpdatedAt = now()
	u.UpdatedBy = newOperator(user)
	return b.d.UserUpdateProfile(context.TODO(), u)
}

func (b *userBiz) Count() (count int, err error) {
	return b.d.UserCount(context.TODO())
}

type UserPrivacy struct {
	ID       string
	Name     string
	Password string `json:"-"`
	Salt     string `json:"-"`
	Type     string
	Status   int32
}

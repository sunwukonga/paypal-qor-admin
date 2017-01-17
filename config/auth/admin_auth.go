package auth

import (
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/qor/roles"
	"github.com/sunwukonga/paypal-qor-admin/app/models"
	"net/http"
)

func init() {
	roles.Register(models.RoleAdmin, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleAdmin
	})
	roles.Register(models.RoleCustomer, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleCustomer
	})
	roles.Register(models.RoleReseller, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleReseller
	})
	roles.Register(models.RoleInfluencer, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleInfluencer
	})
	roles.Register(models.RoleEditor, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleEditor
	})
	roles.Register(models.RoleServicer, func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleServicer
	})
}

type AdminAuth struct {
}

func (AdminAuth) LoginURL(c *admin.Context) string {
	return "/auth/login"
}

func (AdminAuth) LogoutURL(c *admin.Context) string {
	return "/auth/logout"
}

func (AdminAuth) GetCurrentUser(c *admin.Context) qor.CurrentUser {
	userInter, err := Auth.CurrentUser(c.Writer, c.Request)
	if userInter != nil && err == nil {
		if userInter.(*models.User).Role == models.RoleAdmin {
			return userInter.(*models.User)
		}
	}
	return nil
}

package auth

import (
	"github.com/qor/admin"
	"github.com/qor/qor"
	"github.com/sunwukonga/qor-example/app/models"
	"github.com/qor/roles"
	"net/http"
)

func init() {
	roles.Register("admin", func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleAdmin
	})
	roles.Register("customer", func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleCustomer
	})
	roles.Register("reseller", func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleReseller
	})
	roles.Register("influencer", func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleInfluencer
	})
	roles.Register("editor", func(req *http.Request, currentUser interface{}) bool {
		return currentUser != nil && currentUser.(*models.User).Role == models.RoleEditor
	})
	roles.Register("servicer", func(req *http.Request, currentUser interface{}) bool {
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
		return userInter.(*models.User)
	}
	return nil
}

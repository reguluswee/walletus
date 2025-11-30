package service

import "github.com/reguluswee/walletus/common/model"

func IsSuperAdmin(portalUser *model.PortalUser) bool {
	return portalUser.Type == "super_admin"
}

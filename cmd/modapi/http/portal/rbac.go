package portal

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func PortalFuncList(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()
	var portalFuncs []model.PortalFunc
	db.Where("flag = ? and type IN ('menu', 'button')", 0).Find(&portalFuncs)
	res.Data = gin.H{
		"portal_funcs": portalFuncs,
	}

	c.JSON(http.StatusOK, res)
}

func PortalRoleList(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()
	var portalRoles []model.PortalRole
	db.Where("flag = ?", 0).Find(&portalRoles)
	res.Data = gin.H{
		"portal_roles": portalRoles,
	}

	c.JSON(http.StatusOK, res)
}

func PortalRoleCreate(c *gin.Context) {
	var request request.PortalRoleCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	db := system.GetDb()
	var portalRole model.PortalRole = model.PortalRole{
		Name: request.Name,
		Flag: 0,
	}
	db.Create(&portalRole)

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = portalRole

	c.JSON(http.StatusOK, res)
}

func PortalRoleUpdate(c *gin.Context) {
	var request request.PortalRoleCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	db := system.GetDb()

	var portalRole model.PortalRole
	db.Where("id = ?", request.ID).First(&portalRole)
	if portalRole.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "role not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	portalRole.Name = request.Name
	db.Save(&portalRole)

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = portalRole

	c.JSON(http.StatusOK, res)
}

func PortalRoleDelete(c *gin.Context) {
	var request request.PortalRoleCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request",
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	db := system.GetDb()

	var portalRole model.PortalRole
	db.Where("id = ?", request.ID).First(&portalRole)
	if portalRole.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "role not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	portalRole.Flag = 1
	db.Save(&portalRole)

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	res.Data = portalRole

	c.JSON(http.StatusOK, res)
}

func PortalRoleFuncList(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()
	var portalRole model.PortalRole
	db.Where("id = ?", c.Param("role_id")).First(&portalRole)

	if portalRole.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "role not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	var portalFuncs []model.PortalFunc
	db.Table("admin_portal_function f").
		Joins("JOIN admin_portal_role_func rf ON f.id = rf.func_id").
		Where("rf.role_id = ?", portalRole.ID).Find(&portalFuncs)

	res.Data = gin.H{
		"portal_funcs": portalFuncs,
	}

	c.JSON(http.StatusOK, res)
}

func PortalRoleUserList(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()
	var portalRole model.PortalRole
	db.Where("id = ?", c.Param("role_id")).First(&portalRole)

	if portalRole.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "role not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	var portalFuncs []model.PortalUser
	db.Table("admin_portal_user f").
		Joins("JOIN admin_portal_user_role rf ON f.id = rf.user_id").
		Where("rf.role_id = ?", portalRole.ID).Find(&portalFuncs)

	res.Data = gin.H{
		"portal_users": portalFuncs,
	}

	c.JSON(http.StatusOK, res)
}

func PortalRoleFuncBind(c *gin.Context) {
	res := common.Response{
		Timestamp: time.Now().Unix(),
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
	}

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil || portalUser.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	roleIDStr := c.Param("role_id")
	funcIDStr := c.Param("func_id")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil || roleID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid role_id"
		c.JSON(http.StatusOK, res)
		return
	}
	funcID, err := strconv.ParseUint(funcIDStr, 10, 64)
	if err != nil || funcID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid func_id"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var portalRole model.PortalRole
	if err := db.First(&portalRole, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_SECURITY
			res.Msg = "role not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var portalFunc model.PortalFunc
	if err := db.First(&portalFunc, funcID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_SECURITY
			res.Msg = "func not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	bind := model.PortalRoleFunc{
		RoleID: roleID,
		FuncID: funcID,
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "role_id"}, {Name: "func_id"}},
		DoNothing: true, // keep idempotent
	}).Create(&bind).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "bind role & func failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

func PortalRoleUserBind(c *gin.Context) {
	res := common.Response{
		Timestamp: time.Now().Unix(),
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
	}

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	loginUser, ok := mainUser.(*model.PortalUser)
	if !ok || loginUser == nil || loginUser.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	roleIDStr := c.Param("role_id")
	userIDStr := c.Param("user_id")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil || roleID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid role_id"
		c.JSON(http.StatusOK, res)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid user_id"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var role model.PortalRole
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "role not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var bindUser model.PortalUser
	if err := db.First(&bindUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "user not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	userRole := model.PortalUserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "role_id"}},
		DoNothing: true, // keep idempotent
	}).Create(&userRole).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "bind user & role failed"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

func PortalRoleFuncUnbind(c *gin.Context) {
	res := common.Response{
		Timestamp: time.Now().Unix(),
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
	}

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil || portalUser.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	roleIDStr := c.Param("role_id")
	funcIDStr := c.Param("func_id")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil || roleID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid role_id"
		c.JSON(http.StatusOK, res)
		return
	}
	funcID, err := strconv.ParseUint(funcIDStr, 10, 64)
	if err != nil || funcID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid func_id"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var portalRole model.PortalRole
	if err := db.First(&portalRole, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "role not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var portalFunc model.PortalFunc
	if err := db.First(&portalFunc, funcID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "func not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var rf model.PortalRoleFunc
	tx := db.Where("role_id = ? AND func_id = ?", roleID, funcID).Delete(&rf)
	if tx.Error != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "unbind role & func failed"
		c.JSON(http.StatusOK, res)
		return
	}
	if tx.RowsAffected == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "role & func not bound"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

func PortalRoleUserUnbind(c *gin.Context) {
	res := common.Response{
		Timestamp: time.Now().Unix(),
		Code:      codes.CODE_SUCCESS,
		Msg:       "success",
	}

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	loginUser, ok := mainUser.(*model.PortalUser)
	if !ok || loginUser == nil || loginUser.ID == 0 {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	roleIDStr := c.Param("role_id")
	userIDStr := c.Param("user_id")

	roleID, err := strconv.ParseUint(roleIDStr, 10, 64)
	if err != nil || roleID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid role_id"
		c.JSON(http.StatusOK, res)
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid user_id"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var role model.PortalRole
	if err := db.First(&role, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "role not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var bindUser model.PortalUser
	if err := db.First(&bindUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
			res.Msg = "user not existing"
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = "db error"
		}
		c.JSON(http.StatusOK, res)
		return
	}

	var ur model.PortalUserRole
	tx := db.Where("user_id = ? AND role_id = ?", userID, roleID).Delete(&ur)
	if tx.Error != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "unbind user & role failed"
		c.JSON(http.StatusOK, res)
		return
	}
	if tx.RowsAffected == 0 {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "user & role not bound"
		c.JSON(http.StatusOK, res)
		return
	}

	_ = loginUser

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

func PortalUserMenus(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	mainUser, ok := c.Get("main_user")
	if !ok {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}
	portalUser, ok := mainUser.(*model.PortalUser)
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var db = system.GetDb()

	var portalFuncs []model.PortalFunc
	err := db.Table("admin_portal_function f").
		Joins("JOIN admin_portal_role_func rf ON f.id = rf.func_id").
		Joins("JOIN admin_portal_user_role ur ON rf.role_id = ur.role_id").
		Where("ur.user_id = ? and f.type = 'menu'", portalUser.ID).
		Distinct().
		Find(&portalFuncs).Error
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "db error"
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"portal_funcs": portalFuncs,
	}

	c.JSON(http.StatusOK, res)
}

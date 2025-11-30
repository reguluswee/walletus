package portal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/cmd/modapi/security"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
)

const UserLoginTypeMain = 0
const UserLoginTypeWallet = 1

func PortalLogin(c *gin.Context) {
	var request request.PortalLoginRequest
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

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	var db = system.GetDb()
	var portalUsers []model.PortalUser
	db.Where("(login_id = ? or email = ?) and password = ?", request.LoginID, request.LoginID, request.Password).Find(&portalUsers)
	if len(portalUsers) == 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "portal user not existing or password not matching"
		c.JSON(http.StatusOK, res)
		return
	}
	if len(portalUsers) > 1 {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "portal user found more than one"
		c.JSON(http.StatusOK, res)
		return
	}

	portalUser := portalUsers[0]

	expireTs := time.Now().Add(common.TOKEN_DURATION).Unix()

	tokenOrig := fmt.Sprintf("%d|%d|%d", portalUser.ID, UserLoginTypeMain, expireTs)
	tokenEnc, err := security.Encrypt([]byte(tokenOrig))
	if err != nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "token gen error:" + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"token": tokenEnc,
		"user": gin.H{
			"id":       portalUser.ID,
			"login_id": portalUser.LoginID,
			"email":    portalUser.Email,
			"name":     portalUser.Name,
		},
	}

	c.JSON(http.StatusOK, res)
}

func PortalDashboard(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	c.JSON(http.StatusOK, res)
}

func PortalDeptList(c *gin.Context) {
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
	var portalDepts []model.PortalDept
	db.Where("flag = ?", 0).Find(&portalDepts)
	res.Data = gin.H{
		"portal_depts": portalDepts,
	}

	c.JSON(http.StatusOK, res)
}

func PortalDeptCreate(c *gin.Context) {
	var request request.PortalDeptCreateRequest
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

	var db = system.GetDb()
	var portalDept model.PortalDept
	db.Where("name = ?", request.Name).Find(&portalDept)
	if portalDept.ID != 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "portal dept already existing"
		c.JSON(http.StatusOK, res)
		return
	}

	portalDept.Name = request.Name
	portalDept.Desc = request.Desc
	portalDept.AddTime = time.Now()
	portalDept.Flag = 0
	portalDept.Status = "active"
	err := db.Create(&portalDept).Error
	if err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "portal dept create error:" + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	c.JSON(http.StatusOK, res)
}

func PortalDeptUpdate(c *gin.Context) {
	var request request.PortalDeptUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	var db = system.GetDb()
	var portalDept model.PortalDept
	db.Where("id = ?", request.ID).Find(&portalDept)
	if portalDept.ID == 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "portal dept not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	portalDept.Name = request.Name
	portalDept.Desc = request.Desc
	db.Save(&portalDept)

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	c.JSON(http.StatusOK, res)
}

func PortalDeptDelete(c *gin.Context) {
	var request struct {
		ID uint64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	var db = system.GetDb()
	var portalDept model.PortalDept
	db.Where("id = ?", request.ID).Find(&portalDept)
	if portalDept.ID == 0 {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "portal dept not existing"
		c.JSON(http.StatusOK, res)
		return
	}

	portalDept.Flag = 1
	portalDept.Status = "deleted"
	db.Save(&portalDept)

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	c.JSON(http.StatusOK, res)
}

type DeptItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type UserWithDept struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	LoginID  string `json:"login_id"`
	Location string `json:"location"`
	DeptList string `json:"dept_list"`
}

func PortalUserList(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"

	var db = system.GetDb()

	var users []UserWithDept

	err := db.
		Table("admin_portal_user u").
		Select(`
        u.id,
        u.name,
        u.email,
        u.login_id,
        u.location,
        IFNULL(
            (
                SELECT JSON_ARRAYAGG(
                    JSON_OBJECT(
                        'id', d.id,
                        'name', d.name
                    )
                )
                FROM admin_portal_dept_user du
                JOIN admin_portal_dept d ON d.id = du.dept_id
                WHERE du.user_id = u.id
            ),
            JSON_ARRAY()
        ) AS dept_list
    `).
		Scan(&users).Error

	if err != nil {
		log.Error("PortalUserList error: ", err)
	}

	for i, u := range users {
		var deptItems []DeptItem
		// We are just validating it's valid JSON, but sending the string to frontend
		// Alternatively, we could change the struct to use interface{} for DeptList if we wanted to return object
		json.Unmarshal([]byte(u.DeptList), &deptItems)
		// If we want to return parsed objects, we need a different struct.
		// For now, let's just return what we have, but make sure to add it to response.
		_ = i
	}

	res.Data = gin.H{
		"users": users,
	}

	c.JSON(http.StatusOK, res)
}

func PortalUserUpdate(c *gin.Context) {
	var req request.PortalUserUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Data:      nil,
			Timestamp: time.Now().Unix(),
		})
		return
	}
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	db := system.GetDb()
	tx := db.Begin()

	var user model.PortalUser
	if err := tx.First(&user, req.ID).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "user not found"
		c.JSON(http.StatusOK, res)
		return
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Location = req.Location
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to update user: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	// Update Departments
	// 1. Delete existing associations
	if err := tx.Where("user_id = ?", user.ID).Delete(&model.PortalDeptUser{}).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to clear old departments: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	// 2. Insert new associations
	for _, deptID := range req.DeptIDs {
		deptUser := model.PortalDeptUser{
			UserID: user.ID,
			DeptID: deptID,
		}
		if err := tx.Create(&deptUser).Error; err != nil {
			tx.Rollback()
			res.Code = codes.CODE_ERR_STATUS_GENERAL
			res.Msg = "failed to add department: " + err.Error()
			c.JSON(http.StatusOK, res)
			return
		}
	}

	tx.Commit()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

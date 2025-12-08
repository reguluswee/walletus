package portal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	ethutil "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
	"github.com/reguluswee/walletus/cmd/modapi/codes"
	"github.com/reguluswee/walletus/cmd/modapi/common"
	"github.com/reguluswee/walletus/cmd/modapi/request"
	"github.com/reguluswee/walletus/cmd/modapi/security"
	"github.com/reguluswee/walletus/common/model"
	"github.com/reguluswee/walletus/common/system"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

const UserLoginTypeMain = 0
const UserLoginTypeWallet = 1

const PayrollStatusCreate = "create"
const PayrollStatusWaitingApproval = "waiting_approval"
const PayrollStatusApproved = "approved"
const PayrollStatusRejected = "rejected"
const PayrollStatusPaid = "paid"
const PayrollStatusPaying = "paying"

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

	var db = system.GetDb()
	var countPortalUsers, countPortalDepts int64
	var countPendingPayrolls int64
	var countAmountForPendingPayrolls decimal.Decimal

	var payrollsPending []model.PortalPayroll
	db.Model(&model.PortalUser{}).
		Where("flag = ?", 0).
		Count(&countPortalUsers)
	db.Model(&model.PortalDept{}).
		Where("flag = ?", 0).
		Count(&countPortalDepts)
	db.Model(&model.PortalPayroll{}).
		Where("flag = ? and status = 'waiting_approval'", 0).
		Find(&payrollsPending)

	for _, payroll := range payrollsPending {
		countPendingPayrolls++
		countAmountForPendingPayrolls = payroll.TotalAmount.Add(payroll.TotalAmount)
	}

	res.Data = gin.H{
		"count_portal_users": countPortalUsers,
		"count_portal_depts": countPortalDepts,
		"count_payrolls": gin.H{
			"count":  countPendingPayrolls,
			"amount": countAmountForPendingPayrolls,
		},
	}

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
		Where("flag = ?", 0).
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

func PortalUserAdd(c *gin.Context) {
	var req request.PortalUserAddRequest
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
	user.Name = req.Name
	user.Email = req.Email
	user.Location = req.Location
	user.LoginID = req.LoginID
	user.Password = req.Password
	user.AddTime = time.Now()
	user.Type = "common"
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to update user: " + err.Error()
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

func PortalUserDelete(c *gin.Context) {
	res := common.Response{}
	res.Timestamp = time.Now().Unix()

	db := system.GetDb()
	tx := db.Begin()

	var user model.PortalUser
	if err := tx.First(&user, c.Param("user_id")).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to find user: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}
	if user.Type == "super_admin" {
		tx.Rollback()
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "super admin cannot be deleted"
		c.JSON(http.StatusOK, res)
		return
	}
	if err := tx.Where("user_id = ?", user.ID).Delete(&model.PortalDeptUser{}).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to clear old departments: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}
	user.Flag = 1
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to delete user: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	tx.Commit()

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

func PortalPayrollList(c *gin.Context) {
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

	status := c.Query("status")

	var db = system.GetDb()
	var payrollList []model.PortalPayroll
	query := db.Where("flag = ?", 0)
	if len(status) > 0 {
		query = query.Where("status = ?", status)
	}
	query.Order("roll_month DESC").Find(&payrollList)

	res.Data = gin.H{
		"payroll_list": payrollList,
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollCreate(c *gin.Context) {
	var request request.PortalPayrollCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	if !isValidYearMonth(request.RollMonth) {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "payroll month must be in format YYYY-MM"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	newPayroll := model.PortalPayroll{
		RollMonth:   request.RollMonth,
		TotalAmount: request.TotalAmount,
		Flag:        0,
		CreatorID:   portalUser.ID,
		AddTime:     time.Now(),
		Status:      "create",
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		var existingPayroll model.PortalPayroll
		result := tx.
			Where("roll_month = ?", request.RollMonth).
			Where("status <> ?", "rejected").
			First(&existingPayroll)
		if result.Error == nil {
			return fmt.Errorf("payroll for month %s already exists and is not in 'rejected' status (ID: %d)",
				request.RollMonth, existingPayroll.ID)
		}

		if result.Error != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing payroll: %w", result.Error)
		}

		if err := tx.Create(&newPayroll).Error; err != nil {
			return err
		}
		if len(request.Items) > 0 {
			var payslips []model.PortalPayslip
			for _, item := range request.Items {
				payslips = append(payslips, model.PortalPayslip{
					PayrollID:     newPayroll.ID,
					UserID:        item.UserID,
					WalletAddress: item.WalletAddress,
					WalletType:    item.WalletType,
					WalletChain:   item.WalletChain,
					Amount:        item.Amount,
					Flag:          0,
					TransTime:     time.Now(),
					ReceiptHash:   "",
				})
			}
			if err := tx.Create(&payslips).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		if err.Error() == "payroll month already exists" {
			res.Code = codes.CODE_ERR_REPEAT
			res.Msg = err.Error()
		} else {
			res.Code = codes.CODE_ERR_UNKNOWN
			res.Msg = err.Error()
		}
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"payroll": newPayroll,
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollUpdate(c *gin.Context) {
	var request request.PortalPayrollCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	if !isValidYearMonth(request.RollMonth) {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "payroll month must be in format YYYY-MM"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var payroll model.PortalPayroll
	if err := db.First(&payroll, request.ID).Error; err != nil {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "payroll not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if payroll.Status != "create" {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "payroll status must be create"
		c.JSON(http.StatusOK, res)
		return
	}

	err := db.Transaction(func(tx *gorm.DB) error {
		payroll.RollMonth = request.RollMonth
		payroll.TotalAmount = request.TotalAmount
		payroll.Status = request.Status
		if err := tx.Save(&payroll).Error; err != nil {
			return err
		}

		// Delete existing payslips
		if err := tx.Where("payroll_id = ?", payroll.ID).Delete(&model.PortalPayslip{}).Error; err != nil {
			return err
		}

		// Insert new payslips
		if len(request.Items) > 0 {
			var payslips []model.PortalPayslip
			for _, item := range request.Items {
				payslips = append(payslips, model.PortalPayslip{
					PayrollID:     payroll.ID,
					UserID:        item.UserID,
					WalletID:      item.WalletID,
					WalletAddress: item.WalletAddress,
					WalletType:    item.WalletType,
					WalletChain:   item.WalletChain,
					Amount:        item.Amount,
					Flag:          0,
					TransTime:     time.Now(),
					ReceiptHash:   "",
				})
			}
			if err := tx.Create(&payslips).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to update payroll: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"payroll": payroll,
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollDelete(c *gin.Context) {
	var request request.PortalPayrollCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	if !isValidYearMonth(request.RollMonth) {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "payroll month must be in format YYYY-MM"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var payroll model.PortalPayroll
	if err := db.First(&payroll, request.ID).Error; err != nil {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "payroll not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if payroll.Status != "create" {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "payroll status must be create"
		c.JSON(http.StatusOK, res)
		return
	}

	payroll.Flag = 1
	if err := db.Save(&payroll).Error; err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to update payroll: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollSubmit(c *gin.Context) {
	var request request.PortalPayrollCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var payroll model.PortalPayroll
	if err := db.First(&payroll, request.ID).Error; err != nil {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "payroll not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if payroll.Status != "create" {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "payroll status must be create"
		c.JSON(http.StatusOK, res)
		return
	}

	if len(request.Desc) > 0 {
		payroll.Desc = payroll.Desc + "\n" + request.Desc
	}
	payroll.Status = "waiting_approval"
	if err := db.Save(&payroll).Error; err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to update payroll: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollAudit(c *gin.Context) {
	var request request.PortalPayrollCreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}

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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var payroll model.PortalPayroll
	if err := db.First(&payroll, request.ID).Error; err != nil {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "payroll not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if payroll.Status != "waiting_approval" {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "payroll status must be waiting_approval"
		c.JSON(http.StatusOK, res)
		return
	}

	if len(request.Desc) > 0 {
		if len(payroll.Desc) == 0 {
			payroll.Desc = request.Desc
		} else {
			payroll.Desc = payroll.Desc + "\n" + request.Desc
		}
	}
	switch request.Op {
	case "approve":
		payroll.Status = "approved"
	case "reject":
		payroll.Status = "rejected"
	default:
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "invalid op"
		c.JSON(http.StatusOK, res)
		return
	}

	if err := db.Save(&payroll).Error; err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to update payroll: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollPay(c *gin.Context) {
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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var request struct {
		ID uint64 `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "invalid request: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var payroll model.PortalPayroll
	if err := db.First(&payroll, request.ID).Error; err != nil {
		res.Code = codes.CODE_ERR_EXIST_OBJ
		res.Msg = "payroll not found"
		c.JSON(http.StatusOK, res)
		return
	}

	if payroll.Status != "approved" {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "payroll status must be approved"
		c.JSON(http.StatusOK, res)
		return
	}

	var portalSpecs []model.PortalSpec
	db.Where("flag = ? and spec_type = ?", 0, SPEC_TYPE_PAYROLL_SETTINGS).Find(&portalSpecs)
	var result PayrollSettings
	for _, spec := range portalSpecs {
		switch spec.SpecName {
		case "chain":
			result.Chain = spec.SpecValue
		case "pay_contract":
			result.PayContract = spec.SpecValue
		case "pay_token":
			result.PayToken = spec.SpecValue
		}
	}
	if !result.IsValid() {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "invalid payroll settings"
		c.JSON(http.StatusOK, res)
		return
	}
	var payTime = time.Now()
	payroll.Status = "paid"
	payroll.PayTime = &payTime
	if err := db.Save(&payroll).Error; err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to update payroll: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	c.JSON(http.StatusOK, res)
}

type WithWalletStaff struct {
	model.PortalUser
	WalletAddress string `gorm:"column:wallet_address;not null" json:"wallet_address"`
	WalletType    string `gorm:"column:wallet_type;not null" json:"wallet_type"`
	WalletChain   string `gorm:"column:wallet_chain;not null" json:"wallet_chain"`
	WalletID      uint64 `gorm:"column:wallet_id;not null" json:"wallet_id"`
}

func PortalPayrollStaffList(c *gin.Context) {
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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var staffs []WithWalletStaff
	var db = system.GetDb()
	err := db.Table("admin_portal_user u").
		Joins("LEFT JOIN admin_portal_user_wallet w ON u.id = w.user_id").
		Select("u.*, w.wallet_address, w.wallet_type, w.wallet_chain, w.id as wallet_id").
		Where("u.flag = 0").
		Find(&staffs).Error
	if err != nil {
		res.Code = codes.CODE_ERR_STATUS_GENERAL
		res.Msg = "failed to query payroll staff list: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"staff_list": staffs,
	}

	c.JSON(http.StatusOK, res)
}

func PortalPayrollStaffWallet(c *gin.Context) {
	var request request.PortalPayrollStaffWalletRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, common.Response{
			Code:      codes.CODE_ERR_REQFORMAT,
			Msg:       "invalid request: " + err.Error(),
			Timestamp: time.Now().Unix(),
		})
		return
	}
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

	userIDStr := c.Param("user_id")

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil || userID == 0 {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid user_id"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

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

	if !ethutil.IsHexAddress(request.WalletAddress) {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "invalid wallet address"
		c.JSON(http.StatusOK, res)
		return
	}

	var bindWallet model.PortalUserWallet
	db.Where("user_id = ? and flag = 0", bindUser.ID).First(&bindWallet)
	bindWallet.WalletAddress = request.WalletAddress
	bindWallet.WalletType = request.WalletType
	bindWallet.WalletChain = request.WalletChain
	bindWallet.Flag = 0
	bindWallet.UserID = bindUser.ID
	bindWallet.AddTime = time.Now()
	if err := db.Save(&bindWallet).Error; err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "failed to update wallet: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Code = codes.CODE_SUCCESS
	res.Msg = "success"
	c.JSON(http.StatusOK, res)
}

type PortalPayslipDetail struct {
	model.PortalPayslip
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

func PortalPayrollDetail(c *gin.Context) {
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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	payrollIDStr := c.Param("payroll_id")
	payrollID, err := strconv.ParseUint(payrollIDStr, 10, 64)
	if err != nil || payrollID == 0 {
		res.Code = codes.CODE_ERR_REQFORMAT
		res.Msg = "invalid payroll_id"
		c.JSON(http.StatusOK, res)
		return
	}

	db := system.GetDb()

	var payroll model.PortalPayroll
	if err := db.First(&payroll, payrollID).Error; err != nil {
		res.Code = codes.CODE_ERR_OBJ_NOT_FOUND
		res.Msg = "payroll not found"
		c.JSON(http.StatusOK, res)
		return
	}

	var items []PortalPayslipDetail
	err = db.Table("admin_portal_payslip p").
		Joins("LEFT JOIN admin_portal_user u ON p.user_id = u.id").
		Select("p.*, u.name as user_name, u.email as user_email").
		Where("p.payroll_id = ?", payroll.ID).
		Find(&items).Error
	if err != nil {
		res.Code = codes.CODE_ERR_UNKNOWN
		res.Msg = "failed to query payslips: " + err.Error()
		c.JSON(http.StatusOK, res)
		return
	}

	res.Data = gin.H{
		"payroll": payroll,
		"items":   items,
	}

	c.JSON(http.StatusOK, res)
}

type PersonlPayslip struct {
	model.PortalPayslip
	RollMonth string    `gorm:"column:roll_month;type:varchar(255);not null" json:"roll_month"`
	Status    string    `gorm:"column:status;type:varchar(255);not null" json:"status"`
	PayTime   time.Time `gorm:"column:pay_time" json:"pay_time"`
}

func PortalPayslipList(c *gin.Context) {
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
	if !ok || portalUser == nil {
		res.Code = codes.CODE_ERR_SECURITY
		res.Msg = "please login first"
		c.JSON(http.StatusOK, res)
		return
	}

	var result []PersonlPayslip
	var db = system.GetDb()
	db.Table("admin_portal_payslip ps").
		Joins("JOIN admin_portal_payroll pr ON ps.payroll_id = pr.id").
		Where("pr.status <> 'rejected' AND pr.flag = 0 AND ps.user_id = ? and ps.flag = 0", portalUser.ID).
		Select("ps.*, pr.roll_month, pr.status, pr.pay_time").
		Order("pr.roll_month desc").
		Find(&result)

	res.Data = result

	c.JSON(http.StatusOK, res)
}

/******** private method **********/
func isValidYearMonth(s string) bool {
	t, err := time.Parse("2006-01", s)
	if err != nil {
		return false
	}

	return t.Format("2006-01") == s
}

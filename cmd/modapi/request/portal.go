package request

type PortalLoginRequest struct {
	LoginID  string `json:"login_id"`
	Password string `json:"password"`
}

type PortalDeptCreateRequest struct {
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc" binding:"required"`
}

type PortalDeptUpdateRequest struct {
	ID   uint64 `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
	Desc string `json:"desc" binding:"required"`
}

type PortalUserUpdateRequest struct {
	ID       uint64   `json:"id" binding:"required"`
	Name     string   `json:"name" binding:"required"`
	Email    string   `json:"email" binding:"required"`
	Location string   `json:"location"`
	DeptIDs  []uint64 `json:"dept_ids"`
}

type PortalRoleCreateRequest struct {
	ID   uint64 `json:"id"`
	Name string `json:"name" binding:"required"`
}

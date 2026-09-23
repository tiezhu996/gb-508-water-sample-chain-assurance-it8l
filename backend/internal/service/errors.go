package service

import "errors"

var (
	ErrInvalidTransition = errors.New("requested status transition is not allowed")
	ErrInvalidInput      = errors.New("business input validation failed")
	ErrUnauthorized      = errors.New("invalid username or password")
	ErrInactiveUser      = errors.New("user account is inactive")

	// Chain-of-custody handover keeps a dedicated, stable set of rejection
	// reasons so the API mapping and the sample page can show the exact
	// blocking rule without parsing message text.
	ErrCustodyDisposed     = errors.New("样本已处置，禁止再交接保管")
	ErrCustodyNotCustodian = errors.New("仅当前保管人可以提交保管交接")
	ErrCustodySameTarget   = errors.New("目标保管人与当前保管人相同")
	ErrCustodyAccount      = errors.New("目标账号不存在或已停用")
	ErrCustodyTargetRole   = errors.New("目标账号角色不足，须具备 operator 及以上角色")
)

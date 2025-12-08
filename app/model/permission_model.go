package model

type PermissionUser struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
}
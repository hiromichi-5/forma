package entity

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleEditor Role = "editor"
)

type Member struct {
	UserRef
	Role Role
}

func (r Role) Valid() bool {
	switch r {
	case RoleAdmin, RoleEditor:
		return true
	default:
		return false
	}
}

func (r Role) CanEdit() bool {
	return r == RoleAdmin || r == RoleEditor
}

func (r Role) CanAdmin() bool {
	return r == RoleAdmin
}

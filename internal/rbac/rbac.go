package rbac

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
)

func IsPrivileged(role string) bool {
	return role == RoleOwner || role == RoleAdmin
}


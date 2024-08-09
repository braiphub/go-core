package braipsession

type Role string

const (
	RoleSudo         Role = "sudo"
	RoleAdmin        Role = "admin"
	RoleCollaborator Role = "collaborator"
	RoleCommonUser   Role = "common_user"
	RoleEmployee     Role = "employee"
)

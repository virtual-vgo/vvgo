package access

type Role string

func (x Role) String() string { return string(x) }

const (
	RoleUnknown    Role = "unknown"
	RoleVVGOMember Role = "vvgo-member"
)

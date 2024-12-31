package user

type (
	Role string
	User string
)

const (
	Manager    Role = "manager"
	Technician Role = "technician"
)

func (u User) String() string {
	return string(u)
}

func (r Role) String() string {
	return string(r)
}

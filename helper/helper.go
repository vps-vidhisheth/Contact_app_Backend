package helper

type UserData struct {
	IsAdmin  bool
	IsActive bool
}

func IsAuthorizedAdmin(user UserData) bool {
	return user.IsAdmin && user.IsActive
}

func IsAuthorizedStaff(user UserData) bool {
	return user.IsActive
}

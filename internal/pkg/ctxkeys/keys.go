package ctxkeys

type contextKey string

const (
	UserIDKey contextKey = "userID"

	UserUUIDKey contextKey = "userUUID"

	OwnerIDKey contextKey = "ownerID"

	UserRoleKey contextKey = "userRole"
)

package utils

const defaultLimit = 20

func SetPaginationDefaults(limit, page int32) (int32, int32) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return limit, offset
}

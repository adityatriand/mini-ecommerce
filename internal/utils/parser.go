package utils

import "strconv"

func ParseIDFromString(idStr string) (uint, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func ParseUserIDFromString(userIDStr string) (uint, error) {
	uid, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(uid), nil
}
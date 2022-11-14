package model

import (
	"vec/db"
)

func GetPatentIDByVectorID(vecID string) (string, error) {
	v, ok, err := GetRedis(db.Get().Redis, vecID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return v.(string), nil
}

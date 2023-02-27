package model

import (
	"vec/db"
	"vec/model/vector"
)

func GetPatentIDByVectorID(filed string, vecID int64) (string, error) {
	v, ok, err := GetRedis(db.Get().Redis, vector.AddFieldToVectorID(filed, vecID))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return v.(string), nil
}

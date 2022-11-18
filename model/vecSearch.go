package model

import (
	"fmt"

	"vec/db"
)

func GetPatentIDByVectorID(filed, vecID string) (string, error) {
	v, ok, err := GetRedis(db.Get().Redis, fmt.Sprintf("%s-%s", filed, vecID))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return v.(string), nil
}

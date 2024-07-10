package user_repo

import (
	"context"
	"errors"
)

type Repo interface {
	// 从ctx中解析出用户唯一标识
	ParseUID(ctx context.Context) (string, error)
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (repoImpl) ParseUID(ctx context.Context) (string, error) {
	return "", errors.New("please call common_button.SetUserRepo(repo)")
}

package prize_repo

import (
	"context"

	"github.com/zlyuancn/common_button/pb"
)

type Repo interface {
	// 解析奖品id
	ParsePrizeID(ctx context.Context, prizeID string) (*pb.Prize, error)
}

var defRepo Repo = repoImpl{}

func SetRepo(repo Repo) {
	defRepo = repo
}
func GetRepo() Repo {
	return defRepo
}

type repoImpl struct{}

func (repoImpl) ParsePrizeID(ctx context.Context, prizeID string) (*pb.Prize, error) {
	ret := &pb.Prize{
		PrizeId:   prizeID,
		PrizeName: "id=" + prizeID,
		PrizeUrl:  "",
	}
	return ret, nil
}

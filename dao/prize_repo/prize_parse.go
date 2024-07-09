package prize_repo

import (
	"context"

	"github.com/zlyuancn/common_button/pb"
)

// 奖品id解析为奖品数据
type PrizeIDParse func(ctx context.Context, prizeID string) (*pb.Prize, error)

var prizeIDParse PrizeIDParse = defPrizeIDParse

// 设置奖品解析函数
func SetPrizeIDParseFn(fn PrizeIDParse) {
	prizeIDParse = fn
}

// 默认的奖品解析函数
func defPrizeIDParse(ctx context.Context, prizeID string) (*pb.Prize, error) {
	ret := &pb.Prize{
		PrizeId:   prizeID,
		PrizeName: "id=" + prizeID,
		PrizeUrl:  "",
	}
	return ret, nil
}

// 解析奖品id
func ParsePrizeID(ctx context.Context, prizeID string) (*pb.Prize, error) {
	return prizeIDParse(ctx, prizeID)
}

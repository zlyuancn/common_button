package dao

import (
	"context"

	"github.com/zlyuancn/common_button/dao/loopload"
	"github.com/zlyuancn/common_button/pb"
)

type ButtonRepo interface {
	// 根据业务模块id和场景/页面id批量获取按钮, 场景/页面id为空则获取业务模块id下的所有按钮
	GetButtonsByModuleAndScene(ctx context.Context, moduleID pb.ButtonModuleID, sceneIDs []pb.ButtonSceneID) ([]*pb.Button, error)
	// 根据按钮id获取按钮数据
	GetButtonByID(ctx context.Context, buttonID int32) (*pb.Button, error)
}

type buttonRepo struct{}

func (buttonRepo) GetButtonsByModuleAndScene(ctx context.Context, moduleID pb.ButtonModuleID, sceneIDs []pb.ButtonSceneID) ([]*pb.Button, error) {
	return loopload.GetButtonsByModuleAndScene(ctx, moduleID, sceneIDs)
}

func (buttonRepo) GetButtonByID(ctx context.Context, buttonID int32) (*pb.Button, error) {
	return loopload.GetButtonByID(ctx, buttonID)
}

func NewButtonRepo() ButtonRepo {
	return buttonRepo{}
}

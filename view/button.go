package view

import (
	"context"
	"errors"

	"github.com/zly-app/zapp/logger"
	"go.uber.org/zap"

	"github.com/zlyuancn/common_button/dao"
	"github.com/zlyuancn/common_button/pb"
	"github.com/zlyuancn/common_button/util/task_repo"
)

type implCli struct {
	pb.UnimplementedCommonButtonServiceServer
}

func (impl implCli) GetButtonList(ctx context.Context, req *pb.GetButtonListReq) (*pb.GetButtonListRsp, error) {
	bs, err := dao.GetButtonRepo().GetButtonsByModuleAndScene(ctx, req.ModuleId, req.SceneId)
	if err != nil {
		logger.Error(ctx, "GetButtonList call GetButtonsByModuleAndScene err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	bs, err = task_repo.GetRepo().MultiRenderTasksStatus(ctx, bs)
	if err != nil {
		logger.Error(ctx, "GetButtonList call renderTaskStatus err", zap.Any("req", req), zap.Error(err))
		return nil, err
	}

	ret := &pb.GetButtonListRsp{
		Buttons: bs,
	}
	return ret, nil
}

func (implCli) ClickButton(ctx context.Context, req *pb.ClickButtonReq) (*pb.ClickButtonRsp, error) {
	return nil, errors.New("未实现")
}

func (implCli) OneClickFinish(ctx context.Context, req *pb.OneClickFinishReq) (*pb.OneClickFinishRsp, error) {
	return nil, errors.New("未实现")
}

func NewButtonService() pb.CommonButtonServiceServer {
	return &implCli{}
}

package msggateway

import (
	"context"

	"github.com/OpenIMSDK/Open-IM-Server/pkg/proto/msg"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/proto/sdkws"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/rpcclient"
	"github.com/OpenIMSDK/Open-IM-Server/pkg/utils"
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/proto"
)

type Req struct {
	ReqIdentifier int32  `json:"reqIdentifier" validate:"required"`
	Token         string `json:"token" `
	SendID        string `json:"sendID" validate:"required"`
	OperationID   string `json:"operationID" validate:"required"`
	MsgIncr       string `json:"msgIncr" validate:"required"`
	Data          []byte `json:"data"`
}

func (r *Req) String() string {
	return utils.StructToJsonString(r)
}

type Resp struct {
	ReqIdentifier int32  `json:"reqIdentifier"`
	MsgIncr       string `json:"msgIncr"`
	OperationID   string `json:"operationID"`
	ErrCode       int    `json:"errCode"`
	ErrMsg        string `json:"errMsg"`
	Data          []byte `json:"data"`
}

func (r *Resp) String() string {
	return utils.StructToJsonString(r)
}

type MessageHandler interface {
	GetSeq(context context.Context, data Req) ([]byte, error)
	SendMessage(context context.Context, data Req) ([]byte, error)
	SendSignalMessage(context context.Context, data Req) ([]byte, error)
	PullMessageBySeqList(context context.Context, data Req) ([]byte, error)
	UserLogout(context context.Context, data Req) ([]byte, error)
	SetUserDeviceBackground(context context.Context, data Req) ([]byte, bool, error)
}

var _ MessageHandler = (*GrpcHandler)(nil)

type GrpcHandler struct {
	msgRpcClient *rpcclient.MsgClient
	validate     *validator.Validate
}

func NewGrpcHandler(validate *validator.Validate, msgRpcClient *rpcclient.MsgClient) *GrpcHandler {
	return &GrpcHandler{msgRpcClient: msgRpcClient, validate: validate}
}

func (g GrpcHandler) GetSeq(context context.Context, data Req) ([]byte, error) {
	req := sdkws.GetMaxSeqReq{}
	if err := proto.Unmarshal(data.Data, &req); err != nil {
		return nil, err
	}
	if err := g.validate.Struct(&req); err != nil {
		return nil, err
	}
	resp, err := g.msgRpcClient.GetMaxSeq(context, &req)
	if err != nil {
		return nil, err
	}
	c, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (g GrpcHandler) SendMessage(context context.Context, data Req) ([]byte, error) {
	msgData := sdkws.MsgData{}
	if err := proto.Unmarshal(data.Data, &msgData); err != nil {
		return nil, err
	}
	if err := g.validate.Struct(&msgData); err != nil {
		return nil, err
	}
	req := msg.SendMsgReq{MsgData: &msgData}
	resp, err := g.msgRpcClient.SendMsg(context, &req)
	if err != nil {
		return nil, err
	}
	c, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (g GrpcHandler) SendSignalMessage(context context.Context, data Req) ([]byte, error) {
	signalReq := sdkws.SignalReq{}
	if err := proto.Unmarshal(data.Data, &signalReq); err != nil {
		return nil, err
	}
	if err := g.validate.Struct(&signalReq); err != nil {
		return nil, err
	}
	//req := pbRtc.SignalMessageAssembleReq{SignalReq: &signalReq, OperationID: "111"}
	//todo rtc rpc call
	resp, err := g.msgRpcClient.SendMsg(context, nil)
	if err != nil {
		return nil, err
	}
	c, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (g GrpcHandler) PullMessageBySeqList(context context.Context, data Req) ([]byte, error) {
	req := sdkws.PullMessageBySeqsReq{}
	if err := proto.Unmarshal(data.Data, &req); err != nil {
		return nil, err
	}
	if err := g.validate.Struct(data); err != nil {
		return nil, err
	}
	resp, err := g.msgRpcClient.PullMessageBySeqList(context, &req)
	if err != nil {
		return nil, err
	}
	c, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (g GrpcHandler) UserLogout(context context.Context, data Req) ([]byte, error) {
	//todo
	resp, err := g.msgRpcClient.PullMessageBySeqList(context, nil)
	if err != nil {
		return nil, err
	}
	c, err := proto.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (g GrpcHandler) SetUserDeviceBackground(_ context.Context, data Req) ([]byte, bool, error) {
	req := sdkws.SetAppBackgroundStatusReq{}
	if err := proto.Unmarshal(data.Data, &req); err != nil {
		return nil, false, err
	}
	if err := g.validate.Struct(data); err != nil {
		return nil, false, err
	}
	return nil, req.IsBackground, nil
}

//func (g GrpcHandler) call[T any](ctx context.Context, data Req, m proto.Message, rpc func(ctx context.Context, req proto.Message)) ([]byte, error) {
//	if err := proto.Unmarshal(data.Data, m); err != nil {
//		return nil, err
//	}
//	if err := g.validate.Struct(m); err != nil {
//		return nil, err
//	}
//	rpc(ctx, m)
//	req := msg.SendMsgReq{MsgData: &msgData}
//	resp, err := g.notification.Msg.SendMsg(context, &req)
//	if err != nil {
//		return nil, err
//	}
//	c, err := proto.Marshal(resp)
//	if err != nil {
//		return nil, err
//	}
//	return c, nil
//}

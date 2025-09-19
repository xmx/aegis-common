package errcode

import "github.com/xgfone/ship/v5"

var (
	ErrDataNotExists          = ship.ErrNotFound.Newf("数据不存在")
	ErrCertificateUnavailable = ship.ErrBadRequest.Newf("未配置有效的证书")
	ErrCertificateInvalid     = ship.ErrBadRequest.Newf("无效证书")
	ErrBrokerInactive         = ship.ErrBadRequest.Newf("代理节点未上线")
)

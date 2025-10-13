package shipx

import (
	"encoding/base64"
	"encoding/json"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/xgfone/ship/v5"
	"github.com/xmx/aegis-common/contract/problem"
	"github.com/xmx/aegis-common/library/validation"
)

func NotFound(_ *ship.Context) error {
	return ship.ErrNotFound.Newf("资源不存在")
}

func HandleErrorWithHost(host string) func(*ship.Context, error) {
	return func(c *ship.Context, e error) {
		statusCode, title, detail := UnpackError(e)
		pd := &problem.Details{
			Host:     host,
			Title:    title,
			Status:   statusCode,
			Detail:   detail,
			Instance: c.Path(),
			Method:   c.Method(),
			Datetime: time.Now().UTC(),
		}
		_ = c.JSON(statusCode, pd)
	}
}

func HandleError(c *ship.Context, e error) {
	statusCode, title, detail := UnpackError(e)
	pd := &problem.Details{
		Host:     c.Host(),
		Title:    title,
		Status:   statusCode,
		Detail:   detail,
		Instance: c.Path(),
		Method:   c.Method(),
		Datetime: time.Now().UTC(),
	}
	_ = c.JSON(statusCode, pd)
}

func UnpackError(err error) (statusCode int, title string, detail string) {
	statusCode = http.StatusBadRequest
	title = "请求错误"
	detail = err.Error()

	switch ce := err.(type) {
	case ship.HTTPServerError:
		statusCode = ce.Code
	case *ship.HTTPServerError:
		statusCode = ce.Code
	case *validation.ValidError:
		title = "参数校验错误"
	case *time.ParseError:
		detail = "时间格式错误，正确格式：" + ce.Layout
	case *net.ParseError:
		detail = ce.Text + " 不是有效的 " + ce.Type
	case base64.CorruptInputError:
		detail = "Base64 编码错误：" + ce.Error()
	case *json.SyntaxError:
		detail = "错误的 JSON 格式"
	case *json.UnmarshalTypeError:
		detail = ce.Field + " 收到无效的数据类型"
	case *strconv.NumError:
		var msg string
		if sn := strings.SplitN(ce.Func, "Parse", 2); len(sn) == 2 {
			msg = ce.Num + " 不是 " + strings.ToLower(sn[1]) + " 类型"
		} else {
			msg = "类型错误：" + ce.Num
		}
		detail = msg
	case *http.MaxBytesError:
		limit := strconv.FormatInt(ce.Limit, 10)
		detail = "请求报文超过 " + limit + " 个字节限制"
		statusCode = http.StatusRequestEntityTooLarge
	}

	return statusCode, title, detail
}

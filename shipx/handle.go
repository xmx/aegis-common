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
	"github.com/xmx/aegis-common/library/validation"
	"github.com/xmx/aegis-common/problem"
)

func NotFound(_ *ship.Context) error {
	return ship.ErrNotFound.Newf("资源不存在")
}

func HandleError(c *ship.Context, e error) {
	pd := &problem.Details{
		Host:     c.Host(),
		Instance: c.Path(),
		Method:   c.Method(),
		Datetime: time.Now().UTC(),
	}
	ei, _ := AssertError(e)
	if ei != nil {
		pd.Title = ei.Title
		pd.Detail = ei.Detail
		pd.Status = ei.Status
	} else {
		pd.Title = "请求错误"
		pd.Detail = e.Error()
		pd.Status = http.StatusBadRequest
	}

	_ = c.JSON(pd.Status, pd)
}

type ErrorInfo struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func AssertError(err error) (*ErrorInfo, bool) {
	if err == nil {
		return nil, false
	}

	ei := &ErrorInfo{
		Status: http.StatusBadRequest,
		Title:  "请求错误",
		Detail: err.Error(),
	}
	switch e := err.(type) {
	case ship.HTTPServerError:
		ei.Status = e.Code
		return ei, true

	case *ship.HTTPServerError:
		ei.Status = e.Code
		return ei, true

	case *validation.ValidError:
		ei.Title = "参数校验错误"
		return ei, true

	case base64.CorruptInputError:
		ei.Detail = "base64 编码错误：" + e.Error()
		return ei, true

	case *json.SyntaxError:
		ei.Detail = "错误的 JSON 格式"
		return ei, true

	case *json.UnmarshalTypeError:
		ei.Detail = e.Field + " 收到无效的数据类型"
		return ei, true

	case *time.ParseError:
		ei.Detail = "时间格式错误，正确格式：" + e.Layout
		return ei, true

	case *net.ParseError:
		ei.Detail = e.Text + " 不是有效的 " + e.Type
		return ei, true

	case *http.MaxBytesError:
		ei.Status = http.StatusRequestEntityTooLarge
		limit := strconv.FormatInt(e.Limit, 10)
		ei.Detail = "请求报文超过 " + limit + " 个字节限制"
		return ei, true

	case *strconv.NumError:
		if fn, found := strings.CutPrefix(e.Func, "Parse"); found {
			ei.Detail = e.Num + " 不是 " + strings.ToLower(fn) + " 类型"
		} else {
			ei.Detail = "类型错误：" + e.Num
		}
		return ei, true
	}

	return nil, false
}

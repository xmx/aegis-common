package jsmod

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/grafana/sobek"
	"github.com/xmx/jsos/jsvm"
)

func NewConsole() jsvm.Module {
	return &stdConsole{}
}

type stdConsole struct {
	eng jsvm.Engineer
}

func (mod *stdConsole) Preload(eng jsvm.Engineer) (string, any, bool) {
	mod.eng = eng
	vals := map[string]any{
		"debug": mod.stdout,
		"info":  mod.stdout,
		"warn":  mod.stdout,
		"error": mod.stdout,
		"log":   mod.stdout,
	}

	return "console", vals, false
}

func (mod *stdConsole) stdout(call sobek.FunctionCall) sobek.Value {
	rt := mod.eng.Runtime()
	stdout, _ := mod.eng.Output()
	msg, err := mod.format(call)
	if err != nil {
		return rt.NewGoError(err)
	}
	if _, err = stdout.Write(msg); err != nil {
		return rt.NewGoError(err)
	}

	return sobek.Undefined()
}

func (mod *stdConsole) format(call sobek.FunctionCall) ([]byte, error) {
	buf := new(bytes.Buffer)
	for _, arg := range call.Arguments {
		if err := mod.parse(buf, arg); err != nil {
			return nil, err
		}
	}
	buf.WriteByte('\n')

	return buf.Bytes(), nil
}

func (mod *stdConsole) parse(buf *bytes.Buffer, val sobek.Value) error {
	switch {
	case sobek.IsUndefined(val), sobek.IsNull(val):
		buf.WriteString(val.String())
		return nil
	}

	export := val.Export()
	switch v := export.(type) {
	case fmt.Stringer:
		buf.WriteString(v.String())
	case string:
		buf.WriteString(v)
	case int64:
		buf.WriteString(strconv.FormatInt(v, 10))
	case float64:
		buf.WriteString(strconv.FormatFloat(v, 'g', -1, 64))
	case bool:
		buf.WriteString(strconv.FormatBool(v))
	case []byte:
		str := base64.StdEncoding.EncodeToString(v)
		buf.WriteString(str)
	case func(sobek.FunctionCall) sobek.Value:
		buf.WriteString("<Function>")
	case sobek.ArrayBuffer:
		bs := v.Bytes()
		str := base64.StdEncoding.EncodeToString(bs)
		buf.WriteString(str)
	default:
		return mod.reflectParse(buf, v)
	}

	return nil
}

func (*stdConsole) reflectParse(buf *bytes.Buffer, v any) error {
	vof := reflect.ValueOf(v)
	switch vof.Kind() {
	case reflect.String:
		buf.WriteString(vof.String())
	case reflect.Int64:
		buf.WriteString(strconv.FormatInt(vof.Int(), 10))
	case reflect.Float64:
		buf.WriteString(strconv.FormatFloat(vof.Float(), 'g', -1, 64))
	case reflect.Bool:
		buf.WriteString(strconv.FormatBool(vof.Bool()))
	default:
		tmp := new(bytes.Buffer)
		if err := json.NewEncoder(tmp).Encode(v); err == nil && tmp.Len() != 0 {
			_, _ = buf.ReadFrom(tmp)
			return nil
		}
		vts := vof.Type().String()
		buf.WriteString(vts)
	}

	return nil
}

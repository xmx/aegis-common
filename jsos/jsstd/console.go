package jsstd

import (
	"bytes"
	"encoding/base64"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"reflect"
	"strconv"

	"github.com/grafana/sobek"
	"github.com/xmx/aegis-common/jsos/jsvm"
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
		"log":   mod.write(false),
		"debug": mod.write(false),
		"info":  mod.write(false),
		"warn":  mod.write(false),
		"error": mod.write(true),
	}

	return "console", vals, false
}

func (mod *stdConsole) write(writeErr bool) func(call sobek.FunctionCall) sobek.Value {
	return func(call sobek.FunctionCall) sobek.Value {
		rt := mod.eng.Runtime()
		buf, err := mod.format(call)
		if err != nil {
			return rt.NewGoError(err)
		}
		stdout, stderr := mod.eng.Output()
		if writeErr {
			_, err = buf.WriteTo(stderr)
		} else {
			_, err = buf.WriteTo(stdout)
		}
		if err != nil {
			return rt.NewGoError(err)
		}

		return sobek.Undefined()
	}
}

func (mod *stdConsole) format(call sobek.FunctionCall) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	for _, arg := range call.Arguments {
		if err := mod.parse(buf, arg); err != nil {
			return nil, err
		}
	}
	buf.WriteByte('\n')

	return buf, nil
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
	case reflect.Func:
		buf.WriteString("<Function>")
	default:
		enc := jsontext.NewEncoder(buf)
		if err := json.MarshalEncode(enc, v); err == nil {
			return nil
		}

		vts := vof.Type().String()
		buf.WriteString(vts)
	}

	return nil
}

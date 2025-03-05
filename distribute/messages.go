package distribute

import (
	"reflect"
)

func newGeneralMsg(data any) GeneralMsg {
	return GeneralMsg{
		TypeName: reflect.TypeOf(data).Name(),
		Data:     data,
	}
}

func EncodeMsg(msg interface{}) GeneralMsg {
	return GeneralMsg{
		Data:     msg,
		TypeName: reflect.TypeOf(msg).Name(),
	}
}

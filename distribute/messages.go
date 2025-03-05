package distribute

import (
	"reflect"
)

func newGeneralMsg(data any) GeneralMsg {
	// fmt.Println("Sending data of type", reflect.TypeOf(data))
	return GeneralMsg{
		TypeName: reflect.TypeOf(data).Name(),
		Data:     data,
	}
}

// new problem: There is no way to send nested structs...
// Time to do recursive
// func Structify(msg interface{}, target *GeneralMsg) error {
// 	// fmt.Println(msg)
// 	// fmt.Println(target)
// 	jsonEnc, _ := json.Marshal(msg) // NOTE: This may not be needed, maybe try to remove
// 	err := json.Unmarshal(jsonEnc, target)

// 	if err != nil {
// 		fmt.Println("Could not parse message:", msg)
// 		return err
// 	}
// 	return nil
// }

func EncodeMsg(msg interface{}) GeneralMsg {
	// fmt.Println(reflect.TypeOf(msg))
	return GeneralMsg{
		Data:     msg,
		TypeName: reflect.TypeOf(msg).Name(),
	}
}

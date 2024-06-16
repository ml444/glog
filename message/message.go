package message

import "google.golang.org/protobuf/types/known/structpb"

func (x *Message) ExtraMap(m map[string]interface{}) error {
	fields := make(map[string]*structpb.Value)
	for k, v := range m {
		val, err := structpb.NewValue(v)
		if err != nil {
			return err
		}
		fields[k] = val
	}
	x.Extra = fields
	return nil
}

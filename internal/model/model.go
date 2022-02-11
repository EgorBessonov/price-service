//Package model represents models witch are used in price service
package model

import "encoding/json"

//Share type represents Share structure in price service
type Share struct {
	Name      int32
	Bid       float32
	Ask       float32
	UpdatedAt string
}

//MarshalBinary realisation for share model
func (share *Share) MarshalBinary() ([]byte, error) {
	return json.Marshal(share)
}

//UnmarshalBinary realisation for share model
func (share *Share) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &share)
}

//ShareMap type represents share types structure in price service
type ShareMap struct {
	Map map[int32]map[string]*chan *Share
}

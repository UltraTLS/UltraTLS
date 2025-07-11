package protocol

type Codec interface {
	Encode(interface{}) ([]byte, error)
	Decode([]byte) (interface{}, error)
}
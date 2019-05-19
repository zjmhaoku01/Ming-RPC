// 使用自带库gob进行编码/解码
package codec

import (
	"bytes"
	"encoding/gob"
)

// 客户端与服务端之间传输的数据类型
type Data struct {
	Name string        // service名称
	Args []interface{} // 传参列表
	Err  string        // error
}

// Encode
func Encode(data Data) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Decode
func Decode(b []byte) (Data, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var data Data
	if err := decoder.Decode(&data); err != nil {
		return Data{}, err
	}
	return data, nil
}

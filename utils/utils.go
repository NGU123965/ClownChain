package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
)

// IntToHex int64转换成字节数组
func IntToHex(data int64) []byte {
	buffer := new(bytes.Buffer) // 新建一个buffer
	err := binary.Write(buffer, binary.BigEndian, data)
	if nil != err {
		log.Panicf("int to []byte failed! %v\n", err)
	}
	return buffer.Bytes()
}

// JSONToArray windows下JSON转换的标准输入格式：
// bc.exe send -from "[\"Alice\",\"Bob\"]" -to "[\"Bob\",\"troytan\"]" -amount "[\"5\",\"2\"]"
// bc.exe send -from "[\"Alice\"]" -to "[\"Bob\"]" -amount "[\"5\"]"
// 标准JSON格式转数组
func JSONToArray(jsonString string) []string {
	var strArr []string
	// json.Unmarshal
	if err := json.Unmarshal([]byte(jsonString), &strArr); err != nil {
		log.Panicf("json to []string failed! %v\n", err)
	}
	return strArr
}

// Reverse 反转切片
func Reverse(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// GobEncode 将结构体序列化为字节数组
func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if nil != err {
		log.Panicf("encode the data failed! %v\n", err)
	}
	return buff.Bytes()
}

// xxxxx(指令)xxxxx(数据...)

// CommandToBytes 将命令转为字节数组
// 指令长度最长为12位
func CommandToBytes(command string) []byte {
	var bytes [12]byte // 命令长度
	for i, c := range command {
		bytes[i] = byte(c) // 转换
	}
	return bytes[:]
}

// BytesToCommand 将字节数组转成cmd
func BytesToCommand(bytes []byte) string {
	var command []byte // 接收命令
	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}
	return fmt.Sprintf("%s", command)
}

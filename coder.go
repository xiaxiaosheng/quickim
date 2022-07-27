package main

import (
	"bytes"
	"crypto/des"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-netty/go-netty"
	"github.com/go-netty/go-netty/codec"
	"github.com/go-netty/go-netty/utils"
	logger "github.com/opentrx/seata-golang/v2/pkg/util/log"
)

// JSONCodec create a json codec
func StringCodec() codec.Codec {
	return &stringCodec{}
}

type stringCodec struct {
}

func (*stringCodec) CodecName() string {
	return "string-codec"
}

func (j *stringCodec) HandleRead(ctx netty.InboundContext, message netty.Message) {

	bytes := utils.MustToBytes(message)

	message = string(bytes)

	ctx.HandleRead(message)
}

func (j *stringCodec) HandleWrite(ctx netty.OutboundContext, message netty.Message) {
	str, ok := message.(string)
	if !ok {
		panic(errors.New("message type error"))
	}
	// post json
	ctx.HandleWrite([]byte(str))
}

func decode(desKey, str string) string {
	if desKey != "" {
		return DecryptDES_ECB(str, desKey)
	}
	return str
}

func encodeMsg(desKey string, msg map[string]interface{}) string {
	tmp, _ := json.Marshal(msg)
	if desKey != "" {
		return EncryptDES_ECB(string(tmp), desKey)
	}

	return string(tmp)
}

//ECB加密
func EncryptDES_ECB(src, key string) string {
	data := []byte(src)
	keyByte := []byte(key)
	keyByte = keyByte[:8]
	block, err := des.NewCipher(keyByte)
	if err != nil {
		logger.Error(err)
		return ""
	}
	bs := block.BlockSize()
	//对明文数据进行补码
	data = PKCS5Padding(data, bs)
	if len(data)%bs != 0 {
		logger.Error("Need a multiple of the blocksize")
		return ""
	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		//对明文按照blocksize进行分块加密
		//必要时可以使用go关键字进行并行加密
		block.Encrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	return fmt.Sprintf("%X", out)
}

//ECB解密
func DecryptDES_ECB(src, key string) string {
	data, err := hex.DecodeString(src)
	if err != nil {
		logger.Error(err)
		return ""
	}
	keyByte := []byte(key)
	keyByte = keyByte[:8]
	block, err := des.NewCipher(keyByte)
	if err != nil {
		logger.Error(err)
		return ""

	}
	bs := block.BlockSize()
	if len(data)%bs != 0 {
		err = errors.New("crypto/cipher: input not full blocks")
		logger.Error(err)
		return ""

	}
	out := make([]byte, len(data))
	dst := out
	for len(data) > 0 {
		block.Decrypt(dst, data[:bs])
		data = data[bs:]
		dst = dst[bs:]
	}
	out = PKCS5UnPadding(out)
	return string(out)
}

//明文补码算法
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//明文减码算法
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}

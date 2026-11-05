package wxcallback

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	wxbase "github.com/WeixinCloud/wxcloudrun-wxcomponent/comm/wx/base"
	"github.com/gin-gonic/gin"
)

type encryptedEnvelope struct {
	Encrypt string `xml:"Encrypt"`
}

// decodeComponentCallback converts WeChat's AES encrypted XML into the JSON
// shape consumed by the existing component event handlers. JSON is retained
// for compatibility with callers using the older integration format.
func decodeComponentCallback(c *gin.Context, body []byte) ([]byte, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, errors.New("empty callback body")
	}
	if trimmed[0] == '{' {
		return trimmed, nil
	}

	var envelope encryptedEnvelope
	if err := xml.Unmarshal(trimmed, &envelope); err != nil {
		return nil, fmt.Errorf("invalid XML envelope: %w", err)
	}
	var plaintext []byte
	if envelope.Encrypt != "" {
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")
		if err := verifyCallbackSignature(c.Query("msg_signature"), timestamp, nonce, envelope.Encrypt); err != nil {
			return nil, err
		}
		var err error
		plaintext, err = decryptCallback(envelope.Encrypt)
		if err != nil {
			return nil, err
		}
	} else {
		// Some test tools and older integrations send plaintext XML.
		if err := verifyCallbackSignature(c.Query("signature"), c.Query("timestamp"), c.Query("nonce"), ""); err != nil {
			return nil, err
		}
		plaintext = trimmed
	}
	var message componentXMLMessage
	if err := xml.Unmarshal(plaintext, &message); err != nil {
		return nil, fmt.Errorf("invalid decrypted component XML: %w", err)
	}
	return json.Marshal(message)
}

func handleComponentVerification(c *gin.Context) {
	echo := c.Query("echostr")
	if echo == "" {
		c.String(http.StatusBadRequest, "missing echostr")
		return
	}
	var response string
	var err error
	if c.Query("encrypt_type") == "aes" {
		err = verifyCallbackSignature(c.Query("msg_signature"), c.Query("timestamp"), c.Query("nonce"), echo)
		if err == nil {
			var plaintext []byte
			plaintext, err = decryptCallback(echo)
			response = string(plaintext)
		}
	} else {
		err = verifyCallbackSignature(c.Query("signature"), c.Query("timestamp"), c.Query("nonce"), "")
		response = echo
	}
	if err != nil {
		c.String(http.StatusUnauthorized, "invalid signature")
		return
	}
	c.String(http.StatusOK, response)
}

func verifyCallbackSignature(signature, timestamp, nonce, encrypted string) error {
	token := os.Getenv("WX_CALLBACK_TOKEN")
	if token == "" {
		return errors.New("WX_CALLBACK_TOKEN is not configured")
	}
	if signature == "" || timestamp == "" || nonce == "" {
		return errors.New("missing callback signature parameters")
	}
	parts := []string{token, timestamp, nonce}
	if encrypted != "" {
		parts = append(parts, encrypted)
	}
	sort.Strings(parts)
	hash := sha1.Sum([]byte(strings.Join(parts, "")))
	if !strings.EqualFold(signature, hex.EncodeToString(hash[:])) {
		return errors.New("callback signature mismatch")
	}
	return nil
}

func decryptCallback(ciphertext string) ([]byte, error) {
	encodedKey := os.Getenv("WX_ENCODING_AES_KEY")
	if encodedKey == "" {
		return nil, errors.New("WX_ENCODING_AES_KEY is not configured")
	}
	if len(encodedKey) == 43 {
		encodedKey += "="
	}
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != 32 {
		return nil, errors.New("WX_ENCODING_AES_KEY must be the 43-character WeChat EncodingAESKey")
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil || len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, errors.New("invalid encrypted callback payload")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	plain := make([]byte, len(data))
	cipher.NewCBCDecrypter(block, key[:aes.BlockSize]).CryptBlocks(plain, data)
	if len(plain) == 0 {
		return nil, errors.New("empty decrypted callback payload")
	}
	padding := int(plain[len(plain)-1])
	if padding == 0 || padding > 32 || padding > len(plain) {
		return nil, errors.New("invalid encrypted callback padding")
	}
	for _, b := range plain[len(plain)-padding:] {
		if int(b) != padding {
			return nil, errors.New("invalid encrypted callback padding")
		}
	}
	plain = plain[:len(plain)-padding]
	if len(plain) < 20 {
		return nil, errors.New("decrypted callback payload is too short")
	}
	messageLength := binary.BigEndian.Uint32(plain[16:20])
	messageEnd := uint64(20) + uint64(messageLength)
	if messageEnd > uint64(len(plain)) {
		return nil, errors.New("invalid decrypted callback message length")
	}
	message := plain[20:messageEnd]
	appid := string(plain[messageEnd:])
	expectedAppid := wxbase.GetAppid()
	if expectedAppid != "" && appid != expectedAppid {
		return nil, errors.New("decrypted callback appid mismatch")
	}
	if appid == "" {
		return nil, errors.New("decrypted callback is missing appid")
	}
	return append([]byte(nil), message...), nil
}

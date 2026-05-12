package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"sync"
)

// EncryptionService 加密服务
type EncryptionService struct {
	encryptionKey []byte
	gcm           cipher.AEAD
	once          sync.Once
	initErr       error
}

// NewEncryptionService 创建加密服务
func NewEncryptionService() (*EncryptionService, error) {
	es := &EncryptionService{}
	
	// 延迟初始化，获取密钥时再初始化
	return es, nil
}

// initGCM 初始化 GCM 模式
func (e *EncryptionService) initGCM() error {
	e.once.Do(func() {
		// 从环境变量获取加密密钥
		keyStr := os.Getenv("ENCRYPTION_KEY")
		if keyStr == "" {
			// 使用默认密钥（仅用于开发环境）
			keyStr = "ai-gateway-default-key-32-bytes!"
		}
		
		// 解码 Base64 密钥
		key, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			// 如果不是 Base64，直接使用字符串作为密钥
			key = []byte(keyStr)
		}
		
		// 确保密钥长度为 32 字节 (AES-256)
		if len(key) < 32 {
			// 填充到 32 字节
			padded := make([]byte, 32)
			copy(padded, key)
			key = padded
		} else if len(key) > 32 {
			key = key[:32]
		}
		
		e.encryptionKey = key
		
		// 创建 AES cipher
		block, err := aes.NewCipher(e.encryptionKey)
		if err != nil {
			e.initErr = fmt.Errorf("failed to create cipher: %w", err)
			return
		}
		
		// 创建 GCM 模式
		e.gcm, err = cipher.NewGCM(block)
		if err != nil {
			e.initErr = fmt.Errorf("failed to create GCM: %w", err)
			return
		}
	})
	
	return e.initErr
}

// Encrypt 加密明文
func (e *EncryptionService) Encrypt(plaintext string) (string, error) {
	if err := e.initGCM(); err != nil {
		return "", err
	}
	
	// 将明文转换为字节
	plaintextBytes := []byte(plaintext)
	
	// 创建随机 nonce
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// 加密数据
	ciphertext := e.gcm.Seal(nonce, nonce, plaintextBytes, nil)
	
	// 返回 Base64 编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密密文
func (e *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if err := e.initGCM(); err != nil {
		return "", err
	}
	
	// Base64 解码
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}
	
	// 检查长度
	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	// 分离 nonce 和实际密文
	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	
	// 解密
	plaintext, err := e.gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	
	return string(plaintext), nil
}

// GenerateKeyPrefix 生成密钥前缀用于显示
func GenerateKeyPrefix(apiKey string) string {
	if len(apiKey) <= 8 {
		return "sk-****"
	}
	return apiKey[:7] + "****" + apiKey[len(apiKey)-4:]
}

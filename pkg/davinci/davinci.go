//go:generate mockery --all --inpackage --case snake

package davinci

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/scrypt"
	"golang.org/x/crypto/sha3"
	"math/big"
	"strconv"
	"time"
)

type Engine struct {
}

func (dc Engine) GenerateHashValue(
	secretKey string,
	uniqueID string,
	bitLen int,
) (string, error) {
	secretByte, err := base32.StdEncoding.DecodeString(secretKey)
	if err != nil {

		return "", err
	}
	hash := hmac.New(sha3.New224, secretByte)
	_, err = hash.Write([]byte(uniqueID))
	if err != nil {

		return "", err
	}
	hmacBytes := hash.Sum(nil)

	if bitLen > 1 {
		return hex.EncodeToString(hmacBytes[:bitLen]), nil
	}

	return hex.EncodeToString(hmacBytes), nil
}

func (dc Engine) VerifyReferenceNumber(
	secretKey []byte,
	uniqueID string,
	referenceNumber string,
) bool {
	// Generate the hash value using the secret key and transaction identifier
	hash, err := dc.GenerateUniqueKey(secretKey, uniqueID)
	if err != nil {
		return false
	}

	// Compare the generated hash with the reference number
	return referenceNumber == hash
}

func (dc Engine) GenerateCodeTRX() string {
	const (
		prefix     = "TRX-"
		charset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		codeLength = 6
		charsetLen = int64(len(charset))
	)

	// Use the current timestamp to seed randomness.
	timestamp := time.Now().UnixNano()

	// Generate a random code.
	randomCode := make([]byte, codeLength)
	for i := range randomCode {
		// Generate a random index based on the charset length.
		index, err := rand.Int(rand.Reader, big.NewInt(charsetLen))
		if err != nil {
			panic("Failed to generate random index")
		}
		randomCode[i] = charset[index.Int64()]
	}

	return fmt.Sprintf("%s%s-%X", prefix, randomCode, timestamp)
}

func (dc Engine) GenerateUniqueKey(
	secretKey []byte,
	uniqueID string,
) (string, error) {
	return strconv.FormatInt(time.Now().UTC().Unix(), 16), nil
}

func (dc Engine) GenerateHash(secretKey []byte, uniqueID string) (string, error) {
	h := hmac.New(sha256.New, secretKey)
	_, err := h.Write([]byte(uniqueID))
	if err != nil {
		return "", err
	}
	hash := h.Sum(nil)

	// Truncate the hash to 17 bytes (34 characters)
	truncatedHash := hash[:17]

	// Convert the hash to a hex string
	hashStr := hex.EncodeToString(truncatedHash)

	return hashStr, nil
}

func DefaultDavinci() Engine {
	return Engine{}
}

func (dc Engine) GenerateOTPCode(
	secret string,
	counter uint64,
) (int, error) {
	counterByte := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		counterByte[i] = byte(counter & 0xff)
		counter >>= 8
	}

	secretByte, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return 0, fmt.Errorf("StdEncoding.DecodeString: %w", err)
	}
	hash := hmac.New(sha1.New, secretByte)
	_, err = hash.Write(counterByte)
	if err != nil {
		return 0, fmt.Errorf("hash.Write: %w", err)
	}
	hmacBytes := hash.Sum(nil)

	// "Dynamic truncation" in RFC 4226
	// http://tools.ietf.org/html/rfc4226#section-5.4
	offset := hmacBytes[len(hmacBytes)-1] & 0xf
	code := (int(hmacBytes[offset])&0x7f)<<24 |
		(int(hmacBytes[offset+1])&0xff)<<16 |
		(int(hmacBytes[offset+2])&0xff)<<8 |
		(int(hmacBytes[offset+3]) & 0xff)
	code = code % 1000000

	// padding the non 6-digits otp with zero value
	f := fmt.Sprintf("%%0%dd", 6)
	codeStr := fmt.Sprintf(f, code)
	newCode, err := strconv.ParseInt(codeStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(" strconv.ParseInt newcode: %w", err)
	}

	return int(newCode), nil
}

func (dc Engine) HashAndSalt(pwd []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (dc Engine) ComparePassword(hashedPwd string, pwd []byte) (bool, error) {
	HashedByte := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(HashedByte, pwd)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d Engine) DeriveKey(password, salt []byte) ([]byte, []byte, error) {
	if salt == nil {
		salt = make([]byte, 32)
		if _, err := rand.Read(salt); err != nil {
			return nil, nil, err
		}
	}

	key, err := scrypt.Key(password, salt, 32768, 8, 1, 32)
	if err != nil {

		return nil, nil, err
	}

	return key, salt, nil
}

func (d Engine) DecryptMessage(key []byte, p string) (string, error) {
	data, err := hex.DecodeString(p)
	if err != nil {
		return "", err
	}
	salt, data := data[len(data)-32:], data[:len(data)-32]

	key, _, err = d.DeriveKey(key, salt)
	if err != nil {

		return "", err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {

		return "", err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {

		return "", err
	}

	nonce, ciphertext := data[:gcm.NonceSize()], data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {

		return "", err
	}

	return string(plaintext), nil
}

func (d Engine) EncryptMessage(key, data []byte) (string, error) {
	key, salt, err := d.DeriveKey(key, nil)
	if err != nil {

		return "", err
	}

	blockCipher, err := aes.NewCipher(key)
	if err != nil {

		return "", err
	}

	gcm, err := cipher.NewGCM(blockCipher)
	if err != nil {

		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	ciphertext = append(ciphertext, salt...)

	return hex.EncodeToString(ciphertext), nil
}

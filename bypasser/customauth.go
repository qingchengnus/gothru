package bypasser

import (
	"crypto/des"
	"crypto/rand"
	"errors"
)

const (
	CipherTypeAES256 = iota
	CipherTypeSimple
)

type customAuthRequest struct {
	version    byte
	username   []byte
	password   []byte
	cipherType byte
}

type customAuthResponse struct {
	version       byte
	status        byte
	key           []byte
	initialVector []byte
}

type CustomAuthenticator struct {
}

const (
	AES256KEY = "Internat"
)

const (
	CipherTypeInUse = CipherTypeSimple
)

func (c CustomAuthenticator) HandleAuthentication(packet []byte, legalUsers map[string]string) ([]byte, GFWCipher, bool, error) {
	req, ok := formatCustomAuthRequest(packet)
	if !ok {
		return []byte{}, nil, false, errors.New("Invalid username/password packet.")
	}

	result := validateCustomCredentials(req, legalUsers)
	if result {
		if req.cipherType == CipherTypeSimple {
			randomKey := make([]byte, 1)
			rand.Read(randomKey)
			resp := customAuthResponse{req.version, validationStatusSuccess, randomKey, []byte{}}
			return parseCustomAuthResponse(resp), NewShiftCipher(randomKey[0]), true, nil
		} else if req.cipherType == CipherTypeAES256 {
			initialVector := make([]byte, des.BlockSize)
			rand.Read(initialVector)
			mCipher, err := NewAESCTRCipher([]byte(AES256KEY), initialVector)
			if err != nil {
				return []byte{}, nil, false, errors.New("Failed to generate AESCTRCipher." + err.Error())
			}
			resp := customAuthResponse{req.version, validationStatusSuccess, []byte(AES256KEY), initialVector}

			return parseCustomAuthResponse(resp), mCipher, true, nil
		} else {
			return []byte{}, nil, false, errors.New("Unknown cryption method.")
		}
	} else {
		resp := customAuthResponse{req.version, validationStatusFailure, []byte{}, []byte{}}
		return parseCustomAuthResponse(resp), nil, false, nil
	}
}

func parseCustomAuthResponse(resp customAuthResponse) []byte {
	result := []byte{resp.version, resp.status}
	result = append(result, byte(len(resp.key)))
	if len(resp.key) != 0 {
		result = append(result, resp.key...)
	}
	result = append(result, byte(len(resp.initialVector)))
	if len(resp.initialVector) != 0 {
		result = append(result, resp.initialVector...)
	}

	return result
}

func formatCustomAuthRequest(req []byte) (customAuthRequest, bool) {
	if !validateCustomAuthRequest(req) {
		return customAuthRequest{}, false
	} else {
		reqLen := len(req)
		uNameLen := int(req[1])
		return customAuthRequest{req[0], req[2 : 2+uNameLen], req[3+uNameLen : reqLen-1], req[reqLen-1]}, true
	}
}

func formatCustomAuthResponse(resp []byte) customAuthResponse {
	result := customAuthResponse{}
	result.version = resp[0]
	result.status = resp[1]
	result.key = resp[3 : 3+int(resp[2])]
	result.initialVector = resp[4+int(resp[2]):]
	return result
}

func parseCustomAuthRequest(req customAuthRequest) []byte {
	result := []byte{req.version}
	result = append(result, byte(len(req.username)))
	result = append(result, req.username...)
	result = append(result, byte(len(req.password)))
	result = append(result, req.password...)
	result = append(result, req.cipherType)
	return result
}

func validateCustomAuthRequest(req []byte) bool {
	if len(req) < 6 {
		return false
	} else {
		uNameLen := int(req[1])
		if len(req) < 4+uNameLen {
			return false
		}
		pWordLen := int(req[2+uNameLen])
		return (4 + uNameLen + pWordLen) == len(req)
	}
}

func validateCustomCredentials(req customAuthRequest, legalUsers map[string]string) bool {
	pword, ok := legalUsers[string(req.username)]
	if !ok {
		return false
	} else {
		return pword == string(req.password)
	}
}

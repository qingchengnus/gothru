package bypasser

import (
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
	AES256KEY = ""
	SHIFTKEY  = 0x05
)

func (c CustomAuthenticator) HandleAuthentication(packet []byte) ([]byte, GFWCipher, bool, error) {
	req, ok := formatCustomAuthRequest(packet)
	if !ok {
		return []byte{}, nil, false, errors.New("Invalid username/password packet.")
	}

	result := validateCustomCredentials(req)
	if result {
		if req.cipherType == CipherTypeSimple {
			resp := customAuthResponse{req.version, validationStatusSuccess, []byte{SHIFTKEY}, []byte{}}
			return parseCustomAuthResponse(resp), NewShiftCipher(SHIFTKEY), true, nil
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

func validateCustomCredentials(req customAuthRequest) bool {
	return true
}

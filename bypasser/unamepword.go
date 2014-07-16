package bypasser

import (
	"errors"
)

const (
	validationStatusSuccess = 0x00
)

type UnamePwordHandler struct {
}

type usernamePasswordRequest struct {
	version  byte
	username []byte
	password []byte
}

type usernamePasswordResponse struct {
	version byte
	status  byte
}

func validateCredentials(req usernamePasswordRequest) byte {
	return validationStatusSuccess
}

func handleUsernamePassword(req usernamePasswordRequest) usernamePasswordResponse {
	return usernamePasswordResponse{req.version, validateCredentials(req)}
}

func parseUsernamePasswordResponse(resp usernamePasswordResponse) []byte {
	return []byte{resp.version, resp.status}
}

func formatUsernamePasswordRequest(req []byte) (usernamePasswordRequest, bool) {
	if !validateUsernamePasswordRequest(req) {
		return usernamePasswordRequest{}, false
	} else {
		uNameLen := int(req[1])
		return usernamePasswordRequest{req[0], req[2 : 2+uNameLen], req[3+uNameLen:]}, true
	}
}

func validateUsernamePasswordRequest(req []byte) bool {
	if len(req) < 5 {
		return false
	} else {
		uNameLen := int(req[1])
		if len(req) < 3+uNameLen {
			return false
		}
		pWordLen := int(req[2+uNameLen])
		return (3 + uNameLen + pWordLen) == len(req)
	}
}

func (h UnamePwordHandler) Authenticate(p []byte) ([]byte, bool, error) {
	req, ok := formatUsernamePasswordRequest(p)
	if !ok {
		return []byte{}, false, errors.New("Invalid username/password packet.")
	}

	resp := handleUsernamePassword(req)
	return parseUsernamePasswordResponse(resp), resp.status == validationStatusSuccess, nil
}

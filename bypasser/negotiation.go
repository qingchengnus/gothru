package bypasser

type Authenticator interface {
	Authenticate(p []byte) ([]byte, bool, error)
}

const (
	AuthenticatingMethodNo               = 0x00
	AuthenticatingMethodGSSAPI           = 0x01
	AuthenticatingMethodUsernamePassword = 0x02
	AuthenticatingMethodCustom           = 0x0C
	AuthenticatingMethodNotSupported     = 0xff
)

type MethodSelectionRequest struct {
	Version      byte
	NumOfMethods byte
	Methods      []byte
}

type MethodSelectionResponse struct {
	Version        byte
	SelectedMethod byte
}

func getAuthenticatingMethod(req MethodSelectionRequest) byte {
	for _, methodRequested := range req.Methods {
		if isAuthenticatingMethodSupported(methodRequested) {
			return methodRequested
		}
	}
	return AuthenticatingMethodNotSupported
}

func isAuthenticatingMethodSupported(method byte) bool {
	if method == AuthenticatingMethodCustom || method == AuthenticatingMethodUsernamePassword || method == AuthenticatingMethodNo {
		return true
	} else {
		return false
	}
}

func HandleMethodSelection(req MethodSelectionRequest) MethodSelectionResponse {
	return MethodSelectionResponse{req.Version, getAuthenticatingMethod(req)}
}

func FormatMethodSelectionRequest(packet []byte) MethodSelectionRequest {
	return MethodSelectionRequest{packet[0], packet[1], packet[2:]}
}

func ParseMethodSelectionResponse(resp MethodSelectionResponse) []byte {
	return []byte{resp.Version, resp.SelectedMethod}
}

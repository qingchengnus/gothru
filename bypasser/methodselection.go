package bypasser

type MethodSelectionRequest struct {
	Version      byte
	NumOfMethods byte
	Methods      []byte
}

type MethodSelectionResponse struct {
	Version        byte
	SelectedMethod byte
}

func HandleMethodSelectionServer(req []byte) ([]byte, byte) {
	mRequest := formatMethodSelectionRequest(req)
	selectedMethod := getAuthenticatingMethod(mRequest, true)
	return parseMethodSelectionResponse(MethodSelectionResponse{mRequest.Version, selectedMethod}), selectedMethod
}

func HandleMethodSelectionClient(req []byte) []byte {
	mRequest := formatMethodSelectionRequest(req)
	selectedMethod := getAuthenticatingMethod(mRequest, false)
	if selectedMethod == AuthenticatingMethodNo {
		mRequest.Methods = []byte{AuthenticatingMethodCustom}
		mRequest.NumOfMethods = 0x01
		return parseMethodSelectionRequest(mRequest)
	} else {
		return req
	}
}

func formatMethodSelectionRequest(packet []byte) MethodSelectionRequest {
	return MethodSelectionRequest{packet[0], packet[1], packet[2:]}
}

func parseMethodSelectionResponse(resp MethodSelectionResponse) []byte {
	return []byte{resp.Version, resp.SelectedMethod}
}

func parseMethodSelectionRequest(req MethodSelectionRequest) []byte {
	return append([]byte{req.Version, req.NumOfMethods}, req.Methods...)
}

func getAuthenticatingMethod(req MethodSelectionRequest, isServer bool) byte {
	for _, methodRequested := range req.Methods {
		if isAuthenticatingMethodSupported(methodRequested, isServer) {
			return methodRequested
		}
	}
	return AuthenticatingMethodNotSupported
}

func isAuthenticatingMethodSupported(method byte, isServer bool) bool {
	if isServer {
		if method == AuthenticatingMethodCustom {
			return true
		} else {
			return false
		}
	} else {
		if method == AuthenticatingMethodNo {
			return true
		} else {
			return false
		}
	}
}

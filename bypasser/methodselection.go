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

func HandleMethodSelection(req []byte) ([]byte, byte) {
	mRequest := formatMethodSelectionRequest(req)
	selectedMethod := getAuthenticatingMethod(mRequest)
	return parseMethodSelectionResponse(MethodSelectionResponse{mRequest.Version, selectedMethod}), selectedMethod
}

func formatMethodSelectionRequest(packet []byte) MethodSelectionRequest {
	return MethodSelectionRequest{packet[0], packet[1], packet[2:]}
}

func parseMethodSelectionResponse(resp MethodSelectionResponse) []byte {
	return []byte{resp.Version, resp.SelectedMethod}
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
	if method == AuthenticatingMethodCustom {
		return true
	} else {
		return false
	}
}

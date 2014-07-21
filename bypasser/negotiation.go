package bypasser

import (
	"errors"
	"fmt"
	"net"
	"os"
)

const (
	INFO = iota
	DEBUG
	ERROR
)

const (
	statusMethodSelecting = iota
	statusSubNegotiating
	statusRequesting
)

const (
	minPacketLength = 3
	maxPacketLength = 257
)

const (
	AuthenticatingMethodNo               = 0x00
	AuthenticatingMethodGSSAPI           = 0x01
	AuthenticatingMethodUsernamePassword = 0x02
	AuthenticatingMethodCustom           = 0x0C
	AuthenticatingMethodNotSupported     = 0xff
)

type Authenticator interface {
	HandleAuthentication(packet []byte, legalUsers map[string]string) ([]byte, GFWCipher, bool, error)
}

var logger *GFWLogger = NewLogger(os.Stdout, []string{"INFO", "DEBUG", "ERROR"})

func HandleConnectionNegotiationServer(conn *net.TCPConn, users map[string]string) {
	logger.Log(DEBUG, "A new client connected.")
	status := statusMethodSelecting
	var method byte
	var mCipher GFWCipher
	for {
		data := make([]byte, maxPacketLength)
		numOfBytes, err := conn.Read(data)
		if err != nil {
			logger.Log(ERROR, "Connection closed due to failure to read data: "+err.Error())
			conn.Close()
			return
		}
		fmt.Println(data[:numOfBytes])
		packet, parseErr := parsePacket(data[:numOfBytes], status)
		if parseErr != nil {
			logger.Log(ERROR, "Connection closed due to invalid packet: "+parseErr.Error())
			conn.Close()
			return
		}

		switch status {
		case statusMethodSelecting:
			{
				logger.Log(DEBUG, "Handling method selection request.")
				var result []byte
				result, method = HandleMethodSelectionServer(packet)
				fmt.Println(result)
				_, err := conn.Write(result)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to write response: "+err.Error())
					conn.Close()
					return
				}
				if method == AuthenticatingMethodNotSupported {
					logger.Log(DEBUG, "Selected method is not supported, connection closed.")
					conn.Close()
					return
				} else {
					if method == AuthenticatingMethodNo {
						logger.Log(DEBUG, "Selected method is no authentication.")
						status = statusRequesting
					} else {
						logger.Log(DEBUG, "Selected method is some other supported authentication method.")
						status = statusSubNegotiating
					}

				}

			}
		case statusSubNegotiating:
			{
				logger.Log(DEBUG, "Handling authentication request.")
				authenticator := getAuthenticator(method)
				resp, futureCipher, ok, err := authenticator.HandleAuthentication(packet, users)
				mCipher = futureCipher
				if err != nil {
					logger.Log(ERROR, "Connection closed due to authentication error: "+err.Error())
					conn.Close()
					return
				}
				_, err = conn.Write(resp)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to write response: "+err.Error())
					conn.Close()
					return
				}
				if !ok {
					logger.Log(DEBUG, "Client failed to pass the authentication.")
					conn.Close()
					return
				}
				status = statusRequesting
			}
		case statusRequesting:
			{
				logger.Log(DEBUG, "Handling request.")
				resp, err := HandleRequestServer(packet, conn, mCipher)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to HandleRequest error: "+err.Error())
					conn.Close()
					return
				}
				_, err = conn.Write(resp)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to write response: "+err.Error())
					conn.Close()
					return
				}
				return
			}
		}
	}

}

func HandleConnectionNegotiationClient(conn *net.TCPConn, serverAddr *net.TCPAddr, uname, pword string) {
	logger.DisableTag(DEBUG)
	logger.DisableTag(ERROR)
	logger.Log(DEBUG, "A new client connected.")

	connToServer, err := net.DialTCP("tcp", nil, serverAddr)
	if err != nil {
		logger.Log(ERROR, "Connection closed due to failure to connect to server: "+err.Error())
		conn.Close()
		return
	}

	status := statusMethodSelecting
	var method byte
	var mCipher GFWCipher
	for {
		data := make([]byte, maxPacketLength)
		numOfBytes, err := conn.Read(data)
		if err != nil {
			logger.Log(ERROR, "Connection closed due to failure to read data: "+err.Error())
			connToServer.Close()
			conn.Close()
			return
		}
		packet, parseErr := parsePacket(data[:numOfBytes], status)
		if parseErr != nil {
			logger.Log(ERROR, "Connection closed due to invalid packet: "+parseErr.Error())
			connToServer.Close()
			conn.Close()
			return
		}

		switch status {
		case statusMethodSelecting:
			{
				logger.Log(DEBUG, "Handling method selection request.")
				var result []byte
				result = HandleMethodSelectionClient(packet)
				_, err := connToServer.Write(result)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to write to server: "+err.Error())
					connToServer.Close()
					conn.Close()
					return
				}
				logger.Log(DEBUG, "Method selection request sent to server.")
				resp := make([]byte, maxPacketLength)
				size, err := connToServer.Read(resp)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to read from server: "+err.Error())
					connToServer.Close()
					conn.Close()
					return
				}
				logger.Log(DEBUG, "Method selection response back.")
				method = resp[1]
				if method == AuthenticatingMethodNotSupported {
					logger.Log(DEBUG, "Selected method is not supported, connection closed.")
					connToServer.Close()
					conn.Close()
					return
				} else if method == AuthenticatingMethodCustom {
					logger.Log(DEBUG, "Selected method is custom.")
					customReq := customAuthRequest{0x05, []byte(uname), []byte(pword), CipherTypeSimple}
					packet := parseCustomAuthRequest(customReq)
					_, err := connToServer.Write(packet)
					if err != nil {
						logger.Log(ERROR, "Connection closed due to failure to write to server: "+err.Error())
						connToServer.Close()
						conn.Close()
						return
					}
					logger.Log(DEBUG, "Auth request sent.")
					resp := make([]byte, maxPacketLength)
					size, err := connToServer.Read(resp)
					if err != nil {
						logger.Log(ERROR, "Connection closed due to failure to read from server: "+err.Error())
						connToServer.Close()
						conn.Close()
						return
					}

					authResp := formatCustomAuthResponse(resp[:size])
					if authResp.status != validationStatusSuccess {
						logger.Log(DEBUG, "Connection closed due to incorrect credentials.")
						connToServer.Close()
						conn.Close()
						return
					}
					logger.Log(DEBUG, "Auth succeeded.")
					mCipher = NewShiftCipher(authResp.key[0])

					status = statusRequesting

					_, err = conn.Write([]byte{0x05, 0x00})
					if err != nil {
						logger.Log(ERROR, "Connection closed due to failure to write response: "+err.Error())
						connToServer.Close()
						conn.Close()
						return
					}

				} else {
					_, err = conn.Write(resp[:size])
					if err != nil {
						logger.Log(ERROR, "Connection closed due to failure to write response: "+err.Error())
						connToServer.Close()
						conn.Close()
						return
					}
					status = statusSubNegotiating
				}

			}
		case statusSubNegotiating:
			{
				_, err := connToServer.Write(packet)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to write to server: "+err.Error())
					connToServer.Close()
					conn.Close()
					return
				}
				resp := make([]byte, maxPacketLength)
				size, err := connToServer.Read(data)
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to read from server: "+err.Error())
					connToServer.Close()
					conn.Close()
					return
				}

				_, err = conn.Write(resp[:size])
				if err != nil {
					logger.Log(ERROR, "Connection closed due to failure to write to client: "+err.Error())
					connToServer.Close()
					conn.Close()
					return
				}

				status = statusRequesting
			}
		case statusRequesting:
			{
				logger.Log(DEBUG, "Handling request.")
				HandleRequestClient(packet, conn, connToServer, mCipher)
				return
			}
		}
	}

}

func getAuthenticator(method byte) Authenticator {
	switch method {
	case AuthenticatingMethodCustom:
		{
			return CustomAuthenticator{}
		}
	}
	return nil
}

func parsePacket(packet []byte, status int) ([]byte, error) {
	if !validatePacketLength(packet) {
		return []byte{}, errors.New("Invalid packet length.")
	}

	if !validateVersion(packet[0]) {
		return []byte{}, errors.New("Invalid version")
	}

	switch status {
	case statusMethodSelecting:
		{
			if !validateMethodSelectionPacket(packet) {
				return []byte{}, errors.New("Invalid method selection method.")
			} else {
				return packet, nil
			}
		}
	case statusSubNegotiating:
		{
			return packet, nil
		}
	case statusRequesting:
		{
			return packet, nil
		}
	default:
		{
			return []byte{}, errors.New("Unexpected packet.")
		}
	}
	return []byte{}, errors.New("Unexpected packet.")
}

func validatePacketLength(packet []byte) bool {
	pLen := len(packet)
	return pLen >= minPacketLength && pLen <= maxPacketLength
}

func validateVersion(v byte) bool {
	return v == 0x05
}

func validateMethodSelectionPacket(packet []byte) bool {
	return int(packet[1])+2 == len(packet)
}

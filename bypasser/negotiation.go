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
	HandleAuthentication(packet []byte) ([]byte, GFWCipher, bool, error)
}

var logger *GFWLogger = NewLogger(os.Stdout, []string{"INFO", "DEBUG", "ERROR"})

func HandleConnectionNegotiation(conn *net.TCPConn) {
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
				result, method = HandleMethodSelection(packet)
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
				resp, futureCipher, ok, err := authenticator.HandleAuthentication(packet)
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
				resp, err := HandleRequest(packet, conn, mCipher)
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

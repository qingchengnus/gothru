package main

import (
	"errors"
	//"fmt"
	"github.com/qingchengnus/gofw/bypasser"
	"net"
)

const (
	minPacketLength = 3
	maxPacketLength = 128
)

const (
	statusMethodSelecting = iota
	statusSubNegotiating  = iota
	statusRequesting      = iota
	statusConnecting      = iota
)

const (
	packetTypeMethodSelection = iota
	packetTypeSubNegotiation  = iota
	packetTypeRequest         = iota
)

func main() {
	listenAddress, _ := net.ResolveTCPAddr("tcp", ":18888")
	ln, err := net.ListenTCP("tcp", listenAddress)
	if err != nil {
		log(err.Error(), 0)
	} else {
		for {
			conn, err := ln.AcceptTCP()
			log("A client is trying to connect.", 0)
			if err != nil {
				log("A client failed to connect.", 0)
				log(err.Error(), 0)
				continue
			}
			go handleConnection(conn)
		}
	}
}

func handleConnection(conn *net.TCPConn) {
	log("A client connected.", 0)
	status := statusMethodSelecting
	var method byte
	for {
		data := make([]byte, maxPacketLength)
		numOfBytes, err := conn.Read(data)
		result := make([]byte, numOfBytes)
		bypasser.Decrypt(result, data[:numOfBytes], bypasser.EncryptMethodSimple)
		log("Receiving data.", 1)
		if err != nil {
			conn.Close()
			log("Failed to read data, connection closed.", 1)
			return
		}
		packet, parseErr := parsePacket(result, numOfBytes, status)
		if parseErr != nil {
			log("Invalid packet, connection closed. "+parseErr.Error(), 1)
			conn.Close()
			return
		}

		switch status {
		case statusMethodSelecting:
			{
				log("Handling method selection request.", 2)
				methodSelectionRequest := bypasser.FormatMethodSelectionRequest(packet)
				methodSelectionResponse := bypasser.HandleMethodSelection(methodSelectionRequest)
				result := bypasser.ParseMethodSelectionResponse(methodSelectionResponse)
				bypasser.Encrypt(result, result, bypasser.EncryptMethodSimple)
				_, err := conn.Write(result)
				if err != nil {
					conn.Close()
					log("Fail to write response back, connection closed.", 3)
					return
				}
				method = methodSelectionResponse.SelectedMethod
				if method == bypasser.AuthenticatingMethodNotSupported {
					conn.Close()
					log("Selected method is not supported, connection closed.", 3)
					return
				} else {
					if method == bypasser.AuthenticatingMethodNo {
						log("Selected method is no authentication.", 3)
						status = statusRequesting
					} else {
						log("Selected method is some other supported authentication method.", 3)
						status = statusSubNegotiating
					}

				}

			}
		case statusSubNegotiating:
			{
				authenticator := getAuthenticator(method)
				resp, ok, err := authenticator.Authenticate(packet)
				if err != nil {
					conn.Close()
					return
				}
				conn.Write(resp)
				if !ok {
					conn.Close()
					return
				}
				status = statusRequesting
			}
		case statusRequesting:
			{
				log("Handling request.", 2)
				resp, err := bypasser.HandleRequest(packet, conn)
				if err != nil {
					conn.Close()
					log("Failed to handle request, connection closed.", 2)
					log(err.Error(), 3)
					return
				}
				bypasser.Encrypt(resp, resp, bypasser.EncryptMethodSimple)
				conn.Write(resp)
				status = statusConnecting
				//conn.Close()
				return
			}
		}
	}

}

func parsePacket(packet []byte, length int, status int) ([]byte, error) {
	// if length < minPacketLength || length > maxPacketLength {
	// 	return []byte{}, errors.New("Invalid packet length.")
	// }

	if status == statusMethodSelecting {
		if !validateVersion(packet[0]) {
			return []byte{}, errors.New("Invalid version")
		}

		if !validateMethodSelectionPacket(packet, length) {
			return []byte{}, errors.New("Invalid method selection method.")
		}

		return packet[0:length], nil
	} else if status == statusSubNegotiating {
		return packet[0:length], nil
	} else if status == statusRequesting {
		return packet[0:length], nil
	} else if status == statusConnecting {
		return packet[0:length], nil
	}
	return []byte{}, errors.New("Unexpected packet.")

}

func validateVersion(v byte) bool {
	return v == 0x05
}

func validateMethodSelectionPacket(packet []byte, length int) bool {
	return int(packet[1])+2 == length
}

func getAuthenticator(authMethod byte) bypasser.Authenticator {
	if authMethod == bypasser.AuthenticatingMethodUsernamePassword {
		return bypasser.UnamePwordHandler{}
	}
	return nil
}

func log(msg string, lvl int) {
	// blank := ""
	// for i := 0; i < lvl; i++ {
	// 	blank += "   "
	// }
	//fmt.Println("GoFWBypasser:", blank, msg)
}

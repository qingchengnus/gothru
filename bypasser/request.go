package bypasser

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
)

const (
	reserved = 0x00
)

const (
	bufferSize       = 8192
	minRequestLength = 10
)

const (
	commandConnect      = 0x01
	commandBind         = 0x02
	commandUDPAssociate = 0x03
)

const (
	addressTypeIPv4       = 0x01
	addressTypeDomainName = 0x03
	addressTypeIPv6       = 0x04
)

const (
	replySucceeded                 = 0x00
	replyGeneralSOCKSServerFailure = 0x01
	replyConnectionNotAllowed      = 0x02
	replyNetworkUnreachable        = 0x03
	replyHostUnreachable           = 0x04
	replyConnectionRefused         = 0x05
	replyTTLExpired                = 0x06
	replyCommandNotSupported       = 0x07
	replyAddressTypeNotSupported   = 0x08
)

type request struct {
	version            byte
	command            byte
	rsv                byte
	addressType        byte
	destinationAddress []byte
	destinationPort    [2]byte
}

type response struct {
	version      byte
	reply        byte
	rsv          byte
	addressType  byte
	boundAddress []byte
	boundPort    [2]byte
}

func HandleRequestServer(rqst []byte, conn *net.TCPConn, mCipher GFWCipher, username string, udata chan DataUsage) ([]byte, error) {
	req, ok := formatRequest(rqst)
	if !ok {
		return []byte{}, errors.New("Invalid request packet.")
	}
	switch req.command {
	case commandConnect:
		{
			logger.Log(DEBUG, "Request to connect.")
			return parseResponse(handleConnect(req, conn, mCipher, username, udata)), nil
		}
	case commandBind:
		{
			return parseResponse(generateFailureResponse(req.version, replyCommandNotSupported)), nil
		}
	case commandUDPAssociate:
		{
			return parseResponse(generateFailureResponse(req.version, replyCommandNotSupported)), nil
		}
	default:
		{
			return parseResponse(generateFailureResponse(req.version, replyCommandNotSupported)), nil
		}
	}
}

func HandleRequestClient(rqst []byte, conn, connToServer *net.TCPConn, mCipher GFWCipher) {
	req, ok := formatRequest(rqst)
	if ok {
		var method string
		var address string
		switch req.command {
		case commandConnect:
			{
				method = "CONNECT "
			}
		case commandBind:
			{
				method = "BIND "
			}
		case commandUDPAssociate:
			{
				method = "UDPASSOCIATE "
			}
		default:
			{
				method = "UNKNOWN "
			}
		}
		switch req.addressType {
		case addressTypeIPv4:
			{
				address = net.IP(req.destinationAddress).String()
			}
		case addressTypeIPv6:
			{
				address = net.IP(req.destinationAddress).String()
			}
		case addressTypeDomainName:
			{
				address = string(req.destinationAddress[1:]) + ":" + formatPort(req.destinationPort)
			}
		default:
			{
				address = "unknown address type."
			}
		}

		logger.Log(INFO, method+address)

	}
	mCipher.Encrypt(rqst, rqst)
	_, err := connToServer.Write(rqst)
	if err != nil {
		logger.Log(ERROR, "Connection closed due to failure to write to server: "+err.Error())
		connToServer.Close()
		conn.Close()
		return
	}
	resp := make([]byte, maxPacketLength)
	size, err := connToServer.Read(resp)
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

	if resp[1] == replySucceeded {
		buildTunnel(conn, connToServer, mCipher, "", nil)
	} else {
		connToServer.Close()
		conn.Close()
	}

}

func handleConnect(req request, conn *net.TCPConn, cipher GFWCipher, username string, udata chan DataUsage) response {
	switch req.addressType {
	case addressTypeIPv4:
		{
			logger.Log(DEBUG, "REQUEST TO IPv4.")
			return startTcpConnectSession(req.version, req.destinationAddress, addressTypeIPv4, req.destinationPort, conn, cipher, username, udata)
		}
	case addressTypeDomainName:
		{
			logger.Log(INFO, "REQUEST TO "+string(req.destinationAddress))
			return startTcpConnectSession(req.version, req.destinationAddress, addressTypeDomainName, req.destinationPort, conn, cipher, username, udata)
		}
	default:
		{
			return generateFailureResponse(req.version, replyAddressTypeNotSupported)
		}
	}
}

func generateFailureResponse(version byte, reply byte) response {
	return response{version, reply, reserved, 0, []byte{}, [2]byte{0, 0}}
}

func startTcpConnectSession(version byte, addr []byte, addrType byte, port [2]byte, conn *net.TCPConn, cipher GFWCipher, username string, udata chan DataUsage) response {
	var addrString string
	if addrType == addressTypeDomainName {
		addrString = string(addr[1:])
	} else {
		addrString = net.IP(addr).String()
	}
	targetAddr, err := net.ResolveTCPAddr("tcp", addrString+":"+formatPort(port))
	if err != nil {
		resp := generateFailureResponse(version, replyGeneralSOCKSServerFailure)
		return resp
	}
	connToTarget, err := net.DialTCP("tcp", nil, targetAddr)
	if err != nil {
		resp := generateFailureResponse(version, replyHostUnreachable)
		return resp
	} else {
		ipAddr, portNumber, _ := net.SplitHostPort(connToTarget.LocalAddr().String())
		ipAddrBytes := net.ParseIP(ipAddr)
		var addrType byte
		if ipAddrBytes.To4() != nil {
			ipAddrBytes = ipAddrBytes.To4()
			addrType = addressTypeIPv4
		} else {
			addrType = addressTypeIPv6
		}
		logger.Log(DEBUG, "Building tunnel.")
		go buildTunnel(connToTarget, conn, cipher, username, udata)
		return response{version, replySucceeded, reserved, addrType, ipAddrBytes, parsePort(portNumber)}
		// localAddr, _ := net.ResolveTCPAddr("tcp", ":0")
		// ln, err := net.ListenTCP("tcp", localAddr)
		// if err != nil {
		// 	log("Failed to get the listener to listen client.", 5)
		// 	resp := generateFailureResponse(version, replyGeneralSOCKSServerFailure)
		// 	return resp
		// } else {
		// 	log("Listener created.", 5)
		// 	networkAddr := ln.Addr().String()
		// 	//networkAddr := connToTarget.LocalAddr().String()
		// 	log("Local address is: "+networkAddr+".", 5)
		// 	ipAddr, portNumber, splitError := net.SplitHostPort(networkAddr)
		// 	if splitError != nil {
		// 		log("Failed to split the listen's local address.", 5)
		// 		resp := generateFailureResponse(version, replyGeneralSOCKSServerFailure)
		// 		return resp
		// 	} else {
		// 		ipAddr = "127.0.0.1"
		// ipAddrBytes := net.ParseIP(ipAddr)
		// var addrType byte
		// if ipAddrBytes.To4() != nil {
		// 	log("Bound ip address is an IPv4 address", 5)
		// 	addrType = addressTypeIPv4
		// } else {
		// 	log("Bound ip address is an IPv6 address", 5)
		// 	addrType = addressTypeIPv6
		// }
		// log("Bound port is: "+portNumber, 5)
		// 		resp := response{version, replySucceeded, reserved, addrType, ipAddrBytes, parsePort(portNumber)}

		// 		go waitClientToConnect(ln, connToTarget)
		// 		return resp

		// 	}
		// }
	}
}

// func waitClientToConnect(ln *net.TCPListener, connToTarget *net.TCPConn) {
// 	log("Waiting client to connect.", 5)
// 	connToClient, err := ln.AcceptTCP()
// 	log("Client connected.", 6)
// 	if err != nil {
// 		log("Client connected, but error occurs: "+err.Error(), 6)
// 		//break
// 	} else {
// 		log("Start building tunnel.", 6)
// 		go BuildTunnel(connToTarget, connToClient)
// 	}
// }

func buildTunnel(fromTarget, toClient *net.TCPConn, cipher GFWCipher, username string, udata chan DataUsage) {
	tunnelForward := make(chan []byte)
	tunnelBackward := make(chan []byte)
	errorChannelForward := make(chan error)
	errorChannelBackward := make(chan error)
	go handleTunnel(fromTarget, tunnelForward, tunnelBackward, errorChannelForward, errorChannelBackward, true, cipher, username, udata)
	go handleTunnel(toClient, tunnelBackward, tunnelForward, errorChannelBackward, errorChannelForward, false, cipher, "", nil)

	// go func() {
	// 	for {
	// 		io.Copy(fromTarget, toClient)
	// 	}
	// }()
	// go func() {
	// 	for {
	// 		io.Copy(toClient, fromTarget)
	// 	}
	// }()
}

func handleTunnel(target *net.TCPConn, receiver <-chan []byte, sender chan<- []byte, errorChannelF <-chan error, errorChannelB chan<- error, shouldEncrypt bool, cipher GFWCipher, username string, udata chan DataUsage) {
	errChan := make(chan error)
	dataChan := make(chan []byte)
	go func(dch chan []byte, ech chan error) {
		for {
			buf := make([]byte, bufferSize)
			//target.SetReadDeadline(time.Now())
			length, err := target.Read(buf)
			if err != nil {
				ech <- err
				return
			} else {
				if cipher != nil {
					if shouldEncrypt {
						cipher.Encrypt(buf[:length], buf[:length])
						if username != "" {
							dataUsed := DataUsage{username, int64(length)}
							udata <- dataUsed
						}
					} else {
						cipher.Decrypt(buf[:length], buf[:length])
					}
				}
				dch <- buf[:length]
			}

		}
	}(dataChan, errChan)

	for {
		select {
		case data := <-dataChan:
			sender <- data
		case err := <-errChan:
			errorChannelB <- err
			target.Close()
			return
		case <-errorChannelF:
			target.Close()
			return
		case data := <-receiver:
			target.Write(data)
		}
	}

}

func parsePort(p string) [2]byte {
	portNumber, _ := strconv.ParseUint(p, 10, 16)
	var result [2]byte
	binary.BigEndian.PutUint16(result[0:2], uint16(portNumber))
	return result
}

func formatPort(p [2]byte) string {
	return strconv.FormatUint(uint64(binary.BigEndian.Uint16(p[0:2])), 10)
}

func formatRequest(req []byte) (request, bool) {
	reqLen := len(req)
	if reqLen < minRequestLength {
		return request{}, false
	} else {
		rqst, ok := request{req[0], req[1], req[2], req[3], req[4 : reqLen-2], [2]byte{req[reqLen-2], req[reqLen-1]}}, true
		if rqst.addressType == addressTypeDomainName {
			//log()
		}
		return rqst, ok
	}
}

func parseResponse(resp response) []byte {
	//return []byte{resp.version, resp.reply, resp.rsv, resp.addressType, 0, 0, 0, 0, 0, 0}
	return append(append([]byte{resp.version, resp.reply, resp.rsv, resp.addressType}, resp.boundAddress...), resp.boundPort[:]...)
}

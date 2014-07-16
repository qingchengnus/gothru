package bypasser

import (
	//"bytes"
	"encoding/binary"
	"errors"
	//"fmt"
	//"io"
	"net"
	"strconv"
	//"time"
)

const (
	bufferSize       = 4096
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

func HandleRequest(rqst []byte, conn *net.TCPConn) ([]byte, error) {
	req, ok := formatRequest(rqst)
	if !ok {
		return []byte{}, errors.New("Invalid request packet.")
	}
	switch req.command {
	case commandConnect:
		{
			log("Handling command connect.", 3)
			return parseResponse(handleConnect(req, conn)), nil
		}
	default:
		{
			log("Command not supported.", 3)
			return parseResponse(generateFailureResponse(req.version, replyCommandNotSupported)), nil
		}
	}
}

func handleConnect(req request, conn *net.TCPConn) response {
	switch req.addressType {
	case addressTypeIPv4:
		{
			log("Address type is IPv4, start connecting.", 4)
			return startTcpConnectSession(req.version, req.destinationAddress, addressTypeIPv4, req.destinationPort, conn)
		}
	case addressTypeDomainName:
		{
			log("Address type is Domain name, start connecting.", 4)
			//log("Requested domain name is: "+string(req.destinationAddress[1:]), 4)
			//return generateFailureResponse(req.version, replyAddressTypeNotSupported)
			return startTcpConnectSession(req.version, req.destinationAddress, addressTypeDomainName, req.destinationPort, conn)
		}
	default:
		{
			log("Address type is not supported: "+strconv.Itoa(int(req.addressType)), 4)
			return generateFailureResponse(req.version, replyAddressTypeNotSupported)
		}
	}
}

func generateFailureResponse(version byte, reply byte) response {
	return response{version, reply, reserved, 0, []byte{}, [2]byte{0, 0}}
}

func startTcpConnectSession(version byte, addr []byte, addrType byte, port [2]byte, conn *net.TCPConn) response {
	var addrString string
	if addrType == addressTypeDomainName {
		addrString = string(addr[1:])
	} else {
		addrString = net.IP(addr).String()
	}
	log("Start to resolve the address: "+addrString+":"+formatPort(port), 5)
	targetAddr, err := net.ResolveTCPAddr("tcp", addrString+":"+formatPort(port))
	if err != nil {
		log("Fail to resolve tcp address.", 5)
		resp := generateFailureResponse(version, replyGeneralSOCKSServerFailure)
		return resp
	}
	connToTarget, err := net.DialTCP("tcp", nil, targetAddr)
	if err != nil {
		log("Failed to connect to the target.", 5)
		resp := generateFailureResponse(version, replyHostUnreachable)
		return resp
	} else {
		log("Connected to target.", 5)
		ipAddr, portNumber, _ := net.SplitHostPort(connToTarget.LocalAddr().String())
		ipAddrBytes := net.ParseIP(ipAddr)
		var addrType byte
		if ipAddrBytes.To4() != nil {
			log("Bound ip address is an IPv4 address", 5)
			ipAddrBytes = ipAddrBytes.To4()
			addrType = addressTypeIPv4
		} else {
			log("Bound ip address is an IPv6 address", 5)
			addrType = addressTypeIPv6
		}
		log("Bound port is: "+portNumber, 5)
		go buildTunnel(connToTarget, conn)
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

func buildTunnel(fromTarget, toClient *net.TCPConn) {
	// defer fromTarget.Close()

	// if _, err := fromTarget.Write(buf.Bytes()); err != nil {
	// 	panic(err)
	// }

	// data := make([]byte, 1024)
	// n, err := fromTarget.Read(data)
	// if err != nil {
	// 	if err != io.EOF {
	// 		panic(err)
	// 	} else {
	// 		toClient.Write(data[:n])

	// 	}
	// }
	// toClient.Close()
	tunnelForward := make(chan []byte)
	tunnelBackward := make(chan []byte)
	errorChannelForward := make(chan error)
	errorChannelBackward := make(chan error)
	go handleTunnel(fromTarget, tunnelForward, tunnelBackward, errorChannelForward, errorChannelBackward)
	go handleTunnel(toClient, tunnelBackward, tunnelForward, errorChannelBackward, errorChannelForward)

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

func handleTunnel(target *net.TCPConn, receiver <-chan []byte, sender chan<- []byte, errorChannelF <-chan error, errorChannelB chan<- error) {
	errChan := make(chan error)
	dataChan := make(chan []byte)
	go func(dch chan []byte, ech chan error) {
		for {
			// buf := &bytes.Buffer{}
			// for {
			// 	data := make([]byte, 256)
			// 	n, err := target.Read(data)
			// 	if err != nil {
			// 		if err == io.EOF {
			// 			log("End of file encountered: "+err.Error(), 6)
			// 			break
			// 		} else {
			// 			log("Other error encountered: "+err.Error(), 6)
			// 			ech <- err
			// 			return
			// 		}

			// 	}
			// 	buf.Write(data[:n])
			// 	errChan <- errors.New("")
			// 	if data[n-2] == 13 && data[n-1] == 10 {
			// 		log("End of file encountered.", 6)
			// 		break
			// 	}
			// }
			// dch <- buf.Bytes()
			// return
			buf := make([]byte, bufferSize)
			//target.SetReadDeadline(time.Now())
			length, err := target.Read(buf)
			if err != nil {
				log("Error encountered: "+err.Error(), 6)
				ech <- err
				return
			} else {
				//target.SetReadDeadline(time.Time{})
				log(strconv.FormatInt(int64(length), 10)+" bytes of data received, sent to data channel.", 6)

				dch <- buf[:length]
			}

		}
	}(dataChan, errChan)

	for {
		select {
		case data := <-dataChan:
			//log("Data received: "+string(data), 6)
			log("Data received from data channel, pass it to sender.", 6)
			sender <- data
		case err := <-errChan:
			// handle our error then exit for loop
			log("Error received from error channel: "+err.Error()+", notify the other routine to stop.", 6)
			errorChannelB <- err
			log("Tunnel ended.", 6)
			target.Close()
			return
		case err := <-errorChannelF:
			log("Error received from the other routine: "+err.Error()+", closing the connection.", 6)
			log("Tunnel ended.", 6)
			target.Close()
			return
		case data := <-receiver:
			log("Data received from receiver.", 6)
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

func log(msg string, lvl int) {
	blank := ""
	for i := 0; i < lvl; i++ {
		blank += "   "
	}
	fmt.Println("GoFWBypasser:", blank, msg)
}

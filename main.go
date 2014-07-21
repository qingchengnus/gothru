package main

import (
	"encoding/xml"
	"errors"
	"flag"
	"fmt"
	"github.com/qingchengnus/gofw/bypasser"
	"net"
	"os"
	"path/filepath"
)

type clientConfig struct {
	Server_address string `xml:"server_address"`
	Server_port    string `xml:"server_port"`
	Local_port     string `xml:"local_port"`
	Username       string `xml:"username"`
	Password       string `xml:"password"`
}

type serverConfig struct {
	Server_port string `xml:"server_port"`
	Users       []user `xml:"user"`
}

type user struct {
	Username string `xml:"username"`
	Password string `xml:"password"`
}

const (
	defaultString                = "DefAuLTsTRinG"
	shortHandTip                 = " (shorthand)"
	defaultLocalPort             = "16666"
	defaultServerPort            = "18888"
	localPortUsage               = "Set the local port of the client if no config file is present, default port is " + defaultLocalPort + "."
	paramsNotSetError            = "You need to set username, password, server address and localport(optional) when you do not have a config file. Enter gofw -h to see how to set them."
	localPortBindingError        = "Cannot create local listener due to: "
	serverAddrResolvingError     = "Cannot resolve server address due to: "
	localPortResolvingError      = "Cannot resolve local address due to: "
	configFilePathUsage          = "Set the path for configuration file explicitly."
	defaultConfigFileName        = "config.xml"
	loadingConfigFileError       = "Failed to load the config file due to: "
	clientConfigFileInvalidError = "Config file format is not correct, or some of the mandatory parameters are missing."
	serverConfigFileInvalidError = "Config file format is not correct, or you do not have a user."
	serverUsage                  = "Run as a server with server flag on."
	serverListenError            = "Failed to listen due to: "
)

const (
	startMethodDefault = iota
	startMethodConfig
	startMethodError
)

var server bool

func init() {
	flag.BoolVar(&server, "server", false, serverUsage)
	flag.BoolVar(&server, "s", false, serverUsage+shortHandTip)
}

var configFilePath string

func init() {
	flag.StringVar(&configFilePath, "config", defaultString, configFilePathUsage)
	flag.StringVar(&configFilePath, "c", defaultString, configFilePathUsage+shortHandTip)
}

func main() {
	flag.Parse()
	method := checkFlags()
	var realConfigFilePath string
	switch method {
	case startMethodConfig:
		{
			realConfigFilePath = configFilePath
		}
	case startMethodDefault:
		{
			defaultConfigFilePath, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			realConfigFilePath = filepath.Join(defaultConfigFilePath, defaultConfigFileName)
		}
	}

	if server {
		serverPort, users, err := parseServerConfigFile(realConfigFilePath)
		if err != nil {
			fmt.Println(loadingConfigFileError + err.Error())
			return
		}

		listenAddress, _ := net.ResolveTCPAddr("tcp", ":"+serverPort)
		ln, err := net.ListenTCP("tcp", listenAddress)
		if err != nil {
			fmt.Println(serverListenError + err.Error())
		} else {
			for {
				conn, err := ln.AcceptTCP()
				if err != nil {
					continue
				}
				go bypasser.HandleConnectionNegotiationServer(conn, users)
			}
		}

	} else {
		serverAddr, localPort, username, password, err := parseClientConfigFile(realConfigFilePath)
		if err != nil {
			fmt.Println(loadingConfigFileError + err.Error())
			return
		}

		serverAddress, err := net.ResolveTCPAddr("tcp", serverAddr)
		if err != nil {
			fmt.Println(serverAddrResolvingError + err.Error())
			return
		}
		listenAddress, err := net.ResolveTCPAddr("tcp", ":"+localPort)
		if err != nil {
			fmt.Println(localPortResolvingError + err.Error())
			return
		}
		ln, err := net.ListenTCP("tcp", listenAddress)
		if err != nil {
			fmt.Println(localPortBindingError + err.Error())
		} else {
			fmt.Println("Client started listening on port: " + localPort)
			for {
				conn, err := ln.AcceptTCP()
				if err != nil {
					continue
				}
				go bypasser.HandleConnectionNegotiationClient(conn, serverAddress, username, password)
			}
		}
	}
}

func checkFlags() int {
	if flag.Parsed() {
		if configFilePath != defaultString {
			return startMethodConfig
		} else {
			return startMethodDefault
		}

	} else {
		flag.Parse()
		return checkFlags()
	}
}

func parseClientConfigFile(filePath string) (string, string, string, string, error) {
	xmlFile, err := os.Open(filePath)
	if err != nil {
		return "", "", "", "", err
	}
	defer xmlFile.Close()
	decoder := xml.NewDecoder(xmlFile)
	var c clientConfig
	err = decoder.Decode(&c)
	if err != nil {
		return "", "", "", "", err
	} else {
		if c.Server_address == "" || c.Server_port == "" || c.Username == "" || c.Password == "" {
			return "", "", "", "", errors.New(clientConfigFileInvalidError)
		}
		localPort := defaultLocalPort
		if c.Local_port != "" {
			localPort = c.Local_port
		}
		return c.Server_address + ":" + c.Server_port, localPort, c.Username, c.Password, nil
	}
}

func parseServerConfigFile(filePath string) (string, map[string]string, error) {
	xmlFile, err := os.Open(filePath)
	if err != nil {
		return "", nil, err
	}
	defer xmlFile.Close()
	decoder := xml.NewDecoder(xmlFile)
	var s serverConfig
	err = decoder.Decode(&s)
	if err != nil {
		return "", nil, err
	} else {
		if len(s.Users) == 0 {
			return "", nil, errors.New(serverConfigFileInvalidError)
		}
		serverPort := defaultServerPort
		if s.Server_port != "" {
			serverPort = s.Server_port
		}
		usersMap := make(map[string]string)
		for _, u := range s.Users {
			usersMap[u.Username] = u.Password
		}

		return serverPort, usersMap, nil
	}
}

package ftpserver

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
)

func (c *Client) getLocalIpv4Addr() (string, error) {
	ip := strings.Split(c.CmdConn.LocalAddr().String(), ":")[0]

	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "", errors.New("invalid Pasv Ip")
	}

	parsedIP = parsedIP.To4()

	if parsedIP == nil {
		return "", fmt.Errorf("invalid IPv4 passive IP %#v", ip)
	}

	return parsedIP.String(), nil
}

//被动模式(服务器需要建立一个监听第二通道连接)
func (c *Client) handlePASV(param string) error {
	l, err := net.Listen("tcp", "")
	if err != nil {
		return err
	}

	ip, err := c.getLocalIpv4Addr()
	if err != nil {
		return err
	}

	port := l.Addr().(*net.TCPAddr).Port
	msg := fmt.Sprintf("Entering Passive Mode (%s,%d,%d).",
		strings.ReplaceAll(ip, ".", ","),
		port/256,
		port%256)

	return c.reply(StatusEnteringPASV, msg)
}

//主动模式
func (c *Client) handlePORT(param string) error {
	params := strings.Split(param, ",")

	ip := strings.Join(params[:4], ".")
	p1, _ := strconv.Atoi(params[4])
	p2, _ := strconv.Atoi(params[5])

	port := p1<<8 + p2

	return c.reply(StatusOK, fmt.Sprintf("Entering Active Mode, %s:%d.", ip, port))
}

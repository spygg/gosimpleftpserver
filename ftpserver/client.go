package ftpserver

import (
	"fmt"
	"net"
	"strings"
)

const (
	maxCommandSize = 2048
)

func (c *Client) handleClient() error {
	//先发送欢迎消息
	c.reply(StatusServiceReady, "Welcome to Go SimpleFtpServer.")

	//循环解析命令
	for {
		data := make([]byte, maxCommandSize)
		n, err := c.CmdConn.Read(data)

		if err != nil {
			return err
		}

		if n <= 2 || data[n-2] != '\r' || data[n-1] != '\n' {
			fmt.Println("bad command")
			continue
		}

		data = data[:n-2]

		c.parseCommand(string(data))
	}

}

func parseLine(line string) (string, string) {
	params := strings.SplitN(line, " ", 2)
	if len(params) == 1 {
		return params[0], ""
	}

	return params[0], params[1]
}

//解析命令
func (c *Client) parseCommand(line string) {
	command, param := parseLine(line)
	command = strings.ToUpper(command)

	fmt.Println(command, param)

	cmdDesc := commandsMap[command]

	//检查是否实现该命令
	if cmdDesc == nil {
		c.reply(StatusCommandNotImplemented,
			fmt.Sprintf(`Command not implemented "%s".`, command))

		fmt.Printf(`Command not implemented (%s)\n`, command)
		return
	}

	//检查是否需要登录
	if !cmdDesc.NoNeedAuthentication && !c.LoggedIn {
		c.reply(StatusNotLoggedIn,
			fmt.Sprintf(`You must log in first (%s)".`, command))
		return
	}

	//检查是否是只读权限
	if c.server.ReadOnly && cmdDesc.NeedVerifyWritePermission {
		c.reply(StatusActionNotTaken, "Can't do that in read-only mode.")
		return
	}

	//通过第二通道传输的
	if cmdDesc.TransferRelated {
		//将数据传输置位
		c.transferWg.Add(1)

		go func(cmd, param string) {
			//先入栈后执行
			defer c.closeDataConnection()

			//等待传输完成
			defer c.transferWg.Done()

			c.executeCommandFn(cmdDesc, cmd, param)
		}(command, param)

	} else {
		c.executeCommandFn(cmdDesc, command, param)
	}
}

func (c *Client) executeCommandFn(cmdDesc *CommandDescription, command, param string) {
	// Let's prepare to recover in case there's a command error
	defer func() {
		if r := recover(); r != nil {
			c.reply(StatusSyntaxErrorNotRecognised,
				fmt.Sprintf("Unhandled internal error: %s", r))

			fmt.Println(
				"Internal command handling error",
				"err", r,
				"command", command,
				"param", param,
			)
		}
	}()

	if err := cmdDesc.Fn(c, param); err != nil {
		c.reply(StatusSyntaxErrorNotRecognised, fmt.Sprintf("Error: %s", err))
	}
}

func (c *Client) replyString(msg string) error {
	data := fmt.Sprintf("%s\r\n", msg)

	_, err := c.CmdConn.Write([]byte(data))

	return err
}

//发送命令
func (c *Client) reply(stateCode int, msg string) error {
	data := fmt.Sprintf("%d %s\r\n", stateCode, msg)

	_, err := c.CmdConn.Write([]byte(data))

	return err
}

func (c *Client) setupDataConnection() error {
	l, err := net.Listen("tcp", "")

	if err != nil {
		fmt.Println("setupDataConnection error: ", err, l)
	}

	for {

	}
}

func (c *Client) closeDataConnection() error {
	//等待传输完毕
	c.transferWg.Wait()

	if c.DataConn != nil {
		err := c.DataConn.Close()
		if err != nil {
			fmt.Println("data connection close error: ", err)
		}

		c.DataConn = nil
		return err
	}

	return nil
}

func (c *Client) closeCommandConnection() error {
	fmt.Println("Client quit!", c.User, c.CmdConn.RemoteAddr())

	//关闭命令通道
	err := c.CmdConn.Close()
	if err != nil {
		fmt.Println("closeCommandConnection error: ", err)
	}

	return err

}

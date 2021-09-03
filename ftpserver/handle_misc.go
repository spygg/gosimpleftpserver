package ftpserver

import (
	"fmt"
	"strings"
)

func (c *Client) handleAUTH(param string) error {
	//先回复
	c.reply(StatusAuthAccepted, "Initializing SSL connection.")

	return nil
}

func (c *Client) handlePROT(param string) error {

	return nil
}

func (c *Client) handlePBSZ(param string) error {
	return c.reply(StatusOK, "Command okay.")
}

func (c *Client) handleFEAT(param string) error {
	//多行的数据要以 代码-消息:\r\n 行1\r\n  End结尾, 以\r\n换行, 新的一行 要 加上一个空格
	features := fmt.Sprintf("%d,-Here is my features:\r\n", StatusSystemStatus)

	cmds := []string{
		"USER",
		"PASS",
		"AUTH",
		"PROT",
		"PBSZ",
		"FEAT",
		"SYST",
		"NOOP",
		"OPTS",
		"QUIT",
		"SIZE",
		"STAT",
		"RETR",
		"STOR",
		"APPE",
		"DELE",
		"RNFR",
		"RNTO",
		"REST",
	}

	for _, c := range cmds {
		features = features + " " + c + "\r\n"
	}

	features += fmt.Sprintf("%d End", StatusSystemStatus)
	return c.replyString(features)
}

func (c *Client) handleSYST(param string) error {
	return c.reply(StatusSystemType, "linux or windows server!")
}

func (c *Client) handleNOOP(param string) error {
	return c.reply(StatusOK, "Command okay.")
}

//由于GO语言默认是 UTF8, 我们也只支持这个吧...
func (c *Client) handleOPTS(param string) error {
	if strings.ToUpper(param) == "UTF8 ON" {
		return c.reply(StatusOK, "Command okay.")
	} else {
		return c.reply(StatusSyntaxErrorParameters, "Bad Command.")
	}
}

func (c *Client) handleQUIT(param string) error {
	c.reply(StatusClosingControlConn, "Client say bye, so we must quit!")

	return c.closeCommandConnection()
}

//事实上我们没管那么多,只是为了回复好看点....
func (c *Client) handleTYPE(param string) error {
	switch param {
	case "I", "L8":
		c.reply(StatusOK, "Type set to binary")
	case "A", "L7":
		c.reply(StatusOK, "Type set to ASCII")
	default:
		c.reply(StatusNotImplementedParam, "Unsupported transfer type")
	}
	return nil
}

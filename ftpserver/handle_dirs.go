package ftpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//返回本地路径
func (c *Client) localPath(fileName string) string {
	return filepath.Join(c.server.RootPath, c.CurrentPath, fileName)
}

func (c *Client) handleCWD(param string) error {
	p := c.localPath(param)

	fi, err := os.Stat(p)
	if (err == nil || os.IsExist(err)) && fi.IsDir() {
		cp := filepath.Clean(filepath.Join(c.CurrentPath, param))

		//替换windows 路径 的\\符号
		c.CurrentPath = strings.ReplaceAll(cp, "\\", "/")

		return c.reply(StatusFileOK, "Requested file action okay, completed.")
	} else {
		return c.reply(StatusActionNotTaken, "Requested action not taken; file unavailable.")
	}
}

func (c *Client) handlePWD(param string) error {
	return c.reply(StatusPathCreated, fmt.Sprintf(`"%s"`, c.CurrentPath))
}

func (c *Client) handleCDUP(param string) error {
	if c.CurrentPath == "/" {
		return c.reply(StatusFileOK, "Requested file action okay, completed.")
	} else {
		return c.handleCWD("..")
	}
}

func (c *Client) handleMKD(param string) error {
	p := c.localPath(param)
	err := os.Mkdir(p, 0755)

	if err != nil {
		return c.reply(StatusActionNotTaken, "Mkdir Failed.")
	} else {
		return c.reply(StatusPathCreated, fmt.Sprintf("\"%s\" created.", param))
	}
}

func (c *Client) handleRMD(param string) error {
	p := c.localPath(param)
	err := os.Remove(p)

	if err != nil {
		return c.reply(StatusActionNotTaken, "Remove directory action not taken; file unavailable.")
	} else {
		return c.reply(StatusPathCreated, "Remove directory action okay, completed.")
	}
}

func (c *Client) handleNLST(param string) error {
	p := c.localPath(param)

	fmt.Println(p)
	return nil
}

func (c *Client) handleLIST(param string) error {
	p := c.localPath(param)

	fmt.Println("handleLIST", p)
	err := c.closeDataConnection()
	return err
}

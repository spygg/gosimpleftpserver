package ftpserver

func (c *Client) handleUSER(param string) error {
	c.User = param

	return c.reply(StatusUserOK, "User name OK, need password.")
}

func (c *Client) handlePASS(param string) error {
	if c.server.AllowAnonymous || (c.server.User == c.User && c.server.PassWord == param) {

		//改变登录状态
		c.LoggedIn = true

		return c.reply(StatusUserLoggedIn, "230 You are logged in.")
	} else {
		return c.reply(StatusNotLoggedIn, "User name or password was incorrect.")
	}
}

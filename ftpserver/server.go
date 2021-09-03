package ftpserver

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()

		if err != nil {
			return err
		}

		c := &Client{
			CmdConn:     conn,
			CurrentPath: "/",
			server:      s,
		}

		s.clients = append(s.clients, c)

		go c.handleClient()
	}
}

func (s *Server) ListenAndServe() error {
	addr := net.JoinHostPort(s.Ip, strconv.Itoa(s.Port))
	ln, err := net.Listen("tcp", addr)

	if err != nil {
		fmt.Println("listen error :", err)
		return err
	}

	return s.Serve(ln)
}

func ListenAndServe(user string, passWord string, ip string, port int, rootPath string) error {
	if port == 0 {
		port = 21
	}

	fi, err := os.Stat(rootPath)
	if !((err == nil || os.IsExist(err)) && fi.IsDir()) {
		currentDir, _ := os.Getwd()
		rootPath = currentDir
	}

	s := &Server{
		User:           user,
		PassWord:       passWord,
		Ip:             ip,
		Port:           port,
		RootPath:       rootPath,
		clients:        make([]*Client, 0),
		AllowAnonymous: true,
	}

	return s.ListenAndServe()
}

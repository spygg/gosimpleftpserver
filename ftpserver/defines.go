package ftpserver

import (
	"net"
	"sync"
)

// from @stevenh's PR proposal
// https://github.com/fclairamb/ftpserverlib/blob/becc125a0770e3b670c4ced7e7bd12594fb024ff/server/consts.go

// Status codes as documented by:
// https://tools.ietf.org/html/rfc959
// https://tools.ietf.org/html/rfc2428
// https://tools.ietf.org/html/rfc2228
const (
	// 100 Series - The requested action is being initiated, expect another reply before
	// proceeding with a new command.
	StatusFileStatusOK = 150 // RFC 959, 4.2.1

	// 200 Series - The requested action has been successfully completed.
	StatusOK                 = 200 // RFC 959, 4.2.1
	StatusNotImplemented     = 202 // RFC 959, 4.2.1
	StatusSystemStatus       = 211 // RFC 959, 4.2.1
	StatusDirectoryStatus    = 212 // RFC 959, 4.2.1
	StatusFileStatus         = 213 // RFC 959, 4.2.1
	StatusHelpMessage        = 214 // RFC 959, 4.2.1
	StatusSystemType         = 215 // RFC 959, 4.2.1
	StatusServiceReady       = 220 // RFC 959, 4.2.1
	StatusClosingControlConn = 221 // RFC 959, 4.2.1
	StatusClosingDataConn    = 226 // RFC 959, 4.2.1
	StatusEnteringPASV       = 227 // RFC 959, 4.2.1
	StatusEnteringEPSV       = 229 // RFC 2428, 3
	StatusUserLoggedIn       = 230 // RFC 959, 4.2.1
	StatusAuthAccepted       = 234 // RFC 2228, 3
	StatusFileOK             = 250 // RFC 959, 4.2.1
	StatusPathCreated        = 257 // RFC 959, 4.2.1

	// 300 Series - The command has been accepted, but the requested action is on hold,
	// pending receipt of further information.
	StatusUserOK            = 331 // RFC 959, 4.2.1
	StatusFileActionPending = 350 // RFC 959, 4.2.1

	// 400 Series - The command was not accepted and the requested action did not take place,
	// but the error condition is temporary and the action may be requested again.
	StatusServiceNotAvailable      = 421 // RFC 959, 4.2.1
	StatusCannotOpenDataConnection = 425 // RFC 959, 4.2.1
	StatusTransferAborted          = 426 // RFC 959, 4.2.1
	StatusFileActionNotTaken       = 450 // RFC 959, 4.2.1

	// 500 Series - Syntax error, command unrecognized and the requested action did not take
	// place. This may include errors such as command line too long.
	StatusSyntaxErrorNotRecognised = 500 // RFC 959, 4.2.1
	StatusSyntaxErrorParameters    = 501 // RFC 959, 4.2.1
	StatusCommandNotImplemented    = 502 // RFC 959, 4.2.1
	StatusBadCommandSequence       = 503 // RFC 959, 4.2.1
	StatusNotImplementedParam      = 504 // RFC 959, 4.2.1
	StatusNotLoggedIn              = 530 // RFC 959, 4.2.1
	StatusActionNotTaken           = 550 // RFC 959, 4.2.1
	StatusActionAborted            = 552 // RFC 959, 4.2.1
	StatusActionNotTakenNoFile     = 553 // RFC 959, 4.2.1
)

type Server struct {
	User     string `json:"user"`
	PassWord string `json:"password"`

	Ip   string `json:"ip"`
	Port int    `json:"port"`

	RootPath       string `json:"rootpath"`       //根路径
	AllowAnonymous bool   `json:"allowanonymous"` //允许匿名登录
	Ssl            bool   `json:"ssl"`            //启用加密证书
	ReadOnly       bool   `json:"readonly"`       //只读
	clients        []*Client
}

type Client struct {
	User string

	CurrentPath    string //当前路径
	LoggedIn       bool
	ActiveTransfer bool     //主动模式(默认被动模式)
	CmdConn        net.Conn //命令通道链接
	DataConn       net.Conn //数据通道链接

	transferWg sync.WaitGroup //数据通道传输同步
	server     *Server        //需要获取服务器的用户名密码,用来登录对比
}

type CommandDescription struct {
	NoNeedAuthentication      bool //不需要登录即可执行
	NeedVerifyWritePermission bool //需要写权限
	TransferRelated           bool //是否要开启第二通道(数据通道)
	Fn                        func(*Client, string) error
}

var commandsMap = map[string]*CommandDescription{
	//Auth
	"USER": {Fn: (*Client).handleUSER, NoNeedAuthentication: true},
	"PASS": {Fn: (*Client).handlePASS, NoNeedAuthentication: true},

	// TLS handling
	"AUTH": {Fn: (*Client).handleAUTH, NoNeedAuthentication: true},
	"PROT": {Fn: (*Client).handlePROT},
	"PBSZ": {Fn: (*Client).handlePBSZ},

	// Misc
	"FEAT": {Fn: (*Client).handleFEAT, NoNeedAuthentication: true},
	"SYST": {Fn: (*Client).handleSYST},
	"NOOP": {Fn: (*Client).handleNOOP},
	"OPTS": {Fn: (*Client).handleOPTS, NoNeedAuthentication: true},
	"QUIT": {Fn: (*Client).handleQUIT},

	// File access
	"SIZE": {Fn: (*Client).handleSIZE},
	"STAT": {Fn: (*Client).handleSTAT},
	"RETR": {Fn: (*Client).handleRETR, TransferRelated: true},
	"STOR": {Fn: (*Client).handleSTOR, TransferRelated: true, NeedVerifyWritePermission: true},
	"APPE": {Fn: (*Client).handleAPPE, TransferRelated: true, NeedVerifyWritePermission: true},
	"DELE": {Fn: (*Client).handleDELE, NeedVerifyWritePermission: true},
	"RNFR": {Fn: (*Client).handleRNFR, NeedVerifyWritePermission: true},
	"RNTO": {Fn: (*Client).handleRNTO, NeedVerifyWritePermission: true},
	"REST": {Fn: (*Client).handleREST},

	// Directory handling
	"CWD":  {Fn: (*Client).handleCWD},
	"PWD":  {Fn: (*Client).handlePWD},
	"XCWD": {Fn: (*Client).handleCWD},
	"XPWD": {Fn: (*Client).handlePWD},
	"CDUP": {Fn: (*Client).handleCDUP},
	"NLST": {Fn: (*Client).handleNLST, TransferRelated: true},
	"LIST": {Fn: (*Client).handleLIST, TransferRelated: true},
	"MKD":  {Fn: (*Client).handleMKD, NeedVerifyWritePermission: true},
	"RMD":  {Fn: (*Client).handleRMD, NeedVerifyWritePermission: true},

	// Connection handling
	"TYPE": {Fn: (*Client).handleTYPE},
	"PASV": {Fn: (*Client).handlePASV},
	"PORT": {Fn: (*Client).handlePORT},
}

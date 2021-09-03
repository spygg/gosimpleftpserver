package main

import (
	"ftpserver/ftpserver"
)

func main() {

	ftpserver.ListenAndServe("test", "test", "", 21, "")
}

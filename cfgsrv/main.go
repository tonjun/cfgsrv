package main

import (
	"fmt"

	"github.com/tonjun/cfgsrv"
)

func main() {
	fmt.Println("vim-go")

	srv := cfgsrv.NewConfigServer(&cfgsrv.Options{})
	srv.Start()
}

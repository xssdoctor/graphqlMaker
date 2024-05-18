package main

import (
	"fmt"

	"github.com/xssdoctor/graphqlMaker/cli"
)


func main() {
	err := cli.Cli()
	if err != nil {
		fmt.Println(err)
	}
}
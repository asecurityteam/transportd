package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	transportd "github.com/asecurityteam/transportd/pkg"
	"github.com/asecurityteam/transportd/pkg/components"
)

func main() {
	ctx := context.Background()
	plugins := components.Defaults

	// Handle the -h flag and print settings.
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.Usage = func() {}
	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		help, errHelp := transportd.Help(ctx, plugins...)
		if errHelp != nil {
			panic(errHelp.Error())
		}
		fmt.Println(help)
		return
	}

	// The system will accept either a full OpenAPI specification through
	// the environment or the name of a file where the specification is
	// stored. Priority is given to the file if both are present.
	fileName := os.Getenv("TRANSPORTD_OPENAPI_SPECIFICATION_FILE")
	fileContent := []byte(os.Getenv("TRANPSPORTD_OPENAPI_SPECIFICATION_CONTENT"))
	var errRead error
	if fileName != "" {
		fileContent, errRead = ioutil.ReadFile(fileName)
		if errRead != nil {
			panic(errRead)
		}
	}

	// Create and run the system.
	rt, err := transportd.New(ctx, fileContent, plugins...)
	if err != nil {
		panic(err.Error())
	}
	if err := rt.Run(); err != nil {
		panic(err.Error())
	}
}

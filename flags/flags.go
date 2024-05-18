package flags

import "github.com/jessevdk/go-flags"

type Flags struct {
	File string `short:"f" long:"file" description:".js file to parse"`
	Folder string `short:"F" long:"folder" description:"Folder with .js files to parse"`
}

func Init() (Flags, error) {
	var o = Flags{}
	parser := flags.NewParser(&o, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		return Flags{}, err
	}
	return o, nil
}
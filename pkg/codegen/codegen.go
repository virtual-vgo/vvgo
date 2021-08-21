package codegen

import (
	"fmt"
	"io"
	"strings"
)

func Package(writer io.Writer, pkg string) error {
	_, err := fmt.Fprintf(writer, "\npackage %s\n", pkg)
	return err
}

func Imports(writer io.Writer, imports []string) error {
	var statements []string
	for _, im := range imports {
		statements = append(statements, fmt.Sprintf(`import "%s"`, im))
	}
	_, err := fmt.Fprint(writer, "\n"+strings.Join(statements, "\n")+"\n")
	return err
}

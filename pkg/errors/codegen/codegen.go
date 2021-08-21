package errors

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/codegen"
	"io"
	"sort"
	"strings"
)

func Generate(writer io.Writer) error {
	_ = codegen.Package(writer, "errors")
	_ = codegen.Imports(writer, []string{"fmt"})
	statements := generateMethodFailures()
	sort.Strings(statements)
	_, err := fmt.Fprint(writer, "\n"+strings.Join(statements, "\n")+"\n")
	return err
}

func generateMethodFailures() []string {
	var statements []string
	for _, x := range codegen.MethodFailures {
		returnStatement := `return fmt.Errorf("` + x.Failed + `() failed: %w", err)`
		statements = append(statements,
			fmt.Sprintf(`func %s(err error) error { %s }`, x.Method, returnStatement))
	}
	return statements
}

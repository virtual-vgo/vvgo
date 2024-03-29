package logger

import (
	"fmt"
	"github.com/virtual-vgo/vvgo/pkg/codegen"
	"io"
	"sort"
	"strings"
)

func Generate(writer io.Writer) error {
	_ = codegen.Package(writer, "logger")
	_ = codegen.Imports(writer, []string{"context", "github.com/sirupsen/logrus"})

	statements := generateMethods()
	statements = append(statements, generateMethodFailures()...)
	sort.Strings(statements)
	_, _ = fmt.Fprint(writer, strings.Join(statements, "\n"))
	return nil
}

func generateMethods() []string {
	var statements []string
	for _, x := range []struct {
		method      string
		params      string
		args        string
		returnEntry bool
	}{
		{method: `WithField`, params: `key string, value interface{}`, args: `key, value`, returnEntry: true},
		{method: `WithFields`, params: `fields logrus.Fields`, args: `fields`, returnEntry: true},
		{method: `WithError`, params: `err error`, args: `err`, returnEntry: true},
		{method: `WithContext`, params: `ctx context.Context`, args: `ctx`, returnEntry: true},
		{method: `Print`, params: `args ...interface{}`, args: `args...`},
		{method: `Info`, params: `args ...interface{}`, args: `args...`},
		{method: `Fatal`, params: `args ...interface{}`, args: `args...`},
		{method: `Warn`, params: `args ...interface{}`, args: `args...`},
		{method: `Error`, params: `args ...interface{}`, args: `args...`},
		{method: `Println`, params: `args ...interface{}`, args: `args...`},
		{method: `Infoln`, params: `args ...interface{}`, args: `args...`},
		{method: `Warnln`, params: `args ...interface{}`, args: `args...`},
		{method: `Errorln`, params: `args ...interface{}`, args: `args...`},
		{method: `Fatalln`, params: `args ...interface{}`, args: `args...`},
		{method: `Printf`, params: `format string, args ...interface{}`, args: `format, args...`},
		{method: `Infof`, params: `format string, args ...interface{}`, args: `format, args...`},
		{method: `Warnf`, params: `format string, args ...interface{}`, args: `format, args...`},
		{method: `Errorf`, params: `format string, args ...interface{}`, args: `format, args...`},
		{method: `Fatalf`, params: `format string, args ...interface{}`, args: `format, args...`},
	} {
		if x.returnEntry {
			statements = append(statements,
				fmt.Sprintf("func %s(%s) Entry { return Entry{logrus.%s(%s)} }", x.method, x.params, x.method, x.args),
				fmt.Sprintf("func (e Entry) %s(%s) Entry { return Entry{e.Entry.%s(%s)} }", x.method, x.params, x.method, x.args))
		} else {
			statements = append(statements,
				fmt.Sprintf("func %s(%s) { logrus.%s(%s) }", x.method, x.params, x.method, x.args),
				fmt.Sprintf("func (e Entry) %s(%s) { e.Entry.%s(%s) }", x.method, x.params, x.method, x.args))
		}
	}
	return statements
}

func generateMethodFailures() []string {
	var statements []string
	for _, x := range codegen.MethodFailures {
		params := `ctx context.Context, err error`
		args := fmt.Sprintf(`ctx, "%s", err`, x.Failed)
		statements = append(statements,
			fmt.Sprintf("func %s(%s) { MethodFailure(%s) }", x.Method, params, args),
			fmt.Sprintf("func (e Entry) %s(%s)  { e.MethodFailure(%s) }", x.Method, params, args))
	}
	return statements
}

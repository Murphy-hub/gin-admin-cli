package parser

import (
	"bytes"
	"context"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"

	"github.com/gin-admin/gin-admin-cli/v10/internal/utils"
	"go.uber.org/zap"
)

func ModifyModuleMainFile(ctx context.Context, args BasicArgs) ([]byte, error) {
	filename, err := GetModuleMainFilePath(args.ModuleName)
	if err != nil {
		return nil, err
	}

	fullname := filepath.Join(args.Dir, filename)
	zap.S().Infof("[%s] modify struct %s from file %s", args.Flag.String(), args.StructName, fullname)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fullname, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Walk(&astModuleMainVisitor{
		fset:       fset,
		moduleName: args.ModuleName,
		structName: args.StructName,
		flag:       args.Flag,
	}, f)

	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, f); err != nil {
		return nil, err
	}

	buf = utils.Scanner(buf, func(line string) string {
		if strings.HasPrefix(strings.TrimSpace(line), "new(schema") {
			return strings.ReplaceAll(line, "), new(", "),\n \t\tnew(")
		}
		return line
	})

	return buf.Bytes(), nil
}

func ModifyModuleWireFile(ctx context.Context, args BasicArgs, genPackages []string) ([]byte, error) {
	filename, err := GetModuleWireFilePath(args.ModuleName)
	if err != nil {
		return nil, err
	}

	fullname := filepath.Join(args.Dir, filename)
	zap.S().Infof("[%s] modify struct %s from file %s", args.Flag.String(), args.StructName, fullname)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fullname, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Walk(&astModuleWireVisitor{
		fset:        fset,
		moduleName:  args.ModuleName,
		structName:  args.StructName,
		flag:        args.Flag,
		genPackages: genPackages,
	}, f)

	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, f); err != nil {
		return nil, err
	}

	buf = utils.Scanner(buf, func(line string) string {
		if strings.HasPrefix(strings.TrimSpace(line), "wire.Struct") {
			return strings.ReplaceAll(line, "), wire.", "),\n \twire.")
		}
		return line
	})

	return buf.Bytes(), nil
}

func ModifyModsFile(ctx context.Context, args BasicArgs) ([]byte, error) {
	filename := GetModsFilePath()
	fullname := filepath.Join(args.Dir, filename)
	zap.S().Infof("[%s] modify module from file %s", args.Flag.String(), fullname)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fullname, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	ast.Walk(&astModsVisitor{
		fset:       fset,
		dir:        args.Dir,
		moduleName: args.ModuleName,
		flag:       args.Flag,
	}, f)

	buf := new(bytes.Buffer)
	if err := format.Node(buf, fset, f); err != nil {
		return nil, err
	}

	buf = utils.Scanner(buf, func(line string) string {
		if strings.Contains(strings.TrimSpace(line), ".Set, ") {
			return strings.ReplaceAll(line, ".Set, ", ".Set,\n \t")
		}
		return line
	})

	result := bytes.ReplaceAll(buf.Bytes(), []byte("ctx,\n\n\t\t\tv1"), []byte("ctx, v1"))
	result = bytes.ReplaceAll(result, []byte(".\n\t\tInit"), []byte(".Init"))
	result = bytes.ReplaceAll(result, []byte(".\n\t\tRegister"), []byte(".Register"))

	return result, nil
}

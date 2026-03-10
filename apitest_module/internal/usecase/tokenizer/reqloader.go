package tokenizer

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/rdhmuhammad/phisiobook/pkg/logger"
)

func LoadRequest(tokenize *Tokenizer) {
	root := "../" // module root
	for _, g := range tokenize.groups {
		for _, r := range g.routers {
			if r.requestBodyDir == "" {
				continue
			}
			schema, err := findStructInAllowedDirs(root, r.requestBodyDir, []string{
				"internal/core/usecase",
				"shared/payload",
				"iam_module/internal/core/usecase",
			})
			if err != nil {
				panic(err)
			}

			if schema == nil {
				continue
			}

			out, err := json.MarshalIndent(schema, "", "  ")
			if err != nil {
				panic(err)
			}

			r.requestBodyValue = string(out)
			logger.Infof(r.requestBodyValue)
			tokenize.editRouter(r)
		}
	}
}

func splitQualifiedType(full string) (string, string, error) {
	parts := strings.Split(full, ".")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid qualified type: %s", full)
	}
	return parts[0], parts[1], nil
}

func findStructInAllowedDirs(root, fullType string, allowedDirs []string) (*StructSchema, error) {
	pkgName, structName, err := splitQualifiedType(fullType)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()

	for _, relDir := range allowedDirs {
		baseDir := filepath.Join(root, relDir)

		schema, err := scanDirRecursive(fset, baseDir, pkgName, structName)
		if err != nil {
			return nil, err
		}
		if schema != nil {
			return schema, nil
		}
	}

	return nil, nil
}
func scanDirRecursive(fset *token.FileSet, dir, pkgName, structName string) (*StructSchema, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			schema, err := scanDirRecursive(fset, fullPath, pkgName, structName)
			if err != nil {
				return nil, err
			}
			if schema != nil {
				return schema, nil
			}
			continue
		}

		if !isGoFile(entry.Name()) {
			continue
		}

		schema, err := scanFileForStruct(fset, fullPath, pkgName, structName)
		if err != nil {
			return nil, err
		}
		if schema != nil {
			return schema, nil
		}
	}

	return nil, nil
}

func scanFileForStruct(fset *token.FileSet, filePath, pkgName, structName string) (*StructSchema, error) {
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil
	}

	if file.Name == nil || file.Name.Name != pkgName {
		return nil, nil
	}

	return findStructInFile(fset, file, filePath, structName), nil
}

func findStructInFile(fset *token.FileSet, file *ast.File, path, structName string) *StructSchema {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok || typeSpec.Name == nil || typeSpec.Name.Name != structName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			schema := &StructSchema{
				StructName: structName,
				Package:    file.Name.Name,
				File:       path,
				Fields:     extractFields(fset, structType),
			}

			return schema
		}
	}

	return nil
}
func extractFields(fset *token.FileSet, st *ast.StructType) []StructField {
	var fields []StructField

	if st.Fields == nil {
		return fields
	}

	for _, field := range st.Fields.List {
		typeName := exprToString(fset, field.Type)
		jsonName := parseJSONTag(field.Tag)

		// embedded field
		if len(field.Names) == 0 {
			fields = append(fields, StructField{
				Name:     typeName,
				JSONName: jsonName,
				Type:     typeName,
			})
			continue
		}

		for _, name := range field.Names {
			jn := jsonName
			if jn == "" {
				jn = lowerFirst(name.Name)
			}

			fields = append(fields, StructField{
				Name:     name.Name,
				JSONName: jn,
				Type:     typeName,
			})
		}
	}

	return fields
}

func isGoFile(name string) bool {
	return strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
}

func exprToString(fset *token.FileSet, expr ast.Expr) string {
	var b strings.Builder
	_ = printer.Fprint(&b, fset, expr)
	return b.String()
}

func parseJSONTag(tagLit *ast.BasicLit) string {
	if tagLit == nil {
		return ""
	}

	raw, err := strconvUnquote(tagLit.Value)
	if err != nil {
		return ""
	}

	tag := reflect.StructTag(raw)
	jsonTag := tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return ""
	}

	name := strings.Split(jsonTag, ",")[0]
	if name == "" {
		return ""
	}

	return name
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func strconvUnquote(s string) (string, error) {
	// small wrapper so helper section stays compact
	return strconv.Unquote(s)
}

type foundStructError struct {
	Schema *StructSchema
}

type StructSchema struct {
	StructName string        `json:"struct_name"`
	Package    string        `json:"package"`
	File       string        `json:"file"`
	Fields     []StructField `json:"fields"`
}

type StructField struct {
	Name     string `json:"name"`
	JSONName string `json:"json_name"`
	Type     string `json:"type"`
}

func (e *foundStructError) Error() string {
	return "struct found"
}

func foundStruct(schema *StructSchema) error {
	return &foundStructError{Schema: schema}
}

package tokenizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/rdhmuhammad/phisiobook/pkg/logger"
)

func Generate(tokenize *Tokenizer) {
	gofile := os.Getenv("GOFILE")
	root, err := findProjectRoot(filepath.Dir(gofile))
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: project root not found: %v\n", err)
		return
	}

	collectionName, baseURL := readEnvApitest(root)
	outDir := filepath.Join(root, "resource", "apidocs")
	outPath := filepath.Join(outDir, "collection.json")

	docs := buildAPIDocs(tokenize, collectionName, baseURL, outPath)

	out, err := json.MarshalIndent(docs, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate: marshal error: %v\n", err)
		return
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "generate: mkdir error: %v\n", err)
		return
	}
	if err := os.WriteFile(outPath, out, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "generate: write error: %v\n", err)
		return
	}

	fmt.Printf("generate: written %s\n", outPath)
}

func findProjectRoot(start string) (string, error) {
	dir := start
	if dir == "" {
		dir = "."
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(abs, "go.mod")); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			return "", fmt.Errorf("go.mod not found from %s", start)
		}
		abs = parent
	}
}

func readEnvApitest(root string) (name, baseURL string) {
	name = "Phisiobook API"
	baseURL = "http://localhost:8999/api/v1"

	data, err := os.ReadFile(filepath.Join(root, "env.apitest"))
	if err != nil {
		return
	}

	if string(data) != "" {
		logger.Infof("Using user defined environment variable")
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "name":
			name = strings.TrimSpace(val)
		case "base_url":
			baseURL = strings.TrimSpace(val)
		}
	}
	return
}

func buildAPIDocs(tokenize *Tokenizer, collectionName, baseURL, outPath string) APIDocs {
	docs := loadExistingDocs(outPath, collectionName, baseURL)

	for _, g := range tokenize.groups {
		folder := CollectionItem{
			FunIden: g.VarName,
			Name:    folderName(g),
		}
		for _, item := range docs.Item {
			if item.FunIden == g.VarName {
				folder = item
				break
			}
		}

		var newItem []CollectionItem
		var storedItem = make(map[string]bool)
		for _, item := range folder.Item {
			r, ok := tokenize.getRouter(folder.FunIden, item.FunIden)
			if ok {
				colItem := buildItem(g, r)
				storedItem[r.handlerFunc] = true
				newItem = append(newItem, colItem)
			} else {
				newItem = append(newItem, item)
			}
		}

		for _, r := range g.routers {
			if storedItem[r.handlerFunc] {
				continue
			}
			colItem := buildItem(g, r)
			newItem = append(newItem, colItem)
		}

		folder.Item = newItem
		if len(folder.Item) > 0 {
			docs.Item = append(docs.Item, folder)
		}
	}

	return docs
}

func loadExistingDocs(outPath, collectionName, baseURL string) APIDocs {
	defaults := APIDocs{
		Info: CollectionInfo{
			Name:        collectionName,
			Description: "Auto-generated API collection",
			Schema:      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Variable: []CollectionVar{
			{Key: "base_url", Value: baseURL, Type: "string"},
			{Key: "token", Value: "", Type: "string"},
		},
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		logger.Infof("Creating new api docs file")
		return defaults
	}

	var existing APIDocs
	if err := json.Unmarshal(data, &existing); err != nil {
		return defaults
	}

	logger.Infof("Using existing docs file")
	return existing
}

func buildItem(g group, r router) CollectionItem {
	fullPath := g.path + r.path
	rawURL := "{{base_url}}" + fullPath
	pathParts := strings.FieldsFunc(fullPath, func(c rune) bool { return c == '/' })

	req := Request{
		Method: r.method,
		Header: buildHeaders(r),
		Body: &RequestBody{
			Mode: "raw",
			Raw:  r.requestBodyValue,
		},
		URL: RequestURL{
			Raw:  rawURL,
			Host: []string{"{{base_url}}"},
			Path: pathParts,
		},
	}

	return CollectionItem{
		FunIden:  g.VarName + "." + r.handlerFunc,
		Name:     itemName(r),
		Request:  &req,
		Response: []any{},
	}
}

func buildHeaders(r router) []Header {
	var h []Header
	if r.requestBodyValue != "" {
		h = append(h, Header{Key: "Content-Type", Value: "application/json"})
	}
	h = append(h, Header{Key: "Authorization", Value: "Bearer {{token}}"})
	return h
}

func folderName(g group) string {
	path := strings.Trim(g.path, "/")
	if path == "" {
		return g.VarName
	}
	parts := strings.Split(path, "/")
	var words []string
	for _, p := range parts {
		if p != "" {
			words = append(words, capitalizeFirst(p))
		}
	}
	return strings.Join(words, " ")
}

func itemName(r router) string {
	if r.handlerFunc == "" {
		return r.method + " " + r.path
	}
	return splitCamel(r.handlerFunc)
}

func splitCamel(s string) string {
	var b strings.Builder
	for i, c := range s {
		if i > 0 && unicode.IsUpper(c) {
			b.WriteRune(' ')
		}
		b.WriteRune(c)
	}
	return b.String()
}

func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

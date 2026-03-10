package main

import (
	"github.com/rdhmuhammad/phisiobook/apitest_module/internal/usecase/tokenizer"
	"github.com/rdhmuhammad/phisiobook/pkg/logger"
)

func main() {
	logger.DefaultLogger()
	tokenize := tokenizer.Load()
	tokenizer.LoadRequest(tokenize)
	tokenizer.Generate(tokenize)
}

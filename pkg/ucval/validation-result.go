package ucval

type ValidationResult struct {
	IsValid bool
	Message string
}

func Valid() ValidationResult {
	return ValidationResult{IsValid: true}
}

func Invalid(msg string) ValidationResult {
	return ValidationResult{IsValid: false, Message: msg}
}

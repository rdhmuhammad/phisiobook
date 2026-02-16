//go:generate mockery --all --inpackage --case snake

package ucval

type ValidationNode[T any] struct {
	next *ValidationNode[T]
	ValidationValidator[T]
}

type ValidationValidator[T any] interface {
	Validate(input T) ValidationResult
}

func NewValidationNode[T any](node ValidationValidator[T]) *ValidationNode[T] {
	return &ValidationNode[T]{
		ValidationValidator: node,
	}
}

func (th *ValidationNode[T]) GetResult(request T) ValidationResult {
	validateResult := th.Validate(request)
	if !validateResult.IsValid {
		return validateResult
	}
	if th.next == nil {
		return validateResult
	}

	lastNext := th.next
	for lastNext != nil {
		vd := lastNext.Validate(request)
		if !vd.IsValid {
			return vd
		}
		lastNext = lastNext.next
	}

	return validateResult
}

func (th *ValidationNode[T]) Add(node *ValidationNode[T]) *ValidationNode[T] {
	if th.next == nil {
		th.next = node
		return th
	}

	lastNext := th.next
	for lastNext.next != nil {
		lastNext = lastNext.next
	}
	lastNext.next = node
	return node
}

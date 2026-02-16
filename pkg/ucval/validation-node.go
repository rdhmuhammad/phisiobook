//go:generate mockery --all --inpackage --case snake

package ucval

type ValidationNode[T any] struct {
	next     *ValidationNode[T]
	Validate func(input T) ValidationResult
}

func NewNode[T any](validate func(input T) ValidationResult) *ValidationNode[T] {
	return &ValidationNode[T]{
		Validate: validate,
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

func (th *ValidationNode[T]) Add(fn func(input T) ValidationResult) *ValidationNode[T] {
	node := NewNode[T](fn)

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

package cloud

import "fmt"

type notExistsError struct {
	resource string
	search   string
	err      string
}

func (err *notExistsError) Error() string {
	if err.err != "" {
		return fmt.Sprintf("couldn't find %v '%v': %v", err.resource, err.search, err.err)
	} else {
		return fmt.Sprintf("couldn't find %v '%v'", err.resource, err.search)
	}
}

func (err *notExistsError) RuntimeError() {
	panic("implement me")
}

func newNotExistsError(resource string, search string, err error) *notExistsError {
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	return &notExistsError{
		resource: resource,
		search:   search,
		err:      errString,
	}
}

func IsNotExistsError(err error) bool {
	_, ok := err.(*notExistsError)
	return ok
}

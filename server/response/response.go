package response

type Error struct {
	Code        int               `json:"code"`
	Error       string            `json:"error,omitempty"`
	Errors      []string          `json:"errors,omitempty"`
	InputErrors map[string]string `json:"input_errors,omitempty"`
}

type Success struct {
	Message string `json:"message"`
}

func OK(message string) Success {
	return Success{
		Message: message,
	}
}

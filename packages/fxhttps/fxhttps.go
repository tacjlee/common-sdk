package fxhttps

import "github.com/gin-gonic/gin"

func ParsePathParameters(ctx *gin.Context) map[string]string {
	params := make(map[string]string)
	// Extract path parameters (defined in route like /user/:id)
	for _, param := range ctx.Params {
		params[param.Key] = param.Value
	}
	return params
}

func ParseQueryParameters(ctx *gin.Context) map[string]string {
	params := make(map[string]string)
	for key, values := range ctx.Request.URL.Query() {
		// Using the first value if there are multiple
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

func ParseFormParameters(ctx *gin.Context) (map[string]string, error) {
	params := make(map[string]string)
	// Ensure the form is parsed
	err := ctx.Request.ParseForm()
	if err != nil {
		return nil, err
	}
	for key, values := range ctx.Request.PostForm {
		// Using the first value if there are multiple
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params, nil
}

func ParseJsonBody[T any](ctx *gin.Context) (T, error) {
	var result T
	if err := ctx.ShouldBindJSON(&result); err != nil {
		return result, err
	}
	return result, nil
}

package globals

var AuthKeys = map[string]string{
	"gcp":     "Authorization",
	"aws":     "x-api-key",
	"azure":   "x-functions-key",
	"alibaba": "Authorization",
}

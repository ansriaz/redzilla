package api

import (
	"net/http"
	"strings"

	"github.com/ansriaz/redzilla/model"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// JSONError a JSON response in case of error
type JSONError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// errorResponse send an error response with a common JSON message
func errorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, JSONError{code, message})
}

func internalError(c *gin.Context, err error) {
	logrus.Errorf("Internal Error: %s", err.Error())
	logrus.Debugf("%+v", err)
	code := http.StatusInternalServerError
	errorResponse(c, code, http.StatusText(code))
}

func notFound(c *gin.Context) {
	code := http.StatusNotFound
	errorResponse(c, code, http.StatusText(code))
}

func badRequest(c *gin.Context) {
	code := http.StatusNotFound
	errorResponse(c, code, http.StatusText(code))
}

func extractSubdomain(host string, cfg *model.Config) string {
	if len(host) == 0 {
		return ""
	}
	hostname := ""
	if strings.Contains(host, ":") {
		hostname = host[:strings.Index(host, ":")]
	} else {
		hostname = host
	}
	name := strings.Replace(hostname, "."+cfg.Domain, "", -1)
	return name
}

func extractInstanceName(path string, cfg *model.Config) string {
	if len(path) == 0 {
		return ""
	}
	strFind := "/instance/"
	hostname := path[strings.Index(path, strFind)+len(strFind):]
	return hostname
}

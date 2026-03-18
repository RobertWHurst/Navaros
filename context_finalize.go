package navaros

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// finalize is called by the router after all matching handlers have processed
// the request. It is responsible for writing the response to the client. For
// libraries which to extend or encapsulate the functionality of Navaros,
// Finalize can be called with the CtxFinalize function.
func (c *Context) finalize() {
	if c.Error != nil {
		c.Status = 500
		if PrintHandlerErrors {
			fmt.Printf("Error occurred when handling request: %s\n%s", c.Error, c.ErrorStack)
		}
	}

	var finalBodyReader io.Reader
	var redirect *Redirect

	if !c.hasWrittenBody && c.Body != nil {
		if bodyReader, ok := c.Body.(io.Reader); ok {
			finalBodyReader = bodyReader
		} else {
			switch body := c.Body.(type) {
			case *Redirect:
				redirect = body
			case Redirect:
				redirect = &body
			case string:
				finalBodyReader = strings.NewReader(body)
			case []byte:
				finalBodyReader = bytes.NewReader(body)
			default:
				marshalledReader, err := c.marshallResponseBody()
				if err == nil {
					finalBodyReader = marshalledReader
				} else {
					c.Status = 500
					if PrintHandlerErrors {
						fmt.Printf("Error occurred when marshalling response body: %s", err)
					}
				}
			}
		}
	}

	if c.Status == 0 {
		if redirect != nil {
			c.Status = 302
		} else if finalBodyReader == nil {
			c.Status = 404
		} else {
			c.Status = 200
		}
	}

	if redirect != nil {
		to := resolveRedirectLocation(redirect.To, c.Request().URL.Path)
		c.Headers.Set("Location", to)
	}

	writer := c.bodyWriter
	if writer == nil {
		writer = c.responseWriter
		if !c.inhibitResponse {
			for key, values := range c.Headers {
				for _, value := range values {
					writer.Header().Add(key, value)
				}
			}
			for _, cookie := range c.Cookies {
				http.SetCookie(writer, cookie)
			}
		}
	}
	if !c.inhibitResponse {
		writer.WriteHeader(c.Status)
	}

	hasBody := finalBodyReader != nil
	is100Range := c.Status >= 100 && c.Status < 200
	is204Or304 := c.Status == 204 || c.Status == 304

	if !c.inhibitResponse && hasBody {
		if is100Range || is204Or304 {
			fmt.Printf("response with status %d has body but no content is expected", c.Status)
		} else {
			_, err := io.Copy(writer, finalBodyReader)
			if finalBodyReaderCloser, ok := finalBodyReader.(io.Closer); ok {
				if err := finalBodyReaderCloser.Close(); err != nil && PrintHandlerErrors {
					fmt.Printf("Failed to close body read closer: %s", err)
				}
			}
			if err != nil {
				c.Status = 500
				fmt.Printf("error occurred when writing response body: %s", err)
			}
		}
	}

	if closer, ok := writer.(io.Closer); ok {
		if err := closer.Close(); err != nil && PrintHandlerErrors {
			fmt.Printf("Failed to close body writer: %s", err)
		}
	}

	c.FinalError = c.Error
	c.FinalErrorStack = c.ErrorStack
	if c.doneChannel != nil {
		close(c.doneChannel)
	}
}

func resolveRedirectLocation(to string, currentPath string) string {
	toUrl, err := url.Parse(to)
	if err != nil {
		return to
	}
	if toUrl.Scheme != "" || toUrl.Host != "" {
		return to
	}
	if currentPath == "" {
		currentPath = "/"
	}
	if to == "" || to[0] != '/' {
		dir, _ := path.Split(currentPath)
		to = dir + to
	}
	return to
}

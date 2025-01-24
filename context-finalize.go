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
		if bodyReader, ok := c.Body.(io.ReadCloser); ok {
			finalBodyReader = bodyReader
			defer bodyReader.Close()
		} else if bodyReader, ok := c.Body.(io.Reader); ok {
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

	if !c.hasWrittenHeaders {
		for key, values := range c.Headers {
			for _, value := range values {
				c.bodyWriter.Header().Add(key, value)
			}
		}

		if redirect != nil {
			to := redirect.To
			toUrl, err := url.Parse(to)
			if err != nil {
				c.Status = 500
				fmt.Printf("Error occurred when parsing redirect url: %s", err)
			} else {
				if toUrl.Scheme == "" && toUrl.Host == "" {
					currentPath := c.Request().URL.Path
					if currentPath == "" {
						currentPath = "/"
					}
					if to == "" || to[0] != '/' {
						currentChunks, _ := path.Split(currentPath)
						to = currentChunks + to
					}
					query := ""
					if i := strings.Index(to, "?"); i != -1 {
						query = to[i:]
						to = to[:i]
					}
					to += query
				}
			}
			c.bodyWriter.Header().Add("Location", to)
		}

		for _, cookie := range c.Cookies {
			http.SetCookie(c.bodyWriter, cookie)
		}

		if !c.inhibitResponse {
			c.bodyWriter.WriteHeader(c.Status)
		}
	}

	hasBody := finalBodyReader != nil
	is100Range := c.Status >= 100 && c.Status < 200
	is204Or304 := c.Status == 204 || c.Status == 304

	if !c.inhibitResponse && hasBody {
		if is100Range || is204Or304 {
			fmt.Printf("Response with status %d has body but no content is expected", c.Status)
		} else {
			_, err := io.Copy(c.bodyWriter, finalBodyReader)
			if err != nil {
				c.Status = 500
				fmt.Printf("Error occurred when writing response body: %s", err)
			}
		}
	}

	c.FinalError = c.Error
	c.FinalErrorStack = c.ErrorStack
	for _, doneHandler := range c.doneHandlers {
		doneHandler()
	}
}

package navaros

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func (c *Context) finalize() {
	if c.Error != nil {
		c.Status = 500
		if PrintHandlerErrors {
			fmt.Printf("Error occurred when handling request: %s\n%s", c.Error, c.ErrorStack)
		}
	}

	var finalBodyReader io.Reader
	if !c.hasWrittenBody && c.Body != nil {
		switch body := c.Body.(type) {
		case string:
			finalBodyReader = bytes.NewBufferString(body)
		case []byte:
			finalBodyReader = bytes.NewBuffer(body)
		default:
			marshalledReader, err := c.marshallResponseBody()
			if err != nil {
				c.Status = 500
				fmt.Printf("Error occurred when marshalling response body: %s", err)
			} else {
				finalBodyReader = marshalledReader
			}
		}
	}

	if c.Status == 0 {
		if finalBodyReader == nil {
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
		for _, cookie := range c.Cookies {
			http.SetCookie(c.bodyWriter, cookie)
		}
		c.bodyWriter.WriteHeader(c.Status)
	}

	hasBody := finalBodyReader != nil
	is100Range := c.Status >= 100 && c.Status < 200
	is204Or304 := c.Status == 204 || c.Status == 304

	if hasBody {
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

	delete(contextData, c)

	c.FinalError = c.Error
	c.FinalErrorStack = c.ErrorStack
	for _, doneHandler := range c.doneHandlers {
		doneHandler()
	}
}

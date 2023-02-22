package request

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const (
	suborbitalStateHeader     = "X-Suborbital-State"
	suborbitalParamsHeader    = "X-Suborbital-Params"
	suborbitalRequestIDHeader = "X-Suborbital-RequestID"
)

// CoordinatedRequest represents a request whose fulfillment can be coordinated across multiple hosts
// and is serializable to facilitate interoperation with Wasm Modules and transmissible over the wire
type CoordinatedRequest struct {
	Method       string            `json:"method"`
	URL          string            `json:"url"`
	ID           string            `json:"request_id"`
	Body         []byte            `json:"body"`
	Headers      map[string]string `json:"headers"`
	RespHeaders  map[string]string `json:"resp_headers"`
	Params       map[string]string `json:"params"`
	State        map[string][]byte `json:"state"`
	SequenceJSON []byte            `json:"sequence_json,omitempty"`

	bodyValues map[string]interface{}
}

// FromEchoContext creates a CoordinatedRequest from an echo context.
func FromEchoContext(c echo.Context) (*CoordinatedRequest, error) {
	var err error
	reqBody := make([]byte, 0)

	if c.Request().Body != nil { // Read
		reqBody, err = io.ReadAll(c.Request().Body)
		if err != nil {
			return nil, errors.Wrap(err, "io.ReadAll request body")
		}
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset

	flatHeaders := map[string]string{}
	for k, v := range c.Request().Header {
		// we lowercase the key to have case-insensitive lookup later
		flatHeaders[strings.ToLower(k)] = v[0]
	}

	flatParams := map[string]string{}
	for _, p := range c.ParamNames() {
		flatParams[p] = c.Param(p)
	}

	return &CoordinatedRequest{
		Method:      c.Request().Method,
		URL:         c.Request().URL.RequestURI(),
		ID:          c.Request().Header.Get("X-Request-ID"),
		Body:        reqBody,
		Headers:     flatHeaders,
		RespHeaders: map[string]string{},
		Params:      flatParams,
		State:       map[string][]byte{},
	}, nil
}

// UseSuborbitalHeaders adds the values in the state and params headers JSON to the CoordinatedRequest's State and Params
func (c *CoordinatedRequest) UseSuborbitalHeaders(ec echo.Context) error {
	// fill in initial state from the state header
	stateJSON := ec.Request().Header.Get(suborbitalStateHeader)
	if err := c.addState(stateJSON); err != nil {
		return err
	}

	// fill in the URL params from the Params header
	paramsJSON := ec.Request().Header.Get(suborbitalParamsHeader)
	if err := c.addParams(paramsJSON); err != nil {
		return err
	}

	ec.Response().Header()[suborbitalRequestIDHeader] = []string{ec.Request().Header.Get("requestID")}

	return nil
}

// BodyField returns a field from the request body as a string
func (c *CoordinatedRequest) BodyField(key string) (string, error) {
	if c.bodyValues == nil {
		if len(c.Body) == 0 {
			return "", nil
		}

		vals := map[string]interface{}{}

		if err := json.Unmarshal(c.Body, &vals); err != nil {
			return "", errors.Wrap(err, "failed to Unmarshal request body")
		}

		// cache the parsed body
		c.bodyValues = vals
	}

	interfaceVal, ok := c.bodyValues[key]
	if !ok {
		return "", fmt.Errorf("body does not contain field %s", key)
	}

	stringVal, ok := interfaceVal.(string)
	if !ok {
		return "", fmt.Errorf("request body value %s is not a string", key)
	}

	return stringVal, nil
}

// SetBodyField sets a field in the JSON body to a new value
func (c *CoordinatedRequest) SetBodyField(key, val string) error {
	if c.bodyValues == nil {
		if len(c.Body) == 0 {
			return nil
		}

		vals := map[string]interface{}{}

		if err := json.Unmarshal(c.Body, &vals); err != nil {
			return errors.Wrap(err, "failed to Unmarshal request body")
		}

		// cache the parsed body
		c.bodyValues = vals
	}

	c.bodyValues[key] = val

	newJSON, err := json.Marshal(c.bodyValues)
	if err != nil {
		return errors.Wrap(err, "failed to Marshal new body")
	}

	c.Body = newJSON

	return nil
}

// FromJSON unmarshalls a CoordinatedRequest from JSON
func FromJSON(jsonBytes []byte) (*CoordinatedRequest, error) {
	req := CoordinatedRequest{}
	if err := json.Unmarshal(jsonBytes, &req); err != nil {
		return nil, errors.Wrap(err, "failed to Unmarshal request")
	}

	if req.Method == "" || req.URL == "" || req.ID == "" {
		return nil, errors.New("JSON is not CoordinatedRequest, required fields are empty")
	}

	return &req, nil
}

// ToJSON returns a JSON representation of a CoordinatedRequest
func (c *CoordinatedRequest) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CoordinatedRequest) addState(stateJSON string) error {
	if stateJSON == "" {
		return nil
	}

	if c.State == nil {
		c.State = map[string][]byte{}
	}

	state := map[string]string{}

	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return errors.Wrap(err, "failed to Unmarshal state header")
	} else {
		// iterate over the state and convert each field to bytes
		for k, v := range state {
			c.State[k] = []byte(v)
		}
	}

	return nil
}

func (c *CoordinatedRequest) addParams(paramsJSON string) error {
	if paramsJSON == "" {
		return nil
	}

	if c.Params == nil {
		c.Params = map[string]string{}
	}

	state := map[string]string{}

	if err := json.Unmarshal([]byte(paramsJSON), &state); err != nil {
		return errors.Wrap(err, "failed to Unmarshal params header")
	} else {
		// iterate over the state and convert each field to bytes
		for k, v := range state {
			c.Params[k] = v
		}
	}

	return nil
}

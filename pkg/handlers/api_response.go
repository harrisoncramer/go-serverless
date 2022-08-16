package handlers

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// Writes a body and status code to a APIGatewayProxyResponse struct and returns it.
func apiResponse(status int, body interface{}) (*events.APIGatewayProxyResponse, error) {

	resp := events.APIGatewayProxyResponse{
		StatusCode:        status,
		Headers:           map[string]string{"Content-Type": "application/json"},
		MultiValueHeaders: map[string][]string{},
		Body:              "",
		IsBase64Encoded:   false,
	}
	resp.StatusCode = status

	stringBody, _ := json.Marshal(body)
	resp.Body = string(stringBody)

	return &resp, nil
}

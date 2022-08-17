// The controller functions which access the database and the structs which define the data. If this were a larger package these would be split up.
package user

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/harrisoncramer/go-serverless/pkg/validators"
)

var (
	ErrorFailedToFetchRecord     = "Failed to fetch record"
	ErrorInvalidUserData         = "Invalid user data"
	ErrorFailedToUnmarshalRecord = "Failed to unmarshal record"
	ErrorInvalidEmail            = "Invalid Email"
	ErrorCouldNotMarshalItem     = "Could not marshal item"
	ErrorCouldNotDeleteItem      = "Could not delete item"
	ErrorCouldNotPutDynamoItem   = "Could not dynamo put item"
	ErrorUserAlreadyExists       = "User already exists"
	ErrorUserDoesNotExist        = "User does not exist"
)

type User struct {
	Email     string `json: "email"`
	FirstName string `json: "firstName"`
	LastName  string `json: "lastName"`
}

func FetchUser(email string, tablename string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {

	// Construct a query based on the email, similar to MongoDB. We will replace with PG later.
	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tablename),
	}

	// Run the query and capture the result
	res, err := dynaClient.GetItem(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	// Unmarshal the result back into the User struct
	item := new(User)
	err = dynamodbattribute.UnmarshalMap(res.Item, item)

	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	return item, nil

}

func FetchUsers(tableName string, dynaClient dynamodbiface.DynamoDBAPI) (*[]User, error) {

	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	// This is getting all the data in the table
	res, err := dynaClient.Scan(input)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	items := new([]User)
	err = dynamodbattribute.UnmarshalListOfMaps(res.Items, items)
	if err != nil {
		return nil, errors.New(ErrorFailedToFetchRecord)
	}

	return items, nil

}

func CreateUser(req events.APIGatewayProxyRequest, tablename string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User

	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidUserData)
	}

	if !validators.IsEmailValid(u.Email) {
		return nil, errors.New(ErrorInvalidUserData)
	}

	currentUser, _ := FetchUser(u.Email, tablename, dynaClient)

	if currentUser != nil && len(currentUser.Email) != 0 {
		return nil, errors.New(ErrorUserAlreadyExists)
	}

	av, err := dynamodbattribute.MarshalMap(u)

	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tablename),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotPutDynamoItem)
	}

	return &u, nil

}

func UpdateUser(req events.APIGatewayProxyRequest, tablename string, dynaClient dynamodbiface.DynamoDBAPI) (*User, error) {
	var u User

	if err := json.Unmarshal([]byte(req.Body), &u); err != nil {
		return nil, errors.New(ErrorInvalidEmail)
	}

	currentUser, _ := FetchUser(u.Email, tablename, dynaClient)
	if currentUser != nil && len(currentUser.Email) == 0 {
		return nil, errors.New(ErrorUserDoesNotExist)
	}

	av, err := dynamodbattribute.MarshalMap(u)
	if err != nil {
		return nil, errors.New(ErrorCouldNotMarshalItem)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tablename),
	}

	_, err = dynaClient.PutItem(input)
	if err != nil {
		return nil, errors.New(ErrorCouldNotPutDynamoItem)
	}

	return &u, nil

}

func DeleteUser(req events.APIGatewayProxyRequest, tablename string, dynaClient dynamodbiface.DynamoDBAPI) error {

	email := req.QueryStringParameters["email"]
	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(email),
			},
		},
		TableName: aws.String(tablename),
	}

	_, err := dynaClient.DeleteItem(input)
	if err != nil {
		return errors.New(ErrorCouldNotDeleteItem)
	}

	return nil

}

package rabbitmq

import (
	"fmt"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/michaelklishin/rabbit-hole"
)

func pathConfigConnection(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/connection",
		Fields: map[string]*framework.FieldSchema{
			"connection_uri": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "RabbitMQ Management URI",
			},
			"username": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Username of a RabbitMQ management administrator",
			},
			"password": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Password of the provided RabbitMQ management user",
			},
			"verify_connection": &framework.FieldSchema{
				Type:        framework.TypeBool,
				Default:     true,
				Description: `If set, connection_uri is verified by actually connecting to the RabbitMQ management API`,
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.UpdateOperation: b.pathConnectionUpdate,
		},

		HelpSynopsis:    pathConfigConnectionHelpSyn,
		HelpDescription: pathConfigConnectionHelpDesc,
	}
}

func (b *backend) pathConnectionUpdate(req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	uri := data.Get("connection_uri").(string)
	username := data.Get("username").(string)
	password := data.Get("password").(string)

	if uri == "" {
		return logical.ErrorResponse(fmt.Sprintf(
			"'connection_uri' is a required parameter.")), nil
	}

	if username == "" {
		return logical.ErrorResponse(fmt.Sprintf(
			"'username' is a required parameter.")), nil
	}

	if password == "" {
		return logical.ErrorResponse(fmt.Sprintf(
			"'password' is a required parameter.")), nil
	}

	// Don't check the connection_url if verification is disabled
	verifyConnection := data.Get("verify_connection").(bool)
	if verifyConnection {
		// Create RabbitMQ management client
		client, err := rabbithole.NewClient(uri, username, password)
		if err != nil {
			return logical.ErrorResponse(fmt.Sprintf(
				"Error  info: %s", err)), nil
		}

		// Verify provided user is able to list users
		_, err = client.ListUsers()
		if err != nil {
			return logical.ErrorResponse(fmt.Sprintf(
				"Error validating connection info by listing users: %s", err)), nil
		}
	}

	// Store it
	entry, err := logical.StorageEntryJSON("config/connection", connectionConfig{
		URI:      uri,
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(entry); err != nil {
		return nil, err
	}

	// Reset the client connection
	b.ResetClient()

	return nil, nil
}

type connectionConfig struct {
	URI       string `json:"connection_uri"`
	VerifyURI string `json:"verify_connection"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

const pathConfigConnectionHelpSyn = `
Configure the connection URI, username, and password to talk to RabbitMQ management HTTP API.
`

const pathConfigConnectionHelpDesc = `
This path configures the connection properties used to connect to RabbitMQ management HTTP API.
The "connection_uri" parameter is a string that is used to connect to the API. The "username"
and "password" parameters are strings that are used as credentials to the API. The "verify_connection"
parameter is a boolean that is used to verify whether the provided connection URI, username, and password
are valid.

The URI looks like:
"http://localhost:15672"
`

package client

import (
	"github.com/hashicorp/terraform-provider-azurerm/internal/common"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/chaos/sdk/2021-09-15-preview/targets"
)

type Client struct {
	TargetsClient *targets.TargetsClient
}

func NewClient(o *common.ClientOptions) *Client {
	targetsClient := targets.NewTargetsClientWithBaseURI(o.ResourceManagerEndpoint)
	o.ConfigureClient(&targetsClient.Client, o.ResourceManagerAuthorizer)

	return &Client{
		TargetsClient: &targetsClient,
	}
}

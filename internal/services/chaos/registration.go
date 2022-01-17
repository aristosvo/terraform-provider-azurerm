package chaos

import (
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
)

var _ sdk.TypedServiceRegistration = Registration{}

type Registration struct{}

func (r Registration) Name() string {
	return "Chaos Studio"
}

func (r Registration) WebsiteCategories() []string {
	return []string{
		"Chaos Studio",
	}
}

func (r Registration) DataSources() []sdk.DataSource {
	return []sdk.DataSource{}
}

func (r Registration) Resources() []sdk.Resource {
	resources := []sdk.Resource{
		TargetResource{},
	}

	return resources
}

package web_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/web/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type FunctionAppHostKeysResource struct{}

func TestAccFunctionAppHostKeysResource_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_function_app_host_keys", "test")
	r := FunctionAppHostKeysResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: FunctionAppHostKeysResource{}.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("primary_key").Exists(),
				check.That(data.ResourceName).Key("host_keys.0.default").Exists(),
				check.That(data.ResourceName).Key("host_keys.0.password").Exists(),
			),
		},
	})
}

func (r FunctionAppHostKeysResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.FunctionAppID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Web.AppServicesClient.Get(ctx, id.ResourceGroup, id.SiteName)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Function App %q (Resource Group %q): %+v", id.SiteName, id.ResourceGroup, err)
	}

	// The SDK defines 404 as an "ok" status code..
	if utils.ResponseWasNotFound(resp.Response) {
		return utils.Bool(false), nil
	}

	return utils.Bool(true), nil
}

func (d FunctionAppHostKeysResource) basic(data acceptance.TestData) string {
	template := FunctionAppResource{}.basic(data)
	return fmt.Sprintf(`
%s

resource "azurerm_function_app_host_keys" "test" {
  name                = azurerm_function_app.test.name
  resource_group_name = azurerm_resource_group.test.name

  host_keys = {
	  default  = "123234"
	  password = "H8$H1C0RP"
  }
}
`, template)
}

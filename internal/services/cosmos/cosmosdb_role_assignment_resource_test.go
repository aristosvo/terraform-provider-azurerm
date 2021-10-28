package cosmos_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cosmos/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type CosmosDBRoleAssignmentResource struct {
}

func TestAccCosmosDBRoleAssignment_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_role_assignment", "test")
	r := CosmosDBRoleAssignmentResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
	})
}

func TestAccCosmosDBRoleAssignment_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_cosmosdb_role_assignment", "test")
	r := CosmosDBRoleAssignmentResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.basic(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_cosmosdb_role_assignment"),
		},
	})
}

func (t CosmosDBRoleAssignmentResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.SqlRoleAssigmentID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Cosmos.SqlClient.GetSQLRoleAssignment(ctx, id.SqlRoleAssignmentName, id.ResourceGroup, id.DatabaseAccountName)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %+v", id, err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (CosmosDBRoleAssignmentResource) basic(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-cosmos-%d"
  location = "%s"
}

resource "azurerm_cosmosdb_account" "test" {
  name                = "acctest-ca-%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  offer_type          = "Standard"
  kind                = "GlobalDocumentDB"

  consistency_policy {
    consistency_level = "Eventual"
  }

  geo_location {
    location          = azurerm_resource_group.test.location
    failover_priority = 0
  }
}

data "azurerm_client_config" "test" {
}

resource "azurerm_cosmosdb_role_assignment" "test" {
  cosmosdb_account_name = azurerm_cosmosdb_account.test.name
  resource_group_name   = azurerm_resource_group.test.name
  scope                 = azurerm_cosmosdb_account.test.id
  role_definition_name  = "Cosmos DB Built-in Data Contributor"
  principal_id          = data.azurerm_client_config.test.object_id
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

func (r CosmosDBRoleAssignmentResource) requiresImport(data acceptance.TestData) string {
	return fmt.Sprintf(`
%s

resource "azurerm_cosmosdb_role_assignment" "import" {
  name                  = azurerm_cosmosdb_role_assignment.test.name
  cosmosdb_account_name = azurerm_cosmosdb_role_assignment.test.cosmosdb_account_name
  resource_group_name   = azurerm_cosmosdb_role_assignment.test.resource_group_name
  scope                 = azurerm_cosmosdb_role_assignment.test.scope
  role_definition_name  = azurerm_cosmosdb_role_assignment.test.role_definition_name
  principal_id          = azurerm_cosmosdb_role_assignment.test.principal_id
}
`, r.basic(data))
}

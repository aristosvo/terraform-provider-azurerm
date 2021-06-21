package containers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/acceptance/check"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/containers/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

type KubernetesMaintenanceConfigurationResource struct {
}

var kubernetesMaintenanceConfigTests = map[string]func(t *testing.T){
	"maintenanceConfig": testAccKubernetesMaintenanceConfig,
}

func TestAccKubernetesMaintenanceConfig(t *testing.T) {
	checkIfShouldRunTestsIndividually(t)
	testAccKubernetesMaintenanceConfig(t)
}

func testAccKubernetesMaintenanceConfig(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_kubernetes_maintenance_configuration", "test")
	r := KubernetesMaintenanceConfigurationResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			// Enabled
			Config: r.defaultConfig(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
				check.That(data.ResourceName).Key("tags.%").HasValue("0"),
			),
		},
		data.ImportStep(),
	})
}

func (t KubernetesMaintenanceConfigurationResource) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := parse.MaintenanceConfigurationID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.Containers.MaintenanceConfigurationsClient.Get(ctx, id.ResourceGroup, id.ManagedClusterName, id.Name)
	if err != nil {
		return nil, fmt.Errorf("reading Kubernetes Maintenance Configuration (%s): %+v", id.String(), err)
	}

	return utils.Bool(resp.ID != nil), nil
}

func (KubernetesMaintenanceConfigurationResource) templateConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-aks-%d"
  location = "%s"
}

resource "azurerm_kubernetes_cluster" "test" {
  name                = "acctestaks%d"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
  dns_prefix          = "acctestaks%d"

  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_DS2_v2"
  }

  identity {
    type = "SystemAssigned"
  }
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger)
}

func (r KubernetesMaintenanceConfigurationResource) defaultConfig(data acceptance.TestData) string {
	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

%s

resource "azurerm_kubernetes_maintenance_configuration" "test" {
  name                  = "internal"
  kubernetes_cluster_id = azurerm_kubernetes_cluster.test.id

  maintenance_allowed {
    day        = "Friday"
    hour_slots = [ 4 ]
  }

  maintenance_not_allowed_window {
    start = "2021-05-26T03:00:00Z"
    end   = "2021-05-30T12:00:00Z"
  }
}
`, r.templateConfig(data))
}

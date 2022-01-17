package chaos_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/chaos/sdk/2021-09-15-preview/targets"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type ChaosTargetResource struct {
}

func TestAccChaosTarget_cosmos(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_chaos_target", "test")
	r := ChaosTargetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.cosmos(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func TestAccChaosTarget_vm(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_chaos_target", "test")
	r := ChaosTargetResource{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.vm(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).Key("name").Exists(),
			),
		},
		data.ImportStep(),
	})
}

func (r ChaosTargetResource) Exists(ctx context.Context, client *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := targets.ParseTargetID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := client.Chaos.TargetsClient.Get(ctx, *id)
	if err != nil {
		if response.WasNotFound(resp.HttpResponse) {
			return utils.Bool(false), nil
		}
		return nil, fmt.Errorf("retrieving Chaos Target %s: %+v", id, err)
	}
	if response.WasNotFound(resp.HttpResponse) {
		return utils.Bool(false), nil
	}
	return utils.Bool(true), nil
}

func (ChaosTargetResource) vm(data acceptance.TestData) string {
	return fmt.Sprintf(`
# note: whilst these aren't used in all tests, it saves us redefining these everywhere
locals {
  first_public_key  = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC+wWK73dCr+jgQOAxNsHAnNNNMEMWOHYEccp6wJm2gotpr9katuF/ZAdou5AaW1C61slRkHRkpRRX9FA9CYBiitZgvCCz+3nWNN7l/Up54Zps/pHWGZLHNJZRYyAB6j5yVLMVHIHriY49d/GZTZVNB8GoJv9Gakwc/fuEZYYl4YDFiGMBP///TzlI4jhiJzjKnEvqPFki5p2ZRJqcbCiF4pJrxUQR/RXqVFQdbRLZgYfJ8xGB878RENq3yQ39d8dVOkq4edbkzwcUmwwwkYVPIoDGsYLaRHnG+To7FvMeyO7xDVQkMKzopTQV8AuKpyvpqu0a9pWOMaiCyDytO7GGN you@me.com"
}

resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_virtual_network" "test" {
  name                = "acctestnw-%d"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name
}

resource "azurerm_subnet" "test" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.test.name
  virtual_network_name = azurerm_virtual_network.test.name
  address_prefix       = "10.0.2.0/24"
}

resource "azurerm_network_interface" "test" {
	name                = "acctestnic-%d"
	location            = azurerm_resource_group.test.location
	resource_group_name = azurerm_resource_group.test.name
  
	ip_configuration {
	  name                          = "internal"
	  subnet_id                     = azurerm_subnet.test.id
	  private_ip_address_allocation = "Dynamic"
	}
  }

resource "azurerm_linux_virtual_machine" "test" {
	name                = "acctestVM-%d"
	resource_group_name = azurerm_resource_group.test.name
	location            = azurerm_resource_group.test.location
	size                = "Standard_F2"
	admin_username      = "adminuser"
	network_interface_ids = [
	  azurerm_network_interface.test.id,
	]
  
	admin_ssh_key {
	  username   = "adminuser"
	  public_key = local.first_public_key
	}
  
	os_disk {
	  caching              = "ReadWrite"
	  storage_account_type = "Standard_LRS"
	}
  
	source_image_reference {
	  publisher = "Canonical"
	  offer     = "UbuntuServer"
	  sku       = "16.04-LTS"
	  version   = "latest"
	}
  }

  resource "azurerm_chaos_target" "test" {
	name                = "Microsoft-Agent"
	location            = azurerm_resource_group.test.location
	resource_group_name = azurerm_resource_group.test.name
  
	parent_name               = azurerm_linux_virtual_machine.test.name
	parent_resource_type      = "virtualMachines"
	parent_provider_namespace = "Microsoft.Compute"
  
	depends_on = [azurerm_linux_virtual_machine.test]
  }

`, data.RandomInteger, data.Locations.Primary, data.RandomInteger, data.RandomInteger, data.RandomInteger)
}

func (ChaosTargetResource) cosmos(data acceptance.TestData) string {
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
    consistency_level = "BoundedStaleness"
  }

  geo_location {
    location          = azurerm_resource_group.test.location
    failover_priority = 0
  }
}

resource "azurerm_chaos_target" "test" {
  name                = "microsoft-cosmosdb"
  location            = azurerm_resource_group.test.location
  resource_group_name = azurerm_resource_group.test.name

  parent_name               = azurerm_cosmosdb_account.test.name
  parent_resource_type      = "databaseAccounts"
  parent_provider_namespace = "Microsoft.DocumentDB"

  depends_on = [azurerm_cosmosdb_account.test]
}
`, data.RandomInteger, data.Locations.Primary, data.RandomInteger)
}

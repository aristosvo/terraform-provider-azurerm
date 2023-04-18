package costmanagement_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/go-azure-sdk/resource-manager/costmanagement/2022-10-01/scheduledactions"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance"
	"github.com/hashicorp/terraform-provider-azurerm/internal/acceptance/check"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type SubscriptionCostManagementScheduledAction struct{}

func TestAccSubscriptionCostManagementScheduledAction_basic(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subscription_cost_management_scheduled_action", "test")
	r := SubscriptionCostManagementScheduledAction{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.daily(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubscriptionCostManagementScheduledAction_update(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subscription_cost_management_scheduled_action", "test")
	r := SubscriptionCostManagementScheduledAction{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.daily(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.monthDay(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.weekly(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
		{
			Config: r.monthWeekDay(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		data.ImportStep(),
	})
}

func TestAccSubscriptionCostManagementScheduledAction_requiresImport(t *testing.T) {
	data := acceptance.BuildTestData(t, "azurerm_subscription_cost_management_scheduled_action", "test")
	r := SubscriptionCostManagementScheduledAction{}

	data.ResourceTest(t, r, []acceptance.TestStep{
		{
			Config: r.daily(data),
			Check: acceptance.ComposeTestCheckFunc(
				check.That(data.ResourceName).ExistsInAzure(r),
			),
		},
		{
			Config:      r.requiresImport(data),
			ExpectError: acceptance.RequiresImportError("azurerm_subscription_cost_management_scheduled_action"),
		},
	})
}

func (t SubscriptionCostManagementScheduledAction) Exists(ctx context.Context, clients *clients.Client, state *pluginsdk.InstanceState) (*bool, error) {
	id, err := scheduledactions.ParseScopedScheduledActionID(state.ID)
	if err != nil {
		return nil, err
	}

	resp, err := clients.CostManagement.ScheduledActionsClient_v2022_10_01.GetByScope(ctx, *id)
	if err != nil {
		return nil, fmt.Errorf("retrieving (%s): %+v", *id, err)
	}

	return utils.Bool(resp.Model != nil), nil
}

func (SubscriptionCostManagementScheduledAction) daily(data acceptance.TestData) string {
	start := time.Now().AddDate(0, 0, 1).UTC().Format("2006-01-02T00:00:00Z")
	end := time.Now().AddDate(0, 0, 2).UTC().Format("2006-01-02T00:00:00Z")

	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "test" {}

resource "azurerm_subscription_cost_management_scheduled_action" "test" {
  name            = "testcostview%s"
  subscription_id = data.azurerm_subscription.test.id

  view_id = "${data.azurerm_subscription.test.id}/providers/Microsoft.CostManagement/views/ms:CostByService"

  display_name         = "CostByServiceView%s"
  email_subject        = substr("Cost Management Report for ${data.azurerm_subscription.test.display_name} Subscription", 0, 70)
  email_addresses      = ["test@test.com", "hashicorp@test.com"]
  email_address_sender = "test@test.com"

  frequency  = "Daily"
  start_date = "%s"
  end_date   = "%s"
}
`, data.RandomString, data.RandomString, start, end)
}

func (SubscriptionCostManagementScheduledAction) monthWeekDay(data acceptance.TestData) string {
	start := time.Now().AddDate(0, 0, 1).UTC().Format("2006-01-02T00:00:00Z")
	end := time.Now().AddDate(0, 0, 2).UTC().Format("2006-01-02T00:00:00Z")

	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "test" {}

resource "azurerm_subscription_cost_management_scheduled_action" "test" {
  name            = "testcostview%s"
  subscription_id = data.azurerm_subscription.test.id

  view_id = "${data.azurerm_subscription.test.id}/providers/Microsoft.CostManagement/views/ms:CostByService"

  display_name         = "CostByServiceView%s"
  message              = "Hi"
  email_subject        = substr("Cost Management Report for ${data.azurerm_subscription.test.display_name} Subscription", 0, 70)
  email_addresses      = ["test@test.com", "hashicorp@test.com"]
  email_address_sender = "test@test.com"

  days_of_week   = ["Monday"]
  weeks_of_month = ["First"]
  frequency      = "Monthly"
  start_date     = "%s"
  end_date       = "%s"
}
`, data.RandomString, data.RandomString, start, end)
}

func (SubscriptionCostManagementScheduledAction) monthDay(data acceptance.TestData) string {
	start := time.Now().AddDate(0, 0, 1).UTC().Format("2006-01-02T00:00:00Z")
	end := time.Now().AddDate(0, 0, 2).UTC().Format("2006-01-02T00:00:00Z")

	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "test" {}

resource "azurerm_subscription_cost_management_scheduled_action" "test" {
  name            = "testcostview%s"
  subscription_id = data.azurerm_subscription.test.id

  view_id = "${data.azurerm_subscription.test.id}/providers/Microsoft.CostManagement/views/ms:CostByService"

  display_name         = "CostByServiceView%s"
  message              = "Hi"
  email_subject        = substr("Cost Management Report for ${data.azurerm_subscription.test.display_name} Subscription", 0, 70)
  email_addresses      = ["test@test.com", "hashicorp@test.com"]
  email_address_sender = "test@test.com"

  hour_of_day  = 23
  day_of_month = 30
  frequency    = "Monthly"
  start_date   = "%s"
  end_date     = "%s"
}
`, data.RandomString, data.RandomString, start, end)
}

func (SubscriptionCostManagementScheduledAction) weekly(data acceptance.TestData) string {
	start := time.Now().AddDate(0, 0, 1).UTC().Format("2006-01-02T00:00:00Z")
	end := time.Now().AddDate(0, 0, 2).UTC().Format("2006-01-02T00:00:00Z")

	return fmt.Sprintf(`
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "test" {}

resource "azurerm_subscription_cost_management_scheduled_action" "test" {
  name            = "testcostview%s"
  subscription_id = data.azurerm_subscription.test.id

  view_id = "${data.azurerm_subscription.test.id}/providers/Microsoft.CostManagement/views/ms:CostByService"

  display_name         = "CostByServiceView%s"
  message              = "Hi"
  email_subject        = substr("Cost Management Report for ${data.azurerm_subscription.test.display_name} Subscription", 0, 70)
  email_addresses      = ["test@test.com", "hashicorp@test.com"]
  email_address_sender = "test@test.com"

  days_of_week = ["Friday"]
  hour_of_day  = 0
  frequency    = "Weekly"
  start_date   = "%s"
  end_date     = "%s"
}
`, data.RandomString, data.RandomString, start, end)
}

func (SubscriptionCostManagementScheduledAction) requiresImport(data acceptance.TestData) string {
	template := SubscriptionCostManagementScheduledAction{}.daily(data)
	return fmt.Sprintf(`
%s

resource "azurerm_subscription_cost_management_scheduled_action" "import" {
  name            = azurerm_subscription_cost_management_scheduled_action.test.name
  subscription_id = azurerm_subscription_cost_management_scheduled_action.test.subscription_id

  view_id = azurerm_subscription_cost_management_scheduled_action.test.view_id

  display_name         = azurerm_subscription_cost_management_scheduled_action.test.display_name
  email_subject        = azurerm_subscription_cost_management_scheduled_action.test.email_subject
  email_addresses      = azurerm_subscription_cost_management_scheduled_action.test.email_addresses
  email_address_sender = azurerm_subscription_cost_management_scheduled_action.test.email_address_sender

  frequency  = azurerm_subscription_cost_management_scheduled_action.test.frequency
  start_date = azurerm_subscription_cost_management_scheduled_action.test.start_date
  end_date   = azurerm_subscription_cost_management_scheduled_action.test.end_date
}
`, template)
}

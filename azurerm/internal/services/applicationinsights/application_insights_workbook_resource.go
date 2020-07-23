package applicationinsights

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/appinsights/mgmt/2015-05-01/insights"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tags"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceArmApplicationInsightsWorkBook() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmApplicationInsightsWorkBookCreateUpdate,
		Read:   resourceArmApplicationInsightsWorkBookRead,
		Update: resourceArmApplicationInsightsWorkBookCreateUpdate,
		Delete: resourceArmApplicationInsightsWorkBookDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"application_insights_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"location": azure.SchemaLocation(),

			"kind": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					string(insights.SharedTypeKindShared),
					string(insights.SharedTypeKindUser),
				}, false),
			},

			"tags": tags.Schema(),

			"workbook": {
				Type:         schema.TypeString,
				Required:     true,
			},
		},
	}
}

func resourceArmApplicationInsightsWorkBookCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppInsights.WorkbooksClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	log.Printf("[INFO] preparing arguments for AzureRM Application Insights WorkBook creation.")

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	appInsightsID := d.Get("application_insights_id").(string)
	workbook := d.Get("workbook").(string)

	id, err := azure.ParseAzureResourceID(appInsightsID)
	if err != nil {
		return err
	}

	appInsightsName := id.Path["components"]

	if d.IsNewResource() {
		existing, err := client.Get(ctx, resGroup, name)
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("Error checking for presence of existing Application Insights WorkBook %q (Resource Group %q): %s", name, resGroup, err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_application_insights_workbook", *existing.ID)
		}
	}

	location := azure.NormalizeLocation(d.Get("location").(string))
	kind := d.Get("kind").(string)

	t := d.Get("tags").(map[string]interface{})
	tagKey := fmt.Sprintf("hidden-link:/subscriptions/%s/resourceGroups/%s/providers/microsoft.insights/components/%s", client.SubscriptionID, resGroup, appInsightsName)
	t[tagKey] = "Resource"

	workBook := insights.Workbook{
		WorkbookProperties: *insights.WorkbookProperties{
			Category: *string,
			Name:     &name,
			SerializedData: *string,
			SharedTypeKind: insights.SharedTypeKind,
			SourceResourceID: *string,
			UserID: *string,
			Tags: *[]string,
		}
		Location: &location,
		Kind:     insights.SharedTypeKind(kind),
		Tags: tags.Expand(t),
	}

	resp, err := client.CreateOrUpdate(ctx, resGroup, name, workBook)
	if err != nil {
		return fmt.Errorf("Error creating Application Insights WorkBook %q (Resource Group %q): %+v", name, resGroup, err)
	}

	d.SetId(*resp.ID)

	return resourceArmApplicationInsightsWorkBookRead(d, meta)
}

func resourceArmApplicationInsightsWorkBookRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppInsights.WorkbooksClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Reading AzureRM Application Insights WorkBook '%s'", id)

	resGroup := id.ResourceGroup
	name := id.Path["webtests"]

	resp, err := client.Get(ctx, resGroup, name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Application Insights WorkBook %q was not found in Resource Group %q - removing from state!", name, resGroup)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error retrieving Application Insights WorkBook %q (Resource Group %q): %+v", name, resGroup, err)
	}

	appInsightsId := ""
	for i := range resp.Tags {
		if strings.HasPrefix(i, "hidden-link") {
			appInsightsId = strings.Split(i, ":")[1]
		}
	}
	d.Set("application_insights_id", appInsightsId)
	d.Set("name", resp.Name)
	d.Set("resource_group_name", resGroup)
	d.Set("kind", resp.Kind)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if props := resp.WorkBookProperties; props != nil {
		d.Set("synthetic_monitor_id", props.SyntheticMonitorID)
		d.Set("description", props.Description)
		d.Set("enabled", props.Enabled)
		d.Set("frequency", props.Frequency)
		d.Set("timeout", props.Timeout)
		d.Set("retry_enabled", props.RetryEnabled)

		if config := props.Configuration; config != nil {
			d.Set("configuration", config.WorkBook)
		}

		if err := d.Set("geo_locations", flattenApplicationInsightsWorkBookGeoLocations(props.Locations)); err != nil {
			return fmt.Errorf("Error setting `geo_locations`: %+v", err)
		}
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmApplicationInsightsWorkBookDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).AppInsights.WorkbooksClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := azure.ParseAzureResourceID(d.Id())
	if err != nil {
		return err
	}
	resGroup := id.ResourceGroup
	name := id.Path["webtests"]

	log.Printf("[DEBUG] Deleting AzureRM Application Insights WorkBook '%s' (resource group '%s')", name, resGroup)

	resp, err := client.Delete(ctx, resGroup, name)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			return nil
		}
		return fmt.Errorf("Error issuing AzureRM delete request for Application Insights WorkBook '%s': %+v", name, err)
	}

	return err
}

func expandApplicationInsightsWorkBookGeoLocations(input []interface{}) []insights.WorkBookGeolocation {
	locations := make([]insights.WorkBookGeolocation, 0)

	for _, v := range input {
		lc := v.(string)
		loc := insights.WorkBookGeolocation{
			Location: &lc,
		}
		locations = append(locations, loc)
	}

	return locations
}

func flattenApplicationInsightsWorkBookGeoLocations(input *[]insights.WorkBookGeolocation) []string {
	results := make([]string, 0)
	if input == nil {
		return results
	}

	for _, prop := range *input {
		if prop.Location != nil {
			results = append(results, azure.NormalizeLocation(*prop.Location))
		}
	}

	return results
}

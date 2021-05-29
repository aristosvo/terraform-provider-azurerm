package web

import (
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2020-06-01/web"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/azure"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/web/parse"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceFunctionAppHostKeys() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceFunctionAppHostKeysCreateUpdate,
		Read:   resourceFunctionAppHostKeysRead,
		Update: resourceFunctionAppHostKeysCreateUpdate,
		Delete: resourceFunctionAppHostKeysDelete,

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:     pluginsdk.TypeString,
				Required: true,
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"primary_key": {
				Type:      pluginsdk.TypeString,
				Computed:  true,
				Sensitive: true,
			},

			"host_keys": {
				Type:     pluginsdk.TypeMap,
				Optional: true,
				Computed: true,
				Elem: &pluginsdk.Schema{
					Type: pluginsdk.TypeString,
				},
			},
		},
	}
}

func resourceFunctionAppHostKeysCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Web.AppServicesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	if d.HasChange("host_keys") {
		input := d.Get("host_keys").(map[string]interface{})

		for k, v := range input {
			keyInfo := &web.KeyInfo{
				Value: utils.String(v.(string)),
			}
			_, err := client.CreateOrUpdateHostSecret(ctx, resourceGroup, name, "functionKeys", k, *keyInfo)
			if err != nil {
				return fmt.Errorf("updating host secret %s for AzureRM Function App %q (Resource Group %q): %s", k, name, resourceGroup, err)
			}
		}

		// delete removed keys
		old, new := d.GetChange("host_keys")
		newKeys := new.(map[string]interface{})
		for oldKey := range old.(map[string]interface{}) {
			if newKeys[oldKey] == nil || newKeys[oldKey].(string) == "" {
				res, err := client.DeleteHostSecret(ctx, resourceGroup, name, "functionKeys", oldKey)
				if err != nil {
					if utils.ResponseWasNotFound(res) {
						continue
					}
					return fmt.Errorf("removing host secret %s for AzureRM Function App %q (Resource Group %q: %s", oldKey, name, resourceGroup, err)
				}
			}
		}
	}

	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	id := parse.NewFunctionAppID(subscriptionId, resourceGroup, name)
	d.SetId(id.ID())

	return resourceFunctionAppHostKeysRead(d, meta)
}

func resourceFunctionAppHostKeysRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Web.AppServicesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.FunctionAppID(d.Id())
	if err != nil {
		return err
	}

	return pluginsdk.Retry(d.Timeout(pluginsdk.TimeoutCreate), func() *pluginsdk.RetryError {
		res, err := client.ListHostKeys(ctx, id.ResourceGroup, id.SiteName)
		if err != nil {
			if utils.ResponseWasNotFound(res.Response) {
				return pluginsdk.NonRetryableError(fmt.Errorf("AzureRM Function App %q (Resource Group %q) was not found", id.SiteName, id.ResourceGroup))
			}

			return pluginsdk.RetryableError(fmt.Errorf("making Read request on AzureRM Function App Hostkeys %q: %+v", id.SiteName, err))
		}

		d.Set("primary_key", res.MasterKey)
		hostKeys := flattenFunctionKeys(res.FunctionKeys)
		if err := d.Set("host_keys", hostKeys); err != nil {
			return pluginsdk.NonRetryableError(fmt.Errorf("setting `host_keys`: %s", err))
		}

		return nil
	})
}

func resourceFunctionAppHostKeysDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Web.AppServicesClient
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	resourceGroup := d.Get("resource_group_name").(string)
	name := d.Get("name").(string)

	// delete removed keys
	keys := d.Get("host_keys").(map[string]interface{})
	for key := range keys {
		res, err := client.DeleteHostSecret(ctx, resourceGroup, name, "functionKeys", key)
		if err != nil {
			if utils.ResponseWasNotFound(res) {
				continue
			}
			return fmt.Errorf("removing host secret %s for AzureRM Function App %q (Resource Group %q: %s", key, name, resourceGroup, err)
		}

	}

	return nil
}

func flattenFunctionKeys(input map[string]*string) map[string]string {
	output := make(map[string]string)
	for k, v := range input {
		output[k] = *v
	}

	return output
}

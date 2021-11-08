package sql

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/sql/mgmt/v5.0/sql"
	"github.com/gofrs/uuid"

	"github.com/hashicorp/go-azure-helpers/response"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/tf"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/identity"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/mssql/validate"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/sql/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tags"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type managedInstanceIdentity = identity.SystemAssigned

func resourceArmSqlMiServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceArmSqlMiServerCreateUpdate,
		Read:   resourceArmSqlMiServerRead,
		Update: resourceArmSqlMiServerCreateUpdate,
		Delete: resourceArmSqlMiServerDelete,
		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.ManagedInstanceID(id)
			return err
		}),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(24 * time.Hour),
			Read:   schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(24 * time.Hour),
			Delete: schema.DefaultTimeout(24 * time.Hour),
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validate.ValidateMsSqlServerName,
			},

			"location": azure.SchemaLocation(),

			"resource_group_name": azure.SchemaResourceGroupName(),

			"sku_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"GP_Gen4",
					"GP_Gen5",
					"BC_Gen4",
					"BC_Gen5",
				}, false),
			},

			"administrator_login": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"administrator_login_password": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"vcores": {
				Type:     schema.TypeInt,
				Required: true,
				ValidateFunc: validation.IntInSlice([]int{
					4,
					8,
					16,
					24,
					32,
					40,
					64,
					80,
				}),
			},

			"storage_size_in_gb": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(32, 8192),
			},

			"license_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"LicenseIncluded",
					"BasePrice",
				}, true),
			},

			"subnet_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"collation": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "SQL_Latin1_General_CP1_CI_AS",
				ValidateFunc: validation.StringIsNotEmpty,
				ForceNew:     true,
			},

			"public_data_endpoint_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"minimum_tls_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1.2",
				ValidateFunc: validation.StringInSlice([]string{
					"1.0",
					"1.1",
					"1.2",
				}, false),
			},

			"proxy_override": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  string(sql.ManagedInstanceProxyOverrideDefault),
				ValidateFunc: validation.StringInSlice([]string{
					string(sql.ManagedInstanceProxyOverrideDefault),
					string(sql.ManagedInstanceProxyOverrideRedirect),
					string(sql.ManagedInstanceProxyOverrideProxy),
				}, false),
			},

			"timezone_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "UTC",
				ValidateFunc: validation.StringIsNotEmpty,
				ForceNew:     true,
			},

			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"dns_zone_partner_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: azure.ValidateResourceID,
			},

			"identity": managedInstanceIdentity{}.Schema(),

			"azuread_administrator": {
				Type:     pluginsdk.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"login_username": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},

						"object_id": {
							Type:         pluginsdk.TypeString,
							Required:     true,
							ValidateFunc: validation.IsUUID,
						},

						"tenant_id": {
							Type:         pluginsdk.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsUUID,
						},

						"azuread_authentication_only": {
							Type:     pluginsdk.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},

			"tags": tags.Schema(),
		},

		CustomizeDiff: pluginsdk.CustomDiffWithAll(
			// dns_zone_partner_id can only be set on init
			pluginsdk.ForceNewIfChange("dns_zone_partner_id", func(ctx context.Context, old, new, _ interface{}) bool {
				return old.(string) == "" && new.(string) != ""
			}),
		),
	}
}

func resourceArmSqlMiServerCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Sql.ManagedInstancesClient
	adminClient := meta.(*clients.Client).Sql.ManagedInstanceAdministratorsClient
	aadOnlyAuthentictionsClient := meta.(*clients.Client).Sql.ManagedInstanceAzureADOnlyAuthenticationsClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	name := d.Get("name").(string)
	resGroup := d.Get("resource_group_name").(string)
	id := parse.NewManagedInstanceID(subscriptionId, resGroup, name)

	if d.IsNewResource() {
		existing, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
		if err != nil {
			if !utils.ResponseWasNotFound(existing.Response) {
				return fmt.Errorf("checking for presence of existing Managed Instance %q: %s", id.ID(), err)
			}
		}

		if existing.ID != nil && *existing.ID != "" {
			return tf.ImportAsExistsError("azurerm_sql_managed_instance", *existing.ID)
		}
	}

	sku, err := expandManagedInstanceSkuName(d.Get("sku_name").(string))
	if err != nil {
		return fmt.Errorf("error expanding `sku_name` for SQL Managed Instance Server %q: %v", id.ID(), err)
	}

	parameters := sql.ManagedInstance{
		Sku:      sku,
		Location: utils.String(azure.NormalizeLocation(d.Get("location").(string))),
		Tags:     tags.Expand(d.Get("tags").(map[string]interface{})),
		ManagedInstanceProperties: &sql.ManagedInstanceProperties{
			LicenseType:                sql.ManagedInstanceLicenseType(d.Get("license_type").(string)),
			AdministratorLogin:         utils.String(d.Get("administrator_login").(string)),
			AdministratorLoginPassword: utils.String(d.Get("administrator_login_password").(string)),
			SubnetID:                   utils.String(d.Get("subnet_id").(string)),
			StorageSizeInGB:            utils.Int32(int32(d.Get("storage_size_in_gb").(int))),
			VCores:                     utils.Int32(int32(d.Get("vcores").(int))),
			Collation:                  utils.String(d.Get("collation").(string)),
			PublicDataEndpointEnabled:  utils.Bool(d.Get("public_data_endpoint_enabled").(bool)),
			MinimalTLSVersion:          utils.String(d.Get("minimum_tls_version").(string)),
			ProxyOverride:              sql.ManagedInstanceProxyOverride(d.Get("proxy_override").(string)),
			TimezoneID:                 utils.String(d.Get("timezone_id").(string)),
			DNSZonePartner:             utils.String(d.Get("dns_zone_partner_id").(string)),
		},
	}

	if azureADAdministrator, ok := d.GetOk("azuread_administrator"); d.IsNewResource() && ok {
		parameters.ManagedInstanceProperties.Administrators = expandMsSqlInstanceAdministrators(azureADAdministrator.([]interface{}))
	}

	identity, err := expandManagedInstanceIdentity(d.Get("identity").([]interface{}))
	if err != nil {
		return fmt.Errorf(`expanding "identity": %v`, err)
	}
	parameters.Identity = identity

	future, err := client.CreateOrUpdate(ctx, resGroup, name, parameters)
	if err != nil {
		return err
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		if response.WasConflict(future.Response()) {
			return fmt.Errorf("sql managed instance names need to be globally unique and %q is already in use", name)
		}

		return err
	}

	if d.HasChange("azuread_administrator") && !d.IsNewResource() {
		aadOnlyDeleteFuture, err := aadOnlyAuthentictionsClient.Delete(ctx, id.ResourceGroup, id.Name)
		if err != nil {
			if aadOnlyDeleteFuture.Response() == nil || aadOnlyDeleteFuture.Response().StatusCode != http.StatusBadRequest {
				return fmt.Errorf("deleting AD Only Authentications %s: %+v", id.String(), err)
			}
			log.Printf("[INFO] AD Only Authentication is not removed as AD Admin is not set for %s: %+v", id.String(), err)
		} else if err = aadOnlyDeleteFuture.WaitForCompletionRef(ctx, adminClient.Client); err != nil {
			return fmt.Errorf("waiting for deletion of AD Only Authentications %s: %+v", id.String(), err)
		}

		if adminParams := expandMsSqlInstanceAdministrator(d.Get("azuread_administrator").([]interface{})); adminParams != nil {
			adminFuture, err := adminClient.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, *adminParams)
			if err != nil {
				return fmt.Errorf("creating AAD admin %s: %+v", id.String(), err)
			}

			if err = adminFuture.WaitForCompletionRef(ctx, adminClient.Client); err != nil {
				return fmt.Errorf("waiting for creation of AAD admin %s: %+v", id.String(), err)
			}

			if aadOnlyAuthentictionsEnabled := expandMsSqlInstanceAADOnlyAuthentictions(d.Get("azuread_administrator").([]interface{})); aadOnlyAuthentictionsEnabled {
				aadOnlyAuthentictionsParams := sql.ManagedInstanceAzureADOnlyAuthentication{
					ManagedInstanceAzureADOnlyAuthProperties: &sql.ManagedInstanceAzureADOnlyAuthProperties{
						AzureADOnlyAuthentication: utils.Bool(aadOnlyAuthentictionsEnabled),
					},
				}
				aadOnlyEnabledFuture, err := aadOnlyAuthentictionsClient.CreateOrUpdate(ctx, id.ResourceGroup, id.Name, aadOnlyAuthentictionsParams)
				if err != nil {
					return fmt.Errorf("setting AAD only authentication for %s: %+v", id.String(), err)
				}

				if err = aadOnlyEnabledFuture.WaitForCompletionRef(ctx, adminClient.Client); err != nil {
					return fmt.Errorf("waiting for setting of AAD only authentication for %s: %+v", id.String(), err)
				}
			}
		} else {
			adminDelFuture, err := adminClient.Delete(ctx, id.ResourceGroup, id.Name)
			if err != nil {
				return fmt.Errorf("deleting AAD admin  %s: %+v", id.String(), err)
			}

			if err = adminDelFuture.WaitForCompletionRef(ctx, adminClient.Client); err != nil {
				return fmt.Errorf("waiting for deletion of AAD admin %s: %+v", id.String(), err)
			}
		}
	}

	d.SetId(id.ID())

	return resourceArmSqlMiServerRead(d, meta)
}

func resourceArmSqlMiServerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Sql.ManagedInstancesClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ManagedInstanceID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.Name, "")
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[INFO] Error reading SQL Managed Instance %q - removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("reading SQL Managed Instance %q: %v", id.ID(), err)
	}

	d.Set("name", id.Name)
	d.Set("resource_group_name", id.ResourceGroup)

	if location := resp.Location; location != nil {
		d.Set("location", azure.NormalizeLocation(*location))
	}

	if sku := resp.Sku; sku != nil {
		d.Set("sku_name", sku.Name)
	}

	if err := d.Set("identity", flattenManagedInstanceIdentity(resp.Identity)); err != nil {
		return fmt.Errorf("setting `identity`: %+v", err)
	}

	if props := resp.ManagedInstanceProperties; props != nil {
		d.Set("license_type", props.LicenseType)
		d.Set("administrator_login", props.AdministratorLogin)
		d.Set("subnet_id", props.SubnetID)
		d.Set("storage_size_in_gb", props.StorageSizeInGB)
		d.Set("vcores", props.VCores)
		d.Set("fqdn", props.FullyQualifiedDomainName)
		d.Set("collation", props.Collation)
		d.Set("public_data_endpoint_enabled", props.PublicDataEndpointEnabled)
		d.Set("minimum_tls_version", props.MinimalTLSVersion)
		d.Set("proxy_override", props.ProxyOverride)
		d.Set("timezone_id", props.TimezoneID)
		// This value is not returned from the api so we'll just set whatever is in the config
		d.Set("administrator_login_password", d.Get("administrator_login_password").(string))
	}

	return tags.FlattenAndSet(d, resp.Tags)
}

func resourceArmSqlMiServerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Sql.ManagedInstancesClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.ManagedInstanceID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.Delete(ctx, id.ResourceGroup, id.Name)
	if err != nil {
		return fmt.Errorf("deleting SQL Managed Instance %q: %+v", id.ID(), err)
	}

	return future.WaitForCompletionRef(ctx, client.Client)
}

func expandManagedInstanceSkuName(skuName string) (*sql.Sku, error) {
	parts := strings.Split(skuName, "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("sku_name (%s) has the wrong number of parts (%d) after splitting on _", skuName, len(parts))
	}

	var tier string
	switch parts[0] {
	case "GP":
		tier = "GeneralPurpose"
	case "BC":
		tier = "BusinessCritical"
	default:
		return nil, fmt.Errorf("sku_name %s has unknown sku tier %s", skuName, parts[0])
	}

	return &sql.Sku{
		Name:   utils.String(skuName),
		Tier:   utils.String(tier),
		Family: utils.String(parts[1]),
	}, nil
}

func expandManagedInstanceIdentity(input []interface{}) (*sql.ResourceIdentity, error) {
	config, err := managedInstanceIdentity{}.Expand(input)
	if err != nil {
		return nil, err
	}

	return &sql.ResourceIdentity{
		Type: sql.IdentityType(config.Type),
	}, nil
}

func flattenManagedInstanceIdentity(input *sql.ResourceIdentity) []interface{} {
	var config *identity.ExpandedConfig

	if input == nil {
		return []interface{}{}
	}

	principalId := ""
	if input.PrincipalID != nil {
		principalId = input.PrincipalID.String()
	}

	tenantId := ""
	if input.TenantID != nil {
		tenantId = input.TenantID.String()
	}

	config = &identity.ExpandedConfig{
		Type:        identity.Type(string(input.Type)),
		PrincipalId: principalId,
		TenantId:    tenantId,
	}
	return managedInstanceIdentity{}.Flatten(config)
}

func expandMsSqlInstanceAdministrator(input []interface{}) *sql.ManagedInstanceAdministrator {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	admin := input[0].(map[string]interface{})
	sid, _ := uuid.FromString(admin["object_id"].(string))

	adminParams := sql.ManagedInstanceAdministrator{
		ManagedInstanceAdministratorProperties: &sql.ManagedInstanceAdministratorProperties{
			AdministratorType: utils.String("ActiveDirectory"),
			Login:             utils.String(admin["login_username"].(string)),
			Sid:               &sid,
		},
	}

	if v, ok := admin["tenant_id"]; ok && v != "" {
		tid, _ := uuid.FromString(v.(string))
		adminParams.TenantID = &tid
	}

	return &adminParams
}

func expandMsSqlInstanceAdministrators(input []interface{}) *sql.ManagedInstanceExternalAdministrator {
	if len(input) == 0 || input[0] == nil {
		return nil
	}

	admin := input[0].(map[string]interface{})
	sid, _ := uuid.FromString(admin["object_id"].(string))

	adminParams := sql.ManagedInstanceExternalAdministrator{
		AdministratorType: sql.AdministratorTypeActiveDirectory,
		Login:             utils.String(admin["login_username"].(string)),
		Sid:               &sid,
	}

	if v, ok := admin["tenant_id"]; ok && v != "" {
		tid, _ := uuid.FromString(v.(string))
		adminParams.TenantID = &tid
	}

	return &adminParams
}

func flatternMsSqlInstanceAdministrators(admin sql.ManagedInstanceExternalAdministrator) []interface{} {
	var login, sid, tid string
	if admin.Login != nil {
		login = *admin.Login
	}

	if admin.Sid != nil {
		sid = admin.Sid.String()
	}

	if admin.TenantID != nil {
		tid = admin.TenantID.String()
	}

	var aadOnlyAuthentictionsEnabled bool
	if admin.AzureADOnlyAuthentication != nil {
		aadOnlyAuthentictionsEnabled = *admin.AzureADOnlyAuthentication
	}

	return []interface{}{
		map[string]interface{}{
			"login_username":              login,
			"object_id":                   sid,
			"tenant_id":                   tid,
			"azuread_authentication_only": aadOnlyAuthentictionsEnabled,
		},
	}
}

func expandMsSqlInstanceAADOnlyAuthentictions(input []interface{}) bool {
	if len(input) == 0 || input[0] == nil {
		return false
	}
	admin := input[0].(map[string]interface{})
	if v, ok := admin["azuread_authentication_only"]; ok && v != nil {
		return v.(bool)
	}
	return false
}

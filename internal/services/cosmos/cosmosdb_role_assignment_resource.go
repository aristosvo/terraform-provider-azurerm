package cosmos

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2021-06-15/documentdb"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/internal/clients"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/cosmos/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/suppress"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/internal/timeouts"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

func resourceCosmosDbRoleAssignment() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceCosmosDbRoleAssignmentCreateUpdate,
		Read:   resourceCosmosDbRoleAssignmentRead,
		// Update: resourceCosmosDbRoleAssignmentCreateUpdate,
		Delete: resourceCosmosDbRoleAssignmentDelete,

		Importer: pluginsdk.DefaultImporter(),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(30 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			// Update: pluginsdk.DefaultTimeout(30 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.IsUUID,
			},

			"cosmosdb_account_name": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexp.MustCompile("^[-a-z0-9]{3,50}$"),
					"Cosmos DB Account name must be 3 - 50 characters long, contain only lowercase letters, numbers and hyphens.",
				),
			},

			"resource_group_name": azure.SchemaResourceGroupName(),

			"scope": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			"role_definition_id": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"role_definition_name"},
				DiffSuppressFunc: suppress.CaseDifference,
			},

			"role_definition_name": {
				Type:             pluginsdk.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"role_definition_id"},
				DiffSuppressFunc: suppress.CaseDifference,
				ValidateFunc:     validation.StringIsNotEmpty,
			},

			"principal_id": {
				Type:     pluginsdk.TypeString,
				Required: true,
				ForceNew: true,
			},

			// "principal_type": {
			// 	Type:     pluginsdk.TypeString,
			// 	Computed: true,
			// },

			// "skip_service_principal_aad_check": {
			// 	Type:     pluginsdk.TypeBool,
			// 	Optional: true,
			// 	Computed: true,
			// },

			// "delegated_managed_identity_resource_id": {
			// 	Type:         pluginsdk.TypeString,
			// 	Optional:     true,
			// 	ForceNew:     true,
			// 	ValidateFunc: azure.ValidateResourceID,
			// },

			// "description": {
			// 	Type:         pluginsdk.TypeString,
			// 	Optional:     true,
			// 	ForceNew:     true,
			// 	ValidateFunc: validation.StringIsNotEmpty,
			// },
		},
	}
}

func resourceCosmosDbRoleAssignmentCreateUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.SqlClient
	subscriptionId := meta.(*clients.Client).Account.SubscriptionId
	ctx, cancel := timeouts.ForCreateUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()
	log.Printf("[INFO] preparing arguments for AzureRM Cosmos DB Account creation.")

	name := d.Get("name").(string)
	if name == "" {
		uuid, err := uuid.GenerateUUID()
		if err != nil {
			return fmt.Errorf("generating UUID for Role Assignment: %+v", err)
		}

		name = uuid
	}

	id := parse.NewSqlRoleAssigmentID(subscriptionId, d.Get("resource_group_name").(string), d.Get("cosmosdb_account_name").(string), name)

	var roleDefinitionId string
	if v, ok := d.GetOk("role_definition_id"); ok {
		roleDefinitionId = v.(string)
	} else if v, ok := d.GetOk("role_definition_name"); ok {
		roleName := v.(string)
		roleDefinitionsRaw, err := client.ListSQLRoleDefinitions(ctx, id.ResourceGroup, id.DatabaseAccountName)
		if err != nil {
			return fmt.Errorf("loading Role Definition List: %+v", err)
		}
		if roleDefinitionsRaw.Value == nil {
			return fmt.Errorf("loading Role Definition List: could not find role '%s'", roleName)
		}
		for _, roleDefinitionResult := range *roleDefinitionsRaw.Value {
			if roleDefinitionResult.SQLRoleDefinitionResource == nil || roleDefinitionResult.SQLRoleDefinitionResource.RoleName == nil {
				continue
			}
			if *roleDefinitionResult.SQLRoleDefinitionResource.RoleName == roleName {
				roleDefinitionId = *roleDefinitionResult.ID
				break
			}
		}
	} else {
		return fmt.Errorf("either role_definition_id or role_definition_name needs to be set")
	}
	d.Set("role_definition_id", roleDefinitionId)

	roleAssignmentProperties := documentdb.SQLRoleAssignmentCreateUpdateParameters{
		&documentdb.SQLRoleAssignmentResource{
			RoleDefinitionID: utils.String(roleDefinitionId),
			Scope:            utils.String(d.Get("scope").(string)),
			PrincipalID:      utils.String(d.Get("principal_id").(string)),
		},
	}

	future, err := client.CreateUpdateSQLRoleAssignment(ctx, id.SqlRoleAssignmentName, id.ResourceGroup, id.DatabaseAccountName, roleAssignmentProperties)
	if err != nil {
		return fmt.Errorf("creating/updating %q: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for %q to complete creation: %+v", id, err)
	}

	d.SetId(id.ID())

	return resourceCosmosDbRoleAssignmentRead(d, meta)
}

func resourceCosmosDbRoleAssignmentRead(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.SqlClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SqlRoleAssigmentID(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.GetSQLRoleAssignment(ctx, id.SqlRoleAssignmentName, id.ResourceGroup, id.DatabaseAccountName)
	if err != nil {
		return fmt.Errorf("retrieving %q: %+v", id, err)
	}

	d.Set("name", resp.Name)

	if v := resp.SQLRoleAssignmentResource; v != nil {
		d.Set("scope", v.Scope)
		d.Set("principal_id", v.PrincipalID)
		d.Set("role_definition_id", v.RoleDefinitionID)
	}

	return nil
}

func resourceCosmosDbRoleAssignmentDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	client := meta.(*clients.Client).Cosmos.SqlClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.SqlRoleAssigmentID(d.Id())
	if err != nil {
		return err
	}

	future, err := client.DeleteSQLRoleAssignment(ctx, id.SqlRoleAssignmentName, id.ResourceGroup, id.DatabaseAccountName)
	if err != nil {
		return fmt.Errorf("deleting %q: %+v", id, err)
	}

	if err = future.WaitForCompletionRef(ctx, client.Client); err != nil {
		return fmt.Errorf("waiting for %q to delete: %+v", id, err)
	}

	return nil
}

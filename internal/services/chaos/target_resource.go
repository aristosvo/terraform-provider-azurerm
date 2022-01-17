package chaos

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
	"github.com/hashicorp/terraform-provider-azurerm/internal/location"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/chaos/sdk/2021-09-15-preview/targets"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

type TargetResource struct {
}

var _ sdk.Resource = TargetResource{}

type TargetResourceModel struct {
	Name                    string `tfschema:"name"`
	ResourceGroup           string `tfschema:"resource_group_name"`
	ParentResourceType      string `tfschema:"parent_resource_type"`
	ParentName              string `tfschema:"parent_name"`
	ParentProviderNamespace string `tfschema:"parent_provider_namespace"`
	Location                string `tfschema:"location"`
}

func (r TargetResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"resource_group_name": azure.SchemaResourceGroupName(),

		"location": location.Schema(),

		"parent_resource_type": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"parent_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},

		"parent_provider_namespace": {
			Type:     pluginsdk.TypeString,
			Required: true,
			ForceNew: true,
		},
	}
}

func (r TargetResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (r TargetResource) ModelObject() interface{} {
	return &TargetResourceModel{}
}

func (r TargetResource) ResourceType() string {
	return "azurerm_chaos_target"
}

func (r TargetResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return targets.ValidateTargetID
}

func (r TargetResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			var model TargetResourceModel
			if err := metadata.Decode(&model); err != nil {
				return fmt.Errorf("decoding %+v", err)
			}

			client := metadata.Client.Chaos.TargetsClient
			subscriptionId := metadata.Client.Account.SubscriptionId
			id := targets.NewTargetID(subscriptionId, model.ResourceGroup, model.ParentProviderNamespace, model.ParentResourceType, model.ParentName, model.Name)

			existing, err := client.Get(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing Chaos %s: %+v", id, err)
			}

			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			type Indentity struct {
				Type    *string `json:"type,omitempty"`
				Subject *string `json:"subject,omitempty"`
			}

			type TargetProperties struct {
				Identities     *[]Indentity `json:"identities,omitempty"`
				AgentProfileId *string      `json:"agentProfileId,omitempty"`
			}

			target := targets.Target{
				Location:   &model.Location,
				Properties: &TargetProperties{},
			}

			if id.TargetName == "Microsoft-Agent" {
				csi := "CertificateSubjectIssuer"
				subject := "CN=example.subject"

				target.Properties = &TargetProperties{Identities: &[]Indentity{{
					Type:    &csi,
					Subject: &subject,
				}}}
			}

			_, err = client.CreateOrUpdate(ctx, id, target)
			if err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

func (r TargetResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := targets.ParseTargetID(metadata.ResourceData.Id())
			if err != nil {
				return fmt.Errorf("while parsing resource ID: %+v", err)
			}

			client := metadata.Client.Chaos.TargetsClient

			resp, err := client.Get(ctx, *id)
			if err != nil {
				if !response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}
				return fmt.Errorf("while checking for Chaos Target's %q existence: %+v", id.TargetName, err)
			}

			state := TargetResourceModel{
				Name:                    id.TargetName,
				Location:                location.NormalizeNilable(utils.String(*resp.Model.Location)),
				ResourceGroup:           id.ResourceGroupName,
				ParentResourceType:      id.ParentResourceType,
				ParentProviderNamespace: id.ParentProviderNamespace,
				ParentName:              id.ParentResourceName,
			}

			return metadata.Encode(&state)
		},
		Timeout: 5 * time.Minute,
	}
}

func (r TargetResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			id, err := targets.ParseTargetID(metadata.ResourceData.Id())

			if err != nil {
				return fmt.Errorf("while parsing resource ID: %+v", err)
			}

			client := metadata.Client.Chaos.TargetsClient

			_, err = client.Delete(ctx, *id)
			if err != nil {
				return fmt.Errorf("while removing Chaos Target %q: %+v", id.TargetName, err)
			}

			return nil
		},
		Timeout: 30 * time.Minute,
	}
}

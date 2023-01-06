package costmanagement

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/go-azure-helpers/lang/response"
	"github.com/hashicorp/go-azure-sdk/resource-manager/costmanagement/2022-06-01-preview/scheduledactions"
	"github.com/hashicorp/terraform-provider-azurerm/internal/sdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/services/costmanagement/parse"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/validation"
	"github.com/hashicorp/terraform-provider-azurerm/utils"
)

var _ sdk.Resource = AnomalyAlertResource{}

type AnomalyAlertResource struct{}

func (AnomalyAlertResource) Arguments() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"email_subject": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},

		"email_addresses": {
			Type:     pluginsdk.TypeSet,
			Required: true,
			MinItems: 1,
			Elem: &pluginsdk.Schema{
				Type:         pluginsdk.TypeString,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},

		"message": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},
	}
}

func (AnomalyAlertResource) Attributes() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{}
}

func (AnomalyAlertResource) ModelObject() interface{} {
	return nil
}

func (AnomalyAlertResource) ResourceType() string {
	return "azurerm_costmanagement_anomaly_alert"
}

func (r AnomalyAlertResource) Create() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CostManagement.ScheduledActionsClient
			subscriptionId := metadata.Client.Account.SubscriptionId
			id := scheduledactions.NewScopedScheduledActionID(fmt.Sprint("/subscriptions/", subscriptionId), "dailyanomalybyresourcegroup")

			existing, err := client.GetByScope(ctx, id)
			if err != nil && !response.WasNotFound(existing.HttpResponse) {
				return fmt.Errorf("checking for presence of existing %s: %+v", id, err)
			}
			if !response.WasNotFound(existing.HttpResponse) {
				return metadata.ResourceRequiresImport(r.ResourceType(), id)
			}

			emailAddressesRaw := metadata.ResourceData.Get("email_addresses").(*pluginsdk.Set).List()
			emailAddresses := utils.ExpandStringSlice(emailAddressesRaw)
			viewId := parse.NewAnomalyAlertViewIdID(subscriptionId, "ms:DailyAnomalyByResourceGroup")
			schedule := scheduledactions.ScheduleProperties{
				Frequency:  scheduledactions.ScheduleFrequencyDaily,
				HourOfDay:  utils.Int64(int64(12)),
				DayOfMonth: utils.Int64(int64(0)),
			}
			schedule.SetEndDateAsTime(time.Now().AddDate(1, 0, 0))
			schedule.SetStartDateAsTime(time.Now())
			param := scheduledactions.ScheduledAction{
				Kind: utils.ToPtr(scheduledactions.ScheduledActionKindInsightAlert),
				Type: utils.String("Microsoft.CostManagement/ScheduledActions"),
				Properties: &scheduledactions.ScheduledActionProperties{
					DisplayName: metadata.ResourceData.Get("name").(string),
					Status:      scheduledactions.ScheduledActionStatusEnabled,
					ViewId:      viewId.ID(),
					Scope:       utils.String(fmt.Sprint("/subscriptions/", subscriptionId)),
					FileDestination: &scheduledactions.FileDestination{
						FileFormats: &[]scheduledactions.FileFormat{},
					},
					Notification: scheduledactions.NotificationProperties{
						Subject: metadata.ResourceData.Get("email_subject").(string),
						Message: utils.String(metadata.ResourceData.Get("message").(string)),
						To:      *emailAddresses,
					},
					Schedule: schedule,
				},
			}
			if _, err := client.CreateOrUpdateByScope(ctx, id, param); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (r AnomalyAlertResource) Update() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CostManagement.ScheduledActionsClient

			id, err := scheduledactions.ParseScopedScheduledActionID(metadata.ResourceData.Get("id").(string))
			if err != nil {
				return err
			}

			emailAddressesRaw := metadata.ResourceData.Get("email_addresses").(*pluginsdk.Set).List()
			emailAddresses := utils.ExpandStringSlice(emailAddressesRaw)

			subscriptionId := metadata.Client.Account.SubscriptionId
			viewId := parse.NewAnomalyAlertViewIdID(subscriptionId, "ms:DailyAnomalyByResourceGroup")

			schedule := scheduledactions.ScheduleProperties{
				Frequency:  scheduledactions.ScheduleFrequencyDaily,
				HourOfDay:  utils.Int64(int64(12)),
				DayOfMonth: utils.Int64(int64(0)),
			}
			schedule.SetEndDateAsTime(time.Now().AddDate(1, 0, 0))
			schedule.SetStartDateAsTime(time.Now())

			param := scheduledactions.ScheduledAction{
				Kind: utils.ToPtr(scheduledactions.ScheduledActionKindInsightAlert),
				Properties: &scheduledactions.ScheduledActionProperties{
					DisplayName: metadata.ResourceData.Get("name").(string),
					Scope:       utils.String(fmt.Sprint("/subscriptions/", subscriptionId)),
					Status:      scheduledactions.ScheduledActionStatusEnabled,
					ViewId:      viewId.ID(),
					Notification: scheduledactions.NotificationProperties{
						Subject: metadata.ResourceData.Get("email_subject").(string),
						Message: utils.String(metadata.ResourceData.Get("message").(string)),
						To:      *emailAddresses,
					},
					Schedule: schedule,
				},
			}
			if _, err := client.CreateOrUpdateByScope(ctx, *id, param); err != nil {
				return fmt.Errorf("creating %s: %+v", id, err)
			}

			metadata.SetID(id)
			return nil
		},
	}
}

func (AnomalyAlertResource) Read() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 5 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CostManagement.ScheduledActionsClient

			id, err := scheduledactions.ParseScopedScheduledActionID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			resp, err := client.GetByScope(ctx, *id)
			if err != nil {
				if response.WasNotFound(resp.HttpResponse) {
					return metadata.MarkAsGone(id)
				}

				return fmt.Errorf("retrieving %s: %+v", id, err)
			}

			metadata.ResourceData.Set("name", resp.Model.Properties.DisplayName)
			metadata.ResourceData.Set("email_subject", resp.Model.Properties.Notification.Subject)
			metadata.ResourceData.Set("email_addresses", resp.Model.Properties.Notification.To)
			metadata.ResourceData.Set("message", resp.Model.Properties.Notification.Message)

			return nil
		},
	}
}

func (AnomalyAlertResource) Delete() sdk.ResourceFunc {
	return sdk.ResourceFunc{
		Timeout: 30 * time.Minute,
		Func: func(ctx context.Context, metadata sdk.ResourceMetaData) error {
			client := metadata.Client.CostManagement.ScheduledActionsClient

			id, err := scheduledactions.ParseScopedScheduledActionID(metadata.ResourceData.Id())
			if err != nil {
				return err
			}

			_, err = client.DeleteByScope(ctx, *id)
			if err != nil {
				return fmt.Errorf("deleting %s: %+v", *id, err)
			}

			return nil
		},
	}
}

func (AnomalyAlertResource) IDValidationFunc() pluginsdk.SchemaValidateFunc {
	return scheduledactions.ValidateScopedScheduledActionID
}

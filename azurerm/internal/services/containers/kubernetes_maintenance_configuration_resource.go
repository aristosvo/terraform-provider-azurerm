package containers

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2021-03-01/containerservice"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/helpers/tf"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/clients"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/containers/parse"
	containerValidate "github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/services/containers/validate"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/pluginsdk"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/suppress"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/tf/validation"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/internal/timeouts"
	"github.com/terraform-providers/terraform-provider-azurerm/azurerm/utils"
)

func resourceKubernetesMaintenanceConfiguration() *pluginsdk.Resource {
	return &pluginsdk.Resource{
		Create: resourceKubernetesMaintenanceConfigurationCreate,
		Read:   resourceKubernetesMaintenanceConfigurationRead,
		Update: resourceKubernetesMaintenanceConfigurationUpdate,
		Delete: resourceKubernetesMaintenanceConfigurationDelete,

		Importer: pluginsdk.ImporterValidatingResourceId(func(id string) error {
			_, err := parse.MaintenanceConfigurationID(id)
			return err
		}),

		Timeouts: &pluginsdk.ResourceTimeout{
			Create: pluginsdk.DefaultTimeout(60 * time.Minute),
			Read:   pluginsdk.DefaultTimeout(5 * time.Minute),
			Update: pluginsdk.DefaultTimeout(60 * time.Minute),
			Delete: pluginsdk.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*pluginsdk.Schema{
			"name": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},

			"kubernetes_cluster_id": {
				Type:         pluginsdk.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: containerValidate.ClusterID,
			},

			"maintenance_allowed": {
				Type:     pluginsdk.TypeSet,
				Required: true,
				MinItems: 1,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"day": {
							Type:     pluginsdk.TypeString,
							Required: true,
						},
						"hour_slots": {
							Type:     pluginsdk.TypeSet,
							Optional: true,
							Elem: &pluginsdk.Schema{
								Type:         pluginsdk.TypeInt,
								ValidateFunc: validation.IntBetween(0, 24),
							},
						},
					},
				},
				Set: resourceTimeInWeekHash,
			},

			"maintenance_not_allowed_window": {
				Type:     pluginsdk.TypeSet,
				Optional: true,
				Elem: &pluginsdk.Resource{
					Schema: map[string]*pluginsdk.Schema{
						"start": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.IsRFC3339Time,
						},
						"end": {
							Type:             pluginsdk.TypeString,
							Required:         true,
							DiffSuppressFunc: suppress.CaseDifference,
							ValidateFunc:     validation.IsRFC3339Time,
						},
					},
				},
			},
		},
	}
}

func resourceKubernetesMaintenanceConfigurationCreate(d *pluginsdk.ResourceData, meta interface{}) error {
	containersClient := meta.(*clients.Client).Containers
	clustersClient := containersClient.KubernetesClustersClient
	maintenanceConfigurationsClient := containersClient.MaintenanceConfigurationsClient
	ctx, cancel := timeouts.ForCreate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	kubernetesClusterId, err := parse.ClusterID(d.Get("kubernetes_cluster_id").(string))
	if err != nil {
		return err
	}

	resourceGroup := kubernetesClusterId.ResourceGroup
	clusterName := kubernetesClusterId.ManagedClusterName
	name := d.Get("name").(string)

	log.Printf("[DEBUG] Retrieving Kubernetes Cluster %q (Resource Group %q)..", clusterName, resourceGroup)
	cluster, err := clustersClient.Get(ctx, resourceGroup, clusterName)
	if err != nil {
		if utils.ResponseWasNotFound(cluster.Response) {
			return fmt.Errorf("Kubernetes Cluster %q was not found in Resource Group %q!", clusterName, resourceGroup)
		}

		return fmt.Errorf("retrieving existing Kubernetes Cluster %q (Resource Group %q): %+v", clusterName, resourceGroup, err)
	}

	existing, err := maintenanceConfigurationsClient.Get(ctx, resourceGroup, clusterName, name)
	if err != nil {
		if !utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("checking for presence of existing Maintenance Configuration %q (Kubernetes Cluster %q / Resource Group %q): %s", name, clusterName, resourceGroup, err)
		}
	}

	if existing.ID != nil && *existing.ID != "" {
		return tf.ImportAsExistsError("azurerm_kubernetes_maintenance_configuration", *existing.ID)
	}

	props := containerservice.MaintenanceConfiguration{}

	if v, set := d.GetOk("maintenance_allowed"); set && v.(*pluginsdk.Set) != nil {
		props.TimeInWeek = expandTimeInWeek(v.(*pluginsdk.Set))
	}

	if v, set := d.GetOk("maintenance_not_allowed_window"); set {
		props.NotAllowedTime = expandNotAllowedTime(v.(*pluginsdk.Set))
	}

	read, err := maintenanceConfigurationsClient.CreateOrUpdate(ctx, resourceGroup, clusterName, name, props)
	if err != nil {
		return fmt.Errorf("creating/updating Managed Kubernetes Maintenance Configuration %q (Resource Group %q): %+v", name, resourceGroup, err)
	}

	if read.ID == nil {
		return fmt.Errorf("cannot read ID for Managed Kubernetes Maintenance Configuration %q (Resource Group %q)", name, resourceGroup)
	}

	d.SetId(*read.ID)

	return resourceKubernetesMaintenanceConfigurationRead(d, meta)
}

func resourceKubernetesMaintenanceConfigurationUpdate(d *pluginsdk.ResourceData, meta interface{}) error {
	containersClient := meta.(*clients.Client).Containers
	client := containersClient.MaintenanceConfigurationsClient
	ctx, cancel := timeouts.ForUpdate(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MaintenanceConfigurationID(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Retrieving existing Maintenance Configuration %q (Kubernetes Cluster %q / Resource Group %q)..", id.Name, id.ManagedClusterName, id.ResourceGroup)
	existing, err := client.Get(ctx, id.ResourceGroup, id.ManagedClusterName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(existing.Response) {
			return fmt.Errorf("Maintenance Configuration %q was not found in Managed Kubernetes Cluster %q / Resource Group %q", id.Name, id.ManagedClusterName, id.ResourceGroup)
		}

		return fmt.Errorf("retrieving Maintenance Configuration %q (Managed Kubernetes Cluster %q / Resource Group %q): %+v", id.Name, id.ManagedClusterName, id.ResourceGroup, err)
	}

	props := existing.MaintenanceConfigurationProperties
	log.Printf("[DEBUG] Determining delta for existing Maintenance Configuration %q (Kubernetes Cluster %q / Resource Group %q)..", id.Name, id.ManagedClusterName, id.ResourceGroup)

	// delta patching
	if d.HasChange("maintenance_allowed") {
		props.TimeInWeek = expandTimeInWeek(d.Get("maintenance_allowed").(*pluginsdk.Set))
	}

	if d.HasChange("maintenance_not_allowed_window") {
		props.NotAllowedTime = expandNotAllowedTime(d.Get("maintenance_not_allowed_window").(*pluginsdk.Set))
	}

	log.Printf("[DEBUG] Updating existing Maintenance Configuration %q (Kubernetes Cluster %q / Resource Group %q)..", id.Name, id.ManagedClusterName, id.ResourceGroup)
	existing.MaintenanceConfigurationProperties = props
	_, err = client.CreateOrUpdate(ctx, id.ResourceGroup, id.ManagedClusterName, id.Name, existing)
	if err != nil {
		return fmt.Errorf("updating Maintenance Configuration %q (Kubernetes Cluster %q / Resource Group %q): %+v", id.Name, id.ManagedClusterName, id.ResourceGroup, err)
	}

	return resourceKubernetesMaintenanceConfigurationRead(d, meta)
}

func resourceKubernetesMaintenanceConfigurationRead(d *pluginsdk.ResourceData, meta interface{}) error {
	containersClient := meta.(*clients.Client).Containers
	clustersClient := containersClient.KubernetesClustersClient
	client := containersClient.MaintenanceConfigurationsClient
	ctx, cancel := timeouts.ForRead(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MaintenanceConfigurationID(d.Id())
	if err != nil {
		return err
	}

	// if the parent cluster doesn't exist then the managed configuration won't
	cluster, err := clustersClient.Get(ctx, id.ResourceGroup, id.ManagedClusterName)
	if err != nil {
		if utils.ResponseWasNotFound(cluster.Response) {
			log.Printf("[DEBUG] Managed Kubernetes Cluster %q was not found in Resource Group %q - removing from state!", id.ManagedClusterName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Managed Kubernetes Cluster %q (Resource Group %q): %+v", id.ManagedClusterName, id.ResourceGroup, err)
	}

	resp, err := client.Get(ctx, id.ResourceGroup, id.ManagedClusterName, id.Name)
	if err != nil {
		if utils.ResponseWasNotFound(resp.Response) {
			log.Printf("[DEBUG] Maintenance Configuration %q was not found in Managed Kubernetes Cluster %q / Resource Group %q - removing from state!", id.Name, id.ManagedClusterName, id.ResourceGroup)
			d.SetId("")
			return nil
		}

		return fmt.Errorf("retrieving Maintenance Configuration %q (Managed Kubernetes Cluster %q / Resource Group %q): %+v", id.Name, id.ManagedClusterName, id.ResourceGroup, err)
	}

	d.Set("name", id.Name)
	d.Set("kubernetes_cluster_id", cluster.ID)

	if props := resp.MaintenanceConfigurationProperties; props != nil {
		if err := d.Set("maintenace_allowed", flattenTimeInWeek(props.TimeInWeek)); err != nil {
			return fmt.Errorf("setting `maintenance_allowed`: %+v", err)
		}

		if err := d.Set("maintenace_not_allowed_window", flattenNotAllowedTime(props.NotAllowedTime)); err != nil {
			return fmt.Errorf("setting `maintenace_not_allowed_window`: %+v", err)
		}
	}

	return nil
}

func resourceKubernetesMaintenanceConfigurationDelete(d *pluginsdk.ResourceData, meta interface{}) error {
	containersClient := meta.(*clients.Client).Containers
	client := containersClient.MaintenanceConfigurationsClient
	ctx, cancel := timeouts.ForDelete(meta.(*clients.Client).StopContext, d)
	defer cancel()

	id, err := parse.MaintenanceConfigurationID(d.Id())
	if err != nil {
		return err
	}

	_, err = client.Delete(ctx, id.ResourceGroup, id.ManagedClusterName, id.Name)
	if err != nil {
		return fmt.Errorf("deleting Maintenance Configuration %q (Managed Kubernetes Cluster %q / Resource Group %q): %+v", id.Name, id.ManagedClusterName, id.ResourceGroup, err)
	}

	return nil
}

func expandTimeInWeek(input *pluginsdk.Set) *[]containerservice.TimeInWeek {
	timeInWeeks := make([]containerservice.TimeInWeek, 0)
	for _, timeInWeekInput := range input.List() {
		timeInWeekInput := timeInWeekInput.(map[string]interface{})

		slots := make([]int32, 0)
		for _, v := range timeInWeekInput["hour_slots"].(*pluginsdk.Set).List() {
			slots = append(slots, int32(v.(int)))
		}

		timeInWeek := &containerservice.TimeInWeek{
			Day:       containerservice.WeekDayFriday,
			HourSlots: &slots,
		}

		timeInWeeks = append(timeInWeeks, *timeInWeek)
	}
	return &timeInWeeks
}

func expandNotAllowedTime(input *pluginsdk.Set) *[]containerservice.TimeSpan {
	timeSpans := make([]containerservice.TimeSpan, 0)
	for _, timeSpanInput := range input.List() {
		timeSpanInput := timeSpanInput.(map[string]interface{})

		startTime, _ := time.Parse(time.RFC3339, timeSpanInput["start"].(string)) // should be validated by the schema
		endTime, _ := time.Parse(time.RFC3339, timeSpanInput["end"].(string))     // should be validated by the schema

		timeSpan := containerservice.TimeSpan{
			Start: &date.Time{Time: startTime},
			End:   &date.Time{Time: endTime},
		}
		timeSpans = append(timeSpans, timeSpan)

	}
	return &timeSpans
}

func flattenTimeInWeek(input *[]containerservice.TimeInWeek) *pluginsdk.Set {
	timeInWeekElements := &pluginsdk.Set{F: resourceTimeInWeekHash}
	if input == nil {
		return timeInWeekElements
	}

	for _, element := range *input {
		output := map[string]interface{}{}
		if element.Day == "" || len(*element.HourSlots) == 0 {
			continue
		}

		output["day"] = element.Day
		output["slots"] = element.HourSlots

		timeInWeekElements.Add(output)
	}

	return timeInWeekElements
}

func resourceTimeInWeekHash(v interface{}) int {
	var buf bytes.Buffer

	if m, ok := v.(map[string]interface{}); ok {
		buf.WriteString(m["day"].(string))
	}

	return pluginsdk.HashString(buf.String())
}

func flattenNotAllowedTime(input *[]containerservice.TimeSpan) *pluginsdk.Set {
	timeSpanElements := &pluginsdk.Set{F: pluginsdk.HashString}
	if input == nil {
		return timeSpanElements
	}

	for _, element := range *input {
		output := map[string]interface{}{}
		if element.Start.String() == "" || element.End.String() == "" {
			continue
		}

		output["start"] = element.Start.String()
		output["end"] = element.End.String()

		timeSpanElements.Add(output)
	}
	return timeSpanElements
}

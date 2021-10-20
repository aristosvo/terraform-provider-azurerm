package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"testing"

	"github.com/hashicorp/terraform-provider-azurerm/internal/resourceid"
)

var _ resourceid.Formatter = EndpointServiceBusQueueId{}

func TestEndpointServiceBusQueueIDFormatter(t *testing.T) {
	actual := NewEndpointServiceBusQueueID("12345678-1234-9876-4563-123456789012", "resGroup1", "hub1", "serviceBusQueueEndpoint1").ID()
	expected := "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/IotHubs/hub1/Endpoints/serviceBusQueueEndpoint1"
	if actual != expected {
		t.Fatalf("Expected %q but got %q", expected, actual)
	}
}

func TestEndpointServiceBusQueueID(t *testing.T) {
	testData := []struct {
		Input    string
		Error    bool
		Expected *EndpointServiceBusQueueId
	}{

		{
			// empty
			Input: "",
			Error: true,
		},

		{
			// missing SubscriptionId
			Input: "/",
			Error: true,
		},

		{
			// missing value for SubscriptionId
			Input: "/subscriptions/",
			Error: true,
		},

		{
			// missing ResourceGroup
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/",
			Error: true,
		},

		{
			// missing value for ResourceGroup
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/",
			Error: true,
		},

		{
			// missing IotHubName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/",
			Error: true,
		},

		{
			// missing value for IotHubName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/IotHubs/",
			Error: true,
		},

		{
			// missing EndpointName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/IotHubs/hub1/",
			Error: true,
		},

		{
			// missing value for EndpointName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/IotHubs/hub1/Endpoints/",
			Error: true,
		},

		{
			// valid
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/resourceGroups/resGroup1/providers/Microsoft.Devices/IotHubs/hub1/Endpoints/serviceBusQueueEndpoint1",
			Expected: &EndpointServiceBusQueueId{
				SubscriptionId: "12345678-1234-9876-4563-123456789012",
				ResourceGroup:  "resGroup1",
				IotHubName:     "hub1",
				EndpointName:   "serviceBusQueueEndpoint1",
			},
		},

		{
			// upper-cased
			Input: "/SUBSCRIPTIONS/12345678-1234-9876-4563-123456789012/RESOURCEGROUPS/RESGROUP1/PROVIDERS/MICROSOFT.DEVICES/IOTHUBS/HUB1/ENDPOINTS/SERVICEBUSQUEUEENDPOINT1",
			Error: true,
		},
	}

	for _, v := range testData {
		t.Logf("[DEBUG] Testing %q", v.Input)

		actual, err := EndpointServiceBusQueueID(v.Input)
		if err != nil {
			if v.Error {
				continue
			}

			t.Fatalf("Expect a value but got an error: %s", err)
		}
		if v.Error {
			t.Fatal("Expect an error but didn't get one")
		}

		if actual.SubscriptionId != v.Expected.SubscriptionId {
			t.Fatalf("Expected %q but got %q for SubscriptionId", v.Expected.SubscriptionId, actual.SubscriptionId)
		}
		if actual.ResourceGroup != v.Expected.ResourceGroup {
			t.Fatalf("Expected %q but got %q for ResourceGroup", v.Expected.ResourceGroup, actual.ResourceGroup)
		}
		if actual.IotHubName != v.Expected.IotHubName {
			t.Fatalf("Expected %q but got %q for IotHubName", v.Expected.IotHubName, actual.IotHubName)
		}
		if actual.EndpointName != v.Expected.EndpointName {
			t.Fatalf("Expected %q but got %q for EndpointName", v.Expected.EndpointName, actual.EndpointName)
		}
	}
}

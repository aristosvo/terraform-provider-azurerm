package validate

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import "testing"

func TestSubscriptionCostManagementScheduledActionID(t *testing.T) {
	cases := []struct {
		Input string
		Valid bool
	}{

		{
			// empty
			Input: "",
			Valid: false,
		},

		{
			// missing SubscriptionId
			Input: "/",
			Valid: false,
		},

		{
			// missing value for SubscriptionId
			Input: "/subscriptions/",
			Valid: false,
		},

		{
			// missing ScheduledActionName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/providers/Microsoft.CostManagement/",
			Valid: false,
		},

		{
			// missing value for ScheduledActionName
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/providers/Microsoft.CostManagement/scheduledActions/",
			Valid: false,
		},

		{
			// valid
			Input: "/subscriptions/12345678-1234-9876-4563-123456789012/providers/Microsoft.CostManagement/scheduledActions/scheduledaction1",
			Valid: true,
		},

		{
			// upper-cased
			Input: "/SUBSCRIPTIONS/12345678-1234-9876-4563-123456789012/PROVIDERS/MICROSOFT.COSTMANAGEMENT/SCHEDULEDACTIONS/SCHEDULEDACTION1",
			Valid: false,
		},
	}
	for _, tc := range cases {
		t.Logf("[DEBUG] Testing Value %s", tc.Input)
		_, errors := SubscriptionCostManagementScheduledActionID(tc.Input, "test")
		valid := len(errors) == 0

		if tc.Valid != valid {
			t.Fatalf("Expected %t but got %t", tc.Valid, valid)
		}
	}
}

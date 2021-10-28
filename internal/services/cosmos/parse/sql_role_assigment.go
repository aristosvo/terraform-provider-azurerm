package parse

// NOTE: this file is generated via 'go:generate' - manual changes will be overwritten

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-provider-azurerm/helpers/azure"
)

type SqlRoleAssigmentId struct {
	SubscriptionId        string
	ResourceGroup         string
	DatabaseAccountName   string
	SqlRoleAssignmentName string
}

func NewSqlRoleAssigmentID(subscriptionId, resourceGroup, databaseAccountName, sqlRoleAssignmentName string) SqlRoleAssigmentId {
	return SqlRoleAssigmentId{
		SubscriptionId:        subscriptionId,
		ResourceGroup:         resourceGroup,
		DatabaseAccountName:   databaseAccountName,
		SqlRoleAssignmentName: sqlRoleAssignmentName,
	}
}

func (id SqlRoleAssigmentId) String() string {
	segments := []string{
		fmt.Sprintf("Sql Role Assignment Name %q", id.SqlRoleAssignmentName),
		fmt.Sprintf("Database Account Name %q", id.DatabaseAccountName),
		fmt.Sprintf("Resource Group %q", id.ResourceGroup),
	}
	segmentsStr := strings.Join(segments, " / ")
	return fmt.Sprintf("%s: (%s)", "Sql Role Assigment", segmentsStr)
}

func (id SqlRoleAssigmentId) ID() string {
	fmtString := "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DocumentDB/databaseAccounts/%s/sqlRoleAssignments/%s"
	return fmt.Sprintf(fmtString, id.SubscriptionId, id.ResourceGroup, id.DatabaseAccountName, id.SqlRoleAssignmentName)
}

// SqlRoleAssigmentID parses a SqlRoleAssigment ID into an SqlRoleAssigmentId struct
func SqlRoleAssigmentID(input string) (*SqlRoleAssigmentId, error) {
	id, err := azure.ParseAzureResourceID(input)
	if err != nil {
		return nil, err
	}

	resourceId := SqlRoleAssigmentId{
		SubscriptionId: id.SubscriptionID,
		ResourceGroup:  id.ResourceGroup,
	}

	if resourceId.SubscriptionId == "" {
		return nil, fmt.Errorf("ID was missing the 'subscriptions' element")
	}

	if resourceId.ResourceGroup == "" {
		return nil, fmt.Errorf("ID was missing the 'resourceGroups' element")
	}

	if resourceId.DatabaseAccountName, err = id.PopSegment("databaseAccounts"); err != nil {
		return nil, err
	}
	if resourceId.SqlRoleAssignmentName, err = id.PopSegment("sqlRoleAssignments"); err != nil {
		return nil, err
	}

	if err := id.ValidateNoEmptySegments(input); err != nil {
		return nil, err
	}

	return &resourceId, nil
}

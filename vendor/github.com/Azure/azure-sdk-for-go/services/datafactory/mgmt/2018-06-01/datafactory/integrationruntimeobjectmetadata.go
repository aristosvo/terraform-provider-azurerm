package datafactory

// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Code generated by Microsoft (R) AutoRest Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"context"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/Azure/go-autorest/tracing"
	"net/http"
)

// IntegrationRuntimeObjectMetadataClient is the the Azure Data Factory V2 management API provides a RESTful set of web
// services that interact with Azure Data Factory V2 services.
type IntegrationRuntimeObjectMetadataClient struct {
	BaseClient
}

// NewIntegrationRuntimeObjectMetadataClient creates an instance of the IntegrationRuntimeObjectMetadataClient client.
func NewIntegrationRuntimeObjectMetadataClient(subscriptionID string) IntegrationRuntimeObjectMetadataClient {
	return NewIntegrationRuntimeObjectMetadataClientWithBaseURI(DefaultBaseURI, subscriptionID)
}

// NewIntegrationRuntimeObjectMetadataClientWithBaseURI creates an instance of the
// IntegrationRuntimeObjectMetadataClient client using a custom endpoint.  Use this when interacting with an Azure
// cloud that uses a non-standard base URI (sovereign clouds, Azure stack).
func NewIntegrationRuntimeObjectMetadataClientWithBaseURI(baseURI string, subscriptionID string) IntegrationRuntimeObjectMetadataClient {
	return IntegrationRuntimeObjectMetadataClient{NewWithBaseURI(baseURI, subscriptionID)}
}

// Get get a SSIS integration runtime object metadata by specified path. The return is pageable metadata list.
// Parameters:
// resourceGroupName - the resource group name.
// factoryName - the factory name.
// integrationRuntimeName - the integration runtime name.
// getMetadataRequest - the parameters for getting a SSIS object metadata.
func (client IntegrationRuntimeObjectMetadataClient) Get(ctx context.Context, resourceGroupName string, factoryName string, integrationRuntimeName string, getMetadataRequest *GetSsisObjectMetadataRequest) (result SsisObjectMetadataListResponse, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/IntegrationRuntimeObjectMetadataClient.Get")
		defer func() {
			sc := -1
			if result.Response.Response != nil {
				sc = result.Response.Response.StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	if err := validation.Validate([]validation.Validation{
		{TargetValue: resourceGroupName,
			Constraints: []validation.Constraint{{Target: "resourceGroupName", Name: validation.MaxLength, Rule: 90, Chain: nil},
				{Target: "resourceGroupName", Name: validation.MinLength, Rule: 1, Chain: nil},
				{Target: "resourceGroupName", Name: validation.Pattern, Rule: `^[-\w\._\(\)]+$`, Chain: nil}}},
		{TargetValue: factoryName,
			Constraints: []validation.Constraint{{Target: "factoryName", Name: validation.MaxLength, Rule: 63, Chain: nil},
				{Target: "factoryName", Name: validation.MinLength, Rule: 3, Chain: nil},
				{Target: "factoryName", Name: validation.Pattern, Rule: `^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`, Chain: nil}}},
		{TargetValue: integrationRuntimeName,
			Constraints: []validation.Constraint{{Target: "integrationRuntimeName", Name: validation.MaxLength, Rule: 63, Chain: nil},
				{Target: "integrationRuntimeName", Name: validation.MinLength, Rule: 3, Chain: nil},
				{Target: "integrationRuntimeName", Name: validation.Pattern, Rule: `^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`, Chain: nil}}}}); err != nil {
		return result, validation.NewError("datafactory.IntegrationRuntimeObjectMetadataClient", "Get", err.Error())
	}

	req, err := client.GetPreparer(ctx, resourceGroupName, factoryName, integrationRuntimeName, getMetadataRequest)
	if err != nil {
		err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataClient", "Get", nil, "Failure preparing request")
		return
	}

	resp, err := client.GetSender(req)
	if err != nil {
		result.Response = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataClient", "Get", resp, "Failure sending request")
		return
	}

	result, err = client.GetResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataClient", "Get", resp, "Failure responding to request")
		return
	}

	return
}

// GetPreparer prepares the Get request.
func (client IntegrationRuntimeObjectMetadataClient) GetPreparer(ctx context.Context, resourceGroupName string, factoryName string, integrationRuntimeName string, getMetadataRequest *GetSsisObjectMetadataRequest) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"factoryName":            autorest.Encode("path", factoryName),
		"integrationRuntimeName": autorest.Encode("path", integrationRuntimeName),
		"resourceGroupName":      autorest.Encode("path", resourceGroupName),
		"subscriptionId":         autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2018-06-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/json; charset=utf-8"),
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DataFactory/factories/{factoryName}/integrationRuntimes/{integrationRuntimeName}/getObjectMetadata", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	if getMetadataRequest != nil {
		preparer = autorest.DecoratePreparer(preparer,
			autorest.WithJSON(getMetadataRequest))
	}
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// GetSender sends the Get request. The method will close the
// http.Response Body if it receives an error.
func (client IntegrationRuntimeObjectMetadataClient) GetSender(req *http.Request) (*http.Response, error) {
	return client.Send(req, azure.DoRetryWithRegistration(client.Client))
}

// GetResponder handles the response to the Get request. The method always
// closes the http.Response Body.
func (client IntegrationRuntimeObjectMetadataClient) GetResponder(resp *http.Response) (result SsisObjectMetadataListResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

// Refresh refresh a SSIS integration runtime object metadata.
// Parameters:
// resourceGroupName - the resource group name.
// factoryName - the factory name.
// integrationRuntimeName - the integration runtime name.
func (client IntegrationRuntimeObjectMetadataClient) Refresh(ctx context.Context, resourceGroupName string, factoryName string, integrationRuntimeName string) (result IntegrationRuntimeObjectMetadataRefreshFuture, err error) {
	if tracing.IsEnabled() {
		ctx = tracing.StartSpan(ctx, fqdn+"/IntegrationRuntimeObjectMetadataClient.Refresh")
		defer func() {
			sc := -1
			if result.FutureAPI != nil && result.FutureAPI.Response() != nil {
				sc = result.FutureAPI.Response().StatusCode
			}
			tracing.EndSpan(ctx, sc, err)
		}()
	}
	if err := validation.Validate([]validation.Validation{
		{TargetValue: resourceGroupName,
			Constraints: []validation.Constraint{{Target: "resourceGroupName", Name: validation.MaxLength, Rule: 90, Chain: nil},
				{Target: "resourceGroupName", Name: validation.MinLength, Rule: 1, Chain: nil},
				{Target: "resourceGroupName", Name: validation.Pattern, Rule: `^[-\w\._\(\)]+$`, Chain: nil}}},
		{TargetValue: factoryName,
			Constraints: []validation.Constraint{{Target: "factoryName", Name: validation.MaxLength, Rule: 63, Chain: nil},
				{Target: "factoryName", Name: validation.MinLength, Rule: 3, Chain: nil},
				{Target: "factoryName", Name: validation.Pattern, Rule: `^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`, Chain: nil}}},
		{TargetValue: integrationRuntimeName,
			Constraints: []validation.Constraint{{Target: "integrationRuntimeName", Name: validation.MaxLength, Rule: 63, Chain: nil},
				{Target: "integrationRuntimeName", Name: validation.MinLength, Rule: 3, Chain: nil},
				{Target: "integrationRuntimeName", Name: validation.Pattern, Rule: `^[A-Za-z0-9]+(?:-[A-Za-z0-9]+)*$`, Chain: nil}}}}); err != nil {
		return result, validation.NewError("datafactory.IntegrationRuntimeObjectMetadataClient", "Refresh", err.Error())
	}

	req, err := client.RefreshPreparer(ctx, resourceGroupName, factoryName, integrationRuntimeName)
	if err != nil {
		err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataClient", "Refresh", nil, "Failure preparing request")
		return
	}

	result, err = client.RefreshSender(req)
	if err != nil {
		err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataClient", "Refresh", nil, "Failure sending request")
		return
	}

	return
}

// RefreshPreparer prepares the Refresh request.
func (client IntegrationRuntimeObjectMetadataClient) RefreshPreparer(ctx context.Context, resourceGroupName string, factoryName string, integrationRuntimeName string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"factoryName":            autorest.Encode("path", factoryName),
		"integrationRuntimeName": autorest.Encode("path", integrationRuntimeName),
		"resourceGroupName":      autorest.Encode("path", resourceGroupName),
		"subscriptionId":         autorest.Encode("path", client.SubscriptionID),
	}

	const APIVersion = "2018-06-01"
	queryParameters := map[string]interface{}{
		"api-version": APIVersion,
	}

	preparer := autorest.CreatePreparer(
		autorest.AsPost(),
		autorest.WithBaseURL(client.BaseURI),
		autorest.WithPathParameters("/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.DataFactory/factories/{factoryName}/integrationRuntimes/{integrationRuntimeName}/refreshObjectMetadata", pathParameters),
		autorest.WithQueryParameters(queryParameters))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// RefreshSender sends the Refresh request. The method will close the
// http.Response Body if it receives an error.
func (client IntegrationRuntimeObjectMetadataClient) RefreshSender(req *http.Request) (future IntegrationRuntimeObjectMetadataRefreshFuture, err error) {
	var resp *http.Response
	resp, err = client.Send(req, azure.DoRetryWithRegistration(client.Client))
	if err != nil {
		return
	}
	var azf azure.Future
	azf, err = azure.NewFutureFromResponse(resp)
	future.FutureAPI = &azf
	future.Result = func(client IntegrationRuntimeObjectMetadataClient) (somsr SsisObjectMetadataStatusResponse, err error) {
		var done bool
		done, err = future.DoneWithContext(context.Background(), client)
		if err != nil {
			err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataRefreshFuture", "Result", future.Response(), "Polling failure")
			return
		}
		if !done {
			err = azure.NewAsyncOpIncompleteError("datafactory.IntegrationRuntimeObjectMetadataRefreshFuture")
			return
		}
		sender := autorest.DecorateSender(client, autorest.DoRetryForStatusCodes(client.RetryAttempts, client.RetryDuration, autorest.StatusCodesForRetry...))
		somsr.Response.Response, err = future.GetResult(sender)
		if somsr.Response.Response == nil && err == nil {
			err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataRefreshFuture", "Result", nil, "received nil response and error")
		}
		if err == nil && somsr.Response.Response.StatusCode != http.StatusNoContent {
			somsr, err = client.RefreshResponder(somsr.Response.Response)
			if err != nil {
				err = autorest.NewErrorWithError(err, "datafactory.IntegrationRuntimeObjectMetadataRefreshFuture", "Result", somsr.Response.Response, "Failure responding to request")
			}
		}
		return
	}
	return
}

// RefreshResponder handles the response to the Refresh request. The method always
// closes the http.Response Body.
func (client IntegrationRuntimeObjectMetadataClient) RefreshResponder(resp *http.Response) (result SsisObjectMetadataStatusResponse, err error) {
	err = autorest.Respond(
		resp,
		azure.WithErrorUnlessStatusCode(http.StatusOK, http.StatusAccepted),
		autorest.ByUnmarshallingJSON(&result),
		autorest.ByClosing())
	result.Response = autorest.Response{Response: resp}
	return
}

package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/azure/azure-dev/cli/azd/pkg/convert"
	"github.com/azure/azure-dev/cli/azd/pkg/infra"
	"github.com/azure/azure-dev/cli/azd/pkg/input"
	"github.com/azure/azure-dev/cli/azd/test/mocks"
	"github.com/stretchr/testify/require"
)

func Test_promptEnvironmentName(t *testing.T) {
	t.Run("valid name", func(t *testing.T) {
		mockContext := mocks.NewMockContext(context.Background())
		mockContext.Console.WhenPrompt(func(options input.ConsoleOptions) bool {
			return true
		}).SetError(errors.New("prompt should not be called for valid environment name"))

		environmentName := "hello"

		err := ensureValidEnvironmentName(*mockContext.Context, &environmentName, mockContext.Console)

		require.NoError(t, err)
	})

	t.Run("empty name gets prompted", func(t *testing.T) {
		environmentName := ""

		mockContext := mocks.NewMockContext(context.Background())
		mockContext.Console.WhenPrompt(func(options input.ConsoleOptions) bool {
			return true
		}).Respond("someEnv")

		err := ensureValidEnvironmentName(*mockContext.Context, &environmentName, mockContext.Console)

		require.NoError(t, err)
		require.Equal(t, "someEnv", environmentName)
	})

	t.Run("duplicate resource groups ignored", func(t *testing.T) {
		mockDeployment := armresources.DeploymentExtended{
			Properties: &armresources.DeploymentPropertiesExtended{
				Dependencies: []*armresources.Dependency{
					{
						DependsOn: []*armresources.BasicDependency{
							{
								ResourceName: convert.RefOf("groupA"),
								ResourceType: convert.RefOf(string(infra.AzureResourceTypeResourceGroup)),
							},
							{
								ResourceName: convert.RefOf("groupB"),
								ResourceType: convert.RefOf(string(infra.AzureResourceTypeResourceGroup)),
							},
							{
								ResourceName: convert.RefOf("ignoredForWrongType"),
								ResourceType: convert.RefOf(string(infra.AzureResourceTypeStorageAccount)),
							},
						},
					},
					{
						DependsOn: []*armresources.BasicDependency{
							{
								ResourceName: convert.RefOf("groupA"),
								ResourceType: convert.RefOf(string(infra.AzureResourceTypeResourceGroup)),
							},
							{
								ResourceName: convert.RefOf("groupB"),
								ResourceType: convert.RefOf(string(infra.AzureResourceTypeResourceGroup)),
							},
							{
								ResourceName: convert.RefOf("groupC"),
								ResourceType: convert.RefOf(string(infra.AzureResourceTypeResourceGroup)),
							},
						},
					},
				},
			},
		}

		mockContext := mocks.NewMockContext(context.Background())

		mockContext.HttpClient.When(func(request *http.Request) bool {
			return request.Method == http.MethodGet && strings.Contains(
				request.URL.Path,
				"/subscriptions/sub-id/providers/Microsoft.Resources/deployments",
			)
		}).RespondFn(func(request *http.Request) (*http.Response, error) {
			subscriptionsListBytes, _ := json.Marshal(mockDeployment)

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBuffer(subscriptionsListBytes)),
			}, nil
		})

		resourceManager := infra.NewAzureResourceManager(*mockContext.Context)
		groups, err := resourceManager.GetResourceGroupsForDeployment(*mockContext.Context, "sub-id", "deployment-name")
		require.NoError(t, err)

		sort.Strings(groups)
		require.Equal(t, []string{"groupA", "groupB", "groupC"}, groups)
	})
}

func Test_getSubscriptionOptions(t *testing.T) {
	t.Run("no default", func(t *testing.T) {
		mockContext := mocks.NewMockContext(context.Background())
		mockContext.HttpClient.When(func(request *http.Request) bool {
			return strings.Contains(request.URL.Path, "subscriptions/")
		}).RespondFn(func(request *http.Request) (*http.Response, error) {
			return mocks.CreateHttpResponseWithBody(request, 200, armsubscriptions.Subscription{
				ID:             convert.RefOf("SUBSCRIPTION"),
				SubscriptionID: convert.RefOf("SUBSCRIPTION_ID"),
				DisplayName:    convert.RefOf("DISPLAY"),
				TenantID:       convert.RefOf("TENANT"),
			})
		})
		mockContext.HttpClient.When(func(request *http.Request) bool {
			return request.URL.Path == "/subscriptions"
		}).RespondFn(func(request *http.Request) (*http.Response, error) {
			return mocks.CreateHttpResponseWithBody(request, 200, armsubscriptions.ClientListResponse{
				SubscriptionListResult: armsubscriptions.SubscriptionListResult{
					Value: []*armsubscriptions.Subscription{
						{
							ID:             convert.RefOf("SUBSCRIPTION"),
							SubscriptionID: convert.RefOf("SUBSCRIPTION_ID_OTHER"),
							DisplayName:    convert.RefOf("DISPLAY"),
							TenantID:       convert.RefOf("TENANT"),
						},
					},
				},
			})
		})
		subList, result, err := getSubscriptionOptions(*mockContext.Context)

		require.Nil(t, err)
		require.EqualValues(t, 2, len(subList))
		require.EqualValues(t, nil, result)
	})

	t.Run("default", func(t *testing.T) {
		mockContext := mocks.NewMockContext(context.Background())
		mockContext.HttpClient.When(func(request *http.Request) bool {
			return strings.Contains(request.URL.Path, "subscriptions/")
		}).RespondFn(func(request *http.Request) (*http.Response, error) {
			return mocks.CreateHttpResponseWithBody(request, 200, armsubscriptions.Subscription{
				ID:             convert.RefOf("SUBSCRIPTION"),
				SubscriptionID: convert.RefOf("SUBSCRIPTION_ID"),
				DisplayName:    convert.RefOf("DISPLAY"),
				TenantID:       convert.RefOf("TENANT"),
			})
		})
		mockContext.HttpClient.When(func(request *http.Request) bool {
			return request.URL.Path == "/subscriptions"
		}).RespondFn(func(request *http.Request) (*http.Response, error) {
			return mocks.CreateHttpResponseWithBody(request, 200, armsubscriptions.ClientListResponse{
				SubscriptionListResult: armsubscriptions.SubscriptionListResult{
					Value: []*armsubscriptions.Subscription{
						{
							ID:             convert.RefOf("SUBSCRIPTION"),
							SubscriptionID: convert.RefOf("SUBSCRIPTION_ID"),
							DisplayName:    convert.RefOf("DISPLAY"),
							TenantID:       convert.RefOf("TENANT"),
						},
					},
				},
			})
		})
		subList, result, err := getSubscriptionOptions(*mockContext.Context)

		require.Nil(t, err)
		require.EqualValues(t, 2, len(subList))
		require.NotNil(t, result)
		defSub, ok := result.(string)
		require.True(t, ok)
		require.EqualValues(t, " 1. DISPLAY (SUBSCRIPTION_ID)", defSub)
	})
}

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package ux

import (
	"fmt"
	"strings"

	"github.com/azure/azure-dev/cli/azd/pkg/output"
	"github.com/fatih/color"
)

type ShowService struct {
	Name      string
	IngresUrl string
}

type ShowEnvironment struct {
	Name      string
	IsCurrent bool
	IsRemote  bool
}

type Show struct {
	AppName         string
	Services        []*ShowService
	Environments    []*ShowEnvironment
	AzurePortalLink string
}

const (
	cHeader           = "\nShowing deployed endpoints and environments for apps in this directory.\n"
	cHeaderNote       = "To view endpoints for a different environment run "
	cShowDifferentEnv = "azd show -e <environment name>"
	cServices         = "\n  Services:\n"
	cEnvironments     = "\n  Environments:\n"
	cCurrentEnv       = " [Current]"
	cRemoteEnv        = " (Remote)"
	cViewInPortal     = "\n  View in Azure Portal:\n"
)

func (s *Show) ToString(currentIndentation string) string {
	return fmt.Sprintf(
		"%s%s%s%s%s%s%s%s%s    %s\n",
		cHeader,
		cHeaderNote,
		color.BlueString("%s\n\n", cShowDifferentEnv),
		color.MagentaString(s.AppName),
		cEnvironments,
		environments(s.Environments),
		cServices,
		services(s.Services),
		cViewInPortal,
		azurePortalLink(s.AzurePortalLink),
	)
}

func azurePortalLink(link string) string {
	if link == "" {
		return fmt.Sprintf(
			"Application is not yet provisioned. Run %s or %s first.",
			color.BlueString("azd provision"),
			color.BlueString("azd up"),
		)
	}
	return output.WithLinkFormat(link)
}

func services(services []*ShowService) string {
	servicesCount := len(services)
	if servicesCount == 0 {
		return fmt.Sprintf(
			"    You don't have services defined. Add your services to %s.",
			color.BlueString("azure.yaml"),
		)
	}
	lines := make([]string, servicesCount)
	for index, service := range services {
		lines[index] = fmt.Sprintf(
			"    %s  %s",
			color.BlueString(service.Name),
			output.WithLinkFormat(service.IngresUrl),
		)
	}
	return strings.Join(lines, "\n")
}

func environments(environments []*ShowEnvironment) string {
	environmentsCount := len(environments)
	if environmentsCount == 0 {
		return fmt.Sprintf(
			"    You haven't created any environment. Run %s to create one.",
			color.BlueString("azd env new"),
		)
	}

	lines := make([]string, environmentsCount)
	for index, environment := range environments {
		var defaultEnv string
		if environment.IsCurrent {
			defaultEnv = cCurrentEnv
		}
		var isRemote string
		if environment.IsRemote {
			isRemote = cRemoteEnv
		}
		lines[index] = fmt.Sprintf(
			"    %s%s%s",
			color.BlueString(environment.Name),
			defaultEnv,
			output.WithGrayFormat(isRemote),
		)
	}
	return strings.Join(lines, "\n")
}

func (s *Show) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

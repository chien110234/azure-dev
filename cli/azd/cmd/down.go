package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/azure/azure-dev/cli/azd/cmd/actions"
	"github.com/azure/azure-dev/cli/azd/internal"
	"github.com/azure/azure-dev/cli/azd/pkg/account"
	"github.com/azure/azure-dev/cli/azd/pkg/alpha"
	"github.com/azure/azure-dev/cli/azd/pkg/environment"
	"github.com/azure/azure-dev/cli/azd/pkg/environment/azdcontext"
	"github.com/azure/azure-dev/cli/azd/pkg/exec"
	"github.com/azure/azure-dev/cli/azd/pkg/infra/provisioning"
	"github.com/azure/azure-dev/cli/azd/pkg/input"
	"github.com/azure/azure-dev/cli/azd/pkg/output"
	"github.com/azure/azure-dev/cli/azd/pkg/output/ux"
	"github.com/azure/azure-dev/cli/azd/pkg/project"
	"github.com/azure/azure-dev/cli/azd/pkg/tools/azcli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type downFlags struct {
	forceDelete bool
	purgeDelete bool
	global      *internal.GlobalCommandOptions
	envFlag
}

func (i *downFlags) Bind(local *pflag.FlagSet, global *internal.GlobalCommandOptions) {
	local.BoolVar(&i.forceDelete, "force", false, "Does not require confirmation before it deletes resources.")
	local.BoolVar(
		&i.purgeDelete,
		"purge",
		false,
		//nolint:lll
		"Does not require confirmation before it permanently deletes resources that are soft-deleted by default (for example, key vaults).",
	)
	i.envFlag.Bind(local, global)
	i.global = global
}

func newDownFlags(cmd *cobra.Command, global *internal.GlobalCommandOptions) *downFlags {
	flags := &downFlags{}
	flags.Bind(cmd.Flags(), global)

	return flags
}

func newDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down",
		Short: "Delete Azure resources for an application.",
	}
}

type downAction struct {
	flags               *downFlags
	accountManager      account.Manager
	azCli               azcli.AzCli
	azdCtx              *azdcontext.AzdContext
	env                 *environment.Environment
	console             input.Console
	commandRunner       exec.CommandRunner
	projectConfig       *project.ProjectConfig
	userProfileService  *azcli.UserProfileService
	subResolver         account.SubscriptionTenantResolver
	alphaFeatureManager *alpha.FeatureManager
}

func newDownAction(
	flags *downFlags,
	accountManager account.Manager,
	azCli azcli.AzCli,
	azdCtx *azdcontext.AzdContext,
	env *environment.Environment,
	projectConfig *project.ProjectConfig,
	console input.Console,
	commandRunner exec.CommandRunner,
	userProfileService *azcli.UserProfileService,
	subResolver account.SubscriptionTenantResolver,
	alphaFeatureManager *alpha.FeatureManager,
) actions.Action {
	return &downAction{
		flags:               flags,
		accountManager:      accountManager,
		azCli:               azCli,
		azdCtx:              azdCtx,
		env:                 env,
		console:             console,
		commandRunner:       commandRunner,
		projectConfig:       projectConfig,
		userProfileService:  userProfileService,
		subResolver:         subResolver,
		alphaFeatureManager: alphaFeatureManager,
	}
}

func (a *downAction) Run(ctx context.Context) (*actions.ActionResult, error) {
	// silent manager for running Plan()
	infraManager, err := createProvisioningManager(ctx, a, a.console)
	if err != nil {
		return nil, fmt.Errorf("creating provisioning manager: %w", err)
	}

	// Command title
	a.console.MessageUxItem(ctx, &ux.MessageTitle{
		Title:     "Deleting all resources and deployed code on Azure (azd down)",
		TitleNote: "Local application code is not deleted when running 'azd down'.",
	})

	startTime := time.Now()

	destroyOptions := provisioning.NewDestroyOptions(a.flags.forceDelete, a.flags.purgeDelete)
	if _, err = infraManager.Destroy(ctx, destroyOptions); err != nil {
		return nil, fmt.Errorf("deleting infrastructure: %w", err)
	}

	return &actions.ActionResult{
		Message: &actions.ResultMessage{
			Header: fmt.Sprintf("Your application was removed from Azure in %s.", ux.DurationAsText(since(startTime))),
		},
	}, nil
}

func createProvisioningManager(ctx context.Context, a *downAction, console input.Console) (*provisioning.Manager, error) {
	infraManager, err := provisioning.NewManager(
		ctx,
		a.env,
		a.projectConfig.Path,
		a.projectConfig.Infra,
		a.console.IsUnformatted(),
		a.azCli,
		console,
		a.commandRunner,
		a.accountManager,
		a.userProfileService,
		a.subResolver,
		a.alphaFeatureManager,
	)
	return infraManager, err
}

func getCmdDownHelpDescription(*cobra.Command) string {
	return generateCmdHelpDescription(fmt.Sprintf(
		"Delete Azure resources for an application. Running %s will not delete application"+
			" files on your local machine.", output.WithHighLightFormat("azd down")), nil)
}

func getCmdDownHelpFooter(*cobra.Command) string {
	return generateCmdHelpSamplesBlock(map[string]string{
		"Delete all resources for an application." +
			" You will be prompted to confirm your decision.": output.WithHighLightFormat("azd down"),
		"Forcibly delete all applications resources without confirmation.": output.WithHighLightFormat("azd down --force"),
		"Permanently delete resources that are soft-deleted by default," +
			" without confirmation.": output.WithHighLightFormat("azd down --purge"),
	})
}

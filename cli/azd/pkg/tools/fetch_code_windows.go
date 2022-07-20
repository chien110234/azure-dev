// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package tools

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/azure/azure-dev/cli/azd/pkg/executil"
	"github.com/blang/semver/v4"
)

type FetchCodeCli interface {
	ExternalTool
	FetchCode(ctx context.Context, repositoryPath string, branch string, target string) error
}

type fetchCodeCli struct {
}

func NewFetchCodeCli() FetchCodeCli {
	return &fetchCodeCli{}
}

func (cli *fetchCodeCli) versionInfo() VersionInfo {
	return VersionInfo{
		// https://docs.microsoft.com/en-us/powershell/scripting/install/powershell-support-lifecycle
		MinimumVersion: semver.Version{
			Major: 7,
			Minor: 2,
			Patch: 0},
		UpdateCommand: fmt.Sprintf("Visit %s to upgrade", cli.InstallUrl()),
	}
}

func (cli *fetchCodeCli) CheckInstalled(ctx context.Context) (bool, error) {
	powershellBinaryName := "pwsh"
	found, err := toolInPath(powershellBinaryName)
	if !found {
		return false, err
	}
	powershellVersion, err := executeCommand(ctx, powershellBinaryName, "--version")
	if err != nil {
		return false, fmt.Errorf("checking %s version: %w", cli.Name(), err)
	}
	powershellSemver, err := extractSemver(powershellVersion)
	if err != nil {
		return false, fmt.Errorf("converting to semver version fails: %w", err)
	}
	updateDetail := cli.versionInfo()
	if powershellSemver.LT(updateDetail.MinimumVersion) {
		return false, &ErrSemver{ToolName: cli.Name(), versionInfo: updateDetail}
	}
	return true, nil
}

func (cli *fetchCodeCli) InstallUrl() string {
	return "https://docs.microsoft.com/en-us/powershell/scripting/install/installing-powershell"
}

func (cli *fetchCodeCli) Name() string {
	return "powershell"
}

func (cli *fetchCodeCli) FetchCode(ctx context.Context, repositoryPath string, branch string, target string) error {
	args := []string{"clone", "--depth", "1", repositoryPath}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	args = append(args, target)

	res, err := executil.RunCommand(ctx, "git", args...)
	if err != nil {
		return fmt.Errorf("failed to clone repository %s, %s: %w", repositoryPath, res.String(), err)
	}

	if err := os.RemoveAll(filepath.Join(target, ".git")); err != nil {
		return fmt.Errorf("removing .git folder after clone: %w", err)
	}

	return nil
}

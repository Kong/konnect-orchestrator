package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/gateway"
	"github.com/Kong/konnect-orchestrator/internal/git"
	"github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/organization/role"
	"github.com/Kong/konnect-orchestrator/internal/organization/team"
	koUtil "github.com/Kong/konnect-orchestrator/internal/util"
	kk "github.com/Kong/sdk-konnect-go"
	kkComps "github.com/Kong/sdk-konnect-go/models/components"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var rootCmd = &cobra.Command{
	Use:   "koctl",
	Short: "koctl is a CLI tool for managing Konnect resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var loopInterval int

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a Platform team repository",
	Long: `Initialize a Platform team GitHub repository to utilize the Konnect Orchestrator for Konnect resource management. A konnect directory will be created in the specified directory with the default folder structure
and template files required for Konnect orchestration.`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(initCmd)
	applyCmd.Flags().IntVarP(&loopInterval, "loop", "l", 0, "Run in a loop with specified interval in seconds (0 = run once)")
}

var applyCmd = &cobra.Command{
	Use:   "apply [file]",
	Short: "Apply a configuration from file",
	Long: `Apply a configuration from a manifest file. 
The file should be in YAML format and contain the necessary resource definitions.`,
	Args: cobra.ExactArgs(1),
	RunE: runApply,
}

func processService(
	platformRepoDir string,
	orgName string,
	envName string,
	teamName string,
	serviceName string,
	serviceConfig manifest.Service,
	serviceEnvConfig manifest.EnvironmentService) error {

	serviceSpec, err := git.GetRemoteFile(
		serviceConfig.Git,
		serviceEnvConfig.Branch,
		serviceConfig.SpecPath)
	if err != nil {
		return fmt.Errorf("failed to get service spec for %s: %w",
			serviceName, err)
	}

	// Create path in the platform repo: konnect/<org>/envs/<env>/teams/<team>/services/<service-name>
	servicePath := filepath.Join(
		platformRepoDir,
		"konnect",
		orgName,
		"envs",
		envName,
		"teams",
		teamName,
		"services",
		serviceName,
	)

	if err := os.MkdirAll(servicePath, 0755); err != nil {
		return fmt.Errorf("failed to create service directory structure for %s: %w",
			serviceName, err)
	}

	// TODO: Stop waving hands at non-YAML spec files
	if err := os.WriteFile(filepath.Join(servicePath, "openapi.yaml"), serviceSpec, 0644); err != nil {
		return fmt.Errorf("failed to write service spec for %s: %w",
			serviceName, err)
	}

	return nil
}

func processOrganization(
	orgName string,
	platformGit manifest.GitConfig,
	orgConfig manifest.Organization,
	teams map[string]manifest.Team) error {

	// Resolve the organization's access token
	accessToken, err := koUtil.ResolveSecretValue(orgConfig.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to resolve access token for organization %s: %w", orgName, err)
	}

	// Initialize SDK client for this organization
	sdk := kk.New(
		kk.WithSecurity(kkComps.Security{
			PersonalAccessToken: kk.String(accessToken),
		}),
	)

	// Process each environment in the organization
	for envName, envConfig := range orgConfig.Environments {

		fmt.Printf("Processing environment %s in organization %s\n", envName, orgName)

		// Apply the control plane for the team in the environment
		regionSpecificSDK := kk.New(
			kk.WithSecurity(kkComps.Security{
				PersonalAccessToken: kk.String(accessToken),
			}),
			kk.WithServerURL(fmt.Sprintf("https://%s.api.konghq.com", envConfig.Region)),
		)

		// Process teams within this environment
		for teamName, teamEnvironmentConfig := range envConfig.Teams {

			fmt.Printf("-Processing team %s\n", teamName)

			// Get the team configuration from the global teams map
			teamConfig, exists := teams[teamName]
			if !exists {
				return fmt.Errorf("team %s referenced in organization %s environment %s not found in teams configuration",
					teamName, orgName, envName)
			}

			platformRepoDir, err := git.Clone(platformGit)
			if err != nil {
				return fmt.Errorf("failed to clone platform repository: %w", err)
			}

			// create / checkout branch
			branchName := fmt.Sprintf("%s-konnect-orchestrator-apply", envName)
			err = git.CheckoutBranch(platformRepoDir, branchName)
			if err != nil {
				return fmt.Errorf("failed to checkout branch: %w", err)
			}

			// Create folder structure for team services in platform repo
			for serviceName, serviceEnvConfig := range teamEnvironmentConfig.Services {

				fmt.Printf("--Processing service %s\n", serviceName)

				serviceConfig, exists := teamConfig.Services[serviceName]
				if !exists {
					return fmt.Errorf("service %s referenced in team %s in organization %s environment %s not found in team configuration",
						serviceName, teamName, orgName, envName)
				}

				if err := processService(
					platformRepoDir,
					orgName,
					envName,
					teamName,
					serviceName,
					serviceConfig,
					serviceEnvConfig); err != nil {
					return fmt.Errorf("failed to process service %s in team %s in organization %s environment %s: %w",
						serviceName, teamName, orgName, envName, err)
				}

			}

			isClean, err := git.IsClean(platformRepoDir)
			if err != nil {
				return fmt.Errorf("failed to check if platform repository is clean: %w", err)
			}

			if !isClean {

				fmt.Printf("-!! Changes detected for team %s in environment %s\n", teamName, envName)

				err = git.Add(platformRepoDir, ".")
				if err != nil {
					return fmt.Errorf("failed to add files to commit: %w", err)
				}
				// commit changes
				err = git.Commit(platformRepoDir, "Platform changes via Konnect Orchestrator", platformGit.Author)
				if err != nil {
					return fmt.Errorf("failed to commit changes: %w", err)
				}
				// push changes
				err = git.Push(platformRepoDir, platformGit)
				if err != nil {
					return fmt.Errorf("failed to push changes: %w", err)
				}

				_, err := github.CreateOrUpdatePullRequest(
					context.Background(),
					"KongAirlines",
					"platform",
					branchName,
					fmt.Sprintf("[Konnect] [%s] Konnect Orchestrator applied changes", envName),
					fmt.Sprintf("For the %s environment, Konnect Orchestrator has detected changes in upstream service repositories and has generated the associated changes.", envName),
					platformGit.GitHub,
				)
				if err != nil {
					return fmt.Errorf("failed to create or update pull request: %w", err)
				}
			} else {
				fmt.Printf("-No changes for team %s in environment %s\n", teamName, envName)
			}

			// Create/update the team
			teamID, err := team.ApplyTeam(
				context.Background(),
				sdk.Teams,
				sdk.TeamMembership,
				sdk.Users,
				sdk.Invites,
				teamName,
				teamConfig,
			)
			if err != nil || teamID == "" {
				return fmt.Errorf("failed to apply team %s in organization %s environment %s: %w",
					teamName, orgName, envName, err)
			}

			cpID, err := gateway.ApplyControlPlane(
				context.Background(),
				regionSpecificSDK.ControlPlanes,
				envName,
				envConfig,
				teamName)

			if err != nil || cpID == "" {
				return fmt.Errorf("failed to apply control plane for team %s in organization %s environment %s: %w",
					teamName, orgName, envName, err)
			}

			// Apply roles for the team in the environment
			if err := role.ApplyRoles(
				context.Background(),
				sdk.Roles,
				teamID,
				cpID,
				envConfig); err != nil {
				return fmt.Errorf("failed to apply team roles: %w", err)
			}
		}
	}

	fmt.Printf("Successfully applied configuration for organization: %s\n", orgName)
	return nil
}

func runApply(cmd *cobra.Command, args []string) error {
	filePath := args[0]

	applyOnce := func() error {
		// Validate file exists and is readable
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return fmt.Errorf("failed to resolve file path: %w", err)
		}

		if _, err := os.Stat(absPath); err != nil {
			return fmt.Errorf("failed to access file %s: %w", absPath, err)
		}

		// Read and parse the manifest file
		fileContent, err := os.ReadFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		var manifest manifest.Orchestrator
		if err := yaml.Unmarshal(fileContent, &manifest); err != nil {
			// If YAML parsing fails, try JSON
			if jsonErr := json.Unmarshal(fileContent, &manifest); jsonErr != nil {
				return fmt.Errorf("failed to parse manifest as YAML or JSON: %w", err)
			}
		}

		// Process each organization
		for orgName, orgConfig := range manifest.Organizations {
			if err := processOrganization(orgName, manifest.Platform.Git, orgConfig, manifest.Teams); err != nil {
				return err
			}
		}

		fmt.Printf("Successfully applied configuration from: %s\n", absPath)
		return nil
	}

	if loopInterval == 0 {
		return applyOnce()
	}

	for {
		if err := applyOnce(); err != nil {
			return err
		}
		fmt.Printf("--- Waiting %d seconds before next control loop \n", loopInterval)
		time.Sleep(time.Duration(loopInterval) * time.Second)
	}
}

func copyDir(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", src, err)
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dst, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			content, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", srcPath, err)
			}

			if err := os.WriteFile(dstPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", dstPath, err)
			}
		}
	}

	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	targetDir := args[0]

	// Create the base konnect directory
	konnectDir := filepath.Join(targetDir, "konnect")
	if err := os.MkdirAll(konnectDir, 0755); err != nil {
		return fmt.Errorf("failed to create konnect directory: %w", err)
	}

	// Copy template files from resources/platform
	templateFiles := []string{
		"konnect.yaml",
		"oas-file-rules.yaml",
		"deck-file-rules.yaml",
		".gitignore",
	}

	for _, file := range templateFiles {
		srcPath := filepath.Join("resources", "platform", file)
		dstPath := filepath.Join(konnectDir, file)

		content, err := os.ReadFile(srcPath)
		if err != nil {
			return fmt.Errorf("failed to read template file %s: %w", file, err)
		}

		if err := os.WriteFile(dstPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", file, err)
		}
	}

	// Recursively copy .github directory and its contents
	srcGithubDir := filepath.Join("resources", "platform", ".github")
	dstGithubDir := filepath.Join(konnectDir, ".github")
	if err := copyDir(srcGithubDir, dstGithubDir); err != nil {
		return fmt.Errorf("failed to copy .github directory: %w", err)
	}

	fmt.Printf("Successfully initialized Konnect configuration in: %s\n", konnectDir)
	return nil
}

func Execute() error {
	return rootCmd.Execute()
}

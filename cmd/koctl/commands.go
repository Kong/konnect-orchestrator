package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/gateway"
	"github.com/Kong/konnect-orchestrator/internal/git"
	"github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/organization/auth"
	"github.com/Kong/konnect-orchestrator/internal/organization/portal"
	"github.com/Kong/konnect-orchestrator/internal/organization/role"
	"github.com/Kong/konnect-orchestrator/internal/organization/team"
	koUtil "github.com/Kong/konnect-orchestrator/internal/util"
	kk "github.com/Kong/sdk-konnect-go"
	kkInternal "github.com/Kong/sdk-konnect-go-internal"
	kkInternalComps "github.com/Kong/sdk-konnect-go-internal/models/components"
	kkComps "github.com/Kong/sdk-konnect-go/models/components"
	giturl "github.com/kubescape/go-git-url"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var loopInterval int
var platformFileArg string
var teamsFileArg string
var organizationsFileArg string

var rootCmd = &cobra.Command{
	Use:   "koctl",
	Short: "koctl is a CLI tool for managing Konnect resources",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize a Platform team repository",
	Long: `Initialize a Platform team GitHub repository to utilize the Konnect Orchestrator for Konnect resource management. A konnect directory will be created in the specified directory with the default folder structure
and template files required for Konnect orchestration.`,
	Args: cobra.ExactArgs(1),
	RunE: runInit,
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration to Konnect organizations",
	Long: `The orchestrator will apply a Federated API strategy to one to many Konnect organizations
based on the input received in 3 files. A Platform team config, a teams configuration, and an organizations configuration. 
The files should be in YAML format and contain the necessary resource definitions. See the init command for an example of the required structure.`,
	RunE: runApply,
}

func init() {
	applyCmd.Flags().StringVar(&platformFileArg, "platform", "", "Path to the platform configuration file")
	applyCmd.Flags().StringVar(&teamsFileArg, "teams", "", "Path to the teams configuration file")
	applyCmd.Flags().StringVar(&organizationsFileArg, "orgs", "", "Path to the organizations configuration file")
	applyCmd.Flags().IntVarP(&loopInterval, "loop", "l", 0, "Run in a loop with specified interval in seconds (0 = run once)")

	_ = applyCmd.MarkFlagRequired("platform")
	_ = applyCmd.MarkFlagRequired("teams")
	_ = applyCmd.MarkFlagRequired("orgs")

	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(initCmd)

}

func processService(
	platformRepoDir string,
	orgName string,
	envName string,
	envType string,
	teamName string,
	serviceName string,
	serviceConfig manifest.Service,
	serviceEnvConfig manifest.EnvironmentService,
	portalId string,
	region string,
	accessToken string,
	labels map[string]string) error {

	labels["team-name"] = teamName
	labels["service-name"] = *serviceConfig.Name

	serviceSpec, err := git.GetRemoteFile(
		*serviceConfig.Git,
		*serviceEnvConfig.Branch,
		*serviceConfig.SpecPath)
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

	internalRegionSdk := kkInternal.New(
		kkInternal.WithSecurity(kkInternalComps.Security{
			PersonalAccessToken: kkInternal.String(accessToken),
		}),
		kkInternal.WithServerURL(fmt.Sprintf("https://%s.api.konghq.com", region)),
	)

	apiName := serviceConfig.Name
	if envType != "PROD" {
		apiName = kk.String(fmt.Sprintf("%s-%s", *apiName, envName))
	}

	// Apply the API configuration for this service api
	err = portal.ApplyApiConfig(
		context.Background(),
		internalRegionSdk.API,
		internalRegionSdk.APISpecification,
		internalRegionSdk.APIPublication,
		*apiName,
		serviceConfig,
		serviceSpec,
		portalId,
		labels)
	if err != nil {
		return err
	}

	return nil
}

func processPortal(
	accessToken string,
	portalDisplayName string,
	region string,
	envName string,
	envType string,
	labels map[string]string) (string, error) {
	// V3 Portals currently require an internal SDK as the API is not yet GA
	internalRegionSdk := kkInternal.New(
		kkInternal.WithSecurity(kkInternalComps.Security{
			PersonalAccessToken: kkInternal.String(accessToken),
		}),
		kkInternal.WithServerURL(fmt.Sprintf("https://%s.api.konghq.com", region)),
	)

	// Apply the Developer Portal configuration for the environment
	portalId, err := portal.ApplyPortalConfig(context.Background(),
		portalDisplayName,
		envName,
		envType,
		internalRegionSdk.V3Portals,
		internalRegionSdk.API,
		labels)
	if err != nil {
		return "", fmt.Errorf("failed to apply portal configuration: %w", err)
	}
	return portalId, nil
}

func processOrganization(
	orgName string,
	platformGit manifest.GitConfig,
	orgConfig manifest.Organization,
	teams map[string]*manifest.Team) error {

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

	if orgConfig.Authorization != nil {
		fmt.Printf("Applying organization authorization settings to organization %s\n", orgName)
		err = auth.ApplyAuthSettings(
			context.Background(),
			sdk.AuthSettings,
			sdk.AuthSettings,
			sdk.Teams,
			sdk.AuthSettings,
			*orgConfig.Authorization)
		if err != nil {
			return fmt.Errorf("failed to apply auth settings for organization %s: %w", orgName, err)
		}
	}

	// Process each environment in the organization
	for envName, envConfig := range orgConfig.Environments {
		err := processEnvironment(
			envName, orgName,
			accessToken,
			*envConfig, teams, platformGit, sdk)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Successfully applied configuration for organization: %s\n", orgName)
	return nil
}

func processEnvironment(
	envName string,
	orgName string,
	accessToken string,
	envConfig manifest.Environment,
	teams map[string]*manifest.Team,
	platformGit manifest.GitConfig,
	sdk *kk.SDK) error {
	fmt.Printf("Processing environment %s in organization %s\n", envName, orgName)

	labels := map[string]string{
		// 'konnect' is a reserved prefix for labels
		"ko-konnect-orchestrator": "true",
		"env-name":                envName,
		"env-type":                envConfig.Type,
	}

	portalId, err := processPortal(
		accessToken,
		orgName,
		envConfig.Region,
		envName,
		envConfig.Type,
		labels)
	if err != nil {
		return err
	}

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
		err = git.CheckoutBranch(platformRepoDir, branchName, platformGit)
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
				envConfig.Type,
				teamName,
				serviceName,
				*serviceConfig,
				*serviceEnvConfig,
				portalId,
				envConfig.Region,
				accessToken,
				labels); err != nil {
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
			err = git.Commit(platformRepoDir, "Platform changes via Konnect Orchestrator", *platformGit.Author)
			if err != nil {
				return fmt.Errorf("failed to commit changes: %w", err)
			}
			// push changes
			err = git.Push(platformRepoDir, platformGit)
			if err != nil {
				return fmt.Errorf("failed to push changes: %w", err)
			}

			gitURL, err := giturl.NewGitURL(*platformGit.Remote)
			if err != nil {
				return fmt.Errorf("failed to parse Git URL: %w", err)
			}

			_, err = github.CreateOrUpdatePullRequest(
				context.Background(),
				gitURL.GetOwnerName(),
				gitURL.GetRepoName(),
				branchName,
				fmt.Sprintf("[Konnect] [%s] Konnect Orchestrator applied changes", envName),
				fmt.Sprintf("For the %s environment, Konnect Orchestrator has detected changes in upstream service repositories and has generated the associated changes.", envName),
				*platformGit.GitHub,
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
			*teamConfig,
		)
		if err != nil || teamID == "" {
			return fmt.Errorf("failed to apply team %s in organization %s environment %s: %w",
				teamName, orgName, envName, err)
		}

		regionSpecificSDK := kk.New(
			kk.WithSecurity(kkComps.Security{
				PersonalAccessToken: kk.String(accessToken),
			}),
			kk.WithServerURL(fmt.Sprintf("https://%s.api.konghq.com", envConfig.Region)),
		)

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
	return nil
}

func readConfigSection(filePath string, out interface{}) error {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to make absolute path: %w", err)
	}

	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("file not accessible: %w", err)
	}

	bytes, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Try YAML first
	if err := yaml.Unmarshal(bytes, out); err != nil {
		// Fall back to JSON if YAML fails
		if jsonErr := json.Unmarshal(bytes, out); jsonErr != nil {
			return fmt.Errorf("failed to parse file as YAML or JSON: %v", err)
		}
	}
	return nil
}

func runApply(cmd *cobra.Command, args []string) error {
	applyOnce := func() error {
		platformFilePath, err := filepath.Abs(platformFileArg)
		if err != nil {
			return fmt.Errorf("failed to resolve platform file path: %w", err)
		}
		if _, err := os.Stat(platformFilePath); err != nil {
			return fmt.Errorf("failed to access file %s: %w", platformFilePath, err)
		}
		teamsFilePath, err := filepath.Abs(teamsFileArg)
		if err != nil {
			return fmt.Errorf("failed to resolve teams file path: %w", err)
		}
		if _, err := os.Stat(teamsFilePath); err != nil {
			return fmt.Errorf("failed to access file %s: %w", teamsFilePath, err)
		}
		organizationsFilePath, err := filepath.Abs(organizationsFileArg)
		if err != nil {
			return fmt.Errorf("failed to resolve organizations file path: %w", err)
		}
		if _, err := os.Stat(organizationsFilePath); err != nil {
			return fmt.Errorf("failed to access file %s: %w", organizationsFilePath, err)
		}

		var manifest manifest.Orchestrator
		if err := readConfigSection(platformFilePath, &manifest.Platform); err != nil {
			return fmt.Errorf("failed to read platform configuration: %w", err)
		}
		if err := readConfigSection(teamsFilePath, &manifest.Teams); err != nil {
			return fmt.Errorf("failed to read teams configuration: %w", err)
		}
		if err := readConfigSection(organizationsFilePath, &manifest.Organizations); err != nil {
			return fmt.Errorf("failed to read organizations configuration: %w", err)
		}

		// Process each organization
		for orgName, orgConfig := range manifest.Organizations {
			if err := processOrganization(orgName, *manifest.Platform.Git, *orgConfig, manifest.Teams); err != nil {
				return err
			}
		}

		fmt.Printf("Successfully applied configuration from:\n  - %s\n  - %s\n  - %s\n",
			platformFileArg, teamsFileArg, organizationsFileArg)

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

	// Create directory if it doesn't exist (preserve if it does)
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

func mergeGitignore(srcPath, dstPath string) error {
	// Read source .gitignore content
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source .gitignore: %w", err)
	}

	srcLines := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(srcContent)), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			srcLines[line] = true
		}
	}

	var dstLines []string
	// Read destination .gitignore if it exists
	if _, err := os.Stat(dstPath); err == nil {
		dstContent, err := os.ReadFile(dstPath)
		if err != nil {
			return fmt.Errorf("failed to read destination .gitignore: %w", err)
		}
		dstLines = strings.Split(strings.TrimSpace(string(dstContent)), "\n")

		// Remove empty lines and check which source lines already exist
		for _, line := range dstLines {
			if line = strings.TrimSpace(line); line != "" {
				srcLines[line] = false // Mark as already existing
			}
		}
	}

	// Append only new lines from source
	f, err := os.OpenFile(dstPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open destination .gitignore: %w", err)
	}
	defer f.Close()

	// If the file is not empty and doesn't end with a newline, add one
	if len(dstLines) > 0 {
		if _, err := f.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline to .gitignore: %w", err)
		}
	}

	// Write new lines
	for line, shouldAdd := range srcLines {
		if shouldAdd {
			if _, err := f.WriteString(line + "\n"); err != nil {
				return fmt.Errorf("failed to write to .gitignore: %w", err)
			}
		}
	}

	return nil
}

func runInit(cmd *cobra.Command, args []string) error {
	targetDir := args[0]

	// Copy the entire konnect directory structure
	srcKonnectDir := filepath.Join("resources", "platform", "konnect")
	dstKonnectDir := filepath.Join(targetDir, "konnect")
	if err := copyDir(srcKonnectDir, dstKonnectDir); err != nil {
		return fmt.Errorf("failed to copy konnect directory: %w", err)
	}

	// Handle .gitignore specially - merge with existing if present
	srcGitignore := filepath.Join("resources", "platform", ".gitignore")
	dstGitignore := filepath.Join(targetDir, ".gitignore")
	if err := mergeGitignore(srcGitignore, dstGitignore); err != nil {
		return fmt.Errorf("failed to handle .gitignore: %w", err)
	}

	// Copy .github directory to the base target directory
	srcGithubDir := filepath.Join("resources", "platform", ".github")
	dstGithubDir := filepath.Join(targetDir, ".github")
	if err := copyDir(srcGithubDir, dstGithubDir); err != nil {
		return fmt.Errorf("failed to copy .github directory: %w", err)
	}

	fmt.Printf("Successfully initialized Konnect configuration in: %s\n", dstKonnectDir)
	fmt.Printf("GitHub workflows have been added to: %s\n", dstGithubDir)
	fmt.Printf("Updated .gitignore at: %s\n", dstGitignore)
	fmt.Println("\nNext steps:")
	fmt.Println("1. Review and customize the konnect.yaml file in the konnect directory")
	fmt.Println("\t- Configure your team's platform repository in the platform key")
	fmt.Println("\t- Add and configure your organization's teams and their services teams configuration")
	fmt.Println("\t- Define your desired Konnect organizational layout in the organizations key")
	fmt.Println("\t- Commit and push your changes to your platform repository")
	fmt.Println("2. In each of your Konnect organizations, add a System Account named `konnect-orchestrator`")
	fmt.Println("\t- Assign the `konnect-orchestrator` account the `Organization Admin` role")
	fmt.Println("\t- Create a new system token for the `konnect-orchestrator` account and store where available to the orchestrator")
	fmt.Println("3. Configure your Platform GitHub repository with the necessary authorizations for workflows")
	fmt.Println("\t- For each Konnect organization, add a `<ORG_NAME>_KONNECT_TOKEN` secret to the repository in the GitHub secrets")
	fmt.Println("\t- Give Actions read and write permissions in the repository for all scopes. GH Settings")
	fmt.Println("4. Run 'koctl apply <dir>/konnect/konnect.yaml' to apply your configuration")
	return nil
}

func Execute() error {
	return rootCmd.Execute()
}

package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Kong/konnect-orchestrator/internal/deck/patch"
	"github.com/Kong/konnect-orchestrator/internal/gateway"
	"github.com/Kong/konnect-orchestrator/internal/git"
	"github.com/Kong/konnect-orchestrator/internal/git/github"
	"github.com/Kong/konnect-orchestrator/internal/manifest"
	"github.com/Kong/konnect-orchestrator/internal/notification"
	"github.com/Kong/konnect-orchestrator/internal/organization/auth"
	"github.com/Kong/konnect-orchestrator/internal/organization/portal"
	"github.com/Kong/konnect-orchestrator/internal/organization/role"
	"github.com/Kong/konnect-orchestrator/internal/organization/team"
	"github.com/Kong/konnect-orchestrator/internal/reports"
	koUtil "github.com/Kong/konnect-orchestrator/internal/util"
	kk "github.com/Kong/sdk-konnect-go"
	kkInternal "github.com/Kong/sdk-konnect-go-internal"
	kkInternalComps "github.com/Kong/sdk-konnect-go-internal/models/components"
	kkInternalOps "github.com/Kong/sdk-konnect-go-internal/models/operations"
	kkComps "github.com/Kong/sdk-konnect-go/models/components"
	giturl "github.com/kubescape/go-git-url"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

//go:embed resources/platform/* resources/platform/.github/* resources/platform/.gitignore
var resourceFiles embed.FS
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

func applyService(
	platformRepoDir string,
	platformGit manifest.GitConfig,
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
	cpId string,
	labels map[string]string) error {

	labels["team-name"] = teamName
	labels["service-name"] = *serviceConfig.Name

	svcGitCfg := serviceConfig.Git
	if svcGitCfg.Auth == nil {
		// If the user doesn't provide a service level git auth config, we use the platform level git auth
		svcGitCfg.Auth = platformGit.Auth
	}

	// This loads the Service Spec from the teams Git Repository
	// into memory
	serviceSpec, err := git.GetRemoteFile(
		*svcGitCfg,
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

	// This copies the Spec from memory into the Platform team Git repository location
	// TODO: Stop waving hands at non-YAML spec files
	if err := os.WriteFile(filepath.Join(servicePath, "openapi.yaml"), serviceSpec, 0644); err != nil {
		return fmt.Errorf("failed to write service spec for %s: %w",
			serviceName, err)
	}

	apiName := serviceConfig.Name
	if envType != "PROD" {
		apiName = kk.String(fmt.Sprintf("%s-%s", *apiName, envName))
	}

	// Write a patch files for the service adding some metadata so we can relate the API to the GW service later
	apiNameServicePatch := patch.Patch{
		Selectors: []string{"$..services[*]"},
		Values: map[string]interface{}{
			"tags": []string{
				"ko-api-name=" + *apiName,
			},
		},
	}
	koPatchFile := patch.PatchFile{
		FormatVersion: "1.0",
		Patches:       []patch.Patch{apiNameServicePatch},
	}

	// write a patch file to the service directory under the name "ko-patch.yaml"
	koPatchFileBytes, err := yaml.Marshal(koPatchFile)
	if err != nil {
		return fmt.Errorf("failed to marshal patch file for %s: %w", serviceName, err)
	}
	if err := os.WriteFile(filepath.Join(servicePath, "ko-patch.yaml"), koPatchFileBytes, 0644); err != nil {
		return fmt.Errorf("failed to write patch file for %s: %w", serviceName, err)
	}

	internalRegionSdk := kkInternal.New(
		kkInternal.WithSecurity(kkInternalComps.Security{
			PersonalAccessToken: kkInternal.String(accessToken),
		}),
		kkInternal.WithServerURL(fmt.Sprintf("https://%s.api.konghq.com", region)),
	)

	// We can now query for GW Services that have the `ko-api-name` tag, this will require that the
	// APIOps pipeline in the Platform repository has ran, such that the entity is tagged properly so we can find it
	// here. If we can't find the service, we just ignore and proceed.
	resp, err := internalRegionSdk.Services.ListService(context.Background(),
		kkInternalOps.ListServiceRequest{
			ControlPlaneID: cpId,
			Tags:           kkInternal.String("ko-api-name=" + *apiName),
		})
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("failed to list services: response is nil")
	}
	services := resp.Object.GetData()
	if services == nil {
		return fmt.Errorf("failed to list services: data is nil")
	}
	var serviceId string
	if len(services) == 1 {
		serviceId = *services[0].GetID()
	} else {
		fmt.Printf("!!!Found %d serivces for API %s. Cannot create API implementation relation, requires exactly 1 service with `ko-api-name` tag.\n", len(services), *apiName)
		return nil
	}

	_, err = portal.ApplyApiConfig(
		context.Background(),
		internalRegionSdk.API,
		internalRegionSdk.APISpecification,
		internalRegionSdk.APIPublication,
		internalRegionSdk.APIImplementation,
		*apiName,
		serviceConfig,
		serviceSpec,
		portalId,
		cpId,
		serviceId,
		labels)
	if err != nil {
		return err
	}

	return nil
}

func applyPortal(
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

func applyTeam(teamName string,
	accessToken string,
	envConfig manifest.Environment,
	envName string,
	orgName string,
	teamConfig manifest.Team,
	sdk *kk.SDK,
	platformGit manifest.GitConfig,
	teamEnvironmentConfig *manifest.TeamEnvironment,
	portalId string,
	labels map[string]string) error {

	fmt.Printf("-Processing team %s\n", teamName)

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

	// Apply roles for the team in the environment
	if err := role.ApplyRoles(
		context.Background(),
		sdk.Roles,
		teamID,
		cpID,
		envConfig); err != nil {
		return fmt.Errorf("failed to apply team roles: %w", err)
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

	if teamEnvironmentConfig != nil {
		for serviceName, serviceEnvConfig := range teamEnvironmentConfig.Services {

			fmt.Printf("--Processing service %s\n", serviceName)

			serviceConfig, exists := teamConfig.Services[serviceName]
			if !exists {
				return fmt.Errorf("service %s referenced in team %s in organization %s environment %s not found in team configuration",
					serviceName, teamName, orgName, envName)
			}

			if err := applyService(
				platformRepoDir,
				platformGit,
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
				cpID,
				labels); err != nil {
				return fmt.Errorf("failed to process service %s in team %s in organization %s environment %s: %w",
					serviceName, teamName, orgName, envName, err)
			}
		}
	} else {
		for serviceName, serviceConfig := range teamConfig.Services {

			fmt.Printf("--Processing service %s\n", serviceName)

			serviceEnvConfig := manifest.EnvironmentService{}
			if envConfig.Type == "PROD" {
				serviceEnvConfig.Branch = serviceConfig.ProdBranch
			} else {
				serviceEnvConfig.Branch = serviceConfig.DevBranch
			}

			if err := applyService(
				platformRepoDir,
				platformGit,
				orgName,
				envName,
				envConfig.Type,
				teamName,
				serviceName,
				*serviceConfig,
				serviceEnvConfig,
				portalId,
				envConfig.Region,
				accessToken,
				cpID,
				labels); err != nil {
				return fmt.Errorf("failed to process service %s in team %s in organization %s environment %s: %w",
					serviceName, teamName, orgName, envName, err)
			}
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

	return nil
}

func applyEnvironment(
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

	portalId, err := applyPortal(
		accessToken,
		orgName,
		envConfig.Region,
		envName,
		envConfig.Type,
		labels)
	if err != nil {
		return err
	}

	if envConfig.Teams == nil { // By default all teams are added to environments
		for teamName, teamConfig := range teams {
			err := applyTeam(
				teamName,
				accessToken,
				envConfig,
				envName,
				orgName,
				*teamConfig,
				sdk,
				platformGit,
				nil, // nil because we use the default config in the teamConfig
				portalId,
				labels)
			if err != nil {
				return err
			}
		}
	} else { // The user wants to specify individual teams for this environment
		// Process teams within this environment
		for teamName, teamEnvironmentConfig := range envConfig.Teams {

			err := applyTeam(
				teamName,
				accessToken,
				envConfig,
				envName,
				orgName,
				*teams[teamName],
				sdk,
				platformGit,
				teamEnvironmentConfig,
				portalId,
				labels)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func applyOrganization(
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
		fmt.Printf("Applying authorization settings to organization %s\n", orgName)
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

	regions := map[string]struct{}{}

	// Process each environment in the organization
	for envName, envConfig := range orgConfig.Environments {
		err := applyEnvironment(
			envName, orgName,
			accessToken,
			*envConfig, teams, platformGit, sdk)
		if err != nil {
			return err
		}
		regions[envConfig.Region] = struct{}{}
	}

	// Default is true, so create if it's missing or truthy
	if orgConfig.EnableCustomReports == nil || *orgConfig.EnableCustomReports {
		for region := range regions {
			internalRegionSdk := reports.New(
				reports.WithSecurity(kkInternalComps.Security{
					PersonalAccessToken: kk.String(accessToken),
				}),
				reports.WithServerURL(fmt.Sprintf("https://%s.api.konghq.com", region)),
			)
			fmt.Printf("Creating default custom reports for organization %s in region %s\n", orgName, region)
			err = reports.ApplyReports(
				context.Background(),
				internalRegionSdk.CustomReports)
			if err != nil {
				return fmt.Errorf("failed to create custom reports for organization %s: %w", orgName, err)
			}
		}
	}

	internalRegionSdk := kkInternal.New(
		kkInternal.WithSecurity(kkInternalComps.Security{
			PersonalAccessToken: kk.String(accessToken),
		}),
	)

	fmt.Printf("Applying notification configuration settings to organization %s\n", orgName)
	err = notification.ApplyNotificationsConfig(
		context.Background(),
		internalRegionSdk.Notifications,
		orgConfig.Notifications)
	if err != nil {
		return fmt.Errorf("failed to apply notification configurations for organization %s: %w", orgName, err)
	}

	fmt.Printf("Successfully applied configuration for organization: %s\n", orgName)
	return nil
}

// copyFile copies a single file from the embedded FS to the local filesystem
func copyFile(embedFS embed.FS, srcPath, dstPath string) error {
	// Open the embedded file
	srcFile, err := embedFS.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open embedded file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	// Ensure the destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", filepath.Dir(dstPath), err)
	}

	// Create the destination file
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", dstPath, err)
	}
	defer dstFile.Close()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy data to %s: %w", dstPath, err)
	}

	return nil
}

// copyEmbeddedFilesRecursive recursively copies files from an embedded FS to the target directory
func copyEmbeddedFilesRecursive(embedFS embed.FS, srcDir, dstDir string) error {
	entries, err := embedFS.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", srcDir, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(srcDir, entry.Name())
		dstPath := filepath.Join(dstDir, entry.Name())

		if entry.IsDir() {
			// Ensure the destination directory exists
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
			}
			// Recurse into subdirectory
			if err := copyEmbeddedFilesRecursive(embedFS, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy individual file
			if err := copyFile(embedFS, srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyPlatformRepo(gitCfg *manifest.GitConfig) error {

	// Apply changes to the Platform repository, these include Workflow files, configurations, etc...
	platformRepoDir, err := git.Clone(*gitCfg)
	if err != nil {
		return fmt.Errorf("failed to clone platform repository: %w", err)
	}

	// create a branch with a well known name
	branchName := "platform-konnect-orchestrator-apply"
	err = git.CheckoutBranch(platformRepoDir, branchName, *gitCfg)
	if err != nil {
		return fmt.Errorf("failed to checkout branch: %w", err)
	}

	err = copyEmbeddedFilesRecursive(resourceFiles, "resources/platform", platformRepoDir)
	if err != nil {
		return fmt.Errorf("failed to copy resource files to platform repository: %w", err)
	}

	// Detect changes to the repository
	isClean, err := git.IsClean(platformRepoDir)
	if err != nil {
		return fmt.Errorf("failed to check if platform repository is clean: %w", err)
	}
	if !isClean {
		fmt.Println("Changes detected in platform repository")
		err = git.Add(platformRepoDir, ".")
		if err != nil {
			return fmt.Errorf("failed to add files to commit: %w", err)
		}
		err = git.Commit(platformRepoDir, "Platform changes via Konnect Orchestrator", *gitCfg.Author)
		if err != nil {
			return fmt.Errorf("failed to commit changes: %w", err)
		}
		err = git.Push(platformRepoDir, *gitCfg)
		if err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}

		gitURL, err := giturl.NewGitURL(*gitCfg.Remote)
		if err != nil {
			return fmt.Errorf("failed to parse Git URL: %w", err)
		}

		_, err = github.CreateOrUpdatePullRequest(
			context.Background(),
			gitURL.GetOwnerName(),
			gitURL.GetRepoName(),
			branchName,
			"[Konnect] Konnect Orchestrator applied changes - Platform",
			"Makes changes for the Platform repository automations, including GitHub Actions and configurations.",
			*gitCfg.GitHub)
		if err != nil {
			return fmt.Errorf("failed to create or update pull request: %w", err)
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

		err = applyPlatformRepo(manifest.Platform.Git)
		if err != nil {
			return fmt.Errorf("failed to apply platform repository changes: %w", err)
		}

		// Process each organization
		for orgName, orgConfig := range manifest.Organizations {
			if err := applyOrganization(orgName, *manifest.Platform.Git, *orgConfig, manifest.Teams); err != nil {
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

func Execute() error {
	return rootCmd.Execute()
}

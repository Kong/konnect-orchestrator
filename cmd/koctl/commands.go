package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

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
	"github.com/Kong/konnect-orchestrator/internal/platform"
	"github.com/Kong/konnect-orchestrator/internal/reports"
	"github.com/Kong/konnect-orchestrator/internal/server"
	koUtil "github.com/Kong/konnect-orchestrator/internal/util"
	kk "github.com/Kong/sdk-konnect-go"
	kkInternal "github.com/Kong/sdk-konnect-go-internal"
	kkInternalComps "github.com/Kong/sdk-konnect-go-internal/models/components"
	kkInternalOps "github.com/Kong/sdk-konnect-go-internal/models/operations"
	kkComps "github.com/Kong/sdk-konnect-go/models/components"
	giturl "github.com/kubescape/go-git-url"
)

//go:embed resources/platform/* resources/platform/.github/* resources/platform/.gitignore
var resourceFiles embed.FS

const (
	defaultOrchestratorPath = "konnect/"
	defaultPlatformFilePath = defaultOrchestratorPath + "platform.yaml"
	defaultTeamsFilePath    = defaultOrchestratorPath + "teams.yaml"
	defaultOrgsFilePath     = defaultOrchestratorPath + "organizations.yaml"
)

var (
	loopInterval         int
	wholeFileArg         string
	platformFileArg      string
	teamsFileArg         string
	organizationsFileArg string
	version              = "dev"
	commit               = "unknown"
	date                 = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "koctl",
	Short: "koctl is a CLI tool for managing Konnect resources",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cmd.Help()
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a platform repository",
	RunE:  runInit,
}

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration to Konnect organizations",
	Long: `The orchestrator will apply a Federated API strategy to one to many Konnect organizations
based on the input received in 3 files. A Platform team config, a teams configuration, 
and an organizations configuration. The files should be in YAML format and contain the 
necessary resource definitions. See the init command for an example of the required structure.`,
	RunE: runApply,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the orchestrators API server",
	Long:  `The orchestrator can run an API server which can handle HTTP requests related to the state of the platform repository.`,
	RunE:  runRun,
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new resource to the platform repository",
}

var addOrganizationCmd = &cobra.Command{
	Use:   "organization",
	Short: "Add a new organization to the platform repository",
	Long:  `Use this command to add a new organization to the platform repository via a PR filed in the platform repository`,
	RunE:  runAddOrganization,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of koctl",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("koctl version: %s\nCommit: %s\nBuild date: %s\n", version, commit, date)
	},
}

func init() {
	addCmd.AddCommand(addOrganizationCmd)

	applyCmd.Flags().StringVar(&wholeFileArg,
		"file",
		"",
		"Path to the configuration file. This is a convenience flag to apply the whole configuration in one file")
	applyCmd.Flags().StringVar(&platformFileArg,
		"platform",
		"./"+defaultPlatformFilePath,
		"Path to the platform configuration file. Superseded by --file")
	applyCmd.Flags().StringVar(&teamsFileArg,
		"teams",
		"./"+defaultTeamsFilePath,
		"Path to the teams configuration file. Superseded by --file")
	applyCmd.Flags().StringVar(&organizationsFileArg,
		"orgs",
		"./"+defaultOrgsFilePath,
		"Path to the organizations configuration file. Superseded by --file")
	applyCmd.Flags().IntVarP(&loopInterval,
		"loop", "l", 0, "Run apply in a loop with specified interval in seconds (0 = run once)")

	runCmd.Flags().StringVar(&platformFileArg,
		"platform",
		"./"+defaultPlatformFilePath,
		"Path to the platform configuration file.")

	initCmd.Flags().StringVar(&platformFileArg,
		"platform",
		"./"+defaultPlatformFilePath,
		"Path to the platform configuration file.")

	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(initCmd)
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
			return fmt.Errorf("failed to parse file as YAML or JSON: %w", err)
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
	portalID string,
	region string,
	accessToken string,
	cpID string,
	labels map[string]string,
) error {
	labels["team-name"] = teamName
	labels["service-name"] = *serviceConfig.Name

	svcGitCfg := serviceConfig.Git
	if svcGitCfg.Auth == nil {
		// If the user doesn't provide a service level git auth config, we use the platform level git auth
		if platformGit.Auth == nil {
			svcGitCfg.GitHub = platformGit.GitHub
		} else {
			svcGitCfg.Auth = platformGit.Auth
		}
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

	if err := os.MkdirAll(servicePath, 0o755); err != nil {
		return fmt.Errorf("failed to create service directory structure for %s: %w",
			serviceName, err)
	}

	// This copies the Spec from memory into the Platform team Git repository location
	// TODO: Stop waving hands at non-YAML spec files
	if err := os.WriteFile(filepath.Join(servicePath, "openapi.yaml"), serviceSpec, 0o600); err != nil {
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
	koPatchFile := patch.File{
		FormatVersion: "1.0",
		Patches:       []patch.Patch{apiNameServicePatch},
	}

	// write a patch file to the service directory under the name "ko-patch.yaml"
	koPatchFileBytes, err := yaml.Marshal(koPatchFile)
	if err != nil {
		return fmt.Errorf("failed to marshal patch file for %s: %w", serviceName, err)
	}
	if err := os.WriteFile(filepath.Join(servicePath, "ko-patch.yaml"), koPatchFileBytes, 0o600); err != nil {
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
			ControlPlaneID: cpID,
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
	var serviceID string
	if len(services) == 1 {
		serviceID = *services[0].GetID()
	} else {
		fmt.Printf("Warn: Found %d services for API %s. Cannot create API implementation relation, "+
			"requires exactly 1 service with `ko-api-name` tag. APIOps workflows may need to be ran.\n", len(services), *apiName)
	}

	_, err = portal.ApplyAPIConfig(
		context.Background(),
		internalRegionSdk.API,
		internalRegionSdk.APISpecification,
		internalRegionSdk.APIPublication,
		internalRegionSdk.APIImplementation,
		*apiName,
		serviceConfig,
		serviceSpec,
		portalID,
		cpID,
		serviceID,
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
	labels map[string]string,
) (string, error) {
	// V3 Portals currently require an internal SDK as the API is not yet GA
	internalRegionSdk := kkInternal.New(
		kkInternal.WithSecurity(kkInternalComps.Security{
			PersonalAccessToken: kkInternal.String(accessToken),
		}),
		kkInternal.WithServerURL(fmt.Sprintf("https://%s.api.konghq.com", region)),
	)

	// Apply the Developer Portal configuration for the environment
	portalID, err := portal.ApplyPortalConfig(context.Background(),
		portalDisplayName,
		envName,
		envType,
		internalRegionSdk.V3Portals,
		internalRegionSdk.API,
		labels)
	if err != nil {
		return "", fmt.Errorf("failed to apply portal configuration: %w", err)
	}
	return portalID, nil
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
	portalID string,
	labels map[string]string,
) error {
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
				return fmt.Errorf(
					"service %s referenced in team %s in organization "+
						"%s environment %s not found in team configuration",
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
				portalID,
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
				portalID,
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
			fmt.Sprintf("[Konnect Orchestrator] - Changes for [%s] environment", envName),
			fmt.Sprintf(
				"For the %s environment, Konnect Orchestrator has detected changes in upstream service repositories "+
					"and has generated the associated changes.", envName,
			),
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
	sdk *kk.SDK,
) error {
	fmt.Printf("Processing environment %s in organization %s\n", envName, orgName)

	labels := map[string]string{
		// 'konnect' is a reserved prefix for labels
		"ko-konnect-orchestrator": "true",
		"env-name":                envName,
		"env-type":                envConfig.Type,
	}

	portalID, err := applyPortal(
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
				portalID,
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
				portalID,
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
	teams map[string]*manifest.Team,
) error {
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

func runInit(_ *cobra.Command, _ []string) error {
	pc, err := loadPlatformManifest(platformFileArg)
	if err != nil {
		return err
	}
	return platform.Init(pc.Git, resourceFiles)
}

func runAddOrganization(_ *cobra.Command, _ []string) error {
	// TODO: Add an organization to the konnect/organizations.yaml file
	// Pre-requisites:
	// 		We need the platform gith configuration (url and token), either the platform file or args
	//		The repository must exist
	//		We prompt the user to provide the following:
	//			Organization name (we slugify for a YAML key)
	//			Konnect Token:
	// 1. Clone the repository locally
	// 2. Load and add the new organization to the organizations file
	// 3. Modify the .github/workflow/koctl-apply.yaml file to include
	//	    env:
	//			<ORGNAME>_KONNECT_TOKEN: {{ .secrets.<ORGNAME>_KONNECT_TOKEN }}
	// 4. Write the provided Konnect token to the repository secrets API
	// 5. File PR
	return nil
}

func loadPlatformManifest(path string) (*manifest.Platform, error) {
	var rv manifest.Orchestrator
	err := readConfigSection(path, &rv)
	return rv.Platform, err
}

func loadConfigManifest() (*manifest.Orchestrator, error) {
	var man manifest.Orchestrator
	var wholeFilePath, platformFilePath, teamsFilePath, organizationsFilePath string
	if wholeFileArg != "" {
		var err error
		wholeFilePath, err = filepath.Abs(wholeFileArg)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve whole file path: %w", err)
		}
		if _, err := os.Stat(wholeFilePath); err != nil {
			return nil, fmt.Errorf("failed to access file %s: %w", wholeFilePath, err)
		}
	} else {
		var err error
		platformFilePath, err = filepath.Abs(platformFileArg)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve platform file path: %w", err)
		}
		if _, err := os.Stat(platformFilePath); err != nil {
			return nil, fmt.Errorf("failed to access file %s: %w", platformFilePath, err)
		}

		if teamsFileArg != "" {
			teamsFilePath, err = filepath.Abs(teamsFileArg)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve teams file path: %w", err)
			}
			if _, err := os.Stat(teamsFilePath); err != nil {
				return nil, fmt.Errorf("failed to access file %s: %w", teamsFilePath, err)
			}
		}

		if organizationsFileArg != "" {
			organizationsFilePath, err = filepath.Abs(organizationsFileArg)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve organizations file path: %w", err)
			}
			if _, err := os.Stat(organizationsFilePath); err != nil {
				return nil, fmt.Errorf("failed to access file %s: %w", organizationsFilePath, err)
			}
		}
	}

	if wholeFilePath != "" {
		if err := readConfigSection(wholeFilePath, &man); err != nil {
			return nil, fmt.Errorf("failed to read whole configuration: %w", err)
		}
	} else {
		if err := readConfigSection(platformFilePath, &man); err != nil {
			return nil, fmt.Errorf("failed to read platform configuration: %w", err)
		}

		if teamsFilePath != "" {
			if err := readConfigSection(teamsFilePath, &man); err != nil {
				return nil, fmt.Errorf("failed to read teams configuration: %w", err)
			}
		} else {
			man.Teams = make(map[string]*manifest.Team)
		}

		if organizationsFilePath != "" {
			if err := readConfigSection(organizationsFilePath, &man); err != nil {
				return nil, fmt.Errorf("failed to read organizations configuration: %w", err)
			}
		} else {
			man.Organizations = make(map[string]*manifest.Organization)
		}
	}

	return &man, nil
}

func apply(man *manifest.Orchestrator) error {
	err := platform.Init(man.Platform.Git, resourceFiles)
	if err != nil {
		return fmt.Errorf("failed to apply platform repository changes: %w", err)
	}

	// Process each organization
	for orgName, orgConfig := range man.Organizations {
		if err := applyOrganization(orgName, *man.Platform.Git, *orgConfig, man.Teams); err != nil {
			return err
		}
	}

	fmt.Println("Configuration Applied")

	return nil
}

func runApply(_ *cobra.Command, _ []string) error {

	// We're not looping, run once and exit
	if loopInterval == 0 {
		man, err := loadConfigManifest()
		if err != nil {
			return err
		}
		return apply(man)
	}

	// Main processing loop
	for {
		man, err := loadConfigManifest()
		if err != nil {
			return err
		}
		err = apply(man)
		if err != nil {
			fmt.Printf("Error applying configuration: %v\n", err)
			return err
		}

		time.Sleep(time.Duration(loopInterval) * time.Second)
	}
}

func runRun(_ *cobra.Command, _ []string) error {
	platform, err := loadPlatformManifest(platformFileArg)
	if err != nil {
		return err
	}
	return server.RunServer(*platform.Git,
		defaultTeamsFilePath,
		defaultOrgsFilePath,
		version, commit, date)
}

func Execute() error {
	return rootCmd.Execute()
}

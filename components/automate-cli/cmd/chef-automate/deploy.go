// Copyright © 2017 Chef Software

package main

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	dc "github.com/chef/automate/api/config/deployment"
	api "github.com/chef/automate/api/interservice/deployment"
	"github.com/chef/automate/components/automate-cli/pkg/status"
	"github.com/chef/automate/components/automate-deployment/pkg/airgap"
	"github.com/chef/automate/components/automate-deployment/pkg/client"
	"github.com/chef/automate/components/automate-deployment/pkg/manifest"
	mc "github.com/chef/automate/components/automate-deployment/pkg/manifest/client"
	"github.com/chef/automate/lib/version"
)

var deployLong = `Deploy a new Chef Automate instance using the supplied configuration.
	- <CONFIG_FILE> must be a valid path to a TOML formatted configuration file`
var promptMLSA = `
To continue, you'll need to accept our terms of service:

Terms of Service
https://www.chef.io/terms-of-service

Master License and Services Agreement
https://www.chef.io/online-master-agreement

I agree to the Terms of Service and the Master License and Services Agreement
`
var errMLSA = "Chef Software Terms of Service and Master License and Services Agreement were not accepted"
var errProvisonInfra = `Architecture does not match with the requested one. 
If you want to provision cluster then you have to first run provision command.

		chef-automate provision-infra

After that you can run this command`

var invalidConfig = "Invalid toml config file, please check your toml file."

var deployCmdFlags = struct {
	channel                         string
	upgradeStrategy                 string
	keyPath                         string
	certPath                        string
	adminPassword                   string
	hartifactsPath                  string
	overrideOrigin                  string
	manifestDir                     string
	fqdn                            string
	airgap                          string
	skipPreflight                   bool
	acceptMLSA                      bool
	enableChefServer                bool
	enableDeploymentOrderStressMode bool
	enableWorkflow                  bool
	products                        []string
	bootstrapBundlePath             string
}{}

// deployCmd represents the new command
var deployCmd = newDeployCmd()

func init() {
	RootCmd.AddCommand(deployCmd)
}

func newDeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy [/path/to/config.toml]",
		Short: "Deploy Chef Automate",
		Long:  deployLong,
		Annotations: map[string]string{
			NoCheckVersionAnnotation: NoCheckVersionAnnotation,
		},
		Args: cobra.RangeArgs(0, 1),
		RunE: runDeployCmd,
	}

	// flags for Deploy Command
	cmd.PersistentFlags().BoolVar(
		&deployCmdFlags.skipPreflight,
		"skip-preflight",
		false,
		"Deploy regardless of pre-flight conditions")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.overrideOrigin,
		"override-origin",
		"",
		"Optional origin to install local .hart packages from")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.hartifactsPath,
		"hartifacts",
		"",
		"Optional path to cache of local .hart packages")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.manifestDir,
		"manifest-dir",
		"",
		"Optional path to local automate manifest files")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.channel,
		"channel",
		"",
		"Release channel to deploy all services from")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.upgradeStrategy,
		"upgrade-strategy",
		"at-once",
		"Upgrade strategy to use for this deployment.")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.certPath,
		"certificate",
		"",
		"The path to a certificate that should be used for external TLS connections (web and API).")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.keyPath,
		"private-key",
		"",
		"The path to a private key corresponding to the TLS certificate.")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.adminPassword,
		"admin-password",
		"",
		"The password for the initial admin user. Auto-generated by default.")
	cmd.PersistentFlags().BoolVar(
		&deployCmdFlags.acceptMLSA,
		"accept-terms-and-mlsa",
		false,
		"Agree to the Chef Software Terms of Service and the Master License and Services Agreement")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.fqdn,
		"fqdn",
		"",
		"The fully-qualified domain name that Chef Automate can be accessed at. (default: hostname of this machine)")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.airgap,
		"airgap-bundle",
		"",
		"Path to an airgap install bundle")
	cmd.PersistentFlags().BoolVar(
		&deployCmdFlags.enableChefServer,
		"enable-chef-server",
		false,
		"Deploy Chef Server services along with Chef Automate")
	cmd.PersistentFlags().BoolVar(
		&deployCmdFlags.enableDeploymentOrderStressMode,
		"enable-deploy-order-stress-mode",
		false,
		"Deploy services in the order that stresses hab the most")
	cmd.PersistentFlags().BoolVar(
		&deployCmdFlags.enableWorkflow,
		"enable-workflow",
		false,
		"Deploy Workflow services along with Chef Automate")
	cmd.PersistentFlags().StringSliceVar(
		&deployCmdFlags.products,
		"product",
		nil,
		"Product to deploy")
	cmd.PersistentFlags().StringVar(
		&deployCmdFlags.bootstrapBundlePath,
		"bootstrap-bundle",
		"",
		"Path to bootstrap bundle")

	if !isDevMode() {
		for _, flagName := range []string{
			"override-origin",
			"hartifacts",
			"manifest-dir",
			// passwords are not validated until the end of the deploy, which makes this
			// feature dangerous. But we still want to have it in Ci, so we mark it as
			// hidden
			"admin-password",
			"enable-chef-server",
			"enable-deploy-order-stress-mode",
			"enable-workflow",
			"bootstrap-bundle",
		} {
			err := cmd.PersistentFlags().MarkHidden(flagName)
			if err != nil {
				fmt.Printf("failed configuring cobra: %s\n", err.Error())
				panic(":(")
			}
		}
	}
	return cmd
}

func runDeployCmd(cmd *cobra.Command, args []string) error {
	var configPath = ""
	if len(args) > 0 {
		configPath = args[0]
	}
	var deployer, derr = getDeployer(configPath)
	if derr != nil {
		return status.Wrap(derr, status.ConfigError, invalidConfig)
	}
	if deployer != nil {
		return deployer.doDeployWork(args)
	}
	writer.Printf("Automate deployment non HA mode proceeding...")
	if !deployCmdFlags.acceptMLSA {
		agree, err := writer.Confirm(promptMLSA)
		if err != nil {
			return status.Wrap(err, status.InvalidCommandArgsError, errMLSA)
		}

		if !agree {
			return status.New(status.InvalidCommandArgsError, errMLSA)
		}
	}

	if deployCmdFlags.keyPath != "" && deployCmdFlags.certPath == "" {
		msg := "Cannot provide --private-key without also providing --certificate."
		return status.New(status.InvalidCommandArgsError, msg)
	}

	if deployCmdFlags.certPath != "" && deployCmdFlags.keyPath == "" {
		msg := "cannot provide --certificate without also providing --private-key."
		return status.New(status.InvalidCommandArgsError, msg)
	}

	conf := new(dc.AutomateConfig)
	var err error
	if len(args) == 0 {
		// Use default configuration if no configuration file was provided
		conf, err = generatedConfig()
		if err != nil {
			return status.Annotate(err, status.ConfigError)
		}
	} else {
		conf, err = dc.LoadUserOverrideConfigFile(args[0])
		if err != nil {
			return status.Wrapf(
				err,
				status.ConfigError,
				"Loading configuration file %s failed",
				args[0],
			)
		}
	}

	if err = mergeFlagOverrides(conf); err != nil {
		return status.Wrap(
			err,
			status.ConfigError,
			"Merging command flag overrides into Chef Automate config failed",
		)
	}

	adminPassword := deployCmdFlags.adminPassword
	if adminPassword == "" {
		adminPassword, err = dc.GeneratePassword()
		if err != nil {
			return status.Wrap(err, status.ConfigError, "Generating the admin user password failed")
		}
	}
	err = conf.AddCredentials("Local Administrator", "admin", adminPassword)
	if err != nil {
		return status.Wrap(err, status.ConfigError, "Applying the admin user password to configuration failed")
	}

	offlineMode := deployCmdFlags.airgap != ""
	manifestPath := ""
	if offlineMode {
		writer.Title("Installing artifact")
		metadata, err := airgap.Unpack(deployCmdFlags.airgap)
		if err != nil {
			return status.Annotate(err, status.AirgapUnpackInstallBundleError)
		}
		manifestPath = api.AirgapManifestPath

		// We need to set the path for the hab binary so that the deployer does not
		// try to go to the internet to get it
		pathEnv := os.Getenv("PATH")

		err = os.Setenv("PATH", fmt.Sprintf("%s:%s", path.Dir(metadata.HabBinPath), pathEnv))
		if err != nil {
			return err
		}
	} else {
		manifestPath = conf.Deployment.GetV1().GetSvc().GetManifestDirectory().GetValue()
	}

	manifestProvider := manifest.NewLocalHartManifestProvider(
		mc.NewDefaultClient(manifestPath),
		conf.Deployment.GetV1().GetSvc().GetHartifactsPath().GetValue(),
		conf.Deployment.GetV1().GetSvc().GetOverrideOrigin().GetValue())

	err = client.Deploy(writer, conf, deployCmdFlags.skipPreflight, manifestProvider, version.BuildTime, offlineMode, deployCmdFlags.bootstrapBundlePath)
	if err != nil && !status.IsStatusError(err) {
		return status.Annotate(err, status.DeployError)
	}
	return err
}

func generatedConfig() (*dc.AutomateConfig, error) {
	cfg, err := dc.GenerateInitConfig(
		deployCmdFlags.channel,
		deployCmdFlags.upgradeStrategy,
		dc.InitialTLSCerts(deployCmdFlags.keyPath, deployCmdFlags.certPath),
		dc.InitialFQDN(deployCmdFlags.fqdn),
	)
	if err != nil {
		return nil, status.Wrap(err, status.ConfigError, "Generating initial default configuration failed")
	}
	return cfg.AutomateConfig(), nil
}

// mergeFlagOverrides merges the flag provided configuration options into the
// user override config. Because the override configuration will be persisted
// we only want to add overrides for flags that have been specifically set so
// that we don't accidentally set an override with a default value.
func mergeFlagOverrides(conf *dc.AutomateConfig) error {
	overrideOpts := []dc.AutomateConfigOpt{}
	if deployCmdFlags.manifestDir != "" {
		overrideOpts = append(overrideOpts, dc.WithManifestDir(deployCmdFlags.manifestDir))
	}

	if deployCmdFlags.channel != "" {
		overrideOpts = append(overrideOpts, dc.WithChannel(deployCmdFlags.channel))
	}

	if deployCmdFlags.hartifactsPath != "" {
		overrideOpts = append(overrideOpts, dc.WithHartifacts(deployCmdFlags.hartifactsPath))
	}

	if deployCmdFlags.overrideOrigin != "" {
		overrideOpts = append(overrideOpts, dc.WithOrigin(deployCmdFlags.overrideOrigin))
	}

	if deployCmdFlags.enableChefServer {
		overrideOpts = append(overrideOpts, dc.WithChefServerEnabled(true))
	}

	if deployCmdFlags.enableDeploymentOrderStressMode {
		overrideOpts = append(overrideOpts, dc.WithDeploymentOrderStressMode(true))
	}

	if deployCmdFlags.enableWorkflow {
		overrideOpts = append(overrideOpts, dc.WithWorkflowEnabled(true))
	}

	if len(deployCmdFlags.products) > 0 {
		overrideOpts = append(overrideOpts, dc.WithProducts(deployCmdFlags.products))
	}

	return dc.WithConfigOptions(conf, overrideOpts...)
}

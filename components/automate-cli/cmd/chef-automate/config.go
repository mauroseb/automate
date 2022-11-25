package main

import (
	"io/ioutil"
	"reflect"
	"regexp"
	"strings"
	"time"

	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	dc "github.com/chef/automate/api/config/deployment"
	"github.com/chef/automate/components/automate-cli/pkg/status"
	"github.com/chef/automate/components/automate-deployment/pkg/client"
	"github.com/chef/toml"
	"github.com/imdario/mergo"
)

var configCmdFlags = struct {
	overwriteFile bool
	timeout       int64
	acceptMLSA    bool

	automate         bool
	chef_server      bool
	frontend         bool
	opensearch       bool
	postgresql       bool
	getAppliedConfig bool
	file             string
}{}

const (
	FRONTEND_COMMANDS = `
	sudo chef-automate config patch /tmp/%s;
	export TIMESTAMP=$(date +'%s');
	sudo mv /etc/chef-automate/config.toml /etc/chef-automate/config.toml.$TIMESTAMP;
	sudo chef-automate config show > sudo /etc/chef-automate/config.toml`

	BACKEND_COMMAND = `
	export TIMESTAMP=$(date +"%s");
	echo "yes" | sudo hab config apply automate-ha-%s.default  $(date '+%s') /tmp/%s;
	`

	GET_CONFIG = `
	source <(sudo cat /hab/sup/default/SystemdEnvironmentFile.sh);
	automate-backend-ctl show --svc=automate-ha-%s
	`

	GET_APPLIED_CONFIG = `
	source <(sudo cat /hab/sup/default/SystemdEnvironmentFile.sh);
	automate-backend-ctl applied --svc=automate-ha-%s
	`

	dateFormat = "%Y%m%d%H%M%S"

	postgresql       = "postgresql"
	opensearch_const = "opensearch"
)

var configValid = "Config file must be a valid %s config"

func init() {
	configCmd.AddCommand(showConfigCmd)
	configCmd.AddCommand(patchConfigCmd)
	configCmd.AddCommand(setConfigCmd)

	showConfigCmd.Flags().BoolVarP(&configCmdFlags.overwriteFile, "overwrite", "o", false, "Overwrite existing config.toml")
	patchConfigCmd.PersistentFlags().BoolVarP(&configCmdFlags.frontend, "frontend", "f", false, "Patch toml configuration to the all frontend nodes")
	patchConfigCmd.PersistentFlags().BoolVar(&configCmdFlags.frontend, "fe", false, "Patch toml configuration to the all frontend nodes[DUPLICATE]")

	patchConfigCmd.PersistentFlags().BoolVarP(&configCmdFlags.automate, "automate", "a", false, "Patch toml configuration to the automate node")
	patchConfigCmd.PersistentFlags().BoolVar(&configCmdFlags.automate, "a2", false, "Patch toml configuration to the automate node[DUPLICATE]")
	patchConfigCmd.PersistentFlags().BoolVarP(&configCmdFlags.chef_server, "chef_server", "c", false, "Patch toml configuration to the chef_server node")
	patchConfigCmd.PersistentFlags().BoolVar(&configCmdFlags.chef_server, "cs", false, "Patch toml configuration to the chef_server node[DUPLICATE]")
	patchConfigCmd.PersistentFlags().BoolVarP(&configCmdFlags.opensearch, "opensearch", "o", false, "Patch toml configuration to the opensearch node")
	patchConfigCmd.PersistentFlags().BoolVar(&configCmdFlags.opensearch, "os", false, "Patch toml configuration to the opensearch node[DUPLICATE]")
	patchConfigCmd.PersistentFlags().BoolVarP(&configCmdFlags.postgresql, "postgresql", "p", false, "Patch toml configuration to the postgresql node")
	patchConfigCmd.PersistentFlags().BoolVar(&configCmdFlags.postgresql, "pg", false, "Patch toml configuration to the postgresql node[DUPLICATE]")

	configCmd.PersistentFlags().BoolVarP(&configCmdFlags.acceptMLSA, "auto-approve", "y", false, "Do not prompt for confirmation; accept defaults and continue")
	configCmd.PersistentFlags().Int64VarP(&configCmdFlags.timeout, "timeout", "t", 10, "Request timeout in seconds")

	RootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config COMMAND",
	Short: "Chef Automate configuration",
}

var showConfigCmd = &cobra.Command{
	Use:   "show [/path/to/write/config.toml]",
	Short: "show the Chef Automate configuration",
	Long:  "Show the Chef Automate configuration. When given a filepath, the output will be written to the file instead of printed to STDOUT",
	RunE:  runShowCmd,
	Args:  cobra.RangeArgs(0, 1),
}

var patchConfigCmd = &cobra.Command{
	Use:   "patch path/to/config.toml",
	Short: "patch the Chef Automate configuration", Long: "Apply a partial Chef Automate configuration to the deployment. It will take the partial configuration, merge it with the existing configuration, and apply and required changes.",
	RunE: runPatchCommand,
	Args: cobra.ExactArgs(1),
}

var setConfigCmd = &cobra.Command{
	Use:   "set path/to/config.toml",
	Short: "set the Chef Automate configuration",
	Long:  "Set the Chef Automate configuration for the deployment. It will replace the Chef Automate configuration with the given configuration and apply any required changes.",
	RunE:  runSetCommand,
	Args:  cobra.ExactArgs(1),
}

func runShowCmd(cmd *cobra.Command, args []string) error {
	res, err := client.GetAutomateConfig(configCmdFlags.timeout)
	if err != nil {
		return err
	}

	// TODO: should this be done server side?
	res.Config.Redact()

	// Handle writing to a file if a path was given
	if len(args) > 0 && args[0] != "" {
		outFile, err := filepath.Abs(args[0])
		if err != nil {
			return status.Annotate(err, status.FileAccessError)
		}

		if _, err := os.Stat(outFile); err == nil {
			if !configCmdFlags.overwriteFile {
				ok, err := writer.Confirm(fmt.Sprintf("%s file already exists. Do you wish to overwrite it?", outFile))
				if err != nil || !ok {
					if !ok {
						err = errors.New("failed to confirm overwrite")
					}

					return status.Annotate(err, status.FileAccessError)
				}
			}
		}

		err = res.Config.MarshalToTOMLFile(outFile, 0644)
		if err != nil {
			return status.Wrap(
				err,
				status.MarshalError,
				"Marshaling configuration to config.toml failed",
			)
		}

		writer.Successf("Chef Automate configuration file written to: %s", outFile)
		return nil
	}

	t, err := res.Config.MarshalToTOML()
	if err != nil {
		return status.Wrap(
			err,
			status.MarshalError,
			"Marshaling configuration to TOML failed",
		)
	}
	status.GlobalResult = res.Config

	writer.Println(string(t))
	return nil
}
func runPatchCommand(cmd *cobra.Command, args []string) error {
	/*
		incase of a2ha mode of deployment, config file will be copied to /hab/a2_deploy_workspace/configs/automate.toml file
		then automate cluster ctl deploy will patch the config to automate
	*/

	if isA2HARBFileExist() {

		var err error
		infra, err := getAutomateHAInfraDetails()
		if err != nil {
			return err
		}

		sshUser := infra.Outputs.SSHUser.Value
		sskKeyFile := infra.Outputs.SSHKeyFile.Value
		sshPort := infra.Outputs.SSHPort.Value

		timestamp := time.Now().Format("20060102150405")
		sshConfig := &SSHConfig{
			sshUser:    sshUser,
			sshKeyFile: sskKeyFile,
			sshPort:    sshPort,
		}
		sshUtil := NewSSHUtil(sshConfig)
		if configCmdFlags.frontend {
			frontendIps := append(infra.Outputs.ChefServerPrivateIps.Value, infra.Outputs.AutomatePrivateIps.Value...)
			if len(frontendIps) == 0 {
				writer.Error("No frontend IPs are found")
				os.Exit(1)
			}
			const remoteService string = "frontend"
			err = setConfigForFrontEndNodes(args, sshUtil, frontendIps, remoteService, timestamp)
		} else if configCmdFlags.automate {
			frontendIps := infra.Outputs.AutomatePrivateIps.Value
			if len(frontendIps) == 0 {
				writer.Error("No automate IPs are found")
				os.Exit(1)
			}
			const remoteService string = "automate"
			err = setConfigForFrontEndNodes(args, sshUtil, frontendIps, remoteService, timestamp)
		} else if configCmdFlags.chef_server {
			frontendIps := infra.Outputs.ChefServerPrivateIps.Value
			if len(frontendIps) == 0 {
				writer.Error("No chef-server IPs are found")
				os.Exit(1)
			}
			const remoteService string = "chef-server"
			err = setConfigForFrontEndNodes(args, sshUtil, frontendIps, remoteService, timestamp)
		} else if configCmdFlags.postgresql {
			const remoteService string = "postgresql"
			err = setConfigForPostgresqlNodes(args, remoteService, sshUtil, infra, timestamp)
		} else if configCmdFlags.opensearch {
			const remoteService string = "opensearch"
			err = setConfigForOpensearch(args, remoteService, sshUtil, infra, timestamp)

		}
		if err != nil {
			writer.Errorf("%v", err)
			return err
		}

	} else {

		cfg, err := dc.LoadUserOverrideConfigFile(args[0])
		if err != nil {
			return status.Annotate(err, status.ConfigError)
		}
		if err = client.PatchAutomateConfig(configCmdFlags.timeout, cfg, writer); err != nil {
			return err
		}
	}

	// writer.Success("Configuration patched")
	return nil
}

// setConfigForFrontEndNodes patches the configuration for front end nodes in Automate HA
func setConfigForFrontEndNodes(args []string, sshUtil SSHUtil, frontendIps []string, remoteService string, timestamp string) error {
	scriptCommands := fmt.Sprintf(FRONTEND_COMMANDS, remoteService+timestamp, dateFormat)
	for i := 0; i < len(frontendIps); i++ {
		writer.Print("Connecting to the " + remoteService + " node : " + frontendIps[i])
		sshUtil.getSSHConfig().hostIP = frontendIps[i]
		err := sshUtil.copyFileToRemote(args[0], remoteService+timestamp, false)
		if err != nil {
			writer.Errorf("%v", err)
			return err
		}
		output, err := sshUtil.connectAndExecuteCommandOnRemote(scriptCommands, true)
		if err != nil {
			writer.Errorf("%v", err)
			return err
		}
		writer.Printf(output + "\n")
		writer.Success("Patching is completed on " + remoteService + " node : " + frontendIps[i] + "\n")
	}
	return nil
}

// setConfigForPostgresqlNodes patches the config for postgresql nodes in Automate HA
func setConfigForPostgresqlNodes(args []string, remoteService string, sshUtil SSHUtil, infra *AutomteHAInfraDetails, timestamp string) error {
	//checking for log configuration
	err := enableCentralizedLogConfigForHA(args, remoteService, sshUtil, infra.Outputs.PostgresqlPrivateIps.Value)
	if err != nil {
		return err
	}
	//checking database configuration
	existConfig, reqConfig, err := getExistingAndRequestedConfigForPostgres(args, infra, GET_APPLIED_CONFIG, sshUtil)
	if err != nil {
		return err
	}
	isConfigChangedDatabase := isConfigChanged(existConfig, reqConfig)
	//Implementing the config if there is some change in the database configuration
	if isConfigChangedDatabase {
		tomlFile := args[0] + timestamp
		tomlFilePath, err := createTomlFileFromConfig(&reqConfig, tomlFile)
		if err != nil {
			return err
		}
		scriptCommands := fmt.Sprintf(BACKEND_COMMAND, dateFormat, remoteService, "%s", remoteService+timestamp)
		if len(infra.Outputs.PostgresqlPrivateIps.Value) > 0 {
			sshUtil.getSSHConfig().hostIP = infra.Outputs.PostgresqlPrivateIps.Value[0]
			writer.Print("Connecting to the " + remoteService + " node : " + sshUtil.getSSHConfig().hostIP)
			err := sshUtil.copyFileToRemote(tomlFilePath, remoteService+timestamp, true)
			if err != nil {
				writer.Errorf("%v", err)
				return err
			}
			output, err := sshUtil.connectAndExecuteCommandOnRemote(scriptCommands, true)
			if err != nil {
				writer.Errorf("%v", err)
				return err
			}
			writer.Printf(output + "\n")
			writer.Success("Patching is completed on " + remoteService + " node : " + sshUtil.getSSHConfig().hostIP + "\n")
		}

	} else {
		writer.Println("There is no change in the configuration")
	}
	return nil
}

// setConfigForOpensearch patches the config for open-search nodes in Automate HA
func setConfigForOpensearch(args []string, remoteService string, sshUtil SSHUtil, infra *AutomteHAInfraDetails, timestamp string) error {
	//checking for log configuration
	err := enableCentralizedLogConfigForHA(args, remoteService, sshUtil, infra.Outputs.OpensearchPrivateIps.Value)
	if err != nil {
		return err
	}
	//checking database configuration
	existConfig, reqConfig, err := getExistingAndRequestedConfigForOpenSearch(args, infra, GET_APPLIED_CONFIG, sshUtil)
	if err != nil {
		return err
	}
	isConfigChangedDatabase := isConfigChanged(existConfig, reqConfig)
	//Implementing the config if there is some change in the database configuration
	if isConfigChangedDatabase {
		tomlFile := args[0] + timestamp
		// reqConfig.TLS = nil
		tomlFilePath, err := createTomlFileFromConfig(&reqConfig, tomlFile)

		if err != nil {
			return err
		}
		scriptCommands := fmt.Sprintf(BACKEND_COMMAND, dateFormat, remoteService, "%s", remoteService+timestamp)
		if len(infra.Outputs.OpensearchPrivateIps.Value) > 0 {
			sshUtil.getSSHConfig().hostIP = infra.Outputs.OpensearchPrivateIps.Value[0]
			writer.Print("Connecting to the " + remoteService + " node : " + sshUtil.getSSHConfig().hostIP)
			err := sshUtil.copyFileToRemote(tomlFilePath, remoteService+timestamp, true)
			if err != nil {
				writer.Errorf("%v", err)
				return err
			}
			output, err := sshUtil.connectAndExecuteCommandOnRemote(scriptCommands, true)
			if err != nil {
				writer.Errorf("%v", err)
				return err
			}
			writer.Printf(output + "\n")
			writer.Success("Patching is completed on " + remoteService + " node : " + sshUtil.getSSHConfig().hostIP + "\n")
		}

	} else {
		writer.Println("There is no change in the configuration")
	}
	return nil
}

func runSetCommand(cmd *cobra.Command, args []string) error {
	cfg, err := dc.LoadUserOverrideConfigFile(args[0])
	if err != nil {
		return status.Annotate(err, status.ConfigError)
	}

	if err := client.SetAutomateConfig(configCmdFlags.timeout, cfg, writer); err != nil {
		return err
	}

	writer.Success("Configuration set")
	return nil
}

func getMergedOpensearchInterface(rawOutput string, pemFilePath string, remoteService string) (interface{}, error) {

	var src OpensearchConfig
	if _, err := toml.Decode(cleanToml(rawOutput), &src); err != nil {
		return "", err
	}

	pemBytes, err := ioutil.ReadFile(pemFilePath) // nosemgrep
	if err != nil {
		return "", err
	}

	destString := string(pemBytes)
	var dest OpensearchConfig
	if _, err := toml.Decode(destString, &dest); err != nil {
		return "", errors.Errorf(configValid, remoteService)
	}

	mergo.Merge(&dest, src) //, mergo.WithOverride

	return dest, nil
}

func getMergedPostgresqlInterface(rawOutput string, pemFilePath string, remoteService string) (interface{}, error) {

	var src PostgresqlConfig
	if _, err := toml.Decode(cleanToml(rawOutput), &src); err != nil {
		return "", err
	}

	pemBytes, err := ioutil.ReadFile(pemFilePath) // nosemgrep
	if err != nil {
		return "", err
	}

	destString := string(pemBytes)
	var dest PostgresqlConfig
	if _, err := toml.Decode(destString, &dest); err != nil {
		return "", errors.Errorf(configValid, remoteService)
	}

	mergo.Merge(&dest, src) //, mergo.WithOverride

	return dest, nil
}

func getRemoteType(flag string, infra *AutomteHAInfraDetails) (string, string) {
	switch strings.ToLower(flag) {
	case "opensearch", "os", "o":
		return infra.Outputs.OpensearchPrivateIps.Value[0], "opensearch"
	case "postgresql", "pg", "p":
		return infra.Outputs.PostgresqlPrivateIps.Value[0], "postgresql"
	default:
		return "", ""
	}
}

func cleanToml(rawData string) string {
	re := regexp.MustCompile("(?im).*info:.*$")
	tomlOutput := re.ReplaceAllString(rawData, "")
	return tomlOutput
}

func getMergerTOMLPath(args []string, infra *AutomteHAInfraDetails, timestamp string, remoteType string, config string) (string, error) {
	sshconfig := &SSHConfig{}
	tomlFile := args[0] + timestamp
	sshconfig.sshUser = infra.Outputs.SSHUser.Value
	sshconfig.sshKeyFile = infra.Outputs.SSHKeyFile.Value
	sshconfig.sshPort = infra.Outputs.SSHPort.Value

	remoteIP, remoteService := getRemoteType(remoteType, infra)
	sshconfig.hostIP = remoteIP
	sshUtil := NewSSHUtil(sshconfig)
	scriptCommands := fmt.Sprintf(config, remoteService)
	rawOutput, err := sshUtil.connectAndExecuteCommandOnRemote(scriptCommands, true)
	if err != nil {
		return "", err
	}

	var (
		dest interface{}
		err1 error
	)
	if remoteService == "opensearch" {
		dest, err1 = getMergedOpensearchInterface(rawOutput, args[0], remoteService)
	} else {
		dest, err1 = getMergedPostgresqlInterface(rawOutput, args[0], remoteService)
	}
	if err1 != nil {
		return "", err1
	}

	f, err := os.Create(tomlFile)

	if err != nil {
		// failed to create/open the file
		writer.Bodyf("Failed to create/open the file, \n%v", err)
		return "", err
	}
	if err := toml.NewEncoder(f).Encode(dest); err != nil {
		// failed to encode
		writer.Bodyf("Failed to encode\n%v", err)
		return "", err
	}
	if err := f.Close(); err != nil {
		// failed to close the file
		writer.Bodyf("Failed to close the file\n%v", err)
		return "", err
	}

	return tomlFile, nil
}

// getConfigFromRemoteServer gets the config for remote server using the commands
func getConfigFromRemoteServer(infra *AutomteHAInfraDetails, remoteType string, config string, sshUtil SSHUtil) (string, error) {
	remoteIP, remoteService := getRemoteType(remoteType, infra)
	sshUtil.getSSHConfig().hostIP = remoteIP
	scriptCommands := fmt.Sprintf(config, remoteService)
	rawOutput, err := sshUtil.connectAndExecuteCommandOnRemote(scriptCommands, true)
	if err != nil {
		return "", err
	}
	return rawOutput, nil
}

// getDecodedConfig gets the decoded config from the input
func getDecodedConfig(input string, remoteService string) (interface{}, error) {
	if remoteService == postgresql {
		var src PostgresqlConfig
		if _, err := toml.Decode(cleanToml(input), &src); err != nil {
			return nil, err
		}
		return src, nil
	}
	if remoteService == opensearch_const {
		var src OpensearchConfig
		if _, err := toml.Decode(cleanToml(input), &src); err != nil {
			return nil, err
		}
		return src, nil
	}

	return nil, nil
}

// getConfigForArgsPostgresqlAndOpenSearch gets the requested config from the args provided for postgresql or opensearch
func getConfigForArgsPostgresqlOrOpenSearch(args []string, remoteService string) (interface{}, error) {
	pemBytes, err := ioutil.ReadFile(args[0]) // nosemgrep
	if err != nil {
		return nil, err
	}
	destString := string(pemBytes)
	if remoteService == "postgresql" {
		var dest PostgresqlConfig
		if _, err := toml.Decode(destString, &dest); err != nil {
			return dest, errors.Errorf(configValid, remoteService)
		}
		return dest, nil
	} else {
		var dest OpensearchConfig
		if _, err := toml.Decode(destString, &dest); err != nil {
			return dest, errors.Errorf(configValid, remoteService)
		}
		return dest, nil
	}

	return nil, nil

}

// isConfigChanged checks if configuration is changed
func isConfigChanged(src interface{}, dest interface{}) bool {
	if reflect.DeepEqual(src, dest) {
		return false
	}
	return true
}

// getExistingAndRequestedConfigForPostgres get requested and existing config for postgresql
func getExistingAndRequestedConfigForPostgres(args []string, infra *AutomteHAInfraDetails, config string, sshUtil SSHUtil) (PostgresqlConfig, PostgresqlConfig, error) {
	//Getting Existing config from server
	var existingConfig PostgresqlConfig
	var reqConfig PostgresqlConfig
	srcInputString, err := getConfigFromRemoteServer(infra, postgresql, config, sshUtil)
	if err != nil {
		return existingConfig, reqConfig, errors.Wrapf(err, "Unable to get config from the server with error")
	}
	existingConfigInterface, err := getDecodedConfig(srcInputString, postgresql)
	if err != nil {
		return existingConfig, reqConfig, err
	}
	existingConfig = existingConfigInterface.(PostgresqlConfig)

	//Getting Requested Config
	reqConfigInterface, err := getConfigForArgsPostgresqlOrOpenSearch(args, postgresql)
	if err != nil {
		return existingConfig, reqConfig, err
	}
	reqConfig = reqConfigInterface.(PostgresqlConfig)
	mergo.Merge(&reqConfig, existingConfig)
	return existingConfig, reqConfig, nil
}

// getExistingAndRequestedConfigForOpenSearch gets existed and requested config for opensearch
func getExistingAndRequestedConfigForOpenSearch(args []string, infra *AutomteHAInfraDetails, config string, sshUtil SSHUtil) (OpensearchConfig, OpensearchConfig, error) {
	//Getting Existing config from server
	var existingConfig OpensearchConfig
	var reqConfig OpensearchConfig
	srcInputString, err := getConfigFromRemoteServer(infra, opensearch_const, config, sshUtil)
	if err != nil {
		return existingConfig, reqConfig, errors.Wrapf(err, "Unable to get config from the server with error")
	}
	existingConfigInterface, err := getDecodedConfig(srcInputString, opensearch_const)
	if err != nil {
		return existingConfig, reqConfig, err
	}
	existingConfig = existingConfigInterface.(OpensearchConfig)

	//Getting Requested Config
	reqConfigInterface, err := getConfigForArgsPostgresqlOrOpenSearch(args, opensearch_const)
	if err != nil {
		return existingConfig, reqConfig, err
	}
	reqConfig = reqConfigInterface.(OpensearchConfig)

	mergo.Merge(&reqConfig, existingConfig)
	return existingConfig, reqConfig, nil
}

// createTomlFileFromConfig created a toml file where path and struct interface is provided
func createTomlFileFromConfig(config interface{}, tomlFile string) (string, error) {
	f, err := os.Create(tomlFile)

	if err != nil {
		// failed to create/open the file
		writer.Bodyf("Failed to create/open the file, \n%v", err)
		return "", err
	}
	if err := toml.NewEncoder(f).Encode(config); err != nil {
		// failed to encode
		writer.Bodyf("Failed to encode\n%v", err)
		return "", err
	}
	if err := f.Close(); err != nil {
		// failed to close the file
		writer.Bodyf("Failed to close the file\n%v", err)
		return "", err
	}

	return tomlFile, nil

}

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type RootConfig struct {
	baseURL      string
	clientID     string
	clientSecret string
	username     string
	password     string
	grantType    string
	apiVersion   string
}

const envPrefix = "SF_BULK_EXPORTER"

var (
	cfgFile string
	config  RootConfig
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "salesforce-bulk-exporter",
	Short: "CLI Utility for interacting with Salesforce bulk API",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.salesforce-bulk-exporter.yaml)")

	// required configuration
	rootCmd.PersistentFlags().StringVar(&config.baseURL, "base-url", "", "Salesforce base URL")
	rootCmd.PersistentFlags().StringVar(&config.clientID, "client-id", "", "Salesforce client ID")
	rootCmd.PersistentFlags().StringVar(&config.clientSecret, "client-secret", "", "Salesforce client secret")
	rootCmd.PersistentFlags().StringVar(&config.username, "username", "", "Salesforce username")
	rootCmd.PersistentFlags().StringVar(&config.password, "password", "", "Salesforce password")
	rootCmd.PersistentFlags().StringVar(&config.grantType, "grant-type", "password", "Salesforce grant type")
	rootCmd.PersistentFlags().StringVar(&config.apiVersion, "api-version", "55.0", "API version to use when interacting with Salesforce")

	viper.BindPFlag("base-url", rootCmd.PersistentFlags().Lookup("base-url"))
	viper.BindPFlag("client-id", rootCmd.PersistentFlags().Lookup("client-id"))
	viper.BindPFlag("client-secret", rootCmd.PersistentFlags().Lookup("client-secret"))
	viper.BindPFlag("username", rootCmd.PersistentFlags().Lookup("username"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("grant-type", rootCmd.PersistentFlags().Lookup("grant-type"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".salesforce-bulk-exporter" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".salesforce-bulk-exporter")
	}

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix(envPrefix)

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()

	bindFlags(rootCmd)

	validateRequiredConfig()
}

// Bind each cobra flag to its associated viper configuration (config file and environment variable)
// https://github.com/carolynvs/stingoftheviper
func bindFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		// Environment variables can't have dashes in them, so bind them to their equivalent
		// keys with underscores, e.g. --favorite-color to STING_FAVORITE_COLOR
		if strings.Contains(f.Name, "-") {
			envVar := fmt.Sprintf("%s_%s", envPrefix, strings.ToUpper(strings.ReplaceAll(f.Name, "-", "_")))
			viper.BindEnv(f.Name, envVar)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && viper.IsSet(f.Name) {
			val := viper.Get(f.Name)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}

// checks all configuration options and exits if any are empty
func validateRequiredConfig() {
	var missingConfiguration []string

	if config.baseURL == "" {
		missingConfiguration = append(missingConfiguration, "base-url")
	}
	if config.clientID == "" {
		missingConfiguration = append(missingConfiguration, "client-id")
	}
	if config.clientSecret == "" {
		missingConfiguration = append(missingConfiguration, "client-secret")
	}
	if config.username == "" {
		missingConfiguration = append(missingConfiguration, "username")
	}
	if config.password == "" {
		missingConfiguration = append(missingConfiguration, "password")
	}

	if len(missingConfiguration) != 0 {
		fmt.Printf("required configuration is missing: %s\n", strings.Join(missingConfiguration[:], ", "))
		os.Exit(1)
	}
}

package cmd

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vivekschauhan/correlation-service/pkg/client"
	"github.com/vivekschauhan/correlation-service/pkg/config"
	"github.com/vivekschauhan/correlation-service/pkg/server"
	"github.com/vivekschauhan/correlation-service/pkg/service"
	"github.com/vivekschauhan/correlation-service/pkg/util"
	"gopkg.in/yaml.v3"
)

var cfg = &config.Config{}

// NewRootCmd creates a new cobra.Command
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "correlation-service",
		Short:   "Server/client for testing correlation service",
		Version: "0.0.1",
		RunE:    run,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initViperConfig(cmd)
		},
	}

	initFlags(cmd)

	return cmd
}

func initFlags(cmd *cobra.Command) {
	cmd.Flags().String("mode", "server", "Mode: server/client (default: server)")
	cmd.Flags().Uint32("port", 9090, "port (default: 9090)")
	cmd.Flags().String("log_level", "info", "log level (default: info)")
	cmd.Flags().String("log_format", "line", "log format (default: line)")
	cmd.Flags().String("resource_mapping_file", "/data/resource_mapping.yaml", "Yaml file with resource mapping")
}

func initViperConfig(cmd *cobra.Command) error {
	v := viper.New()
	v.AutomaticEnv()
	bindFlagsToViperConfig(cmd, v)

	err := v.Unmarshal(cfg)
	if err != nil {
		return err
	}

	return nil
}

// bindFlagsToViperConfig - For each flag, look up its corresponding env var, and use the env var if the flag is not set.
func bindFlagsToViperConfig(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if err := v.BindPFlag(f.Name, f); err != nil {
			panic(err)
		}

		// Apply the viper config value to the flag when the flag is not set and viper has a value
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
			if err != nil {
				panic(err)
			}
		}
	})
}

func run(_ *cobra.Command, _ []string) error {
	log, err := util.GetLogger(cfg.Level, cfg.Format)

	if err != nil {
		return err
	}
	switch cfg.Mode {
	case "server":
		log.Info("starting correlation service")
		startServer(cfg, log)
	case "client":
		log.Info("starting correlation client")
		startClient(cfg, log)
	default:
		log.Error("unexpected mode")
	}
	return nil
}
func loadResourceMapping(cfg *config.Config) (*service.ResourceMappings, error) {
	buf, err := ioutil.ReadFile(cfg.ResourceMappingFile)
	if err != nil {
		return nil, err
	}
	type mappings struct {
		Resources []service.Resource `yaml:"mapping"`
	}

	rm := &mappings{}
	err = yaml.Unmarshal(buf, rm)
	if err != nil {
		return nil, err
	}

	resMappings := &service.ResourceMappings{
		Resources: make(map[string]service.Resource),
	}
	for _, res := range rm.Resources {
		resMappings.Resources[res.Path] = res
	}

	return resMappings, nil
}

func startServer(cfg *config.Config, log *logrus.Logger) {
	resMapping, err := loadResourceMapping(cfg)
	if err != nil {
		log.Fatal("unable to load resource mapping", err)
	}

	cs := service.NewCorrelationService(log, resMapping)

	s, err := server.NewServer(cfg, log, cs)
	if err != nil {
		log.Fatal("unable to start server", err)
	}
	s.Start()
}

func startClient(cfg *config.Config, log *logrus.Logger) {
	c, err := client.NewClient(context.Background(), cfg)
	if err != nil {
		log.Fatal("failed to initialize client", err)
	}
	resMapping, err := loadResourceMapping(cfg)
	if err != nil {
		log.Fatal("unable to load resource mapping", err)
	}
	for _, res := range resMapping.Resources {
		ret, err := c.GetResource(res.Path)
		if err != nil {
			log.Error("error in getting resource", err)
			continue
		}
		log.Printf("%+v", ret)
	}

}

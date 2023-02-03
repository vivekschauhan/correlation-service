package config

// Config the configuration for the KeyTool
type Config struct {
	Mode                string `mapstructure:"mode"`
	Host                string `mapstructure:"host"`
	Port                uint32 `mapstructure:"port"`
	Level               string `mapstructure:"log_level"`
	Format              string `mapstructure:"log_format"`
	ResourceMappingFile string `mapstructure:"resource_mapping_file"`
}

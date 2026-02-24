package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

var InstanceViper map[string]*ConfigHolder = make(map[string]*ConfigHolder)

type ConfigHolder struct {
	Engine       *viper.Viper
	KeyProcessed map[string]bool
}

type Configurable interface {
	SetDefaults() map[string]any
	SetEnvBindings() map[string]string
}

func LoadDefaultConfig[T Configurable](c T) error {
	return LoadConfig("", c, "config", "yaml", []string{})
}

func LoadDefaultConfigModule[T Configurable](moduleName string, c T) error {
	prefix := getKeyPrefix(moduleName, true)
	return LoadConfig(prefix, c, "config", "yaml", []string{})
}

func LoadConfigModule[T Configurable](moduleName string, c T, file string, ext string, path []string) error {
	prefix := getKeyPrefix(moduleName, true)
	return LoadConfig(prefix, c, file, ext, path)
}

func LoadConfig[T Configurable](prefix string, c T, file string, ext string, path []string) error {
	var holder *ConfigHolder

	replacer := strings.NewReplacer(".", "_")
	name := file + "." + ext

	if prefix != "" && !strings.HasSuffix(prefix, ".") {
		prefix += "."
	}

	if InstanceViper[name] == nil {
		v := viper.New()

		v.SetConfigName(file)
		v.SetConfigType(ext)
		if len(path) == 0 {
			v.AddConfigPath(".")
		} else {
			for _, p := range path {
				v.AddConfigPath(p)
			}
		}

		// Override with environment variables
		v.AutomaticEnv()

		if err := v.ReadInConfig(); err != nil {
			// If config file is not found, use defaults and environment variables
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return err
			}
		}

		// Replace dots with underscores for environment variable keys
		v.SetEnvKeyReplacer(replacer)
		holder = &ConfigHolder{
			Engine:       v,
			KeyProcessed: make(map[string]bool),
		}
		InstanceViper[name] = holder
	} else {
		holder = InstanceViper[name]
	}

	// Set defaults with priority to environment variables
	setPriorityDefaults(c, holder, replacer, prefix)

	if err := holder.Engine.Unmarshal(c); err != nil {
		return err
	}

	return nil
}

func getKeyPrefix(prefix string, ismodule bool) string {
	if prefix != "" {
		if ismodule {
			return "module." + prefix + "."
		} else {
			return prefix
		}
	}
	return ""
}

func setPriorityDefaults(c Configurable, holder *ConfigHolder, replacer *strings.Replacer, prefix string) {
	// var v *viper.Viper
	v := holder.Engine

	modPrefix := prefix

	// Force binding of specific environment variables
	bindings := c.SetEnvBindings()
	for runtimeKey, envKey := range bindings {
		v.BindEnv(runtimeKey, envKey)
	}

	defaults := c.SetDefaults()

	space := "      "
	text := fmt.Sprintf("Scan Values %s with prefix [%s]:\n", v.ConfigFileUsed(), prefix)
	for _, runtimeKey := range v.AllKeys() {
		if holder.KeyProcessed[runtimeKey] {
			// skip yang sudah diproses
			continue
		}

		runtimeValue := v.Get(runtimeKey)
		envFilekey := replacer.Replace(runtimeKey)

		cut := false
		runtimeKeyCut := runtimeKey
		if prefix != "" && strings.HasPrefix(runtimeKey, modPrefix) {
			runtimeKeyCut = strings.TrimPrefix(runtimeKey, modPrefix)
			cut = true
		}

		if runtimeKey != envFilekey {
			if runtimeValue == nil {
				envFileValue := v.Get(envFilekey)
				if envFileValue != nil {
					text += fmt.Sprintf("%s %s = %v -> [%s]\n", space, runtimeKey, envFileValue, envFilekey)
					v.SetDefault(runtimeKey, envFileValue)
					if cut {
						text += fmt.Sprintf("%s   ~ %s = %v -> [%s-CUT-PREFIX]\n", space, runtimeKeyCut, envFileValue, envFilekey)
						v.SetDefault(runtimeKeyCut, envFileValue)
					}
				} else if defValue, ok := defaults[runtimeKey]; ok {
					text += fmt.Sprintf("%s %s = %v -> [DEFAULTS]\n", space, runtimeKey, defValue)
					v.SetDefault(runtimeKey, defValue)
					if cut {
						text += fmt.Sprintf("%s   ~ %s = %v -> [DEFAULTS-CUT-PREFIX]\n", space, runtimeKeyCut, defValue)
						v.SetDefault(runtimeKeyCut, defValue)
					}
				}
			} else {
				text += fmt.Sprintf("%s %s = %v -> [RUNTIME]\n", space, runtimeKey, runtimeValue)
				if cut {
					text += fmt.Sprintf("%s   ~ %s = %v -> [RUNTIME-CUT-PREFIX]\n", space, runtimeKeyCut, runtimeValue)
					v.SetDefault(runtimeKeyCut, runtimeValue)
				}
			}
		}
	}

	// tandai yang sudah di-binding sebagai sudah diproses
	for runtimeKey := range bindings {
		holder.KeyProcessed[runtimeKey] = true
	}

	log.Println(text)
}

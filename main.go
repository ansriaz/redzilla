package main

// Project by Luca Capra (muka)
// URL: https://github.com/muka/redzilla
// Updated by Ans Riaz (ansriazch@gmail.com)

import (
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/ansriaz/redzilla/model"
	"github.com/ansriaz/redzilla/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/onrik/logrus/filename" // Add the file name and line number to the logging utility
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	Buildtimestamp string
	Githash        string
)

func init() {
	if Buildtimestamp != "" && Githash != "" {
		fmt.Printf("Build timestamp: %s Git Hash: %s\n", Buildtimestamp, Githash)
	}

	log.SetFormatter(&prefixed.TextFormatter{})
	// log.SetFormatter(&log.JSONFormatter{})
	filenameHook := filename.NewHook()
	filenameHook.Field = "filename" // Customize source field name
	log.AddHook(filenameHook)
}

func main() {

	viper.SetDefault("Network", "redzilla")
	viper.SetDefault("APIPort", ":3000")
	viper.SetDefault("Domain", "redzilla.localhost")
	viper.SetDefault("ImageName", "nodered/node-red-docker")
	viper.SetDefault("StorePath", "./data/store")
	viper.SetDefault("InstanceDataPath", "./data/instances")
	viper.SetDefault("InstanceConfigPath", "./data/config")
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("Autostart", false)
	viper.SetDefault("EnvPrefix", "")

	viper.SetDefault("AuthType", "none")
	viper.SetDefault("AuthHttpMethod", "GET")
	viper.SetDefault("AuthHttpUrl", "")
	viper.SetDefault("AuthHttpHeader", "Authorization")

	viper.SetEnvPrefix("redzilla")
	viper.AutomaticEnv()

	configFile := "./config.yml"
	if os.Getenv("REDZILLA_CONFIG") != "" {
		configFile = os.Getenv("REDZILLA_CONFIG")
	}

	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		viper.SetConfigFile(configFile)
		err := viper.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Failed to read from config file: %s", err))
		}
	}

	cfg := &model.Config{
		Network:            viper.GetString("Network"),
		APIPort:            viper.GetString("APIPort"),
		Domain:             viper.GetString("Domain"),
		ImageName:          viper.GetString("ImageName"),
		StorePath:          viper.GetString("StorePath"),
		InstanceDataPath:   viper.GetString("InstanceDataPath"),
		InstanceConfigPath: viper.GetString("InstanceConfigPath"),
		LogLevel:           viper.GetString("LogLevel"),
		Autostart:          viper.GetBool("Autostart"),
		EnvPrefix:          viper.GetString("EnvPrefix"),
		AuthType:           viper.GetString("AuthType"),
	}

	if strings.ToLower(cfg.AuthType) == "http" {

		a := new(model.AuthHttp)
		a.Method = viper.GetString("AuthHttpMethod")
		a.URL = viper.GetString("AuthHttpUrl")
		a.Header = viper.GetString("AuthHttpHeader")

		//setup the body template
		rawTpl := viper.GetString("AuthHttpHeader")
		if len(rawTpl) > 0 {
			bodyTemplate, err := template.New("").Parse(rawTpl)
			if err != nil {
				panic(fmt.Errorf("Failed to parse template: %s", err))
			}
			a.Body = bodyTemplate
		}

		cfg.AuthHttp = a
	}

	lvl, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		panic(fmt.Errorf("Failed to parse level %s: %s", cfg.LogLevel, err))
	}
	log.SetLevel(lvl)

	if lvl != log.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}

	log.Debugf("%++v", cfg)

	defer service.Stop(cfg)

	err = service.Start(cfg)
	if err != nil {
		log.Errorf("Error: %s", err.Error())
		panic(err)
	}

}

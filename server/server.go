package server

import (
	"github.com/microbay/microbay/plugin"
	"github.com/microbay/microbay/server/backends"
	//"github.com/fvbock/endless" ----> Hot reloads
	log "github.com/Sirupsen/logrus"
	"github.com/gocraft/web"
	"github.com/spf13/viper"
	"net/http"
)

var Config API

// Creates Root and resources routes and starts listening
func Start() {
	Config = LoadConfig()
	bootstrapLoadBalancer(Config.Resources)
	bootstrapPlugins(Config.Resources)
	rootRouter := web.New(Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Middleware((*Context).ConfigMiddleware).
		Middleware((*Context).RootMiddleware).
		Middleware((*Context).ResourceConfigMiddleware).
		Middleware((*Context).PluginMiddleware).
		Middleware((*Context).BalancedProxy)
	log.Info(Config.Name, " listening on ", viper.GetString("host"), " in ", viper.Get("env"), " mode")
	err := http.ListenAndServe(viper.GetString("host"), rootRouter)
	if err != nil {
		log.Fatal("Failed to start server ", err)
	}
}

func bootstrapPlugins(resources []*Resource) {
	for i := 0; i < len(resources); i++ {
		activePlugins := resources[i].Plugins
		plugins := make([]plugin.Plugin, 0)
		for j := 0; j < len(activePlugins); j++ {
			if p, err := plugin.Get(activePlugins[j]).Bootstrap(Config.plugins["redis-jwt"]); err != nil {
				log.Fatal(activePlugins[j], " plugin failed to bootstrap: ", err)
			} else {
				plugins = append(plugins, p)
			}
		}
		resources[i].Middleware = plugins
	}
}

// Creates linked list (golang Ring) from weighted micros array per resource
func bootstrapLoadBalancer(resources []*Resource) {
	for i := 0; i < len(resources); i++ {
		micros := resources[i].Micros
		flattenedMicros := make([]string, 0)
		for j := 0; j < len(micros); j++ {
			for n := 0; n < micros[j].Weight; n++ {
				flattenedMicros = append(flattenedMicros, micros[j].URL)
			}
		}
		resources[i].Backends = backends.Build("round-robin", flattenedMicros)
	}
}

package loader

import (
	"fmt"
	"os"

	"github.com/ipfs/go-ipfs/plugin"

	logging "gx/ipfs/QmRb5jh8z2E8hMGN2tkvs1yHynUanqnZ3UeKwgN1i9P1F8/go-log"
)

var log = logging.Logger("plugin/loader")

var loadPluginsFunc = func(string) ([]plugin.Plugin, error) {
	return nil, nil
}

// PluginLoader keeps track of loaded plugins
type PluginLoader struct {
	plugins []plugin.Plugin
}

func loadDynamicPlugins(pluginDir string) ([]plugin.Plugin, error) {
	_, err := os.Stat(pluginDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return loadPluginsFunc(pluginDir)
}

// NewPluginLoader creates new plugin loader
func NewPluginLoader(pluginDir string) (*PluginLoader, error) {
	plMap := make(map[string]plugin.Plugin)
	for _, v := range preloadPlugins {
		plMap[v.Name()] = v
	}

	newPls, err := loadDynamicPlugins(pluginDir)
	if err != nil {
		return nil, err
	}

	for _, pl := range newPls {
		if ppl, ok := plMap[pl.Name()]; ok {
			// plugin is already preloaded
			return nil, fmt.Errorf(
				"plugin: %s, is duplicated in version: %s, "+
					"while trying to load dynamically: %s",
				ppl.Name(), ppl.Version(), pl.Version())
		}
		plMap[pl.Name()] = pl
	}

	loader := &PluginLoader{plugins: make([]plugin.Plugin, 0, len(plMap))}

	for _, v := range plMap {
		loader.plugins = append(loader.plugins, v)
	}

	return loader, nil
}

// Initialize all the loaded plugins
func (loader *PluginLoader) Initialize() error {
	return initialize(loader.plugins)
}

// Run all the loaded plugins
func (loader *PluginLoader) Run() error {
	return run(loader.plugins)
}

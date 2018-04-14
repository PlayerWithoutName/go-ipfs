package loader

import (
	"fmt"
	"github.com/ipfs/go-ipfs/core/coredag"
	"github.com/ipfs/go-ipfs/plugin"
	"os"

	logging "gx/ipfs/QmRb5jh8z2E8hMGN2tkvs1yHynUanqnZ3UeKwgN1i9P1F8/go-log"
	opentracing "gx/ipfs/QmWLWmRVSiagqP15jczsGME1qpob6HDbtbHAY2he9W5iUo/opentracing-go"
	ipld "gx/ipfs/Qme5bWv7wtjUNGsK2BNGVUFPKiuxWrsqrtvYwCLRw8YFES/go-ipld-format"
)

var log = logging.Logger("plugin/loader")

var loadPluginsFunc = func(string) ([]plugin.Plugin, error) {
	return nil, nil
}

// PluginLoader keeps track of loaded plugins
type PluginLoader struct {
	plugins []plugin.Plugin
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

//Initialize all loaded plugins
func (loader *PluginLoader) Initialize() error {
	for _, p := range loader.plugins {
		err := p.Init()
		if err != nil {
			return err
		}
	}

	return nil
}

//Run the plugins
func (loader *PluginLoader) Run() error {
	for _, pl := range loader.plugins {
		switch pl := pl.(type) {
		case plugin.PluginIPLD:
			err := runIPLDPlugin(pl)
			if err != nil {
				return err
			}
		case plugin.PluginTracer:
			err := runTracerPlugin(pl)
			if err != nil {
				return err
			}
		default:
			panic(pl)
		}
	}
	return nil
}

func runIPLDPlugin(pl plugin.PluginIPLD) error {
	err := pl.RegisterBlockDecoders(ipld.DefaultBlockDecoder)
	if err != nil {
		return err
	}
	return pl.RegisterInputEncParsers(coredag.DefaultInputEncParsers)
}

func runTracerPlugin(pl plugin.PluginTracer) error {
	tracer, err := pl.InitTracer()
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)
	return nil
}

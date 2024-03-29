package modules

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"

	"kusionstack.io/kusion/pkg/log"
	"kusionstack.io/kusion/pkg/util/kfile"
)

const (
	DefaultModulePathEnv     = "KUSION_MODULE_PATH"
	KusionModuleBinaryPrefix = "kusion-module-"
	Dir                      = "modules"
)

var mu sync.Mutex

// HandshakeConfig is a common handshake that is shared by plugin and host.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "MODULE_PLUGIN",
	MagicCookieValue: "ON",
}

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	PluginKey: &GRPCPlugin{},
}

type Plugin struct {
	// key represents the module key, it consists of two parts: namespace/moduleName@version. e.g. "kusionstack/mysql@v0.1.0"
	key    string
	client *plugin.Client
	// Module represents the real module impl
	Module Module
}

func NewPlugin(key string) (*Plugin, error) {
	if key == "" {
		return nil, fmt.Errorf("module key can not be empty")
	}
	p := &Plugin{key: key}
	err := p.initModule()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Plugin) initModule() error {
	key := p.key
	split := strings.Split(key, "@")
	msg := "init module failed. Invalid plugin module key: %s. " +
		"The correct format for a key should be as follows: namespace/moduleName@version. e.g. kusionstack/mysql@v0.1.0"
	if len(split) != 2 {
		return fmt.Errorf(msg, key)
	}
	prefix := strings.Split(split[0], "/")
	if len(prefix) != 2 {
		return fmt.Errorf(msg, key)
	}

	// build the plugin client
	pluginPath, err := buildPluginPath(prefix[0], prefix[1], split[1])
	if err != nil {
		return err
	}
	pluginName := prefix[0] + "-" + prefix[1]
	client, err := newPluginClient(pluginPath, pluginName)
	if err != nil {
		return err
	}
	p.client = client
	rpcClient, err := client.Client()
	if err != nil {
		return fmt.Errorf("init kusion module plugin: %s failed. %w", key, err)
	}

	// dispense the plugin to get the real module
	raw, err := rpcClient.Dispense(PluginKey)
	if err != nil {
		return err
	}
	p.Module = raw.(Module)

	return nil
}

func buildPluginPath(namespace, resourceType, version string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	// validate the module path
	prefixPath, err := PluginDir()
	if err != nil {
		return "", err
	}
	goOs := runtime.GOOS
	goArch := runtime.GOARCH
	name := resourceType + "_" + version
	p := path.Join(prefixPath, namespace, resourceType, version, goOs, goArch, KusionModuleBinaryPrefix+name)
	_, err = os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("module dir doesn't exist. %s", p)
		} else {
			return "", err
		}
	}
	return p, nil
}

func newPluginClient(modulePluginPath, moduleName string) (*plugin.Client, error) {
	// create the plugin log file
	var logFilePath string
	dir, err := kfile.KusionDataFolder()
	if err != nil {
		return nil, err
	}
	logDir := path.Join(dir, log.Folder, Dir)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
			return nil, fmt.Errorf("failed to create module log dir: %w", err)
		}
	}
	logFilePath = path.Join(logDir, moduleName+".log")
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create module log file: %w", err)
	}

	// write log to a separate file
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   moduleName,
		Output: logFile,
		Level:  hclog.Debug,
	})

	// We're a host! Start by launching the plugin process.Need to defer kill
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         PluginMap,
		Cmd:             exec.Command(modulePluginPath),
		AllowedProtocols: []plugin.Protocol{
			plugin.ProtocolGRPC,
		},
		Logger: logger,
	})
	return client, nil
}

func (p *Plugin) KillPluginClient() error {
	if p.client == nil {
		return fmt.Errorf("plugin: %s client is nil", p.key)
	}
	p.client.Kill()
	return nil
}

func PluginDir() (string, error) {
	if env, found := os.LookupEnv(DefaultModulePathEnv); found {
		return env, nil
	} else if dir, err := kfile.KusionDataFolder(); err == nil {
		return path.Join(dir, Dir), nil
	} else {
		return "", err
	}
}

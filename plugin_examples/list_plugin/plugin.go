package main

import (
	"crypto/tls"
	"net/http"
	"os"
	"regexp"
	"time"

	bhttp "github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix/http"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix/terminal"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/bluemix/trace"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/common/rest"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/plugin"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/plugin_examples/list_plugin/api"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/plugin_examples/list_plugin/commands"
	"github.com/IBM-Bluemix/bluemix-cli-sdk/plugin_examples/list_plugin/i18n"
)

type ListPlugin struct{}

func (p *ListPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "bx-list",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 1,
		},
		Commands: []plugin.Command{
			{
				Name:        "list",
				Description: "List your apps, containers and services in the target space",
				Usage:       "bx list",
			},
		},
	}
}

func (p *ListPlugin) Run(context plugin.PluginContext, args []string) {
	p.init(context)

	ui := terminal.NewStdUI()
	client := &rest.Client{
		HTTPClient:    NewHTTPClient(context),
		DefaultHeader: DefaultHeader(context),
	}

	var err error
	switch args[0] {
	case "list":
		err = commands.NewList(ui,
			context,
			api.NewCCClient(context.APIEndpoint(), client),
			api.NewContainerClient(containerEndpoint(context), client),
		).Run(args[1:])
	}

	if err != nil {
		ui.Failed("%v\n", err)
		os.Exit(1)
	}
}

func (p *ListPlugin) init(context plugin.PluginContext) {
	i18n.T = i18n.Init(context)

	trace.Logger = trace.NewLogger(context.Trace())

	terminal.UserAskedForColors = context.ColorEnabled()
	terminal.InitColorSupport()
}

func NewHTTPClient(context plugin.PluginContext) *http.Client {
	transport := bhttp.NewTraceLoggingTransport(
		&http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: context.IsSSLDisabled(),
			},
		})

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(context.HTTPTimeout()) * time.Second,
	}
}

func DefaultHeader(context plugin.PluginContext) http.Header {
	context.RefreshUAAToken()

	h := http.Header{}
	h.Add("Authorization", context.UAAToken())
	return h
}

func containerEndpoint(context plugin.PluginContext) string {
	return regexp.MustCompile(`(^https?://)?[^\.]+(\..+)+`).ReplaceAllString(context.APIEndpoint(), "${1}containers-api${2}")
}

func main() {
	plugin.Start(new(ListPlugin))
}

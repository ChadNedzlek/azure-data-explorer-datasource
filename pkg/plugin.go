package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"

	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"

	"github.com/grafana/grafana_plugin_model/go/datasource"
	plugin "github.com/hashicorp/go-plugin"
)

var pluginLogger = hclog.New(&hclog.LoggerOptions{
	Name:  "azuredx-datasource",
	Level: hclog.LevelFromString("DEBUG"),
})

func healthcheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

func registerPProfHandlers(r *http.ServeMux) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// pluginGRPCServer provides a default GRPC server with message sizes increased from 4MB to 16MB
func pluginGRPCServer(opts []grpc.ServerOption) *grpc.Server {
	sopts := []grpc.ServerOption{}
	return grpc.NewServer(sopts...)
}

func main() {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", healthcheckHandler)
	registerPProfHandlers(m)
	go func() {
		if err := http.ListenAndServe(":6060", m); err != nil {
			log.Fatal(err)
		}
	}()
	// background pruning of column cache
	go func() {

	}()

	// log.SetOutput(os.Stderr) // the plugin sends logs to the host process on strErr
	pluginLogger.Debug("Running GRPC server")
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: map[string]plugin.Plugin{
			"grafana-azure-data-explorer-datasource": &datasource.DatasourcePluginImpl{Plugin: &GrafanaAzureDXDatasource{logger: pluginLogger}},
		},
		GRPCServer: pluginGRPCServer,
	})
}

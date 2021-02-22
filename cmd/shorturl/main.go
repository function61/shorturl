package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/function61/gokit/app/aws/lambdautils"
	"github.com/function61/gokit/app/dynversion"
	"github.com/function61/gokit/log/logex"
	"github.com/function61/gokit/net/http/httputils"
	"github.com/function61/gokit/os/osutil"
	"github.com/spf13/cobra"
)

func main() {
	if lambdautils.InLambda() {
		lambda.StartHandler(lambdautils.NewLambdaHttpHandlerAdapter(newServerHandler()))
		return
	}

	app := &cobra.Command{
		Use:     os.Args[0],
		Short:   "URL shortening service",
		Version: dynversion.Version,
	}

	app.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "Start server",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			logger := logex.StandardLogger()

			osutil.ExitIfError(runServer(
				osutil.CancelOnInterruptOrTerminate(logger),
				logger))
		},
	})

	osutil.ExitIfError(app.Execute())
}

func runServer(ctx context.Context, logger *log.Logger) error {
	srv := &http.Server{
		Addr:    ":80",
		Handler: newServerHandler(),
	}

	return httputils.CancelableServer(ctx, srv, func() error { return srv.ListenAndServe() })
}

func newServerHandlerWithDb(db map[string]string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/go/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/go/") // "/go/foobar" => "foobar"

		target, found := db[id]
		if !found {
			http.NotFound(w, r)
			return
		}

		http.Redirect(w, r, target, http.StatusFound)
	})

	return mux
}

func newServerHandler() http.Handler {
	return newServerHandlerWithDb(linkdb)
}

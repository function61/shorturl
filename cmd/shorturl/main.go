package main

import (
	"context"
	"net/http"
	"os"

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
				osutil.CancelOnInterruptOrTerminate(logger)))
		},
	})

	osutil.ExitIfError(app.Execute())
}

func runServer(ctx context.Context) error {
	srv := &http.Server{
		Addr:    ":80",
		Handler: newServerHandler(),

		ReadHeaderTimeout: httputils.DefaultReadHeaderTimeout,
	}

	return httputils.CancelableServer(ctx, srv, srv.ListenAndServe)
}

func newServerHandlerWithDb(db map[string]string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/go/{linkid...}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("linkid") // "/go/foobar" => "foobar"

		target, found := db[id]
		if !found {
			http.NotFound(w, r)
		} else {
			http.Redirect(w, r, target, http.StatusFound)
		}
	})

	return mux
}

func newServerHandler() http.Handler {
	return newServerHandlerWithDb(linkdb)
}

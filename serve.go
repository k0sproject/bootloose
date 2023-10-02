// SPDX-FileCopyrightText: 2019 Weaveworks Ltd.
// SPDX-FileCopyrightText: 2023 bootloose authors
// SPDX-License-Identifier: Apache-2.0

package bootloose

import (
	"fmt"
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/k0sproject/bootloose/pkg/api"
	"github.com/k0sproject/bootloose/pkg/cluster"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type serveOptions struct {
	listen       string
	keyStorePath string
	debug        bool
}

// defaultKeyStore is the path where to store the public keys.
const defaultKeyStorePath = "keys"
const defaultListenPort = 2444

func NewServeCommand() *cobra.Command {
	opts := &serveOptions{}
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Launch a bootloose server",
		RunE:  opts.serve,
	}
	cmd.Flags().StringVarP(&opts.listen, "listen", "l", fmt.Sprintf(":%d", defaultListenPort), "Listen address")
	cmd.Flags().StringVar(&opts.keyStorePath, "keystore-path", defaultKeyStorePath, "Path of the public keys store")
	cmd.Flags().BoolVar(&opts.debug, "debug", false, "Enable debug")

	return cmd
}

func baseURI(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", err
	}
	if host == "" || host == "0.0.0.0" || host == "[::]" {
		host = "localhost"
	}
	return fmt.Sprintf("http://%s:%s", host, port), nil
}

func (opts *serveOptions) serve(cmd *cobra.Command, args []string) error {
	baseURI, err := baseURI(opts.listen)
	if err != nil {
		return errors.Wrapf(err, "invalid listen address '%s'", opts.listen)
	}

	log.Infof("Starting server on: %s\n", opts.listen)

	keyStore := cluster.NewKeyStore(opts.keyStorePath)
	if err := keyStore.Init(); err != nil {
		return errors.Wrapf(err, "could not init keystore")
	}

	log.Infof("Key store successfully initialized in path: %s\n", opts.keyStorePath)

	api := api.New(baseURI, keyStore, opts.debug)
	router := api.Router()

	err = http.ListenAndServe(opts.listen, router)
	if err != nil {
		log.Fatalf("Unable to start server: %s", err)
	}

	return nil
}


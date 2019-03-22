// Copyright 2019 tawalaya
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/vulcand/oxy/forward"
	"golang.org/x/crypto/acme/autocert"
)

var logger = logrus.New()
var log = logrus.NewEntry(logger)

func cliSetup() {
	viper.SetDefault("port", 8000)
	viper.SetDefault("author", "admin@localhost")

	flag.Int("port", 8000, "set the port to listen to")
	flag.String("author", "admin@localhost", "the mail used by acme to register this service")

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

}

func main() {
	cliSetup()
	pxy := SSLProxy{port: viper.GetInt("port")}
	pxy.setupServer()
}

type SSLProxy struct {
	address *url.URL
	port    int
	oxy     *forward.Forwarder
}

func (p *SSLProxy) setupServer() {

	//initialize proxy
	oxy, err := forward.New(
		forward.Stream(true),
		forward.PassHostHeader(true),
	)

	if err != nil {
		log.Fatalf("failed to create proxy %+v", err)
	}
	p.oxy = oxy

	m := &autocert.Manager{
		Email:  viper.GetString("author"),
		Prompt: autocert.AcceptTOS,
		HostPolicy: func(ctx context.Context, host string) error {
			//TODO: add sensible host model
			return nil
		},
		Cache: autocert.DirCache(".certs"),
	}

	httpsServer := &http.Server{
		Addr:      ":443",
		Handler:   http.HandlerFunc(p.serve),
		TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
	}

	p.address, err = url.Parse(fmt.Sprintf("http://localhost:%d", p.port))

	err = httpsServer.ListenAndServeTLS("", "")
	if err != nil {
		log.Errorf("httpsSrv.ListendAndServeTLS() failed with %s", err)
	}

}

func (p *SSLProxy) serve(w http.ResponseWriter, req *http.Request) {
	req.URL = p.address
	p.oxy.ServeHTTP(w, req)
}

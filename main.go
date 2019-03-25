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
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/kabukky/httpscerts"
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
	viper.SetDefault("self", true)
	viper.SetDefault("certs", ".")

	flag.Int("port", 8000, "set the port to listen to")
	flag.String("author", "admin@localhost", "the mail used by acme to register this service")
	flag.String("domain", "localhost", "the domain to use for acme")
	flag.Bool("self", false, "use self signed instead")
	flag.String("certs", ".", "location of the certificates")
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
	p.address, err = url.Parse(fmt.Sprintf("http://localhost:%d", p.port))

	if viper.GetBool("self") {
		cert := filepath.Join(viper.GetString("certs"), "cert.pem")
		key := filepath.Join(viper.GetString("certs"), "key.pem")

		err := httpscerts.Check(cert, key)
		if err != nil {
			log.Info("could not load self signed keys - generationg some")
			err = httpscerts.Generate(cert, key, "127.0.0.1:443")
			if err != nil {
				log.Fatal("Error: Couldn't create https certs.")
			}
		}
		httpsServer := &http.Server{
			Addr:    ":443",
			Handler: http.HandlerFunc(p.serve),
		}

		err = httpsServer.ListenAndServeTLS(cert, key)
		if err != nil {
			log.Errorf("httpsSrv.ListendAndServeTLS() failed with %s", err)
		}

	} else {
		m := &autocert.Manager{
			Email:      viper.GetString("author"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(viper.GetStringSlice("domains")...),
			Cache:      autocert.DirCache("."),
		}

		httpsServer := &http.Server{
			Addr:      ":443",
			Handler:   http.HandlerFunc(p.serve),
			TLSConfig: &tls.Config{GetCertificate: m.GetCertificate},
		}

		err = httpsServer.ListenAndServeTLS("", "")
		if err != nil {
			log.Errorf("httpsSrv.ListendAndServeTLS() failed with %s", err)
		}
	}

}

func (p *SSLProxy) serve(w http.ResponseWriter, req *http.Request) {
	req.URL = p.address
	p.oxy.ServeHTTP(w, req)
}

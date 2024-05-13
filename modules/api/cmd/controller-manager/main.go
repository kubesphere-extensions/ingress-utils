/*
Copyright 2024 KubeSphere Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"

	certutil "k8s.io/client-go/util/cert"

	options "github.com/kubesphere-extensions/ingress-utils/cmd/controller-manager/options"
	"github.com/kubesphere-extensions/ingress-utils/pkg/controller"
	"github.com/kubesphere-extensions/ingress-utils/pkg/scheme"
	"k8s.io/client-go/util/keyutil"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	setupLog       = ctrl.Log.WithName("setup")
	defaultCertDir = filepath.Join(os.TempDir(), "k8s-webhook-server", "serving-certs")
)

const (
	// The server key and certificate must be named tls.key and tls.crt, respectively.
	defaultCertName = "tls.crt"
	defaultKeyName  = "tls.key"
)

func main() {
	zapOpts := zap.Options{
		Development: true,
	}
	zapOpts.BindFlags(flag.CommandLine)
	serverOpts := options.NewOptions()
	serverOpts.BindFlags(flag.CommandLine)
	klog.InitFlags(flag.CommandLine)
	flag.Parse()
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))

	if !certsExists() {
		if err := generateCerts(); err != nil {
			setupLog.Error(err, "unable to generate certs")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme.Scheme,
		Metrics:                metricsserver.Options{BindAddress: serverOpts.MetricsAddr},
		WebhookServer:          webhook.NewServer(webhook.Options{Port: serverOpts.Port}),
		HealthProbeBindAddress: serverOpts.ProbeAddr,
		LeaderElection:         serverOpts.LeaderElection,
		LeaderElectionID:       serverOpts.LeaderElectionID,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controller.IngressWebhook{}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ingress")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func certsExists() bool {
	certFile := path.Join(defaultCertDir, defaultCertName)
	keyFile := path.Join(defaultCertDir, defaultKeyName)
	_, err1 := os.Stat(certFile)
	_, err2 := os.Stat(keyFile)
	return err1 == nil && err2 == nil
}

func generateCerts() error {
	certFile := path.Join(defaultCertDir, defaultCertName)
	keyFile := path.Join(defaultCertDir, defaultKeyName)
	cert, key, err := certutil.GenerateSelfSignedCertKeyWithFixtures("localhost", []net.IP{net.ParseIP("127.0.0.1")}, []string{}, "")
	if err != nil {
		return fmt.Errorf("unable to generate self signed cert: %s", err)
	}
	if err := certutil.WriteCert(certFile, cert); err != nil {
		return fmt.Errorf("unable to write self signed cert: %s", err)
	}
	if err := keyutil.WriteKey(keyFile, key); err != nil {
		return fmt.Errorf("unable to write self signed cert: %s", err)
	}
	klog.Infof("Generated self-signed cert (%s, %s)", certFile, keyFile)
	return nil
}

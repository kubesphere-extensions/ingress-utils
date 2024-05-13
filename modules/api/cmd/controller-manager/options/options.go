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

package options

import (
	"flag"
)

type Options struct {
	Port             int
	MetricsAddr      string
	ProbeAddr        string
	LeaderElectionID string
	LeaderElection   bool
}

func NewOptions() *Options {
	return &Options{}
}

// BindFlags will parse the given flagset for server option flags and set the options accordingly
func (o *Options) BindFlags(fs *flag.FlagSet) {
	fs.IntVar(&o.Port, "server-port", 8443, "The server bind port.")
	fs.StringVar(&o.MetricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	fs.StringVar(&o.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.StringVar(&o.LeaderElectionID, "leader-election-id", "controller-manager", "LeaderElectionID determines the name of the resource that leader election will use for holding the leader lock")
	fs.BoolVar(&o.LeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
}

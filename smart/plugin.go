/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

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

package smart

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
)

const (
	PluginName = "smart-disk"
	version    = 7
	pluginType = plugin.CollectorPluginType

	nsVendor = "intel"
	nsClass  = "disk"
	nsType   = "smart"
)

var (
	//procPath source of data for metrics
	procPath = "/proc"
	//devPath source of data for metrics
	devPath = "/dev"

	namespace_prefix = []string{nsVendor, nsClass}
	namespace_suffix = []string{nsType}

	sysUtilProvider SysutilProvider
)

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(
		PluginName,
		version,
		pluginType,
		[]string{plugin.SnapGOBContentType},
		[]string{plugin.SnapGOBContentType},
		plugin.ConcurrencyCount(1),
	)
}

func NewSmartCollector() *SmartCollector {
	logger := log.New()
	return &SmartCollector{
		logger:    logger,
		proc_path: procPath,
		dev_path:  devPath,
	}
}

func makeName(device, metric string) []string {
	splited := strings.Split(metric, "/")

	name := []string{}
	name = append(name, namespace_prefix...)
	name = append(name, device)
	name = append(name, namespace_suffix...)
	name = append(name, splited...)

	return name
}

func parseName(namespace []string) (disk, attribute string) {
	disk = namespace[len(namespace_prefix)]
	smart_namespace := namespace[len(namespace_prefix)+len(namespace_suffix)+1:]
	attribute = strings.Join(smart_namespace, "/")
	return
}

func validateName(namespace []string) bool {
	for i, v := range namespace_prefix {
		if namespace[i] != v {
			return false
		}
	}

	offset := len(namespace_prefix) + 1
	for i, v := range namespace_suffix {
		if namespace[offset+i] != v {
			return false
		}
	}

	return true
}

// Function to check properness of configuration parameters
// and set plugin attribute accordingly
func (sc *SmartCollector) setProcDevPath(cfg interface{}) error {
	procPath, err := config.GetConfigItem(cfg, "proc_path")
	if err == nil && len(procPath.(string)) > 0 {
		procPathStats, err := os.Stat(procPath.(string))
		if err != nil {
			return err
		}
		if !procPathStats.IsDir() {
			return errors.New(fmt.Sprintf("%s is not a directory", procPath.(string)))
		}
		sc.proc_path = procPath.(string)
	}
	devPath, err := config.GetConfigItem(cfg, "dev_path")
	if err == nil && len(devPath.(string)) > 0 {
		devPathStats, err := os.Stat(devPath.(string))
		if err != nil {
			return err
		}
		if !devPathStats.IsDir() {
			return errors.New(fmt.Sprintf("%s is not a directory", devPath.(string)))
		}
		sc.dev_path = devPath.(string)
	}
	if sysUtilProvider == nil {
		sysUtilProvider = NewSysutilProvider(sc.proc_path, sc.dev_path)
	}
	return nil
}

type SmartCollector struct {
	logger    *log.Logger
	proc_path string
	dev_path  string
}

type smartResults map[string]interface{}

// CollectMetrics returns metrics from smart
func (sc *SmartCollector) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	err := sc.setProcDevPath(mts[0])
	if err != nil {
		return nil, err
	}

	buffered_results := map[string]smartResults{}

	results := make([]plugin.MetricType, len(mts))
	errs := make([]string, 0)

	collected := false

	t := time.Now()

	for i, mt := range mts {
		namespace := mt.Namespace().Strings()
		results[i] = plugin.MetricType{
			Namespace_: mt.Namespace(),
			Timestamp_: t,
		}

		if !validateName(namespace) {
			errs = append(errs, fmt.Sprintf("%s is not valid metric", mt.Namespace().String()))
			continue
		}
		disk, attribute_path := parseName(namespace)
		buffered, ok := buffered_results[disk]
		if !ok {
			values, err := ReadSmartData(disk, sysUtilProvider)
			if err != nil {
				return nil, err
			}
			buffered = values.GetAttributes()
			buffered_results[disk] = buffered
		}

		attribute, ok := buffered[attribute_path]

		if !ok {
			errs = append(errs, "Unknown attribute "+attribute_path)
		} else {
			collected = true
			results[i].Data_ = attribute
		}
	}

	errsStr := strings.Join(errs, "; ")
	if collected {
		if len(errs) > 0 {
			log.Printf("Data collected but error(s) occured: %v", errsStr)
		}
		return results, nil
	} else {
		return nil, errors.New(errsStr)
	}
}

// GetMetricTypes returns the metric types exposed by smart
func (sc *SmartCollector) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	err := sc.setProcDevPath(cfg)
	if err != nil {
		return nil, err
	}
	smart_metrics := ListAllKeys()
	devices, err := sysUtilProvider.ListDevices()
	if err != nil {
		return nil, err
	}
	mts := make([]plugin.MetricType, 0, len(smart_metrics)*len(devices))

	for _, device := range devices {
		for _, metric := range smart_metrics {
			path := makeName(device, metric)
			mts = append(mts, plugin.MetricType{Namespace_: core.NewNamespace(path...)})
		}
	}

	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicy
func (p *SmartCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	rule, _ := cpolicy.NewStringRule("proc_path", false, "/proc")
	node := cpolicy.NewPolicyNode()
	node.Add(rule)
	cp.Add([]string{nsVendor, nsClass, nsType}, node)
	rule, _ = cpolicy.NewStringRule("dev_path", false, "/dev")
	node.Add(rule)
	return cp, nil
}

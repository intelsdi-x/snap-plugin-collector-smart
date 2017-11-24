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
	"sync"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"

	"github.com/intelsdi-x/snap-plugin-utilities/config"
	log "github.com/sirupsen/logrus"
)

const (
	PluginName = "smart-disk"
	version    = 9
	pluginType = plugin.CollectorPluginType

	nsVendor = "intel"
	nsClass  = "disk"
	nsType   = "smart"
	devname  = "device"
)

var (
	//procPath source of data for metrics
	procPath = "/proc"
	//devPath source of data for metrics
	devPath = "/dev"

	namespace_prefix = []string{nsVendor, nsClass, nsType}

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
	imutex := new(sync.Mutex)
	return &SmartCollector{
		logger:           logger,
		initializedMutex: imutex,
		proc_path:        procPath,
		dev_path:         devPath,
	}
}

func parseName(namespace []string) (disk, attribute string) {
	disk = namespace[len(namespace_prefix)]
	smart_namespace := namespace[len(namespace_prefix)+1:]
	attribute = strings.Join(smart_namespace, "/")
	return
}

// Function to check properness of configuration parameters
// and set plugin attribute accordingly
func (sc *SmartCollector) setProcDevPath(cfg interface{}) error {
	sc.initializedMutex.Lock()
	defer sc.initializedMutex.Unlock()
	if sc.initialized {
		return nil
	}
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
	sc.initialized = true
	return nil
}

type SmartCollector struct {
	initialized      bool
	initializedMutex *sync.Mutex
	logger           *log.Logger
	proc_path        string
	dev_path         string
}

type smartResults map[string]interface{}

// DiskMetrics returns metrics from smart on given disk
func (sc *SmartCollector) DiskMetrics(ns []core.NamespaceElement,
	t time.Time, disk string, attribute_path string,
	buffered_results map[string]smartResults) (plugin.MetricType, error) {
	var result plugin.MetricType
	buffered, ok := buffered_results[disk]
	if !ok {
		values, err := ReadSmartData(disk, sysUtilProvider)
		if err != nil {
			return result, err
		}
		buffered = values.GetAttributes()
		buffered_results[disk] = buffered
	}
	attribute, ok := buffered[attribute_path]
	if !ok {
		return result, fmt.Errorf("Unknown attribute %s", attribute_path)
	}

	ns1 := make([]core.NamespaceElement, len(ns))
	copy(ns1, ns)
	ns1[3].Value = disk
	result = plugin.MetricType{
		Namespace_: ns1,
		Timestamp_: t,
		Version_:   version,
		Data_:      attribute,
	}

	return result, nil
}

// CollectMetrics returns metrics from smart
func (sc *SmartCollector) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	if err := sc.setProcDevPath(mts[0]); err != nil {
		return nil, err
	}

	buffered_results := map[string]smartResults{}
	results := []plugin.MetricType{}

	t := time.Now()
	for _, mt := range mts {
		ns := mt.Namespace()
		disk, attribute_path := parseName(ns.Strings())
		if disk == "*" {
			// All system disks requested
			devices, err := sysUtilProvider.ListDevices()
			if err != nil {
				return nil, err
			}
			for _, dev := range devices {
				result, err := sc.DiskMetrics(ns, t, dev, attribute_path, buffered_results)
				if err != nil {
					sc.logger.Warning(fmt.Sprintf("Error collecting SMART %s data on %s disk: %v", attribute_path, dev, err))
				} else {
					results = append(results, result)
				}
			}
		} else {
			// Single disk requested
			result, err := sc.DiskMetrics(ns, t, disk, attribute_path, buffered_results)
			if err != nil {
				sc.logger.Warning(fmt.Sprintf("Error collecting SMART %s data on %s disk: %v", attribute_path, disk, err))
			} else {
				results = append(results, result)
			}
		}
	}
	if len(results) == 0 {
		return nil, errors.New("No metrics found")

	}

	return results, nil
}

// GetMetricTypes returns the metric types exposed by smart
func (sc *SmartCollector) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	smart_metrics := ListAllKeys()
	mts := []plugin.MetricType{}
	for _, metric := range smart_metrics {
		ns := core.NewNamespace(namespace_prefix...).AddDynamicElement(devname, "SMART device")
		for _, elt := range strings.Split(metric, "/") {
			ns = ns.AddStaticElement(elt)
		}
		mts = append(mts, plugin.MetricType{
			Namespace_:   ns,
			Description_: "dynamic SMART metric: " + metric,
		})
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

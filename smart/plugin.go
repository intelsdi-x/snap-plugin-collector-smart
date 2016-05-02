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
	"log"
	"os"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
)

const (
	name       = "smart-disk"
	version    = 7
	pluginType = plugin.CollectorPluginType
)

var sysUtilProvider SysutilProvider = NewSysutilProvider()

var namespace_prefix = []string{"intel", "disk"}
var namespace_suffix = []string{"smart"}

func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

func NewSmartCollector() *SmartCollector {
	return &SmartCollector{}
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

type SmartCollector struct {
}

type smartResults map[string]interface{}

// CollectMetrics returns metrics from smart
func (sc *SmartCollector) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	buffered_results := map[string]smartResults{}

	results := make([]plugin.MetricType, len(mts))
	errs := make([]string, 0)

	collected := false

	t := time.Now()
	host, _ := os.Hostname()

	for i, mt := range mts {
		tags := mt.Tags()
		if tags == nil {
			tags = map[string]string{}
		}
		tags["hostname"] = host

		namespace := mt.Namespace().Strings()
		results[i] = plugin.MetricType{
			Namespace_: mt.Namespace(),
			Tags_:      tags,
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
func (sc *SmartCollector) GetMetricTypes(_ plugin.ConfigType) ([]plugin.MetricType, error) {
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
	c := cpolicy.New()
	return c, nil
}

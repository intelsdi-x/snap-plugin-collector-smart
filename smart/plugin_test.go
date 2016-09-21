//
// +build small

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
	"os"
	"strings"
	"sync"
	"testing"

	log "github.com/Sirupsen/logrus"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/core/cdata"

	. "github.com/smartystreets/goconvey/convey"
)

type fakeSysutilProvider2 struct {
	FillBuf []byte
}

func (s *fakeSysutilProvider2) ListDevices() ([]string, error) {
	return []string{"DEV_ONE", "DEV_TWO"}, nil
}

func (s *fakeSysutilProvider2) OpenDevice(device string) (*os.File, error) {
	return nil, nil
}

func (s *fakeSysutilProvider2) Ioctl(fd uintptr, cmd uint, buf []byte) error {
	if cmd == smart_read_values {
		for i, v := range s.FillBuf {
			buf[i] = v
		}
	}
	return nil
}

func sysUtilWithMetrics(metrics []byte) fakeSysutilProvider2 {
	util := fakeSysutilProvider2{FillBuf: make([]byte, 512)}

	for i, m := range metrics {
		util.FillBuf[2+i*12] = m
	}

	return util
}

func TestSmartCollectorPlugin(t *testing.T) {
	Convey("Meta should return Metadata for the plugin", t, func() {
		meta := Meta()
		So(meta.Name, ShouldResemble, PluginName)
		So(meta.Version, ShouldResemble, version)
		So(meta.Type, ShouldResemble, plugin.CollectorPluginType)
	})

	Convey("Create Smart Collector", t, func() {
		sCol := NewSmartCollector()
		Convey("So sCol should not be nil", func() {
			So(sCol, ShouldNotBeNil)
		})
		Convey("So sCol should be of Psutil type", func() {
			So(sCol, ShouldHaveSameTypeAs, &SmartCollector{})
		})
		Convey("sCol.GetConfigPolicy() should return a config policy", func() {
			configPolicy, _ := sCol.GetConfigPolicy()
			Convey("So config policy should not be nil", func() {
				So(configPolicy, ShouldNotBeNil)
			})
			Convey("So config policy should be a cpolicy.ConfigPolicy", func() {
				So(configPolicy, ShouldHaveSameTypeAs, &cpolicy.ConfigPolicy{})
			})
		})
	})
}

func TestGetMetricTypes(t *testing.T) {
	Convey("When having two devices with known smart attribute", t, func() {

		Convey("And system lets you to list devices", func() {
			provider := &fakeSysutilProvider2{}

			orgProvider := sysUtilProvider
			sysUtilProvider = provider

			collector := SmartCollector{
				logger:           log.New(),
				initializedMutex: new(sync.Mutex),
				proc_path:        "/proc",
				dev_path:         "/dev",
			}

			Convey("Both devices should be present in metric list", func() {

				new_hier, is_dynamic := false, false
				metrics, err := collector.GetMetricTypes(plugin.NewPluginConfigType())
				So(err, ShouldBeNil)

				for _, m := range metrics {
					switch m.Namespace().Strings()[2] {
					case "smart":
						new_hier = true
					}
					switch m.Namespace().Strings()[3] {
					case "*":
						is_dynamic = true
					}
				}

				So(new_hier, ShouldBeTrue)
				So(is_dynamic, ShouldBeTrue)

			})

			Reset(func() {
				sysUtilProvider = orgProvider
			})

		})

	})

}

func TestParseName(t *testing.T) {
	Convey("When given correct namespace refering to single word attribute", t, func() {

		disk, attr := parseName([]string{"intel", "disk", "smart", "DEV", "abc"})

		Convey("Device should be correctly extracted", func() {

			So(disk, ShouldEqual, "DEV")

		})

		Convey("Attribute should be correctly extracted", func() {

			So(attr, ShouldEqual, "abc")

		})

	})

	Convey("When given correct namespace refering to multi level attribute", t, func() {

		disk, attr := parseName([]string{"intel", "disk", "smart", "DEV",
			"abc", "def"})

		Convey("Device should be correctly extracted", func() {

			So(disk, ShouldEqual, "DEV")

		})

		Convey("Attribute should be correctly extracted", func() {

			So(attr, ShouldEqual, "abc/def")

		})

	})

}

func TestValidateName(t *testing.T) {
	Convey("When given namespace with invalid prefix", t, func() {

		test := validateName([]string{"intel", "cake", "smart", "DEV",
			"abc", "def"})

		Convey("Validation should fail", func() {

			So(test, ShouldBeFalse)

		})

	})

	Convey("When given namespace with invalid suffix", t, func() {

		test := validateName([]string{"intel", "disk", "dumb", "DEV",
			"abc", "def"})

		Convey("Validation should fail", func() {

			So(test, ShouldBeFalse)

		})

	})

	Convey("When given correct namespace refering to single word attribute", t, func() {

		test := validateName([]string{"intel", "disk", "smart", "DEV", "abc"})

		Convey("Validation should pass", func() {

			So(test, ShouldBeTrue)

		})

	})

	Convey("When given correct namespace refering to multi level attribute", t, func() {

		test := validateName([]string{"intel", "disk", "smart", "DEV",
			"abc", "def"})
		Convey("Validation should pass", func() {

			So(test, ShouldBeTrue)

		})

	})
}

func TestCollectMetrics(t *testing.T) {
	Convey("Using fake system", t, func() {

		orgReader := ReadSmartData
		orgProvider := sysUtilProvider

		sc := SmartCollector{
			logger:           log.New(),
			initializedMutex: new(sync.Mutex),
			proc_path:        "/proc",
			dev_path:         "/dev",
		}
		cfg := cdata.NewNode()

		metric_id, metric_name := firstKnownMetric()
		metric_ns := strings.Split(metric_name, "/")

		Convey("When asked about metric not in valid namespace", func() {

			_, err := sc.CollectMetrics([]plugin.MetricType{
				{
					Namespace_: core.NewNamespace("cake"),
					Config_:    cfg,
				},
			})

			Convey("Returns error", func() {

				So(err, ShouldNotBeNil)

				Convey("Error is about invalid metric", func() {

					So(err.Error(), ShouldContainSubstring, "not valid metric")

				})

			})

		})

		Convey("When asked about metric in valid namespace but unknown to reader", func() {

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				return nil, errors.New("x not valid disk")
			}

			_, err := sc.CollectMetrics([]plugin.MetricType{
				{
					Namespace_: core.NewNamespace("intel", "disk", "smart", "x", "y"),
					Config_:    cfg,
				},
			})

			Convey("Returns error", func() {

				So(err, ShouldNotBeNil)

				Convey("Error is about invalid metric", func() {

					So(err.Error(), ShouldContainSubstring, "not valid disk")

				})
			})

		})

		Convey("When asked about metric in valid namespace but reading fails", func() {

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				return nil, errors.New("Something")
			}

			_, err := sc.CollectMetrics([]plugin.MetricType{
				{
					Namespace_: core.NewNamespace("intel", "disk", "smart", "sda", "y"),
					Config_:    cfg,
				},
			})

			Convey("Returns error", func() {

				So(err, ShouldNotBeNil)

			})

		})

		Convey("When asked about metric in valid namespace", func() {

			drive_asked := ""

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				drive_asked = device

				result := SmartValues{}
				result.Values[0].Id = metric_id

				return &result, nil
			}

			metrics, err := sc.CollectMetrics([]plugin.MetricType{
				{
					Namespace_: core.NewNamespace("intel", "disk", "smart", "my_disk").AddStaticElements(metric_ns...),
					Config_:    cfg,
				},
			})

			Convey("Asks reader to read metric from correct drive", func() {
				So(err, ShouldBeNil)
				So(len(metrics), ShouldEqual, 1)
				So(drive_asked, ShouldEqual, "my_disk")

				Convey("Returns value of metric from reader", func() {
					So(len(metrics), ShouldBeGreaterThan, 0)

					//TODO: Value is correct

				})

			})

		})

		Convey("When asked about metrics in valid namespaces", func() {

			asked := map[string]int{"x": 1, "y": 2}

			ReadSmartData = func(device string,
				sysutilProvider SysutilProvider) (*SmartValues, error) {
				asked[device]++

				result := SmartValues{}
				result.Values[0].Id = metric_id

				return &result, nil
			}
			sc.CollectMetrics([]plugin.MetricType{
				{
					Namespace_: core.NewNamespace("intel", "disk", "smart", "sda").AddStaticElements(metric_ns...),
					Config_:    cfg,
				},
				{
					Namespace_: core.NewNamespace("intel", "disk", "smart", "sdb").AddStaticElements(metric_ns...),
					Config_:    cfg,
				},
				{
					Namespace_: core.NewNamespace("intel", "disk", "smart", "sdb").AddStaticElements(metric_ns...),
					Config_:    cfg,
				},
				{
					Namespace_: core.NewNamespace("intel", "disk", "smart", "sda").AddStaticElements(metric_ns...),
					Config_:    cfg,
				},
			})

			Convey("Reader is asked once per drive", func() {
				So(asked["x"], ShouldEqual, 1)
				So(asked["y"], ShouldEqual, 2)

			})

		})

		Reset(func() {
			sysUtilProvider = orgProvider
			ReadSmartData = orgReader
		})

	})
}

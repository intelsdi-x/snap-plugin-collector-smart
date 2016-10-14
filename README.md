# snap collector plugin - SMART

This plugin monitors storage systems from Intel's SSDs. Raw data interpretation is based on [State Drive DC S3700 Series specification](http://www.intel.com/content/dam/www/public/us/en/documents/product-specifications/ssd-dc-s3700-spec.pdf). 
Other disks may have different attributes or different raw data formats.

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Installation](#installation)
  * [Configuration and Usage](configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license-and-authors)
6. [Acknowledgements](#acknowledgements)

## Getting Started

Plugin directly reads underlying device parameters using [ioctl(2)](http://man7.org/linux/man-pages/man2/ioctl.2.html)

### System Requirements
* [golang 1.5+](https://golang.org/dl/)  (needed only for building)

### Operating systems
All OSs currently supported by plugin:
* Linux/amd64

### Configuration and Usage

**Enable SMART support in BIOS**

### Installation
#### Download SMART plugin binary:
You can get the pre-built binaries for your OS and architecture at snap's [GitHub Releases](https://github.com/intelsdi-x/snap/releases) page.

#### To build the plugin binary:
Fork https://github.com/intelsdi-x/snap-plugin-collector-smart  
Clone repo into `$GOPATH/src/github.com/intelsdi-x/`:

```
$ git clone https://github.com/<yourGithubID>/snap-plugin-collector-smart.git
```

Build the plugin by running make within the cloned repo:
```
$ make
```
This builds the plugin in `/build/rootfs/`


## Documentation

### Collected Metrics
This plugin has the ability to gather the following metrics:

Namespace | Data Type | Description
----------|-----------|-----------------------
/intel/disk/smart/\<device_name\>/reallocatedsectors | | number of retired blocks
/intel/disk/smart/\<device_name\>/reallocatedsectors/normalized | | shows percent remaining of allowable grown defect count
/intel/disk/smart/\<device_name\>/poweronhours | | cumulative power-on time in hours
/intel/disk/smart/\<device_name\>/poweronhours/normalized | | always 100
/intel/disk/smart/\<device_name\>/powercyclecount | | cumulative number of power cycle events
/intel/disk/smart/\<device_name\>/powercyclecount/normalized | | always 100
/intel/disk/smart/\<device_name\>/availablereservedspace | | available reserved space
/intel/disk/smart/\<device_name\>/availablereservedspace/normalized | | undocumented
/intel/disk/smart/\<device_name\>/programfailcount | | total count of program fails
/intel/disk/smart/\<device_name\>/programfailcount/normalized | | percent remaining of allowable program fails
/intel/disk/smart/\<device_name\>/erasefailcount | | total count of erase fails
/intel/disk/smart/\<device_name\>/erasefailcount/percent | | remaining of allowable erase fails
/intel/disk/smart/\<device_name\>/unexpectedpowerloss | | cumulative number of unclean shutdowns
/intel/disk/smart/\<device_name\>/unexpectedpowerloss/normalized | | always 100
/intel/disk/smart/\<device_name\>/powerlossprotectionfailure | | last test result as microseconds to discharge capacitor
/intel/disk/smart/\<device_name\>/powerlossprotectionfailure/sincelast | | minutes since last test
/intel/disk/smart/\<device_name\>/powerlossprotectionfailure/tests | | lifetime number of tests
/intel/disk/smart/\<device_name\>/powerlossprotectionfailure/normalized | | 1 on test failure, 11 if capacitor tested in excessive temperature, otherwise 100
/intel/disk/smart/\<device_name\>/satadownshifts | | number of times SATA interface selected lower signaling rate due to error
/intel/disk/smart/\<device_name\>/satadownshifts/normalized | | always 100
/intel/disk/smart/\<device_name\>/e2eerrors | | number of LBA tag mismatches in end-to-end data protection path
/intel/disk/smart/\<device_name\>/e2eerrors/normalized | | always 100
/intel/disk/smart/\<device_name\>/uncorrectableerrors | | number of errors that could not be recovered using Error Correction Code
/intel/disk/smart/\<device_name\>/uncorrectableerrors/normalized | | always 100
/intel/disk/smart/\<device_name\>/casetemperature | | SSD case temperature in Celsius
/intel/disk/smart/\<device_name\>/casetemperature/min | | minimal value
/intel/disk/smart/\<device_name\>/casetemperature/max | | maximal value
/intel/disk/smart/\<device_name\>/casetemperature/overcounter | | number of times sampled temperature exceeds drive max operating temperature spec.
/intel/disk/smart/\<device_name\>/casetemperature/normalized | | value (100-temperature in Celsius)
/intel/disk/smart/\<device_name\>/unsafeshutdowns | | cumulative number of unsafe shutdowns
/intel/disk/smart/\<device_name\>/unsafeshutdowns/normalized | | always 100
/intel/disk/smart/\<device_name\>/internaltemperature | | device internal temperature in Celsius. Reading from PCB.
/intel/disk/smart/\<device_name\>/internaltemperature/normalized | | (150 temperature in Celsius) or 100 if temperature is less than 50.
/intel/disk/smart/\<device_name\>/pendingsectors | | number of current unrecoverable read errors that will be re-allocated on next write.
/intel/disk/smart/\<device_name\>/pendingsectors/normalized | | always 100.
/intel/disk/smart/\<device_name\>/crcerrors | | total number of encountered SATA CRC errors.
/intel/disk/smart/\<device_name\>/crcerrors/normalized | | always 100
/intel/disk/smart/\<device_name\>/hostwrites | | total number of sectors written by the host system
/intel/disk/smart/\<device_name\>/hostwrites/normalized | | always 100
/intel/disk/smart/\<device_name\>/timedworkload | |
/intel/disk/smart/\<device_name\>/timedworkload/mediawear | | measures the wear seen by the SSD (since reset of the workload timer, see timedworkload/time), as a percentage of the maximum rated cycles.
/intel/disk/smart/\<device_name\>/timedworkload/mediawear/normalized | | always 100
/intel/disk/smart/\<device_name\>/timedworkload/readpercent | | shows the percentage of I/O operations that are read operations (since reset of the workload timer, see timedworkload/time)
/intel/disk/smart/\<device_name\>/timedworkload/readpercent/normalized | | always 100
/intel/disk/smart/\<device_name\>/timedworkload/time | | number of minutes since starting workload timer
/intel/disk/smart/\<device_name\>/timedworkload/time/normalized | | always 100
/intel/disk/smart/\<device_name\>/reservedblocks | | number of reserved blocks remaining
/intel/disk/smart/\<device_name\>/reservedblocks/normalized | | percentage of reserved space available
/intel/disk/smart/\<device_name\>/wearout | | always 0
/intel/disk/smart/\<device_name\>/wearout/number | | of cycles the NAND media has undergone. Declines linearly from 100 to 1 as the average erase cycle count increases from 0 to the maximum rated cycles. Once it reaches 1 the number will not decrease, although it is likely that significant additional wear can be put on the device.
/intel/disk/smart/\<device_name\>/thermalthrottle | | percent throttle status
/intel/disk/smart/\<device_name\>/thermalthrottle/eventcount | | number of times thermal throttle has activated. Preserved over power cycles.
/intel/disk/smart/\<device_name\>/thermalthrottle/normalized | | always 100
/intel/disk/smart/\<device_name\>/totallba | |
/intel/disk/smart/\<device_name\>/totallba/written | | total number of sectors written by the host system
/intel/disk/smart/\<device_name\>/totallba/written/normalized | | always 100
/intel/disk/smart/\<device_name\>/read | | total number of sectors read by the host system
/intel/disk/smart/\<device_name\>/read/normalized | | always 100

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-smart/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-smart/pulls).

## Community Support
This repository is one of **many** plugins in **snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support)

## Contributing
We love contributions

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Lukasz Mroz](https://github.com/lmroz)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

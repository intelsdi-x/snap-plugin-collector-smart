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
You can get the pre-built binaries for your OS and architecture at Snap's [GitHub Releases](https://github.com/intelsdi-x/snap/releases) page.

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
This builds the plugin in `/build/`


## Documentation

### Collected Metrics

List of collected metrics is described in [METRICS.md](METRICS.md).

### Roadmap
There isn't a current roadmap for this plugin, but it is in active development. As we launch this plugin, we do not have any outstanding requirements for the next release. If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-smart/issues/new) and/or submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-smart/pulls).

## Community Support
This repository is one of **many** plugins in **Snap**, a powerful telemetry framework. See the full project at http://github.com/intelsdi-x/snap
To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support) or visit [Slack](http://slack.snap-telemetry.io).

## Contributing
We love contributions

There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).

## Acknowledgements

* Author: [Lukasz Mroz](https://github.com/lmroz)

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Copy event**: New event type for copying files to/from containers (`copy_to`, `copy_from`)
- **Shell event**: New event type for executing shell commands in containers
- **Tool scripts**: Added utility scripts in `tool/` directory

### Fixed
- **Network cleanup on error**: Ensure containerlab network is destroyed even when after() encounters errors during log collection
- **Config event container name**: Use correct containerlab container name format (`clab-<topo>-<host>`) for vtysh config changes
- **Config event vtysh execution**: Use `vtysh -c` multiple times instead of `vtysh -f` with temp file (fixes `Unknown command` errors)
- **Copy event destination directory**: Auto-create destination directory when copying files from container
- **Tcpdump log directory**: Use `MkdirAll` to create parent directories when saving tcpdump logs

### Improved
- **Error messages**: Enhanced error messages with detailed context (container name, file paths, command output) for easier debugging

### Changed
- **update_yaml_paths.sh**: Support both JSON and YAML scenario file formats

## [0.0.2] - 2025-08-02

### Added
- **Stress command**: Pumba stress command for CPU/memory stress testing (PR #6)
- **Rate command**: Pumba rate command for bandwidth limiting (PR #5)
- **Data path in scenario**: Added `data` field to scenario file for specifying dot2net data.json path (PR #3)

### Fixed
- **DestroyNetwork function**: Fixed error preventing network from being destroyed properly (PR #4)
- **TcpdumpLog function**: Ensure log directory is created before writing log files (PR #2)

### Changed
- **Scenario file structure**: Updated scenario file format with clear paths for topo, data, and logPath
- **Example files**: Updated example scenario files with correct paths

## [0.0.1] - 2023-09-22

### Added
- Initial release
- Basic scenario execution with JSON/YAML support
- Containerlab integration for network emulation
- Pumba integration for network fault injection (delay, loss, corrupt, duplicate)
- tcpdump logging on target hosts
- Log file collection after scenario completion

[Unreleased]: https://github.com/cpflat/netroub/compare/v0.0.2...HEAD
[0.0.2]: https://github.com/cpflat/netroub/compare/v0.0.1...v0.0.2
[0.0.1]: https://github.com/cpflat/netroub/releases/tag/v0.0.1

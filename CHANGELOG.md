# ChangeLog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added

- Added option to set `--cache-freq`.
- Added option to set `--zap-log-level`.
- Added more data fields to existing collections -- expanded node and job
  states, memory usage.
- Added accounting data collection -- job states, TRES usage.
- Added scheduler statistics collection.
- Added effective CPU and memory totals for nodes.

### Fixed

- Fixed image tag incorrectly defaulting to appVersion instead of version.
- Fixed update strategies employing `Recreate` when unnecessary.

### Changed

- Changed `--server` default to localhost URL.
- Changed slurm-client version to 0.3.0

### Removed

- Removed `--per-user-metrics` toggle, enabled automatically.

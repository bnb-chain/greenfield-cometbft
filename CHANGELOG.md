# CHANGELOG
## v1.3.2
This release updates the dependencies in this repo

Features:
* [#73](https://github.com/bnb-chain/greenfield-cometbft/pull/73) update prysm version
* [#77](https://github.com/bnb-chain/greenfield-cometbft/pull/77) chore: update dependencies and fix workflows
* [#85](https://github.com/bnb-chain/greenfield-cometbft/pull/85) feat: update btcec

## v1.3.0
This release fixes the vulnerabilities in the repo

* [#69](https://github.com/bnb-chain/greenfield-cometbft/pull/69) chore: upgrade deps for fixing vulnerabilities

## v1.2.0
This release supports state sync at specific height and custom blocks to rollback

Features:
* [#54](https://github.com/bnb-chain/greenfield-cometbft/pull/54) feat: add support for state sync at specific height
* [#55](https://github.com/bnb-chain/greenfield-cometbft/pull/55) feat: add support for custom blocks to rollback

## v1.1.0
This release supports 6 new json rpc queries and also resolves a replay issue.

Features:
* [#42](https://github.com/bnb-chain/greenfield-cometbft/pull/42) feat: add support for some json rpc queries

Bugfixes:
* [#37](https://github.com/bnb-chain/greenfield-cometbft/pull/37) fix: replay issue with mock app

Chores:
* [#49](https://github.com/bnb-chain/greenfield-cometbft/pull/49) chore: bump golang net and grpc lib to secure version

## v1.0.0
This release includes 1 bugfix.

Bugfixes:
* [#36](https://github.com/bnb-chain/greenfield-cometbft/pull/36) fix: trim the prefix 0 for eth_chainId


## v0.0.3
This release includes the features and bugfixes in the v0.0.3 alpha versions and 1 new bugfix.

Bugfixes:
* [#28](https://github.com/bnb-chain/greenfield-cometbft/pull/28) fix: fix dependency security issues

## v0.0.3-alpha.1
This release adapts the repository to the upstream v0.37.2 version.

Features:
* [#21](https://github.com/bnb-chain/greenfield-cometbft/pull/21) feat: add event types for op cross chain events
* [#22](https://github.com/bnb-chain/greenfield-cometbft/pull/22) feat: add flag to skip app hash verification

Bugfixes:
* [#20](https://github.com/bnb-chain/greenfield-cometbft/pull/20) fix: merge upstream v0.37.2


## v0.0.2
This release mainly includes following feature, bugfix and improvement:

Features:
* [#11](https://github.com/bnb-chain/greenfield-cometbft/pull/11) feat: add option to disable tx event indexing
* [#14](https://github.com/bnb-chain/greenfield-cometbft/pull/14) feat: add websocket client

Bugfixes:
* [#4](https://github.com/bnb-chain/greenfield-cometbft/pull/4) fix: infinite re-entry of MarshalJSON method
* [#6](https://github.com/bnb-chain/greenfield-cometbft/pull/6) fix: rollback LastRandaoMix was incorrect
* [#16](https://github.com/bnb-chain/greenfield-cometbft/pull/16) fix: block include nonce mismatch tx issue

Improvements:
* [#5](https://github.com/bnb-chain/greenfield-cometbft/pull/5) perf: performance improvement
* [#8](https://github.com/bnb-chain/greenfield-cometbft/pull/8) perf: remove some local client's mutex


## v0.0.1

This release mainly includes following feature:
1. vote pool for cross-chain and data availability challenges
2. randao feature for data availability challenges
3. support EVM json-rpc requests

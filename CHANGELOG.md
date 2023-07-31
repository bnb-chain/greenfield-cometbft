# CHANGELOG

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

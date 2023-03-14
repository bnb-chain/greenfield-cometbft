# Greenfield Tendermint

Greenfield Tendermint, forked from [tendermint](https://github.com/tendermint/tendermint),
is the consensus layer of Greenfield blockchain.
Tendermint Core is a Byzantine Fault Tolerant (BFT) middleware that takes a
state transition machine - written in any programming language - and securely
replicates it on many machines.

For protocol details, refer to the [Tendermint Specification](./spec/README.md).

For detailed analysis of the consensus protocol, including safety and liveness
proofs, read our paper, "[The latest gossip on BFT
consensus](https://arxiv.org/abs/1807.04938)".

## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Key features

We implement several key features based on the Tendermint fork:

* Vote Pool. Vote pool is used to collect votes from different validators for off-chain consensus.
Currently, it is mainly used for cross chain and data availability challenge in Greenfield blockchain.
* RANDAO. RANDAO is introduced for on-chain randomness. Overall, the idea is very similar to the RANDAO
in Ethereum beacon chain, you can refer to [here](https://eth2book.info/altair/part2/building_blocks/randomness)
for more information. It has some limitations, please use it with caution.

## Minimum requirements

| Requirement | Notes             |
|-------------|-------------------|
| Go version  | Go 1.18 or higher |

### Install

See the [install instructions](./docs/introduction/install.md).

### Quick Start

- [Single node](./docs/introduction/quick-start.md)
- [Local cluster using docker-compose](./docs/tools/docker-compose.md)
- [Remote cluster using Terraform and Ansible](./docs/tools/terraform-and-ansible.md)

## Contributing

Please abide by the [Code of Conduct](CODE_OF_CONDUCT.md) in all interactions.

Before contributing to the project, please take a look at the [contributing
guidelines](CONTRIBUTING.md) and the [style guide](STYLE_GUIDE.md). You may also
find it helpful to read the [specifications](./spec/README.md), and familiarize
yourself with our [Architectural Decision Records
(ADRs)](./docs/architecture/README.md) and
[Request For Comments (RFCs)](./docs/rfc/README.md).


## Resources

### Libraries

- [Cosmos SDK](http://github.com/cosmos/cosmos-sdk); A framework for building
  applications in Golang
- [Tendermint in Rust](https://github.com/informalsystems/tendermint-rs)
- [ABCI Tower](https://github.com/penumbra-zone/tower-abci)

### Applications

- [Cosmos Hub](https://hub.cosmos.network/)
- [Terra](https://www.terra.money/)
- [Celestia](https://celestia.org/)
- [Anoma](https://anoma.network/)
- [Vocdoni](https://docs.vocdoni.io/)

### Research

- [The latest gossip on BFT consensus](https://arxiv.org/abs/1807.04938)
- [Master's Thesis on Tendermint](https://atrium.lib.uoguelph.ca/xmlui/handle/10214/9769)
- [Original Whitepaper: "Tendermint: Consensus Without Mining"](https://tendermint.com/static/docs/tendermint.pdf)
- [Tendermint Core Blog](https://medium.com/tendermint/tagged/tendermint-core)
- [Cosmos Blog](https://blog.cosmos.network/tendermint/home)

## License

To be added.

[bft]: https://en.wikipedia.org/wiki/Byzantine_fault_tolerance
[smr]: https://en.wikipedia.org/wiki/State_machine_replication
[Blockchain]: https://en.wikipedia.org/wiki/Blockchain
[version-badge]: https://img.shields.io/github/tag/tendermint/tendermint.svg
[version-url]: https://github.com/tendermint/tendermint/releases/latest
[api-badge]: https://camo.githubusercontent.com/915b7be44ada53c290eb157634330494ebe3e30a/68747470733a2f2f676f646f632e6f72672f6769746875622e636f6d2f676f6c616e672f6764646f3f7374617475732e737667
[api-url]: https://pkg.go.dev/github.com/tendermint/tendermint
[go-badge]: https://img.shields.io/badge/go-1.18-blue.svg
[go-url]: https://github.com/moovweb/gvm
[discord-badge]: https://img.shields.io/discord/669268347736686612.svg
[discord-url]: https://discord.gg/cosmosnetwork
[license-badge]: https://img.shields.io/github/license/tendermint/tendermint.svg
[license-url]: https://github.com/tendermint/tendermint/blob/main/LICENSE
[sg-badge]: https://sourcegraph.com/github.com/tendermint/tendermint/-/badge.svg
[sg-url]: https://sourcegraph.com/github.com/tendermint/tendermint?badge
[tests-url]: https://github.com/tendermint/tendermint/actions/workflows/tests.yml
[tests-badge]: https://github.com/tendermint/tendermint/actions/workflows/tests.yml/badge.svg?branch=main
[lint-badge]: https://github.com/tendermint/tendermint/actions/workflows/lint.yml/badge.svg
[lint-url]: https://github.com/tendermint/tendermint/actions/workflows/lint.yml

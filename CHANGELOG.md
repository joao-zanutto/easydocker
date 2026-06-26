# Changelog

## [1.3.0](https://github.com/joao-zanutto/easydocker/compare/v1.2.0...v1.3.0) (2026-06-26)


### Features

* add compose support ([#54](https://github.com/joao-zanutto/easydocker/issues/54)) ([8d0e38b](https://github.com/joao-zanutto/easydocker/commit/8d0e38bc1fca65090cce54809c1b2b9686e0acad))
* add config via env var, cli and file ([#124](https://github.com/joao-zanutto/easydocker/issues/124)) ([f4dece4](https://github.com/joao-zanutto/easydocker/commit/f4dece42d4d77f4022b60f25bd565b4d3f01350a))
* add esc menu ([#67](https://github.com/joao-zanutto/easydocker/issues/67)) ([2211b2b](https://github.com/joao-zanutto/easydocker/commit/2211b2b5c62f3cf2803f104e4c131f44be095851))
* add file-based logging and web based profiling ([#123](https://github.com/joao-zanutto/easydocker/issues/123)) ([dff02f9](https://github.com/joao-zanutto/easydocker/commit/dff02f928142004ece3827f8d4c0b83d922dda5a))
* add visual cues to compose rows ([#55](https://github.com/joao-zanutto/easydocker/issues/55)) ([85c6991](https://github.com/joao-zanutto/easydocker/commit/85c699126c3e3e4d11cd6bf99cd3ce95dbfe1f2b))
* **easydocker:** add interactive terminal access for containers  ([#11](https://github.com/joao-zanutto/easydocker/issues/11)) ([721999d](https://github.com/joao-zanutto/easydocker/commit/721999d3b0f1052fb939f20340ac8c70c972e831))
* implement gomock ([#89](https://github.com/joao-zanutto/easydocker/issues/89)) ([7159deb](https://github.com/joao-zanutto/easydocker/commit/7159deb90161e5f6c2f2c45d6c44f56963fd83ea))
* implement inspect ([#66](https://github.com/joao-zanutto/easydocker/issues/66)) ([c112202](https://github.com/joao-zanutto/easydocker/commit/c1122023c5d2dc9c0ed0cb84d0b57a06e9d1bfb8))
* intelligent JSON filtering for inspect results ([#111](https://github.com/joao-zanutto/easydocker/issues/111)) ([1172760](https://github.com/joao-zanutto/easydocker/commit/1172760bb0c66a4bfb54e55f3368a3aae6b4bce9))
* load data in parallel ([#88](https://github.com/joao-zanutto/easydocker/issues/88)) ([a9c5a85](https://github.com/joao-zanutto/easydocker/commit/a9c5a85f4f8b0a24ed187ad6ed9573ee9ee1e207))
* **ui:** header and footer overhaul ([#99](https://github.com/joao-zanutto/easydocker/issues/99)) ([5042770](https://github.com/joao-zanutto/easydocker/commit/50427702c5b00881f757afbd17059e404d440b2c))


### Bug Fixes

* "follow" indicator disappearing from header when turned off ([#78](https://github.com/joao-zanutto/easydocker/issues/78)) ([ebb2180](https://github.com/joao-zanutto/easydocker/commit/ebb2180f1b3eb1ff8785bb935fc126d0b180af86))
* adjust app paddings ([#74](https://github.com/joao-zanutto/easydocker/issues/74)) ([2965662](https://github.com/joao-zanutto/easydocker/commit/2965662f2e404d8520b6385b4f9e43a2798ecee2))
* adjust container count in image details ([#116](https://github.com/joao-zanutto/easydocker/issues/116)) ([d0d7416](https://github.com/joao-zanutto/easydocker/commit/d0d741667db7cde7ad064ed9dbfd713cd03a8b2e))
* adjust visible rows calculation and sync viewport data in render methods ([#126](https://github.com/joao-zanutto/easydocker/issues/126)) ([52f85bc](https://github.com/joao-zanutto/easydocker/commit/52f85bcc3d01db49cca1d588b52106826d611ba5))
* aggressive mem allocation ([#101](https://github.com/joao-zanutto/easydocker/issues/101)) ([98c42d8](https://github.com/joao-zanutto/easydocker/commit/98c42d8149de6893b507ce6e1e0b256ded363279))
* color state label rendering by removing ANSI reset codes ([#91](https://github.com/joao-zanutto/easydocker/issues/91)) ([8921a0b](https://github.com/joao-zanutto/easydocker/commit/8921a0bb14d76bab2fce49fd066acca4abdefd85))
* **docs:** rebuild docs on new releases ([#52](https://github.com/joao-zanutto/easydocker/issues/52)) ([707634f](https://github.com/joao-zanutto/easydocker/commit/707634f7ec9a014f9be6f218337c15a1959a0a5b))
* help window height when scrolling down ([#73](https://github.com/joao-zanutto/easydocker/issues/73)) ([5bb9180](https://github.com/joao-zanutto/easydocker/commit/5bb91802776b20779ea0cfeeb29ebc1644915c81))
* interactive shell not starting on inspect mode ([#72](https://github.com/joao-zanutto/easydocker/issues/72)) ([8b974bb](https://github.com/joao-zanutto/easydocker/commit/8b974bb3cd7e22cc57dbb7ed2f7196dc2225d7d9))
* remove scroll keys from footer ([#76](https://github.com/joao-zanutto/easydocker/issues/76)) ([e7c43ae](https://github.com/joao-zanutto/easydocker/commit/e7c43ae7eeb4dad816c8e20ec36c45e6bc4b915e))
* reset horizontal position when entering viewer screen ([#81](https://github.com/joao-zanutto/easydocker/issues/81)) ([cc54a48](https://github.com/joao-zanutto/easydocker/commit/cc54a4890caa447b50ba98f265e1f4ae363fe7ec))
* **ui:** regression and add tests ([#103](https://github.com/joao-zanutto/easydocker/issues/103)) ([58c88bd](https://github.com/joao-zanutto/easydocker/commit/58c88bddfe6843c1834a59f6afd6994488043f63))
* user position after log history load ([#65](https://github.com/joao-zanutto/easydocker/issues/65)) ([cf2f2e6](https://github.com/joao-zanutto/easydocker/commit/cf2f2e6eedc2619fe666341ce9719fcc94176f3e))


### Performance Improvements

* reduce allocations and add caching in hot paths ([#96](https://github.com/joao-zanutto/easydocker/issues/96)) ([5962832](https://github.com/joao-zanutto/easydocker/commit/5962832597de8682d862b333a91a3f6e4a681397))

## [1.2.0](https://github.com/joao-zanutto/easydocker/compare/v1.1.0...v1.2.0) (2026-04-24)


### Features

* add changelog to docs ([#49](https://github.com/joao-zanutto/easydocker/issues/49)) ([48db33d](https://github.com/joao-zanutto/easydocker/commit/48db33da31f24f5efe6fe3c1996a676508a4ddd8))
* add github pages website ([#44](https://github.com/joao-zanutto/easydocker/issues/44)) ([5f14ae5](https://github.com/joao-zanutto/easydocker/commit/5f14ae5006bdc4ffa710245b67839f8dc0674cfe))
* add horizontal scroll indicator ([#38](https://github.com/joao-zanutto/easydocker/issues/38)) ([59298d2](https://github.com/joao-zanutto/easydocker/commit/59298d2c9d6208cbfec0bcba299b0f3f96b0df6d))
* add MIT license ([#33](https://github.com/joao-zanutto/easydocker/issues/33)) ([2875578](https://github.com/joao-zanutto/easydocker/commit/2875578e975062f402e517c2d0d7929daa411885))
* add readme badges ([#40](https://github.com/joao-zanutto/easydocker/issues/40)) ([cf50c28](https://github.com/joao-zanutto/easydocker/commit/cf50c2862c02aabb0abfeaea9f259dd369e08508))
* implement log line wrapping ([#32](https://github.com/joao-zanutto/easydocker/issues/32)) ([a57b908](https://github.com/joao-zanutto/easydocker/commit/a57b908240b08302dfb5517e9d50ebed28900f70))
* split repo and tags from image view ([#30](https://github.com/joao-zanutto/easydocker/issues/30)) ([7d878be](https://github.com/joao-zanutto/easydocker/commit/7d878be160ae5dc921f0ba137591ea8d95640fc8))
* update image sorting policy ([#41](https://github.com/joao-zanutto/easydocker/issues/41)) ([e5ef905](https://github.com/joao-zanutto/easydocker/commit/e5ef9052189a690b28ce74047751d7f812053a96))
* update spinners ([#26](https://github.com/joao-zanutto/easydocker/issues/26)) ([fa1a849](https://github.com/joao-zanutto/easydocker/commit/fa1a84942932b2899f118e08e1bbd44c6614efc3))


### Bug Fixes

* adjust spacing reintroduced by horizontal scroll indicator ([#39](https://github.com/joao-zanutto/easydocker/issues/39)) ([bf9e2d7](https://github.com/joao-zanutto/easydocker/commit/bf9e2d75267294e0dda37172f4abb332d644ac49))
* **ci:** refactor to prevent app/docs pipelines running when there are no changes on it ([#51](https://github.com/joao-zanutto/easydocker/issues/51)) ([a088613](https://github.com/joao-zanutto/easydocker/commit/a088613cc988b072d5caba795db6c72a3dca5403))
* **docs:** adjust gif location ([#46](https://github.com/joao-zanutto/easydocker/issues/46)) ([426ff32](https://github.com/joao-zanutto/easydocker/commit/426ff327af66b40fb5b96ab5246cb1da4b886567))
* **docs:** broken logo and gif ([#45](https://github.com/joao-zanutto/easydocker/issues/45)) ([5382601](https://github.com/joao-zanutto/easydocker/commit/538260140e25cb229614fe3a713110e38ac996bc))
* **docs:** gif reference in README ([#47](https://github.com/joao-zanutto/easydocker/issues/47)) ([24d0602](https://github.com/joao-zanutto/easydocker/commit/24d0602c59137d47c4bfe37cf96c7c0e3a650e1b))
* log loading inconsistency ([#29](https://github.com/joao-zanutto/easydocker/issues/29)) ([3326551](https://github.com/joao-zanutto/easydocker/commit/3326551c9b8e898db34893a06f01f8b3d9fb8b20))
* remove leftover spacing on right border of log viewport ([#35](https://github.com/joao-zanutto/easydocker/issues/35)) ([00a2289](https://github.com/joao-zanutto/easydocker/commit/00a228924351edc8746777abb5aaa800eff09558))

## [1.1.0](https://github.com/joao-zanutto/easydocker/compare/v1.0.0...v1.1.0) (2026-04-16)


### Features

* add resource filtering mode ([#21](https://github.com/joao-zanutto/easydocker/issues/21)) ([ead1438](https://github.com/joao-zanutto/easydocker/commit/ead1438387dcc02886a7409a473155b75c3276d6))
* implement log filtering ([#22](https://github.com/joao-zanutto/easydocker/issues/22)) ([65afc00](https://github.com/joao-zanutto/easydocker/commit/65afc0090e78e7d9bc01ba2a36c8d3f142f33ee2))
* replace static footer with bubbles help ([#16](https://github.com/joao-zanutto/easydocker/issues/16)) ([60a277d](https://github.com/joao-zanutto/easydocker/commit/60a277dfb6a4f8bddb53e9c7ad1cef6a9a88a7ae))


### Bug Fixes

* state coloring being disabled when the screen/column size is reduced ([#19](https://github.com/joao-zanutto/easydocker/issues/19)) ([bd92341](https://github.com/joao-zanutto/easydocker/commit/bd9234160e52e9367cb4d57a45226dcd7e2832fa))

## 1.0.0 (2026-04-12)


### Features

* add install script and refactor docs ([#8](https://github.com/joao-zanutto/easydocker/issues/8)) ([fd72123](https://github.com/joao-zanutto/easydocker/commit/fd721231db18d3928f78f0f857f67ab326bed67a))
* add vhs gif with tui usage and enrich docs ([#10](https://github.com/joao-zanutto/easydocker/issues/10)) ([057374a](https://github.com/joao-zanutto/easydocker/commit/057374a649cec6f0449146c9c09dc9fb09dc6d12))
* first release refactor ([#1](https://github.com/joao-zanutto/easydocker/issues/1)) ([6be1f69](https://github.com/joao-zanutto/easydocker/commit/6be1f69e821617aed91a9cf5fd8367b1b9cf3b0c))
* replace static loading with spinners and stabilize metrics ([#7](https://github.com/joao-zanutto/easydocker/issues/7)) ([267dce1](https://github.com/joao-zanutto/easydocker/commit/267dce1572e9dc23d19de4e660be79199e3748a6))


### Bug Fixes

* **ci:** run pr workflow only on go files ([#9](https://github.com/joao-zanutto/easydocker/issues/9)) ([c1d64e6](https://github.com/joao-zanutto/easydocker/commit/c1d64e63ba02ee603781f6eafc27327d6f0de405))
* delete changelog ([c2e7a1f](https://github.com/joao-zanutto/easydocker/commit/c2e7a1f14ae8679ea336506d36619de95ab6459d))
* gorelease version ([9d03238](https://github.com/joao-zanutto/easydocker/commit/9d03238c5dcf926faff6317721ea840c05abb16e))
* update release-please authentication to use pat ([#3](https://github.com/joao-zanutto/easydocker/issues/3)) ([1205c82](https://github.com/joao-zanutto/easydocker/commit/1205c823ae88ac83b0ee4121741e978c5bb40dff))

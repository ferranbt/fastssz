# 0.1.5 (Unreleased)

- fix: Array of fixed size of bytes with size in external package [GH-181](https://github.com/ferranbt/fastssz/pull/181)

# 0.1.4 (7 Aug, 2024)

- fix: Do not skip intermediate hashes in multi-proof [GH-173](https://github.com/ferranbt/fastssz/issues/173)]
- feat: Add dot graph generation [[GH-172](https://github.com/ferranbt/fastssz/issues/172)]
- fix: Fix spurious allocation in hasher.Merkleize [[GH-171](https://github.com/ferranbt/fastssz/issues/171)]
- feat: Increase performance for repeated proving [[GH-168](https://github.com/ferranbt/fastssz/issues/168)]
- fix: Infer size for fixed []byte without tags [[GH-155](https://github.com/ferranbt/fastssz/issues/155)]
- fix: Unmarshaling of fixed sized custom types [[GH-152](https://github.com/ferranbt/fastssz/issues/152)]
- feat: Support list of non-ptr containers [[GH-151](https://github.com/ferranbt/fastssz/issues/151)]
- feat: Support uin32 lists [[GH-149](https://github.com/ferranbt/fastssz/issues/149)]
- fix: Fix chunk count in merkleize [[GH-147](https://github.com/ferranbt/fastssz/issues/147)]
- feat: Add deneb fork to specs [[GH-139](https://github.com/ferranbt/fastssz/issues/139)]
- fix: Sszgen incorrect output for nested []byte types [[GH-127](https://github.com/ferranbt/fastssz/issues/127)]
- fix: Sszgen do not import package references if not used [[GH-137](https://github.com/ferranbt/fastssz/issues/137)]

# 0.1.3 (8 Feb, 2023)

- fix: Tree proof memory out of bounds [[GH-119](https://github.com/ferranbt/fastssz/issues/119)]
- fix: Double merkleization of `BeaconState` bytes field [[GH-119](https://github.com/ferranbt/fastssz/issues/119)]
- fix: Depth calculation for basic types merkleization (eip-4844) [[GH-111](https://github.com/ferranbt/fastssz/issues/111)]
- fix: Deterministic hash digest for generated file [[GH-110](https://github.com/ferranbt/fastssz/issues/110)]
- feat: Skip unit tests and `ssz` generate objects during parsing [[GH-114](https://github.com/ferranbt/fastssz/issues/114)]
- fix: Add support for non-literal array lengths [[GH-108](https://github.com/ferranbt/fastssz/issues/108)]
- feat: Add `--suffix` command line option [[GH-113](https://github.com/ferranbt/fastssz/issues/113)]

# 0.1.2 (26 Aug, 2022)

- feat: Add `HashFn` abstraction and introduce `gohashtree` hashing [[GH-95](https://github.com/ferranbt/fastssz/issues/95)]
- feat: `sszgen` for alias to byte array [[GH-55](https://github.com/ferranbt/fastssz/issues/55)]
- feat: `sszgen` include version in generated header file [[GH-101](https://github.com/ferranbt/fastssz/issues/101)]
- feat: support `time.Time` type as native object [[GH-100](https://github.com/ferranbt/fastssz/issues/100)]
- fix: Allocate nil data structures in HTR [[GH-98](https://github.com/ferranbt/fastssz/issues/98)]
- fix: Allocate uint slice if len is 0 instead of nil [[GH-96](https://github.com/ferranbt/fastssz/issues/96)]
- feat: Simplify the logic of the merkleizer [[GH-94](https://github.com/ferranbt/fastssz/issues/94)]

# 0.1.1 (1 July, 2022)

- Struct field not as a pointer [[GH-54](https://github.com/ferranbt/fastssz/issues/54)]
- Embed container structs [[GH-86](https://github.com/ferranbt/fastssz/issues/86)]
- Introduce `GetTree` to return the tree proof of the generated object [[GH-64](https://github.com/ferranbt/fastssz/issues/64)]
- Update to go `1.8` version [[GH-80](https://github.com/ferranbt/fastssz/issues/80)]
- Fix `alias` should not be considered objects but only used as types [[GH-76](https://github.com/ferranbt/fastssz/issues/76)]
- Fix the exclude of types from generation if they are set with the `exclude-objs` flag [[GH-76](https://github.com/ferranbt/fastssz/issues/76)]
- Add `version` command to `sszgen` [[GH-74](https://github.com/ferranbt/fastssz/issues/74)]
- Support `bellatrix`, `altair` and `phase0` forks in spec tests command to `sszgen` [[GH-73](https://github.com/ferranbt/fastssz/issues/73)]

# 0.1.0 (15 May, 2022)

- Initial public release.

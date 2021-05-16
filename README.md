Run EVM code against a database at a certain block height -

# Example

```
./run-evm-code \
	-blknum 1200000 \
	-db_dir ~/eth-mainnet-db/geth-attempt/geth/chaindata \
	-receiver 0xc02aaa39b223fe8d0a0e5c4f27ead9083c756cc2 \
	-sender 0x2485Aaa7C5453e04658378358f5E028150Dc7606 <ABI_ENCODED_DATA>
```

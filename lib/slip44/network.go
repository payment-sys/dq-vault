package slip44

// SLIP-0044 coin type constants for BIP-44 derivation paths
// Based on https://github.com/satoshilabs/slips/blob/master/slip-0044.md

const (
	// Bitcoin
	Bitcoin         uint16 = 0
	TestNet         uint16 = 1
	Litecoin        uint16 = 2
	Dogecoin        uint16 = 3
	Reddcoin        uint16 = 4
	Dash            uint16 = 5
	Peercoin        uint16 = 6
	Namecoin        uint16 = 7
	Feathercoin     uint16 = 8
	Counterparty    uint16 = 9
	Blackcoin       uint16 = 10
	NuShares        uint16 = 11
	NuBits          uint16 = 12
	Mazacoin        uint16 = 13
	Viacoin         uint16 = 14
	ClearingHouse   uint16 = 15
	Rubycoin        uint16 = 16
	Groestlcoin     uint16 = 17
	Digitalcoin     uint16 = 18
	Cannacoin       uint16 = 19
	DigiByte        uint16 = 20
	OpenAssets      uint16 = 21
	Monacoin        uint16 = 22
	Clams           uint16 = 23
	Primecoin       uint16 = 24
	Neoscoin        uint16 = 25
	Jumbucks        uint16 = 26
	ZiftCoin        uint16 = 27
	Vertcoin        uint16 = 28
	NXT             uint16 = 29
	Burst           uint16 = 30
	MonetaryUnit    uint16 = 31
	Zoom            uint16 = 32
	Vpncoin         uint16 = 33
	CanadaeCoin     uint16 = 34
	ShadowCash      uint16 = 35
	ParkByte        uint16 = 36
	Pandacoin       uint16 = 37
	StartCOIN       uint16 = 38
	MOIN            uint16 = 39
	Argentum        uint16 = 40
	Libertas        uint16 = 41
	Posw            uint16 = 42
	Shreeji         uint16 = 43
	GlobalCurrency  uint16 = 44
	Novacoin        uint16 = 45
	Asiacoin        uint16 = 46
	Bitcoindark     uint16 = 47
	Dopecoin        uint16 = 48
	Templecoin      uint16 = 49
	AIB             uint16 = 50
	EDRCoin         uint16 = 51
	Syscoin         uint16 = 52
	Solarcoin       uint16 = 53
	Smileycoin      uint16 = 54
	Ethereum        uint16 = 60
	Ether           uint16 = 60 // Alias for Ethereum
	EthereumClassic uint16 = 61
	Icap            uint16 = 62
	Expanse         uint16 = 63
	Olympus         uint16 = 64
	Ellaism         uint16 = 65
	Ethersocial     uint16 = 66
	Callisto        uint16 = 67
	Musicoin        uint16 = 68
	Bitshares       uint16 = 69
	Ubiq            uint16 = 70
	Klassic         uint16 = 71
	Boolberry       uint16 = 72
	Syscoin2        uint16 = 73
	Skycoin         uint16 = 74
	Nekonium        uint16 = 75
	Vivo            uint16 = 76
	Whitecoin       uint16 = 77
	Stratis         uint16 = 78
	Saga            uint16 = 79
	Pinkcoin        uint16 = 80
	Pivx            uint16 = 81
	Flashcoin       uint16 = 82
	Zencash         uint16 = 83
	Vcash           uint16 = 84
	Japan           uint16 = 85
	Megacoin        uint16 = 86
	Zcoin           uint16 = 87
	Zcash           uint16 = 88
	ZClassic        uint16 = 89
	Hush            uint16 = 90
	Komodo          uint16 = 91
	Hempcoin        uint16 = 92
	Spectrecoin     uint16 = 93
	Bitzeny         uint16 = 94
	Chainz          uint16 = 95
	Denarius        uint16 = 96
	Adcoin          uint16 = 97
	Particl         uint16 = 98
	Blocknet        uint16 = 99
	Whitecoin2      uint16 = 100
	Zoin            uint16 = 101
	Crave           uint16 = 102
	Ryo             uint16 = 103
	Arqma           uint16 = 104
	Bitcoinz        uint16 = 105
	Ritocoin        uint16 = 106
	Sugarchain      uint16 = 107
	Monero          uint16 = 128
	Aeon            uint16 = 129
	Sumokoin        uint16 = 130
	Loki            uint16 = 131
	Masari          uint16 = 132
	Grin            uint16 = 133
	Beam            uint16 = 134
	Tron            uint16 = 195
	Stellar         uint16 = 148
	Ripple          uint16 = 144
	Cardano         uint16 = 1815
	Cosmos          uint16 = 118
	Binance         uint16 = 714
	Polkadot        uint16 = 354
	Kusama          uint16 = 434
	Solana          uint16 = 501
	Avalanche       uint16 = 9000
	Polygon         uint16 = 966
	Fantom          uint16 = 1007
	Harmony         uint16 = 1023
	Near            uint16 = 397
	Algorand        uint16 = 283
	Filecoin        uint16 = 461
	Hedera          uint16 = 3030
	Theta           uint16 = 500
	Vechain         uint16 = 818
	Elrond          uint16 = 508
	Zilliqa         uint16 = 313
	Waves           uint16 = 5741
	Nano            uint16 = 165
	Iota            uint16 = 4218
	Ontology        uint16 = 1024
	Qtum            uint16 = 2301
	Icon            uint16 = 74
	Tezos           uint16 = 1729
	Chainlink       uint16 = 60 // Uses Ethereum's coin type
	Uniswap         uint16 = 60 // Uses Ethereum's coin type
	Compound        uint16 = 60 // Uses Ethereum's coin type
	Aave            uint16 = 60 // Uses Ethereum's coin type
	Yearn           uint16 = 60 // Uses Ethereum's coin type
	Sushi           uint16 = 60 // Uses Ethereum's coin type
	Curve           uint16 = 60 // Uses Ethereum's coin type
	Maker           uint16 = 60 // Uses Ethereum's coin type
	Synthetix       uint16 = 60 // Uses Ethereum's coin type
	TheGraph        uint16 = 60 // Uses Ethereum's coin type
	Bancor          uint16 = 60 // Uses Ethereum's coin type
	Balancer        uint16 = 60 // Uses Ethereum's coin type
	Inch            uint16 = 60 // Uses Ethereum's coin type (1inch)
	Enjin           uint16 = 60 // Uses Ethereum's coin type
	Sandbox         uint16 = 60 // Uses Ethereum's coin type
	Mana            uint16 = 60 // Uses Ethereum's coin type (Decentraland)
	Axie            uint16 = 60 // Uses Ethereum's coin type (Axie Infinity)
	Immutable       uint16 = 60 // Uses Ethereum's coin type
	Loopring        uint16 = 60 // Uses Ethereum's coin type
	Storj           uint16 = 60 // Uses Ethereum's coin type
	Basic           uint16 = 60 // Uses Ethereum's coin type (Basic Attention Token)
	Omisego         uint16 = 60 // Uses Ethereum's coin type
	Zrx             uint16 = 60 // Uses Ethereum's coin type (0x Protocol)
	Augur           uint16 = 60 // Uses Ethereum's coin type
	Golem           uint16 = 60 // Uses Ethereum's coin type
	Status          uint16 = 60 // Uses Ethereum's coin type
	Gnosis          uint16 = 60 // Uses Ethereum's coin type
	Dai             uint16 = 60 // Uses Ethereum's coin type
	Usdc            uint16 = 60 // Uses Ethereum's coin type
	Usdt            uint16 = 60 // Uses Ethereum's coin type
	Weth            uint16 = 60 // Uses Ethereum's coin type
	Wbtc            uint16 = 60 // Uses Ethereum's coin type
)

// IsTestnet returns true if the coin type is for a testnet
func IsTestnet(coinType uint16) bool {
	return coinType == TestNet
}

// GetCoinName returns the name of the coin for a given coin type
func GetCoinName(coinType uint16) string {
	switch coinType {
	case Bitcoin:
		return "Bitcoin"
	case TestNet:
		return "Bitcoin Testnet"
	case Litecoin:
		return "Litecoin"
	case Dogecoin:
		return "Dogecoin"
	case Ethereum: // Ether is an alias for Ethereum (both are 60)
		return "Ethereum"
	case EthereumClassic:
		return "Ethereum Classic"
	case Bitshares:
		return "BitShares"
	case Zcash:
		return "Zcash"
	case Monero:
		return "Monero"
	case Stellar:
		return "Stellar"
	case Ripple:
		return "Ripple"
	case Cardano:
		return "Cardano"
	case Cosmos:
		return "Cosmos"
	case Binance:
		return "Binance Chain"
	case Polkadot:
		return "Polkadot"
	case Solana:
		return "Solana"
	case Avalanche:
		return "Avalanche"
	case Polygon:
		return "Polygon"
	case Fantom:
		return "Fantom"
	case Harmony:
		return "Harmony"
	case Near:
		return "NEAR Protocol"
	case Algorand:
		return "Algorand"
	case Filecoin:
		return "Filecoin"
	case Tezos:
		return "Tezos"
	case Qtum:
		return "Qtum"
	case Icon:
		return "ICON"
	case Waves:
		return "Waves"
	case Nano:
		return "Nano"
	case Iota:
		return "IOTA"
	case Ontology:
		return "Ontology"
	case Zilliqa:
		return "Zilliqa"
	case Vechain:
		return "VeChain"
	case Theta:
		return "Theta"
	case Hedera:
		return "Hedera"
	case Elrond:
		return "Elrond"
	case Tron:
		return "Tron"
	case Kusama:
		return "Kusama"
	case Grin:
		return "Grin"
	case Beam:
		return "Beam"
	default:
		return "Unknown"
	}
}

// IsSupportedCoinType returns true if the coin type is supported
func IsSupportedCoinType(coinType uint16) bool {
	switch coinType {
	case Bitcoin, TestNet, Ethereum, EthereumClassic, Bitshares, Litecoin, Dogecoin, Zcash, Monero,
		Stellar, Ripple, Cardano, Cosmos, Binance, Polkadot, Solana, Avalanche, Polygon, Fantom,
		Harmony, Near, Algorand, Filecoin, Tezos, Qtum, Icon, Waves, Nano, Iota, Ontology, Zilliqa,
		Vechain, Theta, Hedera, Elrond, Tron, Kusama, Grin, Beam:
		return true
	default:
		return false
	}
}

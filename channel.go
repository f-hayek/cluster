package main

type Channel struct {
	shortChannelID string
	active         bool
	capacity       uint64
	localBalance   uint64
	remoteBalance  uint64
	commitFee      uint64
	localNodeID    string
	remoteNodeID   string
	remoteAlias    string
	outbound       uint64
	localBaseFee   uint64
	localFeeRate   uint64
	remoteBaseFee  uint64
	remoteFeeRate  uint64
	lastForward    float64
	opener         string
	localFees      uint64
	private        bool
	peerConnected  bool
}

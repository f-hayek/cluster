package main

type Channel struct {
	shortChannelID string
	active         bool
	capacity       int64
	localBalance   int64
	remoteBalance  int64
	commitFee      int64
	localNodeID    string
	remoteNodeID   string
	remoteAlias    string
	outbound       int64
	localBaseFee   int64
	localFeeRate   int64
	remoteBaseFee  int64
	remoteFeeRate  int64
	lastForward    float64
	opener         string
	localFees      int64
	private        bool
	peerConnected  bool
}

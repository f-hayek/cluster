package main

type Channel struct {
	shortChannelID string
	state          string
	active         bool
	capacity       int64
	localBalance   int64
	remoteBalance  int64
	commitFee      int64
	localNodeID    string
	remoteNodeID   string
	remoteAlias    string
	localBaseFee   int64
	localFeeRate   int64
	remoteBaseFee  int64
	remoteFeeRate  int64
	lastForward    float64
	opener         string
	localFees      int64
	remoteFees     int64
	private        bool
	peerConnected  bool
	block          int64
	age            int64
}

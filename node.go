package main

type OptionWillFund struct {
	leaseFeeBaseMsat                     int64
	leaseFeeBasis                        int64
	fundingWeight                        int64
	channelFeeMaxBaseMsat                int64
	channelFeeMaxProportionalThousandths int64
	compactLease                         string
}
type Node struct {
	id             string
	alias          string
	color          string
	blockheight    int64
	optionWillFund *OptionWillFund
}

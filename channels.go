package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"math"
	"sort"
	"strings"
	"time"
)

func getBalance(channel Channel) string {
	//fmt.Println("chan_id = ", channel.shortChannelID)
	//fmt.Println("localBalance = ", channel.localBalance)
	//fmt.Println("capacity = ", channel.capacity)
	send := int(10 * channel.localBalance / (channel.capacity - uint64(channel.commitFee)))
	recv := 10 - send
	bar := "[red]" + strings.Repeat(".", recv) +
		"[white]|" +
		"[green]" + strings.Repeat(".", send)
	return bar
}

func formatSats(v uint64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", v)
}

func formatDaysSince(ts float64) float64 {
	if ts > 0.0 {
		parsed := time.Unix(int64(ts), 0)
		return time.Since(parsed).Hours() / 24.0
	} else {
		return 0.0
	}

}
func channelsPage(ui *UI) *Table {
	t := NewTable()
	t.SetTitle(" Channels ")
	t.SetBorder(true)
	t.SetBorderColor(MainColor)
	t.SetSelectable(true, false)
	t.Select(4, 0).SetFixed(4, 12)
	t.AddColumnHeader("\n[bold]inbound", tview.AlignRight)
	t.AddColumnHeader("\nbalance", tview.AlignCenter)
	t.AddColumnHeader("\noutbound", tview.AlignRight)
	t.AddColumnHeader("local\nbase_fee\n(msat)", tview.AlignRight)
	t.AddColumnHeader("local\nfee_rate\n(ppm)", tview.AlignRight)
	t.AddColumnHeader("remote\nbase_fee\n(msat)", tview.AlignRight)
	t.AddColumnHeader("remote\nfee_rate\n(ppm)", tview.AlignRight)
	t.AddColumnHeader("last\nforward\n(days)", tview.AlignRight)
	t.AddColumnHeader("local\nfees\n(sat)", tview.AlignRight)
	t.AddColumnHeader("remote\nfees\n(estimate)", tview.AlignRight)
	t.AddColumnHeader("\nopener", tview.AlignCenter)
	t.AddColumnHeader("\nalias", tview.AlignLeft)
	t.AddHeaderSeparator()

	t.SetDoneFunc(func(key tcell.Key) {
		ui.pages.SwitchToPage("dash")
		ui.FocusMenu()
	})


	for row, channel := range getChannels(ui) {
		var opener string
		if channel.opener == "local" {
			opener = "[greenyellow]local"
		} else {
			opener = "[darkviolet]remote"
		}
		lastForward := formatDaysSince(channel.lastForward)
		var lastForwardFormatted string
		if lastForward > 0.0 {
			lastForwardFormatted = fmt.Sprintf("%.1f", lastForward)
		} else {
			lastForwardFormatted = "never"
		}
		t.SetCell(row+4, 0,
			tview.NewTableCell("[red]"+formatSats(channel.remoteBalance)).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 1,
			tview.NewTableCell(getBalance(channel)).SetAlign(tview.AlignCenter))
		t.SetCell(row+4, 2,
			tview.NewTableCell("[green]"+formatSats(channel.localBalance)).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 3,
			tview.NewTableCell("[deepskyblue]"+formatSats(uint64(channel.localBaseFee))).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 4,
			tview.NewTableCell("[deepskyblue]"+formatSats(uint64(channel.localFeeRate))).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 5,
			tview.NewTableCell("[lightyellow]"+formatSats(uint64(channel.remoteBaseFee))).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 6,
			tview.NewTableCell("[lightyellow]"+formatSats(uint64(channel.remoteFeeRate))).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 7,
			tview.NewTableCell("[grey]" + lastForwardFormatted).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 8,
			tview.NewTableCell("[lightcyan]" + formatSats(channel.localFees)).SetAlign(tview.AlignRight))
		t.SetCell(row+4, 10,
			tview.NewTableCell(opener).SetAlign(tview.AlignCenter))
		t.SetCell(row+4, 11,
			tview.NewTableCell("[white]"+channel.remoteAlias))
	}

	return t
}
func getChannels(ui *UI) []Channel {
	client := NewClient(ui)

	getInfo, err := client.Call("getinfo")
	if err != nil {
		ui.log.Warn("error: " + err.Error() + "\n")
	}

	localNode := Node{
		getInfo.Get("id").String(),
		getInfo.Get("alias").String(),
		getInfo.Get("color").String(),
	}

	ui.log.Info("listpeers ")

	peers, err := client.CallNamed("listpeers")
	if err != nil {
		ui.log.Warn(" error: " + err.Error() + "\n")
	}

	ui.log.Ok("OK\n")

	var channels []Channel

	forwards := getForwards(ui, map[string]interface{} {
		"status": "settled",
	}).Get("forwards.#.{resolved_time,in_channel,out_channel,fee}").Array()

	for _, peer := range peers.Get("peers").Array() {
		channel := peer.Get("channels.0")
		shortChannelID := channel.Get("short_channel_id").String()

		if shortChannelID == "" {
			continue
		}
		capacity := channel.Get("msatoshi_total").Uint() / 1000
		localBalance := channel.Get("msatoshi_to_us").Uint() / 1000
		lastTxFee := channel.Get("last_tx_fee").Uint()

		chanInfo := getChannel(ui, shortChannelID)
		chanLen := chanInfo.Get("channels.#").Uint()

		var localFee Fee
		var remoteFee Fee
		var node1Fee Fee
		var node2Fee Fee

		if chanLen > 0 {
			node1Fee = Fee{
				chanInfo.Get("channels.0.base_fee_millisatoshi").Uint(),
				chanInfo.Get("channels.0.fee_per_millionth").Uint(),
			}
			if chanLen > 1 {
				node2Fee = Fee{
					chanInfo.Get("channels.1.base_fee_millisatoshi").Uint(),
					chanInfo.Get("channels.1.fee_per_millionth").Uint(),
				}
				if localNode.id != chanInfo.Get("channels.0.source").String() {
					remoteFee = node1Fee
					localFee = node2Fee
				}

			}
			if chanLen > 1 {
				if localNode.id != chanInfo.Get("channels.1.source").String() {
					localFee = node1Fee
					remoteFee = node2Fee
				}
			} else {
				localFee = node1Fee
				remoteFee = Fee{0, 0}
			}

		}
		remoteNodeID := peer.Get("id").String()
		remoteAlias := getNode(ui, remoteNodeID).alias

		lastForward := 0.0
		fees := uint64(0)

		for _, forward := range forwards {
			inChan := forward.Get("in_channel").String()
			outChan := forward.Get("out_channel").String()
			// last forward
			if shortChannelID == inChan || shortChannelID == outChan {
				lastForward = math.Max(forward.Get("resolved_time").Float(), lastForward)
			}
			// local fees earned
			if shortChannelID == outChan {
				fees += forward.Get("fee").Uint() / 1000
			}
		}
		channels = append(channels, Channel{
			shortChannelID: channel.Get("short_channel_id").String(),
			active:         channel.Get("state").String() == "CHANNELD_NORMAL",
			opener:         channel.Get("opener").String(),
			localNodeID:    localNode.id,
			remoteNodeID:   remoteNodeID,
			remoteAlias:    remoteAlias,
			capacity:       capacity,
			localBalance:   localBalance,
			remoteBalance:  capacity - localBalance,
			commitFee:      lastTxFee,
			outbound:       200,
			localBaseFee:   localFee.base,
			localFeeRate:   localFee.rate,
			remoteBaseFee:  remoteFee.base,
			remoteFeeRate:  remoteFee.rate,
			lastForward:    lastForward,
			localFees:      fees,
		})

	}

	sort.Slice(channels, func(i, j int) bool {
		c1 := channels[i]
		c2 := channels[j]
		return float32(c1.localBalance)/float32(c1.capacity-c1.commitFee) < float32(c2.localBalance)/float32(c2.capacity-c2.commitFee)
	})
	return channels

}

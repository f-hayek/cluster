package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SortFunc func(channels []Channel) []Channel
type SortFuncs map[string]SortFunc

var sortFuncs = SortFuncs {
	"Channel balance": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.localBalance)/float32(c1.capacity-c1.commitFee) < float32(c2.localBalance)/float32(c2.capacity-c2.commitFee)
		})
		return channels
	},
	"Inbound liquidity": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.remoteBalance) > float32(c2.remoteBalance)
		})
		return channels
	},
	"Outbound liquidity": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.localBalance) > float32(c2.localBalance)
		})
		return channels
	},
	"Local base fee": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.localBaseFee) > float32(c2.localBaseFee)
		})
		return channels
	},
	"Local fee rate": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.localFeeRate) > float32(c2.localFeeRate)
		})
		return channels
	},
	"Remote base fee": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.remoteBaseFee) > float32(c2.remoteBaseFee)
		})
		return channels
	},
	"Remote fee rate": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.remoteFeeRate) > float32(c2.remoteFeeRate)
		})
		return channels
	},
	"Last forward": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.lastForward) > float32(c2.lastForward)
		})
		return channels
	},
	"Local fees earned": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.localFees) > float32(c2.localFees)
		})
		return channels
	},
	"Remote fees earned": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return float32(c1.remoteFees) > float32(c2.remoteFees)
		})
		return channels
	},
	"Alias": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return c1.remoteAlias < c2.remoteAlias
		})
		return channels
	},
	"Channel age (youngest first)": func(channels []Channel) []Channel {
		sort.Slice(channels, func(i, j int) bool {
			c1 := channels[i]
			c2 := channels[j]
			return c1.block > c2.block
		})
		return channels
	},

}

var selectedSortFunc = "Channel balance"

func getBalance(channel Channel) string {
	//fmt.Println("chan_id = ", channel.shortChannelID)
	//fmt.Println("localBalance = ", channel.localBalance)
	//fmt.Println("capacity = ", channel.capacity)
	send := int(10 * channel.localBalance / (channel.capacity - int64(channel.commitFee)))
	recv := 10 - send
	bar := "[red]" + strings.Repeat(".", recv) +
		"[white]|" +
		"[green]" + strings.Repeat(".", send)
	return bar
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

	t.AddColumnHeader("\n[bold]inbound", tview.AlignRight)
	t.AddColumnHeader("\nbalance", tview.AlignCenter)
	t.AddColumnHeader("\noutbound", tview.AlignRight)
	t.AddColumnHeader("local\nbase_fee\n(msat)", tview.AlignRight)
	t.AddColumnHeader("local\nfee_rate\n(ppm)", tview.AlignRight)
	t.AddColumnHeader("remote\nbase_fee\n(msat)", tview.AlignRight)
	t.AddColumnHeader("remote\nfee_rate\n(ppm)", tview.AlignRight)
	t.AddColumnHeader("last\nforward\n(days)", tview.AlignRight)
	t.AddColumnHeader("local\nfees earned\n(sat)", tview.AlignRight)
	t.AddColumnHeader("remote\nfees earned\n(estimate)", tview.AlignRight)
	t.AddColumnHeader("\nstatus", tview.AlignCenter)
	t.AddColumnHeader("\nage\n(blocks)", tview.AlignCenter)
	t.AddColumnHeader("\nalias", tview.AlignLeft)
	t.Separator(11)
	rowOffset := t.GetRowCount()
	t.Select(rowOffset, 0)
	t.SetFixed(rowOffset, t.GetColumnCount())

	t.SetDoneFunc(func(key tcell.Key) {
		ui.pages.SwitchToPage("dash")
		ui.FocusMenu()
	})

	channels := getChannels(ui)

	t.SetSelectedFunc(func(row, column int) {
		ui.log.Info("Selected channel id: " + fmt.Sprintf("%s", channels[row - rowOffset].shortChannelID) + "\n")
		ui.log.Info("Current commit fee: " + fmt.Sprintf("%d", channels[row - rowOffset].commitFee) + " sats\n")
		ui.log.Info("Remote base fee: " + fmt.Sprintf("%d", channels[row - rowOffset].remoteBaseFee) + "\n")
		ui.log.Info("Block: " + fmt.Sprintf("%d", channels[row - rowOffset].block) + "\n")

	})

	// Do not allow to select the header
	t.SetSelectionChangedFunc(func(row, column int) {
		if row < rowOffset {
			t.Select(row + 1, column)
		}
	})

	// Keyboard handler
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 's':
			if ui.HasPage("channelSort") {
				ui.DeletePage("channelSort")
			} else {
				ui.AddPage("channelSort", ui.NewChannelSortPage(), true, true)
				//ui.pages.SwitchToPage("channelSort")
				ui.SetFocus("channelSort")
			}
		case 'h':
			help := []string{
			"j/k   - Scroll down/up              ",
			"G/g   - Scroll to bottom/top        ",
			"Enter - Go to channel details page  ",
			"o     - Open new channel(s)         ",
			"c     - Close selected channel      ",
			"s     - Sort channels by            ",
			"h     - Toggle this help screen     ",
			"ESC   - Go back                     ",
			}
			if ui.HasPage("help") {
				ui.DeletePage("help")
			} else {
				ui.AddPage("help", ui.NewHelpPage(help), true, true)
			}

		}
		return event
	})

	totalInbound := int64(0)
	totalOutbound := int64(0)
	localFees := int64(0)
	remoteFees := int64(0)

	for row, channel := range channels {
		var state string
		switch channel.state {
		case "CHANNELD_AWAITING_LOCKIN":
			state = "[orange]opening"
		case "CHANNELD_NORMAL":
			if channel.peerConnected {
				state = "[green]online"
			} else {
				state = "[grey]offline"
			}
		case "AWAITING UNILATERAL":
			state = "[orange]awaiting unilateral"
		case "CHANNELD_SHUTTING_DOWN", "CLOSINGD_SIGEXCHANGE", "CLOSINGD_COMPLETE":
			state = "[lightgrey]closing"
		case "ONCHAIN":
			state = "[lightgrey]onchain"
		case "CLOSED":
			state = "[grey]closed"
		}

		var aliasColor string
		if channel.peerConnected {
			if channel.opener == "local" {
				aliasColor = "[greenyellow]"
			} else {
				aliasColor = "[darkviolet]"
			}
		} else {
			if channel.opener == "local" {
				aliasColor = "[#9DB27C]"
			} else {
				aliasColor = "[#71577C]"
			}

		}
		lastForward := formatDaysSince(channel.lastForward)
		var lastForwardFormatted string
		if lastForward > 0.0 {
			if lastForward > 60 {
				lastForwardFormatted = fmt.Sprintf("%s%.1f", "[red]", lastForward)
			} else {
				lastForwardFormatted = fmt.Sprintf("%s%.1f", "[white]", lastForward)
			}
		} else {
			lastForwardFormatted = "never"
		}

		currentRow := row + rowOffset
		t.SetCell(currentRow, 0,
			tview.NewTableCell("[red]"+formatSats(channel.remoteBalance)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 1,
			tview.NewTableCell(getBalance(channel)).SetAlign(tview.AlignCenter))
		t.SetCell(currentRow, 2,
			tview.NewTableCell("[green]"+formatSats(channel.localBalance)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 3,
			tview.NewTableCell("[deepskyblue]"+formatSats(channel.localBaseFee)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 4,
			tview.NewTableCell("[deepskyblue]"+formatSats(channel.localFeeRate)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 5,
			tview.NewTableCell("[lightyellow]"+formatSats(channel.remoteBaseFee)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 6,
			tview.NewTableCell("[lightyellow]"+formatSats(channel.remoteFeeRate)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 7,
			tview.NewTableCell(lastForwardFormatted).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 8,
			tview.NewTableCell("[deepskyblue]" + formatSats(channel.localFees)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 9,
			tview.NewTableCell("[lightyellow]" + formatSats(channel.remoteFees)).SetAlign(tview.AlignRight))
		t.SetCell(currentRow, 10,
			tview.NewTableCell(state).SetAlign(tview.AlignCenter))
		t.SetCell(currentRow, 11,
			tview.NewTableCell(fmt.Sprintf("%d", channel.age)).SetAlign(tview.AlignCenter))
		t.SetCell(currentRow, 12,
			tview.NewTableCell(aliasColor + channel.remoteAlias))

		t.Separator(11)

		// totals
		totalInbound += channel.remoteBalance
		totalOutbound += channel.localBalance
		localFees += channel.localFees
		remoteFees += channel.remoteFees


	}
	currentRow := t.GetRowCount()

	// Total inbound
	t.SetCell(currentRow, 0,
		tview.NewTableCell("[red]" + formatSats(totalInbound)).SetAlign(tview.AlignRight))
	// Total outbound
	t.SetCell(currentRow, 2,
		tview.NewTableCell("[green]" + formatSats(totalOutbound)).SetAlign(tview.AlignRight))
	// Total local fees
	t.SetCell(currentRow, 8,
		tview.NewTableCell("[deepskyblue]" + formatSats(localFees)).SetAlign(tview.AlignRight))
	// Total remote fees
	t.SetCell(currentRow, 9,
		tview.NewTableCell("[lightyellow]" + formatSats(remoteFees)).SetAlign(tview.AlignRight))

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
		getInfo.Get("blockheight").Int(),
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
	}).Get("forwards.#.{resolved_time,in_channel,out_channel,in_msatoshi,fee}").Array()

	for _, peer := range peers.Get("peers").Array() {
		channel := peer.Get("channels.0")
		peerConnected := peer.Get("connected").Bool()
		state := channel.Get("state").String()
		shortChannelID := channel.Get("short_channel_id").String()

		if shortChannelID == "" {
			continue
		}
		// extract the block from shortChannelID
		parsedBlock := strings.Split(shortChannelID, "x")
		var block int64
		var age int64
		if len(parsedBlock) == 3 {
			b, err := strconv.ParseInt(parsedBlock[0], 10, 64)
			if err == nil {
				block = b
				age = localNode.blockheight - block
			}
		}
		capacity := channel.Get("msatoshi_total").Int() / 1000
		localBalance := channel.Get("msatoshi_to_us").Int() / 1000
		lastTxFee, err := Mstoi(channel.Get("last_tx_fee").String())
		if err != nil {
			lastTxFee = 0
		} else {
			lastTxFee = lastTxFee / 1000
		}
		private := channel.Get("private").Bool()

		chanInfo := getChannel(ui, shortChannelID)
		chanLen := chanInfo.Get("channels.#").Uint()

		var localFee Fee
		var remoteFee Fee
		var node1Fee Fee
		var node2Fee Fee

		if chanLen > 0 {
			node1Fee = Fee{
				chanInfo.Get("channels.0.base_fee_millisatoshi").Int(),
				chanInfo.Get("channels.0.fee_per_millionth").Int(),
			}
			if chanLen > 1 {
				node2Fee = Fee{
					chanInfo.Get("channels.1.base_fee_millisatoshi").Int(),
					chanInfo.Get("channels.1.fee_per_millionth").Int(),
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
		remoteNode := getNode(ui, remoteNodeID)

		var remoteAlias string

		if remoteNode.alias != "" {
			remoteAlias = remoteNode.alias
		} else {
			remoteAlias = remoteNodeID
		}

		lastForward := 0.0
		localFees := int64(0)
		remoteFees := int64(0)

		for _, forward := range forwards {
			inChan := forward.Get("in_channel").String()
			outChan := forward.Get("out_channel").String()
			amountIn := forward.Get("in_msatoshi").Int() / 1000
			// last forward
			if shortChannelID == inChan || shortChannelID == outChan {
				lastForward = math.Max(forward.Get("resolved_time").Float(), lastForward)
			}
			// local fees earned
			if shortChannelID == outChan {
				localFees += forward.Get("fee").Int() / 1000
			}
			// remote fees (estimate if base/rate fee changed)
			if shortChannelID == inChan {
				remoteFees += (remoteFee.base + remoteFee.rate * amountIn / 1000) / 1000
			}
		}
		channels = append(channels, Channel{
			state:     state,
			shortChannelID: channel.Get("short_channel_id").String(),
			active:         state == "CHANNELD_NORMAL",
			opener:         channel.Get("opener").String(),
			localNodeID:    localNode.id,
			remoteNodeID:   remoteNodeID,
			remoteAlias:    remoteAlias,
			capacity:       capacity,
			localBalance:   localBalance,
			remoteBalance:  capacity - localBalance,
			commitFee:      lastTxFee,
			localBaseFee:   localFee.base,
			localFeeRate:   localFee.rate,
			remoteBaseFee:  remoteFee.base,
			remoteFeeRate:  remoteFee.rate,
			lastForward:    lastForward,
			localFees:      localFees,
			remoteFees:     remoteFees,
			private:        private,
			peerConnected:  peerConnected,
			block:          block,
			age:            age,
		})

	}

	return sortFuncs[selectedSortFunc](channels)

}

func (ui *UI) NewChannelSortPage() tview.Primitive {

	form := tview.NewForm()
	form.SetBorder(true)
	form.SetTitle(" Sort channels by ")
	form.SetBorderColor(MainColor)
	options := []string{
		"Channel balance",
		"Inbound liquidity",
		"Outbound liquidity",
		"Local base fee",
		"Local fee rate",
		"Remote base fee",
		"Remote fee rate",
		"Last forward",
		"Local fees earned",
		"Remote fees earned",
		"Remote alias",
		"Channel age (youngest first)",
	}

	initialOption := 0
	for idx, option := range options {
		if option == selectedSortFunc {
			initialOption = idx
			break
		}
	}

	form.AddDropDown("Order channels by ", options, initialOption, func(option string, optionIdx int) {
		if option != selectedSortFunc {
			selectedSortFunc = option
			ui.DeletePage("channelSort")
			ui.AddPage("channels", channelsPage(ui), true, true)
			ui.pages.SwitchToPage("channels")
			ui.SetFocus("channels")
		}
	})

	return ui.Modal(form, 80, 40)

}
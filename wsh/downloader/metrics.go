// Copyright 2015 The go-wiseplat Authors
// This file is part of the go-wiseplat library.
//
// The go-wiseplat library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wiseplat library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wiseplat library. If not, see <http://www.gnu.org/licenses/>.

// Contains the metrics collected by the downloader.

package downloader

import (
	"github.com/wiseplat/go-wiseplat/metrics"
)

var (
	headerInMeter      = metrics.NewMeter("wsh/downloader/headers/in")
	headerReqTimer     = metrics.NewTimer("wsh/downloader/headers/req")
	headerDropMeter    = metrics.NewMeter("wsh/downloader/headers/drop")
	headerTimeoutMeter = metrics.NewMeter("wsh/downloader/headers/timeout")

	bodyInMeter      = metrics.NewMeter("wsh/downloader/bodies/in")
	bodyReqTimer     = metrics.NewTimer("wsh/downloader/bodies/req")
	bodyDropMeter    = metrics.NewMeter("wsh/downloader/bodies/drop")
	bodyTimeoutMeter = metrics.NewMeter("wsh/downloader/bodies/timeout")

	receiptInMeter      = metrics.NewMeter("wsh/downloader/receipts/in")
	receiptReqTimer     = metrics.NewTimer("wsh/downloader/receipts/req")
	receiptDropMeter    = metrics.NewMeter("wsh/downloader/receipts/drop")
	receiptTimeoutMeter = metrics.NewMeter("wsh/downloader/receipts/timeout")

	stateInMeter   = metrics.NewMeter("wsh/downloader/states/in")
	stateDropMeter = metrics.NewMeter("wsh/downloader/states/drop")
)

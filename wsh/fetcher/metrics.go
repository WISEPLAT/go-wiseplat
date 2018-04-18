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

// Contains the metrics collected by the fetcher.

package fetcher

import (
	"github.com/wiseplat/go-wiseplat/metrics"
)

var (
	propAnnounceInMeter   = metrics.NewMeter("wsh/fetcher/prop/announces/in")
	propAnnounceOutTimer  = metrics.NewTimer("wsh/fetcher/prop/announces/out")
	propAnnounceDropMeter = metrics.NewMeter("wsh/fetcher/prop/announces/drop")
	propAnnounceDOSMeter  = metrics.NewMeter("wsh/fetcher/prop/announces/dos")

	propBroadcastInMeter   = metrics.NewMeter("wsh/fetcher/prop/broadcasts/in")
	propBroadcastOutTimer  = metrics.NewTimer("wsh/fetcher/prop/broadcasts/out")
	propBroadcastDropMeter = metrics.NewMeter("wsh/fetcher/prop/broadcasts/drop")
	propBroadcastDOSMeter  = metrics.NewMeter("wsh/fetcher/prop/broadcasts/dos")

	headerFetchMeter = metrics.NewMeter("wsh/fetcher/fetch/headers")
	bodyFetchMeter   = metrics.NewMeter("wsh/fetcher/fetch/bodies")

	headerFilterInMeter  = metrics.NewMeter("wsh/fetcher/filter/headers/in")
	headerFilterOutMeter = metrics.NewMeter("wsh/fetcher/filter/headers/out")
	bodyFilterInMeter    = metrics.NewMeter("wsh/fetcher/filter/bodies/in")
	bodyFilterOutMeter   = metrics.NewMeter("wsh/fetcher/filter/bodies/out")
)

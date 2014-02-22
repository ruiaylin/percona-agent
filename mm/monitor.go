/*
   Copyright (c) 2014, Percona LLC and/or its affiliates. All rights reserved.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>
*/

package mm

import (
	"time"
)

/**
 * A Monitor collects one or more Metric, usually many.  The MySQL monitor
 * (mysql/monitor.go) collects most SHOW STATUS variables, each as its own
 * Metric.  Each Metric collected during a single period are sent as a
 * Collection to an Aggregator (aggregator.go).  The Aggregator keeps Stats
 * for each unique Metric in a Metrics map/hash table.  When it's time to
 * report, the Stats are summarized and the Metrics are encoded in a Report
 * and sent to a Spooler (data/spooler.go).
 */

// Using given config, collect metrics when tickChan ticks, and send to collecitonChan.
type Monitor interface {
	Start(config []byte, tickChan chan time.Time, collectionChan chan *Collection) error
	Stop() error
	Status() map[string]string
	TickChan() chan time.Time
}

type MonitorFactory interface {
	Make(mtype, name string) (Monitor, error)
}

type Metric struct {
	Name   string  // mysql/status/Threads_running
	Type   byte    // see below
	Number float64 // Type=NUMBER|COUNTER
	String string  // Type=STRING
}

/**
 * Metric.Type is one of:
 *    NUMBER: standard metric type for which we calc full Stats: pct5, min, med, etc.
 *   COUNTER: value only increases or decreases; we only calc rate; e.g. Bytes_sent
 *    STRING: value is a string, used to collect config/setting values
 */
const (
	_ byte = iota
	NUMBER
	COUNTER
	STRING
)

type Collection struct {
	Ts      int64 // UTC Unix timestamp
	Metrics []Metric
}

type Metrics map[string]*Stats

type Report struct {
	Ts       time.Time // start, UTC
	Duration uint      // seconds
	Metrics  Metrics
}

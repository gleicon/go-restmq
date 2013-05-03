// Copyright 2009-2013 The RestMQ Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package restmq

var QUEUESET = "QUEUESET"
var UUID_SUFFIX = ":UUID"
var QUEUE_SUFFIX = ":queue"

func queue_name(qn string) string {
	return qn + QUEUE_SUFFIX
}

func queue_uuid(qn string) string {
	return qn + UUID_SUFFIX
}

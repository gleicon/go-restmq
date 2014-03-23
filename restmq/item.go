// Copyright 2009-2013 restmq authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package restmq

import (
	"encoding/json"
	"log"
	"net/http"
)

// Item is a queue item.
type Item struct {
	Id    int    `json:"id"`
	Count int    `json:"count"`
	Value string `json:"value"`
}

// WriteJSON writes a JSON representation of the Item into w.
// It should also set the Content-Type HTTP header.
func (i *Item) WriteJSON(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	return enc.Encode(i)
}

// JSON returns the JSON representation of the Item as a string.
func (i *Item) JSON() string {
	if b, err := json.Marshal(i); err != nil {
		log.Fatal("json.Marshal(Item) failed:", err)
	} else {
		return string(b)
	}
	return ""
}

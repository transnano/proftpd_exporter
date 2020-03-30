// Copyright 2020 Transnano
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"testing"
	"time"
)

func TestCalculateUptime1(t *testing.T) {
	s := "standalone FTP daemon [12257], up for 0 min"
	val := calculateUptime(s)
	if val != 0 {
		t.Errorf("calc: %0.f", val)
	}
}

func TestCalculateUptime2(t *testing.T) {
	s := "standalone FTP daemon [15697], up for  6 hrs 47 min"
	val := calculateUptime(s)
	if val != (6*time.Hour + 47*time.Minute).Seconds() {
		t.Errorf("calc: %0.f", val)
	}
}

func TestCalculateUptime3(t *testing.T) {
	s := "standalone FTP daemon [29875], up for 9 days,  2 hrs 14 min"
	val := calculateUptime(s)
	if val != (9*time.Hour*24 + 2*time.Hour + 14*time.Minute).Seconds() {
		t.Errorf("calc: %0.f", val)
	}
}

func TestCalculateUptime4(t *testing.T) {
	s := "standalone FTP daemon [15697], up for 78 days, 28 min"
	val := calculateUptime(s)
	if val != (78*time.Hour*24 + 0*time.Hour + 28*time.Minute).Seconds() {
		t.Errorf("calc: %0.f", val)
	}
}

func TestParseUptime1(t *testing.T) {
	s := "1677h1"
	val := parseUptime(s)
	if val != (1677*time.Hour + 1*time.Minute) {
		t.Errorf("calc: %0.s", val)
	}
}

func TestParseUptime2(t *testing.T) {
	s := "0m6s"
	val := parseUptime(s)
	if val != (6 * time.Second) {
		t.Errorf("calc: %0.s", val)
	}
}

func TestParseUptime3(t *testing.T) {
	s := "3h32m"
	val := parseUptime(s)
	if val != (3*time.Hour + 32*time.Minute) {
		t.Errorf("calc: %0.s", val)
	}
}

// Copyright 2022 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/palantir/tenablesc-client/tenablesc"
	"inet.af/netaddr"
)

// line-by-line logging for error discovery is far too interesting to not have loglines.
// Including an entire log framework to get .Caller() type functions also seemed like overkill, so
// welcome to the cheapest log formatting we could manage.

func init() {
	log.SetOutput(os.Stderr)
}

const (
	logDebug = "DEBUG"
	logTrace = "TRACE"
	logError = "ERROR"
	//logWarn  = "WARN"
	logInfo = "INFO"
)

var RecastAcceptRiskProtocolIDMap = map[string]string{
	"tcp":     "6",
	"udp":     "17",
	"icmp":    "1",
	"unknown": "0",
	"any":     "any",
}

// Intended for use by logging functions that want to identify where in code
// things are happening.
func trace(skip int) (string, int, string) {
	// if called with 0, we still want the function it's called from... not skip itself.
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return "?", 0, "?"
	}

	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return file, line, "?"
	}

	return file, line, fn.Name()
}

func Logf(level, message string, args ...interface{}) {
	skipLogf(1, level, message, args...)
}

func skipLogf(skip int, level, message string, args ...interface{}) {
	fileName, lineNumber, funcName := trace(skip + 1)

	formattedMessage := fmt.Sprintf(message, args...)

	for _, line := range strings.Split(formattedMessage, "\n") {
		log.Printf("[%s] %s:%d in %s, %s\n", level, fileName, lineNumber, funcName, line)
	}
}

// tenablesc client wraps errors that are for missing records in a custom struct.
// We will consistently want to treat not-found as a nonerror case so we can
// do basic drift handling and corrective plans.
func handleNotFoundError(d *schema.ResourceData, err error) diag.Diagnostics {
	if err != nil {
		nfe := tenablesc.NotFoundError{}

		if errors.As(err, &nfe) {
			// Someone already deleted it, not an error.
			d.SetId("")
			skipLogf(2, logInfo, "Got NotFoundError response, assuming resource has been deleted.")
			return nil
		}
		return diag.FromErr(err)
	}
	return nil
}

// DiffSuppressParsedTimes handles tenable's attempts to be incredibly helpful by
//
//	returning timestamps in local time instead of UTC.
const (
	tenableTime = time.RFC3339
)

func DiffSuppressParsedTimes(k, oldValue, newValue string, d *schema.ResourceData) bool {
	// 2022-12-31T16:00:00-08:00
	// 2023-01-01T00:00:00Z

	oldTime, err := time.Parse(tenableTime, oldValue)
	if err != nil {
		return false
	}

	newTime, err := time.Parse(tenableTime, newValue)
	if err != nil {
		return false
	}

	return oldTime.Equal(newTime)
}

// diffSuppressNormalizedIPSet takes an old and new string which _may_ be a set of IPs.
//
//	If at any point it fails to parse, it'll return false quietly.
//	Tenable's upstream libraries perform _some_ kind of hilarious normalization on IP sets.
//	it's our job to parse oldValue and newValue into IPSets and ask if they're equal.
func diffSuppressNormalizedIPSet(k, oldValue, newValue string, d *schema.ResourceData) bool {

	oldSet, err := buildIPSetForTenableFormat(oldValue)
	if err != nil {
		Logf(logDebug, err.Error())
		return false
	}

	newSet, err := buildIPSetForTenableFormat(newValue)
	if err != nil {
		Logf(logDebug, err.Error())
		return false
	}

	if oldSet.Equal(newSet) {
		return true
	}

	Logf(logDebug, "not equivalent ipsets: oldSet=%v, newset=%v", oldSet, newSet)

	return false
}

// IPSets may start with quotes. instead of trying to get clever, just split on quotes too.
var tenableIPSetDelimiter = regexp.MustCompile(`[",\n]`)

func buildIPSetForTenableFormat(iplist string) (*netaddr.IPSet, error) {

	ipSplit := tenableIPSetDelimiter.Split(iplist, -1)

	builder := &netaddr.IPSetBuilder{}

	for _, ipString := range ipSplit {
		if len(ipString) == 0 {
			// continuing the lazy logic, quotes at start and end will result in empty elements. these should
			// not error, just skip.
			continue
		}
		if ip, err := netaddr.ParseIP(ipString); err == nil {
			builder.Add(ip)
		} else if ipRange, err := netaddr.ParseIPRange(ipString); err == nil {
			builder.AddRange(ipRange)
		} else if ipPrefix, err := netaddr.ParseIPPrefix(ipString); err == nil {
			builder.AddPrefix(ipPrefix)
		} else {
			return nil, fmt.Errorf("unable to parse %s as ip, range, or prefix", ipString)
		}
	}

	return builder.IPSet()
}

func validateRecastAcceptRiskProtocol(protocol any, path cty.Path) (diags diag.Diagnostics) {

	if _, err := getRecastAcceptRiskProtocolID(protocol.(string)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("failed to get protocol id: %s", err),
			AttributePath: path,
		})
	}
	return
}

func DiffSuppressCase(k, old, new string, d *schema.ResourceData) bool {
	if strings.ToLower(old) == strings.ToLower(new) {
		return true
	}
	return false
}

func getRecastAcceptRiskProtocolID(protocol string) (string, error) {
	id, ok := RecastAcceptRiskProtocolIDMap[strings.ToLower(protocol)]
	if !ok {
		return "", fmt.Errorf("invalid protocol '%s'", protocol)
	}

	return id, nil
}

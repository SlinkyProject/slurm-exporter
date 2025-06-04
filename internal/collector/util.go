// SPDX-FileCopyrightText: Copyright (C) SchedMD LLC.
// SPDX-License-Identifier: Apache-2.0

package collector

import (
	"fmt"
	"math"

	"k8s.io/utils/ptr"

	api "github.com/SlinkyProject/slurm-client/api/v0041"
)

func ParseUint64NoVal(noVal *api.V0041Uint64NoValStruct) uint64 {
	if noVal == nil {
		return 0
	}
	if isSet := ptr.Deref(noVal.Set, false); !isSet {
		return 0
	}
	if isInfinite := ptr.Deref(noVal.Infinite, false); isInfinite {
		return math.MaxUint64
	}
	number := uint64(ptr.Deref(noVal.Number, 0))
	return number
}

func ParseUint32NoVal(noVal *api.V0041Uint32NoValStruct) uint32 {
	if noVal == nil {
		return 0
	}
	if isSet := ptr.Deref(noVal.Set, false); !isSet {
		return 0
	}
	if isInfinite := ptr.Deref(noVal.Infinite, false); isInfinite {
		return math.MaxUint32
	}
	number := uint32(ptr.Deref(noVal.Number, 0))
	return number
}

// parseNodeList parses Slurm node list formats
// Examples:
//   - "node1" -> ["node1"]
//   - "node1,node2" -> ["node1", "node2"]
//   - "node[1-3]" -> ["node1", "node2", "node3"]
//   - "node[1,3,5]" -> ["node1", "node3", "node5"]
//   - "node-[913,928,955]" -> ["node-913", "node-928", "node-955"]
func parseNodeList(nodeList string) []string {
	var nodes []string

	// Check if it contains bracket notation
	bracketStart := strings.Index(nodeList, "[")
	bracketEnd := strings.Index(nodeList, "]")

	if bracketStart != -1 && bracketEnd != -1 && bracketEnd > bracketStart {
		// Extract prefix and suffix
		prefix := nodeList[:bracketStart]
		suffix := ""
		if bracketEnd+1 < len(nodeList) {
			suffix = nodeList[bracketEnd+1:]
		}

		// Extract the content inside brackets
		bracketContent := nodeList[bracketStart+1 : bracketEnd]

		// Parse the bracket content
		parts := strings.Split(bracketContent, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.Contains(part, "-") {
				// Handle range notation like "1-3"
				rangeParts := strings.Split(part, "-")
				if len(rangeParts) == 2 {
					start := 0
					end := 0
					_, err1 := fmt.Sscanf(rangeParts[0], "%d", &start)
					_, err2 := fmt.Sscanf(rangeParts[1], "%d", &end)
					if err1 == nil && err2 == nil {
						for i := start; i <= end; i++ {
							nodes = append(nodes, fmt.Sprintf("%s%d%s", prefix, i, suffix))
						}
					}
				}
			} else {
				// Single number
				nodes = append(nodes, fmt.Sprintf("%s%s%s", prefix, part, suffix))
			}
		}
	} else if strings.Contains(nodeList, ",") {
		// Simple comma-separated list
		parts := strings.Split(nodeList, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part != "" {
				nodes = append(nodes, part)
			}
		}
	} else {
		// Single node
		nodeList = strings.TrimSpace(nodeList)
		if nodeList != "" {
			nodes = append(nodes, nodeList)
		}
	}

	return nodes
}

package git

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// ComputeMerkle builds a Merkle tree from the provided files and their
// contents, returning the hex-encoded root hash.  The same inputs always
// produce the same hash; changing any file path or its content produces a
// different hash.  Returns an empty string when files is empty.
func ComputeMerkle(files []string, contents map[string][]byte) string {
	if len(files) == 0 {
		return ""
	}

	// Deterministic ordering.
	sorted := make([]string, len(files))
	copy(sorted, files)
	sort.Strings(sorted)

	// Leaf hashes: SHA-256(path + content).
	leaves := make([][]byte, len(sorted))
	for i, path := range sorted {
		h := sha256.New()
		h.Write([]byte(path))
		h.Write(contents[path])
		leaves[i] = h.Sum(nil)
	}

	// Build tree bottom-up.
	layer := leaves
	for len(layer) > 1 {
		layer = merkleLayer(layer)
	}

	return hex.EncodeToString(layer[0])
}

// merkleLayer reduces a slice of hashes to the next tree level by pairing
// adjacent nodes.  An odd node is promoted unchanged.
func merkleLayer(nodes [][]byte) [][]byte {
	next := make([][]byte, 0, (len(nodes)+1)/2)
	for i := 0; i < len(nodes); i += 2 {
		if i+1 == len(nodes) {
			// Odd node: promote as-is.
			next = append(next, nodes[i])
			continue
		}
		h := sha256.New()
		h.Write(nodes[i])
		h.Write(nodes[i+1])
		next = append(next, h.Sum(nil))
	}
	return next
}

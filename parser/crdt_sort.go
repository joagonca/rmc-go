package parser

import (
	"container/heap"
)

// sortedItemsKey is used by the priority queue to maintain deterministic ordering
// when multiple items become ready simultaneously. Matches rmscene's use of heapq.
type sortedItemsKey struct {
	id    CrdtID
	isEnd bool // "__end" marker sorts after all real IDs
}

// compareKey returns true if a should come before b.
func compareKey(a, b sortedItemsKey) bool {
	if a.isEnd != b.isEnd {
		return !a.isEnd // non-end before end
	}
	if a.id.Part1 != b.id.Part1 {
		return a.id.Part1 < b.id.Part1
	}
	return a.id.Part2 < b.id.Part2
}

type keyHeap []sortedItemsKey

func (h keyHeap) Len() int            { return len(h) }
func (h keyHeap) Less(i, j int) bool  { return compareKey(h[i], h[j]) }
func (h keyHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *keyHeap) Push(x interface{}) { *h = append(*h, x.(sortedItemsKey)) }
func (h *keyHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[:n-1]
	return x
}

// SortedItems returns the sequence's items in CRDT-resolved logical order
// (topological sort via Kahn's algorithm using LeftID/RightID pointers).
//
// reMarkable's .rm v6 format stores scene items as a CRDT sequence where
// physical storage order does NOT equal render order. Each item points to
// its left/right neighbors in the logical sequence, and the correct order
// is obtained by topological sort.
//
// Ported from rmscene/crdt_sequence.py toposort_items().
func (cs *CrdtSequence) SortedItems() []CrdtSequenceItem {
	if cs == nil || len(cs.Items) == 0 {
		return nil
	}

	// Build lookup of known item IDs.
	itemByID := make(map[CrdtID]CrdtSequenceItem, len(cs.Items))
	for _, item := range cs.Items {
		itemByID[item.ItemID] = item
	}

	// Resolve a side reference to either a known item ID or a virtual start/end marker.
	// Unknown refs (not in itemByID) collapse to start (for left) or end (for right).
	resolveSide := func(item CrdtSequenceItem, side string) sortedItemsKey {
		var sideID CrdtID
		if side == "left" {
			sideID = item.LeftID
		} else {
			sideID = item.RightID
		}
		if _, ok := itemByID[sideID]; ok {
			return sortedItemsKey{id: sideID, isEnd: false}
		}
		// Unknown → virtual start (for left) or end (for right)
		return sortedItemsKey{isEnd: side == "right"}
	}

	startKey := sortedItemsKey{isEnd: false} // virtual start (zero CrdtID, isEnd=false)
	endKey := sortedItemsKey{isEnd: true}

	// Build dependency graph
	inDegree := make(map[sortedItemsKey]int)
	dependents := make(map[sortedItemsKey][]sortedItemsKey)

	// Ensure start and end exist in graph
	inDegree[startKey] = 0
	inDegree[endKey] = 0

	for _, item := range cs.Items {
		itemKey := sortedItemsKey{id: item.ItemID, isEnd: false}
		leftKey := resolveSide(item, "left")
		rightKey := resolveSide(item, "right")

		// item depends on left (item comes after left)
		inDegree[itemKey]++
		dependents[leftKey] = append(dependents[leftKey], itemKey)

		// right depends on item (right comes after item)
		inDegree[rightKey]++
		dependents[itemKey] = append(dependents[itemKey], rightKey)
	}

	// Kahn's algorithm with priority queue for determinism
	ready := &keyHeap{}
	heap.Init(ready)
	heap.Push(ready, startKey)

	var resultIDs []CrdtID
	processed := make(map[sortedItemsKey]bool)

	for ready.Len() > 0 {
		current := heap.Pop(ready).(sortedItemsKey)
		if processed[current] {
			continue
		}
		processed[current] = true

		// Record real item IDs in order (skip virtual start/end markers)
		if !current.isEnd && !(current.id.Part1 == 0 && current.id.Part2 == 0 && !keyExists(itemByID, current.id)) {
			if _, ok := itemByID[current.id]; ok {
				resultIDs = append(resultIDs, current.id)
			}
		}

		for _, dep := range dependents[current] {
			inDegree[dep]--
			if inDegree[dep] == 0 {
				heap.Push(ready, dep)
			}
		}
	}

	// Map back to CrdtSequenceItem
	result := make([]CrdtSequenceItem, 0, len(resultIDs))
	for _, id := range resultIDs {
		result = append(result, itemByID[id])
	}
	return result
}

func keyExists(m map[CrdtID]CrdtSequenceItem, id CrdtID) bool {
	_, ok := m[id]
	return ok
}

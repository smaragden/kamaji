package kamaji

import (
    "sort"
    "time"
)

type SortableItem interface {
    GetPrio() int
    GetCreated() time.Time
}


func prio(c1, c2 SortableItem) bool {
    return c1.GetPrio() > c2.GetPrio()
}

func created(c1, c2 SortableItem) bool {
    return c1.GetCreated().UnixNano() > c2.GetCreated().UnixNano()
}

type lessFunc func(p1 SortableItem, p2 SortableItem) bool

// multiSorter implements the Sort interface, sorting the changes within.
type multiSorter struct {
    jobs []SortableItem
    less []lessFunc
}

// Sort sorts the argument slice according to the less functions passed to OrderedBy.
func (ms *multiSorter) Sort(jobs []SortableItem) {
    ms.jobs = jobs
    sort.Sort(ms)
}

// OrderedBy returns a Sorter that sorts using the less functions, in order.
// Call its Sort method to sort the data.
func OrderedBy(less ...lessFunc) *multiSorter {
    return &multiSorter{
        less: less,
    }
}

// Len is part of sort.Interface.
func (ms *multiSorter) Len() int {
    return len(ms.jobs)
}

// Swap is part of sort.Interface.
func (ms *multiSorter) Swap(i, j int) {
    ms.jobs[i], ms.jobs[j] = ms.jobs[j], ms.jobs[i]
}

// Less is part of sort.Interface. It is implemented by looping along the
// less functions until it finds a comparison that is either Less or
// !Less. Note that it can call the less functions twice per call. We
// could change the functions to return -1, 0, 1 and reduce the
// number of calls for greater efficiency: an exercise for the reader.
func (ms *multiSorter) Less(i, j int) bool {
    p, q := ms.jobs[i], ms.jobs[j]
    // Try all but the last comparison.
    var k int
    for k = 0; k < len(ms.less) - 1; k++ {
        less := ms.less[k]
        switch {
        case less(p, q):
            // p < q, so we have a decision.
            return true
        case less(q, p):
            // p > q, so we have a decision.
            return false
        }
        // p == q; try the next comparison.
    }
    // All comparisons to here said "equal", so just return whatever
    // the final comparison reports.
    return ms.less[k](p, q)
}

package chord

import "math/big"

func stabilize() {

}

func fixFingers() {

}

// returns true if elt is between start and end, accounting for the right
// if inclusive is true, it can match the end
func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

// while looping maintenance tasks
func runMaintenance() {
	// kör stablize fixFingers jämna mellan rum
}

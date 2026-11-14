// Package proc reads the process tree, so lightsout can tell which app it is
// running inside.
package proc

// maxDepth bounds Ancestors, in case the process tree reports a cycle.
const maxDepth = 64

// Ancestors returns start followed by its parent, grandparent and so on,
// looked up with parent. It stops at PID 1 (or 0), when a lookup fails, or
// when a PID repeats, so it always returns at least start.
func Ancestors(start int, parent func(pid int) (int, error)) []int {
	chain := []int{start}
	seen := map[int]bool{start: true}
	for pid := start; len(chain) < maxDepth; {
		next, err := parent(pid)
		if err != nil || next <= 1 || seen[next] {
			break
		}
		chain = append(chain, next)
		seen[next] = true
		pid = next
	}
	return chain
}

// Package game implements planning-poker rules without I/O, serialization, clocks
// or concurrency. The hub owns each Room in a single goroutine.
//
// View omits card values until reveal; domain errors are exported sentinels for
// callers to map to their own protocols.
package game

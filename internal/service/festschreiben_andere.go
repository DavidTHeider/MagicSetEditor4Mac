//go:build !unix

package service

// leichtFestschreiben ist außerhalb von Unix dasselbe wie das volle: dort
// gibt es kein schwächeres (`docs/PORTIERUNG.md`).
var leichtFestschreiben = func(d nachbardatei) error { return d.Sync() }

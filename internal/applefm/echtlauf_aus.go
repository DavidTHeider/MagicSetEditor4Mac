//go:build !fmlauf

package applefm

// echtlauf sagt, ob dieser Bau `fm` in einem Testlauf starten darf — **im
// gewöhnlichen Bau nie**. Der Echtlauf übersetzt mit `-tags fmlauf`.
const echtlauf = false

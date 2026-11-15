//go:build !darwin

package app

import "io"

// systemprotokoll gibt es außerhalb von macOS nicht.
//
// **Und das ist der Stand von §12**, keine Lücke: gebaut wird für
// darwin/arm64. Für Linux beschreibt `docs/PORTIERUNG.md` den Weg über
// journald — beschrieben, nicht gebaut, nicht geprüft.
//
// Die **Datei** schreibt überall; sie ist die Senke, auf die Verlass ist.
func systemprotokoll() io.Writer { return io.Discard }

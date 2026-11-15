package app

import (
	"runtime"
	"strings"
)

// **Wer rief — für die ausführliche Stufe des Protokolls.**
//
// Aus `profilnaht.go` herausgelöst, als sie mit der abbrechbaren Sperre
// (Zweitblick über A und B, Befund 4) über 400 Zeilen ging (§2).

// imTakt sagt, ob das Fenster diese Bindung **von selbst und im Takt** fragt
// — Nachlauf über Phase 6.4, N5. `Datenstand` alle zwei Sekunden
// (`useBestandsstand.ts`) gab rund 43.000 Zeilen am Tag, und die
// ausführliche Stufe hält höchstens 1 MB je Datei. Ein Ruf, den niemand
// ausgelöst hat, erklärt nichts, was jemand sucht.
func imTakt(stelle string) bool {
	return strings.HasSuffix(stelle, ".Datenstand")
}

// aufrufer nennt die Bindung, die [App.imProfil] rief — ohne ihr Paket.
func aufrufer() string {
	stelle, _, _, _ := runtime.Caller(2)
	name := runtime.FuncForPC(stelle).Name()

	return name[strings.LastIndex(name, "/")+1:]
}

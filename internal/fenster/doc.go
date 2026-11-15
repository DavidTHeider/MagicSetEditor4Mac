// Package fenster ist die Hülle um Wails 3 — ENG-02
// (#1727,
// Vermerk 0132).
//
// # Was es tut
//
// Es öffnet das Fenster und setzt die Nähte der Anwendung darauf: die
// Menüleiste aus dem Modell von `internal/app`, Meldungen, Dialoge, die
// Zwischenablage, das Ablegen von Dateien, den Bildweg und den gemerkten
// Fensterstand. `internal/app` beschreibt, **was** das Fenster tun soll; hier
// steht, **wie** Wails es tut.
//
// # Warum eine eigene Schicht
//
// Unter Wails 2 standen die Rufe in `internal/app`, und jeder Lauf über die
// Menüleiste las Wails' Verbünde. Wails 3 gibt einen Menüeintrag so nicht
// mehr her — und baut unter Linux nur mit cgo; eine Anwendungsschicht, die es
// importierte, fiele aus dem Kreuzbau der Pipeline. Nur `main.go` kennt diese
// Schicht (§2).
//
// # Was gemessen wird und was nicht
//
// Alles, was entscheidet, steht in eigenen Dateien und wird zu 100 %
// gemessen: das Umsetzen der Leiste (`menue.go`), Fensteroptionen und
// Ereigniswerte (`ereignisse.go`) und die Verdrahtung (`verdrahtung.go`); die
// Weiche zum Bildweg steht bei den Bildwegen (`vorschauweg.Weg.Vor`), damit
// diese Schicht kein `net/http` braucht (#54). **Allein `wails.go` greift ins
// laufende Fenster** und ist in `.covignore` begründet; was dort geschieht,
// zeigt nur das Fenster — die Liste der Abnahmen steht in `docs/WAILS3.md`.
package fenster

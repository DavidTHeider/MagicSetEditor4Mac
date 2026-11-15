// Package app ist die Fassade zu den Wails-Bindings (AGENTS.md §2).
//
// Es übersetzt zwischen Wails und den Services: DTOs, Fehler, Kontext. Eine
// fachliche Entscheidung gehört hier nicht hinein — ein `if` über einen
// Domänenzustand ist ein Befund, und `internal/schichten` kann ihn nicht
// finden, weil er in keinem Importgraphen steht.
//
// Die eine benannte Ausnahme ist der Composition Root: hier entstehen die
// Repositories und werden den Services gereicht. Irgendwo muss die Verdrahtung
// stehen; sie an einer Stelle zu halten ist der Grund, warum die übrigen
// Schichten einander nicht kennen müssen.
//
// Was dieses Paket **nicht** enthält: die Menüleiste ist eine Ausnahme von der
// Regel „keine Oberfläche in Go". Sie steht hier, weil sie unter macOS nativ
// ist und Wails sie als Datenstruktur entgegennimmt — und weil sie sich damit
// prüfen lässt, ohne ein Fenster zu öffnen.
package app

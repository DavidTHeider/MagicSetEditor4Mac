# 0144 — Keine neue Endung für Deck und Sammlung; der Umschlag wird erklärt

**Stand:** entschieden am 2026-09-29 · **Betrifft:** FMT-04
(#1693),
AK-02 · **Ergänzt** [0061](0061-was-ein-deck-hier-ist.md),
[0052](0052-die-sammlung-ist-nicht-das-set.md)

## Die Frage

FMT-04.AK-02 erlaubt „Deckdatei" und „Sammlungsdatei" als verständliche
Bezeichnungen an der Oberfläche und verlangt für **neue** Endungen einen
definierten Behälter, eine Registrierung und ein Kompatibilitätsverhalten.
Braucht es eine neue Endung — etwa `.mse-deck` —, und stimmt, was es schon
gibt?

## Nachgesehen am 2026-09-29

| Was | Behälter | Registriert | Kompatibilität |
| --- | --- | --- | --- |
| **Umschlag** `.msepackage` (N4, #1457) — trägt ein Deck, eine Sammlungssicht, eine Karte, ein Set, eine Vorlage oder eine Leihe | ZIP mit Manifest, Prüfsummen und Unterschrift (`internal/austausch`) | **nein** — `build/darwin/Info.plist` kannte nur `.mse-set` | `austausch.Fassung`: eine fremde Fassung wird abgewiesen, nicht heruntergehandelt |
| **Sammlung als CSV** | Text, Spalten in `service/sammlungsausfuhr.go` | als Allerweltstyp | eine unbekannte Spalte wird übergangen; eine fehlende Pflichtspalte ist ein Befund |
| **Deckliste** als Text | die übliche „Menge Name (SET) Nummer"-Form | — | gedeutet, nicht geraten; was nicht eindeutig ist, steht als Frage da (T8, #1473) |

**MSE2 kennt weder Deck noch Sammlung** — nachgesehen in [0061](0061-was-ein-deck-hier-ist.md)
und [0052](0052-die-sammlung-ist-nicht-das-set.md), in `src/` der Fassung
`f9d9356`. Eine Endung der Gestalt `.mse-deck` versprach also etwas, das es
nicht gibt: MSE2 öffnete sie nicht.

## Die Entscheidung

1. **Keine neue Endung in 0.6.0.** „Deckdatei" heißt an der Oberfläche der
   Umschlag mit einem Deck, „Sammlungsdatei" die CSV oder der Umschlag mit
   einer Sammlungssicht. Die Behälter haben Manifest, Fassung und Grenzen —
   eine zweite Endung für dasselbe wäre eine zweite Verabredung.
2. **Der Umschlag wird im Bündel erklärt** (`UTExportedTypeDeclarations`,
   `com.wails.MagicSetEditor4Mac.umschlag`, Beschreibung
   „MagicSetEditor4Mac-Umschlag (kein MSE2-Paket)"). Seine Endung trägt
   „mse" im Namen und sah im Finder wie ein MSE-Paket aus — genau die
   **Scheinkompatibilität**, gegen die FMT-04 steht. Ein Wächter hält jede
   eigene Endung gegen das `Info.plist`
   (`internal/werkzeuge/eigeneendungen_test.go`).
3. **`.msepackage` bleibt.** Umbenennen bräche jeden Umschlag, der seit 0.5.0
   verschickt wurde; die Erklärung im Bündel nimmt dem Namen die Täuschung,
   ohne ihn zu ändern.
4. **Kein Doppelklick-Öffnen.** Ein Umschlag wird unter *Sets* angenommen —
   mit Unterschrift, Quarantäne und der Frage, wohin (N4). Ein Öffnen aus dem
   Finder ginge an diesem Weg vorbei; deshalb steht der Typ **nicht** in
   `CFBundleDocumentTypes`.

## Verworfen

- **`.mse-deck` und `.mse-sammlung`** — siehe oben: der Name verspräche
  MSE2-Kompatibilität, die es nicht gibt.
- **`.msepackage` umbenennen** — der Preis wären alle vorhandenen Umschläge,
  der Gewinn ein Name.

## Was daraus folgt

Die Abbildung zwischen Kartendefinition, Deckplatz und Sammlungsstück steht
in [`docs/DECKUNDSAMMLUNG.md`](../DECKUNDSAMMLUNG.md), samt dem, was welcher
Behälter trägt und was nicht.

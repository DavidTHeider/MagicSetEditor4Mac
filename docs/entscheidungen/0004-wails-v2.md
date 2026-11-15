# 0004 — Wails v2 statt v3

**Stand:** angenommen
**Datum:** 2026-08-08

## Lage

Die Anwendung ist eine Desktop-App mit Go im Rücken und einem WebView vorn.
Wails ist dafür der naheliegende Rahmen. Es gibt zwei Reihen: v2 als stabile
Fassung und v3, das mehr Fenster- und Menüfähigkeiten mitbringt.

Ein Schwesterprojekt derselben Bauart (Projektkompass) arbeitet seit über einem
Jahr mit v2. Werkzeugpinnung, Pipeline, Bindings-Erzeugung und Release-Weg sind
dort erprobt und lassen sich übernehmen.

## Entscheidung

**Wails v2.**

Die Fähigkeiten, die v3 hinzufügt — mehrere Fenster, reichere Menüs,
Systemtray —, sind hier nicht der Engpass. Der Engpass ist der
Skript-Interpreter (Phase 3) und die Kartenausgabe (Phase 4), und beide sind von
der Wahl des Fensterrahmens unberührt.

## Verworfen

**Wails v3.** Verworfen, nicht abgelehnt: die Mehrfenstertauglichkeit wäre für
eine Anwendung, in der man mehrere Sets nebeneinander bearbeitet, ein echter
Gewinn. Aber sie wäre in Phase 5 zu haben, und bis dahin kostete jede
Unwägbarkeit einer jungen Fassung Zeit an einer Stelle, an der ohnehin genug
Unbekanntes liegt.

**Tauri.** Verworfen — der Rücken wäre Rust, und die fachliche Entscheidung für
Go steht bereits (SQLite ohne cgo, geteilte Werkzeugkette mit dem
Schwesterprojekt).

**Electron.** Verworfen — Bündelgröße und Speicherbedarf stehen in keinem
Verhältnis, und der Rücken müsste dann ebenfalls in JavaScript liegen oder als
zweiter Prozess danebenlaufen.

**Eine native Oberfläche in Swift.** Verworfen. Sie wäre auf dem Mac die
schönste Lösung und macht `docs/PORTIERUNG.md` gegenstandslos — Windows und
Linux wären damit endgültig aus. Da sie ausdrücklich beschrieben bleiben sollen,
fällt diese Möglichkeit weg.

## Folgen

- Ein Fenster. Mehrere Sets nebeneinander sind in Phase 5 über die Oberfläche zu
  lösen, nicht über Fenster.
- Die macOS-Menüleiste kommt aus Wails' Menü-API. Was sie nicht kann, wird nicht
  nachgebaut.
- Der Wails-CLI liegt in einem **eigenen** Go-Modul (`tools/wails/`): seine
  Abhängigkeiten kollidieren mit denen von golangci-lint.
- Ein Wechsel auf v3 bleibt möglich und wäre ein eigener Vermerk hier — er
  beträfe `app/`, `main.go` und den Bau, nicht die Fachschichten. Dass er das
  nicht mehr täte, wäre ein Zeichen, dass §2 verletzt ist.

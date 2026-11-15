# 0180 — Der Pfadvertrag bleibt am Rohtext

**Stand:** angenommen am 2026-10-07 · SEC-02
(#2137).

## Die Frage

REST und MCP prüfen jeden Pfad **bevor** ein Weg ihn bekommt, am Rohtext des
Rumpfs (`domain.PfadeIn`, `internal/rest/pfadvertrag.go`,
`internal/mcp/pfadvertrag.go`): jeder Name, der in irgendeiner Schreibweise
ein Pfadfeld meint, und für fünf Wege zusätzlich `quelle` oder `ort`. SEC-02
fragte, ob diese fünf Zusatzwege besser am **einmal dekodierten DTO** prüfen
— eine Dekodierung statt zwei, und die Prüfung genau an dem Wert, den der Weg
benutzt. Die AK verlangten dabei **gleiche Pfade, gleiche Fehler und gleiche
Dateisystemwirkung** vor und nach der Umstellung.

## Was gemessen ist

Am 2026-10-07 mit Go 1.27.1, `encoding/json` und `DisallowUnknownFields`, ein
DTO `struct{ Quelle string \`json:"quelle"\` }` gegen die Rohwache:

| Rumpf | Wert im DTO | die Rohwache prüft | heute | mit DTO-Prüfung |
| --- | --- | --- | --- | --- |
| `{"quelle":"/etc","Quelle":"/ok"}` | `/ok` | `/etc` und `/ok` | abgewiesen | angenommen, wirkt auf `/ok` |
| `{"quelle":"/ok","quelle":null}` | `/ok` | `""` | abgewiesen | angenommen, wirkt auf `/ok` |
| `{"quelle":null}` | `""` | `""` | abgewiesen | je nach Regel für Leeres |
| `{"QUELLE":"/etc"}` | `/etc` | `/etc` | abgewiesen | abgewiesen |

**Kein Pfad entkommt der Rohwache**: sie prüft jeden Wert unter jedem Namen,
der in das Feld fallen kann, auch den, der am Ende überschrieben wird. Eine
DTO-Prüfung prüft nur den letzten — sicher auch sie, aber **nachgiebiger**:
zwei Eingabeklassen wechselten von „abgewiesen" zu „angenommen und
ausgeführt".

## Die Entscheidung

**Die Rohwache bleibt der Vertrag, für alle Wege, die fünf eingeschlossen.**
Eine Umstellung auf das DTO findet nicht statt.

- **Weil die AK sich am Verhalten des Dekodierers widersprechen.** Gleiche
  Fehler und gleiche Wirkung gibt es nur, wenn doppelte Schlüssel und `null`
  an Pfadfeldern **zusätzlich** am Rohtext abgewiesen werden — also mit genau
  der zweiten Prüfung, die die Umstellung abschaffen sollte.
- **Weil eine Lockerung an einer Schutzwache eine Entscheidung nach §16
  wäre**, an REST und MCP einzuordnen, und nichts sie verlangt: kein Weg
  braucht doppelte Schlüssel oder `null` als Pfad.
- **Weil der Gewinn nur Pflege ist.** Die Rohwache kostet eine zweite
  Dekodierung je Anfrage unter dem Rumpfdeckel (`rumpfgrenze`) — **nicht
  gemessen**, und kein Lauf zeigt sie als Last. Die Zuordnung „welches Feld
  ist ein Pfad" hält `internal/werkzeuge/pfadstrukturen_test.go` schon heute
  an jedem Feld fest.

## Verworfen

- **Das DTO mit eigener Ablehnung doppelter und leerer Pfadfelder.** Gleiche
  Fehler, aber zwei Prüfungen statt einer — der Zustand, den SEC-02 beheben
  wollte, nur an anderer Stelle.
- **Das DTO allein, als bewusste Lockerung.** Sicher, aber eine Änderung des
  Vertrags ohne Anlass, und zwei Klassen von Anfragen, die heute an der Grenze
  scheitern, gingen dann bis auf die Platte.

## Folgen

- `wegpfadfelder` und `Werkzeug.Pfadfelder` bleiben die Aufzählung der fünf
  Zusatzwege; ihr Wächter bleibt.
- Wer eine neue Fläche baut, die Pfade nimmt, prüft am Rohtext mit
  `domain.PfadeIn` — derselbe Vertrag, dieselben Fehler.

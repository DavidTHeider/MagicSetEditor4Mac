# 0015 — Bezeichner folgen dem Vokabular, nicht der Sichtbarkeit

**Stand:** angenommen
**Datum:** 2026-08-15

## Lage

`AGENTS.md` §9 sagte: **exportierte** Typen, Felder, Methoden und Funktionen
sind englisch. Begründet mit zwei Sätzen — die exportierte Schicht grenze an
Go-Standardbibliothek und Wails-Bindings, und MSE-Bezeichner seien englisch.

#115 hat die
Regel an `internal/storage` angestrichen: `SetSpeicher`, `Setkoepfe`,
`Rohbloecke`, `Oeffnen`, `Migriere`.

## Was die Zählung ergibt

Die Regel ist **nirgends** befolgt worden, nicht nur in `storage`. Gezählt über
alle exportierten Bezeichner in `internal/` (Produktivcode, ohne Tests):

| Paket | deutsch | englisch |
| --- | ---: | ---: |
| `app` | 72 | 57 |
| `service` | 34 | 20 |
| `storage` | 28 | 11 |
| `msepaket` | 26 | 36 |
| `render` | 22 | 26 |
| `msescript` | 19 | 25 |
| `domain` | 14 | 31 |
| `mtgjson` | 11 | 16 |

**Auch `app`**, und das ist der Punkt: das sind die **Wails-Bindungen** — die
Namen, die das Frontend aufruft. `BildSchreiben`, `Bildnamen`, `Bestandsleser`.
Genau die Schicht, für die der erste Begründungssatz geschrieben wurde, ist die
mit den meisten deutschen Namen.

Ein Umbau wäre damit kein Aufräumen in einer Schicht, sondern ein Umbenennen
quer durch alle acht — samt neu erzeugten Bindungen und jedem Aufruf im
Frontend. Für null sichtbares Verhalten.

## Entscheidung

**Die Sprache eines Bezeichners folgt seinem Vokabular, nicht seiner
Sichtbarkeit.**

| Der Bezeichner benennt | Sprache | Beispiele |
| --- | --- | --- |
| einen **fremden** Begriff — MSE, MTG JSON, Go, Wails | englisch | `Card`, `StyleField`, `FindByID`, `SearchQuery` |
| einen **eigenen** Begriff dieses Projekts | deutsch | `Planbauer`, `Bogenleger`, `Zerteile`, `Bildquelle` |
| Paketinternes und Lokales | deutsch | `zeileLesen`, `umschliessen` |
| Tests | deutsch | `TestCardValidateBefunde` |

Der **zweite** Begründungssatz aus §9 bleibt damit vollständig in Kraft, und er
war immer der tragende: ein Feld heißt in jeder Vorlage `casting cost`, und ein
deutscher Typname daneben verschleiert, dass es dasselbe ist. Deshalb heißt die
Entität `Card` und nicht `Karte`, und `Value.Field` nicht `Wert.Feld`.

Der **erste** Satz — „grenzt an Go und Wails" — trug nie: Wails bindet, was
dasteht, und die Go-Standardbibliothek stellt keine Namensforderung. Er wird
gestrichen.

## Was das kostet

**Ein Bezeichner ist damit nicht mehr an einer Stelle nachschlagbar, sondern
verlangt eine Frage:** ist das ein fremder Begriff oder unserer? An den Rändern
gibt es Fälle, in denen beide Antworten vertretbar sind — `Quelle` in
`msepaket` etwa: die Datei ist unsere Sicht auf ein fremdes Paket.

Die alte Regel war in diesem einen Sinn besser: sie war mechanisch prüfbar.
Sie war nur nie geprüft worden und nie befolgt, und eine mechanische Regel, an
die sich niemand hält, ist keine Regel, sondern ein Vorwurf, den man später
gegen jeden erheben kann.

**Ein Wächter darüber gibt es nicht**, und das ist beabsichtigt: „fremder
Begriff" ist keine Eigenschaft, die ein Programm ansieht. Das gehört ins
Review.

## Verworfen: umbenennen

Acht Pakete, rund 230 exportierte Namen, die erzeugten Bindungen und jeder
Aufruf im Frontend. Ein Umbau ohne Verhalten, der jeden offenen Zweig
kollidieren lässt und dessen einziger Gewinn eine Regel ist, deren Begründung
nicht trägt.

## Verworfen: die Regel auf die äußeren Schichten einschränken

Das war der zweite Vorschlag in #115 — englisch nur für `domain` und `app`.
Die Zählung widerlegt ihn: `app` ist mit 72 deutschen Namen die deutscheste
Schicht überhaupt, und es sind die Bindungen selbst. Die Einschränkung hätte
genau dort den größten Umbau verlangt.

## Wer das ändern will

ändert zuerst diesen Vermerk. Der wahrscheinlichste Anlass wäre eine zweite
Person im Projekt, für die die Frage „fremd oder eigen?" nicht dieselbe Antwort
hat — dann ist nicht die Regel falsch, sondern sie braucht Beispiele.

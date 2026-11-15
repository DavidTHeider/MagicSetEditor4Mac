# 0041 — Die Setgrenze liegt im Dienst, und RLS gibt es nicht

**Stand:** entschieden, nachgeschärft am 2026-08-22 · **Datum:** 2026-08-22 ·
**Betrifft:** #592,
#801 ·
**Ersetzt nichts** · **Ergänzt** [0010](0010-mcp-als-eigener-prozess.md)

## Die Frage

Phase N verlangt eine Aussage über die Zugriffsgrenze: **wo liegt sie, und wer
hält sie?** Der Anlass ist, dass SQLite kein Row Level Security kennt — eine
Serverdatenbank brächte eine Regel mit, die sagt, welche Zeilen ein Aufrufer
überhaupt sieht. Hier gibt es das nicht.

## Was hier überhaupt zu schützen ist

**Keine Nutzenden, keine Rollen, keine Rechte.** Dies ist eine Anwendung auf
einem Rechner, und wer sie startet, darf alles. Eine Rechteverwaltung wäre
nicht nur überflüssig, sie wäre irreführend: sie behauptete eine Trennung, die
es nicht gibt.

**Was es gibt, ist eine Grenze, und sie ist fachlich:** eine Karte gehört zu
genau einem Set. Ein Vorgang, der eine Karte meint, meint sie **in einem Set** —
und ob beide Angaben zusammenpassen, ist eine Frage, die jemand beantworten
muss.

## Was nachgemessen war

Die Schreibwege des Lebenszyklus nahmen bis zum 2026-08-22 eine blanke
Kartenkennung:

```go
KarteKopieren(ctx context.Context, karteID int64) (int64, error)
KarteLoeschen(ctx context.Context, karteID int64) error
KartenTauschen(ctx context.Context, eineID, andereID int64) error
```

Kein Set. Und `internal/app` reichte sie über `mcpnaht` **unmittelbar aus
`storage`** durch — der Lebenszyklus lief an `internal/service` vorbei. Eine
Grenze, die dort geprüft werden soll, konnte es nicht, weil der Weg nicht dort
entlangführte.

`KartenTauschen` hatte als einziges eine Prüfung, und sie ist **relativ**:
beide Karten müssen im *selben* Set liegen — nicht in dem, das der Aufrufer
nennt. Sie steht außerdem in `storage`.

## Warum das im Fenster nicht auffiel

Dort kommt die Kartenkennung aus der Liste des offenen Sets. Ein Mensch, ein
Prozess, eine Sicht — eine falsche Kennung entsteht nicht.

**Über MCP schon.** `schreiben.karte.loeschen` nimmt sie von einem
Sprachmodell entgegen. Eine verwechselte, erfundene oder aus einem früheren
Gespräch übernommene Kennung ist dort kein Angriff, sondern ein **gewöhnlicher
Fehlerfall** — und sie löschte eine Karte in einem Set, das niemand geöffnet
hatte. Zurückholen ließ sie sich nur über den Verlauf des Fensters, das von der
Änderung nichts weiß.

Das ist die Naht, die [0010](0010-mcp-als-eigener-prozess.md) beschreibt:
**zwei Schreiber auf einem Bestand.** Sie ist dort für den Verlauf bedacht
worden und nicht für die Zugehörigkeit.

## Die Entscheidung

1. **Die Grenze ist das Set**, und sie wird in `internal/service` geprüft —
   `service.Kartenaufsicht`. Jeder Vorgang des Lebenszyklus nennt das Set, und
   die Aufsicht hält die Karte dagegen.
2. **Der Befund ist `ErrNotFound`, nicht `ErrValidation`.** Aus der Sicht
   dessen, der ein Set genannt hat, gibt es diese Karte dort nicht. Dieselbe
   Antwort bekommt, wer eine Kennung nennt, die es überhaupt nicht gibt: wer
   nach einer fremden Karte fragt, erfährt nicht, dass es sie gibt.
3. **Kein RLS, und kein Ersatz dafür.** Nicht, weil SQLite es nicht kann,
   sondern weil es die falsche Ebene wäre: die Zuordnung Karte→Set ist eine
   fachliche Regel, und §2 sagt, wo fachliche Regeln leben. Eine Regel in der
   Datenbank wäre zusätzlich unsichtbar für jeden, der den Dienst liest.

## Was das nicht leistet

**Es ist keine Sicherheitsgrenze.** Wer den Bestand mit `sqlite3` aufmacht,
schreibt daran vorbei, und das soll so sein — es ist die Datei der Nutzenden.
Die Aufsicht schützt vor einer **Verwechslung**, nicht vor einem Willen.

**Sie prüft nicht, ob es das Set gibt.** `KarteAnlegen` hat keine Karte, deren
Zugehörigkeit zu halten wäre; ob das Set besteht, entscheidet der Fremdschlüssel
darunter. Eine zweite Abfrage davor wäre ein Zeitfenster, kein Schutz.

**Sie kostet eine Abfrage je Vorgang.** Beim Tausch zwei. Das ist bei einem
Lebenszyklusvorgang nicht messbar — bei einem Weg, der je Tastendruck läuft,
wäre es die falsche Bauform.

## Nachgeschärft am selben Tag — und was daran falsch war

Hier stand zuerst, die Regel gelte, **und** der Feldweg bleibe ausgenommen:
„das ist kein Versehen, sondern der nächste Schritt … gehört in ein eigenes
Werkstück". Ebenso war der Fenster-Lebenszyklus ausgenommen; es hieß, wer ihn
„künftig" aus dem Fenster rufe, rufe die Aufsicht.

**Der Zweitblick über Phase N hat beides angestrichen, und er hat recht:**

> Die in 0041 definierte Regel ist deshalb nur für den MCP-Lebenszyklus
> umgesetzt. […] „Eigenes Werkstück" ist eine Umfangsverschiebung, keine
> fachliche Begründung für die Ausnahme.

Eine Regel, die für jede Kartenänderung zugesagt und für vier von sieben Wegen
gehalten wird, ist keine Regel — sie ist eine Absicht mit einem Wächter daneben.

**Und der naheliegende Einwand trägt nicht.** Er lautete: im Fenster komme die
Kennung aus der Liste des offenen Sets, also könne sie nicht fremd sein. Das
widerlegt §7 Regel 1 selbst: „Ein Wechsel des Bezugsobjekts räumt den offenen
Kontext" steht in der Leitlinie, **weil** eine stehengebliebene Kennung nach
einem Setwechsel vorgekommen ist. Genau die zeigt auf eine Karte des vorigen
Sets. Ein Wächter, der die dokumentierte Fehlerklasse der eigenen Anwendung
auslässt, lässt das Wahrscheinlichste aus.

Seit #808
gilt die Regel für **jeden** kartenbezogenen Schreibweg:

| Weg | Über |
| --- | --- |
| `schreiben.feld`, `schreiben.karte.*`, `schreiben.karten.tauschen` (MCP) | `app.mcpnaht` → Aufsicht |
| Feld setzen, Massenänderung (Fenster) | `App.FeldSetzen`, `App.FelderSetzen` → Aufsicht |
| kopieren, löschen, tauschen (Fenster) | Aufsicht statt Bestand |
| Bild einsetzen | dieselbe Naht wie Feld setzen |

**Die Prüfung sitzt im Weg, nicht daneben.** Ein Wächter, den ein Aufrufer erst
rufen muss, ist einer, den der nächste Aufrufer vergisst — deshalb steht die
Aufsicht **an der Stelle des Bestands** und nicht als zusätzlicher Aufruf davor.

## Was weiterhin nicht geprüft wird

**Lesende Wege.** `Vorschau` und `Symbolpalette` nehmen weiterhin eine blanke
Kartenkennung. Das ist eine Aussage über den Zuschnitt und keine über den Wert:
diese Entscheidung handelt vom **Schreiben**, und eine Vorschau hinterlässt
nichts. Wer sie ausdehnen will, tut es als eigenes Werkstück — und dann steht
hier, dass es geschah.

## Folgen

- Die MCP-Werkzeuge verlangen `setID` — die vier des Lebenszyklus und
  `schreiben.feld`. Das ist eine Änderung an einer Schnittstelle und steht in
  `CHANGELOG.md`.
- `app.Kartenschreibnaht` und sechs Wails-Bindungen tragen die Setkennung mit;
  die Fassade des Frontends ebenso.
- Eine unlesbare Setkennung meldet sich am Feld `setID` und nicht mehr am
  gemeinsamen `id` — bei zwei gleichartigen Angaben ist der gemeinsame Name
  keiner (§8).

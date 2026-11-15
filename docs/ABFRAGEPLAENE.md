# Die Abfragepläne der leseintensiven Listen

**Gemessen, nicht erklärt** — ENG-05.AK-02
(#1730).
Die Invarianten, zu denen die Indizes gehören, stehen in
[`INVARIANTEN.md`](INVARIANTEN.md).

## Wie gemessen wird

**An den echten Abfragen, mit SQLites eigenem Zähler.**
`TestDieListenLesenNichtDenGanzenBestand` (`internal/storage/abfrageplaene_intern_test.go`)
füllt einen Bestand mit zwei Profilen — das fremde **hundertmal größer**
(20 gegen 2.000 Sets, Decks, Turniere, Formate und Sammlungszeilen, 10.100
Karten mit Namen, Marken und Verlaufsschritten, 10.100 Preisstände) —, hängt
an die eine Verbindung des Bestands einen Rückruf und lässt die Listen des
kleinen Profils laufen. Nach jeder Anweisung liest der Rückruf
`SQLITE_STMTSTATUS_FULLSCAN_STEP`: wie viele Schritte sie in einem
vollständigen Durchlauf einer Tafel gegangen ist.

**Die Schwelle ist keine Schätzung, sondern die Größe des Fremden**: liest
eine Liste mehr Zeilen am Stück, als das fremde Profil hält, hat sie die
Tafel gelesen und nicht ihr Profil. Und **der Zähler beweist sich selbst**:
vor den Listen läuft eine Abfrage, die die Tafel ganz lesen muss, und der
Lauf verlangt, dass sie gezählt wird — blieben die Zähler sonst bei null,
sagten die Nullen nichts.

**Was der Lauf nicht misst**: Zeit (sie hängt an der Maschine, die Zahl der
Schritte nicht) und Schreibwege — die stehen unten, samt ihrer Messung.

## Stand am 2026-09-23 (Wanderung 0080)

Mit `go test ./internal/storage -run TestDieListenLesenNichtDenGanzenBestand -v`
steht der Plan jeder Anweisung im Protokoll des Laufs.

| Liste | Vollscan-Schritte | Plan |
| --- | ---: | --- |
| Sets | 0 | `sets_profil (profil_id=?)`, je Set die Karten über `cards_nach_set`; Ordnung über einen Zwischenbaum |
| Kartenzeilen | 0 | `cards_nach_set (set_id=?)`, Name und Zahl der Werte über den Schlüssel von `card_values` |
| Suche, Volltext | 0 | FTS5 (`card_suche MATCH`), dann Karte und Set über ihre Kennung |
| Suche, Marken | 0 | **vom Profil aus**: `sets_profil` → `cards_nach_set` → Schlüssel von `karten_marken` |
| Sammlung | 0 | die Eindeutigkeit `(profil_id, …)` für Ordnung und Fenster |
| Decks | 0 | `decks_profil`, die Plätze über `deckplaetze_deck` |
| Turniere | 0 | `turniere_profil`; Ordnung über einen Zwischenbaum |
| Formate | 0 | `formate_name_je_spiel (profil_id=?)` |
| Verlauf | 0 | `verlauf_nach_set (set_id=?)` |
| Preisverlauf | 0 | `preisstaende_neueste` über alle acht Merkmale der Reihe |
| Handel (TRADE-01, 2026-09-30) | 0 | `handelsrevisionen_abdruck (profil_id=?)`, je Vorgang Anfang, Archiv und Karten über ihre Schlüssel; die Karten über `json_each` der Positionen, **dann** `druckkennungen (system=? AND wert=?)` |

**Ein Zwischenbaum für `ORDER BY`** ist hier kein Befund: er sortiert, was
schon auf das Profil eingegrenzt ist.

## Was die Messung gefunden hat

**Die Markensuche las `karten_marken` ganz — 10.099 Schritte für 10.100
Marken**, über jedes Profil. Die Ursache war kein fehlender Index, sondern
eine fehlende Bedingung: ohne Set klemmten weder die Volltextsuche noch die
Markensuche noch die Feldnamen auf das Profil, und die Suche fand die Karten
anderer Nutzer. Behoben in Commit `29143ef7`; gehalten von
`TestUeberAlleSetsHeisstUeberDieSetsDesProfils`. Die Klasse dahinter —
Abkömmlinge, die über eine fremde Kennung erreichbar sind — ist
#1815.

**Die Handelsliste las je Vorgang alle Scryfall-Kennungen des Bestands**
(gemessen am 2026-09-30, TRADE-01): SQLite legte `druckkennungen` vor die
Positionen und suchte über `system=?` allein — ein Bereich so groß wie die
Sammlung, **kein Vollscan** im Sinn des Zählers und deshalb von ihm nicht
gesehen. Behoben mit `CROSS JOIN`, der die Reihenfolge festhält; gehalten
von derselben Messung, die für diese Liste zusätzlich `system=? AND wert=?`
im Plan verlangt. **Die Klasse dahinter**: ein Zähler für Vollscans sieht
keinen zu breiten Indexbereich — wo eine Liste über eine zweite Tafel
verbindet, gehört der erwartete Schlüssel in die Messung.

## Die Schreibwege: Indizes für die Fremdschlüssel

**Ohne Index sucht SQLite beim Löschen der Elternzeile jede Kindzeile ab**,
einmal je gelöschter Zeile (sqlite.org/foreignkeys.html, Abschnitt 3 —
„Required and Suggested Database Indexes"). Acht Fremdschlüssel hatten
keinen Index, der mit ihren Spalten beginnt; Wanderung 0080 gibt ihnen
einen, und `TestJederFremdschluesselHatSeinenIndex` hält, dass kein neuer
ohne dazukommt.

| Fremdschlüssel | Index aus 0080 |
| --- | --- |
| `deckplaetze.karte_id` → `cards` | `deckplaetze_karte` |
| `decks.format_id` → `formate` | `decks_format` |
| `ki_feldherkunft.karte_id` → `cards` | `ki_feldherkunft_karte` |
| `reservierungen.exemplar_id` → `sammlung` | `reservierungen_nach_exemplar` |
| `turnierdeckplaetze.teilnehmer_id` → `turnierteilnehmer` | `turnierdeckplaetze_teilnehmer` |
| `turnierpaarungen.a` → `turnierteilnehmer` | `turnierpaarungen_a` |
| `turnierpaarungen.b` → `turnierteilnehmer` | `turnierpaarungen_b` |
| `turnierteilnehmer.deck_id` → `decks` | `turnierteilnehmer_deck` |

**Seit Wanderung 0088** kommt `turniere.format_id` → `formate` dazu, gleich mit
seinem Index `turniere_format` (DOM-04.AK-01, #1699), **seit 0089**
`turnierprotokoll.turnier_id` → `turniere` mit `turnierprotokoll_turnier`
(DOM-04.AK-04).

**Gemessen am 2026-09-23** mit der Kommandozeile von SQLite 3.54.0 an einer
Kopie des Schemas: ein Set mit 2.000 Karten zu löschen, während 60.000
Deckplätze auf sie zeigen, dauerte **1,92 s ohne** den Index auf
`deckplaetze.karte_id` und **0,11 s mit** ihm. Die Zeit gilt für diese
Maschine; der Unterschied ist der zwischen einem Durchlauf je Karte und einer
Suche je Karte.

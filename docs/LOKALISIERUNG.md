# Was von der Sprache abhängt — und was ausdrücklich nicht

**Das Feldinventar aus LOC-03.AK-01**
(#1689):
*„Ein Feldinventar umfasst Dezimal-/Tausenderzeichen, Geld, Datum/Uhrzeit,
Prozent, Einheiten, Sortierung und CSV-Trennzeichen."*

---

## Die drei Klassen

Jede Stelle in diesem Programm, an der Sprache eine Rolle spielen **könnte**,
gehört in genau eine von drei Klassen. Die Klasse entscheidet, was richtig
ist — und ob eine Stelle, die sich an der Sprache orientiert, ein Fehler ist.

| Klasse | Was gilt | Woran man sie erkennt |
| --- | --- | --- |
| **Kanonische Form** | **Keine Sprache.** So wird gespeichert und über die vier Flächen geredet: `19.99`, `2026-09-22`, `,` als CSV-Trennzeichen. Wer hier lokalisierte, machte aus derselben Datei in zwei Ländern zwei Bedeutungen | Es geht in den Bestand, in ein CSV oder über REST/MCP |
| **Menschentext** | **Die Sprache des Fensters.** Zahlen, Beträge, Zeitpunkte und die **Ordnung** von Namen folgen ihr — nicht der des Betriebssystems und nicht einer fest eingebauten | Ein Mensch liest es |
| **Technischer Schlüssel** | **Keine Sprache, aber eine feste Ordnung.** Kartenkennungen, Zonennamen, Befehlsnamen: sie werden sortiert, damit die Reihenfolge **stabil** ist. Eine Ordnung, die sich mit der Systemsprache ändert, ist hier ein Fehler — dieselbe Liste sähe auf zwei Rechnern verschieden aus | Es steht nie in einem Satz, den jemand liest |

---

## Vier Einstellungen und nicht eine

**LOC-03 verlangt, dass Sprache, regionale Darstellung, Zeitzone und
Währung getrennt einstellbar sind** — und dass die volle Locale erhalten
bleibt (AK-03). Hier stand allein die Sprache, und das Fenster kürzte
`navigator.language` auf ihren Sprachanteil: aus `de-CH` wurde `de`.

| Einstellung | Was sie entscheidet | Ohne Wahl | Ebene |
| --- | --- | --- | --- |
| `sprache` | die Wörter — welcher Katalog | die Sprache des Systems, sonst Deutsch | Mensch |
| `darstellung` | wie Zahlen, Beträge und Zeitpunkte aussehen — **eine volle Locale** wie `de-CH` | die Locale des Systems, **ganz** | Mensch |
| `zeitzone` | in welcher Zone ein Zeitpunkt erscheint | die des Rechners | Gerät |
| `vorgabe.waehrung` | womit das Währungsfeld einer neuen Sammlungszeile anfängt — **umgerechnet wird nichts** (AK-02) | leer | Profil |

**Gemessen, was die Gegend ändert** (2026-09-22): `de-CH` schreibt
»1'234.50« und »39%«, `de-DE` »1.234,50« und »39 %«; `en-GB` setzt den Tag
vor den Monat und rechnet in 24 Stunden, `en-US` nicht.

**Die Flächen ohne Fenster** — Kommandozeile, REST, MCP — lesen und setzen
alle vier wie jede andere Einstellung. **Wirksam** werden Darstellung und
Zeitzone nur im Fenster: die anderen Flächen reden kanonisch, und das ist
die ausdrückliche Ausnahme, nach der das Issue fragt.

## Das Inventar

### Geld

| Wo | Klasse | Was gilt |
| --- | --- | --- |
| `internal/domain/geld.go` | kanonisch | `GeldAus`/`GeldText`: Punkt als Trennzeichen, **kein** Tausenderzeichen, so viele Stellen, wie die Währung hat ([0127](entscheidungen/0127-ein-betrag-steht-in-der-kleinsten-einheit.md)) |
| `frontend/src/geld.ts` | Menschentext | `Intl.NumberFormat` mit der Sprache des Fensters. **Geteilt wird durch die Stellen, die Go speichert**, nicht durch die von `Intl` — die beiden sind bei 17 Währungen verschieden ([0127](entscheidungen/0127-ein-betrag-steht-in-der-kleinsten-einheit.md)) |
| `frontend/src/waehrungsgestalt.ts` | kanonisch und Menschentext | **Der Code ist kanonisch** — drei Buchstaben wie in `internal/domain/geld.go`, und nichts wird still umgeschrieben. **Die Namen in den Vorschlägen sind Menschentext**: `Intl.DisplayNames` in der **Sprache** des Fensters und nicht in der Darstellung — ein Name ist ein Wort (UX-06, #1678) |
| `frontend/src/drucksprachen.ts` | technischer Schlüssel | Scryfalls zwanzig Sprachcodes in **ihrer** Reihenfolge; die Namen dazu kommen aus dem Katalog, denn `Intl` kennt weder Phyrexianisch noch Zwergisch |
| `frontend/src/bausteine/kifeldauftrag.ts` | Menschentext | **Der Auftrag eines Feldvorschlags** (AI-05, #2042): sein Satz kommt aus dem Katalog in der **Sprache des Fensters** und verlangt die Antwort in ihr — das Modell schlägt einen Namen in der Sprache vor, in der jemand arbeitet. **Die Feldnamen der Karte bleiben, wie das Paket sie führt** (`name: …`, §9) |
| `frontend/src/waehrungsstellen.ts` | kanonisch | Dieselben null- und dreistelligen Währungen wie `internal/domain/geld.go`; `internal/werkzeuge/waehrungsstellen_test.go` hält beide gleich |
| `internal/domain/preisstand.go` (`MarktpreisAus`) | kanonisch | Marktpreise derselbe Vertrag, mit zwei Unterschieden: ein **leerer** Text ist hier ein Fehler, und ein **negativer** Preis auch ([0127](entscheidungen/0127-ein-betrag-steht-in-der-kleinsten-einheit.md)) |

### Dezimal- und Tausenderzeichen, Prozent

| Wo | Klasse | Was gilt |
| --- | --- | --- |
| `frontend/src/lokal.ts` (`lesarten`) | Menschentext → kanonisch | Eine Eingabe in der Sprache des Fensters wird kanonisch — **nur, wenn diese Sprache sie selbst so schriebe**. Hat sie zwei Werte (»1.234« Dinar im Deutschen), wird sie **erklärt** und geht nicht hinüber; hat sie keinen, geht sie unverändert, und Go nennt, was dasteht |
| `frontend/src/seiten/Diagnosepaket.tsx` | Menschentext | Größen in KB und MB, mit `toLocaleString` in der Darstellung. **Stand bis zum 2026-09-22 nicht hier**: der Wächter kannte `toLocaleString` nicht |
| `frontend/src/bausteine/Betriebszahlen.tsx` | Menschentext | Zeitpunkte, Bytes, Millisekunden und Zählungen der Metriken (OPS-05, #1722) — mit `toLocaleString` und `alsZeitpunkt` in der Darstellung |
| `frontend/src/seiten/Betriebshinweise.tsx` | Menschentext | der freie Platz in Gigabyte, dezimal wie im Finder — mit `toLocaleString` in der Darstellung (OPS02, #2348) |
| `frontend/src/seiten/Zwischenspeicher.tsx` | Menschentext | Trefferquote über `alsProzent`, Größen über `alsGroesse`, Zählungen mit `toLocaleString` (OPS-08, #1725) |
| `frontend/src/seiten/Bogenzahlen.tsx` | Menschentext | Papiermaße in Millimetern, mit `toLocaleString` in der Darstellung: `215,9 × 279,4 mm` in einer deutschen, `215.9 × 279.4 mm` in einer englischen Oberfläche. Die Zahlen daneben (Plätze, Seiten, Auflösung) sind ganze Zahlen, und die Wörter kommen aus dem Katalog (UXR-F7, #2015) |
| `frontend/src/seiten/Hilfebetrieb.tsx` | Menschentext | Bauzeitpunkt über `alsZeitpunkt`, die Zahl der Migrationen mit `toLocaleString`; Belegpfade und Dienstnamen bleiben, wie sie heißen (OPS-07, #1724) |
| `frontend/src/bausteine/dauer.ts` | Menschentext | Millisekunden und — ab tausend — Sekunden, mit `toLocaleString` in der Darstellung (OPS-05, #1722) |
| `frontend/src/bausteine/KIModelle.tsx` | Menschentext | Das Kontextfenster je Modell über `alsZahl` in der Darstellung — »8.192« im Deutschen, »8,192« im Englischen (AI-01.AK-02, #1742) |
| `frontend/src/bausteine/Laufzeitentafel.tsx` | Menschentext | Rufe, Fehlerrate über `alsProzent` und Mittelwerte in der Darstellung |
| `frontend/src/lokal.ts` (`alsProzent`) | Menschentext | Das Prozentzeichen gehört zur Zahl: »42 %« mit geschütztem Leerzeichen im Deutschen, »42%« im Englischen |
| `frontend/src/bausteine/Sammlungsformular.tsx` | Menschentext → kanonisch | Der Erwerbspreis beim Anlegen — gelesen über `preiseingabe.ts` |
| `frontend/src/bausteine/Sammlungsinspector.tsx` | Menschentext → kanonisch | Der Erwerbspreis beim Bearbeiten — über dieselbe Lesart, und ins Feld in der Schreibweise des Fensters. **Hier stand „die einzige freie Zahleneingabe mit Nachkommastellen“**: das Bearbeiten war eine zweite, und sie las nicht (Zweitblick über Phase 6.3, S6) |

**Die Übersetzung geschieht im Fenster und nur dort** — das ist die
ausdrückliche Ausnahme zwischen den Flächen, nach der LOC-03 fragt.
Kommandozeile, REST und MCP nehmen die kanonische Form unmittelbar: wer ein
Skript schreibt, soll nicht davon abhängen, welche Sprache der Rechner hat,
auf dem es läuft.

### Datum und Uhrzeit

| Wo | Klasse | Was gilt |
| --- | --- | --- |
| `domain.Exemplar.Erwerbsdatum` | kanonisch | `JJJJ-MM-TT`, ein **Kalendertag** ohne Uhrzeit und ohne Zone |
| Zeitstempel über die Flächen | kanonisch | RFC 3339 in UTC |
| `frontend/src/sprache/darstellung.ts` | Menschentext | Welche Locale und welche Zone gelten: die Wahl, sonst das System. **Was `Intl` nicht versteht, fällt zurück** — ein `RangeError` beim Zeichnen nähme jede Seite mit einem Datum mit |
| `frontend/src/zeitpunkt.ts` | Menschentext | `Intl.DateTimeFormat` mit der Sprache des Fensters. **Die Zone kennt nur der Browser** — Go reicht jedes fremde Datum unverändert durch. Die Reihenfolge ist in jeder der drei gemessenen Sprachen eine andere: Tag zuerst, Monat zuerst, Jahr zuerst |

### Sortierung

| Wo | Klasse | Was gilt |
| --- | --- | --- |
| `internal/domain/sammlungssicht.go` | Menschentext | `collate` nach der Sprache der Sicht; ohne Angabe Deutsch — das ist die Vorgabe der Flächen ohne Fenster. **Für Deutsch und Englisch ist die Ordnung dieselbe**, und das ist gemessen: beide legen »ä« neben »a«. Der Unterschied fängt bei der dritten Sprache an — im Schwedischen steht »ä« hinter »z«, im Japanischen 稲妻 vor 光, und ein Lauf hält beides fest |
| `frontend/src/sprache/vergleich.ts` | Menschentext | **Die Naht** (F01, #2243): `vergleicher(sprache)` ordnet, `falte(text, sprache)` schreibt für eine Suche klein — beide in der Sprache des Fensters. Ein Wächter (`frontend/src/sprachnaht.test.ts`) hält, dass keine Stelle sonst `localeCompare` oder `toLocale…Case` ohne Sprache ruft |
| `internal/domain/handelsfilter.go` | Menschentext | Namen der Gegenüber in der Liste der Handel (TRADE-01, #1715): `collate` nach der Sprache des Filters, ohne Angabe Deutsch — dieselbe Vorgabe wie bei der Sammlung. Das Fenster schickt seine Sprache mit (`useHandelsliste`); ein Gegenüber ohne Namen steht unter seiner Kennung, und bei gleichen Namen entscheidet die letzte Änderung |
| `frontend/src/seiten/paketsicht.ts` | Menschentext | Namen, Versionen, Herkunft und Arten von Paketen, über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/seiten/paketbaum.ts` | Menschentext | Namen im Paketbaum, über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/seiten/markenauswahl.ts` | Menschentext | Namen von Marken, über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/bausteine/partienauswahl.ts` | Menschentext | Namen von Partien, geordnet und für die Suche gefaltet über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/bausteine/einstellungstreffer.ts` | Menschentext | Die Suche in den Einstellungen faltet über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/rahmen/Palette.tsx` | **technischer Schlüssel** | Befehle der Palette nach ihrem Namen — `lesen.sets` ist eine Kennung, geordnet mit `schluesselvergleich` nach Codepunkten (F01; hier stand Menschentext, und die Ordnung folgte der Sprache des Systems) |
| `frontend/src/bausteine/Gestaltungsbeiwerk.tsx` | Menschentext | Kontrastverhältnisse der eigenen Farben, in der Darstellung des Fensters (UX-09) |
| `frontend/src/bausteine/listenordnung.ts` | Menschentext | Namen von Sets und Decks, geordnet **in der Sprache des Fensters** (UX-11, #1683) — die erste Stelle, die ihr `localeCompare` eine Sprache mitgibt |
| `frontend/src/bausteine/Listenansicht.tsx` | Menschentext | Die Suche über Sets und Decks faltet über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/seiten/Hilfesuche.tsx` | Menschentext | Die Suche in der Hilfe faltet Namen und Sätze über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/bausteine/Deckkartenwahl.tsx` | Menschentext | Die Suche über die Karten eines Sets in der Kartenwahl des Decks faltet über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/bausteine/Druckwahl.tsx` | Menschentext | Namen von Drucken, geordnet mit `localeCompare` **in der Sprache des Fensters** (die Bindung gibt sie nach der Kennung des Drucks), und die Suche darin faltet Bezeichnung und Kennung über `sprache/vergleich.ts` in der Sprache des Fensters (F01) (UXR-F18, #2033) |
| `frontend/src/seiten/Kuerzeltafel.tsx` | Menschentext | Namen der Seiten in der Auswahl der Ziffernbelegung (UX-02.AK-03), über `sprache/vergleich.ts` in der Sprache des Fensters (F01) |
| `frontend/src/ordnung.ts` | **technischer Schlüssel** | `schluesselvergleich` — zeichenweise nach Codepunkten, ohne Sprache |
| `frontend/src/paketordnung.ts` | **technischer Schlüssel** | Pakete beim neuen Set nach Rang, dann nach ihrem **Dateinamen** — wie MSE2, das nach `name()` und nicht nach dem angezeigten Kurznamen ordnet. In Go dieselbe Regel: `domain.VergleicheNachRang` |
| `frontend/src/tisch/weltbild.ts` | **technischer Schlüssel** | Markenarten am Tisch. Sie werden **angezeigt** und trotzdem nicht übersetzt: »+1/+1« ist ein Schlüssel aus dem Spielzustand, und ihre Ordnung muss auf zwei Rechnern dieselbe sein |
| `frontend/src/bausteine/Tischkartenknopf.tsx` | **technischer Schlüssel** | dieselben Markenarten, dieselbe Zusage |

### Namen, die das Programm vergibt

| Wo | Klasse | Was gilt |
| --- | --- | --- |
| `internal/storage/grundlinie/saat.sql` (aus `0034_nutzer_und_profile.sql`) | kanonisch | „Ich" und „Mein Profil" — der erste Nutzer und sein Profil, angelegt beim **ersten Öffnen** des Bestands |
| `internal/storage/nutzer.go` (`Erstprofil`) | kanonisch | „Mein Profil" — das erste Profil jedes weiteren Nutzers, dieselbe Regel |

**Kanonisch und nicht Menschentext, und das ist eine Entscheidung**
(#1963,
gefunden an der Hüllenprobe `leisteenglisch`: im englischen Fenster stand
„Ich · Mein Profil"). Drei Gründe:

1. **Ein Name ist ein Wert und keine Darstellung.** Er wird gespeichert und
   an allen vier Flächen gleich gezeigt, ist je Nutzer eindeutig
   (`UNIQUE (nutzer_id, name)`), wird beim Löschen **verglichen** — wer ein
   Profil löscht, nennt seinen Namen, und `namePruefen` in
   `internal/storage/nutzerpflege.go` hält ihn gegen den gespeicherten — und
   geht abgeschrieben in die Herkunft einer Übernahme ein („Zweiter · Mein
   Profil", [0071](entscheidungen/0071-woher-ein-deck-kommt.md)). Im Fenster
   übersetzt hieße dasselbe Profil dort anders als an der Kommandozeile: die
   zwei Bedeutungen, vor denen die kanonische Form schützt.
2. **In der Sprache des Fensters anzulegen geht nicht.** Der erste Nutzer
   entsteht in der Wanderung, beim ersten Öffnen des Bestands — vor jedem
   Fenster, und oft ohne eines: `mse`, `mse-rest` und `mse-mcp` wandern den
   Bestand genauso. Ein späteres Umbenennen „solange er noch so heißt"
   änderte die Daten eines Menschen hinter seinem Rücken und hielte jemanden,
   der sein Profil selbst so genannt hat, für unberührt.
3. **Gleich bleibt er auch in einer anderen Sprache** — wie jeder Name, den
   ein Mensch getippt hat. Die Sprache des Fensters zu wechseln benennt
   nichts um.

**Was das kostet**: wer das Fenster auf Englisch stellt, liest „Ich" und
„Mein Profil", bis er sie umbenennt — das geht im Fenster wie an der
Kommandozeile (`mse nutzer-aendern`, `mse profil-aendern`). Ein Vorgabename,
den die Wanderung in der Sprache des Systems setzte, bliebe dieselbe
Entscheidung mit einer anderen Sprache darin; wer sie will, ändert zuerst
diesen Absatz.

### Schrift

| Wo | Klasse | Was gilt |
| --- | --- | --- |
| `frontend/src/erscheinungsbild.ts` (`schriftketten`) und `frontend/src/schriftreihen.css` (`:lang(zh)`) | Menschentext | **Welche Han-Schrift zuerst kommt, entscheidet die Sprache des Elements** — G18 aus dem Zweitblick über Phase 6.3 (#1664). Jede Anzeigeschrift trägt japanische vor chinesischen Rückfällen, weil Kana nur japanische haben (UX-09.AK-02); steht davor eine Schrift ohne Han-Zeichen, bekäme ein chinesischer Name japanische Zeichenformen. Es gibt deshalb je Kette drei Reihen, und `:lang(zh)` wählt die vereinfachte, `:lang(zh-Hant)`/`zh-TW`/`zh-HK`/`zh-MO` die traditionelle — für die Oberfläche in ihrer Sprache und für jedes Element mit eigener. Am Fenster geprüft mit der Hüllenprobe `g18`. **Heute trägt außer der Wurzel kein Element eine eigene Sprache**; eine Karte aus einem chinesischen Druck bekäme sie erst, wenn ihr Name in dieser Sprache vorliegt |

### Was ausdrücklich nicht lokalisiert wird

| Was | Warum |
| --- | --- |
| **Das CSV-Trennzeichen** | Ein Komma, immer. Ein Semikolon „für Deutsch" machte aus derselben Datei zwei Formate, und beim Zurücklesen entschiede die Einstellung des Lesenden, was in der Datei steht. Wer ein Komma **im Wert** hat, setzt ihn in Anführungszeichen — so schreibt es auch die Ausfuhr |
| **Einheiten** | `250 ms` und `26 MB` schreiben Deutsch, Englisch und Japanisch gleich — gemessen, und ein Lauf hält es fest. **Die Sekunde nicht**: `Intl` schreibt »12 Sek.«, »12 sec« und »12 秒«. Die Kataloge schreiben das SI-Zeichen »s«, und das ist in jeder Sprache richtig — eine Wahl und kein Versehen (`sprache/gleichezeichen.ts`) |
| **Bezeichner aus MSE-Paketen** | Feldnamen, Spielnamen, Stilbezeichnungen und Skriptbezeichner stehen so in der Datei; eine Übersetzung machte jedes vorhandene Paket unlesbar (AGENTS.md §9) |
| **Regelnummern** | `CR 608` heißt in jeder Sprache so |

---

## Die Rundläufe

**LOC-03.AK-01 verlangt sie für de-DE, en-US und eine CJK-Locale.** Die
CJK-Locale ist Japanisch, und warum gerade sie, ist gemessen: bei den
Trennzeichen bringt sie **nichts** Neues — `ja-JP` schreibt Zahlen wie
`en-US` —, dafür die Reihenfolge eines Datums, das breite Yenzeichen
(U+FFE5) und die Ordnung von Kanji.

| Was | Wo | Geprüft wird |
| --- | --- | --- |
| Dezimal- und Tausenderzeichen, Geld, Datum und Uhrzeit, Prozent, Einheiten | `frontend/src/lokalrundlauf.test.ts` | Was das Fenster in einer der drei Sprachen zeigt, liest sich als derselbe Wert zurück |
| Geld, kanonisch | `internal/domain/geld_test.go` | EUR, USD, JPY und BHD hin und zurück |
| CSV-Trennzeichen | `internal/service/sammlungssprachen_test.go` | Was drei Sprachen in eine Notiz schreiben — ein japanisches und ein breites Komma eingeschlossen —, kommt aus dem CSV heil zurück, samt Yen und Dinar |
| Sortierung | `internal/domain/sammlungsordnung_test.go` | Die Sprache der Sicht wirkt — beim Schwedischen und beim Japanischen |

**Das Fenster lässt sich seit LOC-01 auf Japanisch stellen**
(#1687, geschlossen): acht Sprachkataloge liegen unter
`frontend/src/sprache/` (de, en, es, fr, it, ja, ru, zh). Die Läufe oben prüfen
die Schicht, die übersetzt, für de-DE, en-US und ja-JP — nicht jede Sprache der
Oberfläche. Hier stand bis zum 2026-10-11 „kann heute nicht auf Japanisch
gestellt werden“; das galt bis LOC-01.

## Was noch offen ist

**Geschlossen am 2026-10-10** (F01,
#2243): hier
stand, die acht Stellen mit `localeCompare` und die Suchen mit
`toLocaleLowerCase()` nähmen die Sprache des **Systems** — benannt, bis eine
dritte Sprache käme. Es kamen sechs. Gemessen (Node 24.15, ICU) ordnet eine
japanische Oberfläche auf einem englischen System `あ ア か カ` statt
`ア あ か カ`, eine chinesische `中 北 安` statt nach Pinyin `安 北 中`, eine
russische Lateinisch vor Kyrillisch. Jetzt geht jede Stelle über
`sprache/vergleich.ts`; die Palette ordnet Kennungen und ist damit ein
technischer Schlüssel. Eine zweite Ordnung der Sammlung im Fenster —
`sammlungssicht.ts` mit `gesehen` — hatte keinen Aufrufer mehr und ist
entfernt; die Sammlung ordnet der Bestand nach der Sprache des Fensters.

## Was dieses Dokument hält — und was nicht

**Ein Wächter hält die Vollständigkeit**
(`internal/werkzeuge/lokalisierung_test.go`): jede Datei, die `Intl.`,
`localeCompare`, `collate.New` oder `toLocale…` benutzt, steht oben.
**`toLocale…` kam am 2026-09-22 dazu** — drei Dateien fehlten, und der Lauf
war grün. Wer eine neue Stelle
anlegt und sie nicht einträgt, macht die Pipeline rot.

**Und ein zweiter hält die Naht** (`frontend/src/sprachnaht.test.ts`, F01):
kein Aufruf von `localeCompare` mit weniger als zwei Argumenten und keiner
von `toLocaleLowerCase`/`toLocaleUpperCase` ohne Sprache — auch nicht mit
`undefined` an ihrer Stelle. **Gelesen wird der Syntaxbaum und keine Regex**
(#2269): Leerraum in der Klammer, ein Umbruch, eine tiefe Verschachtelung und
ein Kommentar mit demselben Wortlaut täuschten die Regex. Er hat seine
Gegenprobe eingebaut und sieht, **dass** eine Sprache dasteht, nicht welche;
ein Aufruf mit Spread gilt im Zweifel als gedeckt.

**Was er nicht hält**, steht hier: **ob die Klasse stimmt**. Dass
`weltbild.ts` Schlüssel sortiert und `paketsicht.ts` Namen, sieht nur ein
Mensch — und wer es verwechselt, bekommt eine Liste, die auf zwei Rechnern
verschieden aussieht, ohne dass ein Lauf rot wird.

**Und er prüft nicht, in welcher Sprache eine Stelle läuft.** LOC-01 und
LOC-02 (#1687,
#1688, beide geschlossen)
haben aus zwei Sprachen acht gemacht (de, en, es, fr, it, ja, ru, zh); was hier
steht, gilt als Klasse für alle — die Sprache der Sicht entscheidet seit F01
(2026-10-10) `sprache/vergleich.ts`. Hier stand bis zum 2026-10-11 „gilt für
Deutsch und Englisch“.

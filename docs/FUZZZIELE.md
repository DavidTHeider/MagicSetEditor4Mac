# Fuzzziele — das Register

**Welche Eingaben von außen kommen, welche davon ein Fuzzziel haben, und
was jedes Ziel zusagt** — QA-06
(#1737).
Die Aufrufe mit Budget stehen in [`RELEASE.md`](RELEASE.md), die Kampagnen und
ihre Zahlen in [`belege/fuzzkampagne.md`](belege/fuzzkampagne.md), die Ebene
im Ganzen in [`HAERTUNG.md`](HAERTUNG.md#3--fuzzen-mit-zeit--außerhalb-der-pipeline).

## Wie gezählt wurde

**Eine Fläche ist ein Einstiegspunkt mit einem Eingabekanal** — eine Funktion
oder ein Bündel, das dieselbe fremde Eingabe liest. Inventarisiert am
2026-10-01 über `internal/` und `cmd/` (Stand `2e654b9f`), lesend und ohne
Lauf, in sechs Klassen:

| Klasse | Was | Flächen |
| --- | --- | ---: |
| 1 | Paket-, Skript-, Archiv- und Sprachdateien | 30 |
| 2 | Netzrahmen, Handschlag, Einladung, Entdeckung, Partie- und Handelsnachrichten | 29 |
| 3 | Import- und Konfigurationsdateien, Austausch | 16 |
| 4 | Regel- und Handelszustand durch fremde Befehle | 5 |
| 5 | Nutzereingaben und lokal erreichbare Schnittstellen | 15 |
| 6 | Antworten fremder Dienste, je Paket | 14 |
| | **zusammen** | **109** |

**Hoch** heißt: ein fremder Peer oder eine geladene Datei erreicht die Fläche,
ohne dass ein Mensch vorher hinsieht — 58 der 109.

## Der Anteil — und was er heißt

| | vor QA-06 | nach QA-06 | nach QA-06b | nach QA-06c |
| --- | ---: | ---: | ---: | ---: |
| Ziele | 27 in 23 Paketen | 42 in 25 Paketen | 53 in 26 Paketen | **60 in 27 Paketen** |
| Flächen mit eigenem Ziel | 18 (16,5 %) | 32 (29,4 %) | 43 (39,4 %) | **52 (47,7 %)** |
| dazu teilweise gedeckt | 14 | 14 | 14 | 14 |
| hohe Flächen mit eigenem Ziel | 9 von 58 | 23 von 58 | 34 von 58 | **43 von 58** |

**Nach QA-06 stand hier „42 in 26 Paketen"** — gezählt über die Aufrufe in
`RELEASE.md` waren es 25. QA-06b
(#1967)
brachte elf Ziele und mit `internal/austausch` das sechsundzwanzigste Paket,
QA-06c (#1976)
sieben und mit `internal/app` das siebenundzwanzigste.

**25 bis 30 % ist ein Anteil der Flächen mit Ziel und keine
Testabdeckung des Codes** (QA-06.AK-02). Gezählt ist eine Fläche nur, wenn
ein Ziel genau ihren Einstieg fährt; ein Ziel, das sie nur auf einem Umweg
erreicht, zählt als „teilweise".

**Was der Anteil nicht sagt**: dass die kritischen Leser vollständig sind.
Der **Draht** ist es seit QA-06 — Rahmen, Hülle, Kennung, Einladung,
Einlösung, Adresse, Entdeckung, Kontrolle, Partienachricht, Folgen,
Wiederaufnahme, Chat und Handelsumschlag haben je ein Ziel. Auf der
**Paketseite** haben seit QA-06b Installer, Paketpfade, Einbinden, Spiel-
und Stilmodell, Bilder, Symbole, Umschlag und Manifest und das Sprachpaket
eines, seit QA-06c auch Bildquelle, Text- und Kartenskripte, die
verschlüsselte Sicherung und der Umschlaginhalt. Die Kampagnen stehen in
[`belege/fuzzkampagne-paketseite-erstlauf.md`](belege/fuzzkampagne-paketseite-erstlauf.md),
[`belege/fuzzkampagne-paketseite.md`](belege/fuzzkampagne-paketseite.md) und
[`belege/fuzzkampagne-paketseite-nachtest.md`](belege/fuzzkampagne-paketseite-nachtest.md),
die von QA-06c in [`belege/fuzzkampagne-qa06c.md`](belege/fuzzkampagne-qa06c.md)
und [`belege/fuzzkampagne-qa06c-nachtest.md`](belege/fuzzkampagne-qa06c-nachtest.md).

## Der Anteil je Ziel — QA-06.AK-04

**Seit QA-06c nennt jeder Kampagnenbericht je Ziel einen Anteil**: den
Anteil der Anweisungen des Pakets, den Saat und erarbeitetes Korpus betreten
(`scripts/fuzzkampagne.mjs`, Kopf). Gos Fuzzer zählt Kanten, gibt sie aber
nicht heraus; `new interesting` ist eine Zahl ohne Nenner. Gemessen wird
deshalb nach dem Fuzzen ein Lauf über das ganze Korpus mit
`-coverprofile -coverpkg=<paket>`.

**Was er nicht ist**: Kantenabdeckung — eine betretene Anweisung sagt nicht,
über welche Kante sie erreicht wurde. **Und der Nenner ist das ganze Paket**:
`FuzzBildangabe` betritt 0,3 % von `internal/service`, weil das Paket
Hunderte Dateien hat, die mit einer Bildangabe nichts zu tun haben;
`FuzzHandschlag` betritt 74,3 % von `internal/peer`. Verglichen wird ein
Anteil deshalb **mit demselben Ziel in einem späteren Lauf**, nicht mit
einem anderen Ziel.

**Jeder Teil einer Kampagne ist erforderlich** (Z5 des Zweitblicks über Phase
6.11, `scripts/fuzzbericht.mjs`): der Fuzzlauf, die Übernahme des erarbeiteten
Korpus und der Abdeckungslauf. Ein Ziel, bei dem einer gescheitert ist, steht
im Bericht als **unvollständig** — mit dem Grund —, sein Anteil ist ein Strich
und keine Messung, und die Kampagne endet mit Exitcode **3** (bei einem Fund
mit **1**; der Fund wiegt schwerer). Vorher stand dort „0 mit Fund", ein leerer
Anteil und Exitcode 0, wenn die Instrumentierung nach einem grünen Fuzzlauf
scheiterte. **Mit Fund darf der Abdeckungslauf rot enden**: die gespeicherte
Eingabe lässt `-run ^<Ziel>$` scheitern, das ist ihr Zweck.

## Die Ziele

Was ein Fund ist, steht im Kopf jedes Ziels — festgelegt **vor** dem ersten
Lauf. Hier steht die Kurzform.

| Ziel | Paket | Fläche | Invariante in einem Satz |
| --- | --- | --- | --- |
| `FuzzZerteile` | `internal/msepaket` | 1 Paketdatei | abgelehnt heißt ohne Baum und mit Stelle; Zerteilen∘Schreiben∘Zerteilen ist die Identität |
| `FuzzBeschaedigtesPaket` | `internal/msepaket` | 2 Archiv und Ordner | kein Panic; keine Bytes gegen die CRC; ein genannter Name wird nie still abgewiesen |
| `FuzzInstaller` | `internal/msepaket` | 3 Installer | nichts außerhalb des Ziels, kein Verweis; jede Datei trägt die Bytes eines Eintrags ihres Pakets; zwei Paketnamen sind nie gleich im Sinne der Zuordnung |
| `FuzzPaketpfad` | `internal/msepaket` | 4 Paketpfade | angenommen heißt: das genannte Paket und ein Weg, den ein `fs.FS` öffnet; Zerlegen∘Zusammensetzen ist die Identität, eigener und voller Verweis ergeben dasselbe |
| `FuzzEinbinden` | `internal/msepaket` | 5 Einbinden | kein Verweis überlebt — auch nicht in kanonischer Schreibweise; die Eingabe bleibt unberührt; zweimal dasselbe Ergebnis |
| `FuzzSpielmodell` | `internal/msepaket` | 6 Modelle aus dem Baum | jeder Feldblock wird Feld oder Befund; `boolean` beginnt mit `yes`, `no`; Farben nur am Farbfeld |
| `FuzzStilmodell` | `internal/msepaket` | 6 Modelle aus dem Baum | jede Option wird Option oder Befund; Feldstile nach `z index` |
| `FuzzDekodiereBild` | `internal/msepaket` | 7 Bilder | ausgepackt wird nie mehr, als vermessen wurde, und nie über 4 Byte je Bildpunkt; Formate wie im Programm |
| `FuzzSymbolzeichnung` | `internal/msepaket` | 8 Symbole | angenommen heißt in den Grenzen; Schreiben ist ein Fixpunkt — was gelesen wurde, liest sich geschrieben wieder |
| `FuzzZerteileSprache` | `internal/msescript` | 12 Skript | abgelehnt heißt ohne Baum und zweimal dieselbe Meldung |
| `FuzzAuswerten` | `internal/msescript` | 13 Skript auswerten | kein Hänger (Frist); Fehler heißt ohne Wert; derselbe Baum, dasselbe Ergebnis |
| `FuzzEntpacken` | `internal/sicherung` | 22 Sicherung | keine Datei außerhalb des Ziels; Beschreibung oder Fehler |
| `FuzzUmschlag` | `internal/austausch` | 27/28 Umschlag und Manifest | unterschrieben und trotzdem in Quarantäne: angenommen heißt, jedes Objekt liegt im Verzeichnis und ist so groß und so gebaut, wie das Manifest sagt — gemessen an den ausgepackten Bytes |
| `FuzzSprachpaketEinlesen` | `internal/service` | 30 Sprachpaket | abgewiesen heißt mit Feld und Schlüssel und ohne Schreiben; angenommen besteht seine Prüfung; Einlesen∘Ausgeben ist die Identität |
| `FuzzAuspacken` | `internal/bezug` | 25 tar aus dem Netz | nur gewöhnliche Dateien, nichts außerhalb — geprüft an der Platte |
| `FuzzRahmen` | `internal/leitung` | 31 Längenpräfix | Länge 0 oder über der Grenze wird abgewiesen, bevor zugeteilt wird; jede Nachricht ist genau so lang, wie ihr Rahmen sagt |
| `FuzzAuspacke` | `internal/peer` | 34 Hülle | über der Höchstlänge nie gelesen; angenommen heißt unterschrieben vom Absender; ein Byte daneben, und sie geht nicht mit anderem Inhalt durch |
| `FuzzFingerabdruck` | `internal/peer` | 38 Kennung | kein Panic; einen Fingerabdruck gibt es nur für eine lesbare Kennung, und eine fremde Unterschrift gilt nicht für sie |
| `FuzzZerlegeKontakteinladung` | `internal/domain` | 42 Einladungscode | gelesen oder mit Klasse abgewiesen, nie halb; gelesen heißt mit Daten und Unterschrift |
| `FuzzZerlegeEinloesung` | `internal/domain` | 43 Einlösung | über 4 KiB nie gelesen; Fassung in 1–1; Zerlegen∘Kodieren ist die Identität |
| `FuzzZerlegeEinloeseantwort` | `internal/domain` | 43 Antwort | ein Ergebnis aus der Menge; Zerlegen∘Kodieren ist die Identität |
| `FuzzZerlegeEinladung` | `internal/domain` | 45 tailcat-Einladung | jede Adresse kommt unverstümmelt zurück; angenommen heißt mit Adresse |
| `FuzzZerlegeAdresse` | `internal/domain` | 47 Adresse | was zum Wählen gilt, hat Wirt und Port ab eins; Zerlegen∘Zusammensetzen ist die Identität |
| `FuzzDeute` | `internal/entdeckung` | 48 mDNS-Antwort | angenommen heißt mit Kennung, Port und Wirt; ein Name ohne Steuer- und Formatzeichen, höchstens 64 Zeichen |
| `FuzzLiesKontrolle` | `internal/domain` | 49 Kontrollumschlag | nur erwartete Arten; Lesen∘Verpacken ist die Identität |
| `FuzzPartienachricht` | `internal/domain` | 50 Partienachricht | zweimal dasselbe Urteil; angenommen heißt mit Epoche |
| `FuzzEmpfaengerstand` | `internal/domain` | 51 Folgen | der Stand springt nie zurück und lässt nichts aus, auch an den Rändern von `uint64` |
| `FuzzChat` | `internal/domain` | 55 Satz am Tisch | angenommen heißt ohne Steuerzeichen und ohne Richtungsüberschreibung |
| `FuzzWiederaufnahme` | `internal/domain` | 56 Rückkehr | wer weiter ist als der Wirt, spielt nicht weiter; nachgereicht wird nur, was der Wirt hat |
| `FuzzHandelsumschlag` | `internal/domain` | 58 Handelsumschlag | kein gültiges Feld trägt den Trenner der Unterschriftsform; zwei Umschläge, zwei Formen |
| `FuzzZerlegeAustausch` | `internal/domain` | 60 Einstellungsdatei | unbekanntes Feld ist ein Befund, fremde Fassung eine Ablehnung |
| `FuzzManifestAus` | `internal/aktualisierung` | 66 Freigabemanifest | „bereit" heißt verträglich, ohne Hindernis, gleiche Plattform |
| `FuzzLiesListe` | `internal/sammelliste` | 68 Deckliste | Befunde mit Zeile; zweimal dasselbe Ergebnis |
| `FuzzLiesCSV` | `internal/sammelliste` | 69 Sammlungs-CSV | Lesen∘Schreiben ist die Identität |
| `FuzzLiesInhalt` | `internal/mtgjson` | 71 MTG JSON | abgelehnt heißt ohne Quelle; kein leerer Befund |
| `FuzzLoeseAuf` | `internal/domain` | 73 Druckverweis | eindeutig nur bei genau einem Druck |
| `FuzzZugfolge` | `internal/service` | 77 Priorität und Stapel | zwei und drei Sitze; nach jedem Schritt einer Befehlsfolge: `Pruefe()`, Priorität und „gepasst" nur bei Sitzen im Spiel, keine zwei Objekte an derselben Stapelstelle; die Eingabe des Auslegers bleibt unberührt |
| `FuzzHandelsfolge` | `internal/domain` | 78 Handelsangebot | Schritt, Vorschlag und Stand der Gegenseite wechseln nur entlang der Matrix; kein Abschluss ohne Nachweis; abgelehnt lässt den Stand stehen |
| `FuzzBildangabe` | `internal/service` | 9 Bildquelle | kein Verweis ins Set-Paket ohne die eigene Marke — eine Vorlage nennt keinen Ort auf der Platte; zweimal dasselbe Ergebnis |
| `FuzzTextskript` | `internal/msescript` | 15 Textskripte | kein Fehler mit Wert; zweimal dasselbe; gültiges UTF-8 ohne `{` und `\` kommt unverändert zurück |
| `FuzzKartenskript` | `internal/service` | 19 Kartenskripte | abgelehnt heißt mit Klasse und ohne Ruf an die Ablage; angelegt nur, was `new_card` setzen darf, und höchstens `Kartenskriptgrenze` Karten; kein Lauf über die Frist |
| `FuzzEntpackenVerschluesselt` | `internal/sicherung` | 23 verschlüsselte Sicherung | kein Arbeitsfaktor über 2^18 wird gerechnet; mit falschem Kennwort kommt nichts heraus; keine Datei außerhalb des Ziels |
| `FuzzUmschlagdeck` | `internal/app` | 29 Umschlaginhalt | keine eigene Karte des Absenders und kein Platz unter eins im Deck; **jeder Platz liegt im Deck oder ist genannt** |
| `FuzzHandschlag` | `internal/peer` | 35/36/37 Angebot, Annahme, Abschluss | an einer Stelle eine fremde Nachricht: angenommen nur derselbe Inhalt; kein Angebot zweimal; nach dem Abschluss eine Sitzung auf beiden Seiten |
| `FuzzSpielbefehl` | `internal/service` | 76 Spielbefehl | alle fünfzig Arten mit freier Nutzlast, wie vom Draht: `Pruefe()` danach, Eingabe unberührt, zweimal dasselbe, Ablehnung mit Klasse |
| `FuzzDokument` | `internal/pdf` | 75 Maße aus Stilen | das PDF bleibt wohlgeformt, kein NaN |
| `FuzzDienstruf` | `internal/rest` | 87 REST (teilweise) | fremde Marke nie unter 400; kein 5xx auf einen kaputten Rumpf |
| `FuzzTokenPruefen` | `internal/service` | 90 Token | jede Abweisung außer der am Scope sieht gleich aus |
| `FuzzPruefeRueckruf` | `internal/anmeldung` | 91 OAuth-Rückruf | nie angenommen ohne eigenes `state` |
| `FuzzLiesKarte` | `internal/scryfall` | 96 (teilweise) | Befunde statt Abbruch; zweimal dasselbe Ergebnis |
| `FuzzLiesKarten` | `internal/mtgapi` | 97 | dito |
| `FuzzLies` | `internal/mtggraphql` | 98 | dito |
| `FuzzLies` | `internal/mtgjsonpreise` | 99 | dito; der Deckel liegt auf dem ausgepackten Strom |
| `FuzzLiesGepackt` | `internal/mtgjsonpreise` | 99 gzip | Packbombe und abgeschnittener Strom enden am Budget |
| `FuzzAngebote` | `internal/cardtrader` | 100 (teilweise) | kein Angebot ohne Blueprint |
| `FuzzLiesKarten` | `internal/scrydex` | 101 (teilweise) | Befunde statt Abbruch |
| `FuzzModelle` | `internal/anthropic` | 102 (teilweise) | jeder Befund nennt seine Stelle |
| `FuzzModelle` | `internal/openai` | 103 (teilweise) | dito |
| `FuzzModelle` | `internal/ollama` | 104 (teilweise) | dito |
| `FuzzModelle` | `internal/xai` | 105 (teilweise) | dito |

**Teilweise gedeckt ohne eigenes Ziel**: 14 (Muster), 18 (Skriptbibliothek),
61 (Einstellungsvorschau), 64 (Zurückspielen), 70 (Listeneinfuhr) und 72
(Setimport) erreicht je ein Ziel auf einem Umweg.

## Ohne Ziel, hoch — die nächsten

| Klasse | Flächen |
| --- | --- |
| Handschlag | 32 Zeugnis |
| Partie und Handel | 52 Partienetz · 53/54 Wirt und Gast · 57 Handelspost · 59 Belegurkunde |
| Dateien | 62 Partiedatei · 65 fremde SQLite-Datei · 67 Prüfsummenliste und Signatur |
| Schnittstellen | 85 Dienstpfad · 88 MCP-Zeilen |

**Folgen mit Zustand** (35–37, 76–78) sind Ziele, die eine Befehlsfolge
fahren und nach jedem Schritt `Pruefe()` verlangen. **77 und 78 haben seit
QA-06b eines** (`FuzzZugfolge`, `FuzzHandelsfolge`); `FuzzZugfolge` fährt
seit #1971
auch drei Sitze — dort behielt ein Gehender die Priorität, bis CR 800.4a
und 800.4j gebaut waren. **Seit QA-06c haben auch der Handschlag (35–37,
`FuzzHandschlag`) und der Spielbefehl als Ganzes (76, `FuzzSpielbefehl`)
eines.**

## Was das Register nicht hält

- **Ob ein Ziel etwas taugt.** Der Wächter
  (`internal/werkzeuge/fuzzziele_test.go`) hält nur, dass jedes Ziel hier und
  in `RELEASE.md` steht. Die Invariante liest ein Mensch.
- **Die Vollständigkeit der Flächen.** Das Inventar ist gelesen und nicht
  gezählt; eine neue Fläche kommt mit ihrem Werkstück hierher, so wie eine
  neue Bindung nach `FAEHIGKEITEN.md`.

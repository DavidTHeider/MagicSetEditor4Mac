# Fassungen — was über einen Wechsel hinweg gilt

**Stand: 2026-09-23** — ENG-06
(#1731).

Diese Datei führt **jede Fassung**, die dieses Programm dauerhaft speichert oder
nach außen trägt, an **einer** Stelle: wo sie steht, was eine **neuere**
Anwendung mit einem **älteren** Stand tut, was eine **ältere** mit einem
**neueren** tut — und welcher Lauf es belegt. Die Regeln dahinter stehen in
[0033](entscheidungen/0033-was-ueber-fassungen-hinweg-zugesagt-wird.md); hier
steht, **wo** sie gelten.

**Ein Wächter hält die Tabellen gegen den Quelltext**
(`internal/werkzeuge/fassungen_test.go`): jede Konstante, die eine Fassung
trägt — dem Namen nach oder als Kennung der Gestalt `mse4mac/<wort>/<zahl>` —,
steht hier mit ihrem Wert, und jeder genannte Lauf existiert. **Was er nicht
hält**, ist, ob eine Zeile das Verhalten richtig beschreibt; dafür stehen die
Läufe daneben. Und ein Format **ohne** Fassung findet er nicht — die bekannten
stehen unten unter »Ohne eigene Fassung«.

## Die Anwendung und ihr Bestand

| Gegenstand | Fassung | steht in | eine neuere liest einen älteren Stand | eine ältere liest einen neueren Stand | Beleg |
| --- | --- | --- | --- | --- | --- |
| **Die Anwendung** | `1.0.0` | `app.Fassung` | — | — | `TestDieFassungImQuelltextStimmtMitWailsUeberein` |
| **Der Bestand** (SQLite) | die **Grundlinie** bis `0097_faktorhaushalt.sql` — 97 Schritte mit Name und Prüfsumme, wie 0.7.0 sie ausgeliefert hat —, danach Wanderungen als Datei, die erste `0098_sicherungsvermerke.sql` (M01, #2345); ein **leerer** entsteht in einem Zug (J02, #2266, [0189](entscheidungen/0189-nur-die-grundlinie.md)) | `internal/storage/grundlinie/`, `internal/storage/migrationen/` | **einen aus 0.7.0 öffnet sie ohne einen Schritt**, einen jüngeren wandert sie vorwärts, nachdem er gesichert ist; **einen von vor 0.7.0 weist sie mit Rat ab**, ohne ihn anzufassen — der Weg führt über 0.7.0 | **weist ihn ab** und nennt die Wanderungen, die sie nicht kennt | `TestEineSicherungAusNullSiebenWirdZurueckgespieltUndGeoeffnet`, `TestEinBestandVorDerGrundlinieWirdMitRatAbgewiesen`, `TestEineKuenftigeWanderungLaeuftAufDerGrundlinie`, `TestEinNeuererBestandWirdMitStandAbgewiesen`, `TestEinAltbestandBekommtDieSicherungsvermerke`; am nativen Fenster `aufstieg09` (0.8.0 → heute, 25.000 Karten, Rückweg durch das Fenster von 0.8.0) |
| Eine nachträglich geänderte Wanderung | Prüfsumme je Wanderung | Buchführung im Bestand | **weist ab** — dieselbe Nummer, anderer Inhalt | dito | `TestEineNachtraeglichGeaenderteMigrationWirdAbgewiesen` |
| **Eine Sicherung** des Bestands | die der Anwendung und das Schema | `sicherung.json` im Archiv | **spielt sie zurück**, wenn sie auf der Grundlinie steht, und wandert sie; **eine von vor 0.7.0 weist sie mit Rat ab** | **weist sie ab**, bevor sie etwas ersetzt | `TestEineSicherungAusNullSiebenWirdZurueckgespieltUndGeoeffnet`, `TestEineSicherungAusNullFuenfWirdMitRatAbgewiesen`, `TestSicherungZurueckspielenWeistEineNeuereFassungAb` |

**Zwischen zwei Freigaben trägt jeder Bau dieselbe Fassung.** Ein Bau nach
0.6.0 heißt `0.6.0` und kann schon Wanderungen mehr kennen. Deshalb entscheidet
beim Bestand **das Schema** und nicht die Nummer — und deshalb nennt die
Meta-Auskunft zusätzlich den **Bau** (unten).

### Der Aufstieg

**Seit 0.8.0 gibt es nur die Grundlinie**
([0189](entscheidungen/0189-nur-die-grundlinie.md)). Die 97 Wanderungen bis
`0097_faktorhaushalt.sql` liegen nicht mehr als Datei vor; ein Bestand führt
sie in seiner Buchführung wie bisher, und `grundlinie/umfang.txt` trägt Name
und Prüfsumme — dieselben, die 0.7.0 ausgeliefert hat, verglichen am
2026-10-12 für alle 97.

- **Ein Bestand aus 0.7.0 öffnet sich ohne einen Schritt.** Belegt an einer
  Sicherung, die 0.7.0 selbst geschrieben hat (`testdaten/fassung-0.7.0/`):
  zurückgespielt behält sie Sets, Karten, Deck und Sprache, und die
  Buchführung nennt genau, was diese Fassung kennt.
- **Ein Bestand von vor 0.7.0 wird mit Rat abgewiesen** — bevor irgendetwas
  an ihm geschieht, auch an seiner Buchführung
  (`TestEinBestandOhnePruefsummenspalteBleibtUnberuehrt`). Der Weg führt
  über 0.7.0: dort einmal geöffnet, steigt er auf. Die Sicherung, die 0.5.0
  geschrieben hat (`testdaten/fassung-0.5.0/`), belegt die Abweisung.
- **Was nach der Grundlinie kommt**, ist wieder eine Wanderung als Datei und
  läuft Schritt für Schritt, genau einmal
  (`TestEineKuenftigeWanderungLaeuftAufDerGrundlinie`).

**Einstellungsdatei und Umschlag** hängen nicht am Schema: eine
Einstellungsdatei aus 0.5.0 wird ohne Befund übernommen, ein Umschlag, den ein
0.5.0-Peer unterschrieben hat, ebenso — **die Unterschrift von damals trägt
heute**.

### Der Abstieg

**Ein Abstieg heißt: die Sicherung zurückspielen** — und nicht rückwärts
gerichtetes SQL (ENG-06.AK-02). Vor jeder Wanderung legt die Anwendung eine
Sicherung mit der Fassung im Namen an (`app.SicherungVorMigration`); eine
**ältere** Fassung spielt sie zurück. Einen **neueren** Bestand öffnet sie
nicht und eine **neuere** Sicherung spielt sie nicht zurück — beides benannt
und bevor etwas ersetzt ist. Was nach der Wanderung geschrieben wurde, ist
mit dem Abstieg fort; der Rücksprung beziffert es, bevor er etwas tut
(`TestDerVergleichNenntJedenGegenstandEinzeln`).

**Vor einem Update** prüft die Anwendung, ob die neue Fassung diesen Bestand
öffnen kann — Wanderungen, Prüfsummen, seit 0.8.0 die Grundlinie
(`HindernisVorDerGrundlinie`), die Fassungen der Formate, Platz, laufende
Vorgänge und Sicherung (`domain.Vertraeglichkeitsurteil`, O6). **Eine Fassung
vor 0.8.0 kennt die Grundlinie nicht**: ihre Prüfung nennt 0.8.0 für einen
Bestand aus 0.6.0 verträglich, und erst der Start weist ihn ab.

## Dateien, die jemand weitergibt

| Gegenstand | Fassung | steht in | eine neuere liest eine ältere Datei | eine ältere liest eine neuere Datei | Beleg |
| --- | --- | --- | --- | --- | --- |
| **Einstellungen als Datei** | `1` | `domain.AustauschFassung` | **liest** jede Fassung von 1 bis zur eigenen | **weist sie ab** und nennt, welche sie kennt | `TestEineNeuereFassungWirdAbgewiesen`, `TestEineEinstellungsdateiAusNullFuenfWirdUebernommen` |
| **Eine Partie als Datei** | `1` | `domain.PartieaustauschFassung` | **liest** jede Fassung von 1 bis zur eigenen — die erste gibt es seit 0.6.0 | **weist sie ab** | `TestZerlegePartieWeistAb` |
| **Ein Turnier als Datei** | `1` | `domain.TurnieraustauschFassung` | **liest** jede Fassung von 1 bis zur eigenen — die erste gibt es seit 0.6.0 (DOM-04.AK-03, Vermerk [0140](entscheidungen/0140-ein-turnier-als-datei.md)) | **weist sie ab** | `TestEineFalscheTurnierdateiIstEinNein` |
| **Der Umschlag** — Set, Karte, Vorlage, Deck, Leihe | `1` | `austausch.Fassung` | **verlangt Gleichheit** und weist eine andere ab, statt sie herunterzuhandeln | dito | `TestEineFremdeFassungWirdAbgewiesen`, `TestEinUmschlagAusNullFuenfWirdUebernommen` |
| **Der Druckverweis** in einer Sammlungsliste | `1` | `domain.Verweisfassung` | **liest** auch eine Liste ohne Verweis | die Spalte nennt ihre Fassung je Zeile | `TestEinDruckWirdZurPortablenReferenz`, `TestEinAltesCSVOhneReferenzLiestSich` |
| **Ein Installer** (`.mse-installer`, das Format von MSE2) | `0.3.7` | `msepaket.installerfassung` | **liest** jeden — die Fassung steht in `mse_version:` der Hauptdatei und wird nicht geprüft; gelesen wird, was MSE2 in `package:` schreibt (FMT-02.AK-04, #1691) | dito — geschrieben wird immer `0.3.7`, `file_version_installer` von MSE2 (`src/util/version.cpp`, Zeile 94, Fassung f9d9356 vom 2021-07-06) | `TestEinInstallerTraegtDiePaketeWieMSE2`, `TestEinInstallerAusMSE2WirdGelesen` |
| **Eine eigene Übersetzung als Datei** | `mse4mac/sprachpaket/1` | `domain.Sprachpaketformat` | **verlangt Gleichheit**: eine Datei mit `mse4mac/sprachpaket/2` oder höher nennt ihr Format und wird abgewiesen, eine mit fremdem Format ist keine Übersetzung (LOC-02, Vermerk [0147](entscheidungen/0147-eigene-uebersetzungen.md)) | dito | `TestEineFremdeUebersetzungsdateiWirdAbgewiesen` |

**Beim Umschlag steht die Fassung innen, und das mit Absicht**: das
unterschriebene Manifest (`austausch.json`) trägt sie. Der Behälter darum
(`umschlag.json`) ist eine Liste von Prüfsummen und nennt keine — eine
Fassung außerhalb der Unterschrift ließe sich ändern, ohne dass es auffiele.

## Der Weg zu anderen Rechnern

**Hier gibt es keine Aushandlung, sondern Gleichheit** — und eine Abweisung,
die sagt, woran es liegt (ENG-06.AK-03). Wer zwei Fassungen kann, bietet sie
als zwei Protokolle an; eine Einigung im Ablauf wäre die Stelle, an der ein
Dritter die schwächere wählt (`peer.Handschlagfassung`).

| Gegenstand | Fassung | steht in | Verhalten | Beleg |
| --- | --- | --- | --- | --- |
| **Die Einladung** | `1` | `domain.Einladungsfassung` | nimmt jede Fassung bis zur eigenen; eine **neuere** wird benannt abgewiesen — und damit auch ein tailcat-Drahtformat, das diese Fassung nicht kennt ([0114](entscheidungen/0114-tailcat-gepinnt.md)) | `TestEineNeuereEinladungWirdBenanntAbgewiesen` |
| **Die Kontakteinladung** | `1` | `domain.Kontakteinladungsfassung` | nimmt jede Fassung bis zur eigenen; eine **neuere** wird als `einladungZuNeu` benannt abgewiesen. **Die Fassung geht in die Unterschrift ein** (`mse4mac-kontakt/1` als Vorsatz), eine andere unterschreibt etwas anderes (NET-01, #1705) | `TestKontakteinladungZerlegenLehntAb` |
| **Die Einlösung einer Kontakteinladung** | `1` | `domain.Einloesefassung` | nimmt jede Fassung bis zur eigenen; eine **neuere** wird als `einloesefassung` benannt — der Einladende antwortet „fassung", und wer annimmt, liest den Rückfall auf zwei Einladungen. **Eine eigene Zahl neben der des Handschlags**: die Einlösung kann sich ändern, ohne dass sich der Handschlag ändert. Eine ältere Gegenstelle erkennt man **vor** dem Wählen: sie stellt keine einlösbare Einladung aus (NET-01b, #1954, Vermerk [0156](entscheidungen/0156-eine-einladung-genuegt.md)) | `TestEineNeuereEinloesefassungWirdBenannt`, `TestDieEinladungSagtObSieEinloesbarIst` |
| Die Ableitung des Vergleichscodes | `mse4mac-vergleich/1` | `domain.vergleichsfassung` | **muss auf beiden Seiten gleich sein** — zwei Fassungen zeigen für dasselbe Paar verschiedene Ziffern, und jeder Vergleich scheiterte. Deshalb ist der Wert gegen eine unabhängige Rechnung eingefroren | `TestVergleichscode` |
| **Der Handschlag** zweier Peers | `1` | `peer.Handschlagfassung` | **Gleichheit** — eine fremde Fassung schließt keinen Handschlag | `TestEineFremdeFassungWirdAbgewiesen` |
| Das Protokoll der Paarung | `/mse4mac/pair/1` | `peer.ProtokollPaarung` | ein Zweck gilt nur in seinem Protokoll | `TestEinZweckGiltNurInSeinemProtokoll`, `TestJederZweckNenntSeinProtokoll` |
| Das Protokoll des Teilens und Handelns | `/mse4mac/share/1` | `peer.ProtokollTeilen` | dito | `TestEinZweckGiltNurInSeinemProtokoll` |
| Das Protokoll der Partie | `/mse4mac/game/1` | `peer.ProtokollSpiel` | dito | `TestEinZweckGiltNurInSeinemProtokoll` |
| Der Bereich der Unterschrift einer Hülle | `mse4mac/huelle/1` | `peer.bereich` | geht in jede Unterschrift ein; eine andere Fassung unterschreibt etwas anderes | `TestEineVerfaelschteNachrichtEndetAnDerHuelle` |
| Die Kanalbindung | `mse4mac/kanal/1` | `leitung.Kanalbereich` | Bezeichner der Ableitung aus TLS 1.3; ein anderer ergibt einen anderen Kanal | `TestEinHandschlagGiltNurAufSeinemKanal` |
| Der Kanal des Handels | `mse4mac/trade/1` | `service.Handelskanal` | der Name der Leitung, über die ein Handel geht | `TestEineNachrichtGehtDurchDieSchleife` |
| **Der Handelsumschlag** | `1` | `domain.Handelsprotokollfassung` | **Gleichheit** — ein Umschlag anderer Fassung ist ungültig. **Seit P11 (#2379) zwölf Arten**: `sicht` beantwortet `sichtanfrage` auf derselben Sitzung — **erweiternd ohne neue Fassung**, denn eine ältere fragt nie und bekommt deshalb nie eine Antwort; fragt eine neuere eine ältere, endet die Sitzung ohne Antwort, und der Fragende sagt es | `TestWasEinenUmschlagUngueltigMacht`, `TestDerFragendeNimmtKeineFremdeAntwort` |
| **Der Handelsvertrag** | `1` | `domain.Handelsvertragsfassung` | **Gleichheit**, und die Fassung geht in die Unterschrift ein | `TestEinAngebotWirdGeprueft` |

**Seit 0.5.0 hat sich keine dieser Fassungen bewegt** — nachgemessen am
2026-10-05 mit `git diff v0.5.0`: jede Konstante oben steht unverändert, neu
hinzugekommen sind die Kontakteinladung und ihre Einlösung. **Der Code darum
hat sich bewegt**: elf Dateien außerhalb der Tests in `internal/leitung` und
`internal/fernleitung` (eigene Zeugnisse aus NET-03, Proxy und Relais aus
NET-02, der Schleifenpuffer aus dem Zweitblick über UX-12 und UX-13); in
`internal/peer` nur Kommentare und die Grenze der Nonceliste aus dem
Zweitblick über Phase 6.7. Hier stand bis dahin „keine Änderung außerhalb der
Tests" — wahr am Tag, an dem es geschrieben wurde, und danach nicht mehr.

**Was eine Datei nicht belegen kann, ist der Handschlag selbst.** Er ist an
seine Leitung und an zwei frisch gewürfelte Zahlen gebunden, und ein
Mitschnitt lässt sich absichtlich nicht wieder vorlegen. Ein Handschlag
zwischen einem 0.5.0 und einem 0.6.0 auf zwei Rechnern ist deshalb
**Handabnahme**; belegt ist an einer Datei die Unterschrift
(`TestEinUmschlagAusNullFuenfWirdUebernommen`).

**Seit 0.6.0 spricht der Wirt einer Partie zuerst — und das Protokoll
bleibt `/mse4mac/game/1`** (UX-08.AK-04,
#1680).
Gleich nach der Aufnahme schickt er das Tischbild des vergebenen Sitzes;
dessen Umschlag nennt die Partie, und ein Gast, der sie nicht kennt, liest
sie daraus (`service.Partieerfragen`). **Eine neue Fassung wäre die falsche
Antwort**: jede Nachricht war schon vorher gültig, und ein Tischbild nimmt
jeder Gast seit 0.5.0 entgegen. Was sich zwischen den Fassungen verhält:

| Gast | Wirt | Was geschieht |
| --- | --- | --- |
| 0.5.0, nennt die Partie | 0.6.0 | ein Tischbild mehr, genommen wie jedes andere — die Gastseite nimmt es seit 0.5.0 unverändert entgegen (`Partiegast.kontrolle`); über die Leitung belegt ist es mit der heutigen (`TestEinGastSetztSichUndSeinZugKommtAn`) |
| 0.6.0, nennt die Partie | 0.5.0 | wie bisher — der Gast spricht zuerst |
| 0.6.0, **ohne** Partie | 0.6.0 | die Partie kommt mit dem ersten Wort (`TestEinGastOhnePartieSitztAmOffenenTischDesWirts`) |
| 0.6.0, **ohne** Partie | 0.5.0 | **der Gast wartet**, denn ein Wirt vor 0.6.0 begrüßt nicht; das Fenster sagt es dort, wo der Tisch fehlt, und die Partie lässt sich unter „Die Partie selbst nennen" eintragen |

Die letzte Zeile ist **Handabnahme** wie der Handschlag darüber: sie braucht
ein 0.5.0 am anderen Ende.

## Die Flächen nach außen

**MCP, Kommandozeile und REST sind außen** (0033, Regel 5): dort steht ein
Gegenüber, das nicht mitgeliefert wird. **Jede Änderung wird eingeordnet** —
verträglich, erweiternd, brechend — und eine brechende braucht eine
Hauptfassung (Regeln 6 und 7).

| Gegenstand | Fassung | steht in | Verhalten | Beleg |
| --- | --- | --- | --- | --- |
| **MCP** — das Protokoll | `2025-06-18` | `mcp.fassung` | nennt sie in `initialize`; über HTTP gilt die Kopfzeile `MCP-Protocol-Version`, und eine Fassung, die er nicht spricht, wird mit den Fassungen abgewiesen, die er spricht | `TestDieGenannteFassungWirdGeprueft`, `TestDieAbsageNenntDieFassungenDiesesServers` |
| MCP — die Vorgabe ohne Kopfzeile | `2025-03-26` | `mcp.Vorgabefassung` | gilt, wenn eine Anfrage über HTTP keine nennt — so will es die Spezifikation | `TestDieVorgabefassungGiltOhneKopfzeile` |
| **MCP, Kommandozeile, REST** — die Verträge | keine eigene | `app.Fassung` | folgen der Anwendung; `docs/FAEHIGKEITEN.md` führt, was jede Fläche kann | `TestDieFassungImQuelltextStimmtMitWailsUeberein` |

## Innen — Hälften derselben Auslieferung

| Gegenstand | Fassung | steht in | Verhalten | Beleg |
| --- | --- | --- | --- | --- |
| **Die Wails-Bindungen** | Abdruck, keine Fassung | `app.Vertragsabdruck` | **innen** (0033, Regel 5): sie ändern sich mit jeder Fassung. Gehalten wird, dass das Fenster jede Änderung **quittiert** hat (unten) | `TestDasFensterHatDenBindungsvertragQuittiert`, `TestDieHuelleBindetDieFassade` |
| **Der Rasterzwischenspeicher** | `7` | `render.Fassung` | Teil des Schlüssels: eine neue Fassung verwirft die alten Bilder, statt sie zu zeigen. **`2` seit QA-03** — Text wird seitdem anders gezeichnet (Kerning, vierfacher Weg, Ligaturen, Glättung). **`3` seit dem 2026-09-24** — ausgewichene Zeichen in voller Breite, und was über die Zeile ragt (die Punkte eines Ä), bleibt stehen. **`4` seit dem 2026-09-26** — der Ebenenspeicher zählt und lässt sich leeren (OPS-08); kein Punkt fällt anders. **`5` seit dem 2026-10-03** — jede Rasterung fragt eine Pforte nach einem Renderplatz (#2020); kein Punkt fällt anders. **`6` seit dem 2026-10-11** — Textschatten und das Schachbrett eines leeren Bildfelds (F02, #2244); ein Plan ohne beides zeichnet wie vorher. Im selben Zug, noch vor jeder Auslieferung: die einzige leere Zeile eines Satzes ist eine Zeile hoch statt null (#2298) — sie zeichnet nichts, ändert aber `content_height`, und das geht über den Plan, nicht über den Rasterer; und ein kursiver Abschnitt ohne `italic name:` nimmt den Schnitt „Italic" seiner Familie (Befund 3 des Zweitblicks über E/F) — das ändert Bildpunkte; und die Maske eines Textfelds bestimmt die Breite jeder Zeile (#2299) — ein Feld ohne Maske bricht wie vorher um, nur eine Zeile, neben der nichts Platz hat, rückt hinunter; und ein Symbol im Text misst in ganzen Punkten und schrumpft mit einem verkleinerten Text (#2300); und ein zu langer Text wird gesucht wie bei MSE2, auf einen Schritt unter dem größten passenden Maßstab (#2244); und der rechte Absatzrand verkürzt den Umbruch nicht mehr, und ein Wort, das allein nicht in seine Zeile passt, wird zwischen zwei Zeichen gebrochen (Befunde 2 und 3 des Nachlaufs über E/F); und eine Wahl meldet das Maß ihres vorbereiteten Bilds (#2303); und der obere Rand des ersten Absatzes gilt nicht (#2302) — das geht über den Plan, der Rasterer bekommt dafür nur ein Maß aus demselben Bildweg; und ein einzeiliges Feld (`multi line: false` des Spiels) bricht nur an einem harten Umbruch und läuft sonst über (J05, #2269) — ein Feld, dessen Text passt, zeichnet wie vorher; und der Ebenenspeicher merkt sich nur noch Bilder aus der Bildquelle, kein erzeugtes (E06, #2278) — kein Punkt fällt anders; und ein gestauchtes Feld bricht um wie jedes andere, die Stauchung folgt aus der breitesten Zeile (#2308) — ein einzeiliges Feld ohne harten Umbruch zeichnet wie vorher; und gestaucht wird nach der harten Breite ohne weichen Schwanz, und ob eine Zeile füllt und links beginnt, entscheidet jede für sich (Befunde 10 und 11 des Zweitblicks über die Folgen) — eine kurze Zeile neben einer überlaufenden bleibt in ihrer Ausrichtung; und jede Schriftgröße liegt auf ¼ Punkt, wie MSE2 sie als wxFont setzt (#2302) — eine ganze oder halbe Punktgröße zeichnet wie vorher, ein verkleinerter Text bis zu ⅛ Punkt anders; und eine Zeile wird je Abschnitt in einem Zug gezeichnet, wie MSE2 es je Element tut (#2302) — eine Schrift, die über die Lücke kernt, zeichnet ihr Paar jetzt mit; und der Puffer eines Zugs ist so breit wie sein gezeichneter Text (Nachlauf über die Folgen, Befund 3) — ein ungestauchter Zug zeichnet wie vorher **`7` seit dem 2026-10-14** — der Plan nennt das Paket, mit dem er gebaut ist (#2314); kein Punkt fällt anders | `TestDieRasterfassungPasstZumRasterer` |
| **Das Protokollformat** | `1` | `domain.Protokollfassung` | steht im Freigabemanifest; nennt eine Freigabe sie nicht, bleibt die Verträglichkeit **ungeklärt** | `TestEinManifestOhneFormatfassungenIstUngeklaert` |

## Etiketten an einem gespeicherten Stand

**Diese Fassungen sagen, unter welchen Regeln ein Stand entstand** — sie werden
mit ihm gespeichert und nicht nachgeführt.

| Gegenstand | Fassung | steht in | Verhalten | Beleg |
| --- | --- | --- | --- | --- |
| **Das Grundprofil der Regeln** | `22` | `domain.Grundprofilfassung` | eine Partie friert das ganze Profil ein und behält es; ein Griff auf eine andere Fassung als die eingefrorene wird abgewiesen | `TestEinePartieTraegtIhrProfilUndDerGriffZeigtDarauf`, `TestEinePartieEntstehtUndLaesstSichSpielen` |
| **Das Commanderprofil** | `1` | `domain.Commanderfassung` | dito, mit eigener Geschichte | `TestEinGenanntesCommanderwerkGibtDasProfil` |
| **Die Kontextrichtlinie der KI** | `kontext-1` | `domain.Kontextpolicyfassung` | steht im Manifest jedes Kontextpakets | `TestDasManifestNenntSeineZaehlweise` |
| Das Schema des Suchindex | `1` | `domain.RetrievalschemaFassung` | steht im Manifest des Index | `TestEinIndexNenntDieFassungenSeinesAufbaus` |
| Die Fassung des Suchindex | `1` | `domain.RetrievalindexFassung` | dito | `TestEinIndexNenntDieFassungenSeinesAufbaus` |
| Die Zerteilung des Korpus | `karten-seiten-1` | `service.Zerteilfassung` | dito | `TestEinIndexNenntDieFassungenSeinesAufbaus` |

**Beim Suchindex ist das Etikett alles.** Nichts vergleicht es beim Suchen: ein
Index, der unter einer älteren Zerteilung gebaut wurde, bleibt aktiv, bis
jemand neu aufbaut. Das ist eine **Grenze**, keine Zusage — wer die Zerteilung
ändert, sagt es in den Hinweisen der Freigabe.

## Fremde Fassungen

**Sie gehören anderen** — einem Dateiformat, einem Dienst. Dieses Programm
liest oder schreibt sie, bestimmt sie aber nicht.

| Gegenstand | Fassung | steht in | Verhalten | Beleg |
| --- | --- | --- | --- | --- |
| **MSEs Setdatei** | `2.0.0` | `service.standardfassung` | die Fassung, die ein neu zusammengesetztes Set nennt; ein gelesenes behält seine | `TestEinSetOhneQuelleWirdFrischZusammengesetzt` |
| **MSEs Symboldatei** | `0.3.5` | `msepaket.SymbolFassung` | geschrieben, wie MSE2 sie schreibt, und von MSE2 gezeichnet | `TestEineGezeichneteZeichnungWirdZuEinerDatei`, `TestMse2ZeichnetUnsereSymboldatei` |
| **MTG JSON** | Hauptfassung `5` | `mtgjson.bekannteHauptfassung` | eine andere Hauptfassung wird **gelesen** und als Befund geführt, nicht verweigert | `TestDieFassungWirdGelesen` |
| **Anthropic** | `2023-06-01` | `anthropic.Fassung` | Kopfzeile `anthropic-version` jeder Anfrage | `TestJedeAnfrageNenntSchluesselUndFassung` |

## Ohne eigene Fassung

**Was hier steht, trägt keine Zahl — und aus einem Grund.**

| Gegenstand | warum keine eigene |
| --- | --- |
| **Zusätze in MSE-Dateien** | es gibt keine: geschrieben wird, was gelesen wurde, **Byte für Byte** (`TestDerRundlaufUeberAllePakete`), und was hier dazukommt — die Herkunft eines Wertes etwa — steht im Bestand, nicht in der Datei |
| Die Ankündigung im eigenen Netz | `_mse4mac._tcp` mit den Einträgen `kennung` und `name`; eine Antwort, die nicht so aussieht, wird **gezählt** und nicht gedeutet (§16). Seit 0.5.0 unverändert |
| Das Diagnosepaket | geht an einen Menschen und wird nicht zurückgelesen; es nennt die Fassung der Anwendung und den Bau |
| Die Auskunft über die eigenen Daten | geht an den Menschen, dem sie gehören, und wird nicht zurückgelesen |
| Bild, SVG, PDF, MTG JSON als Ausfuhr | ihre Fassung ist die des jeweiligen Formats |

## Bau und Bindungsvertrag

**Zur Diagnose und nicht als zweite Produktfassung** (ENG-06.AK-01, AK-04).
`mse meta`, `GET /meta`, das MCP-Werkzeug `lesen.meta`, der Diagnosetext auf
der Technikseite und das Diagnosepaket nennen **dieselben** Angaben aus
derselben Quelle: die Fassung, den **Commit**, aus dem gebaut wurde, ob der
Baum dabei sauber war, und den **Abdruck des Bindungsvertrags**
(`TestMetaUndDiagnoseNennenDenselbenBau`, `TestDasDiagnosepaketNenntDenselbenBau`,
`TestDieMetaauskunftReichtBauUndVertragDurch`).

**Die Bindungen von Wails können nicht veralten** — unter Wails 2 entstand
`frontend/wailsjs/` bei jedem Bau neu, und seit Wails 3 entsteht gar keine
Datei mehr: gerufen wird über den Namen (`anbindung.ts`, ENG-02). **Was
veralten kann, ist die Fassade, die ruft**: sie ist von Hand geschrieben
(`frontend/src/api.ts` sagt warum), und eine umbenannte JSON-Marke übersetzt
`tsc` klaglos. Deshalb hält `frontend/src/bindungsvertrag.json` den Abdruck
jeder gebundenen Methode fest, wie das Fenster ihn zuletzt gesehen hat; ändert
sich einer in Go, ist der Wächter rot und nennt die Methode. **Was er nicht
hält**: ob die Typen im Fenster danach stimmen — er erzwingt einen Blick, den
Weg selbst geht der Nahtlauf.

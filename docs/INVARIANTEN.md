# Die Invarianten des Bestands

**Was im Bestand immer gilt, wo es gehalten wird und woran das zu sehen ist**
— ENG-05 (#1730).
Die Zahlen sind gezählt am 2026-09-23 an einem frisch migrierten Bestand
(Wanderung 0080): 84 eigene Tafeln und 2 virtuelle, 76 Fremdschlüssel,
77 Eindeutigkeiten, 68 `CHECK`-Bedingungen, 116 Indizes, 19 Auslöser.

## Drei Orte, und keiner ersetzt den anderen

| Ort | Was er hält | Warum dort |
| --- | --- | --- |
| **Domäne** (`Validate`, Wertetypen) | jede Regel über **eine** Entität — Menge, Zustand, Währung, Name | sie sagt, **welches Feld** und **was zu tun ist** (§6, §8) |
| **Ablage und Dienst** | Regeln über **mehrere** — Fassung, Rechte, „der letzte bleibt" — und jede Klemme auf das Profil | sie brauchen den Bestand und eine Transaktion (§2) |
| **Schema** | die zweite Reihe: Fremdschlüssel, Eindeutigkeit, `CHECK`, `NOT NULL`, benannte Auslöser | sie gilt auch für einen Weg, der an der Domäne vorbeiführt — eine Wanderung, ein Fehler im Code, ein zweiter Prozess |

**Kritisch** heißt eine Regel, deren Bruch **Besitz, Geld oder Rechte**
falsch macht. Sie steht an zwei Orten, und beide haben einen Lauf: die Domäne,
damit die Meldung am Feld steht, und das Schema, damit sie auch dann gilt, wenn
die Domäne übergangen wird. Die übrigen stehen an einem Ort, und die Tabelle
sagt, an welchem.

**Was ein Schemabruch an den Flächen heißt** (`storage/fehler.go`,
`storage/eindeutigkeit.go`): Eindeutigkeit → `doppelt`, `doppelterName` oder
`doppeltesExemplar`, mit Feld; `CHECK`, `NOT NULL`, Auslöser und Typ →
Validierung; Fremdschlüssel → nicht gefunden; voller Datenträger → `voll`.
**Keiner davon kommt mehr als „unerwarteter Fehler" an**, und nur ein
überholter Stand heißt noch „lade neu".

## Die Wächter

Jede Klasse unten hat einen Lauf, der **das Schema selbst** befragt —
`pragma_table_info`, `pragma_foreign_key_list`, `pragma_index_list` an einem
frisch migrierten Bestand, nicht eine Liste daneben
(`internal/storage/invarianten_intern_test.go`). Eine neue Tafel, eine neue
Menge, ein neuer Fremdschlüssel erzwingt eine Entscheidung: eine Bedingung
im Schema, oder eine Zeile in den Listen dort, **mit Grund**.

**Was sie nicht halten**: ob eine Bedingung das Richtige verlangt (dafür
stehen die Verhaltensläufe in der Tabelle), ob ein Grund stimmt (das Review),
und die Übergänge eines Zustands (die Domäne).

## Fremdschlüssel

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Jede Verbindung schaltet `foreign_keys` ein, und eine Verletzung heißt „nicht gefunden" | Datenquelle (`datenbank.go`), Einordnung | `TestEineFremdschluesselverletzungIstNichtGefunden`, `TestOhneFremdschluesselBleibtEineWaiseZurueck` |
| Jede Kennungsspalte (`*_id`) verweist — oder steht mit Grund in `kennungenOhneVerweis` (sechs, darunter der Verlauf, der seine Karte überdauert, damit ihr Löschen sich zurücknehmen lässt) | Schema | `TestJedeKennungsspalteVerweistOderStehtBegruendet` |
| **Jeder Fremdschlüssel hat einen Index, der mit seinen Spalten beginnt** — sonst sucht SQLite beim Löschen der Elternzeile jede Kindzeile ab | Schema (0080) | `TestJederFremdschluesselHatSeinenIndex`, Messung in [`ABFRAGEPLAENE.md`](ABFRAGEPLAENE.md) |
| Nach einer Wanderung zeigt kein Fremdschlüssel ins Leere | Schema | `TestEineSicherungAusNullSiebenWirdZurueckgespieltUndGeoeffnet` (0.7.0 bis heute; ältere steigen seit [0189](entscheidungen/0189-nur-die-grundlinie.md) über 0.7.0 auf), Datenpflege `fremdschluessel` |

**Offen**: `matchbindungen.partie_id` und `matchmeldungen.partie_id` verweisen
nicht, und 0067 nennt keinen Grund. Die Datenpflege löscht beendete Partien
nach der Frist; die Bindung bleibt und zeigt dann ins Leere. Ob das so sein
soll, entscheidet eine spätere Durchsicht — bis dahin steht es hier und nicht
still.

## Eindeutigkeit

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Ein Name je Deck, Turnier und Format je Profil und Spiel; je Profil eines Nutzers; je Nutzer; je Token eines Profils | Schema, Meldung am Namensfeld | `TestEineVerletzteBedingungSagtWelche`, `TestEinZweitesDeckGleichenNamensStehtAmNamen`, `doppeltername.test.tsx` |
| **Ein Druck steht in einem Zustand, einer Sprache und einer Ausführung einmal in der Sammlung** — derselbe zweimal ist eine Dublette, keine fremde Änderung (F08) | Schema, Meldung `doppeltesExemplar` | `TestDerselbeDruckZweimalIstEineDublette` |
| **Ein Name je Symbol der Bibliothek und Profil**, ohne Rücksicht auf Groß und Klein — wer „Krone" und „krone" nebeneinander hat, weiß nicht, welches er einsetzt (DOM-02) | Schema (0091, `UNIQUE` über `name COLLATE NOCASE`) | `TestEinSymbolGehtInDieBibliothekUndKommtWieder` |
| Zwei zugleich mit demselben Namen: einer gewinnt, der andere hört „gibt es schon" | Schema und `busy_timeout` | `TestZweiZugleichMitDemselbenNamenGebenEinenDoppelten` |
| Jede Zuordnung von Tafel und Spalten zu einem Feld zeigt auf eine Eindeutigkeit, die es gibt | `eindeutigkeit.go` | `TestJedeZuordnungZeigtAufEineEindeutigkeit` |
| **Eine eindeutige Bedingung an einer Tafel mit `profil_id` nennt das Profil** — oder steht mit Grund in `schluesselOhneProfil` (drei, und jeder sagt, **wer** die Kennung vergibt). Sonst träfe dieselbe Kennung in zwei Profilen **eine** Zeile: eine Leihe von einem Profil ins andere schrieb den Stand des Entleihers in die Zeile des Verleihers (Zweitblick über Phase 6.1, Befund 1), und die Probe desselben Modells galt im zweiten Profil als fremder Auftrag (Nachlauf, Befund 4) | Schema (0081, 0082) | `TestJederSchluesselEinerProfiltafelNenntDasProfil`, `TestEineLeiheZwischenZweiProfilenBleibtZweiZeilen`, `TestDieselbeKennungInZweiProfilenSindZweiAuftraege` |

## Mengen

| Regel | Sitz | Beleg |
| --- | --- | --- |
| **Eine Sammlungszeile hat mindestens ein Stück** — sonst gäbe sie einen Besitz vor, den es nicht gibt | Domäne **und** Schema (Auslöser `sammlung_regeln_*`, 0080) — **kritisch** | `TestExemplarValidateBefunde`, `TestDieSammlungNimmtNurGueltigeZeilen` |
| Deckplätze, Deckzuordnungen, Zusagen, Turnierdeckplätze, Journalzeilen und Rückkehrmengen haben mindestens eines | Schema (`CHECK`) | `TestJedeMengeUndJederBetragHatIhreBedingung` |
| **Ein Symbol hat höchstens 4.096 Teile, 65.536 Punkte, 32 Ebenen und 1 MiB** — gemessen am installierten Pack lag das Größte bei 73 Punkten und 11 KB; ein Name hat höchstens 80 Zeichen (DOM-02.AK-04) | Paketleser (`msepaket.PruefeSymbolumfang`, `GrenzeSymboldatei`), Domäne (`PruefeSymbolname`) und Schema (`CHECK`, 0091) | `TestZuVieleTeilePunkteOderEbenenWerdenAbgewiesen`, `TestEineZuGrosseOderZuVerwickelteDateiWirdNichtGelesen`, `TestEinSymbolnameWirdGeprueft` |
| Die Anzahl eines Packpostens | nur gelesen — fremde Daten, MSE2 kennt keine Untergrenze (`src/data/pack.hpp`, Zeile 68, Fassung f9d9356 vom 2021-07-06) | dito, als Ausnahme |

## Währung und Skala

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Ein Betrag steht als ganze Zahl in der kleinsten Einheit seiner Währung ([0127](entscheidungen/0127-ein-betrag-steht-in-der-kleinsten-einheit.md)) | Domäne (`Betrag`, `Geld`) | `geld_test.go` |
| Kostenbudgets und -posten: nicht negativ, Währung mit drei Zeichen | Domäne **und** Schema (`CHECK`) — **kritisch** | `TestEineVerletzteBedingungSagtWelche` (CHECK), `TestEinKrummerAbschlussbetragKommtNichtDurch` |
| **Die Grenze eines Nachweises ist nicht negativ, und mit Betrag nennt sie ihre Währung** | Domäne **und** Schema (Auslöser `freigabenachweise_grenze`, 0080) — **kritisch** | `TestOhneGrenzeZaehltDerBetragNicht`, `TestEinNachweisTraegtEineGueltigeGrenze` |
| Erwerbspreis und -währung einer Sammlungszeile | Domäne (`Exemplar.Preisbefunde`); kein Geldfluss hängt daran | Ausnahme mit Grund. Die Vorgabe `''` der Spalte stammt aus 0018 und wird von keinem Schreibweg genommen — jeder schreibt eine Zahl |
| Preisstände | fremde Daten eines Dienstes, gelesen, wie sie kommen (§16) | Ausnahme mit Grund |

## Zustände und ihre Übergänge

**Die Übergänge stehen in der Domäne und nirgends sonst** — ein Auslöser,
der sie nachbaute, wäre die „versteckte zweite Geschäftslogik", die AK-02
ausschließt.

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Zustand und Ausführung einer Sammlungszeile kommen aus der Aufzählung der Domäne — sonst ginge „Foil" neben „foil" an der Dublettenerkennung vorbei | Domäne **und** Schema (0080) — **kritisch** | `TestDieSammlungNimmtNurGueltigeZeilen` schreibt **jeden** Wert der Domäne; ein neuer ohne Wanderung wird dort rot |
| Eine alte Zeile bleibt bearbeitbar: geprüft wird, was sich ändert, nicht was dasteht | Schema (0080, in der Grundlinie) | `TestEineAlteZeileBleibtBearbeitbar` — die Zeile wird seit [0189](entscheidungen/0189-nur-die-grundlinie.md) am Auslöser vorbei geschrieben |
| Ein überholter Stand wird abgewiesen, nicht überschrieben (Fassung an Deck, Format, Exemplar, Nutzer, Profil, Token) | Ablage | `TestEinUeberholterStandHeisstSo` |
| Ein Kostenposten ist `reserviert`, `abgerechnet`, `freigegeben` oder `unklar`; ein zweiter Abschluss bucht nicht zweimal | Schema (`CHECK … IN`) und Ablage | `TestEinZweiterAbschlussIstEinKonflikt` |
| Eine Suchgeneration `baut`, ist `fertig` oder `verworfen`, und höchstens eine ist aktiv | Schema (`CHECK`, teilweise Eindeutigkeit) | `retrievalaufbau_test.go` |
| Ein abgeschlossenes Turnier ändert sich nicht — kein Anmelden, keine Runde, kein Ergebnis —, bis es wiedereröffnet ist (Vermerk 0139) | Dienst (`Turnierleitung`) **und** Ablage (`fortschreiben` trifft kein abgeschlossenes Turnier) | `TestEinAbgeschlossenesTurnierWeistJedeAenderungAb`, `TestEinAbgeschlossenesTurnierNimmtNichtsMehrAn` |
| Jede Änderung an einem Turnier schreibt eine Zeile ins Protokoll, im selben Zug; der Vorgang kommt aus einer festen Aufzählung | Ablage (`fortschreiben`, `Anlegen`, `Einlesen`) und Schema (`CHECK … IN`, 0089) | `TestDasProtokollNenntJedeAenderung` |
| Nur ein **Entwurf** — Revision 1 im Zustand `entwurf`, ohne zweite — wird ersetzt oder gelöscht; ins Archiv geht nur ein beendeter Handel (Vermerk 0148) | Ablage (die Bedingung steht **in** der Anweisung, denn zwei Prozesse schreiben) **und** Dienst (`Handelsverzeichnis`, für den Satz) | `TestNurEinEntwurfLaesstSichErsetzenUndLoeschen`, `TestNurEinEntwurfWirdBearbeitet`, `TestLoeschenUndArchiv` |
| Ein eingelesenes Turnier entsteht **ganz oder gar nicht** — Turnier, Teilnehmende, Runden und Ergebnisse in einem Zug; eine Paarung, deren Seite es nicht gibt, bricht ihn ab (Vermerk 0140) | Ablage (`Turnierablage.Einlesen`) und Domäne (`ZerlegeTurnier` prüft Runden ab 1 ohne Lücke, je Runde mindestens eine Paarung und höchstens einen Bye) | `TestEinFehlerhaftesEinlesenLaesstNichtsZurueck`, `TestEineLueckeImVerlaufBrichtDasEinlesenAb` |
| **Ein TOTP-Schritt gilt einmal, und Codeversuche haben ein Budget je Nutzer** — nur ein späterer Schritt als der zuletzt angenommene gilt; jeder Codeweg meldet seinen Versuch vorab an, ab dem fünften sperrt `domain.Faktorsperre`, ein Erfolg löscht den Haushalt ([0178](entscheidungen/0178-ein-code-gilt-einmal-und-raten-hat-ein-budget.md)) | Ablage (`SchrittVerbrauchen` mit der Bedingung im `WHERE`, `FaktorversuchBeginnen` in einer Transaktion) und Schema (`CHECK`, 0097) | `TestEinTOTPSchrittGiltEinmal`, `TestGleichzeitigeVersucheTeilenSichDasBudget`, `TestFehlversucheHabenEinBudget`, `TestDieWurzelVerbrauchtDenCodeUndZaehltDieFehlversuche` |
| **Eine ausgestellte Einladung wird einmal eingelöst** — eine fremde, verbrauchte oder abgelaufene Nonce legt nichts an; wer einlöst, wird **unbestätigt** aufgenommen, ein Bestätigter bleibt es, ein Gesperrter löst nichts ein und die Nonce bleibt unverbraucht. `eingeloest` und `kennung` stehen zusammen oder gar nicht ([0156](entscheidungen/0156-eine-einladung-genuegt.md)) | Ablage (`Kontaktablage.Einloesen`, eine Transaktion) und Schema (`CHECK`, 0096) | `TestEineEinloesungLegtDenKontaktUnbestaetigtAn`, `TestFremdeUndAbgelaufeneNoncesLegenNichtsAn`, `TestEinGesperrterLoestNichtsEin`, `TestEinloesungBrichtAbOhneSpur` |

## Sicherungsvermerke

**Die erste Wanderung nach der Grundlinie** (`0098`, M01,
#2345):
wann der Bestand gesichert wurde. Wer vermerkt, entscheidet der Aufrufer
(OPS01, #2347); was ein Vermerk ist, die Domäne.

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Ein Vermerk nennt Zeit (UTC, Millisekunden), Anlass (`hand`, `plan`, `wanderung`), Fläche (`fenster`, `cli`, `mcp`, `rest`), **den Namen der Datei und keinen Pfad** — der Bestand wird weitergegeben, und ein Pfad trüge den Namen des Benutzerkontos —, eine Größe über null und SHA-256 klein und hexadezimal | Domäne (`Sicherungsvermerk.Pruefe`) **und** Schema (`CHECK`, 0098) | `TestEinSicherungsvermerkWirdAbgewiesen`, `TestDieTafelWeistUngueltigeVermerkeAb` (Gegenprobe: ohne die Pfadbedingung rot) |
| Der jüngste ist der jüngste nach der Zeit, nicht nach dem Eintragen — zwei Prozesse sichern, und der langsamere vermerkt zuletzt | Ablage (`Sicherungsvermerke.Letzter`) | `TestDerLetzteVermerkIstDerJuengste` |
| **Unbekannt heißt unbekannt**: ein Bestand ohne Vermerk nennt keinen Zeitpunkt; ein Zeitpunkt, den der `GLOB` durchlässt, der aber kein Datum ist, ist ein Fehler und nicht „nie" | Ablage | `TestEinAltbestandBekommtDieSicherungsvermerke`, `TestEinUnlesbarerVermerkIstEinFehler`, `TestOhneDieTafelIstEsEinFehler` |
| **Ein Vermerk steht erst nach zurückgelesener und geprüfter Datei** — und nur für eine, die ihren Namen bekam; Größe und Prüfsumme sind die der Datei auf der Platte (OPS01) | Dienst (`Bestandsweg.vermerken` nach `sicherung.PackenGeprueft`) | `TestEineGeprueftSicherungWirdVermerkt`, `TestEineGescheiterteSicherungVermerktNichts`, `TestEineVerdorbeneSicherungBekommtIhrenNamenNicht` |
| **Das Alter fragt nach einer Sicherung, die man wegtragen kann** — die vor einer Wanderung zählt nicht (OPS01) | Ablage (`Sicherungsvermerke.Letzter`) | `TestDerLetzteVermerkUebergehtDieSicherungVorDerWanderung` |

## Der Stand eines Sets

**Die zweite Wanderung nach der Grundlinie** (`0099`, PERF02,
#2352): ein
Zähler je Set, den jede Änderung an einer gezeichneten Karte hebt. Eine
Problemprüfung trägt den Stand, mit dem sie begann; wer liest, vergleicht.

| Regel | Sitz | Beleg |
| --- | --- | --- |
| **Jeder Weg, der eine gezeichnete Karte verändern kann, hebt den Zähler** — Karte, Kartenwert, Wert, Stiloption und Schlüsselwort des Sets, Spiel, Stil, Name, der Ort der Bilder und die Fassungen; auch ein Weg ohne Verlaufsschritt | Schema (siebzehn Auslöser `set_aenderungen_*`, 0099 und 0100) | `TestJederWegZurKarteHebtDenSetstand` (Gegenprobe: ohne die Wanderung rot; Stiloptionen und Ort der Bilder fehlten bis 0100) |
| **Das Aufschlagen hebt ihn nicht**, und eine Änderung an einem Set hebt nur dessen | Schema (`set_aenderungen_set` nur über die genannten Spalten) | dito |
| **Wer schreibt, was schon dasteht, hebt ihn nicht** — jeder Auslöser auf `UPDATE` vergleicht `OLD` mit `NEW`, und zwar **jede** Spalte seiner Tafel außer den benannten von `sets` | Schema (`WHEN … IS NOT …`, 0100) | `TestUnveraendertesSchreibenHebtDenSetstandNicht`, `TestJederAusloeserVergleichtJedeSpalte` (Gegenprobe: ein Vergleich weniger macht ihn rot) |
| Jedes Set hat genau eine Zeile; ein gelöschtes keine mehr | Schema (Fremdschlüssel mit `ON DELETE CASCADE`, `set_aenderungen_set_neu`) | `TestEinUnbekanntesSetHatKeinenStand` |
| Ein Stand, dessen Abdruck nicht der heutige ist, ist veraltet; eine abgebrochene Prüfung ist unvollständig und nie „0 Probleme" | Domäne (`Problemstand.Urteil`, `Problemstand.Zahl`) | `TestDerProblemstandSagtWieWeitErGilt`, `TestEinUnsichererStandBehauptetKeinNull` |

## Abgewiesene Gegenbelege

**Die vierte Wanderung nach der Grundlinie** (`0101`,
#2500):
ein Gegenbeleg, der nicht passt, wird am eigenen Beleg festgehalten — von wem,
welchen Abdruck er nannte, woran er scheiterte und mit welchem Satz.

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Ein Befund nennt Vorgang, Absender, einen der fünf Mängel (`unterschrift`, `vorgang`, `revision`, `parteien`, `zeilen`) und einen Grund | Domäne (`Gegenbelegbefund.Pruefe`) **und** Schema (`CHECK`, 0101) | `TestEinGegenbelegbefundNenntVorgangAbsenderMangelUndGrund`, `TestEinUngueltigerBefundKommtNichtInDenBestand` |
| **Ohne eigenen Beleg kein Befund** — wogegen wäre geprüft worden? Er geht mit dem Beleg | Schema (Fremdschlüssel auf `handelsbelege` mit `ON DELETE CASCADE`, 0101) | `TestOhneBelegGibtEsKeinenBefund` |
| **Einer je Vorgang, der jüngste gilt** — ein Gegenüber, das immer neue Belege schickt, füllt den Bestand nicht | Schema (Primärschlüssel) und Ablage (`Abweisen` ersetzt) | `TestDerJuengsteBefundGilt` |
| **Abgewiesen ist nicht gegengezeichnet**: ein Befund setzt `beidseitig` nicht, und er steht aus demselben Lesen wie der Beleg | Ablage (`Abschlussablage.Beleg`) | `TestEinAbgewiesenerGegenbelegStehtAmBeleg` |

## Objektbesitz

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Jede Zeile gehört einem Profil: über einen Fremdschlüssel mit `CASCADE` oder als benannte Wurzel (0035) | Schema | `TestJedeWurzelGehoertEinemProfil` |
| Eine Wurzel mit der Vorgabe 0 nimmt keine Zeile ohne Profil an | Schema (acht Auslöser `*_profil_*`) | dito |
| **Eine Profilkennung wird nie wieder vergeben** — sonst fielen die Wurzeln eines gelöschten Profils dem nächsten neuen zu | Schema (`AUTOINCREMENT` an `profile` und `nutzer`) | dito |
| Jede Abfrage auf eine Wurzel klemmt auf das Profil | Ablage | `TestJedeWurzelabfrageKlemmtAufEinProfil` |
| „Über alle Sets" heißt über die Sets des Profils | Ablage | `TestUeberAlleSetsHeisstUeberDieSetsDesProfils`, `TestDieSucheBleibtImProfil`, `TestDieSucheFindetKeineFremdeKarte` |
| **Ein Abkömmling fragt nach seiner Wurzel** — Karte, Kartenwert, Marke, Verweis, Verlauf, Deckplatz, Zuordnung, Herkunft, Teilnehmer, Runde, Paarung, Matchbindung, Partieereignis: eine fremde Kennung ist „nicht gefunden", beim Schreiben im selben Zug ([0131](entscheidungen/0131-ein-abkoemmling-fragt-nach-seiner-wurzel.md), #1815) | Ablage (`besitz.go`) | `TestEineFremdeKarteTrifftNichts`, `TestEinFremdesDeckTrifftNichts`, `TestEinFremdesTurnierTrifftNichts`, `TestEineFremdePartieTrifftNichts`, `TestEinNutzerSchreibtKeineFremdeKarteUm` |
| Eine Suche oder ein Nachschlagen über mehrere Kennungen gibt für eine fremde Kennung **nichts** — wie für eine fehlende | Ablage | `TestEineFremdeKennungGibtEineLeereAntwort` |
| **Eine Datei neben einem Set** — die Symbolzeichnung — wird nur gelesen, geschrieben oder nachgezeichnet, wenn das Set dem gezeigten Profil gehört; MCP und Dienst reichen dafür den Kontext des Rufs durch statt `context.Background()` | Anwendung (`symbolwerkstattnaht.go`, `symbolnachzeichnennaht.go`, `symbol.go`) | `TestDasSymbolEinesFremdenSetsTrifftNichts`, `TestDasFensterFasstKeinFremdesSymbolAn`, `TestDieSymbolpaletteUeberJedeFlaeche` |
| Eine Verbindung bleibt im Profil: kein Deckplatz auf eine fremde Karte, keine Zuordnung zu einem Platz eines anderen Decks, keine Paarung mit einem Teilnehmer eines anderen Turniers, keine Matchbindung an eine fremde Partie | Ablage (`platzImDeck`, `teilnehmerImTurnier`, `paarungImTurnier`) | `TestEinEigenesDeckNimmtNichtsFremdes`, `TestEinEigenesTurnierZeigtAufNichtsFremdes` |
| Keine Methode einer Abkömmlingsablage fehlt in diesen Läufen, und jede exportierte Methode der Ablage erreicht eine Klemme — oder ist begründet ausgenommen | Wächter | `TestJedeAbkoemmlingsmethodeWirdAusEinemFremdenProfilGerufen`, `TestJedeAblagemethodeFragtNachDemProfil` |

| **Wer ein Profil oder ein Token nennt, braucht die Rolle an dessen Profil** — Zugriffstoken, Vorgaben, Freigaben; im gezeigten Profil sagt ein Schreibweg ohne Rolle „nur lesen" wie jeder andere (#1817) | Anwendung (`zielprofil.go`) | `TestTokenUndVorgabenBrauchenDieRolleAmZielprofil`, `TestEinLeserLegtKeinTokenAn`, `TestJedeBindungMitZielprofilFragtNachDerRolle` |

## Referenzierte Dateien

| Regel | Sitz | Beleg |
| --- | --- | --- |
| Ein Bild des Nutzers, ein Abzeichen, ein Kartenbild steht als **Name** im Bestand, nie als Pfad — ein Pfad zeigte nach der Weitergabe auf einen fremden Rechner | Schema-Kommentar und Ablage (`Bildablage.Wegnehmen` weist einen Weg ab) | `bildablage_test.go` |
| Eine Datei wird abgelegt, **bevor** ein Feld auf sie zeigt, und bei einem Fehlschlag wieder weggenommen (#1163) | Dienst | `bildablage_test.go` |
| Eine Datei wird ganz oder gar nicht ersetzt | Dienst (`schreibeAtomar`) | `TestEinAbbruchMittenImSchreibenLaesstDieAlteDateiStehen` |

## Die Auslöser

**Jeder Auslöser steht hier, mit der einen Invariante, die er hält.** Der
Lauf `TestJederAusloeserStehtImKatalog` liest diese Tabelle und das Schema
und wird rot, wenn einer fehlt — in der einen oder der anderen Richtung.

| Auslöser | Tafel | Invariante |
| --- | --- | --- |
| `sets_profil_pflicht` | `sets` | keine Zeile ohne Profil |
| `sets_profil_bleibt` | `sets` | dito beim Ändern |
| `decks_profil_pflicht` | `decks` | keine Zeile ohne Profil |
| `decks_profil_bleibt` | `decks` | dito beim Ändern |
| `turniere_profil_pflicht` | `turniere` | keine Zeile ohne Profil |
| `turniere_profil_bleibt` | `turniere` | dito beim Ändern |
| `formate_profil_pflicht` | `formate` | keine Zeile ohne Profil |
| `formate_profil_bleibt` | `formate` | dito beim Ändern |
| `sammlung_regeln_neu` | `sammlung` | Menge ab eins, Zustand und Ausführung aus der Aufzählung, eine Sprache |
| `sammlung_regeln_um` | `sammlung` | dito, für die Spalte, die sich ändert |
| `freigabenachweise_grenze` | `freigabenachweise` | die Grenze ist nicht negativ und nennt mit Betrag ihre Währung |
| `freigabenrevision_zeiger_neu` | `bestand` | die Revision der Rechte zählt jeden Zeigerwechsel (0047) |
| `freigabenrevision_zeiger_um` | `bestand` | dito |
| `freigabenrevision_freigabe_neu` | `profilfreigaben` | die Revision zählt jede Freigabe |
| `freigabenrevision_freigabe_um` | `profilfreigaben` | dito |
| `freigabenrevision_freigabe_weg` | `profilfreigaben` | dito |
| `card_values_eingefuegt` | `card_values` | der Volltextindex folgt jedem Kartenwert |
| `card_values_geaendert` | `card_values` | dito |
| `card_values_geloescht` | `card_values` | dito |
| `objektmarken_set_weg` | `sets` | wenn ein Objekt geht, gehen seine Marken (0135) — keine Marke an einem Set, das es nicht gibt |
| `objektmarken_exemplar_weg` | `sammlung` | dito, am Exemplar |
| `objektmarken_deck_weg` | `decks` | dito, am Deck |
| `objektmarken_paket_weg` | `packages` | dito, am Paket — sobald es nicht mehr eingebunden ist; ein neu bezogenes behält seine Katalogzeile und damit seine Marken |
| `objektmarken_turnier_weg` | `turniere` | dito, am Turnier |
| `objektmarken_partie_weg` | `partien` | dito, an der Partie |
| `objektmarken_handel_weg` | `handelsfolge` | dito, am Handelsvorgang, und nur im Profil, dem er gehört |
| `objektmarken_symbol_weg` | `symbolbibliothek` | dito, am Symbol der Bibliothek (DOM-02, 0091) |
| `handelsrevisionen_letzte_weg` | `handelsrevisionen` | geht die letzte Revision eines Vorgangs — ein gelöschter Entwurf —, gehen seine Marken und sein Archivvermerk mit (TRADE-01, 0093); ein Entwurf hat keine `handelsfolge`, an der der Auslöser darüber hängt |
| `handelsprotokoll_neu` | `handelsrevisionen` | jede neue Revision hinterlässt eine Protokollzeile „'' → Zustand" (TRADE-03.AK-03, 0095) — auf jedem Weg, der eine Revision schreibt: Leitung, Briefkasten, Abschluss |
| `handelsprotokoll_wechsel` | `handelsrevisionen` | ändert sich der Zustand einer Revision, hält das Protokoll „vorher → nachher" fest; dieselbe Zeile noch einmal ohne Wechsel hinterlässt nichts (0095) |
| `handelsprotokoll_bleibt` | `handelsprotokoll` | ein Protokolleintrag wird nicht geändert (0095) |
| `handelsprotokoll_geht_mit_dem_profil` | `handelsprotokoll` | ein Protokolleintrag wird nur mit seinem Profil gelöscht — die Bedingung fragt, ob es das Profil noch gibt (0095) |
| `handelsbeleg_bleibt` | `handelsbelege` | ein Beleg wird nicht umgeschrieben; der einzige Wechsel ist `beidseitig` von 0 auf 1 durch die Gegenzeichnung (0095) |
| `handelsbeleg_geht_mit_dem_profil` | `handelsbelege` | ein Beleg wird nur mit seinem Profil gelöscht (0095) |
| `handelsjournal_bleibt` | `handelsjournal` | eine Journalzeile wird nicht geändert — eine Korrektur ist eine neue Zeile mit Anlass `korrektur` (0074, 0095) |
| `handelsjournal_geht_mit_dem_profil` | `handelsjournal` | eine Journalzeile wird nur mit ihrem Profil gelöscht (0095) |
| `handelsfremdbeleg_bleibt` | `handelsfremdbelege` | ein Fremdbeleg wird nicht geändert; ein zweiter, abweichender ist ein Konflikt im Dienst und kein stilles `DO NOTHING` (0095) |
| `handelsfremdbeleg_geht_mit_dem_profil` | `handelsfremdbelege` | ein Fremdbeleg wird nur mit seinem Profil gelöscht (0095) |
| `set_aenderungen_set_neu` | `sets` | ein neues Set bekommt seine Zeile im Setstand (0099) |
| `set_aenderungen_set` | `sets` | Spiel, Stil, Name, der Ort der Bilder oder ihre Fassung heben den Setstand, **wenn sie sich ändern** — das Aufschlagen nicht (0099, seit 0100 mit `herkunft` und Vergleich) |
| `set_aenderungen_karte_neu` | `cards` | eine neue Karte hebt den Stand ihres Sets (0099) |
| `set_aenderungen_karte_weg` | `cards` | dito, eine gelöschte |
| `set_aenderungen_karte` | `cards` | dito, eine verschobene — in beiden Sets, und nur, wenn sie sich wirklich bewegt (seit 0100) |
| `set_aenderungen_wert_neu` | `card_values` | ein neuer Kartenwert hebt den Stand des Sets (0099) |
| `set_aenderungen_wert` | `card_values` | dito, ein geänderter — auch gleich lang, und nicht derselbe Wert noch einmal (seit 0100) |
| `set_aenderungen_wert_weg` | `card_values` | dito, ein gelöschter |
| `set_aenderungen_setwert_neu` | `set_values` | ein neuer Wert des Sets hebt seinen Stand (0099) |
| `set_aenderungen_setwert` | `set_values` | dito, ein geänderter — nicht derselbe noch einmal (seit 0100) |
| `set_aenderungen_setwert_weg` | `set_values` | dito, ein gelöschter |
| `set_aenderungen_wort_neu` | `set_keywords` | ein neues Schlüsselwort hebt den Stand des Sets (0099) |
| `set_aenderungen_wort` | `set_keywords` | dito, ein geändertes — nicht dasselbe noch einmal (seit 0100) |
| `set_aenderungen_wort_weg` | `set_keywords` | dito, ein gelöschtes |
| `set_aenderungen_stil_neu` | `styling_values` | eine neue Stiloption hebt den Stand des Sets (0100) |
| `set_aenderungen_stil` | `styling_values` | dito, eine geänderte — nicht dieselbe noch einmal |
| `set_aenderungen_stil_weg` | `styling_values` | dito, eine gelöschte |

## Erprobt: Wanderung, zwei Prozesse, Abbruch, Integrität (AK-04)

| Frage | Lauf |
| --- | --- |
| Öffnet sich ein Bestand aus 0.7.0 hier — heil und ohne Verweis ins Leere? | `TestEineSicherungAusNullSiebenWirdZurueckgespieltUndGeoeffnet` (`integrity_check`, `foreign_key_check`) |
| Scheitert eine Wanderung, bleibt der alte Stand? | `TestEinFehlschlagInDerMitteLaesstNichtsZurueck`, `TestEineAbweisendeBestandspruefungNimmtDieFolgeZurueck` |
| Stirbt ein Prozess mitten im offenen Geschäft — ist die Datei heil, steht nichts davon da, schreibt der Nächste sofort? | `TestEinAbbruchMittenImGeschaeftLaesstDenAltenStand` (SIGKILL an einen Kindprozess, Gegenprobe: ein festgeschriebenes Geschäft macht ihn rot) |
| Stirbt ein Prozess nach der Antwort — bleibt, was quittiert war? | `TestEinHartGetoeteterProzessLaesstDenBestandBrauchbar` |
| Zwei Programme an einem Bestand | [0045](entscheidungen/0045-zwei-programme-an-einem-bestand.md), `belegt_test.go`, `TestZweiZugleichMitDemselbenNamenGebenEinenDoppelten` |
| Ein voller Datenträger | `TestEinVollerDatentraegerLaesstDenAltenStandStehen` — und seit #1812 heißt er so |
| Hat ein Zugriffstoken eine halbe Sitzbindung — eine Partie ohne Sitz? | Nein: `CHECK ((sitz_partie = '') = (sitz = ''))` in Wanderung 0083, vorher `Sitzbindung.Validate` am Feld (#1819); `TestEineSitzbindungGehtInDenBestandUndZurueck` |
| Scheitert eine Deckübernahme mitten drin — steht dann ein halbes Deck? | Nein, seit #1820: Deck, Plätze und Herkunft in **einer** Transaktion (`storage/deckuebernahme.go`). `TestEineGescheiterteUebernahmeHinterlaesstNichts` (dritter Platz, Deck, Herkunft), an allen vier Flächen `TestEineGescheiterteUebernahmeHinterlaesstKeinDeck` und `TestEineGescheiterteUebernahmeHinterlaesstAmFensterKeinDeck` |
| Neue Datengrenzen nur im Fenster? | Keine: jede Regel oben steht in Domäne oder Schema; die Oberfläche zeigt sie nur |

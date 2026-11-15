# Lastmessung — was Dienst und Werkzeugserver unter Last tun

**QA-08** (#1739):
Profile für Lesen, Schreiben, Rendern, Einfuhr und lange Aufträge, MCP über
stdio und über HTTP getrennt, gemischte Klienten (AK-01); Grenzen je
Identität, teurer Operation, Parallelauftrag und Ressource (AK-02); die
abschaltbaren Kontingente (AK-03,
[0161](entscheidungen/0161-die-kontingente-lassen-sich-abschalten.md));
Durchsatz, p95/p99, Fehlerrate, Speicher und Erholung (AK-04).

**Die Zahlen stehen im Beleg** —
[`belege/lastlauf.md`](belege/lastlauf.md) und daneben als JSON, mit Commit,
Maschine, Eingaben und jedem gefahrenen Fall. Hier steht, **wie** gemessen
wird, **was** die Zahlen sagen und **was nicht**.

## Wie gemessen wird

`node scripts/lastlauf.mjs` (`make lastlauf`). **Gegen frisch gebaute
Programme und einen eigenen Bestand**, nie gegen den echten: der Lauf baut
`mse`, `mse-rest` und `mse-mcp` in einen Ordner unter `TMPDIR`, legt dort ein
eigenes Zuhause an (`HOME`), darin den Bestand (`MSE_BESTAND`) mit dem
Probeset aus `testdaten/` und der Saat als Paketen, und gibt jedem Prozess
einen Schlüsselbund, den es nicht gibt (`MSE_SCHLUESSELBUND`). Zwei Token
unter zwei Maschinennamen geben die zweite Identität.

**Ein Fall ist ein Profil gegen frisch gestartete Dienste** — `mse-rest` und
`mse-mcp` über HTTP, beide auf `127.0.0.1`. Frisch, weil Guthaben und
Verteilung im Arbeitsspeicher eines Prozesses leben: ein zweites Profil gegen
denselben Prozess erbte das leere Guthaben des ersten. **Jedes Profil läuft
zweimal**: mit Kontingenten (die Vorgabe) und ohne (`grenzen.kontingente
aus`). Ohne zeigt, was die Maschine kann; mit zeigt, was der Schutz daraus
macht. Die Vorgabe ist 8 s je Fall und 8 Klienten (`DAUER`, `KLIENTEN`).

| Profil | Frage | Weg | Klienten |
| --- | --- | --- | --- |
| `rest-lesen` | Lesen | `GET /sets/{id}/karten` und `GET /sets` im Wechsel | 8, ungeduldig |
| `rest-schreiben` | Schreiben | `PUT …/felder/name` reihum über die Karten | 8, ungeduldig |
| `rest-rendern` | Rendern | `POST /karten/{id}/bild` in eine eigene Datei | 8, ungeduldig |
| `rest-einfuhr` | Einfuhr | `POST /mtgjson` mit `testdaten/mtgjson/probe.json` — **nur der erste Ruf legt das Set an**, jeder weitere liest dieselbe Quelle und meldet `schonOffen` (am 2026-10-03 gegen einen frischen `mse-rest` nachgelesen: `angelegt` einmal `["1"]`, danach `[]`); die Zahl misst nach dem ersten Ruf die Erkennung eines Duplikats | 8, ungeduldig |
| `rest-auftraege` | lange Aufträge | `POST /sets/{id}/bogen` mit `Prefer: respond-async`, dann der Stand bis zum Ende; dazu ein Abbruch | 8, warten eine Drosselung ab |
| `mcp-stdio` | MCP über stdio | `lesen.karten`, je Prozess ein Klient | 4 Prozesse |
| `mcp-http` | MCP über HTTP | `lesen.karten` an `/mcp` | 8, ungeduldig |
| `gemischt` | gemischte Klienten | REST lesen 3, schreiben 2, rendern 1, MCP über HTTP 2, MCP über stdio 1 — zugleich | 9 |
| `fairness` | zwei Identitäten | Token A flutet `GET /sets` mit 6 Klienten, Token B fragt alle 250 ms | 7 |
| `ueberlast` | Überlast und Erholung | ein höflicher Klient (10 je s, wartet `Retry-After` ab) 2 s, dann 32 ungeduldige, dann wieder der höfliche 3 s | 1 / 32 / 1 |
| `langsam` | langsame Klienten | 20 Verbindungen mit halbem Kopf je Dienst, daneben je ein Klient alle 200 ms, 12 s | 1 je Dienst |
| `ruhend` | ruhende Verbindungen | ein Pool mit 16 Verbindungen fragt je einmal und lässt sie offen, daneben ein Klient alle 200 ms | 1 |

**Dazu einmal `rest-lesen` mit ausführlichem Protokoll und Unified Log** —
die Nachschärfung zum Issue verlangt es, weil dort einmal die Last zu finden
war und nicht im Behandler.

**Ungeduldig** heißt: ein Klient schickt den nächsten Ruf, sobald der vorige
zurück ist, und wartet eine Drosselung **nicht** ab. So zeigt die Spalte
*gedrosselt*, wie viel ein solcher Aufrufer abgewiesen bekommt — und wie
billig die Absage ist. **Der Klient hält höchstens sechzehn Verbindungen**
zum Dienst offen, so viele, wie der annimmt; was darüber geht, wartet beim
Klienten.

### Zwei Uhren

**Die Uhr des Klienten** misst jeden Ruf vom Schreiben der Anfrage bis zum
letzten Byte der Antwort, samt Leitung und Warteschlange — p50, p95 und p99
sind echte Perzentile über alle Rufe einer Gruppe. **Die Uhr des Dienstes**
steht in `GET /metriken` und `lesen.metriken` (OPS-05): sie misst ab dem
Behandler und kennt nur Eimer — 1, 5, 25, 100, 500, 2.500 ms und darüber.
Ein p95 von dort ist deshalb eine **Obergrenze** („höchstens 25 ms"). Das
Issue verlangt, p95/p99 dort abzulesen; der Beleg nennt beide. **Für MCP über
stdio** gibt es die Uhr des Dienstes nur im Profil `mcp-stdio` — dort wird
`lesen.metriken` am ersten der vier Prozesse gefragt.

### Was als Fehler zählt — und was nicht

**Eine Drosselung nicht.** `429` über REST und „zu viele Rufe" über MCP sind
die erwartete Schutzreaktion (AK-04) und stehen in einer eigenen Spalte. Ein
Fehler ist, was weder gelang noch gedrosselt wurde: eine abgerissene
Verbindung, eine abgelaufene Frist, ein 5xx, ein anderer 4xx, ein `isError`
aus einem anderen Grund.

**Und ein Ruf ohne Wirkung ist keiner, der gelang** (Z3 des Zweitblicks über
Phase 6.11, `scripts/lastlaufwirkung.mjs`). Ein 2xx allein zählte als Erfolg:
antwortete der Dienst 201 und schrieb kein Bild, stand der Ruf im Durchsatz.
Jede Probe trägt jetzt, was ihr Ruf zusagt, und **nach dem Profil** wird es an
einer Stichprobe nachgeprüft — 25 Rufe je Gruppe, gleichmäßig verteilt:

| Profil | Zusage | Prüfung |
| --- | --- | --- |
| `rest-rendern` | ein Bild am Ziel | die Datei ist da und ein **vollständiges** PNG (`pngpruefung.mjs`) |
| `rest-schreiben` | ein Feld im Bestand | `GET /sets/{id}/karten` nach dem Lauf: die Karte trägt einen der Werte, die für sie gelangen — ein späterer Schreiber darf ihn überschrieben haben |
| `rest-einfuhr` | ein Set aus der Datei | ein Set des Namens, das vorher nicht da war, mit den Karten der Datei |
| `rest-auftraege` | ein Bogen | bei einem fertigen Auftrag: ein PDF mit `%PDF-` am Anfang und `%%EOF` am Ende |
| Lesen über REST und MCP | eine Kartenliste, eine Setliste, die Grenzen | die Antwort ist, was sie sein soll — mit der Kartenzahl der Saat |

Was fehlt, ist ein Fehler (nicht eine Drosselung), steht im Befund des Falls —
„Wirkung fehlt: 25 von 25 geprüften Rufen" — und lässt den Lauf mit Exitcode 1
enden. **Es ist eine Stichprobe**: gezählt werden die geprüften Rufe, nicht
eine Hochrechnung; ein Dienst, der jeden hundertsten Ruf verschluckt, fällt
womöglich durch das Netz.

**Die Prozesse des Laufs stehen in einem Register** (`lastlaufdienste.mjs`)
und werden beendet, wo der Lauf endet: in `finally`, bei einem Signal und
beim Verlassen. Scheitert der zweite Dienst beim Start, wird der erste
beendet; ein Ruf über stdio, dessen Prozess stirbt oder dessen Antwort
ausbleibt (Frist 120 s), wird abgewiesen, statt ewig zu warten. Gegen
`SIGKILL` am Läufer selbst hilft das nicht.

### Speicher

**Der höchste belegte Arbeitsspeicher (RSS) beider Dienste**, alle 250 ms über
`ps` genommen, dazu der Heap aus `/metriken` am Ende des Falls. Die
Werkzeugserver über stdio sind eigene Prozesse und stehen nicht darin.

## Was gemessen ist — am 2026-10-02

**Commit `e1184f1c`**, Apple M5 mit 10 Kernen und 32 GiB, macOS 27.0, Go
1.27.1, Node 24.15.0; je Fall 8 s und 8 Klienten. **Die Maschine war nicht
still**: die Last lag vorher bei 23 und nachher bei 41 (Mittel über eine
Minute), weil daneben fremde Testläufe liefen. Die Zahlen sind deshalb eine
**Untergrenze** dessen, was der Dienst allein schafft; **kein Ruf scheiterte**
— die Spalte *Fehler* steht in allen 43 Gruppen auf 0 %. Alle Zahlen und die
übrigen Spalten im [Beleg](belege/lastlauf.md).

### Ohne Kontingente — was die Maschine kann

| Profil | gelungen je s | p50 / p95 / p99 (ms) | Dienst p95 / p99 | RSS höchstens |
| --- | ---: | --- | --- | ---: |
| REST lesen | 2.462 | 2,5 / 6,3 / 19 | ≤ 5 / ≤ 25 ms | 54 MiB |
| REST schreiben | 284 | 27 / 44 / 52 | ≤ 100 / ≤ 100 ms | 56 MiB |
| REST rendern | 7,2 | 968 / 1.606 / 1.630 | ≤ 2.500 / ≤ 2.500 ms | 2.756 MiB |
| REST Einfuhr (1 KB MTG JSON) | 995 | 7 / 15 / 26 | ≤ 25 / ≤ 25 ms | 56 MiB |
| REST Aufträge (Bogen, ganzer Auftrag) | 1,2 | 3.405 / 3.588 / 3.588 | ≤ 5 / ≤ 5 ms | 488 MiB |
| MCP über stdio, 4 Prozesse | 7.031 | 0,3 / 0,9 / 3,3 | ≤ 1 / ≤ 1 ms | 41 MiB |
| MCP über HTTP | 1.662 | 4,1 / 8,1 / 20 | ≤ 5 / ≤ 25 ms | 55 MiB |
| Überlast, 32 Klienten | 2.571 | 10 / 25 / 59 | ≤ 25 / ≤ 25 ms | 54 MiB |
| REST lesen, Protokoll ausführlich | 2.425 | 2,6 / 6,2 / 21 | ≤ 5 / ≤ 25 ms | 54 MiB |

- **Die Plätze stellen schwere Arbeit in eine Schlange — für das Rendern
  erst seit #2020.** Als diese Zeilen gemessen wurden, belegte **keine
  Rasterung** einen Platz (unten, *Was #2020 geändert hat*): die acht Klienten,
  die rendern, liefen zugleich, der Median von knapp einer Sekunde bei 7
  Rasterungen je Sekunde war das Teilen der Kerne und keine Schlange, und die
  frühere Deutung „bekommen zwei Plätze" traf nicht zu. **Diese Zeile ist
  nicht wiederholt.** Bei den Aufträgen ist es das Auftragsbuch — vier
  laufen, der fünfte bekommt `429` mit `Retry-After: 5` (38 % der
  Versuche).
- **Gemischt teilen sich die Flächen die Maschine**: REST liest dann 267 je
  s mit p95 50 ms, MCP über HTTP 853 je s mit p95 3,6 ms — der Dienst rendert
  und schreibt daneben, der Werkzeugserver nicht.
- **Erholung ohne Schutz ist sofort**: nach der Überlast antwortet der
  höfliche Klient wieder mit p95 1,2 ms (vorher 2,1 ms), der erste Ruf kam
  1 ms nach dem Ende.
- **Das ausführliche Protokoll kostet hier nichts Messbares** (2.425 gegen
  2.462 je s, p95 gleich) — die Last, die die Nachschärfung zum Issue fand,
  ist seit N5 behoben.

### Mit Kontingenten — was der Schutz daraus macht

**Gemessen mit den Kosten von damals**: eine Karte zeichnen kostete an
beiden Flächen 2, ein Abbruch 2. Seit
#2003
kostet das Zeichnen 5 und ein Abbruch nichts (unten, *Was der Lastlauf
gefunden hat*); die Zeilen *REST rendern* und *Erholung* gelten für die
alten Kosten, bis der nächste Lauf sie neu misst.

| Profil | gelungen je s | gedrosselt | Absagen je s | p95 aller Rufe | Retry-After |
| --- | ---: | ---: | ---: | ---: | --- |
| REST lesen | 16,9 | 99,7 % | 6.005 | 2,7 ms | 1 s |
| REST schreiben | 8,3 | 99,7 % | 2.784 | 5,1 ms | 1 s |
| REST rendern | 8,4 | 94 % | 132 | 734 ms | 1 s |
| REST Einfuhr | 2 | 99,9 % | 2.836 | 4,9 ms | 1–4 s |
| MCP über stdio, 4 Prozesse | 59,4 | 99,4 % | 10.105 | 0,6 ms | „in 1 s wieder" |
| MCP über HTTP | 16,9 | 99,2 % | 2.169 | 6,8 ms | „in 1 s wieder" |
| Fairness: A flutet | 16,6 | 99,6 % | 4.518 | 2,5 ms | 1 s |
| Fairness: B fragt alle 250 ms | 3,9 | **0 %** | 0 | 4,9 ms | — |

- **Ein Principal bekommt, was ihm zusteht, und nicht mehr**: 120 Einheiten
  am Stück und 2 je Sekunde, also in 8 s rund 136 — 135 Lesen, 67 Schreiben,
  16 Einfuhren. **Über MCP ist jeder Prozess ein Principal**: vier über stdio
  bekommen viermal so viel, alle Klienten über HTTP ohne Token teilen sich
  einen.
- **Eine Absage ist billig**: ein ungeduldiger Klient bekommt bis zu 6.000
  davon je Sekunde, mit p95 zwischen 2,5 und 7 ms und ohne dass der Speicher
  des Dienstes wächst (53–56 MiB wie ohne Last).
- **Fair ist es**: B merkt von A nichts — kein Ruf gedrosselt, p95 4,9 ms.
- **Die Erholung ist die angesagte**: nach der Überlast gelang der erste Ruf
  des höflichen Klienten **1.009 ms** nach dem Ende, also nach dem
  `Retry-After` von einer Sekunde; danach bekommt er 1,7 je s — der Zulauf.
- **Ein Abbruch kostete Guthaben**: mit leerem Guthaben bekam `DELETE
  /auftraege/{id}` selbst ein `429`, und der Auftrag lief zu Ende (Befund
  unten, **behoben mit #2003**). Ohne Kontingente stand er 24 ms nach dem
  `DELETE` auf `abgebrochen`.

### Langsame und ruhende Verbindungen

| Profil | REST p50 / p95 | MCP über HTTP p50 / p95 |
| --- | --- | --- |
| 20 Verbindungen mit halbem Kopf | 0,9 ms / **9,8 s** | 1,1 ms / 3,3 ms |
| 16 ruhende Verbindungen eines Pools | 0,8 ms / 1,0 ms | — |
| dieselben **vor** dem Befund unten (erster Lauf) | **15 s, Frist abgelaufen** | — |

### Speicher

**Nachgemessen am 2026-10-15 im Fußabdruck** (PERF01, `LEISTUNGSSCHNITTE.md`
§10, `scripts/dienstspeicher.mjs`): unter 10.746 Rasterungen in 30 Sekunden
belegt `mse-rest` höchstens 44 MiB und nach zwei Minuten Ruhe 31 MiB. **Die
Zahlen darunter sind RSS**, das unter macOS zurückgegebene Seiten mitzählt.

**Lesen, Schreiben, Einfuhr und beide Werkzeugserver bleiben bei 41–56 MiB
RSS**, der Heap nach jedem Fall bei 4–8 MiB — auch unter Überlast und mit
Hunderttausenden Absagen. **Aufträge mit Bögen** belegen bis 488 MiB, **der
gemischte Fall** bis 590 MiB. **Rendern** ist die Ausnahme: bis **3,7 GiB**
RSS mit Kontingenten und 2,7 GiB ohne, bei einem Heap von 110–286 MiB
danach (Befund unten).

## Die Grenzen nach AK-02 — wo sie sitzen, wo sie zu sehen sind, was sie hält

**Nachgelesen am Stand `52a58825`**, um den Schalter aus AK-03 ergänzt und
**mit #2003 fortgeschrieben**: Kostentafel, Auftrags- und Verbindungsgrenze in
jeder Auskunft, die Verbindungsgrenze auch für `mse-mcp`, ein Abbruch ohne
Kosten. Die letzte Spalte nennt den Lauf, der die Zusage hält — einen Test
oder ein Profil des Lastlaufs.

| Frage | Grenze | Sitz | Sichtbar | Gehalten von |
| --- | --- | --- | --- | --- |
| **je Identität** | Guthaben je Principal — je Maschine eines Tokens, die Prozessmarke als eine, je MCP-Prozess eine: 120 Einheiten, 120 je Minute; abschaltbar | `service.Kontingente` | `principal` und `kontingente` in `GET /grenzen`, `lesen.grenzen`, `mse grenzen`, Hilfe | `TestEinPrincipalVerbrauchtSeinBudgetUndDerAndereArbeitetWeiter`, `TestDasGuthabenZaehltNachDerMaschine`; Profil `fairness` |
| **teure Operation** | Kosten je Ruf: Lesen 1, Schreiben 2, Bild 3, Rendern 5, Ausfuhr 8, Einfuhr 8, Masse 10, Abbruch 0 — **welche Fähigkeit teuer ist, sagt eine Tafel** mit ihren Wegen an REST, MCP und Kommandozeile; was nicht darin steht, kostet nach seiner Art | `domain.Kostentafel`, gefragt über `domain.KostenklasseFuer` vom Dienst (mit dem Muster des Wegs) und vom Werkzeugserver | `kosten` und `kostentafel` | `TestDieTeureOperationKostetUeberJedeFlaecheDasselbe` am gebauten Programm; `TestJederWegDerTafelBuchtSeineKlasse`, `TestJedesWerkzeugHatSeineKostenklasse`; `TestDieKostentafelNenntNurWegeDesKatalogs` hält jeden Weg gegen `FAEHIGKEITEN.md`; Profile `rest-einfuhr`, `rest-schreiben` |
| **Parallelauftrag** | Aufträge des Dienstes: vier laufend, sechzehn behalten, eine Stunde | `domain.Vorgabegrenzen` (`Auftraege`), gelesen von `internal/rest/auftraege.go` | `auftraege` in `GET /grenzen`, `lesen.grenzen`, `mse grenzen` — und im Satz der Ablehnung | `TestZuVieleAuftraegeWerdenAbgewiesen`, `TestDieGrenzenSagenAnJederFlaecheDasselbe`; Profil `rest-auftraege` |
| **Ressource** | Plätze je Klasse: Rendern 2, Einfuhr 1, Ausfuhr 1, Bild 4, Masse 1, Suche 4, Indexbau 1 — je Prozess; **Rendern belegt seit #2020 jede Rasterung** — Kartenbild, Bogen, SVG-Teile, Vorschau, Tisch — an einer Pforte des Rasterers (unten, *Was #2020 geändert hat*) | `service.Haushalt`; für das Rendern `render.Rasterer.MitPforte`, `service.Rasterpforte`, gehängt in `app.NeuesZeichenwerk` | `parallel` | `TestDerHaushaltVergibtPlaetzeJeKlasse`, `TestJedeSchwereArbeitWartetAufIhrenPlatz`; `TestJederZeichenwegRastertMitBelegtemRenderplatz`, `TestMitEinemPlatzZeichnetNurEinKartenbildZugleich`; Profile `rest-rendern`, `rest-einfuhr` |
| | Rumpf und Antwort 32 MiB; eine MCP-Nachricht 8 MiB | `internal/rest/dienst.go`, `internal/mcp/ueberhttp.go` | `antwortBytes` | `TestEineZuGrosseAntwortBrichtAb`, `TestEinZuGrosserRumpfWirdAbgewiesen` |
| | sechzehn Verbindungen — `mse-rest` **und `mse-mcp` über HTTP**, mit demselben Lauscher | `domain.Vorgabegrenzen` (`Verbindungen`), `internal/rest/gedraenge.go`; gesetzt in `cmd/mse-rest/bedienen.go` und `cmd/mse-mcp/ueberhttp.go` | `verbindungen` in jeder Auskunft, `NETZEIGENSCHAFTEN.md` | `TestMehrAlsDieGrenzeWirdNichtAngenommen`, `TestUeberHTTPNimmtNichtMehrVerbindungenAlsDieGrenze`, `TestDerWerkzeugserverUeberHTTPHaeltDieVerbindungsgrenze`; Profile `langsam`, `ruhend` |
| **Fairness** | ein Guthaben je Principal **und** eines für alle; die Tabelle hält 256 Principals und verdrängt den ältesten | `service.Kontingente` | `global`, `reserve` | `TestDieTabelleDerPrincipalsBleibtBegrenzt`; Profil `fairness` |
| **Warteschlange** | wer keinen Platz bekommt, wartet — bis einer frei wird oder sein Kontext endet; eine Verbindung über sechzehn wartet im Rückstau des Betriebssystems | `Haushalt.Belegen`, `gedraenge.Accept` | — | `TestDerHaushaltWartetBisEinPlatzFreiWirdOderDerRufEndet`, `TestEineZweiteRasterungWartetBisDieErsteFreigibt`, `TestSindAlleRenderplaetzeBelegtWartetDieRasterungBisEinerFreiWird`; Profile `rest-rendern`, `langsam` |
| **Ablehnung** | REST `429` mit Code `drossel` oder `auftraege`; MCP ein Werkzeugfehler „zu viele Rufe" mit `isError` und, seit dem Zweitblick über Phase 1 von 0.9.0, `structuredContent.fehler` mit Code `drossel` | `rest.zuViele`, `mcp.Server` | im Rumpf und im Satz | `TestAuchMitMarkeGiltEineGrenze`, `TestEinRufUeberDemGuthabenBekommtDieWartezeit`; jedes Profil mit Kontingenten |
| **Retry-Hinweis** | `Retry-After` als echte Restzeit, aufgerundet, nie 0; bei Aufträgen fest 5 s; MCP „in N s wieder" | `rest.zuViele`, `rest.angenommen`, `mcp.Server` | Kopfzeile und Satz | `TestDieWartezeitLaeuftMitDerUhr`; Spalte *Retry-After* im Beleg |
| **Abbruch** | ein endender Kontext gibt den Platz frei; `DELETE /auftraege/{id}` und `DELETE /index/aufbau`; über MCP `notifications/cancelled`, über HTTP das Schließen des Stroms. **Ein Abbruch kostet nichts und wird nie gedrosselt** — auch nicht unter der Reserve | `Haushalt.Belegen`, `internal/rest/auftraege.go`, `internal/mcp/abbruch.go`; `KostenAbbruch` in `domain.Kostentafel`, `Kontingente.Abbuchen` | `abbruch` in `kosten` | `TestEinAuftragLaesstSichAbbrechenUndWiederholen`, `TestEinAbgebrochenerRufBekommtKeineAntwort`, `TestEinAbbruchWirdNieGedrosselt`, `TestWasNichtsKostetGehtImmerDurch`, `TestEinAbbruchWirdAmDienstNieGedrosselt`; Profil `rest-auftraege` |
| **Raten** | Sperre je vorgezeigtem Nachweis (drei frei, dann 1 s verdoppelt bis 5 min), dahinter 60 Fehlversuche je Minute für alle | `service.Zugangssperre`, `rest.Drossel` | `ADMINISTRATION.md` | `TestDieSperreTrifftNurDenRatendenNachweis`, `TestEinFremderKannDenBerechtigtenNichtAussperren`, `TestDieKontingenteLassenSichAbschalten` (bleibt auch ohne Kontingente) |

## Was der Lastlauf gefunden hat

**Zwei Fehler, beide behoben und mit benanntem Lauf** — und beide hätte
keine Pipeline gefunden, denn kein Test hielt einen Pool mit sechzehn
Verbindungen offen.

1. **Sechzehn ruhende Verbindungen sperrten jeden anderen aus.** Ein Klient
   mit einem Pool fragte je einmal und ließ die Verbindungen offen, wie jeder
   Pool mit `keepAlive`. Danach nahm `mse-rest` keine Verbindung mehr an,
   bis der Leerlauf von zwei Minuten sie schloss — die Lesefrist griff nicht,
   denn ein Kopf kam ja. Gemessen im ersten Lauf
   ([`belege/lastlauf-erstlauf.md`](belege/lastlauf-erstlauf.md), Profil
   `ruhend`): der zweite Klient gab nach 15 s auf. **Jetzt** geht eine
   ruhende Verbindung, sobald eine neue wirklich auf einen Platz wartet
   (`internal/rest/gedraenge.go`); gehalten von
   `TestRuhendeVerbindungenSperrenNiemandenAus` und am gebauten Dienst von
   `TestRuhendeVerbindungenSperrenAmDienstNiemandenAus`.
2. **Die erste Korrektur riss Verbindungen unter Volllast ab.** Sie schloss
   bei vollem Haus jede Verbindung, die auch nur zwischen zwei Rufen ruhte.
   Gemessen im zweiten Lauf
   ([`belege/lastlauf-zwischenlauf.md`](belege/lastlauf-zwischenlauf.md),
   Profil `ueberlast`): **24 bis 39 % der Rufe abgerissen**. Jetzt wird erst
   angenommen und dann nach dem Platz gefragt; gehalten von
   `TestUnterVollerLastBrichtKeineVerbindungAb` (vorher 164 von 600 Rufen
   abgerissen).

**Und Befunde, die beim Lauf offen blieben.** Vier davon sind seit
#2003
behoben, zwei bleiben — keiner ist ein Absturz, jeder war eine Zusage, die
nicht ganz stimmte.

- **Was ein Ruf kostete, hing an der Fläche — behoben.** REST buchte nach
  einer kurzen Liste von Wegen (`rest.KostenklasseFuer`), MCP nach Wörtern im
  Namen eines Werkzeugs (`mcpnaehte.kostenklasseFuer`). Eine Karte zeichnen
  kostete an beiden 2 statt 5 — der Lastlauf ließ mit Kontingenten 67
  Rasterungen durch, wo 5 Einheiten je Ruf rund 27 erlaubten —, den Bogen,
  MTG JSON schreiben, ein Paket und eine Sammlung einlesen buchte REST mit 2
  und MCP mit 8, und eine Vorschau der Einstellungen kostete über MCP 5,
  weil ihr Name „vorschau" trägt. **Jetzt steht je teurer Fähigkeit eine
  Zeile in `domain.Kostentafel`**, mit ihren Wegen an jeder Fläche, und beide
  Flächen fragen sie; `mse grenzen` nennt dieselbe Tafel. Gehalten am
  gebauten Programm von `TestDieTeureOperationKostetUeberJedeFlaecheDasselbe`:
  zeichnen ohne Karte, bis das Guthaben leer ist — über REST und MCP gehen
  je 24 Rufe durch, 120 durch 5, und `mse grenzen` sagt 5.
- **Zwei Grenzen standen in keiner Auskunft — behoben.** `grenzen` nennt an
  jeder Fläche `verbindungen` und `auftraege` (laufend, behalten,
  Aufbewahrung), und die Zahlen stehen bei den übrigen in
  `domain.Vorgabegrenzen` statt als Konstanten im Dienst.
  `TestDieGrenzenSagenAnJederFlaecheDasselbe` vergleicht die Auskunft von
  Kommandozeile, MCP und REST Feld für Feld.
- **`mse-mcp` über HTTP hatte keine Verbindungsgrenze — behoben.** Er nimmt
  jetzt durch denselben Lauscher wie `mse-rest` höchstens sechzehn an; eine
  darüber wartet, und eine ruhende gibt ihren Platz her, sobald jemand
  wartet.
- **Ein Abbruch kostete Guthaben — behoben.** `DELETE /auftraege/{id}` buchte
  wie ein Schreiben; wer sein Guthaben aufgebraucht hatte, konnte seinen
  eigenen Auftrag erst nach `Retry-After` abbrechen, und der Auftrag lief zu
  Ende. Ein Abbruch ist die eine Anfrage, die Last **nimmt**: er kostet jetzt
  nichts, der Dienst fragt das Guthaben dafür gar nicht erst, und das
  Guthaben lässt, was nichts kostet, immer durch — auch unter der Reserve.
- **Zwanzig halbe Köpfe halten `mse-rest` zehn Sekunden auf** (Profil
  `langsam`, p95 9,8 s): sechzehn belegen die Plätze, bis die Lesefrist sie
  schließt. Das ist der Preis der Verbindungsgrenze, und er ist bewusst
  ([`NETZEIGENSCHAFTEN.md`](NETZEIGENSCHAFTEN.md)); eine kürzere Frist
  träfe langsame Leitungen. **Seit #2003 gilt er auch für `mse-mcp` über
  HTTP** — derselbe Lauscher, derselbe Preis.
- **Rendern belegt viel Speicher**: bis 3,7 GiB RSS bei zwei Plätzen,
  während der Heap danach bei 110–286 MiB liegt — Go gibt Speicher verzögert
  ans System zurück, und eine Rasterung teilt große Bilder zu. Die Plätze
  begrenzen, wie viele zugleich rastern, nicht wie viel eine belegt; ob ein
  einzelner Ruf weniger bräuchte, ist hier nicht untersucht. **Die 3,7 GiB
  sind außerdem die von acht Rasterungen zugleich**: gemessen wurde, bevor
  eine Rasterung einen Platz belegte (#2020), und was zwei statt acht
  ändern, steht noch aus. **Gemessen ist
  seither**, dass jeder Ruf sein Zeichenwerk neu baute und dabei jede
  Schriftdatei des Packs las; seit #2005 behalten Dienst und Werkzeugserver
  es ([`LEISTUNGSVERGLEICH.md`](LEISTUNGSVERGLEICH.md) §8). Eine
  `mse-mcp`-Sitzung stand danach bei 482 MB statt 502–553 MB — **dieser
  Lastlauf ist nicht wiederholt**, und für zwei Plätze unter Last gilt die
  Zahl darüber, bis er es ist.

**Was #2003 offen ließ, und was davon #2020 geändert hat** — nachgelesen am
2026-10-02, fortgeschrieben am 2026-10-03 ([0163](entscheidungen/0163-zeichnen-belegt-einen-platz-und-lesen-kostet-lesen.md)):

- **Ein REST-Weg, der nur liest, aber `POST` heißt, kostete 2 statt 1 —
  behoben.** Die Art eines Wegs las der Dienst an seiner Methode, MCP am
  Werkzeug: `POST /einstellungen/vorschau` kostete 2,
  `lesen.einstellungen.vorschau` 1 — einundzwanzig solcher Paare stehen in
  `FAEHIGKEITEN.md`, alle Vorschauen und Deutungen. **Jetzt stehen sie mit
  der Klasse Lesen in `domain.Kostentafel`** (`kostenlesen.go`, neunzehn
  Zeilen), und beide Flächen fragen dieselbe Tafel; die Methode entscheidet nur
  noch dort, wo die Tafel schweigt. **Die Methodenregel für Token
  (`schreibend` in `internal/rest/tokenwache.go`) bekam keine zweite Art**: ein
  Token, das nur lesen darf, kommt an diese Wege weiter nicht heran — über MCP
  kann ein lesendes Profil sie rufen, über REST nicht, und das ist eine
  Abweichung in der Erlaubnis, die offen bleibt. Gehalten von
  `TestDieKostenFolgenDerFaehigkeitAnRESTUndMCP`: jedes Paar der Tabelle
  gegen beide Flächen, rot vorher bei genau diesen einundzwanzig.
- **Kein Dienst belegte einen Platz der Klasse `rendern` — behoben.**
  `Vorgabegrenzen` nennt zwei, und `GET /grenzen` sagt es; ein `Belegen` oder
  `Versuchen` mit `RessourceRendern` stand außerhalb der Tests nirgends
  (gesucht über `internal/` und `cmd/` am 2026-10-02). Die Deutung der
  Messung — acht Klienten, die rendern, „bekommen zwei Plätze" — trug
  deshalb nicht. **Jetzt fragt jede Rasterung eine Pforte** (`render.Pforte`),
  und `service.Rasterpforte` belegt dahinter einen von zwei Plätzen; gehängt
  ist sie einmal, in `app.NeuesZeichenwerk`, durch das das Fenster und jede
  Fläche ohne Fenster gehen. Es sind Kartenbild, Druckbogen, die Rasterteile
  einer SVG, die Vorschau, der Spieltisch und die Kacheln der Symbolpalette —
  jeder Aufruf von `Rasterer.Zeichne`. **Nicht durch die Pforte** geht das
  Rastern eines einzelnen Symbols oder Bildbaums. Ein Druckbogen hält den
  Platz der Klasse `ausfuhr` und belegt je Karte einen `rendern`: erst
  `ausfuhr`, dann `rendern`, nie umgekehrt — es gibt keinen Kreis. Gehalten von
  `TestJederZeichenwegRastertMitBelegtemRenderplatz` (mitten in der Rasterung
  ist genau ein Platz der Maschine belegt, für beide Wurzeln und jeden Weg),
  `TestMitEinemPlatzZeichnetNurEinKartenbildZugleich` (zwei Rufe bei einem
  Platz: der zweite wartet, bis der erste freigibt) und
  `TestDasFensterRastertMitBelegtemRenderplatz`.
  **Der Lastlauf ist nicht wiederholt**: die Zeilen *REST rendern* und der
  Speicher darunter sind die von damals, und was zwei Plätze statt acht
  gleichzeitiger Rasterungen an Durchsatz, Median und Arbeitsspeicher ändern,
  steht aus — dazu gehört der Tisch des Fensters, der sich die zwei Plätze mit
  der Vorschau teilt.
- **Wer teurer sein könnte, ist nicht neu verhandelt.** Ein Set öffnen
  (`POST /sets`) belegt denselben Einfuhrplatz wie MTG JSON lesen, ein Paket
  beziehen, installieren, übernehmen oder weitergeben und eine Vorlage
  ausführen arbeiten wie eine Ein- oder Ausfuhr — und kosten an jeder Fläche
  gleich, nämlich 2. Das ist eine Frage über Preise und nicht über Parität.
- **Eine Wiederholung kostet, was ein Schreiben kostet**:
  `POST /auftraege/{id}/wiederholung` fährt einen Bogen ein zweites Mal und
  bucht 2 statt 8 — gebucht wird vor dem Weg, und welcher Auftrag wiederholt
  wird, weiß erst der Weg.

## Was die Messung nicht sagt

- **Eine zweite Maschine.** Klient und Dienst teilen sich diesen Rechner; ein
  schneller Klient nimmt dem Dienst Kerne weg, und alles andere, was gerade
  läuft, auch. Der Beleg nennt deshalb die Last der Maschine vorher und
  nachher ([0007](entscheidungen/0007-kein-entfernter-lauf.md)).
- **TLS.** Alles läuft auf der Schleife, wo TLS nicht verlangt ist; jenseits
  davon käme der Handschlag dazu.
- **Große Bestände.** Die Saat ist das Probeset mit wenigen Karten; ein Set
  mit tausend Karten liest und schreibt anders. Die Einfuhr liest eine Datei
  von einem Kilobyte.
- **Das Fenster.** Es hat kein Guthaben und keine Uhr am Eingang
  (`internal/service/laufzeitwerk.go`); dass es neben einer Überlast
  bedienbar bleibt, ist hier nicht gemessen.
- **Dauerlast.** Ein Fall dauert Sekunden; ob Speicher über Stunden wächst,
  sagt er nicht.

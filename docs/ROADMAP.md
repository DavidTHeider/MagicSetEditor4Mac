# Roadmap

**Was offen ist — und der Messstand, an dem sich zeigt, wie weit es ist.**
Was erledigt wurde, steht als Arbeit in [`CHRONIK.md`](CHRONIK.md) — bis 0.9.0
im Entwicklungsarchiv, ab 1.0.0 dort.

Die Trennung ist keine Formsache: offene Arbeit zwischen erledigter zu suchen,
lässt Erledigtes als offen führen.

**Der Messstand gehört trotzdem hierher**, und das ist eine Entscheidung
(#615). Hier
stand nur „was offen ist", und darunter standen die Phasenberichte mit ihren
Zahlen — die erste Zeile stimmte damit nicht mit dem Inhalt überein. Sie zu
befolgen und die Berichte in die Chronik zu räumen hätte den **Verlauf**
zerschnitten: 26,714 → 22,426 → 20,679 ist eine Kurve, und eine Kurve, deren
Punkte auf zwanzig Einträge verteilt sind, liest niemand. Die Zeile ist deshalb
genauer geworden, nicht der Inhalt kleiner.

**Repository:** [DavidTHeider/MagicSetEditor4Mac](https://github.com/DavidTHeider/MagicSetEditor4Mac) ·
**Issues:** [alle offenen](https://github.com/DavidTHeider/MagicSetEditor4Mac/issues). Ein Verweis `#N`
aus der Zeit bis 1.0.0 nennt ein Issue oder einen PR des privaten
Entwicklungsarchivs ([0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)).

---

## Stand

| | |
| --- | --- |
| **Freigegeben** | **1.0.0** am 2026-11-15 — [v1.0.0](https://github.com/DavidTHeider/MagicSetEditor4Mac/releases/tag/v1.0.0), die erste Freigabe für Anwendende: der Programmstand von 0.9.0, neu gebaut mit zwei Sicherheitsaktualisierungen und einem eigenen Freigabeschlüssel ([0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)). **Mit derselben benannten Grenze wie 0.9.0**: das Tor des Reifeprofils hält nicht ganz — D1, D4 und D6 an der Sache, D10, D12 und D13 an Mensch oder Hardware ([`REIFEPROFIL.md`](REIFEPROFIL.md)). **Keine** der sieben Handprüfungen ist am Abbild 1.0.0 eingetragen — sie entstehen nach dem Bau (Tabelle unten, [HANDPRUEFUNGEN.md](HANDPRUEFUNGEN.md)) |
| **Als Nächstes** | **1.1.0** — keine Phase geschnitten; ihr Zuschnitt entsteht mit ihrem Auftrag |
| **Der Abstand zu MSE2** | **9,319** im Kanonmittel der festen Kohorte (4.266 Karten), **9,293** in der Zufallsstichprobe (4.582 Karten) mit 95-%-Band [9,186 · 9,399] — gemessen zur Freigabe 0.9.0 am 2026-10-21 an der Rasterung, die 1.0.0 unverändert trägt; Methode in [`MESSMETHODIK.md`](MESSMETHODIK.md) |
| **Zuletzt aktualisiert** | 2026-11-15, mit der Freigabe 1.0.0 |
| **Lizenz** | seit 2026-08-16 **GPL-2.0-or-later** statt MIT ([0019](entscheidungen/0019-gpl-2-0-statt-mit.md)). Damit darf aus MSE2 portiert werden — die Semantik von MSE-Script ist nicht mehr aus Beobachtung zu erschließen |
| **Leitlinien** | [`AGENTS.md`](../AGENTS.md) — verbindlich, gleichlautend mit `CLAUDE.md` |

### Offene Phasen

**Keine.** Geschnitten wird die nächste Iteration erst mit ihrem Auftrag.

---

## Nach 1.0.0 — was offen bleibt

**1.0.0 ist am 2026-11-15 freigegeben.** Kein Paket ist offen; was bleibt,
braucht einen Menschen, zweite Hardware oder einen Lauf am fertigen Abbild und
steht in der Tabelle unten und in [HANDPRUEFUNGEN.md](HANDPRUEFUNGEN.md).

**Offen ist noch:**

| Was | Warum es liegen bleibt |
| --- | --- |
| **Die Peerverbindung zwischen zwei echten Rechnern** (N1, #1454) | **Zwischen zwei VM-Gästen gelaufen** — über beide Transportadapter: mDNS und TLS im eigenen Netz, der Fernweg über tailcat gewöhnlich und mit erzwungenem DERP, auch zwischen tailcat v0.6.0 und v0.7.0 (`belege/sec22-vmlauf.md`, K01 #2187); die gebaute Anwendung hat im Gast gefunden und ist gefunden worden (`belege/k02-lokalesnetz.md`). **Offen bleiben zwei physische Rechner** — dasselbe, was [0007](entscheidungen/0007-kein-entfernter-lauf.md) für jeden Lauf hier sagt: eine Maschine, keine Plattform |
| **Direkte Route und erzwungener DERP-Rückfall zwischen zwei Macs hinter getrennten Routern** (N3, #1456) | Das Issue sagt es selbst: „Eine lokale Demo ist ausdrücklich kein Internet-Abnahmebeleg." **Zwischen zwei Gästen im selben Netz ist beides gelaufen** — direkt über die LAN-Adresse, mit `TS_DEBUG_ALWAYS_USE_DERP=1` nur über das Relais, und seit tailcat v0.7.0 nennt der Fernweg den Weg (K01). **Was fehlt, sind zwei Netze**: zwei Rechner hinter verschiedenen Routern, also eine echte NAT-Durchquerung |
| **Handprüfung 1 — der Lauf über fremde Pakete am Stand 1.0.0** | Sein Beleg nennt einen Commit, und der Stand von 0.9.0 ist im neuen Repositorium keiner; er wird am Wurzelstand neu gefahren (`make fremdpakete`) |
| **Handprüfung 2 — der Kanonlauf am Stand 1.0.0** | Dieselbe Bindung an den Commit; die Rasterung ist die von 0.9.0, die Zahl im Stand oben ist bis dahin die der Freigabe 0.9.0 |
| **Handprüfung 3 — das ausgelieferte Abbild hell** | Sie gilt für **ein** Abbild, und 1.0.0 ist ein neues ([`RELEASE.md`](RELEASE.md)) |
| **Handprüfung 4 — das ausgelieferte Abbild dunkel** | Dieselbe Bindung an das Abbild |
| **Handprüfung 7 — der Bestand nach dem Start** | Am neuen Abbild mit einem Bestand von 0.9.0 nachzustellen |
| **Handprüfung 6 — der Erststart aus dem `.dmg` samt Gatekeeper-Weg** | Er öffnet die Anwendung mit dem echten Bestand und den Vorgaben dieses Rechners, und der Weg über die Systemeinstellungen braucht eine Hand — offen in jedem Freigabebeleg seit 0.6.0 |
| **Handprüfung 5 — ein Druckbogen, wirklich gedruckt und mit dem Lineal nachgemessen**                                                                                                                       | Es geht nicht ohne Hand. `BogenLegen` öffnet im Fenster **immer** einen Speichern-Dialog — eine der Fähigkeiten am nativen Dateidialog, und über Kommandozeile, MCP und Dienst gibt es sie nicht. `System Events` darf Prozesse auflisten, aber nicht bedienen. Und der Kern der Prüfung — ob eine gedruckte Karte 63 mm misst — braucht Papier und ein Lineal                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                   |

**Zurückgestellt, je mit Vermerk** — keine offene Arbeit, aber Fragen, die
wiederkommen: **Quick Look für Setdateien** (machbar und nicht gebaut,
[0186](entscheidungen/0186-quick-look-ist-machbar-und-wird-nicht-gebaut.md))
und **die Bindung im Schlüsselbund** (kein eigener Herausgeber,
[0187](entscheidungen/0187-kein-eigener-herausgeber.md)). Beide nennen, unter
welcher Bedingung sie wieder aufgenommen werden.

---

## Der Messstand im Verlauf

**Der Abstand zur Referenz, mit der Menge, über die er erhoben wurde** (§10 —
die Reichweite einer Messung ist die Reichweite ihrer Aussage). Die
Einzelheiten standen in den Phasenberichten bis 0.9.0 und liegen mit ihnen im
Entwicklungsarchiv.

| Datum      |    Abstand | Menge                      | Was ihn bewegt hat                                                                                                           |
| ---------- | ---------: | -------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| 2026-08-17 |     26,714 | Stichprobe, Phase J        | der Stand vor der Diagnose (#541)                           |
| 2026-08-17 |     23,043 | dieselbe                   | #528 — die leere Schlusszeile zählte zur Blockhöhe          |
| 2026-08-17 |     22,646 | dieselbe                   | #532 — der Skalierfilter war vertauscht                     |
| 2026-08-19 |     22,542 | dieselbe                   | #531 — eine `boolean`-Stiloption stand auf „nein"           |
| 2026-08-19 |     22,478 | dieselbe                   | #547 — eine Karte bringt eigene Stiloptionen mit            |
| 2026-08-19 |     22,426 | dieselbe                   | #556 — eine Mehrfachwahl stand in der Reihenfolge der Datei |
| 2026-08-20 |     21,553 | **573 Karten, vier Sätze** | #546 — Zeilenhöhe und Grundlinie rasten auf Viertelpunkte   |
| 2026-08-20 |     20,692 | dieselbe                   | dito, zweiter Teil                                                                                                           |
| 2026-08-20 | **20,679** | dieselbe                   | #597 — ohne `hhea` gilt die `OS/2`-typo-Metrik              |
| 2026-08-28 | **20,226** | **25.141 Karten, 10 Sets** | der Freigabelauf zu 0.3.0 — der **ganze** Kanon, nicht die Stichprobe                                                        |
| 2026-09-07 | **20,056** | feste Kohorte, 4.265 Karten (16,96 % von 25.148) | der Stand zu 0.4.0; zur Freigabe 0.5.0 am 2026-09-20 bis auf die vierte Nachkommastelle wiedergefunden ([`MESSMETHODIK.md`](MESSMETHODIK.md) §5a, `ABNAHME-0.5.0.md` (Entwicklungsarchiv)) |
| 2026-09-23 | **13,792** | feste Kohorte, 4.266 Karten; Zufallsstichprobe 4.582 | QA-03 (#1656): Kerning, MSEs vierfacher Textweg, Ligaturen und die Glättung von Quartz (#1805, #1806); Beleg am Stand `7972c05` ([`MESSMETHODIK.md`](MESSMETHODIK.md) §5b) |
| 2026-10-11 | **11,470** | feste Kohorte, 4.266 Karten | F02 (#2244): Portierungen aus MSE2 — Maske des Textfelds, Symbole in ganzen Punkten, Maßstabssuche, Textschatten, Kursivschnitte, das Maß einer Wahl; Beleg am Stand `1a45973` (`belege/fremdpakete.json`) |
| 2026-10-12 | **10,686** | feste Kohorte, 4.266 Karten; Zufallsstichprobe 10,630 | der Stand `e48f6d9`: J05 (#2269) — ein einzeiliges Feld bricht nur am harten Umbruch, `symbol_variation` nimmt Farben als Text; getrennt gemessen ist der Anteil beider nicht |
| 2026-10-12 | **10,540** | feste Kohorte, 4.266 Karten; Zufallsstichprobe 10,495 | die Folgen der Kontrolle J05 (`9e15231`): ein gestauchtes Feld bricht erst um und wird dann gestaucht (#2308), jede Schriftgröße liegt auf ¼ Punkt; die Namen der Vanguard-Avatare wurden dabei schlechter (#2302) |
| 2026-10-12 | **9,427** | feste Kohorte, 4.266 Karten | eine Zeile wird je Abschnitt in einem Zug gezeichnet, wie `TextViewer::draw` es tut (#2302) — der Text trägt 17 statt 29 Prozent, Vanguard 2,83 statt 6,83; Karte für Karte gegen 0188 sechs schlechter (#2315) |
| 2026-10-14 | **9,328** | feste Kohorte, 4.266 Karten; Zufallsstichprobe 9,299 | `in_context` wie MSE2 und POSIX-Klassen auf der Rückfalllinie (#2315) — die sechs Karten aus #2315 verkleinern nicht anders, sie zeichneten falsche Symbole; Karte für Karte gegen 0188 zwei schlechter, beide begründet (Warlock Class, Cleric Class; #2396) |
| 2026-10-20 | **9,319** | feste Kohorte, 4.266 Karten | gespeichert wird, was gerechnet ist, und das Zeichnen rechnet in der Reihenfolge des Spiels wie MSE2 beim Laden (#2396, [0203](entscheidungen/0203-gespeichert-wird-was-gerechnet-ist.md)) — Karte für Karte gegen `f892cfd6b` (9,328): **zwölf näher an MSE2, keine weiter weg**, zehn Planeswalker, Warlock Class und Cleric Class; die zwei sind gegen 0188 nicht mehr schlechter. Gemessen mit `TestDerAbstandZuMSE2AmKanon`, **kein Beleg** — `fremdpakete.json` und die Zufallsstichprobe folgen mit QA01 |

**Die ersten sechs, die folgenden drei, die Zeile vom 2026-08-28 und die
vom 2026-09-07 sind nicht vergleichbar.** Der Wechsel der Menge steht in der
Spalte daneben, und wer die Zahlen ohne sie nebeneinanderlegt, sieht Sprünge,
die keine sind. **Ab dem 2026-09-07 ist es dieselbe feste Kohorte** (4.265,
seit dem 2026-09-23 4.266 Karten laut `MESSMETHODIK.md`); die Schritte dazwischen
sind in den Zeilen genannt.

**Die Zeile vom 2026-08-28 ist der Freigabebeleg zu 0.3.0**, und sie ist die
erste über den **ganzen** Kanon: 25.141 Karten aus zehn Sets, 5.021.040.837
Bildpunkte, in 2 Stunden 51 Minuten. Zerlegt: Text 55,13 % Anteil bei 4,14 %
der Fläche, Bild 41,42 % bei 93,34 %, Symbolzug 3,17 %, Fläche 0,29 %. Bester
Satz 8,310, schlechtester 73,538, Mitte 19,420.

**Die schlechteste Vorlage ist `m15-mainframe-dfc`** mit 34,586 über 322
Karten, gefolgt von `m15-flip` (33,166) und `m15-aftermath` (28,684) — die
doppelseitigen und die geteilten. `m15-altered` trägt mit 24.049 Karten die
Masse und liegt bei 20,026.

**Und 2.809 Karten gewännen durch einen Versatz von höchstens zwei Punkten**,
im Mittel 7,355 — bei 524 davon liegt der beste Versatz am Rand des
Suchbereichs, dort ist die Zahl also eine Untergrenze. Über das **ganze** Bild
gewinnt dagegen **keine einzige** Karte durch einen Versatz: die Rahmen sitzen,
der Text steht daneben.

**Und der Messstand trägt einen Fehler der Referenz.** An der Symbolfläche
zeigt MSE2 auf macOS die Randfarbe statt der Füllung
(#458); das
kostet **höchstens 0,918** von 20,679 — beziffert am 2026-08-20
([0032](entscheidungen/0032-der-fehler-der-referenz-bleibt-im-mass.md)).

---

## Was schon offen ist, ohne Phase

| Punkt                                               | Grund                                                                                | Was zu tun ist                                                                                                                                                                                                                                                                                                                                                                         |
| --------------------------------------------------- | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Serverseitiger Schutz von `main`                    | setzt bei privaten Repositories GitHub Pro voraus                                    | Bis dahin greift der `pre-push`-Hook. Wird das Repository öffentlich oder das Konto auf Pro umgestellt, gehört ein Ruleset mit erzwungenem PR und grüner Pipeline nachgezogen                                                                                                                                                                                                          |
| **Geprüft ist eine Maschine, nicht eine Plattform** | Es gibt keinen entfernten Lauf ([0007](entscheidungen/0007-kein-entfernter-lauf.md)) | Nichts — das ist eine Entscheidung. Was sie kostet, steht im Vermerk und gehört in jede Release-Notiz: kein Test lief je auf einer anderen Plattform, kein frischer Klon auf einer zweiten Maschine                                                                                                                                                                                    |
| `docs/FORMATE.md` ist überwiegend Vermutung         | erschlossen an **einem** Paket                                                       | **Diese Zeile versprach den Beleg für Phase 2, und Phase 2 hat ihn nicht erbracht.** Der Wächter zählt nach, dass jede Feststellung ihre vier Angaben trägt und jede Vermutung ihren Testfall — nicht, dass sie an einem zweiten Paket gesehen wurde. Belegt wird an den Paketen, die Phase 3 ohnehin liest; der Gegenlauf hinter dem Build-Tag `fremdpakete` ist weiter ungeschrieben |

---

## Die vier Risiken

Aus dem ersten Entwurf (§4, im Entwicklungsarchiv). Sie stehen hier, damit sie niemand später für eine
Überraschung hält.

| Risiko                                | Woran man es merkt                                    | Was dagegen steht                                                                                                       |
| ------------------------------------- | ----------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| **MSE-Script ist größer als gedacht** | Phase 3 wächst über ihren Zuschnitt                   | Die Sprache wird an echten Paketen erschlossen. Was nicht belegt ist, bekommt einen Testfall, der die Vermutung benennt |
| **Rendering trifft MSE2 nicht**       | Goldbilder weichen sichtbar ab                        | Abweichungen werden benannt, nicht weggeräumt. Ziel ist eine brauchbare Karte, nicht Pixelgleichheit mit wxWidgets      |
| **Die Rundlaufzusage bricht**         | ein zurückgeschriebenes Set öffnet in MSE2 nicht mehr | `raw_blocks` ab Phase 2 mitbauen, nicht nachrüsten                                                                      |
| **Ohne Vorlagen ist nichts zu sehen** | die Saat reicht nicht, um etwas zu tun                | Phase 7 ist keine Zugabe                                                                                                |

---

## Was ausdrücklich nicht kommt

Damit niemand darauf wartet:

- **Keine Synchronisierung** des Bestands zwischen Rechnern — die Zusage aus
  [0010](entscheidungen/0010-mcp-als-eigener-prozess.md); gegen den heutigen
  Austausch von Umschlägen, Partien und Handel nicht neu geprüft am
  2026-10-11.

  **Hier stand bis zum 2026-10-11 „Kein Dienst, der von außen erreichbar ist,
  keine Anmeldung".** Beides stimmt nicht mehr. Eine Anmeldung gibt es seit
  Phase 4.12 (Nutzer, Profile und die Anmeldung über fremde Anbieter,
  #1351). Und
  die Dienste sind erreichbar, wenn man es wählt: `mse-rest` und `mse-mcp`
  binden seit N6
  (#1572) und
  G08 (#2272)
  an eine **einstellbare** Adresse — vorgegeben nur die Rückschleife, auf
  Wunsch ausgewählte Adressen oder alle —, und jenseits der Rückschleife
  gelten TLS und ein Zugang ohne Ausnahme
  ([0115](entscheidungen/0115-die-schleife-ist-eine-vorgabe.md),
  [`NETZEIGENSCHAFTEN.md`](NETZEIGENSCHAFTEN.md), „Als Dienst“). Beide starten
  nicht von selbst. Der **MCP-Server** spricht in der Vorgabe über stdio, ohne
  Port und ohne Adresse ([0010](entscheidungen/0010-mcp-als-eigener-prozess.md));
  über HTTP lauscht er erst auf Zuruf.

  **Davor** stand dort bis zum 2026-08-23 „kein Netzwerkdienst" (Phase U baute
  den REST-Dienst auf `127.0.0.1`,
  [0046](entscheidungen/0046-doch-ein-dienst-auf-der-rueckschleife.md), der
  [0039](entscheidungen/0039-kein-dienst-auf-der-rueckschleife.md) umkehrt) und
  danach „Ein Zugang von einem anderen Rechner kommt weiterhin nicht“ — beides
  gilt nicht mehr.

- Kein Plugin-System für Go-Code — erweitert wird über MSE-Pakete.
- Keine Erweiterung von MSE-Script, die vorhandene Vorlagen brechen könnte.
- Kein Bau für Windows und Linux ([`PORTIERUNG.md`](PORTIERUNG.md)).
- Keine Notarisierung ([0003](entscheidungen/0003-keine-notarisierung.md)).
- Keine fremden Vorlagen im Repository ([0005](entscheidungen/0005-keine-fremden-vorlagen.md)).

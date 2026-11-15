# Chronik

**Was erledigt wurde.** Was offen ist, steht in [`ROADMAP.md`](ROADMAP.md).

Die Trennung ist keine Formsache: offene Arbeit zwischen erledigter zu suchen,
lässt Erledigtes als offen führen — und umgekehrt. Zwei Dateien, zwei Fragen.

**Diese Chronik beginnt mit 1.0.0.** Die Einträge bis 0.9.0 — rund 230
Abschnitte über die Arbeit von 0.1.0 bis zur Freigabe 0.9.0 — liegen im
Entwicklungsarchiv, ebenso die Historie, aus der dieses Repositorium
neu angelegt wurde (Vermerk
[0205](entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)). Hier bleiben zwei
Dinge, auf die sich Bleibendes beruft: die **Phasen bis 1.0.0** als Liste —
sie hält `luecken_test.go` gegen die Pläne in `PARITAET.md` — und die
**Anlässe der Leitlinien**, auf die `AGENTS.md` verweist.

---

## 2026-11-15 — Freigabe 1.0.0

- **1.0.0 ist der Programmstand von 0.9.0** — dieselben Fähigkeiten und dieselben
  Grenzen, neu gebaut mit der Fassung 1.0.0. Was die Fassung nicht belegt, steht
  in der Release-Notiz und in [`REIFEPROFIL.md`](REIFEPROFIL.md).
- **Zwei Sicherheitsaktualisierungen**: Go 1.27.4 (GO-2026-6859 in `net/http`)
  und `klauspost/compress` v1.20.2 (GO-2026-6865) — `govulncheck` meldete beide
  am Stand 0.9.0 als erreicht.
- **Ein eigener Freigabeschlüssel**, getrennt vom Commitschlüssel, wie Vermerk
  [0194](entscheidungen/0194-der-freigabeschluessel-wechselt-ueber-benannte-anker.md)
  es vorsah; beide tragen nur die Noreply-Adresse von GitHub.
- **Ein neues Repositorium unter demselben Namen**: eine Wurzel statt 2.805
  Commits, nur die Freigabe 1.0.0. Das bisherige ist privat archiviert, samt
  Issues, PRs und den Freigaben 0.1.0 bis 0.9.0. Verweise der Gestalt `#N`
  nennen Issues und PRs dort.
- **Am Abbild geprüft**: die Handprüfungen 3, 4 und 7 — hell, dunkel und ein
  Bestand aus 0.7.0 — ohne die echten Orte anzufassen
  ([`belege/freigabebuendel-1.0.0.md`](belege/freigabebuendel-1.0.0.md)); 1
  und 2 von 0.9.0 übernommen, weil die Referenz aus MSE2 hier nicht mehr lag.
  Vorher ist der Baum dreifach auf sensible Daten geprüft worden
  ([`belege/sensibledaten-1.0.0.md`](belege/sensibledaten-1.0.0.md)).
- **Aus der Versionierung genommen**: die Entwürfe der Zieloberfläche, die
  Audits, Abnahmen und alten Pläne, die Chronik und die Historie bis 0.9.0 —
  Vermerk 0205 zählt auf, was ging und welche Wächter mit ihm gingen.

## Bis 1.0.0 — die abgeschlossenen Phasen

Je Phase der Tag, an dem ihr Eintrag entstand, und seine Überschrift; der
Eintrag selbst liegt im Entwicklungsarchiv.

- 2026-10-04 — Phase 6.11: zwei Fenster, ein Tausch (S12.NAT-04)
- 2026-10-04 — Phase 6.11: zwei Fenster, ein Tisch (S12.NAT-04)
- 2026-10-04 — Phase 6.11: Zweitblick und Nachlauf
- 2026-10-04 — Phase 6.11, erster Zug: was aus Phase 6.3 hier lag, und drei Folgen der KI-Phase
- 2026-10-04 — Phase 6.13: abgeschlossen; die KI-Phase 6.10 gebaut
- 2026-10-03 — Phase 6.13, neunter Zug: die Folgen der Durchsicht
- 2026-10-03 — Phase 6.13, achter Zug: Tisch und Zustände, gemessen
- 2026-10-03 — Phase 6.13, siebter Zug: Turniere, Gegenüber, die eigene Abschlussstimme und der Leistungsbeleg
- 2026-10-02 — Phase 6.13, sechster Zug: Rahmen, Spielen, Zeichennaht und der Rest aus Phase 6.5
- 2026-10-02 — Phase 6.13, fünfter Zug: Handeln und drei Folgen
- 2026-10-02 — Phase 6.13, vierter Zug: Sets, Karteneditor, Nutzer und Profile
- 2026-10-02 — Phase 6.13, zweiter Zug: Start, Karten, Werkzeuge, Ausgeben, Sammlung
- 2026-10-02 — Phase 6.13, erster Zug: Prüfmaß und fünf Flächen entschlackt (#1906)
- 2026-10-01 — Phase 6.12: Schnittstellen einrichten und Bereiche schalten (#1905)
- 2026-10-01 — Phase 6.7: Lokale Identität, Verbindung und Teilen (#1668)
- 2026-09-30 — Phase 6.8: Auffindbarer und konsistenter Handel (#1669)
- 2026-09-30 — Phase 6.5: Dateien, Ressourcen und Sicherung (#1666)
- 2026-09-27 — Phase 6.6: Sammlung, Decks und Turniere (#1667)
- 2026-09-26 — Phase 6.4: Konfiguration, Diagnose und Caches (#1665)
- 2026-09-26 — Phase 6.3: Gemeinsame Oberfläche und Bedienmuster (#1664)
- 2026-09-25 — Phase 6.9: Spieltisch und Regelunterstützung (#1670)
- 2026-09-24 — Phase 6.2: Plattform und Wartbarkeit (#1663)
- 2026-09-24 — Phase 6.1: Rechte, Integrität und dringende Fehler (#1662)
- 2026-09-23 — Phase 6.0: Ausgangslage, Verträge und Nachweise (#1661)
- 2026-09-20 — Phase 5.7: Gesamtabnahme und Freigabe (#1411)
- 2026-09-19 — Phase 5.6: Handeln mit ⌘0 (#1410)
- 2026-09-18 — Phase 5.5: der Spieltisch und die Regelmaschine (#1409)
- 2026-09-15 — Phase 5.4: gemeinsame Peer-Verbindung und Objektaustausch (#1408)
- 2026-09-14 — Phase 5.3: externe Daten, semantische Recherche, KI und Agenten (#1407)
- 2026-09-11 — Phase 5.2: Berechtigungen, Konfiguration und Betrieb (#1406)
- 2026-09-10 — Phase 5.1: die Oberfläche bekommt Arbeitsplätze (#1405)
- 2026-09-07 — Phase 4.13: der Lebenszyklus, eine Anwendung, die im Betrieb nicht überrascht (#1352)
- 2026-09-06 — Phase 4.12: Nutzer und Profile, wem die Daten gehören (#1351)
- 2026-09-03 — Phase 4.10: die Oberfläche für Menschen (#1325)
- 2026-09-02 — Phase 4.9: MSE-Script vollständig, und ein Editor dafür (#1324)
- 2026-09-02 — Phase 4.8: drei weitere Kartendienste (#1321)
- 2026-09-01 — Phase 4.7: Turniere (#1294)
- 2026-09-01 — Phase 4.6: Decks und Formatprüfung (#1293)
- 2026-08-31 — Phase 4.11: sechs Einheiten aus MSE2 (#1349)
- 2026-08-31 — Phase 4.5: die Sammlung als tägliches Werkzeug (#1292)
- 2026-08-30 — Phase 4.4: Restparität — vollständig, an jeder Fläche (#1291)
- 2026-08-30 — Phase 4.3: die hundertste Karte (#1290)
- 2026-08-30 — Phase 4.2: Fuzzing, und was ein Fund ist (#1289)
- 2026-08-29 — Phase 4.1: die Wails-Naht wirklich prüfen (#1288)
- 2026-08-29 — Phase 4.0: die Baseline, und was sie über Wächter gelernt hat (#1287)
- 2026-08-28 — Phase 3.8: die Flächenzusage einlösen (#1169)
- 2026-08-28 — Phase 3.8b: Der Sicherungsweg gehört in den Dienst (#1175)
- 2026-08-28 — Phase 3.7: MSE2-Parität nachziehen (#1168)
- 2026-08-27 — Phase 3.6: Magic-Funktionen (#1167)
- 2026-08-27 — Phase 3.5: die Arbeitsfläche für Menschen (#1166)
- 2026-08-27 — Phase 3.4: Sammlung und Preise (#1165)
- 2026-08-27 — Phase 3.3: der Arbeitsbereich für Vorlagen und Skripte (#1164)
- 2026-08-26 — Phase 3.2: Ressourcen, Atomarität, Prozessgrenzen (#1163)
- 2026-08-26 — Phase 3.1: die Sicherung trägt die Bilder (#1162)
- 2026-08-26 — Phase Z: die Aktualisierung, und was der Zweitblick fand
- 2026-08-25 — Phase Ω: originalgetreue Karten, vermessen
- 2026-08-25 — Phase Y: macOS, wie man es erwartet
- 2026-08-25 — Phase X: der Symboleditor
- 2026-08-25 — Phase W: was MSE2 kann, vermessen
- 2026-08-24 — Phase V: fremde Dienste, gekapselt
- 2026-08-23 — Phase U: der Dienst auf der Rückschleife
- 2026-08-23 — Phase T: ein Fachkern, drei Flächen
- 2026-08-23 — Phase S: der Bestand hat einen Rückweg
- 2026-08-23 — Phase R: was ausgeliefert ist, stimmt
- 2026-08-22 — Phase Q: fünf Sätze, die etwas anderes sagten als der Rechner
- 2026-08-22 — Phase P: ein Paket mit Tests ist keine Anwendungsfunktion
- 2026-08-22 — Phase O: fünf Wächter, die weniger hielten als ihr Kommentar
- 2026-08-20 — Phase K: Offen bleibt nur, was besser werden soll
- 2026-08-08 — Phase 0: Vorbereitung
- 2026-08-08 — Phase 1, Werkstück 1: Werkzeugpinnung
- 2026-08-08 — Phase 1, Werkstück 2: Frontend-Gerüst
- 2026-08-08 — Phase 1, Werkstück 3: Wails-Gerüst
- 2026-08-08 — Phase 1, Werkstück 4: Sprachkataloge und der Wächter gegen Text im JSX
- 2026-08-08 — Phase 1, Werkstück 5: Schichtenwächter
- 2026-08-08 — Phase 1, Werkstück 6: Fehlergerüst und Schlüsselvertrag
- 2026-08-08 — Phase 1, Werkstück 7: SQLite-Grundlage
- 2026-08-08 — Phase 1, Werkstück 8: Freigabeweg
- 2026-08-08 — Phase 1: der Zweitblick und was er gekostet hat
- 2026-08-08 — Phase 1 abgeschlossen
- 2026-08-08 — Phase 2 abgeschlossen
- 2026-08-08 — Phase 3 abgeschlossen
- 2026-08-08 — Phase 4 abgeschlossen
- 2026-08-08 — Phase 5: die Werkstücke
- 2026-08-09 — Phase 5 abgeschlossen
- 2026-08-16 — Phase 9 wirklich abgeschlossen
- 2026-08-16 — Phase 8: Freigabe
- 2026-08-16 — Phase 7: Paketverwalter und Saat
- 2026-08-15 — Phase 9: Fremdformate und Schnittstelle
- 2026-08-15 — Phase 6, Werkstück: ein Set zurückschreiben
- 2026-08-15 — Phase 6, Werkstück: Bilder ausgeben
- 2026-08-15 — Phase 6, Werkstück: der Druckbogen
- 2026-08-15 — Phase 6, Werkstück: der Zeichenplan aus einem echten Stil
- 2026-08-15 — Phase 6, Werkstück: Rückgängig über Sitzungen hinweg
- 2026-08-15 — Phase 6, Werkstück: Bilder, Masken und verborgene Felder
- 2026-08-15 — Phase 6, Werkstück: eine MSE-Datei ist nicht notwendig ein Baum
- 2026-08-15 — Phase 6, Werkstück: ein Wächter über die SQLite-Pragmas
- 2026-08-15 — Phase 6, Werkstück: eine zurückgenommene Entfernung behält ihre Stelle
- 2026-08-15 — Phase 6, Werkstück: die Bezeichnerregel sagt, was sie meint
- 2026-08-15 — Phase 6, Werkstück: der Stilplan bekommt die Umgebung des Stils
- 2026-08-15 — Phase 6, Werkstück: der Schreiber entfernt und ergänzt Feldzeilen
- 2026-08-15 — Phase 6, Werkstück: ⌘S, die Fensternaht und das Meldungsband
- 2026-08-15 — Phase 6, Werkstück: der Zwischenspeicher der Bildquelle
- 2026-08-15 — Phase 6, Werkstück: die Laufweite zählt Buchstaben
- 2026-08-15 — Phase 6, Werkstück: Exportvorlagen lesen und ausführen
- 2026-08-15 — Phase 6, Werkstück: die Prozesse bemerken einander
- 2026-08-15 — Phase 6, Werkstück: die Rücknahme über den MCP-Server ist vollständig
- 2026-08-15 — Phase 6: die Sichtprüfung und was sie fand
- 2026-08-15 — Phase 6: der Zweitblick durch Codex
- 2026-08-15 — Phase 6: der Nachlauf (#211)
- 2026-08-15 — Phase 6 ist abgeschlossen
- 2026-08-16 — Phase A: der Abstand zu MSE2 ist beziffert
- 2026-08-16 — Phase A: die Naht wird geprüft, nicht das Zwischenmodell
- 2026-08-16 — Phase A abgeschlossen, und was ihr Zweitblick fand
- 2026-08-16 — Phase B: der zweistufige Satz, und was er nicht bewegt
- 2026-08-16 — Phase C: der Zeichenplan trägt einen Baum
- 2026-08-16 — Phase D: die Auszeichnungen stehen nicht mehr roh da
- 2026-08-16 — Phase E: die Wahlbilder lösen auf
- 2026-08-19 — Phase I: der Abstand bekommt eine Adresse
- 2026-08-19 — Phase J: acht Ursachen, 26,714 auf 22,426
- 2026-08-19 — Phase G abgeschlossen: Die Oberfläche wird benutzbar

---

## 2026-10-07 — Anlässe der Leitlinien (WART-01)

**Die Leitlinien behalten ihre Regeln und deren Gründe; die Geschichte, aus
der eine Regel entstand, steht hier** (#2172).
Bis dahin stand sie in `AGENTS.md` neben der Regel, und wer die Regel suchte,
las sich durch ihren Anlass.

**§8, der Grund in einer Meldung.** Der Satz stand ohne die Einschränkung auf zwei Stellen, während die zwei
Stellen längst so arbeiteten; angestrichen vom Nachlauf zum Zweitblick über
Phase 4.13.

**§10, „Ein Kopf, der mehr verspricht".** Der Satz steht hier, weil er am 2026-08-30 in Phase 4.2 **siebenmal** zutraf
(#1289):
„zweimal dieselbe Eingabe, zwei Ergebnisse" stand als Fund im Kopf von vier
Fuzzzielen, und verglichen wurden zwei Felder von vielen. Der Nachlauf zeigte
ihn eine Ebene tiefer: die Korrektur verglich das ganze Ergebnis — auf dem
**Erfolgsweg**. Die eingebaute Prüfung der Ablehnung wurde sofort rot und
legte einen echten Fehler frei: dieselbe Antwort von Scryfall nannte zweimal
ein anderes fehlendes Pflichtfeld, weil die Prüfung über ein `map`-Literal
ging.

**§10, „Eine Gegenprobe, die nichts ändert".** Der Satz steht hier, weil er am 2026-08-20 in
#661
**zweimal** falsch gelesen wurde. Einmal lag dieselbe Zusage an **zwei**
Stellen — `Feldzeile` und `Feldeingabe` prüfen beide, ob das Feld leer ist —,
und das Entschärfen der einen ließ die andere stehen. Einmal wurde ein Zweig
entfernt, den der Fall gar nicht durchläuft. Beide Male sah es aus wie ein
Befund, und beide Male war der Test in Ordnung.

**§10, „Die Reichweite einer Messung".** Der Satz steht hier, weil derselbe Fehler in einer einzigen Iteration
**dreimal** vorkam: `965 recolor_image → 0` war Buchführung statt Wirkung,
der erste Nahtwächter maß „nicht weiß" statt „Rahmen", und „im Pack liegt
keine einzige `.mse-symbol`-Datei" war eine Aussage über ein Verzeichnis. Der
dritte Fall kostete am meisten: er stand als **Begründung** dafür, dass ein
Issue keine Abnahme nennen könne — und es kann eine.

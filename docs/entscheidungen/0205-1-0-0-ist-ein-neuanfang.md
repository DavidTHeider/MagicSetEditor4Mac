# 0205 — 1.0.0 ist ein Neuanfang: der Programmstand von 0.9.0 in einem neuen Repositorium

**Datum:** 2026-11-15
**Stand:** angenommen — entschieden vom Auftraggeber am 2026-11-15, Schritt
für Schritt in derselben Sitzung
**Anlass:** die erste Freigabe für Anwendende, und vor ihr die Frage, was ein
öffentliches Repositorium enthalten darf (Vermerk
[0190](0190-das-repositorium-wird-oeffentlich.md), „Vor dem Öffnen“)

## Lage

0.9.0 war am 2026-10-21 freigegeben, begrenzt (Reifeprofil, Handprüfungen 5
und 6 beim Menschen). Das Repositorium war privat und trug, was neun
Entwicklungsfassungen hinterlassen hatten: 2.805 Commits und 1,48 GiB
Historie, zwölf Tags, neun Freigaben, 1.243 Issues und 1.300 PRs — und Dinge,
die nicht öffentlich gehören. Erhoben am 2026-11-15:

- **Der eingebettete Freigabeschlüssel trug eine private E-Mail-Adresse als
  Kennung** — in `internal/aktualisierung/freigabeschluessel.asc`, damit in
  jedem gebauten Programm seit dem 2026-08-26, und im README seit dem
  2026-08-25. Derselbe Schlüssel signierte die Commits.
- **GitHub hält jeden PR samt seiner Commits** (`refs/pull/*/head`, nur
  lesbar). Ein zurückgesetzter Zweig im selben Repositorium hätte die alte
  Historie über die PRs erreichbar gelassen — und mit ihr die Adresse. GitHub
  entfernt solche Verweise nach eigener Auskunft nur, wo sich ein Risiko nicht
  durch das Wechseln von Zugangsdaten beheben lässt
  (`docs.github.com`, „Removing sensitive data from a repository“, gelesen am
  2026-11-15).
- **`govulncheck` meldete am Stand 0.9.0 zwei erreichte Schwachstellen**:
  GO-2026-6859 in `net/http` und GO-2026-6865 in `klauspost/compress`.

## Entscheidung

| Was | Festlegung |
| --- | --- |
| **Programmstand** | der von 0.9.0. **Neu gebaut** mit der Fassung 1.0.0 — ein unverändertes Abbild hätte die private Adresse in jedem Programm getragen, sich als 0.9.0 gemeldet, und seine Aktualisierungsprüfung hätte 1.0.0 für neuer gehalten und an der signierten Liste verworfen ([0194](0194-der-freigabeschluessel-wechselt-ueber-benannte-anker.md), Punkt 5) |
| **Sicherheit** | Go 1.27.4 und `klauspost/compress` v1.20.2 — ohne sie wäre die Stufe `security` rot und kein Bau möglich |
| **Schlüssel** | zwei neue, beide ohne Passphrase wie der bisherige, beide nur mit der Noreply-Adresse von GitHub: **Freigabe** `44C59428DE1F766E` (Fingerabdruck `97721629ABDC00BA1711488244C59428DE1F766E`), eingebettet als einziger Anker; **Commits** `1503F87D4C128638`, als `user.signingkey` dieses Repositoriums. Damit gilt die Trennung, die 0194 Punkt 1 vorsah. Erzeugt hat sie der Agent auf ausdrückliche Bitte des Auftraggebers — anders als 0194 es vorsah; die Widerrufszertifikate liegen unter `~/.gnupg/openpgp-revocs.d/` und bleiben auf Entscheidung des Auftraggebers auf diesem Rechner (Nachtrag zu 0194) |
| **Repositorium** | **neu angelegt unter demselben Namen** `DavidTHeider/MagicSetEditor4Mac`; das bisherige heißt `MagicSetEditor4Mac-archiv` und bleibt privat — mit Historie, Issues, PRs und den Freigaben 0.1.0 bis 0.9.0. Kanal der Aktualisierung, Go-Modulpfad und Bundle-Kennung bleiben, wie sie sind |
| **Historie** | eine Wurzel; was danach kommt, geht wie bisher über Zweig und PR (§5). Der erste Push in das leere Repositorium geht ohne PR, wie §5 es für den ersten Push vorsieht |
| **Freigaben** | nur 1.0.0. Die Entwicklungsfassungen sind nicht mehr veröffentlicht; `CHANGELOG.md` fasst sie in einer Tafel zusammen |
| **Verweise `#N`** | ein Issue oder PR des Archivs. Die 15.505 Verweise mit Adresse sind zu Klartext geworden; Adressen ohne Nummer (`…/issues`, `…/releases`) gelten im neuen Repositorium weiter |
| **Aus der Versionierung genommen** | die Entwürfe der Zieloberfläche (`PROTOTYP-0.6.0.md`, `PROTOTYP-0.9.0.md`, `docs/prototyp/`), `SEITENKRITIK.md`, `WEBAPP-GUTACHTEN.md`; die Audits und Abnahmen (`docs/audits/`, `AUDIT-0.2.0.md`, `AUDIT-0.3.0.md`, `ABNAHME-*.md`), `CODEX-NAECHSTE-RUNDE.md`, `ZIELVORSCHLAG.md`, `docs/superpowers/specs/`; `HISTORIE.md`; die Chronik, Roadmap und das Protokoll bis 0.9.0; ein Bildschirmabzug, auf den nichts mehr verweist (`magic-m15-rahmen.png`). Sie liegen im Archiv und beim Auftraggeber lokal — **außerhalb des Arbeitsbaums**, damit ein Wächter dort dasselbe sieht wie in einem frischen Klon |
| **Was weiterlebt** | das Reifeprofil als `docs/REIFEPROFIL.md`; aus `UIREVIEW.md` das Prüfmaß R1 bis R7, auf das sich Code und Tests berufen; aus der Chronik die Liste der Phasen und die Anlässe der Leitlinien |

### Die Wächter, die mit ihrem Gegenstand gingen oder ihn wechselten

| Wächter | Was geschah |
| --- | --- |
| `prototyp_test.go`, `prototyp09_test.go` | entfallen mit den Entwürfen; die Liste der Seiten aus `navigation.ts` lebt als `navigationsseiten` in `fensterseiten_test.go` weiter |
| `chronikumzug_test.go` | entfallen — er hielt den Umzug der Phasenberichte in die Chronik, und die Berichte liegen im Archiv |
| `inventurwege_test.go` | entfallen — er hielt die Arbeitswege des Audits 0.6.0 gegen die Inventur |
| `scripts/befundzuordnung.*` | entfallen — Skript und Wächter über die Zuordnung der Audit-Befunde 0.9.0 |
| `luecken_test.go` | liest die abgeschlossenen Phasen aus der Liste der Chronik; eine „fehlt“-Zeile darf kein Issue aus dem Archiv nennen, und eine Fassung gilt als erschienen, wenn sie oder eine spätere eine datierte Überschrift hat |
| `changelog_test.go` | hält den Abschnitt 1.0.0 statt 0.1.0 |
| `pruefstand_test.go` | weist einen Beleg ab, der ein Issue des Archivs nennt — Anwendende öffnen ihn nicht |
| `scripts/releasesignieren.sh` | signiert mit dem Schlüssel, den die Anwendung nennt, statt mit `user.signingkey` |

## Verworfen

- **Das Abbild von 0.9.0 unverändert als 1.0.0.** Siehe oben: Adresse im
  Programm, falsche Fassung, eine Aktualisierungsprüfung, die sich selbst
  widerspricht.
- **Dasselbe Repositorium mit zurückgesetztem Zweig.** Billiger, und die
  Verweise blieben klickbar — aber die alte Historie bliebe über 1.300 PRs
  erreichbar.
- **Die Verweise aufs Archiv umbiegen.** Für den Auftraggeber klickbar, für
  alle anderen 18.000 Verweise ins Leere.
- **Ein Übergang über zwei Freigaben** (0194 Punkt 3). Ihn braucht, wer eine
  installierte Fassung mitnehmen muss; einen öffentlich installierten Bestand
  gab es nicht.

## Was das nicht leistet

- **Was vorher draußen war, bleibt draußen.** `https://github.com/DavidTHeider.gpg`
  lieferte den bisherigen Schlüssel samt privater Adresse ohne Anmeldung aus;
  wer ihn abgerufen hat, hat ihn. **Seit dem 2026-11-15 liegt er nicht mehr im
  Konto** — der neue Commitschlüssel liegt dort, und die Adresse liefert nur
  noch Schlüssel mit Noreply-Kennung. Die Commits im Archiv bleiben dabei
  verifiziert: GitHub hält eine einmal geprüfte Signatur fest, auch wenn der
  Schlüssel später geht (`docs.github.com`, „About commit signature
  verification", gelesen am 2026-11-15).
- **Eine Nummer `#N` ist nicht eindeutig.** Das neue Repositorium zählt von
  vorn; ein `#12` aus der Zeit bis 1.0.0 meint das Archiv, eines danach dieses
  Repositorium — wer ein neues Issue nennt, nennt es mit Adresse.
- **Eine installierte Entwicklungsfassung nimmt 1.0.0 nicht über die
  Aktualisierung an.** Sie kennt nur den alten Anker; 1.0.0 kommt dann als
  Abbild von Hand.
- **Die Belege bis 0.9.0 nennen Commits, die es hier nicht gibt.** Wer einen
  davon aufschlagen will, findet ihn im Archiv; die Belege, die ein Lauf am
  Stand 1.0.0 erneuert, nennen die neue Wurzel.

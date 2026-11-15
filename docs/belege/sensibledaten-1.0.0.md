# Beleg: der Baum vor der Veröffentlichung, auf sensible Daten geprüft — 1.0.0

**Geprüft am 2026-11-15** am Zweig `release/1.0.0`, der zur Wurzel von 1.0.0
wurde; die beiden Prüfer lasen den Stand `d2922c48f`, die letzte Gegenprobe
lief am Stand `1f87cb900`. Gegenstand sind die **verfolgten** Dateien
(`git ls-files`, 12.039), nicht die Historie: die beginnt mit 1.0.0 neu
(Vermerk [0205](../entscheidungen/0205-1-0-0-ist-ein-neuanfang.md)). Die
gebauten Artefakte prüft ein eigener Abschnitt im Freigabebeleg 1.0.0.

## Drei Prüfer, und keiner kannte die Funde des anderen

| Prüfer | Wie | Reichweite |
| --- | --- | --- |
| **Ein Skript**, deterministisch | 24 Muster über jede verfolgte Datei, binäre Dateien wie `strings` | erschöpfend für die Muster, blind für alles ohne Muster |
| **Ein Agent (Sonnet)**, in die Tiefe | die 68 Bildschirmabzüge per OCR, 29 davon angesehen; die 84 Bilder unter `docs/belege/` per OCR, ein Dutzend angesehen; `docs/belege/` nach Mustern und gezielt gelesen; die Sicherungen unter `testdaten/` ausgepackt und ihre Datenbanken geöffnet | 396 Belegdateien gesucht, nicht alle gelesen; der Rest der Bilder nur per OCR |
| **Codex (`gpt-6.1-terra`, Denkstufe hoch)**, in die Breite | Mustersuche über den ganzen Baum, ZIP-Archive inventarisiert, SQLite-Sicherungen tabellenweise gesichtet, Bildmetadaten | keine Volltextlesung, keine vollständige Rechteprüfung |

**Gesucht** nach privaten E-Mail-Adressen, personenbezogenen Daten (Namen,
Adressen, Telefonnummern, Kontakte, echter Bestand), Geheimnissen (Schlüssel,
Marken, Kennwörter, private Schlüssel aller Formate), Kennungen des
Heimnetzes (Rechnernamen, Seriennummern, Hardware-UUIDs, MAC-Adressen,
öffentliche Adressen eines Anschlusses) und — zweitrangig — fremdem Material.

## Was gefunden wurde, und was damit geschah

| Fund | Gefunden von | Verbleib |
| --- | --- | --- |
| Eine private E-Mail-Adresse als Kennung des eingebetteten Freigabeschlüssels und im README | Erhebung vor dem Plan | entfernt; der Schlüssel ist ersetzt (Vermerk 0205) |
| Der echte Rechnername des Wirts (`…-von-<Vorname>.fritz.box`) in Doku, einem Kommentar und Tests | Skript | durch `MacBook-Air-von-Ada.fritz.box` ersetzt, die Gestalt des Messblocks bleibt |
| Das **öffentliche IPv6-Präfix des Anschlusses** und das eindeutig lokale Präfix der Fritz!Box in `huellenprobe/g08.md` und zwei Tests | Sonnet | durch `2001:db8::/32` und `fd12:3456:789a::/48` ersetzt; die Tests prüfen dieselbe Ordnung und sind grün |
| **Zwei vollständige Einladungen** aus dem Fernweg zwischen zwei VM-Gästen, mit dem vorab geteilten Schlüssel | Codex | gekürzt; der Lauf schreibt die Einladung nicht mehr ins Protokoll, nur ihre Länge |
| Rechnernamen der VM-Gäste in `sec22-vmlauf.md` | Codex, Sonnet | durch `gast-a.local` und `gast-b.fritz.box` ersetzt |

## Was bleibt, und warum

- **Der Name des Urhebers** — in `LICENSE`, im Bündel und in Vermerk 0190; die
  GPL verlangt einen Urhebervermerk. „David" steht außerdem als Name eines
  Mitspielers in Tests.
- **Der Pfadteil `/Users/david`** in Skripten, Doku und Belegen — er verrät den
  Kontonamen, der im öffentlichen Handle steckt; `/Users/lume` ist der
  Standardnutzer der VM.
- **Private Adressen** (`192.168.178.x`, `10.x`) in Tests und Belegen — sie
  sind nicht routbar und hängen an keiner Identität mehr.
- **Peerkennungen** aus VM-Läufen und Hüllenproben — öffentliche Schlüssel von
  Identitäten, die für den Lauf in einem Wegwerfbund entstanden.
- **Testwerte**: `einlangesgeheimnis`, `+4915112345678`, Adressen unter
  `example.org`, das Geheimnis aus RFC 6238, der Wegwerfbund mit dem Kennwort
  `leer`.
- **Karten mit Kunst und Rahmen von Wizards of the Coast** in den
  Bildschirmabzügen — Inhalte Dritter, für die die Abgrenzung in `LICENSE`,
  `README.md` und `DRITTANBIETER.md` gilt.
- **Bildmetadaten**: in 132 Abzügen nur Pixelmaße und der Kommentar
  „Screenshot", kein Autor, kein Ort, keine Gerätekennung.

**Unsicher und dem Auftraggeber vorgelegt**: der Name des Gegenübers
„Turfer" im Handelsbild, in der Befehlspalette, in Vermerk 0153 und in Tests.
Er sieht erfunden aus; ob er es ist, weiß nur der Auftraggeber.

## Die Muster des Skripts und ihre Zählung am Stand `1f87cb900`

Gemeldet steht für Treffer außerhalb der Erlaubnisliste, die danach einzeln
eingeordnet wurden; **null** steht für keine einzige Fundstelle.

| Muster | gemeldet | Einordnung |
| --- | ---: | --- |
| private Adresse (t-online, gmail, icloud …) | 0 | — |
| E-Mail außerhalb der Platzhalter | 10 | Bytefolgen in PNG-Daten, zwei Adressen mit Benutzerangabe in Tests gegen Marktverweise |
| Heimnetz (`*.home`, `fritz.box`, `*.ts.net`) | 37 | Platzhalter wie oben und `.HOME` in Skripten |
| private IPv4 | 163 | siehe „Was bleibt" |
| Tailscale-Bereich `100.64/10` | 3 | der Bereich selbst in der Netzlogik |
| MAC-Adresse, Seriennummer, Rechnername eines Mac | 0 | — |
| Schlüssel von Anthropic, OpenAI, xAI, GitHub, AWS, Google, Slack | 0 | — |
| privater Schlüssel, age-Geheimnis, JWT, Bearer | 0 | — |
| Kennwort- oder Markenzuweisung | 91 | Feldnamen, Katalogtexte, Testwerte |
| Telefonnummer | 45 | Testwerte und Ziffernfolgen in Messdaten |
| IBAN | 0 | — |
| Nutzerpfad außer `/Users/david` | 235 | `/Users/lume` und Testnamen |
| öffentlicher PGP-Schlüssel | 1 | der Freigabeschlüssel mit Noreply-Kennung |

**Was dieser Beleg nicht sagt**: dass nichts übersehen ist. Eine Mustersuche
findet, wofür sie ein Muster hat; zwei Agenten lasen gezielt und nicht jede
Datei. Was die drei zusammen nicht gesehen haben, ist nicht geprüft.

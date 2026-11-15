# Hüllenprobe: fix-f02

**bestanden** — Commit `53cf6c652`,
2026-10-17 09:41 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-17T09-41-37Z-21620`, Bauabdruck `6fbb736c3d21`.


**Ein Set, nativ angelegt: Name, eine Karte mit Text und Bild, ein Symbol**

- ✓ der Stil ist »Mit Kunstfenster« (probe-kunst)
- ✓ die Karte bietet die Felder der Saatvorlage — kunst, name, text
- ✓ »Bild wählen« öffnet den Dialog von macOS
- Bild: `/Users/david/mse-arbeit/w33/.huellenprobe/laeufe/2026-10-17T09-41-37Z-21620/bilder/fix-f02/f02-karte.png`

**⌘S bei einem Set ohne Datei fragt, wohin — S01.NAT-01 und NAT-03**

- ✓ ⌘S öffnet den Dialog von macOS — Set sichern unter, Vorschlag »set.mse-set«
- ✓ der Abbruch schrieb nichts
- im Archiv: kunst1, set, symbol1.mse-symbol
- ✓ das Set trägt das Bild der Karte — kunst1, set, symbol1.mse-symbol
- ✓ und das gezeichnete Symbol — kunst1, set, symbol1.mse-symbol
- ✓ und den Namen und den Regeltext der Karte

**„Sichern unter …" geht denselben Weg — S01.NAT-02**

- ✓ beide Dateien tragen dasselbe Set

**Ein Schreibfehler nennt die Ursache und lässt die alte Datei stehen — NAT-03**

- ✓ die Meldung nennt die Ursache und den nächsten Schritt — Das Set „Nachtprobe“ wurde nicht gesichert — Auf diesen Ort darf dieses Programm nicht zugreifen — er ist gesperrt. Wähle einen anderen oder gib ihn frei, in den Rechten der Datei oder unter Systemeinstellungen › Datensc
- ✓ die alte Datei steht unverändert
- ✓ ⌘Q beendet

**In einem frischen Bestand öffnen — NAT-02**

- ✓ ⌘O öffnet den Dialog von macOS
- ✓ das Set öffnet mit seiner Karte — 1	0	Nachtfalter
- ✓ ⌘Q beendet

**Unbekannte Felder bleiben — NAT-02**

- ✓ die unbekannte Zeile in der Karte bleibt
- ✓ und der unbekannte Block
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

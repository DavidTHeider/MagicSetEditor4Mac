# Hüllenprobe: eintraege

**bestanden** — Commit `1c0b74221`,
2026-10-09 21:59 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.


**Ein Fensterbild wählen — UX-09.AK-01, und der Fokus nach dem Abbruch (UX-08.AK-03)**

- ✓ »Bild wählen …« (Klick) öffnet den Dialog von macOS
- ✓ »Bild wählen …« (⏎) öffnet den Dialog von macOS
- der Fokus nach dem Abbruch, ohne Schlüsselfenster: »Bild wählen …«
- ✓ das gewählte Bild steht in der Vorschau und hinter der Oberfläche — url("/fensterbild/1.png?stand=3")
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/eintraege/fensterbild.png`

**Darstellung zurücksetzen — Farbwelt und Fensterbild fallen auf die Vorgabe**

- ✓ vorher steht die Farbwelt aus dem Profil — schiefer
- ✓ danach steht die Vorgabe — pergament
- ✓ und hinter der Oberfläche steht kein Bild mehr
- ✓ und im Profil steht keine Farbwelt mehr

**Ein Set aufschlagen — die Kartenliste steht da**

- die Brotkrume nach dem Aufschlagen: »Langtext«

**Rückgängig und Wiederherstellen — am Bestand gelesen**

- ✓ »Rückgängig« nimmt die Änderung im Bestand zurück — 1	0	Langtext
- nach dem Rücknehmen: 1	0	Langtext
- ✓ »Wiederherstellen« bringt sie wieder — 1	0	Langtext Neu
- der Name vor dem Tippen war »1	0	Langtext«, danach »1	0	Langtext Neu«

**Alle Karten anhaken**

- ✓ »Alle Karten anhaken« hakt die fünf Karten an — 0 → 5

**Die Einträge, die auf eine Seite führen**

- ✓ »Ablage › Ausgeben › Über eine Exportvorlage …« führt auf »Ausgeben«
- ✓ »Ablage › Ausgeben › Als Bilder …« führt auf »Bilder«
- ✓ »Ablage › Ausgeben › Druckbogen …« führt auf »Druckbogen«
- ✓ »Ablage › Drucken …« führt auf »Druckbogen«
- ✓ »Set › Suchen …« führt auf »Suche«
- ✓ »Set › Ersetzen …« führt auf »Ersetzen«
- ✓ »Hilfe › MagicSetEditor4Mac-Hilfe« führt auf »Hilfe«

**»Über« führt auf den Abschnitt mit Lizenz und Rechten Dritter**

- ✓ er nennt die Lizenz und die Inhalte Dritter — Über MagicSetEditor4Mac · Fassung · 0.7.0

**Zurück und Vorwärts gehen den Gang der Seiten**

- ✓ »Zurück« führt auf die Seite davor — Ersetzen
- ✓ »Vorwärts« führt wieder hin — Hilfe

**Die Seitenleiste ein- und ausblenden**

- ✓ der Eintrag klappt die Seitenleiste — Seitenleiste ausblenden → Seitenleiste einblenden
- ✓ ⌥⌘S klappt sie zurück

**Zuletzt benutzt schlägt das Set auf**

- ✓ »Schließen« schließt das Set, die Liste der Sets steht da — Sets
- ✓ »Zuletzt benutzt« nennt das Set — Werkstatt-Testkarten
- ✓ »Zuletzt benutzt« schlägt es auf — Langtext Neu

**Bestand sichern — der Dateidialog, dann die Sicherung**

- ✓ »Bestand sichern …« öffnet den Dialog von macOS — Bestand sichern, Vorschlag »mse-sicherung.msesicherung«
- vorgeschlagen: »mse-sicherung.msesicherung«
- ✓ »Bestand sichern …« schreibt an den gewählten Ort — 274016 Bytes

**Kernweg 3 — ein Verzeichnis für MTG JSON wählen**

- ✓ »Verzeichnis öffnen …« öffnet den Dialog von macOS
- ✓ der Dialog wählt Ordner — {"art":"oeffnen","ordner":"~/Documents","ordnerwahl":true}
- ✓ das Verzeichnis kommt als Sets in den Bestand — 1 → 3 Zeilen
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

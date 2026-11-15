# Hüllenprobe: uxrf6

**bestanden** — Commit `1c0b74221`,
2026-10-09 22:20 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.

- normal: {"X":95,"Y":119,"B":1280,"H":800}
- auf der Seite: {"titel":"Werkstatt-Testkarten","zeilen":5,"gitter":1,"reihen":5,"karten":5,"sichtbar":5}
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf6/karten-vorher.png`
- Karten vorher: ["Langtext","Sonderlayout — ein Name, der so lang ist, dass er das Namensfeld sprengt","Zweiseitig","Ohne Bild","Sonderzeichen ÄÖÜ ß «» ✦ ☙ 漢字 🃏"]
- ✓ ohne Wahl steht keine Zeile über der Liste

**Die Verteilung des Felds »farbe« — die Balken sind Knöpfe**

- Balken: [{"text":"rot 3 (60 %)","gedrueckt":"false","tabstelle":0},{"text":"blau 2 (40 %)","gedrueckt":"false","tabstelle":0}]
- ✓ zwei Balken, beide Knöpfe in der Tabfolge und keiner gewählt
- ✓ der Satz „Einen Balken wählen …" steht da
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf6/statistik-balken-ungewaehlt.png`

**Wählen mit der Maus — und aufheben mit demselben Klick**

- ✓ blau bleibt ungewählt
- ✓ die Wahl steht in Worten: »Gewählt: »rot« · Karten: 3«
- ✓ *Karten anzeigen* ist der einzige Hauptknopf der Schublade — ["Karten anzeigen"]
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf6/statistik-balken-rot.png`
- ✓ derselbe Klick hebt die Wahl auf, und der Satz steht wieder da

**Wählen mit der Tastatur — Tab-Stelle und Enter**

- ✓ der Balken »blau« nimmt den Fokus
- ✓ Enter wählt »blau«, und rot ist ungewählt

**Der Weg zur Liste — nur die Karten des Balkens**

- Karten gefiltert: ["Sonderlayout — ein Name, der so lang ist, dass er das Namensfeld sprengt","Ohne Bild"]
- ✓ die zwei blauen Karten, und nur sie
- Zeile über der Liste: "Aus der Statistik: »farbe« = »blau« 2 von 5 Karten Zur Statistik Filter aufheben"
- ✓ die Zeile nennt den Filter und zählt „2 von 5 Karten"
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf6/karten-gefiltert.png`

**Der Rückweg — die Wahl bleibt, die Schublade geht auf**

- ✓ der Balken »blau« steht gewählt da
- ✓ *Wahl aufheben* gibt alle fünf Karten wieder, und die Zeile ist fort

**Aufheben auf der Liste — der Knopf in der Zeile**

- ✓ drei rote Karten
- ✓ *Filter aufheben* gibt alle fünf Karten wieder

**Die Auszählung nach der erklärten Dimension »farbe« — dieselbe Wahl**

- Zeile über der Liste (Dimension): "Aus der Statistik: »farbe« = »blau« 2 von 5 Karten Zur Statistik Filter aufheben"
- ✓ dieselben zwei Karten, wenn nach der Dimension gewählt ist

**Die Belegung bleibt zum Lesen**

- ✓ kein Balken der Belegung ist ein Knopf
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf6/statistik-belegung.png`
- ✓ ⌘Q beendet die Anwendung
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

# Hüllenprobe: ap5kontext09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 10:18 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**hell: der Fenstertitel ohne und mit offenem Set**

- ✓ ohne offenes Set heißt das Fenster wie die Anwendung — MagicSetEditor4Mac
- ✓ mit offenem Set nennt der Titel das Set vor der Anwendung — Werkstatt-Testkarten — MagicSetEditor4Mac

**hell: die Kürzel im Menü »Karte«**

- ✓ »Karte anlegen« trägt ⌘↩ — ["Karte anlegen","⌘↩"]
- ✓ »Karte duplizieren« trägt ⌘D — ["Karte duplizieren","⌘D"]
- ✓ »Karte löschen …« trägt ⌘⌫ — ["Karte löschen","⌘⌫"]
- ✓ »Vorige Karte« trägt ⇞ — ["Vorige Karte","⇞"]
- ✓ »Nächste Karte« trägt ⇟ — ["Nächste Karte","⇟"]
- Bild der Leiste: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/menue-karte-hell.png`

**hell: ⌘⌫ in einem Feld löscht Text, keine Karte**

- ✓ das Namensfeld steht da
- ✓ das Feld hat den Fokus und Text — "Langtext Probe"
- ✓ ⌘⌫ im Feld nimmt Text weg — "Langtext Probe" → ""
- ✓ ⌘⌫ im Feld lässt die Karten stehen — 5 → 5
- ✓ ⌘⌫ im Feld stellt keine Löschfrage

**hell: ⌘D, ⌘↩, ⇟, ⇞ und ⌘⌫ außerhalb der Felder**

- ✓ ⌘D verdoppelt die Karte — 5 → 6
- ✓ ⌘↩ legt eine Karte an — 6 → 7
- ✓ ⇞ blättert zur vorigen Karte — 7 → 6
- ✓ ⇟ blättert zur nächsten Karte — 6 → 7
- ✓ ⌘⌫ außerhalb löscht nicht ohne Rückfrage — 7 → 7
- ✓ ⌘⌫ außerhalb stellt die Löschfrage
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/loeschfrage-hell.png`

**hell: Rechtsklick an einer anderen Karte der Liste**

- ✓ einer anderen Karte der Liste steht da — {"x":433,"y":405.046875}
- ✓ das Menü an einer anderen Karte der Liste ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-anderen-Karte-der-Liste-hell.png`
- ✓ der Rechtsklick wählt die andere Karte — 1
- ✓ das Menü gehört ihr — karte/1
- ✓ es nennt Anlegen, Verdoppeln und Löschen — +Karte anlegen⌘↩ · ⧉Karte duplizieren⌘D · ⎘Kopieren (1)⌘C · Ausschneiden (1) · Einfügen · Vorige Karte⇞ · Nächste Karte⇟ · Nach oben · Nach unten · Bezug zu einem Druck … · Teilen … · ✕Karte löschen⌘⌫
- ✓ der untersten Zeile steht da — {"x":433,"y":410.546875}
- ✓ das Menü an der untersten Zeile ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-der-untersten-Zeile-hell.png`
- ✓ auch dort gehört das Menü der Karte darunter — karte/7

**hell: ⇧F10 an der Zeile, Escape gibt ihr den Fokus zurück**

- ✓ ⇧F10: der Fokus steht im Menü
- ✓ Escape gibt der Zeile den Fokus zurück — <div role="row" aria-rowindex="1" aria-selected="true" aria-label="Ohne Namen" t

**hell: ein Wechsel über die Leiste schließt ein offenes Menü**

- ✓ ein Objektwechsel schließt das Menü — 2

**hell: Rechtsklick an einer Setkarte**

- ✓ einer Setkarte steht da — {"x":439.2421875,"y":292.5625}
- ✓ das Menü an einer Setkarte ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Setkarte-hell.png`
- ✓ das Menü der Setkarte gehört dem Set — set/1
- ✓ es nennt Sichern und Löschen — ⤓Sichern · ⤓Sichern unter … · Als Umschlag ablegen · Über eine Exportvorlage … · ✕Löschen …

**hell: Rechtsklick an einer Paketzeile und einer Gruppe**

- ✓ einer Paketzeile steht da — {"x":821,"y":393.5546875}
- ✓ das Menü an einer Paketzeile ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Paketzeile-hell.png`
- ✓ das Menü der Paketzeile gehört ihrem Paket — probe → paket/mse-game/probe
- ✓ der Rechtsklick wählt das Paket — probe
- ✓ einer Paketgruppe steht da — {"x":821,"y":350.0078125}
- ✓ an einer Paketgruppe geht kein Menü der Seite auf

**hell: Rechtsklick an einer Vorlagendatei und am Paket im Baum**

- ✓ einer Vorlagendatei steht da — {"x":441,"y":339.3828125}
- ✓ das Menü an einer Vorlagendatei ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Vorlagendatei-hell.png`
- ✓ das Menü gehört der Datei — dunkel.png → datei//var/folders/8x/tzk9pq5s2g52zjr78n05fyg00000gn/T/huelle-P5sEaV/zuhause/Library/Application Support/MagicSetEditor4Mac/pakete/probe-schlicht-kopie.mse-style/dunkel.png
- ✓ der Rechtsklick hat sie geöffnet — dunkel.png
- ✓ es nennt »Neue Datei« und »Umbenennen« — Neue Datei · +Bild oder Schrift hinzufügen … · ✎Umbenennen · ✕Löschen …
- ✓ einem Paket im Baum steht da — {"x":433,"y":271.5625}
- ✓ an einem Paket im Baum geht kein Menü der Seite auf

**hell: Rechtsklick an einer Partienzeile**

- ✓ einer Partienzeile steht da — {"x":1133.1796875,"y":218.1640625}
- ✓ das Menü an einer Partienzeile ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Partienzeile-hell.png`
- ✓ das Menü gehört der Partie — partie/118e905e6ad82888add84b0140192648
- ✓ es nennt Weiterspielen, Zuschauen und Duplizieren — Weiterspielen als Ada · Weiterspielen als Bo · Zuschauen (ohne Sitz) · ⧉Duplizieren · Ausführen · ✕Löschen · Marken
- ✓ ⌘Q beendet die Anwendung

**dunkel: der Fenstertitel ohne und mit offenem Set**

- ✓ ohne offenes Set heißt das Fenster wie die Anwendung — MagicSetEditor4Mac
- ✓ mit offenem Set nennt der Titel das Set vor der Anwendung — Werkstatt-Testkarten — MagicSetEditor4Mac

**dunkel: die Kürzel im Menü »Karte«**

- ✓ »Karte anlegen« trägt ⌘↩ — ["Karte anlegen","⌘↩"]
- ✓ »Karte duplizieren« trägt ⌘D — ["Karte duplizieren","⌘D"]
- ✓ »Karte löschen …« trägt ⌘⌫ — ["Karte löschen","⌘⌫"]
- ✓ »Vorige Karte« trägt ⇞ — ["Vorige Karte","⇞"]
- ✓ »Nächste Karte« trägt ⇟ — ["Nächste Karte","⇟"]
- Bild der Leiste: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/menue-karte-dunkel.png`

**dunkel: ⌘⌫ in einem Feld löscht Text, keine Karte**

- ✓ das Namensfeld steht da
- ✓ das Feld hat den Fokus und Text — "Langtext Probe"
- ✓ ⌘⌫ im Feld nimmt Text weg — "Langtext Probe" → ""
- ✓ ⌘⌫ im Feld lässt die Karten stehen — 5 → 5
- ✓ ⌘⌫ im Feld stellt keine Löschfrage

**dunkel: ⌘D, ⌘↩, ⇟, ⇞ und ⌘⌫ außerhalb der Felder**

- ✓ ⌘D verdoppelt die Karte — 5 → 6
- ✓ ⌘↩ legt eine Karte an — 6 → 7
- ✓ ⇞ blättert zur vorigen Karte — 7 → 6
- ✓ ⇟ blättert zur nächsten Karte — 6 → 7
- ✓ ⌘⌫ außerhalb löscht nicht ohne Rückfrage — 7 → 7
- ✓ ⌘⌫ außerhalb stellt die Löschfrage
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/loeschfrage-dunkel.png`

**dunkel: Rechtsklick an einer anderen Karte der Liste**

- ✓ einer anderen Karte der Liste steht da — {"x":433,"y":405.046875}
- ✓ das Menü an einer anderen Karte der Liste ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-anderen-Karte-der-Liste-dunkel.png`
- ✓ der Rechtsklick wählt die andere Karte — 1
- ✓ das Menü gehört ihr — karte/1
- ✓ es nennt Anlegen, Verdoppeln und Löschen — +Karte anlegen⌘↩ · ⧉Karte duplizieren⌘D · ⎘Kopieren (1)⌘C · Ausschneiden (1) · Einfügen · Vorige Karte⇞ · Nächste Karte⇟ · Nach oben · Nach unten · Bezug zu einem Druck … · Teilen … · ✕Karte löschen⌘⌫
- ✓ der untersten Zeile steht da — {"x":433,"y":410.546875}
- ✓ das Menü an der untersten Zeile ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-der-untersten-Zeile-dunkel.png`
- ✓ auch dort gehört das Menü der Karte darunter — karte/7

**dunkel: ⇧F10 an der Zeile, Escape gibt ihr den Fokus zurück**

- ✓ ⇧F10: der Fokus steht im Menü
- ✓ Escape gibt der Zeile den Fokus zurück — <div role="row" aria-rowindex="1" aria-selected="true" aria-label="Ohne Namen" t

**dunkel: ein Wechsel über die Leiste schließt ein offenes Menü**

- ✓ ein Objektwechsel schließt das Menü — 2

**dunkel: Rechtsklick an einer Setkarte**

- ✓ einer Setkarte steht da — {"x":439.2421875,"y":292.5625}
- ✓ das Menü an einer Setkarte ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Setkarte-dunkel.png`
- ✓ das Menü der Setkarte gehört dem Set — set/1
- ✓ es nennt Sichern und Löschen — ⤓Sichern · ⤓Sichern unter … · Als Umschlag ablegen · Über eine Exportvorlage … · ✕Löschen …

**dunkel: Rechtsklick an einer Paketzeile und einer Gruppe**

- ✓ einer Paketzeile steht da — {"x":821,"y":393.5546875}
- ✓ das Menü an einer Paketzeile ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Paketzeile-dunkel.png`
- ✓ das Menü der Paketzeile gehört ihrem Paket — probe → paket/mse-game/probe
- ✓ der Rechtsklick wählt das Paket — probe
- ✓ einer Paketgruppe steht da — {"x":821,"y":350.0078125}
- ✓ an einer Paketgruppe geht kein Menü der Seite auf

**dunkel: Rechtsklick an einer Vorlagendatei und am Paket im Baum**

- ✓ einer Vorlagendatei steht da — {"x":441,"y":339.3828125}
- ✓ das Menü an einer Vorlagendatei ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Vorlagendatei-dunkel.png`
- ✓ das Menü gehört der Datei — dunkel.png → datei//var/folders/8x/tzk9pq5s2g52zjr78n05fyg00000gn/T/huelle-PiB1a9/zuhause/Library/Application Support/MagicSetEditor4Mac/pakete/probe-schlicht-kopie.mse-style/dunkel.png
- ✓ der Rechtsklick hat sie geöffnet — dunkel.png
- ✓ es nennt »Neue Datei« und »Umbenennen« — Neue Datei · +Bild oder Schrift hinzufügen … · ✎Umbenennen · ✕Löschen …
- ✓ einem Paket im Baum steht da — {"x":433,"y":271.5625}
- ✓ an einem Paket im Baum geht kein Menü der Seite auf

**dunkel: Rechtsklick an einer Partienzeile**

- ✓ einer Partienzeile steht da — {"x":1133.1796875,"y":218.1640625}
- ✓ das Menü an einer Partienzeile ist ganz zu sehen
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/ap5kontext09/kontext-einer-Partienzeile-dunkel.png`
- ✓ das Menü gehört der Partie — partie/12dbc1cf64d3b8860b6121be88e53093
- ✓ es nennt Weiterspielen, Zuschauen und Duplizieren — Weiterspielen als Ada · Weiterspielen als Bo · Zuschauen (ohne Sitz) · ⧉Duplizieren · Ausführen · ✕Löschen · Marken
- ✓ ⌘Q beendet die Anwendung
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

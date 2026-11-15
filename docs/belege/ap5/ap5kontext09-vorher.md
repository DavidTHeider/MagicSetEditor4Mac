# Hüllenprobe: ap5kontext09

**19 Befund(e)** — Commit `da7059586` **mit 2 ungesicherten Änderung(en)**,
2026-10-16 22:07 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-16T22-05-33Z-4482`, Bauabdruck `38b37d68a5f4`.


**hell: der Fenstertitel ohne und mit offenem Set**

- ✓ ohne offenes Set heißt das Fenster wie die Anwendung — MagicSetEditor4Mac
- ✗ der Titel nennt das Set (nach 10000 ms) — null
- ✗ mit offenem Set nennt der Titel das Set vor der Anwendung — MagicSetEditor4Mac

**hell: die Kürzel im Menü »Karte«**

- ✗ »Karte anlegen« trägt ⌘↩ — {"Karte duplizieren":"","Kopieren (1)":"","Ausschneiden (1)":"","Einfügen":"","Nach oben":"","Nach unten":"","Bezug zu einem Druck …":"","Teilen …":"","Karte löschen":""}
- ✗ »Karte duplizieren« trägt ⌘D — ["Karte duplizieren",""]
- ✗ »Karte löschen …« trägt ⌘⌫ — ["Karte löschen",""]
- ✗ »Vorige Karte« trägt ⇞ — {"Karte duplizieren":"","Kopieren (1)":"","Ausschneiden (1)":"","Einfügen":"","Nach oben":"","Nach unten":"","Bezug zu einem Druck …":"","Teilen …":"","Karte löschen":""}
- ✗ »Nächste Karte« trägt ⇟ — {"Karte duplizieren":"","Kopieren (1)":"","Ausschneiden (1)":"","Einfügen":"","Nach oben":"","Nach unten":"","Bezug zu einem Druck …":"","Teilen …":"","Karte löschen":""}
- Bild der Leiste: `/Users/david/mse-arbeit/w31v/.huellenprobe/laeufe/2026-10-16T22-05-33Z-4482/bilder/ap5kontext09/menue-karte-hell.png`

**hell: ⌘⌫ in einem Feld löscht Text, keine Karte**

- ✓ das Namensfeld steht da
- ✓ das Feld hat den Fokus und Text — "Langtext Probe"
- ✓ ⌘⌫ im Feld nimmt Text weg — "Langtext Probe" → ""
- ✓ ⌘⌫ im Feld lässt die Karten stehen — 5 → 5
- ✓ ⌘⌫ im Feld stellt keine Löschfrage

**hell: ⌘D, ⌘↩, ⇟, ⇞ und ⌘⌫ außerhalb der Felder**

- ✗ ⌘D verdoppelt die Karte (nach 15000 ms) — false
- ✗ ⌘D verdoppelt die Karte — 5 → 5
- ✗ ⌘↩ legt eine Karte an (nach 15000 ms) — false
- ✗ ⌘↩ legt eine Karte an — 6 → 5
- ✗ ⇞ wählt eine andere Karte (nach 10000 ms) — null
- ✗ ⇞ blättert zur vorigen Karte — 1 → null
- ✗ ⇟ wählt wieder die von vorhin (nach 10000 ms) — null
- ✗ ⇟ blättert zur nächsten Karte — null → 2
- ✓ ⌘⌫ außerhalb löscht nicht ohne Rückfrage — 5 → 5
- ✗ die Löschfrage steht (nach 10000 ms) — false
- ✗ ⌘⌫ außerhalb stellt die Löschfrage
- Bild: `/Users/david/mse-arbeit/w31v/.huellenprobe/laeufe/2026-10-16T22-05-33Z-4482/bilder/ap5kontext09/loeschfrage-hell.png`

**hell: Rechtsklick an einer anderen Karte der Liste**

- ✓ einer anderen Karte der Liste steht da — {"x":433,"y":404}
- ✗ der Rechtsklick an einer anderen Karte der Liste öffnet ein Menü (nach 10000 ms) — false
- ✗ das Drehbuch lief durch — abgebrochen: der Rechtsklick an einer anderen Karte der Liste öffnet ein Menü
- Bild: `/Users/david/mse-arbeit/w31v/.huellenprobe/laeufe/2026-10-16T22-05-33Z-4482/bilder/ap5kontext09/abbruch.png`
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

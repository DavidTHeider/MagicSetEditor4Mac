# Hüllenprobe: uxrf15

**bestanden** — Commit `1c0b74221`,
2026-10-09 22:18 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.


**1 — kein Textfeld, sondern ein Knopf und ein Satz**

- ✓ der Knopf heißt „Ort wählen …" — Ort wählen …
- ✓ daneben steht „Noch kein Ort gewählt." — Noch kein Ort gewählt.
- ✓ es gibt kein Textfeld für einen Pfad
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf15/ohne-wahl.png`

**2 — der Sichern-Dialog von macOS: Name und Endung, und ein Abbruch ändert nichts**

- ✓ »Ort wählen …« öffnet den Dialog von macOS — Umschlag der Leihe ablegen, Vorschlag »leihe.msepackage«
- ✓ der Dialog schlägt Namen und Endung vor — leihe.msepackage
- ✓ und sagt, was er fragt — Umschlag der Leihe ablegen
- ✓ nach dem Abbruch steht weiter „Noch kein Ort gewählt." — Noch kein Ort gewählt.
- ✓ und keine Meldung
- ✓ der Knopf ist wieder frei

**3 — der gewählte Pfad steht lesbar daneben**

- ✓ der Pfad steht im Ganzen da — /var/folders/8x/tzk9pq5s2g52zjr78n05fyg00000gn/T/uxrf15-dcx94y/leihe-fuer-bert.msepackage
- ✓ der Knopf heißt jetzt „Anderen Ort wählen …" — Anderen Ort wählen …
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf15/ziel-gewaehlt.png`

**4 — ein Abbruch lässt den gewählten Pfad stehen**

- ✓ »Anderen Ort wählen …« öffnet den Dialog von macOS — Umschlag der Leihe ablegen, Vorschlag »leihe.msepackage«
- ✓ der Pfad ist derselbe — /var/folders/8x/tzk9pq5s2g52zjr78n05fyg00000gn/T/uxrf15-dcx94y/leihe-fuer-bert.msepackage
- ✓ und der Knopf frei

**5 — eine zweite Wahl ändert ihn**

- ✓ der neue Pfad ersetzt den alten — /var/folders/8x/tzk9pq5s2g52zjr78n05fyg00000gn/T/uxrf15-dcx94y/zweite-leihe.msepackage

**6 — der Umschlag geht dorthin, wo der Pfad steht**

- vor dem Verleihen: {"auswahl":["Deck=","Deck=","Bert · 7678-c377-742=bafzaajaiaejcb2ua6atsbcrtq6mdj6izp6pivqmwr6qlif4ofegzyfqfwyovxrue","Welches Deck=1"],"knopf":false}
- ✓ der Umschlag liegt am gewählten Ort
- ✓ und nicht am ersten
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf15/abgelegt.png`

**7 — der Umschlag, der angenommen wird, kommt aus dem Öffnen-Dialog**

- ✓ der Knopf heißt „Umschlag wählen …" — Umschlag wählen …
- ✓ daneben steht „Noch kein Umschlag gewählt." — Noch kein Umschlag gewählt.
- ✓ „Leihe annehmen" ist angehalten
- ✓ »Umschlag wählen …« öffnet den Dialog von macOS
- ✓ nach dem Abbruch steht weiter „Noch kein Umschlag gewählt."
- ✓ der Pfad steht im Ganzen da — /var/folders/8x/tzk9pq5s2g52zjr78n05fyg00000gn/T/uxrf15-dcx94y/zweite-leihe.msepackage
- ✓ der Knopf heißt jetzt „Anderen Umschlag wählen …" — Anderen Umschlag wählen …
- ✓ „Leihe annehmen" ist frei
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/uxrf15/umschlag-gewaehlt.png`

**8 — ein Abbruch lässt auch den Umschlag stehen**

- ✓ »Anderen Umschlag wählen …« öffnet den Dialog von macOS
- ✓ der Pfad ist derselbe — /var/folders/8x/tzk9pq5s2g52zjr78n05fyg00000gn/T/uxrf15-dcx94y/zweite-leihe.msepackage
- ✓ und keine Meldung
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

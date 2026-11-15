# Hüllenprobe: netzpartie09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:19 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**Zwei Anwendungen, einander bekannt — ungezählt, wie in den Ausgangswerten**

- ✓ Gast ist bestätigt — ""
- ✓ Gast darf eine Partie spielen — ""
- ✓ Wirt ist bestätigt — ""
- ✓ Wirt darf eine Partie spielen — ""
- ✓ der Wirt horcht — 127.0.0.1:56155

**S12 — der Wirt lädt ein: »Gast einladen« auf der Deckseite**

- ✓ S12 (bereich, maus): 2 Handlungen, 0 Fehlversuche, 1,0 s Maschinenzeit
- ✓ S12 (bereich, maus): Wirkung beobachtet
- ✓ der Tisch steht offen, ohne »Tisch öffnen«
- ✓ und sagt, wer erwartet wird
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/netzpartie09/s12-erwartet.png`

**Ohne Deck vor dem Anfang abgewiesen — mit Grund**

- ✓ ohne Deck: der Grund des Wirts, kein Tisch — []
- ✓ und er steht als Absage, nicht als Abriss
- ✓ und kein Tisch steht mehr da
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/netzpartie09-gast/ohne-deck-abgewiesen.png`

**S13 — der Gast setzt sich mit seinem Deck: »Bei Wirt mitspielen«**

- ✓ S13 (bereich, maus): 2 Handlungen, 0 Fehlversuche, 1,0 s Maschinenzeit
- ✓ S13 (bereich, maus): Wirkung beobachtet

**S13 — die Eröffnung des Gastes**

- ✓ S13 (eroeffnung, maus): 2 Handlungen, 0 Fehlversuche, 0,9 s Maschinenzeit
- ✓ S13 (eroeffnung, maus): Wirkung beobachtet
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/netzpartie09-gast/s13-gast-mit-eigenem-deck.png`

**Das Gastdeck: beim Gast die eigenen Zeilen, beim Wirt keine**

- ✓ Bos Hand liegt beim Wirt ohne Zeile, mit Bildkennung — [["",true],["",true],["",true],["",true],["",true],["",true],["",true]]
- ✓ jede Kennung in der Sicht des Gastes ist eine Sichtkennung — ["s-P_5_3E5ppOeEl2EYZu","s-EI5wy86iT7MbB6IdmW","s-xZzuNxgcihNLLWgA7-","s-WFNrIlQWym9EhYOtR8"]
- ✓ der Gast zeichnet seine Hand selbst — 2

**Der Wirt behält seine Hand — erst danach ist Spiel (CR 103.8, Fassung 2026-09-25)**

- ✓ beim Wirt steht keine Eröffnung mehr zur Wahl

**Kartenzug: eine Karte des Gastes auf den Tisch — der Wirt sieht sie mit dem Bild des Gastes**

- ✓ der Wunsch geht hinaus — ""
- Karte des Gastes: beim Wirt vom Gast 37549:b183f6b6, beim Gast gezeichnet 37549:b183f6b6 / 41854:c9565df
- ✓ der Wirt zeigt byteweise ein Bild, das der Gast zeichnet — 37549:b183f6b6 · 37549:b183f6b6 / 41854:c9565df
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/netzpartie09/wirt-sieht-gastkarte.png`

**Ein abgewiesener Wunsch beendet die Sitzung nicht (#2478)**

- ✓ der Wunsch geht hinaus — ""
- ✓ der Grund des Kerns steht am Tisch
- ✓ und der Gast sitzt weiter
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/netzpartie09-gast/wunsch-abgewiesen.png`

**Abbruch und Wiederverbindung — derselbe Sitz, dieselben Karten**

- ✓ nach dem Wiederverbinden dieselben Karten am selben Sitz — 7 von 7
- ✓ und keine Fehlermeldung am Tisch — []
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/netzpartie09-gast/wiederverbunden.png`
- Datensatz: /Users/david/mse-arbeit/p13/docs/belege/netzpartie09-wege.json
- ✓ ⌘Q beendet den Gast
- ✓ ⌘Q beendet den Wirt
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

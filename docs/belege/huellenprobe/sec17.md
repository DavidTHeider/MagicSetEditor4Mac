# Hüllenprobe: sec17

**bestanden** — Commit `1c0b74221`,
2026-10-09 22:10 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.

- ✓ die Partie ist angelegt — Partie aa0e74105d1af765283cc00159965045 (laeuft), Fassung 1, Sicht für ada.

**1 — beobachtend: die Policy meldet, hält aber nichts auf**

- beobachtend: {"inline":"lief","eval":"erlaubt","verstoesse":["script-src-elem inline","script-src-elem https://sec17.invalid/fremd.js"]}
- ✓ beobachtend läuft das künstliche Skript
- ✓ beobachtend: das Kartenbild ist geladen — 375 px breit
- ✓ beobachtend: am Tisch steht keine Fehlermeldung — []
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/sec17/sec17-tisch-beobachtend.png`
- beobachtend gemeldet (2): time=2026-10-10T00:10:37.397+02:00 level=WARN msg=CSP-Verstoß anweisung=script-src-elem abgewiesen=inline stelle=index.html:776 | time=2026-10-10T00:10:37.397+02:00 level=WARN msg=CSP-Verstoß anweisung=script-src-elem abgewiesen=https://sec17.invalid stelle=index.html:781
- ✓ ⌘Q beendet

**2 — durchgesetzt, hell**

- hell: {"inline":"abgewiesen","eval":"abgewiesen","verstoesse":["script-src-elem inline","script-src eval","script-src-elem https://sec17.invalid/fremd.js"]}
- ✓ hell: das künstliche Inline-Skript ist abgewiesen
- ✓ hell: eval ist abgewiesen
- ✓ hell: das fremde Skript ist gemeldet — script-src-elem inline · script-src eval · script-src-elem https://sec17.invalid/fremd.js
- ✓ hell: das Kartenbild ist geladen — 375 px breit
- ✓ hell: am Tisch steht keine Fehlermeldung — []
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/sec17/sec17-tisch-hell.png`
- hell: das Protokoll nennt 3 Verstöße
- ✓ hell: das Protokoll nennt die künstlichen Verstöße — 3
- ✓ hell: und keinen anderen
- ✓ ⌘Q beendet

**3 — durchgesetzt, dunkel**

- dunkel: {"inline":"abgewiesen","eval":"abgewiesen","verstoesse":["script-src-elem inline","script-src eval","script-src-elem https://sec17.invalid/fremd.js"]}
- ✓ dunkel: das künstliche Inline-Skript ist abgewiesen
- ✓ dunkel: eval ist abgewiesen
- ✓ dunkel: das fremde Skript ist gemeldet — script-src-elem inline · script-src eval · script-src-elem https://sec17.invalid/fremd.js
- ✓ dunkel: das Kartenbild ist geladen — 375 px breit
- ✓ dunkel: am Tisch steht keine Fehlermeldung — []
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/sec17/sec17-tisch-dunkel.png`
- dunkel: das Protokoll nennt 3 Verstöße
- ✓ dunkel: das Protokoll nennt die künstlichen Verstöße — 3
- ✓ dunkel: und keinen anderen
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

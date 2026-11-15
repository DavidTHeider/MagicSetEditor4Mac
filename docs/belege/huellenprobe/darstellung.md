# Hüllenprobe: darstellung

**bestanden** — Commit `1c0b74221`,
2026-10-09 19:52 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.

- breit: {"X":15,"Y":85,"B":1440,"H":838}

**breit — 1: Titel und Aktivierung stimmen mit dem Bild überein**

- Leiste: {"seitenleiste":{"titel":"Seitenleiste ausblenden","aktiv":true},"navigator":{"titel":"Navigator ausblenden","aktiv":true},"inspector":{"titel":"Inspector ausblenden","aktiv":true}}
- ✓ alle drei Einträge sind wählbar
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/darstellung/darstellung-breit.png`

**breit — 2: »Navigator ausblenden« klappt, und die Leiste folgt**

- ✓ »Navigator ausblenden« klappt (navigator: true → false), und der Eintrag heißt jetzt »Navigator einblenden«
- ✓ derselbe Eintrag klappt navigator zurück

**breit — 3: ⌥⌘[ klappt genau einmal**

- NICHT GEPRÜFT: ⌥⌘[ — AppKit zeigt es auf dieser Belegung als ⌥⌘Ö; von Hand prüfen

**breit — 2: »Inspector ausblenden« klappt, und die Leiste folgt**

- ✓ »Inspector ausblenden« klappt (inspector: true → false), und der Eintrag heißt jetzt »Inspector einblenden«
- ✓ derselbe Eintrag klappt inspector zurück

**breit — 3: ⌥⌘] klappt genau einmal**

- NICHT GEPRÜFT: ⌥⌘] — AppKit zeigt es auf dieser Belegung als ⌥⌘Ä; von Hand prüfen

**breit — 2: »Seitenleiste ausblenden« klappt, und die Leiste folgt**

- ✓ »Seitenleiste ausblenden« klappt (seitenleiste: true → false), und der Eintrag heißt jetzt »Seitenleiste einblenden«
- ✓ derselbe Eintrag klappt seitenleiste zurück

**breit — 3: ⌥⌘S klappt genau einmal**

- ✓ ⌥⌘S klappt seitenleiste einmal und nicht zweimal — false → false

**breit — 4: auf einer Seite ohne Arbeitsplatz sind Navigator und Inspector grau**

- ✓ Navigator und Inspector sind grau
- ✓ die Seitenleiste bleibt wählbar
- ✓ ein Klick auf den grauen Eintrag wird abgewiesen — klick: kein solcher Eintrag: »Navigator einblenden« ist nicht aktiv
- ✓ derselbe Stand wie vorher — {"seitenleiste":true,"navigator":true,"inspector":true} → {"seitenleiste":true,"navigator":true,"inspector":true}

**breit — 5: unter einer Schublade sind beide grau und nennen die Spalte darunter**

- ✓ unter der Schublade sind Navigator und Inspector grau
- ✓ und sie sagen weiter, wie die Spalte darunter steht — Navigator ausblenden · Inspector ausblenden
- ✓ ohne Schublade sind sie wieder wählbar
- breit — am Ende stimmen Leiste und Bild überein: Seite {"seitenleiste":true,"navigator":true,"inspector":true}, Leiste {"seitenleiste":{"titel":"Seitenleiste ausblenden","aktiv":true},"navigator":{"titel":"Navigator ausblenden","aktiv":true},"inspector":{"titel":"Inspector ausblenden","aktiv":true}}
- ✓ breit — am Ende stimmen Leiste und Bild überein
- ✓ ⌘Q beendet die Anwendung (breit)
- eng: {"X":223,"Y":168,"B":1024,"H":700}

**eng — 1: Titel und Aktivierung stimmen mit dem Bild überein**

- Leiste: {"seitenleiste":{"titel":"Seitenleiste ausblenden","aktiv":true},"navigator":{"titel":"Navigator einblenden","aktiv":true},"inspector":{"titel":"Inspector ausblenden","aktiv":true}}
- ✓ alle drei Einträge sind wählbar
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/darstellung/darstellung-eng.png`

**eng — 2: »Navigator einblenden« klappt, und die Leiste folgt**

- ✓ »Navigator einblenden« klappt (navigator: false → true), und der Eintrag heißt jetzt »Navigator ausblenden«
- ✓ derselbe Eintrag klappt navigator zurück

**eng — 3: ⌥⌘[ klappt genau einmal**

- NICHT GEPRÜFT: ⌥⌘[ — AppKit zeigt es auf dieser Belegung als ⌥⌘Ö; von Hand prüfen

**eng — 2: »Inspector ausblenden« klappt, und die Leiste folgt**

- ✓ »Inspector ausblenden« klappt (inspector: true → false), und der Eintrag heißt jetzt »Inspector einblenden«
- ✓ derselbe Eintrag klappt inspector zurück

**eng — 3: ⌥⌘] klappt genau einmal**

- NICHT GEPRÜFT: ⌥⌘] — AppKit zeigt es auf dieser Belegung als ⌥⌘Ä; von Hand prüfen

**eng — 2: »Seitenleiste ausblenden« klappt, und die Leiste folgt**

- ✓ »Seitenleiste ausblenden« klappt (seitenleiste: true → false), und der Eintrag heißt jetzt »Seitenleiste einblenden«
- ✓ derselbe Eintrag klappt seitenleiste zurück

**eng — 3: ⌥⌘S klappt genau einmal**

- ✓ ⌥⌘S klappt seitenleiste einmal und nicht zweimal — false → false

**eng — 6: der Navigator weicht über das Zentrum, und Escape schließt ihn**

- ✓ der Navigator steht bei Enge nicht neben dem Zentrum — false
- ✓ der Eintrag öffnet ihn über dem Zentrum
- ✓ Escape schließt ihn, und der Eintrag heißt wieder »einblenden«

**eng — 4: auf einer Seite ohne Arbeitsplatz sind Navigator und Inspector grau**

- ✓ Navigator und Inspector sind grau
- ✓ die Seitenleiste bleibt wählbar
- ✓ ein Klick auf den grauen Eintrag wird abgewiesen — klick: kein solcher Eintrag: »Navigator einblenden« ist nicht aktiv
- ✓ derselbe Stand wie vorher — {"seitenleiste":true,"navigator":false,"inspector":true} → {"seitenleiste":true,"navigator":false,"inspector":true}
- eng — am Ende stimmen Leiste und Bild überein: Seite {"seitenleiste":true,"navigator":false,"inspector":true}, Leiste {"seitenleiste":{"titel":"Seitenleiste ausblenden","aktiv":true},"navigator":{"titel":"Navigator einblenden","aktiv":true},"inspector":{"titel":"Inspector ausblenden","aktiv":true}}
- ✓ eng — am Ende stimmen Leiste und Bild überein
- ✓ ⌘Q beendet die Anwendung (eng)
- englisch: {"X":15,"Y":82,"B":1440,"H":841}

**englisch — 1: Titel und Aktivierung stimmen mit dem Bild überein**

- Leiste: {"seitenleiste":{"titel":"Hide Sidebar","aktiv":true},"navigator":{"titel":"Hide Navigator","aktiv":true},"inspector":{"titel":"Hide Inspector","aktiv":true}}
- ✓ alle drei Einträge sind wählbar
- Bild: `/Users/david/MagicSetEditor4Mac/.huellenprobe/laeufe/2026-10-09T19-47-35Z-46912/bilder/darstellung/darstellung-englisch.png`

**englisch — 2: »Hide Navigator« klappt, und die Leiste folgt**

- ✓ »Hide Navigator« klappt (navigator: true → false), und der Eintrag heißt jetzt »Show Navigator«
- ✓ derselbe Eintrag klappt navigator zurück

**englisch — 3: ⌥⌘[ klappt genau einmal**

- NICHT GEPRÜFT: ⌥⌘[ — AppKit zeigt es auf dieser Belegung als ⌥⌘Ö; von Hand prüfen

**englisch — 2: »Hide Inspector« klappt, und die Leiste folgt**

- ✓ »Hide Inspector« klappt (inspector: true → false), und der Eintrag heißt jetzt »Show Inspector«
- ✓ derselbe Eintrag klappt inspector zurück

**englisch — 3: ⌥⌘] klappt genau einmal**

- NICHT GEPRÜFT: ⌥⌘] — AppKit zeigt es auf dieser Belegung als ⌥⌘Ä; von Hand prüfen

**englisch — 2: »Hide Sidebar« klappt, und die Leiste folgt**

- ✓ »Hide Sidebar« klappt (seitenleiste: true → false), und der Eintrag heißt jetzt »Show Sidebar«
- ✓ derselbe Eintrag klappt seitenleiste zurück

**englisch — 3: ⌥⌘S klappt genau einmal**

- ✓ ⌥⌘S klappt seitenleiste einmal und nicht zweimal — false → false

**englisch — 4: auf einer Seite ohne Arbeitsplatz sind Navigator und Inspector grau**

- ✓ Navigator und Inspector sind grau
- ✓ die Seitenleiste bleibt wählbar
- ✓ ein Klick auf den grauen Eintrag wird abgewiesen — klick: kein solcher Eintrag: »Show Navigator« ist nicht aktiv
- ✓ derselbe Stand wie vorher — {"seitenleiste":true,"navigator":true,"inspector":true} → {"seitenleiste":true,"navigator":true,"inspector":true}
- englisch — am Ende stimmen Leiste und Bild überein: Seite {"seitenleiste":true,"navigator":true,"inspector":true}, Leiste {"seitenleiste":{"titel":"Hide Sidebar","aktiv":true},"navigator":{"titel":"Hide Navigator","aktiv":true},"inspector":{"titel":"Hide Inspector","aktiv":true}}
- ✓ englisch — am Ende stimmen Leiste und Bild überein
- ✓ ⌘Q beendet die Anwendung (englisch)
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

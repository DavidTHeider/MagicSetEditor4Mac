# Hüllenprobe: netzbild09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:19 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.

- ✓ die Partie ist angelegt — Partie e79859cb2078db740876512cea171dc3 (laeuft), Fassung 1, Sicht für ada.

**Zwei Anwendungen, zwei Bestände, dieselben Zeilen — andere Karten**

- ✓ Gast ist bestätigt — ""
- ✓ Gast darf eine Partie spielen — ""
- ✓ Wirt ist bestätigt — ""
- ✓ Wirt darf eine Partie spielen — ""
- ✓ der Wirt horcht — 127.0.0.1:56143

**Der Wirt öffnet seinen Tisch**


**Der Gast setzt sich**


**Der Gast zeigt die Bilder des Wirts — und fragt den eigenen Bestand nach keiner Zeile**

- Bos Hand beim Wirt: 7 Karten, 2 Bilder, Zeilen ["2","2","2","2","2","2","4"]
- Rufbuch des Gastes: Fremdkartenbild {"anzahl":2,"gelungen":2}, Spieltextur {"anzahl":0}, Rueckentextur {"anzahl":2,"gelungen":2}
- ✓ der Gast hat jedes Bild seiner Hand beim Wirt erfragt, je einmal — {"anzahl":2,"gelungen":2} für 2 Bilder
- ✓ und keine Zeile im eigenen Bestand nachgeschlagen — {"anzahl":0}
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/netzbild09-gast/netzbild09-gast-sieht-die-bilder.png`

**Byteweise dasselbe Bild — und der eigene Bestand hätte ein anderes gezeigt**

- ✓ in Bos Hand liegt ein Bergpfad mit Bildkennung — ["2","2","2","2","2","2","4"]
- Zeile 2: beim Wirt 32467:7ae3b009, beim Gast vom Wirt 32467:7ae3b009, beim Gast aus dem eigenen Bestand 41854:c9565df
- ✓ das Bild am Gast ist das Bild des Wirts — 32467:7ae3b009 · 32467:7ae3b009
- ✓ dieselbe Zeile im eigenen Bestand wäre eine andere Karte — 41854:c9565df · 32467:7ae3b009

**Wer die Bildkennung einer verdeckten Karte kennt, bekommt trotzdem keine**

- Antwort auf Adas Handkarte: {"adresse":"","befunde":["dieses Objekt ist für dich nicht zu sehen"],"fehler":null}
- ✓ der Wirt zeichnet Adas Hand nicht für Bo — {"textur":{"schluessel":"","adresse":"","guete":"tisch","seite":"vorne","breite":0,"hoehe":0,"befunde":["dieses Objekt ist für dich nicht zu sehen"],"zweiseitig":false},"fehler":null}
- ✓ am fremden Tisch steht keine Fehlermeldung — []
- ✓ ⌘Q beendet den Gast
- ✓ ⌘Q beendet den Wirt
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

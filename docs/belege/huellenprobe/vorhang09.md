# Hüllenprobe: vorhang09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:30 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.


**S1 — aus dem Bereich, am neuen Bündel**

- ✓ S1 (bereich, maus): 3 Handlungen, 0 Fehlversuche, 1,4 s Maschinenzeit
- ✓ S1 (bereich, maus): Wirkung beobachtet

**S2 — bis Zug 1: behalten, der Vorhang, behalten**

- ✓ hinter dem Vorhang kein Tisch und keine Karte — {"tisch":false,"namen":[],"sicht":"sitz1","laedt":false,"vorhang":true}
- ✓ der Vorhang wechselt die Sicht nicht von selbst — ["sitz1","sitz1","sitz1","sitz1","sitz1","sitz1","sitz1","sitz1","sitz1","sitz1","sitz1","sitz1"]
- ✓ der Fokus steht auf »Ich bin …«, Esc ist angesagt — {"fokus":"Ich bin Spieler 2","esc":"Weiter als Mein Profil"}
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/vorhang09/vorhang.png`
- ✓ axe in WebKit: keine Verstöße gegen WCAG 2.2 AA am Vorhang
- ✓ S2 (tisch, maus): 3 Handlungen, 0 Fehlversuche, 2,5 s Maschinenzeit
- ✓ S2 (tisch, maus): Wirkung beobachtet

**Esc lässt den bisherigen Sitz weiterhandeln**

- ✓ nach Esc steht der Tisch des bisherigen Sitzes — {"tisch":true,"namen":["Bergpfad","Glutläufer","Rindenwall"],"sicht":"sitz2","laedt":false,"vorhang":false}

**↩ und verzögertes Nachladen — kein Bild mit dem Tisch des vorigen Sitzes**

- ✓ der Fokus steht auf »Ich bin …«
- ✓ während des Ladens zeigt kein Bild den Tisch des vorigen Sitzes — 31 Bilder, 30 mit »wird geladen«, 0 mit fremdem Tisch
- ✓ das Laden war sichtbar verzögert — 30
- ✓ danach steht der neue Sitz — {"tisch":true,"namen":["Bergpfad","Glutläufer","Rindenwall"],"sicht":"sitz1","laedt":false,"vorhang":false}
- ✓ der Fokus steht danach an der Überschrift des Tisches

**Gescheitertes Nachladen — die Meldung und kein Tisch**

- ✓ nach einem gescheiterten Laden kein Tisch und keine Karte — {"tisch":false,"namen":[],"sicht":"sitz1","laedt":false,"vorhang":false}
- ✓ die Meldung steht da

**Der Zuschauer übernimmt keinen Sitz**

- ✓ das Fenster hat den Wechsel der Priorität gelesen — sitz2 → sitz1, Fassung 9
- ✓ der Zuschauer bleibt Zuschauer, ohne Vorhang — {"tisch":true,"namen":[],"sicht":"","laedt":false,"vorhang":false}

**„Offene Hände“ folgt ohne Vorhang**

- ✓ die Sicht folgt ohne Vorhang — 2 Bilder, zuletzt {"tisch":true,"namen":["Bergpfad","Glutläufer","Rindenwall"],"sicht":"sitz2","laedt":false,"vorhang":false}

**Ein Profilwechsel lässt keinen Tisch des vorigen Profils stehen**

- ✓ vor dem Wechsel steht ein Tisch mit Karten — {"tisch":true,"namen":["Bergpfad","Glutläufer","Rindenwall"],"sicht":"sitz2","laedt":false,"vorhang":false}
- ✓ das Profil ist gewechselt — 2	Zweitprofil
- ✓ nach dem Profilwechsel kein Tisch und keine Karte — {"tisch":false,"namen":[],"sicht":null,"laedt":false,"vorhang":false}
- ✓ ⌘Q beendet die Anwendung
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

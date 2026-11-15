# Hüllenprobe: tischhand09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:29 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.

- ✓ die Partie steht in Idas erster Hauptphase
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/tischhand09/tisch-vorher.png`

**S3 — ein Klick spielt ein Land, das zweite wird mit Grund abgewiesen**

- ✓ S3 (tisch-klick, maus): 1 Handlungen, 0 Fehlversuche, 0,7 s Maschinenzeit
- ✓ S3 (tisch-klick, maus): Wirkung beobachtet
- ✓ das zweite Land ist abgewiesen, mit Grund (CR 305.2) — Das ging nicht: in diesem Zug ist schon ein Land gespielt (CR 305.2). Berichtige die Angabe und versuche es noch einmal.
- gestellt: 2.60 objekt.land {"objekt":"s-636wDRZxpkKA8wGmwK"} @9 · 3.29 objekt.land {"objekt":"s-kts7kP2js3QNsImh-m"} @10
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/tischhand09/zweites-land.png`

**S5 — ein Klick tappt, ein Doppelklick tappt einmal für Mana**

- ✓ S5 (tisch-klick, maus): 1 Handlungen, 0 Fehlversuche, 0,6 s Maschinenzeit
- ✓ S5 (tisch-klick, maus): Wirkung beobachtet
- ✓ ein zweiter Klick enttappt
- ✓ der Doppelklick kam als einer an (detail 1, 2) — [1,2]
- ✓ der Doppelklick tappt einmal für Mana und nicht zweimal — getappt: true, Vorrat: »G«

**T an der Karte, ⌘-Klick wählt**

- ✓ die Karte nennt ihre Einzeltasten — T Backspace E
- ✓ T stellt die Primärhandlung
- ✓ ⌘-Klick wählt und stellt nichts

**Die gewählte Karte in der Menüleiste**

- ✓ die Menüleiste führt dieselben Einträge wie der Rechtsklick — Leiste: Exilieren · Auf den Friedhof · Weitere Handlungen … | Menü: Exilieren · Auf den Friedhof · Weitere Handlungen …

**S6 — Rechtsklick, „Auf den Friedhof“**

- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/tischhand09/kartenmenue.png`
- ✓ S6 (tisch-menue, maus): 2 Handlungen, 0 Fehlversuche, 1,3 s Maschinenzeit
- ✓ S6 (tisch-menue, maus): Wirkung beobachtet
- ✓ ⇧F10: das Menü ist offen, der Fokus darin

**S7 — Leben: dreimal −, dann ⇧ +, dann „Wert setzen“**

- ✓ S7 (tisch-regler, maus): 3 Handlungen, 0 Fehlversuche, 1,5 s Maschinenzeit
- ✓ S7 (tisch-regler, maus): Wirkung beobachtet
- ✓ ⇧-Klick auf + gibt fünf — 22
- ✓ „Wert setzen“ setzt 17 — 17
- ✓ axe in WebKit: keine Verstöße gegen WCAG 2.2 AA am Tisch

**Drei schnelle Gesten bei verzögerter Antwort**

- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/tischhand09/drei-gesten-wartend.png`
- ✓ das Warten steht am Tisch, höchstens drei — Wartet auf den Tisch: 2
- ✓ drei Gesten, drei Wirkungen, keine still verloren — []
- gestellt: 2.60 objekt.land {"objekt":"s-636wDRZxpkKA8wGmwK"} @9 · 3.29 objekt.land {"objekt":"s-kts7kP2js3QNsImh-m"} @10 · 3.93 objekt.tappen {"objekt":"s-636wDRZxpkKA8wGmwK"} @10 · 4.28 objekt.enttappen {"objekt":"s-636wDRZxpkKA8wGmwK"} @11 · 4.45 objekt.mana {"objekt":"s-636wDRZxpkKA8wGmwK"} @12 · 5.82 objekt.enttappen {"objekt":"s-636wDRZxpkKA8wGmwK"} @13 · 7.37 objekt.ziehen {"objekt":"s-JiIDMXF7BFNoL6TXbc","zone":"friedhof","sitz":"ada"} @14 · 8.17 spieler.leben {"um":-1} @15 · 8.64 spieler.leben {"um":-1} @16 · 9.11 spieler.leben {"um":-1} @17 · 9.66 spieler.leben {"um":5} @18 · 10.94 spieler.leben {"um":-5} @19 · 11.44 objekt.ziehen {"objekt":"s-UcO1FHGRvECn2Werky","zone":"friedhof","sitz":"ada"} @20 · 12.35 objekt.ziehen {"objekt":"s-WQue-txhzphz0zAeAq","zone":"friedhof","sitz":"ada"} @21 · 13.26 objekt.ziehen {"objekt":"s-m2SvSpjkwxizYT16ZM","zone":"friedhof","sitz":"ada"} @22

**Bildabstand in Ruhe und beim Ziehen**

- in Ruhe: {"n":118,"median":17,"p99":19,"max":19,"lang":[]}
- ✓ Bildabstand beim Ziehen: Median ≤ 17 ms, p99 ≤ 20 ms (123 Bilder) — {"n":123,"median":17,"p99":19,"max":19,"lang":[]}

**Nur mit der Tastatur, an der Liste**

- ✓ S3 (liste, tastatur): 5 Handlungen, 0 Fehlversuche, 1,9 s Maschinenzeit
- ✓ S3 (liste, tastatur): Wirkung beobachtet
- ✓ S5 (liste, tastatur): 15 Handlungen, 0 Fehlversuche, 5,7 s Maschinenzeit
- ✓ S5 (liste, tastatur): Wirkung beobachtet
- ✓ S6 (liste-menue, tastatur): 78 Handlungen, 0 Fehlversuche, 28,6 s Maschinenzeit
- ✓ S6 (liste-menue, tastatur): Wirkung beobachtet
- ✓ S7 (liste, tastatur): 36 Handlungen, 0 Fehlversuche, 13,4 s Maschinenzeit
- ✓ S7 (liste, tastatur): Wirkung beobachtet

**Die Geste: Leertaste, Pfeil, Eingabe — und Escape**

- ✓ Escape bricht ab, der Glutläufer bleibt, wo er war — Abgebrochen. Nichts wurde bewegt. — erreicht: true, vorher: hand, nachher: hand
- ✓ S6 (liste-geste, tastatur): 88 Handlungen, 0 Fehlversuche, 32,9 s Maschinenzeit
- ✓ S6 (liste-geste, tastatur): Wirkung beobachtet

**Aufgeben mit der Tastatur — und das Ergebnis statt einer Aufforderung**

- ✓ die beendete Partie zeigt ihr Ergebnis
- ✓ und keine Starthand- oder Prioritätsaufforderung — [false,false,false]
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/tischhand09/beendet.png`
- ✓ ⌘Q beendet die Anwendung
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

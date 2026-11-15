# Hüllenprobe: fix-f03

**bestanden** — Commit `1c0b74221`,
2026-10-09 21:59 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-09T19-47-35Z-46912`, Bauabdruck `1b6e1cccf72a`.


**S03.NAT-01 — ein künstliches CSV mit neuer Kennung und Menge 3**

- Bericht: 1 neu, 0 erhöht, 3 Stück aufgenommen.
- ✓ genau drei Stück des neuen Drucks — 5d7752c0-5c9b-4fcc-a185-85cac7096f3f,3,nearMint,en,nonfoil,false,,,,,,,,,false,n

**NAT-02 — eine Dublette erhöht, statt einen zweiten Eintrag anzulegen**

- Bericht: 0 neu, 1 erhöht, 2 Stück aufgenommen.
- ✓ ein Eintrag mit fünf Stück — 5d7752c0-5c9b-4fcc-a185-85cac7096f3f,5,nearMint,en,nonfoil,false,,,,,,,,,false,normal,genau,1,scryfall,5d7752c0-5c9b-4fc

**S03.NAT-03 — eine kaputte Zeile gibt einen fachlichen Bericht**

- ✓ der Bericht nennt die Zeile — 0 neu, 0 erhöht, 0 Stück aufgenommen. | Zeile 2: "drei" ist keine Stückzahl
- ✓ ohne technischen Wortlaut
- ✓ und ohne unerklärte Teiländerung

**S03.NAT-02 — was die Anwendung ausgibt, liest sie wieder ein**

- ✓ ⌘Q beendet
- ✓ im frischen Bestand steht dieselbe Sammlung — 2 Zeilen
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

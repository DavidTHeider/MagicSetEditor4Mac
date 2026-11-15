# Hüllenprobe: p13abschluss09

**bestanden** — Commit `718bb0ca9`,
2026-10-20 11:20 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-20T10-13-25Z-83475`, Bauabdruck `43e643558f16`.

- ✓ das Fenster horcht — 127.0.0.1:56251
- ✓ das Fenster horcht — 127.0.0.1:56252

**H1 — am Tisch anbieten; das Angebot trägt Annas Zustimmung und geht hinüber**

- ✓ H1 (bereich, maus): 5 Handlungen, 0 Fehlversuche, 2,5 s Maschinenzeit
- ✓ H1 (bereich, maus): Wirkung beobachtet
- H1: 5 Handlungen (P12 5)
- ✓ H1 ≤ 5 — 5
- ✓ drüben trägt das Angebot Annas Zustimmung — {"vorgang":"h-6ed72eec882aeaff","revision":1,"abdruck":"0da958980aaabd5d554ae7c1fccc741a09a3fb9526a28f3ac62e3533bb9b2854","art":"tausch","initiator":"bafzaajaiaejcbz6ljk72g4fgbqtsnzmlpogbwbmd5dtbnudcibbgs5qdjqwmcavz","parteien":["bafzaajaiaejcbz6ljk72g4fgbqtsnzmlpogbwbmd5dtbnudcibbgs5qdjqwmcavz","bafzaajaiaejcaymf73i2zmuaob24o6wepdytlerf54t6zhmxqspizhbxdybrwqdq"],"ich":"bafzaajaiaejcaymf73i2zmuaob24o6wepdytlerf54t6zhmxqspizhbxdybrwqdq","zustand":"angeboten","positionen":[{"von":"bafzaajaiaejcbz6ljk72g4fgbqtsnzmlpogbwbmd5dtbnudcibbgs5qdjqwmcavz","printing":"00000000-0000-4000-8000-0000000a1301","druckId":0,"menge":1,"guete":"nearMint","sprache":"de","finish":"nonfoil","abbild":"","bewertung":0},{"von":"bafzaajaiaejcaymf73i2zmuaob24o6wepdytlerf54t6zhmxqspizhbxdybrwqdq","printing":"00000000-0000-4000-8000-0000000b1303","druckId":0,"menge":1,"guete":"nearMint","sprache":"de","finish":"nonfoil","abbild":"","bewertung":0}],"bewertungGeteilt":false,"zustimmungen":["bafzaajaiaejcbz6ljk72g4fgbqtsnzmlpogbwbmd5dtbnudcibbgs5qdjqwmcavz"],"ende":false,"beschluss":"zustimmen","lage":"dran","naechster":"zustimmen","moegliche":["gegenangebot","zustimmen","ablehnen","ablaufen"],"phase":"angebot"}
- ✓ Anna sieht, dass es hinausging
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/p13abschluss09/anna-angeboten.png`

**H4 — Bert nimmt an**

- ✓ H4 (bereich, maus): 2 Handlungen, 0 Fehlversuche, 1,0 s Maschinenzeit
- ✓ H4 (bereich, maus): Wirkung beobachtet
- H4: 2 Handlungen (Ausgang 3)
- ✓ H4 ≤ 2 — 2

**H5 — bis beidseitig verbucht, je Seite**

- ✓ H5 (handel, maus): 3 Handlungen, 0 Fehlversuche, 4,5 s Maschinenzeit
- ✓ H5 (handel, maus): Wirkung beobachtet

**Ein Teilfehler: Anna horcht nicht, als Bert die Gegenseite bestätigt**

- ✓ H5 (handel-gegenueber, maus): 3 Handlungen, 0 Fehlversuche, 5,0 s Maschinenzeit
- ✓ H5 (handel-gegenueber, maus): Wirkung beobachtet
- H5: Anna 3, Bert 3 (Ausgang 35 und 30)
- ✓ H5 Anna ≤ 3 — 3
- ✓ H5 Bert ≤ 3 — 3
- ✓ bei Bert eingetragen, und das Fenster sagt, dass Anna es noch nicht hat — ["Bei dir ist es eingetragen, Anna hat es noch nicht: Das Gegenüber ist gerade nicht zu erreichen. Horcht es, und stimmt seine Adresse? Sonst versuche es später noch einmal."]
- ✓ und dass Anna gerade nicht zu erreichen ist, ohne „lade die Liste neu" — Bei dir ist es eingetragen, Anna hat es noch nicht: Das Gegenüber ist gerade nicht zu erreichen. Horcht es, und stimmt seine Adresse? Sonst versuche es später noch einmal.
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/p13abschluss09-bert/bert-unterwegs.png`
- ✓ das Fenster horcht — 127.0.0.1:56264
- ✓ und „Jetzt übertragen" bringt es hinaus

**Am Ende beider Seiten dasselbe**

- ✓ Anna: der Beleg trägt beide Unterschriften — {"vorgang":"h-6ed72eec882aeaff","abdruck":"0da958980aaabd5d554ae7c1fccc741a09a3fb9526a28f3ac62e3533bb9b2854","unterschrift":"afbcf3478ed69680e6138e7bd688032cabbc357daa458684daa939430841fcce70ba6de8a888a5cbc09aaf772c97fea9654871a6fa655235370aeaf5280edb05","beidseitig":true,"gegenzeichner":"bafzaajaiaejcaymf73i2zmuaob24o6wepdytlerf54t6zhmxqspizhbxdybrwqdq","gegengezeichnet":"2026-10-20T11:20:55Z","gebucht":"2026-10-20T11:20:52Z","befund":null}
- ✓ Anna: im Journal zwei Bewegungen, keine doppelt — ["abgang 00000000-0000-4000-8000-0000000a1301","zugang 00000000-0000-4000-8000-0000000b1303"]
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/p13abschluss09/anna-abschluss.png`
- ✓ Bert: der Beleg trägt beide Unterschriften — {"vorgang":"h-6ed72eec882aeaff","abdruck":"0da958980aaabd5d554ae7c1fccc741a09a3fb9526a28f3ac62e3533bb9b2854","unterschrift":"b8c070087441f17095ed907366835d31e1f92a3105b3aa2fc3053a4b85cd6f1b19747f5d2180fae4ff6085bc391289035668483bfe10c7dd943093de2bd0060e","beidseitig":true,"gegenzeichner":"bafzaajaiaejcbz6ljk72g4fgbqtsnzmlpogbwbmd5dtbnudcibbgs5qdjqwmcavz","gegengezeichnet":"2026-10-20T11:20:54Z","gebucht":"2026-10-20T11:20:53Z","befund":null}
- ✓ Bert: im Journal zwei Bewegungen, keine doppelt — ["abgang 00000000-0000-4000-8000-0000000b1303","zugang 00000000-0000-4000-8000-0000000a1301"]
- Bild: `/Users/david/mse-arbeit/p13/.huellenprobe/laeufe/2026-10-20T10-13-25Z-83475/bilder/p13abschluss09-bert/bert-abschluss.png`
- ✓ Anna hat den Sturmrufer und nicht mehr den Glutläufer
- ✓ Bert hat den Glutläufer und nicht mehr den Sturmrufer
- ✓ ⌘Q beendet Anna
- ✓ ⌘Q beendet Bert
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

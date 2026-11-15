# Hüllenprobe: profilwechsel

**bestanden** — Commit `c2c359c4a`,
2026-10-17 14:21 UTC, Mac17,3, macOS 27.0.1,
das Bündel aus `scripts/bauen.sh abnahme`, gestartet über LaunchServices.
Lauf `2026-10-17T14-21-17Z-50185`, Bauabdruck `4e7fdc1a9454`.

- Bob ist Nutzer 2
- ✓ Bobs zweiter Faktor steht

**Wechseln ohne Code fragt nach ihm**

- gearbeitet wird bei »Mein Profil«
- ✓ der Wechsel verlangt den zweiten Faktor und bleibt, wo er war — Mein Profil

**Ein falscher Code wird abgewiesen**

- ✓ der falsche Code wird abgewiesen — Dieser Code stimmt nicht — oder er ist schon verbraucht. Prüfe den Code auf dem Zettel — jeder gilt einmal.
- ✓ und gearbeitet wird weiter im alten Profil

**Der richtige Code wechselt**

- ✓ der richtige Code wechselt in Bobs Profil — Mein Profil → I Ich · Mein Profil von Bob nur lesen › Profil wechseln …
- ✓ als derselbe Mensch, mit der Rolle aus Bobs Profil
- mse profil: 2	Mein Profil
- ✓ ⌘Q beendet
- ✓ die echten Orte dieses Rechners blieben unberührt
- ✓ die echten Orte ließen sich vollständig lesen
- ✓ die Schutzrichtlinie wies nichts ab, was das Drehbuch braucht

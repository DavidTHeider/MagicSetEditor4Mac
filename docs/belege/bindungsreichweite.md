# Beleg: Reichweite der nativen Bindungen — B01

**Lauf `2026-10-09T19-47-35Z-46912`** — Commit `1c0b74221`, 2026-10-09 22:27 UTC, Mac17,3, macOS 27.0.1, 106 Drehbücher.
Ausgewertet mit `node scripts/bindungsreichweite.mjs .huellenprobe/laeufe/2026-10-09T19-47-35Z-46912 --beleg docs/belege/bindungsreichweite.md`.

Gezählt ist, was ein Drehbuch **am nativen Fenster** gerufen hat (`internal/fenster/rufbuch.go`), gegen **jede** Bindung des Vertrags. Lesend und schreibend stehen in `docs/SCHREIBWEGE.md`, die Dateidialoge kommen aus dem Aufrufgraphen (`scripts/bindungssignaturen.go`); Ausnahmen mit Grund stehen neben der Quote. **Erreicht** heißt: der Ruf kam in der Bindung an (gelungen, fachlich abgelehnt oder abgebrochen); **gelungen** heißt: mindestens ein Ruf kam ohne Fehler zurück.

| Klasse | Nenner | erreicht | Quote | gelungen | Quote | ausgenommen |
| --- | --- | --- | --- | --- | --- | --- |
| alle | 495 | 495 | 100,0 % | 454 | 91,7 % | 37 |
| schreibend | 268 | 268 | 100,0 % | 254 | 94,8 % | 25 |
| lesend | 227 | 227 | 100,0 % | 200 | 88,1 % | 12 |
| Dateidialog | 45 | 45 | 100,0 % | 45 | 100,0 % | 1 |

**Ohne den Vertragsrundgang** — nur was ein Bedienweg erreichte:

| Klasse | Nenner | erreicht | Quote | gelungen | Quote | ausgenommen |
| --- | --- | --- | --- | --- | --- | --- |
| alle | 495 | 254 | 51,3 % | 254 | 51,3 % | 37 |
| schreibend | 268 | 104 | 38,8 % | 104 | 38,8 % | 25 |
| lesend | 227 | 150 | 66,1 % | 150 | 66,1 % | 12 |
| Dateidialog | 45 | 22 | 48,9 % | 22 | 48,9 % | 1 |

**Abgewiesen am Tor oder unbekannt** (2): `Datenstand` (4), `Einstellungen` (1).

**Gerufen, aber nicht im Vertrag**: keine.

## Nicht erreichte schreibende Bindungen (0)

—

## Ausnahmen

- `Anmelden` — Anmeldung bei einem fremden Anbieter im Browser
- `BiometrieEinrichten` — Touch ID — ein Systemdialog, den ein Mensch beantwortet
- `DateiAnsehen` — öffnet die Übersicht von macOS
- `DateiImFinderZeigen` — öffnet ein anderes Programm (Finder)
- `DateiMitProgrammOeffnen` — öffnet ein anderes Programm
- `EntsperrenMitSensor` — Touch ID — ein Systemdialog, den ein Mensch beantwortet
- `FernHorchStarten` — verbindet über tailcat mit fremden Netzen (DERP)
- `HorchStarten` — öffnet einen Lauscher im lokalen Netz — macOS fragt nach der Netzerlaubnis
- `IndexAufbauen` — rechnet Einbettungen mit einem lokalen Modell — Minuten, nicht Sekunden (AI-Phase)
- `KIBatchVorschlagen` — rechnet mit einem KI-Modell — Laufzeit und Ergebnis hängen am Modell (AI-Phase)
- `KIBildVorschlagen` — rechnet mit einem Bildmodell — Laufzeit und Ergebnis hängen am Modell (AI-Phase)
- `KIEinschaetzen` — rechnet mit einem KI-Modell — Laufzeit und Ergebnis hängen am Modell (AI-Phase)
- `KIKarteErkennen` — rechnet mit einem KI-Modell — Laufzeit und Ergebnis hängen am Modell (AI-Phase)
- `KIVorschlagen` — rechnet mit einem KI-Modell — Laufzeit und Ergebnis hängen am Modell (AI-Phase)
- `KontoAnmelden` — Anmeldung bei einem fremden Anbieter im Browser
- `Marktangebote` — fremder Dienst CardTrader — braucht einen Zugang
- `Marktkonto` — fremder Dienst CardTrader — braucht einen Zugang
- `MarktlinkOeffnen` — öffnet den Browser
- `NachNeuerFassungSehen` — fragt GitHub — I03 (#2261)
- `NeueFassungLaden` — lädt eine Fassung von GitHub — I03 (#2261)
- `PaketBeziehen` — lädt ein Vorlagenpaket aus dem Netz
- `PaketHerkunft` — fragt den Bezugsort im Netz
- `Preisbulk` — fremder Dienst (MTGJSON-Preise)
- `PreiseHolen` — fremder Dienst (MTGJSON-Preise)
- `SchnittstelleEinrichten` — registriert einen LaunchAgent in der echten Sitzung (launchctl) — eigenes Drehbuch in G08 (#2272)
- `SchnittstelleWiderrufen` — nimmt einen LaunchAgent aus der echten Sitzung — eigenes Drehbuch in G08 (#2272)
- `ScryfallBild` — fremder Dienst Scryfall
- `ScryfallBildEinsetzen` — fremder Dienst Scryfall
- `ScryfallKarteAnlegen` — fremder Dienst Scryfall
- `ScryfallManakosten` — fremder Dienst Scryfall
- `ScryfallRegeln` — fremder Dienst Scryfall
- `ScryfallSammeldatenHolen` — lädt die Sammeldaten von Scryfall (über 100 MB)
- `ScryfallSatzEinlesen` — fremder Dienst Scryfall
- `ScryfallSuche` — fremder Dienst Scryfall
- `ScryfallUebernehmen` — fremder Dienst Scryfall
- `ScryfallVorschlag` — fremder Dienst Scryfall
- `SichtbarWerden` — meldet sich per mDNS im lokalen Netz — macOS fragt nach der Netzerlaubnis

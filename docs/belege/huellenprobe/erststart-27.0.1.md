# Beleg: HAND6 — der Erststart aus dem Abbild auf macOS 27.0.1

**Stand:** Commit `df9eb1fa8` · **Datum:** 2026-10-19 19:28 UTC · **Befehl:** `node scripts/erststartgast.mjs MagicSetEditor4Mac-0.8.0.dmg mse090-r02` · **Rückgabe:** 0

**Maschine:** in der Lume-VM `mse090-r02` — macOS 27.0.1 (26A434), VirtualMac2,1, 4 Kerne.

- **Abbild**: `MagicSetEditor4Mac-0.8.0.dmg`, SHA-256 `c106a56f936c69d2580014664017352b1925fe0dd9f4bbbc78bad557d5d935a6`, gebaut am Stand `df9eb1fa8` mit `MSE_FREIGABE_NUR_BAU=1 ./scripts/release.sh` — **kein Freigabebau**: ohne Pipeline, Prüfsummenliste und Signatur.
- **Bündel**: `CFBundleIdentifier` = `io.github.davidtheider.magicseteditor4mac`, `CFBundleShortVersionString` = `0.8.0`, `LSMinimumSystemVersion` = `15.2`.
- **Quarantäne**: kenntlich synthetisch, Absender `mse4mac-hand6`; am herauskopierten Bündel `0083;6ad66f6c;mse4mac-hand6;`. **Kein `xattr -d`** in diesem Lauf.

| Teil | Schritt | Ergebnis | ms | Wert oder Fehler |
| --- | --- | --- | ---: | --- |
| Abbild | Prüfsumme | gelungen | 3558 | c106a56f936c69d2 |
| Abbild | Quarantäne am Abbild | gelungen | 15 | 0083;6ad66f6c;mse4mac-hand6; |
| Abbild | einhängen | gelungen | 5437 | ERSTER START — bitte lesen.txt, MagicSetEditor4Mac.app, Programme |
| Abbild | ins Programme-Verzeichnis ziehen | gelungen | 2374 | 1 |
| Abbild | die Quarantäne erbt | gelungen | 9 | 0083;6ad66f6c;mse4mac-hand6; |
| Gatekeeper | spctl | gelungen | 392 | Rückgabe 3: /tmp/hand6-ir7ZN1/Programme/MagicSetEditor4Mac.app: rejected  |
| Gatekeeper | Siegel | gelungen | 39 | flags=0x10002(adhoc,runtime) |
| Gatekeeper | ein Doppelklick startet nicht | gelungen | 8286 | nicht gestartet, kein Bestand |
| Erststart | ohne Quarantäne | gelungen | 3 | keine |
| Erststart | das Programm läuft | gelungen | 17 | 48 ms |
| Erststart | der Bestand entsteht | gelungen | 1653 | 1701 ms |
| Erststart | das Programm läuft noch | gelungen | 20 | 1 |
| Erststart | beenden | gelungen | 38 | beendet |

## Was entstand

Schema: 100 Wanderungen.

| Pfad | Art | Rechte | Größe |
| --- | --- | --- | --- |
| `.fl7D26BE60` | Datei | `0444` | 0 |
| `Sitzungen` | Verzeichnis | `0700` |  |
| `Sitzungen/sitzungen.sperre` | Datei | `0600` | 0 |
| `bestand.db` | Datei | `0600` | 1077248 |
| `bestand.db.eintritt` | Datei | `0600` | 0 |
| `bestand.db.sperre` | Datei | `0600` | 0 |
| `bestand.db.wanderung` | Datei | `0600` | 0 |
| `mse.log` | Datei | `0600` | 360 |
| `mse.log.sperre` | Datei | `0600` | 0 |
| `zuhause` | Verzeichnis | `0755` |  |
| `zuhause/Library` | Verzeichnis | `0750` |  |
| `zuhause/Library/Application Support` | Verzeichnis | `0750` |  |
| `zuhause/Library/Application Support/MagicSetEditor4Mac` | Verzeichnis | `0750` |  |
| `zuhause/Library/Application Support/MagicSetEditor4Mac/pakete` | Verzeichnis | `0750` |  |
| `zuhause/Library/Caches` | Verzeichnis | `0700` |  |
| `zuhause/Library/Caches/MagicSetEditor4Mac` | Verzeichnis | `0700` |  |
| `zuhause/Library/Caches/MagicSetEditor4Mac/bildvorschlaege` | Verzeichnis | `0700` |  |
| `zuhause/Library/Caches/MagicSetEditor4Mac/scryfall` | Verzeichnis | `0750` |  |
| `zuhause/Library/Caches/MagicSetEditor4Mac/scryfall/bilder` | Verzeichnis | `0700` |  |
| `zuhause/Library/Caches/MagicSetEditor4Mac/scryfall/sammeldaten` | Verzeichnis | `0700` |  |

## Was syspolicyd sagte

```
Timestamp               Ty Process[PID:TID]
2026-10-19 12:28:59.336 Df syspolicyd[233:4e5] [com.apple.syspolicy.exec:default] GK evaluateScanResult: 0, PST: (path: 5b7f835f7c58f0fa), (team: (null)), (id: io.github.davidtheider.magicseteditor4mac), (bundle_id: io.github.davidtheider.magicseteditor4mac), 1, 0, 1, 0, 0, 0, 0
2026-10-19 12:28:59.337 Df syspolicyd[233:4e5] [com.apple.syspolicy.exec:default] Prompt shown (6, 0), waiting for response: PST: (path: 5b7f835f7c58f0fa), (team: (null)), (id: io.github.davidtheider.magicseteditor4mac), (bundle_id: io.github.davidtheider.magicseteditor4mac)
```

## Grenzen

- »Dennoch öffnen« unter Datenschutz & Sicherheit verlangt das Kennwort eines Verwalters — gegeben von einem Menschen, nicht von diesem Lauf (REL, #2385). Teil 4 fährt das Bündel ohne Quarantäne, wie nach dieser Freigabe.
- **Der Bezug über den Freigabekanal ohne Zugang zum Repositorium** ist hier nicht gelaufen: das Abbild kam über das geteilte Verzeichnis. Er braucht ein öffentliches Repositorium (Vermerk 0190) und gehört zum Freigabegate (#2385).

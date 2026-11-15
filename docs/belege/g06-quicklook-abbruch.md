# Beleg: G06 — die Quick-Look-Probe räumt nach einem Abbruch auf

**Stand:** Commit `c6e379ba6` · **Datum:** 2026-10-11 06:54 UTC · **Befehl:** `MSE_PROBE_ABBRECHEN=1 node scripts/quicklookgast.mjs mse080-152` · **Rückgabe:** 130

**Maschine:** Lume-VM `mse080-152` — macOS 15.2 (24C101). Das Probebündel lag nur dort, wurde wieder abgemeldet und gelöscht (aufgeräumt: ja).

**Ergebnis:** die ad hoc signierte Erweiterung **lief nicht**; die Farbe sagt nichts über die Pakete.

| | Antwort von `QLThumbnailGenerator` für `probe.mse-set` (256 × 256 verlangt) |
| --- | --- |
| vorher, ohne Erweiterung (Gegenprobe) | symbol, 256 × 256, Mitte rgb(242,242,242) |
| nachher, mit Erweiterung | kein Bild — ? |

- **Registrierung** (`pluginkit -m -v`): ` de.mse4mac.probe.quicklook.vorschaubild(1.0) 224C8F6C-7E1B-4689-93DB-D384776A62B4 2026-10-11 06:54:41 +0000 /Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex (1 plug-in)`
- **Signatur der Erweiterung**: `Executable=/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex/Contents/MacOS/MSEVorschau Identifier=de.mse4mac.probe.quicklook.vorschaubild Format=bundle with Mach-O thin (arm64) CodeDirectory v=20400 size=480 flags=0x2(adhoc) hashes=4+7 location=embedded Signature=adhoc Info.p`
- **Berechtigungen**: `<dict><key>com.apple.security.app-sandbox</key><true/></dict>`
- **Größe der Erweiterung**: 68 KB (ohne Renderer — er zeichnet eine Farbfläche).
- **Abbruch**: Signal TERM — **der Lauf ist unvollständig**

## Das Protokoll

```text
$ open -W -a Messer.app (vorher)
rc=0
$ ditto '/Volumes/My Shared Files/g06-gast-BUGCqE/ql/MSEProbe.app' '/Users/lume/Applications/MSEProbe.app'

rc=0
$ codesign -dvv '/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex'
Executable=/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex/Contents/MacOS/MSEVorschau
Identifier=de.mse4mac.probe.quicklook.vorschaubild
Format=bundle with Mach-O thin (arm64)
CodeDirectory v=20400 size=480 flags=0x2(adhoc) hashes=4+7 location=embedded
Signature=adhoc
Info.plist entries=8
TeamIdentifier=not set
Sealed Resources version=2 rules=13 files=0
Internal requirements count=0 size=12
rc=0
$ codesign -d --entitlements - --xml '/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex'
Executable=/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex/Contents/MacOS/MSEVorschau
<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "https://www.apple.com/DTDs/PropertyList-1.0.dtd"><plist version="1.0"><dict><key>com.apple.security.app-sandbox</key><true/></dict></plist>
rc=0
$ du -sk '/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex'
68	/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex
rc=0
$ '/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister' -f -R '/Users/lume/Applications/MSEProbe.app'

rc=0
$ pluginkit -a '/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex'

rc=0
$ pluginkit -m -v -i de.mse4mac.probe.quicklook.vorschaubild
     de.mse4mac.probe.quicklook.vorschaubild(1.0)	224C8F6C-7E1B-4689-93DB-D384776A62B4	2026-10-11 06:54:41 +0000	/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex
 (1 plug-in)
rc=0
$ pluginkit -r '/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex'

rc=0
$ '/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister' -u '/Users/lume/Applications/MSEProbe.app'

rc=0
$ pluginkit -m -i de.mse4mac.probe.quicklook.vorschaubild

rc=0

```

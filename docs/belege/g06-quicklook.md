# Beleg: G06 — Quick Look für Setdateien, die Machbarkeit

**Stand:** Commit `c6e379ba6` · **Datum:** 2026-10-11 06:54 UTC · **Befehl:** `node scripts/quicklookgast.mjs mse080-152` · **Rückgabe:** 0

**Maschine:** Lume-VM `mse080-152` — macOS 15.2 (24C101). Das Probebündel lag nur dort, wurde wieder abgemeldet und gelöscht (aufgeräumt: ja).

**Ergebnis:** die ad hoc signierte Erweiterung **lief**; aus der Sandbox waren die Vorlagenpakete **nicht** lesbar.

| | Antwort von `QLThumbnailGenerator` für `probe.mse-set` (256 × 256 verlangt) |
| --- | --- |
| vorher, ohne Erweiterung (Gegenprobe) | symbol, 256 × 256, Mitte rgb(242,242,242) |
| nachher, mit Erweiterung | vorschaubild, 183 × 256, Mitte cyan |

- **Registrierung** (`pluginkit -m -v`): ` de.mse4mac.probe.quicklook.vorschaubild(1.0) 9BCE6C43-D718-4010-83A6-C7C6BD6B71B9 2026-10-11 06:55:10 +0000 /Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex (1 plug-in)`
- **Signatur der Erweiterung**: `Executable=/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex/Contents/MacOS/MSEVorschau Identifier=de.mse4mac.probe.quicklook.vorschaubild Format=bundle with Mach-O thin (arm64) CodeDirectory v=20400 size=480 flags=0x2(adhoc) hashes=4+7 location=embedded Signature=adhoc Info.p`
- **Berechtigungen**: `<dict><key>com.apple.security.app-sandbox</key><true/></dict>`
- **Größe der Erweiterung**: 68 KB (ohne Renderer — er zeichnet eine Farbfläche).

## Das Protokoll

```text
$ open -W -a Messer.app (vorher)
rc=0
$ ditto '/Volumes/My Shared Files/g06-gast-WfACv5/ql/MSEProbe.app' '/Users/lume/Applications/MSEProbe.app'

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
     de.mse4mac.probe.quicklook.vorschaubild(1.0)	9BCE6C43-D718-4010-83A6-C7C6BD6B71B9	2026-10-11 06:55:10 +0000	/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex
 (1 plug-in)
rc=0
$ open -W -a Messer.app (nachher)
rc=0
$ log show --last 3m --style compact --predicate 'eventMessage CONTAINS[c] "MSEVorschau" OR eventMessage CONTAINS[c] "de.mse4mac.probe"' | tail -8
2026-10-10 23:55:12.995 Df runningboardd[163:426] [com.apple.runningboard:process] [xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}:630] Result 45 setting darwin role to None: Operation not supported, falling back to setting priority
2026-10-10 23:55:12.995 Df runningboardd[163:426] [com.apple.runningboard:ttl] [xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}:630] Set darwin priority to: PRIO_DARWIN_BG
2026-10-10 23:55:12.995 Df runningboardd[163:a38] [com.apple.runningboard:process] [xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}:630] check if suspended process is holding locks
2026-10-10 23:55:12.995 Df WindowServer[144:4af] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:12.995 Df WindowServer[144:4af] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:12.995 Df UIKitSystem[430:a84] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:12.995 Df UIKitSystem[430:a84] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:13.021 Df log[634:11a7] [com.apple.log:] log run noninteractively, parent: 633 (bash), args: 'log' 'show' '--last' '3m' '--style' 'compact' '--predicate' 'eventMessage CONTAINS[c] "MSEVorschau" OR eventMessage CONTAINS[c] "de.mse4mac.probe"'
rc=0
$ pluginkit -r '/Users/lume/Applications/MSEProbe.app/Contents/PlugIns/MSEVorschau.appex'

rc=0
$ '/System/Library/Frameworks/CoreServices.framework/Frameworks/LaunchServices.framework/Support/lsregister' -u '/Users/lume/Applications/MSEProbe.app'

rc=0
$ pluginkit -m -i de.mse4mac.probe.quicklook.vorschaubild

rc=0
2026-10-10 23:55:12.995 Df runningboardd[163:426] [com.apple.runningboard:process] [xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}:630] Result 45 setting darwin role to None: Operation not supported, falling back to setting priority
2026-10-10 23:55:12.995 Df runningboardd[163:426] [com.apple.runningboard:ttl] [xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}:630] Set darwin priority to: PRIO_DARWIN_BG
2026-10-10 23:55:12.995 Df runningboardd[163:a38] [com.apple.runningboard:process] [xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}:630] check if suspended process is holding locks
2026-10-10 23:55:12.995 Df WindowServer[144:4af] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:12.995 Df WindowServer[144:4af] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:12.995 Df UIKitSystem[430:a84] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:12.995 Df UIKitSystem[430:a84] [com.apple.runningboard:monitor] Received state update for 630 (xpcservice<de.mse4mac.probe.quicklook.vorschaubild([osservice<com.apple.quicklook.ThumbnailsAgent(501)>:379])(501)>{vt hash: 187071660}, running-suspended-NotVisible
2026-10-10 23:55:13.021 Df log[634:11a7] [com.apple.log:] log run noninteractively, parent: 633 (bash), args: 'log' 'show' '--last' '3m' '--style' 'compact' '--predicate' 'eventMessage CONTAINS[c] "MSEVorschau" OR eventMessage CONTAINS[c] "de.mse4mac.probe"'

```

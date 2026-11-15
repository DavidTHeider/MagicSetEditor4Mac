# 0184 — Die Untergrenze ist macOS 15.2, weil nur sie belegt ist

**Stand:** angenommen am 2026-10-11 · I02
(#2260).
**Rücknehmbar**, sobald ein Lauf auf 13 oder 14 vorliegt — der Weg steht unten.

## Die Frage

`build/darwin/Info.plist` sagte seit Phase 1 **macOS 13**, und gebaut wurde
dafür (`scripts/macos.sh`, `minos 13.0` an allen vier Programmen,
`vtool -show-build`). **Gelaufen ist die Anwendung dort nie**: die
Abnahmebelege von 0.6.0 und 0.7.0 führen die Zeile als „ungeprüft", und
belegt waren der Wirt (macOS 27) und ein Gast mit 15.2 (B02). I02 verlangt
**entweder** den Kernweg mit dem unveränderten Bündel auf 13 und 14 **oder**
die Mindestversion auf den belegten Stand zu heben, mit Entscheidung.

## Was versucht wurde — gemessen am 2026-10-11

| Weg | Ergebnis |
| --- | --- |
| `lume create --ipsw macos-13.6.ipsw --unattended sequoia` (lume 0.5.3, Abbild 22G120, SHA-256 geprüft) | Installation lief durch; die **Offline-Einrichtung scheiterte**: `diskutil mount disk7s5 failed … This is an encrypted and locked APFS Volume`. Das Datenvolume eines Gastes auf Apples Virtualisierung ist verschlüsselt, und entsperren hieße ein Kennwort eingeben. lume räumte den halben Gast selbst weg |
| dasselbe mit `macos-14.6.1.ipsw` (23G93, SHA-256 geprüft) | **an derselben Stelle gescheitert**, mit demselben Satz, drei Minuten nach dem Beginn der Installation |
| Ein fertiges Abbild für 13 oder 14 | **Gibt es nicht**: trycua führt in der Registry nur `macos-sequoia-vanilla` (`15.2`, `latest`); `macos-sonoma-vanilla` und `macos-ventura-vanilla` antworten `DENIED` |
| Die Einrichtung von Hand über den Setup-Assistenten | verlangt ein **Kontokennwort**, und Kennwörter gibt dieser Auftrag nicht ein, auch nicht im Gast ([`FREIGABEN.md`](../FREIGABEN.md)); ohne Kennwort öffnet `lume ssh` nicht |
| Ein Rechner mit 13 oder 14 | gibt es in dieser Umgebung nicht (`out of scope`, wie zwei physische Macs) |

Die Vorgaben der unbeaufsichtigten Einrichtung heißen bei lume ohnehin
`sequoia` und `tahoe`.

## Was belegt ist

- **macOS 15.2** (24C101), Lume-Gast, vier Kerne, 4 GB: der Kernweg mit dem
  unveränderten Abnahmebündel — Kommandozeile und Fenster, elf Schritte
  ([`belege/huellenprobe/untergrenze-15.2.md`](../belege/huellenprobe/untergrenze-15.2.md)),
  dazu die Starts aus B02 ([`belege/huellenprobe/budgets-gast.md`](../belege/huellenprobe/budgets-gast.md)).
- **macOS 27.0.1** am Wirt, derselbe Treiber als Positivkontrolle
  ([`belege/huellenprobe/untergrenze-wirt.md`](../belege/huellenprobe/untergrenze-wirt.md)).

**Und die Grenze greift** — Gegenprobe am Wirt: eine Kopie des Bündels mit
`LSMinimumSystemVersion` 28.0, ad hoc neu signiert, weist macOS 27 mit
`_LSOpenURLsWithCompletionHandler() … error -10825`
(`kLSIncompatibleSystemVersionErr`) ab; kein Prozess startet.

## Die Entscheidung

**`LSMinimumSystemVersion` und `MACOSX_DEPLOYMENT_TARGET` werden 15.2.**
Zugesagt wird, was ein Lauf belegt — nicht weniger, nicht mehr. §10: eine
unbedingte Zusage mit stillen Ausnahmen ist schlimmer als eine engere.

## Verworfen

- **13 behalten, „zugesagt und ungeprüft".** So stand es seit 0.6.0 in der
  Releasenotiz, und genau das ist eine Zusage ohne Einlösung. Gebaut wurde
  für 13 — ob Go 1.27, Wails v3.0.0-beta.27 und die Aufrufe in WebKit und
  AppKit dort tragen, sagt kein Bau.
- **15.0.** Belegt ist 15.2; ein Unterschied zwischen 15.0 und 15.2 ist
  nicht gemessen, und ein Abstand, der nicht gemessen ist, ist kein Beleg
  für seine Abwesenheit.
- **Einrichten mit leerem Kennwort.** Nur eine Lesart der Freigaben, und
  `lume ssh` meldet sich ohne Kennwort nicht an.

## Der Preis

**Wer macOS 13, 14, 15.0 oder 15.1 hat, kann diese Fassung nicht starten** —
der Finder sagt es mit seinem eigenen Dialog. Ob 0.7.0 dort lief, weiß
niemand; diese Fassung sagt es jetzt, statt es zu behaupten. Der Eintrag
steht unter **Geändert** im CHANGELOG und in der Releasenotiz.

## Der Rückweg

Läuft der Kernweg auf 13 oder 14 — auf einem Rechner oder in einem Gast,
den ein Mensch eingerichtet hat —, wird die Untergrenze dorthin gesenkt.
**Gemessen wird ein Prüfbündel, nicht das Produkt** (Befund 4 des
Zweitblicks über Phase I): das Bündel dieser Fassung verlangt 15.2 und
startet auf 13 gar nicht. Also in einem eigenen Zweig `build/darwin/Info.plist`
**und** `scripts/macos.sh` auf die Zahl setzen, die geprüft werden soll,
dann

    ./scripts/bauen.sh abnahme && node scripts/untergrenzegast.mjs <vm>

Der Lauf schreibt `belege/huellenprobe/untergrenze-<system>.md`; erst wenn
alle Schritte gelingen, wird die Zusage im Produkt gesenkt — mit diesem
Prüfbündel als Beleg.

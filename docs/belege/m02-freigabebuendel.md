# Beleg: Sichtprüfung am ausgelieferten Abbild von 0.7.0 — M02

**Gefahren am 2026-10-08** zwischen Bau und Veröffentlichung, am Abbild
`MagicSetEditor4Mac-0.7.0.dmg` (SHA-256 `577d25f752988ec3e881f51dd8cda57220d749175110f45bbc26c43945e5f7a7`,
Commit `ad53f5e3c`), auf diesem Rechner (Apple M5, macOS 27.0.1). Es sind die
Handprüfungen 3 und 4 des Freigabebelegs.

## Warum nicht das Abnahmebündel

Die Hüllenprobe läuft am **Abnahme**bündel — eigene Kennung, eigener
Steuerkanal. Ausgeliefert wird ein anderes: `scripts/bauen.sh freigabe`,
Kennung `com.wails.MagicSetEditor4Mac`, dazu Vorlagen, Schriften und
Werkzeuge aus `release.sh`. Ob **dieses** Bündel zeichnet, sah bis hierher
niemand — seit 0.5.0 führen die Freigabebelege die Prüfungen 3 und 4 offen.

## Wie, ohne den Rechner anzufassen

**Das Bündel trägt die echte Kennung**, und eine installierte Anwendung
dieses Rechners teilt mit ihm Vorgaben, Webspeicher und Caches. Deshalb:

- das Abbild eingehängt (`hdiutil attach -nobrowse -readonly`), die
  Anwendung **aus ihm** gestartet, über LaunchServices;
- `HOME` **und** `CFFIXED_USER_HOME` auf ein Wegwerfzuhause umgelenkt — das
  zweite nimmt Foundation für Vorgaben, WebKit-Daten und Caches;
  `MSE_SCHLUESSELBUND` auf einen eigenen Bund;
- **vorher und nachher gemessen**: `Preferences`, `WebKit`, `Caches`,
  `Saved Application State`, `HTTPStorages` und `Application Support` der
  echten Anwendung — unverändert, und im Wegwerfzuhause entstanden genau
  diese Orte.

> **Berichtigt am 2026-10-09** (A01,
> #2215).
> Der Satz oben gilt für **diesen** Lauf und trägt nicht weiter. Der Audit zu
> 0.8.0 (Befund F-02) startete dasselbe Bündel mit `HOME` **und**
> `CFFIXED_USER_HOME` umgelenkt und sah nach dem **Speichern-Dialog** die
> echte `~/Library/Preferences/com.wails.magicseteditor4mac.plist` wachsen
> (1017 → 1125 Bytes): der Dialog läuft in einem eigenen Dienst, und
> `cfprefsd` legt die Vorgaben unter der **Kennung** ab, nicht unter einem
> Zuhause. **Hier wurde kein Dateidialog geöffnet** — deshalb blieb der
> Abdruck leer. Eine sichere Prüfgrenze für das Freigabebündel ist allein
> eine **eigene Kennung** oder ein **Gast**; die Hüllenprobe nimmt die
> eigene Kennung, und ihr Abdruck sieht seit A01 alle sieben Orte unter der
> Produktkennung in voller Tiefe (`scripts/huellenorte.mjs`).

Erscheinungsbild, Sprache und die „Ersten Schritte" setzt das `mse` aus dem
Abbild im selben Zuhause; Bilder nimmt `screencapture -l` vom Fenster.

## Was zu sehen war

| Bild | Was es zeigt |
| --- | --- |
| Erster Start, hell, Deutsch | die Startseite mit den „Ersten Schritten" darüber — „Acht Vorlagenpakete sind schon da" |
| Nach dem Öffnen des mitgelieferten `Vanguard.mse-set` | 1 Set, 107 Karten |
| Startseite hell und **dunkel**, ohne die „Ersten Schritte" | „Weiterarbeiten" mit dem Set, „Was Aufmerksamkeit braucht"; lesbar, geordnet, die Leiste mit allen Arbeitsbereichen |
| `mse zeichnen` aus dem Abbild, erste Karte | *Akroma, Angel of Wrath Avatar* — Rahmen, Bild, **Titel, Typzeile, Regeltext und Zahlen in den mitgelieferten Schriften**, 375 × 523 in 0,7 s |

**Die vierte Zeile ist der Befund von 0.6.0, nachgeprüft**: das erste Abbild
von 0.6.0 zeichnete Karten ohne Text, weil die Schriften nicht im Bündel
lagen (#2071).

## Was dieser Beleg nicht sagt

- **Die übrigen Seiten am ausgelieferten Bündel.** Bedienen ließ es sich hier
  nicht: das Freigabebündel hat keinen Steuerkanal, und dieser Prozess darf
  keine Ereignisse senden. Die übrigen Seiten, hell und dunkel, belegt die
  Hüllenprobe am selben Quelltext (98 Drehbücher grün).
- **Den Gatekeeper-Weg und den Bestand eines Rechners, der schon einen hat**
  — Handprüfungen 6 und 7; das Abbild trug hier keine Quarantäne.
- **Den Druckbogen** — Handprüfung 5.

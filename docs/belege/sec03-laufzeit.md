# Beleg: Hardened Runtime am Abnahmebündel — SEC03

**Gemessen am 2026-10-16**, 04:19 bis 04:32 UTC, auf Mac17,3 (Apple M5) mit
macOS 27.0.1. Der Zweig war `feat/2356-sec03-signaturkette` am Commit
`534148913`. Das Bündel kam aus `scripts/bauen.sh abnahme`, Bauabdruck
`a9807ee120fe`, Kennung `io.github.davidtheider.magicseteditor4mac.abnahme`.
Issue: #2356.
Entscheidung: [0195](../entscheidungen/0195-hardened-runtime-ohne-entitlements.md).

**Eingaben**: Jedes Drehbuch lief mit eigenem Bestand, eigenem Schlüsselbund
und eigenem Zuhause unter `$TMPDIR/huelle-*` (`p.ort()`). Die Werkzeuge liefen
in einem Wegwerf-Zuhause mit `MSE_BESTAND`, `MSE_SCHLUESSELBUND` und
`MSE_DIENST_GESPERRT=an`. Keine echten Bestände, keine Konten, kein Netz außer
der Rückschleife.

## Das Siegel

Neu gesiegelt mit `codesign --force --deep --options runtime --sign -`, ohne
Entitlements (`codesign -d --entitlements -` gibt keine aus). Alle vier
Programme tragen danach `flags=0x10002(adhoc,runtime)`:
`MagicSetEditor4Mac`, `mse`, `mse-mcp` und `mse-rest`.
`codesign --verify --deep --strict` hält.

## Was lief

| Weg | Was ihn trägt | Lauf | Gehärtet | Ohne Runtime |
| --- | --- | --- | --- | --- |
| Fenster, WebKit, Wails' MCP-Server (401 ohne Schlüssel), ⌘Q, Leiste | `start` | `2026-10-16T04-19-41Z-135` | 11 von 12 | `2026-10-16T04-20-18Z-1254`: **derselbe** eine Fund |
| Kernwege: Set, Karte, Vorschau, Sichern | `kernwege` | `2026-10-16T04-25-54Z-2955` | bestanden | — |
| purego: Speicher der WebKit-Prozesse | `betriebsband09` | dito | bestanden | — |
| `tmc/apple`: Teilen-Dialog (`NSSharingServicePicker`) | `net04` | dito | bestanden | — |
| Jede Bindung unmittelbar | `vertrag` | `2026-10-16T04-28-33Z-3936` | 31 von 32 Gruppen | `2026-10-16T04-29-01Z-4377`: **dieselben** 31, derselbe Fund |
| LocalAuthentication ohne Dialog (`Faktoren` → `biometrie.Stand`) | Wegwerf-Drehbuch, nicht eingecheckt | `2026-10-16T04-31-24Z-5859` | `biometrie: true`, `Touch ID` | — |
| Leiste, Dateidialoge (NSOpenPanel, NSSavePanel), Ordnerwahl | `eintraege` | `2026-10-16T04-31-56Z-6110` | 34 von 34 | — |
| `mse hilfe`, `mse sets` (SQLite, neu migriert) | Hand, Wegwerf-Zuhause | — | Exit 0, „Kein Set im Bestand." | — |
| `mse-mcp`: `initialize` und `tools/list` | Hand | — | `magicseteditor4mac 0.8.0`, 329 Werkzeuge | — |
| `mse-rest -port 0 -ansage …` | Hand | — | `GET /sets`: 200 mit Marke, 401 ohne | — |

**Die zwei Funde liegen nicht an der Runtime.** Beide stehen im selben Bündel
auch ohne `runtime`:

- **`start`, „Wie zuletzt"**: Nach ⌘Q bleiben `fensterbreite` und
  `fensterhoehe` leer. Das ist eine Regression aus F02 (#2341):
  `fenstergroesse` fragt in einer eigenen Goroutine, Wails reicht `Size()`
  an den Hauptfaden, und der wartet im Schlussrückruf selbst
  (`pkg/application/application.go:981`, v3.0.0-beta.27) — #2439.
- **`vertrag`, `Nachbarn`**: „ins lokale Netz ging nichts hinaus". Das
  Abnahmebündel hat unter seiner neuen Kennung auf diesem Rechner nie die
  Erlaubnis für das lokale Netz bekommen. Das ist die offene Handprüfung
  aus Phase 0.

## Was nicht lief

- **Touch ID mit Finger**: Der Systemdialog braucht einen Menschen. Gemessen
  ist nur die Abfrage ohne Dialog.
- **Quick Look**, **Apples Gerätemodell** (`fm serve`), **Ollama**, der
  **Fernweg über tailcat** und die **Erlaubnis für das lokale Netz**.
- **Ein frischer Gast**: Gemessen ist dieser Rechner. Die Gesamtabnahme in
  Phase 9 fährt alle Drehbücher mit dem Bündel, das `bauen.sh` jetzt so
  siegelt.

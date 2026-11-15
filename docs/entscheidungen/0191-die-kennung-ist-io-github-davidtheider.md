# 0191 — Die Kennung ist `io.github.davidtheider.magicseteditor4mac`

**Stand:** angenommen am 2026-10-15 · #2332 (D02).
**Entschieden vom Auftraggeber** am 2026-10-15 auf die Rückfrage vom
2026-10-13: „ja, variante 1. Das Repository wird später öffentlich gestellt."
**Umgesetzt** wird der Umzug in R02
(#2334), als
eigener Zug mit Rückweg (§5, „Drei Migrationen, drei Züge").

## Lage

Befund A-09 aus dem Audit zur Produktionsreife 0.9.0
(Zweitblick (Entwicklungsarchiv)): Bündel, Dienst und
Dokumenttyp tragen den Namensraum der Vorlage, aus der Wails ein Programm
anlegt — `com.wails.MagicSetEditor4Mac` (`build/darwin/Info.plist`),
`com.wails.MagicSetEditor4Mac.dienst` (`internal/app/schnittstellenziele.go`),
`com.wails.MagicSetEditor4Mac.umschlag` (der exportierte Typ für
`.msepackage`). Dieser Namensraum gehört nicht diesem Projekt. Jede andere
Anwendung, die aus derselben Vorlage entsteht und ihren Namen gleich wählt,
träfe dieselben Vorgaben, dasselbe Etikett im `launchd` und denselben
WebKit-Speicher.

Zur Wahl standen nach der Rückfrage zwei Formen:

1. **`io.github.davidtheider.magicseteditor4mac`** — der Namensraum hängt am
   Konto auf GitHub, und das Repositorium wird dort öffentlich
   ([0190](0190-das-repositorium-wird-oeffentlich.md)); eine Domain braucht es
   nicht.
2. **Eine eigene Domain** in umgekehrter Schreibweise — nur tragfähig, wenn
   die Domain dauerhaft gehalten wird.

## Entscheidung

**Variante 1.** Daraus folgen vier Kennungen, alle klein geschrieben, wie der
Auftraggeber sie gewählt hat:

| Was | Heute | Künftig |
| --- | --- | --- |
| Bündel | `com.wails.MagicSetEditor4Mac` | `io.github.davidtheider.magicseteditor4mac` |
| Abnahmebündel (`scripts/bauen.sh abnahme`) | `com.wails.MagicSetEditor4Mac.abnahme` | `io.github.davidtheider.magicseteditor4mac.abnahme` |
| Etikett des LaunchAgents | `com.wails.MagicSetEditor4Mac.dienst` | `io.github.davidtheider.magicseteditor4mac.dienst` |
| Exportierter Typ für `.msepackage` | `com.wails.MagicSetEditor4Mac.umschlag` | `io.github.davidtheider.magicseteditor4mac.umschlag` |

**Was nicht umzieht**: der Name des Programms (`MagicSetEditor4Mac`), der
Ort des Bestands (`~/Library/Application Support/MagicSetEditor4Mac/`),
der Name, unter dem sich der Werkzeugserver bei Klienten einträgt
(`magicseteditor4mac`), und die Einträge des Schlüsselbunds. Keiner davon hängt
an der Bündelkennung.

## Was der Umzug kostet

- **Vorgaben, WebKit-Speicher, gesicherter Fensterzustand** liegen unter der
  Kennung (`~/Library/Preferences/<kennung>.plist`, `~/Library/WebKit/…`,
  `~/Library/HTTPStorages/…`, `~/Library/Saved Application State/…`). Sie
  ziehen **einmal** um, beim ersten Start der neuen Fassung, und nur, was es
  dort gibt. Wie, misst R02 an zwei isolierten Heimatverzeichnissen.
- **Ein eingerichteter Dienst** trägt das alte Etikett. Der Umzug lädt ihn
  unter dem neuen und entlädt den alten, und danach läuft **genau einer**.
- **Erlaubnisse aus „Datenschutz & Sicherheit" gehen nicht mit** — lokales
  Netz, Kontakte. macOS bindet sie an die Kennung und fragt neu. Das ist eine
  Eigenschaft des Systems und steht in der Release-Notiz, nicht in einer
  Fußnote.

## Verworfen

- **Bei `com.wails.*` bleiben.** Der Satz im Kopf von `Info.plist` — eine
  neue Kennung hieße für jeden Nutzer eine neue Anwendung — ist richtig, und
  er ist der Grund, **jetzt** umzuziehen: vor der ersten Freigabe an
  Anwendende über einen öffentlichen Kanal (0190) ist die Zahl der Betroffenen
  am kleinsten.
- **Eine eigene Domain.** Es gibt keine, die dauerhaft gehalten wird. Eine
  Kennung, deren Namensraum verfällt, ist dieselbe Lage wie heute.
- **Die Schreibweise `io.github.DavidTHeider.MagicSetEditor4Mac`.** Gewählt
  ist die kleine. **Heute stehen beide Schreibweisen auf der Platte** —
  nachgesehen am 2026-10-15 auf dem Rechner, an dem gebaut wird:
  `~/Library/Preferences/com.wails.magicseteditor4mac.plist` klein und
  `com.wails.MagicSetEditor4Mac.abnahme.plist` gemischt. Auf APFS ohne
  Unterscheidung der Schreibweise ist das derselbe Pfad, ein Etikett im
  `launchd` ist es nicht. Der Umzug in R02 sucht deshalb beide.

## Folgen

- **R02** (#2334)
  zieht um: Kennungen, Umzug der Vorgaben, Wechsel des LaunchAgents, ein
  Wächter, der `com.wails.*` als **eigene** Kennung abweist und als **alte**
  erkennt (`internal/service/schnittstellenjson.go`), und der Rückweg zum
  alten Bündel.
- **Die Abnahme braucht ein frisches macOS** — Erlaubnisse und LaunchServices
  lassen sich auf dem Rechner, an dem gebaut wird, nicht unberührt prüfen.
  Sie läuft im Gast (`lume`), mit künstlichen Daten.

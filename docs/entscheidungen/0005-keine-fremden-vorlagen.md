# 0005 — Keine fremden Vorlagen im Repository

**Stand:** angenommen
**Datum:** 2026-08-08

## Lage

Ein Karteneditor ohne Vorlagen ist ein leeres Fenster. Die Aufgabe nannte
[Full-Magic-Pack](https://github.com/MagicSetEditorPacks/Full-Magic-Pack) als
Saat — 528 Pakete, praktisch alles, was es für MSE öffentlich gibt.

Nachgesehen am 2026-08-08:

| Befund | Zahl |
| --- | --- |
| Pakete in `data/` | 528 |
| Belegter Platz | ~1,7 GB |
| Lizenzdatei | **keine** |
| Fremde Windows-Binärdateien | `mse.exe`, `magicseteditor.exe`, `unins000.exe` |

## Entscheidung

**Fremde Vorlagensammlungen werden nicht eingecheckt und nicht mitgeliefert.**

Stattdessen zweigleisig:

1. **Ein Paketverwalter** (Phase 7) bezieht Vorlagen auf dem Rechner der
   Nutzenden — aus einer vorhandenen MSE2-Installation, aus einem Archiv oder
   von einer angegebenen Adresse. Bezogen wird von der jeweiligen Quelle, nicht
   über uns.
2. **Eine eigene Saat** unter `saat/`: kleine, selbst erstellte Vorlagen ohne
   geschützte Grafiken. Genug, um die Anwendung ohne fremdes Material zu benutzen
   und zu prüfen.

## Verworfen

**Full-Magic-Pack als Submodul oder Kopie einchecken.** Verworfen, dreifach —
jeder Grund für sich hinreichend:

- **Keine Lizenz.** Ohne Lizenz gilt das Urheberrecht in seiner Vollform.
  Weiterverbreitung ist nicht gestattet; ein „ist doch öffentlich auf GitHub" ist
  keine Erlaubnis.
- **Geschützte Inhalte.** Kartenrahmen, Mana-Symbole und Schriften von Wizards of
  the Coast. Sie in ein MIT-Repository zu legen, hieße, sie unter MIT zu stellen —
  was wir nicht dürfen.
- **Fremde ausführbare Dateien.** Drei Windows-Binärdateien unbekannter Herkunft
  gehören in kein Repository, das jemand klont.

**Nur die Vorlagen ohne die Grafiken einchecken.** Verworfen. Eine `.mse-style`
ohne ihre Bilder rendert nichts; man hätte den Lizenzbruch vermieden und nichts
gewonnen.

**Die Pakete beim ersten Start automatisch herunterladen.** Verworfen. Ein
Programm, das ungefragt über ein Gigabyte lädt, überrascht seine Nutzenden — und
es verlagert nur die Frage, ob der Bezug zulässig ist, von uns auf sie, ohne sie
zu beantworten. Der Paketverwalter lädt auf **ausdrückliche** Anweisung.

## Folgen

- Ohne Zutun startet die Anwendung mit der Saat aus `saat/`. Sie muss also gut
  genug sein, um etwas damit zu tun — nicht nur, um nicht leer zu sein.
- Die Tests von `msepaket` laufen gegen selbst geschriebene Minimalpakete unter
  `testdaten/`, nicht gegen fremde. Wer gegen ein bezogenes Paket prüfen will,
  tut das hinter dem Build-Tag `fremdpakete`; das läuft nicht in der Pipeline.
- Der Paketverwalter braucht Prüfsummen und eine sichtbare Herkunftsangabe je
  Paket. Wer nicht weiß, woher eine Vorlage kommt, kann ihre Rechtslage nicht
  einschätzen.
- `DRITTANBIETER.md` §3 hält denselben Sachverhalt für Außenstehende fest.

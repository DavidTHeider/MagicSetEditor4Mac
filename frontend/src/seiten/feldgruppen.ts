// **Felder einer Reihe stehen beieinander** — Werkstück 3 von Phase 3.6
// (#1167).
//
// # Der Befund
//
// Die Massenänderung bietet rund hundert Felder in **einer** Liste an. Darin
// stehen `Name`, `Name 2`, `Name 3` und `Name 4` als vier gleichrangige
// Einträge — und wer die Rückseite meint, muss wissen, dass sie die zweite
// Stelle ist. Bei einer zweiseitigen Vorlage ist genau das die Frage.
//
// # Warum die Regel nicht hier steht
//
// **Sie steht in Go** (`domain.Game.Feldreihen`), und `reihe`/`stelle` kommen
// fertig an der Fassade an. Eine zweite Fassung in TypeScript liefe an dem Tag
// auseinander, an dem jemand eine davon ändert (§2).
import type { FeldbeschreibungDTO } from '../api'

// Feldgruppe ist entweder ein einzelnes Feld oder eine Reihe.
export interface Feldgruppe {
  // reihe ist das Grundfeld — leer bei einem Feld ohne Geschwister.
  reihe: string

  // aufschrift ist die des Grundfeldes — die Überschrift der Gruppe.
  aufschrift: string

  felder: FeldbeschreibungDTO[]
}

// gruppiereFelder fasst aufeinanderfolgende Felder derselben Reihe zusammen.
//
// # Warum aufeinanderfolgende und nicht alle
//
// **Weil die Reihenfolge die des Spielpakets ist**, und die ist die, in der
// MSE die Felder zeigt. Wer über die ganze Liste sammelte, zöge ein `Name 4`
// aus der Mitte nach oben — und änderte damit eine Ordnung, die aus der Datei
// kommt. In `magic.mse-game` stehen die Glieder einer Reihe ohnehin
// beieinander; wo nicht, bleibt es bei zwei Gruppen, und das ist die
// ehrlichere Antwort.
//
// **Eine Gruppe aus einem Feld ist keine.** Ein Feld, dessen Geschwister die
// Liste gar nicht anbietet — weil sie nicht editierbar sind —, steht für sich
// und bekommt keine Überschrift.
export function gruppiereFelder(
  felder: readonly FeldbeschreibungDTO[],
): Feldgruppe[] {
  const gruppen: Feldgruppe[] = []

  for (const feld of felder) {
    const letzte = gruppen.at(-1)

    // **`letzte !== undefined` steht ausdrücklich da.** Ein `letzte?.reihe ===
    // feld.reihe` ist bei zwei fehlenden Werten wahr — und griff dann auf die
    // Gruppe zu, die es noch nicht gibt. Gefunden an einer Vorrichtung, die
    // `reihe` gar nicht setzte: der Fall entsteht nicht aus dem Betrieb, aber
    // der Zweig ist derselbe (§3).
    if (
      letzte !== undefined &&
      feld.reihe !== '' &&
      letzte.reihe === feld.reihe
    ) {
      letzte.felder.push(feld)

      continue
    }

    gruppen.push({
      reihe: feld.reihe,
      aufschrift: feld.stelle === 1 ? aufschriftVon(feld) : feld.reihe,
      felder: [feld],
    })
  }

  // **Was allein geblieben ist, verliert seine Reihe wieder.** Sonst stünde
  // eine Überschrift über einem einzigen Eintrag, und die Liste wäre länger
  // ohne mehr zu sagen.
  return gruppen.map((gruppe) =>
    gruppe.felder.length > 1 ? gruppe : { ...gruppe, reihe: '' },
  )
}

// aufschriftVon gibt die Aufschrift eines Feldes — sonst seinen Namen.
//
// **Die Aufschrift und nicht der Bezeichner** (§9): das Paket nennt beides,
// und `caption:` ist das, was ein Mensch lesen soll.
export function aufschriftVon(feld: FeldbeschreibungDTO): string {
  return feld.aufschrift === '' ? feld.name : feld.aufschrift
}

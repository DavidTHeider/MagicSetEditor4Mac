// **Welche Eingabe ohne Knopf schreibt, und ob sie den Leser kennt** — #2290
// (#2290).
//
// Neben Knöpfen (`schreibknoepfe.test.ts`) und Menüzeilen
// (`schreibeintraege.test.ts`) schreiben Eingaben selbst — beim Verlassen
// oder beim Wählen. **Was dieser Wächter hält**: jede genannte Datei fragt
// den Leser **samt Grund** über `useLesegrund` (H01, #2255: ein Vorleser
// hört sonst »gesperrt« und nicht, warum), jede offene fragt gar nicht. **Was er nicht hält**: ob die Tafel
// vollständig ist, und ob die Sperre an **jeder** Eingabe der Datei greift —
// das zeigen `feldeingabeleser.test.tsx` und `eingabenleser.test.tsx`.
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const quelle = (datei: string) => readFileSync(join(__dirname, datei), 'utf8')

const gesperrt: readonly (readonly [string, string])[] = [
  [
    'seiten/Feldeingabe.tsx',
    'Kartenfelder und Setangaben — schreiben beim Verlassen oder Wählen',
  ],
  ['seiten/Kartennotizen.tsx', 'die Notiz einer Karte'],
  ['seiten/Schluesselwortfeld.tsx', 'Ein Schlüsselwort ändern'],
  ['seiten/Stiloption.tsx', 'Stiloptionen lesen und setzen'],
  ['seiten/Autoersetzungen.tsx', 'Eine Autoersetzungsregel setzen'],
  [
    'seiten/Wortlistenwahl.tsx',
    'ein Wort aus der Liste — schreibt beim Wählen',
  ],
  [
    'seiten/Symbolpalette.tsx',
    'ein Symbol in den Entwurf — übernimmt ihn in die Vorschau',
  ],
]

const offen: readonly (readonly [string, string])[] = [
  [
    'seiten/Wahlfreitext.tsx',
    'eine Einstellung — ihr Träger ist je Ebene Profil oder Mensch',
  ],
  ['bausteine/Quelltextfeld.tsx', 'Paketdateien gehören dem Rechner'],
]

describe('Selbstschreibende Eingaben für einen Leser', () => {
  it.each(gesperrt)('%s fragt den Leser samt Grund (%s)', (datei) => {
    expect(quelle(datei)).toMatch(/useLesegrund\(\)/)
    expect(quelle(datei)).not.toMatch(/useNurLesend\(\)/)
  })

  it.each(offen)('%s bleibt offen (%s)', (datei) => {
    expect(quelle(datei)).not.toMatch(/useNurLesend\(\)|useLesegrund\(\)/)
  })
})

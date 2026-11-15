// **Was über dem Rahmen aufgeht** — die Befehlspalette und die
// Kommandozentrale.
//
// **Aus `Rahmen.tsx` herausgezogen**, als die Leiste oben dazukam
// (UX-14.AK-02, #1686)
// und die Datei an der Zeilengrenze stand (§2). **Der Zustand bleibt im
// Rahmen** — ⌘K öffnet von dort —; hier steht, was er zeigt.
import type { ComponentProps } from 'react'

import { Kommandozentrale } from './Kommandozentrale'
import { Palette } from './Palette'
import type { BefehlDTO } from '../apibefehle'
import type { Seitenschluessel } from '../navigation'

interface Rahmendialogeigenschaften {
  paletteOffen: boolean
  befehl: BefehlDTO | undefined
  schliessePalette: () => void
  waehlen: (seite: Seitenschluessel) => void
  waehleBefehl: (befehl: BefehlDTO) => void
  schliesseBefehl: () => void
  wege?: ComponentProps<typeof Palette>['wege']
}

export function Rahmendialoge({
  paletteOffen,
  befehl,
  schliessePalette,
  waehlen,
  waehleBefehl,
  schliesseBefehl,
  wege,
}: Rahmendialogeigenschaften) {
  return (
    <>
      {/* **Abgebaut statt ausgeblendet**: eine Palette, die nur `null`
          zurückgibt, behält ihren Zustand und zeigt beim nächsten Öffnen die
          alte Suche. */}
      {paletteOffen && (
        <Palette
          schliessen={schliessePalette}
          waehlen={waehlen}
          waehleBefehl={waehleBefehl}
          wege={wege}
        />
      )}
      {befehl !== undefined && (
        <Kommandozentrale befehl={befehl} schliessen={schliesseBefehl} />
      )}
    </>
  )
}

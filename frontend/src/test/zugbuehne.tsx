// **Die Schritte, wie sie im Raum stehen** — P13
// (#2381):
// `Handelsschritte` und darunter die Meldung des Zugs, der dem Raum gehört.
// Geteilt von `handelsschrittezug.test.tsx` und
// `handelsschrittezugfehler.test.tsx`.
import { render } from '@testing-library/react'
import { vi } from 'vitest'

import { Handelsschritte } from '../bausteine/Handelsschritte'
import { Zugmeldung } from '../bausteine/Zugmeldung'
import { useHandelszug } from '../haken/useHandelszug'
import { useVorgang } from '../haken/useVorgang'
import { du, einStand, ich } from './handelszugprobe'
import { Sprachraum } from '../sprache/kontext'

import type { Handelsvorgang as Vorgangsstand } from '../apihandelsleitung'

function Buehne({
  stand,
  geaendert,
  gebucht,
}: {
  stand: Vorgangsstand
  geaendert: (neu: Vorgangsstand) => void
  gebucht: () => void
}) {
  const lauf = useVorgang()
  const zug = useHandelszug(du, lauf)

  // **Wie im Raum**: die Meldung des Zugs steht unter den Schritten.
  return (
    <Sprachraum sprache="de">
      <Handelsschritte
        kennung={du}
        stand={stand}
        geaendert={geaendert}
        gebucht={gebucht}
        zug={zug}
        lauf={lauf}
      />
      <Zugmeldung
        ausgang={zug.ausgang}
        name="Bert"
        angehalten={zug.laeuft}
        nochmal={(bisher, legen) => {
          void zug.nochmal(bisher, legen)
        }}
      />
    </Sprachraum>
  )
}

export function zeige(stand: Vorgangsstand) {
  const geaendert = vi.fn()
  const gebucht = vi.fn()
  render(<Buehne stand={stand} geaendert={geaendert} gebucht={gebucht} />)

  return { geaendert, gebucht }
}

export const vorbereitet = einStand({
  zustand: 'vorbereitet',
  beschluss: 'uebergeben',
  zustimmungen: [du],
  naechster: 'uebergeben',
  moegliche: ['uebergeben', 'abbrechen'],
  phase: 'zusagen',
})

export const folge = {
  uebergeben: einStand({ zustand: 'uebergabeOffen', naechster: 'verbuchen' }),
  verbuchen: einStand({ zustand: 'lokalVerbucht', naechster: 'abgleichen' }),
  abgleichen: einStand({
    zustand: 'abgleichOffen',
    beschluss: 'bestaetigen',
    naechster: 'bestaetigen',
  }),
  [`bestaetigen:${ich}`]: einStand({
    zustand: 'abgleichOffen',
    beschluss: 'bestaetigen',
    zustimmungen: [ich],
    lage: 'wartet',
  }),
}

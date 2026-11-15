// **Wer jetzt entscheidet** — P7
// (#2375),
// Befund C-05: der Sitz, an den der Vorhang das Gerät weitergibt. Im Beginn
// der nächste, der noch nicht behalten hat (CR 103.5), danach wer Auslöser
// ordnet (CR 603.3b), eine Legende wählt (CR 704.5j), abwirft (CR 514.1) oder
// Priorität hat (CR 117.1) — alles Fassung 2026-09-25.
import { describe, expect, it } from 'vitest'

import type { Partie, Sitz } from '../apipartie'
import { einePartie, einSitz, einZug } from '../seiten/partieprobe'
import { entscheider } from './entscheider'

function beginn(fuer: string, ...stande: Sitz['eroeffnung'][]): Partie {
  return einePartie({
    fuer,
    sitze: stande.map((eroeffnung, stelle) =>
      einSitz({ kennung: `sitz${String(stelle + 1)}`, eroeffnung }),
    ),
  })
}

describe('wer entscheidet', () => {
  it('ist im Beginn der gezeigte Sitz, solange er nicht behalten hat', () => {
    expect(entscheider(beginn('sitz1', 'offen', 'offen'))).toBe('sitz1')
    expect(entscheider(beginn('sitz1', 'ausstehend', 'offen'))).toBe('sitz1')
  })

  it('ist danach der nächste in der Reihe, der noch entscheidet', () => {
    expect(entscheider(beginn('sitz1', 'behalten', 'offen'))).toBe('sitz2')
    expect(entscheider(beginn('sitz2', 'offen', 'behalten', 'behalten'))).toBe(
      'sitz1',
    )
    expect(entscheider(beginn('sitz2', 'behalten', 'behalten', 'offen'))).toBe(
      'sitz3',
    )
  })

  // **Der Zuschauer steht vor dem ersten Sitz** — die Reihe beginnt dort.
  it('beginnt für den Zuschauer beim ersten Sitz', () => {
    expect(entscheider(beginn('', 'behalten', 'offen'))).toBe('sitz2')
  })

  it('ist nach dem Beginn, wer Priorität hat', () => {
    const partie = einePartie({
      zug: einZug({ nummer: 1, aktiv: 'sitz1', prioritaet: 'sitz2' }),
    })

    expect(entscheider(partie)).toBe('sitz2')
  })

  it('geht vor der Priorität an den, der Auslöser ordnet', () => {
    const partie = einePartie({
      ausloeserreihung: 'sitz2',
      zug: einZug({ nummer: 1, aktiv: 'sitz1', prioritaet: 'sitz1' }),
    })

    expect(entscheider(partie)).toBe('sitz2')
  })

  it('geht vor der Priorität an den, der eine Legende wählt', () => {
    const partie = einePartie({
      legendenwahlen: [{ sitz: 'sitz2', name: 'Isamaru' }],
      zug: einZug({ nummer: 1, aktiv: 'sitz1', prioritaet: 'sitz1' }),
    })

    expect(entscheider(partie)).toBe('sitz2')
  })

  it('ist beim Abwerfen der aktive Sitz', () => {
    const partie = einePartie({
      zug: einZug({ nummer: 3, aktiv: 'sitz2', prioritaet: '', abwurf: 2 }),
    })

    expect(entscheider(partie)).toBe('sitz2')
  })

  it('ist niemand im Sandkasten ohne Zug', () => {
    expect(entscheider(einePartie())).toBeUndefined()
  })

  it('ist niemand nach dem Ende', () => {
    const partie = einePartie({
      stand: 'beendet',
      zug: einZug({ nummer: 4, aktiv: 'sitz1', prioritaet: 'sitz1' }),
    })

    expect(entscheider(partie)).toBeUndefined()
  })
})

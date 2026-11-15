// **Eine Lage ist kein Befehl** — PLAY-04.AK-03
// (#1712).
//
// # Warum hier und nicht an der Seite
//
// **Weil ein gezogener Zeiger in einem Lauf ohne Grafikkarte nicht
// ankommt.** Ein erster Anlauf zog am `Partietisch` — und der Steuerlauf
// „legt nichts ab, solange der Tisch selbst ordnet" war **grün, weil gar
// nichts geschah**: ohne WebGL weicht der Tisch auf die Knopfliste aus, und
// die Zeigerkette liegt still. Aufgefallen ist es erst an einem zweiten
// Steuerlauf („zieht überhaupt"), der ebenfalls rot wurde.
//
// **Das ist die Lehre aus §10 in Reinform**: ein Lauf, der grün ist, weil
// sein Ereignis ausfällt, prüft nichts — und sagt es nicht.
import { describe, expect, it, vi } from 'vitest'

import { zugBehandeln } from './tischzug'

function mitSpionen() {
  return { lege: vi.fn(), befehle: vi.fn(), waehle: vi.fn() }
}

describe('zugBehandeln', () => {
  it('legt eine Lage ab, statt zu befehlen', () => {
    const handlung = mitSpionen()

    zugBehandeln({ art: 'ablegen', kennung: 'k', x: 10, y: 20 }, handlung)

    expect(handlung.lege).toHaveBeenCalledWith('k', { x: 10, y: 20 })
    expect(handlung.befehle).not.toHaveBeenCalled()
  })

  // **Eine Geste deutet der Kern** — P5 (#2373) über P3 (#2371): sie geht
  // als `objekt.ziehen` hinaus, mit dem Ziel-Sitz, der bis dahin wegfiel.
  it('reicht einen Zug in eine andere Zone als Geste an den Wirt', () => {
    const handlung = mitSpionen()

    zugBehandeln(
      { art: 'zone', kennung: 'k', zone: 'friedhof', sitz: 'ich' },
      handlung,
    )

    expect(handlung.befehle).toHaveBeenCalledWith('objekt.ziehen', {
      objekt: 'k',
      zone: 'friedhof',
      sitz: 'ich',
    })
    expect(handlung.lege).not.toHaveBeenCalled()
  })

  // **⌥ heißt frei** — ausdrücklich schieben statt deuten lassen.
  it('nennt einen freien Zug frei', () => {
    const handlung = mitSpionen()

    zugBehandeln(
      { art: 'zone', kennung: 'k', zone: 'tisch', sitz: 'ich', frei: true },
      handlung,
    )

    expect(handlung.befehle).toHaveBeenCalledWith('objekt.ziehen', {
      objekt: 'k',
      zone: 'tisch',
      sitz: 'ich',
      frei: true,
    })
  })

  it('häuft über den Wirt', () => {
    const handlung = mitSpionen()

    zugBehandeln({ art: 'haufen', kennung: 'k', worauf: 'z' }, handlung)

    expect(handlung.befehle).toHaveBeenCalledWith('objekt.haeufen', {
      objekt: 'k',
      worauf: 'z',
    })
  })

  it('enthäuft über den Wirt', () => {
    const handlung = mitSpionen()

    zugBehandeln({ art: 'enthaufen', kennung: 'k' }, handlung)

    expect(handlung.befehle).toHaveBeenCalledWith('objekt.enthaeufen', {
      objekt: 'k',
    })
  })

  // **Nach jedem Zug ist die Auswahl weg** (§7, Regel 1): sie zeigte auf
  // einen Stand, den es nicht mehr gibt.
  it('räumt die Auswahl, was immer der Zug war', () => {
    for (const ziel of [
      { art: 'ablegen', kennung: 'k', x: 1, y: 2 },
      { art: 'enthaufen', kennung: 'k' },
    ] as const) {
      const handlung = mitSpionen()

      zugBehandeln(ziel, handlung)

      expect(handlung.waehle).toHaveBeenCalledWith(undefined)
    }
  })
})

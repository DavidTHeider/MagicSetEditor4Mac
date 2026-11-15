// **Der Weg vom Ausschnitt zum Bild** — PLAY-05.AK-03 und AK-04
// (#1713).
import { describe, expect, it } from 'vitest'

import { abzugMachen } from './tischabzug'
import { einMaler } from '../test/malerprobe'
import { einePartie } from '../seiten/partieprobe'
import { tischplan } from './tischplan'
import { zweiSitze } from '../test/planprobe'

const plan = tischplan(einePartie(zweiSitze))

const kamera = {
  mitteX: plan.breite / 2,
  mitteY: plan.hoehe / 2,
  massstab: 1,
  drehung: 0,
}

const flaeche = { breite: 800, hoehe: 600 }

describe('abzugMachen', () => {
  it('gibt das Bild samt seinen Maßen', async () => {
    const maler = einMaler(flaeche)

    const ergebnis = await abzugMachen(
      { art: 'da', maler },
      plan,
      kamera,
      flaeche,
      { art: 'tisch' },
      150,
    )

    // **Die Maße stehen als Zahlen da und nicht als Platzhalter**: 150 dpi
    // über den ganzen Tisch samt Rand sind eine feste Rechnung, und ein
    // `expect.any(Number)` sagte über sie nichts.
    expect(ergebnis).toEqual({
      bild: 'data:image/png;base64,AAA',
      breite: Math.round(((plan.breite + 8) / 25.4) * 150),
      hoehe: Math.round(((plan.hoehe + 8) / 25.4) * 150),
    })

    // **Der Ausschnitt geht in Millimetern hinein, die Auflösung in Punkten
    // je Millimeter** — 150 dpi sind 150/25,4.
    expect(maler.abzug).toHaveBeenCalledWith(
      expect.objectContaining({ x: -4, y: -4 }),
      150 / 25.4,
    )
  })

  // **Ohne Grafikkarte gibt es kein Bild** — und das ist ein eigener Befund,
  // kein leeres Ergebnis: die Zonenliste daneben steht weiter, und wer den
  // Abzug will, soll hören, warum es ihn nicht gibt (§8).
  it('meldet, wo nicht gezeichnet wird', async () => {
    expect(
      await abzugMachen(
        { art: 'ohne', grund: 'kein-webgl' },
        plan,
        kamera,
        flaeche,
        { art: 'tisch' },
        150,
      ),
    ).toEqual({ fehler: 'keinMaler' })
  })

  // **Eine Zone, die dieser Tisch nicht führt, gibt nichts** — und keinen
  // weißen Abzug, den jemand für einen leeren Tisch hielte.
  it('meldet eine Zone, die es nicht gibt', async () => {
    expect(
      await abzugMachen(
        { art: 'da', maler: einMaler(flaeche) },
        plan,
        kamera,
        flaeche,
        { art: 'zone', zone: 'ante', sitz: 'ich' },
        150,
      ),
    ).toEqual({ fehler: 'keinAusschnitt' })
  })

  // **Und eine Auflösung außerhalb der Grenzen kommt gar nicht erst zur
  // Grafikkarte** — AK-04: „DPI-Grenzen sind begrenzt und fehlerrobust".
  it('fragt die Grafikkarte nicht, wo die Maße nicht gehen', async () => {
    const maler = einMaler(flaeche)

    expect(
      await abzugMachen(
        { art: 'da', maler },
        plan,
        kamera,
        flaeche,
        { art: 'tisch' },
        1200,
      ),
    ).toEqual({ fehler: 'dpi' })

    expect(maler.abzug).not.toHaveBeenCalled()
  })
})

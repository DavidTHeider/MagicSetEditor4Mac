// **Die zwei Wege, die der Haken zu einem Ziel kennt** — PLAY-05.AK-05
// (#1713).
//
// **Eine eigene Datei**, weil `useTisch.test.ts` bei 270 Zeilen steht und die
// Zeilengrenze aus §2 keine Näherung kennt. Der Schnitt folgt der Frage: dort,
// **dass** der Haken zeichnet und aufräumt — hier, **worauf** er zielt.
//
// `kamerazielen.test.ts` hält die Rechnung; hier steht die Verdrahtung: dass
// der Knopf die Kamera wirklich versetzt, und dass er nichts tut, wo es das
// Ziel nicht gibt.
import { act, renderHook, waitFor } from '@testing-library/react'
import { describe, expect, it } from 'vitest'

import { einObjekt, einePartie, einSitz } from '../seiten/partieprobe'
import { tischplan } from './tischplan'
import { useTisch } from './useTisch'

import type { Tischanschluss } from './useTisch'
import type { Tischmalerei } from './maler'
import { einMaler as einMalerOhne } from '../test/malerprobe'

const einPlan = (fuer: string, ...teile: Parameters<typeof einObjekt>[0][]) =>
  tischplan(
    einePartie({
      sitze: [
        einSitz({ kennung: 'ich', name: 'Ich', deckabdruck: 'a', karten: 2 }),
        einSitz({ kennung: 'du', name: 'Du', deckabdruck: 'b', karten: 2 }),
      ],
      fuer,
      objekte: teile.map((eines) =>
        einObjekt({ eigner: 'ich', beherrscher: 'ich', ...eines }),
      ),
    }),
  )

// haenge hängt den Haken an eine echte Leinwand und einen Maler.
const haenge = async (
  plan: ReturnType<typeof tischplan>,
  gewaehlt?: string,
) => {
  const maler = einMalerOhne({ breite: 800, hoehe: 600 })
  const anschluss: Tischanschluss = {
    plan,
    bild: () => undefined,
    malerei: () => Promise.resolve({ art: 'da', maler } as Tischmalerei),
    leinwand: { current: document.createElement('canvas') },
    ...(gewaehlt === undefined ? {} : { gewaehlt }),
  }

  const lauf = renderHook(() => useTisch(anschluss))

  await waitFor(() => {
    expect(maler.zeichne).toHaveBeenCalled()
  })

  return lauf
}

describe('zeigeAuswahl', () => {
  it('holt die gewählte Karte heran', async () => {
    const lauf = await haenge(
      einPlan('ich', { kennung: 'a', zone: 'tisch' }),
      'a',
    )
    const vorher = lauf.result.current.kamera

    act(() => {
      lauf.result.current.zeigeAuswahl()
    })

    expect(lauf.result.current.kamera.massstab).toBeGreaterThan(vorher.massstab)
  })

  // **Ohne Auswahl bewegt sich nichts** — und das ist die Zusage, nicht eine
  // Auslassung: ein Sprung auf den Ursprung wäre eine Bewegung, die niemand
  // angeordnet hat.
  it('lässt die Kamera stehen, wo nichts gewählt ist', async () => {
    const lauf = await haenge(einPlan('ich', { kennung: 'a', zone: 'tisch' }))
    const vorher = lauf.result.current.kamera

    act(() => {
      lauf.result.current.zeigeAuswahl()
    })

    expect(lauf.result.current.kamera).toEqual(vorher)
  })
})

describe('zeigeEigenenSitz', () => {
  // **Die Zusage ist die Mitte und nicht der Maßstab.**
  //
  // Der erste Anlauf dieses Laufs verlangte einen größeren Maßstab und wurde
  // rot — zu Recht: eine Sitzfläche ist **1200 × 300** und damit **so breit
  // wie der ganze Tisch**. Bei zwei Sitzen begrenzt beide Male dieselbe
  // Breite, und der Maßstab ist gleich. Was sich ändert, ist der Blickpunkt.
  it('rückt den eigenen Sitz in die Mitte', async () => {
    const plan = einPlan('ich')
    const lauf = await haenge(plan)
    const eigene = plan.sitze.find((eine) => eine.eigen)

    if (!eigene) throw new Error('kein eigener Sitz')

    act(() => {
      lauf.result.current.zeigeEigenenSitz()
    })

    expect(lauf.result.current.kamera.mitteY).toBeCloseTo(
      eigene.y + eigene.hoehe / 2,
      6,
    )
    expect(lauf.result.current.kamera.mitteY).not.toBeCloseTo(plan.hoehe / 2, 1)
  })

  // **Ein Zuschauer hat keinen eigenen Sitz** — `sitzflaeche.ts` setzt `eigen`
  // dann nirgends.
  it('lässt die Kamera stehen, wo niemand einen eigenen Sitz hat', async () => {
    const lauf = await haenge(einPlan('zuschauer'))
    const vorher = lauf.result.current.kamera

    act(() => {
      lauf.result.current.zeigeEigenenSitz()
    })

    expect(lauf.result.current.kamera).toEqual(vorher)
  })
})

// **Ein Zug endet, wo der Zeiger losgelassen wird — auch neben dem Tisch** —
// Befund 3 des Zweitblicks über Phase 6.9 (PLAY-05.AK-04,
// #1713).
//
// # Der Befund
//
// Das Tischbild nahm den Szenenbaum, wie er gerade stand — und der zeigt
// während eines Zugs den gezogenen Plan. Nachgesehen fand sich eine zweite
// Lücke darunter: die Fläche hörte nur auf ihr **eigenes** Loslassen. Wer
// eine Karte über den Rand hinaus zog und daneben losließ, hatte einen Zug,
// der nie endete — die Karte blieb verschoben stehen, folgte beim nächsten
// Überfahren dem Zeiger ohne gedrückte Taste und ging so auch ins Tischbild.
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { eingepasst, nachPunkten } from '../tisch/kamera'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { Tischflaeche } from './Tischflaeche'
import { tischplan } from '../tisch/tischplan'
import { einMaler } from '../test/malerprobe'

import type { Tischmalerei } from '../tisch/maler'

const eigene = kataloge.de.seiten.spielen

const plan = tischplan(
  einePartie({
    objekte: [
      einObjekt({
        kennung: 'a',
        zone: 'tisch',
        name: 'Wald',
        eigner: 'sitz1',
        beherrscher: 'sitz1',
        geheim: false,
      }),
    ],
  }),
)

const lage = plan.lagen.find((eine) => eine.kennung === 'a')

// amOrt rechnet über dieselbe Kamera, die der Haken nimmt — geratene
// Bildschirmpunkte träfen zufällig etwas.
const amOrt = (x: number, y: number) => {
  const punkt = nachPunkten(
    eingepasst(plan, { breite: 800, hoehe: 600 }, 16),
    { x, y },
    { breite: 800, hoehe: 600 },
  )

  return { clientX: punkt.x, clientY: punkt.y }
}

async function zeige() {
  if (!lage) throw new Error('keine Lage')

  const maler = einMaler({ breite: 800, hoehe: 600 })
  const aufZug = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Tischflaeche
        plan={plan}
        bild={() => undefined}
        malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        aufZug={aufZug}
      />
    </Sprachraum>,
  )

  const flaeche = await screen.findByRole('application')
  flaeche.getBoundingClientRect = () =>
    ({ left: 0, top: 0, width: 800, height: 600 }) as DOMRect

  // **Gewartet wird auf das erste Bild** — der Maler kommt asynchron, und ein
  // Vergleich davor verglich nichts mit nichts.
  await waitFor(() => {
    expect(maler.zeichne).toHaveBeenCalled()
  })

  // gezeichnetBei sagt, wo die Karte im zuletzt gezeichneten Plan liegt.
  const gezeichnetBei = () => {
    const gezeichnet = maler.zeichne.mock.lastCall?.[0].plan.lagen.find(
      (eine) => eine.kennung === 'a',
    )

    return gezeichnet ? { x: gezeichnet.x, y: gezeichnet.y } : undefined
  }

  const mitte = amOrt(lage.x + lage.breite / 2, lage.y + lage.hoehe / 2)

  // fasse zieht die Karte um 40 Millimeter nach rechts, ohne loszulassen.
  const fasse = () => {
    fireEvent.pointerDown(flaeche, { button: 0, ...mitte })
    fireEvent.pointerMove(
      flaeche,
      amOrt(lage.x + 40 + lage.breite / 2, lage.y + lage.hoehe / 2),
    )
  }

  return { maler, aufZug, flaeche, gezeichnetBei, fasse, mitte }
}

describe('ein Zug, der neben dem Tisch endet', () => {
  it('legt die Karte zurück und meldet nichts', async () => {
    const { aufZug, flaeche, gezeichnetBei, fasse, mitte } = await zeige()
    const vorher = gezeichnetBei()

    fasse()

    // **Dass jemand zuhört**: gezogen ist sie.
    expect(gezeichnetBei()).not.toEqual(vorher)

    fireEvent.pointerUp(document.body)

    expect(gezeichnetBei()).toEqual(vorher)
    expect(aufZug).not.toHaveBeenCalled()

    // **Und sie folgt dem Zeiger danach nicht mehr** — ohne gedrückte Taste
    // ist nichts gefasst.
    fireEvent.pointerMove(flaeche, {
      clientX: mitte.clientX + 90,
      clientY: mitte.clientY,
    })

    expect(gezeichnetBei()).toEqual(vorher)
  })
})

// **Escape bricht einen Zug ab** — A04
// (#2218):
// die Abnahme verlangt den Abbruch am echten Fenster, und bis dahin hörte
// der Tisch während eines Zugs auf keine Taste.
describe('ein Zug, den Escape abbricht', () => {
  it('legt die Karte zurück, und das Loslassen danach meldet nichts', async () => {
    const { aufZug, flaeche, gezeichnetBei, fasse } = await zeige()
    const vorher = gezeichnetBei()

    fasse()
    expect(gezeichnetBei()).not.toEqual(vorher)

    const taste = fireEvent.keyDown(window, { key: 'Escape' })

    expect(gezeichnetBei()).toEqual(vorher)
    // **Und die Taste ist genommen** — eine Schublade darunter schlösse
    // sonst mit (G12: das Innerste zuerst).
    expect(taste).toBe(false)

    fireEvent.pointerUp(flaeche, {
      button: 0,
      clientX: 400,
      clientY: 300,
    })

    expect(aufZug).not.toHaveBeenCalled()
  })

  it('lässt eine andere Taste den Zug nicht abbrechen', async () => {
    const { gezeichnetBei, fasse } = await zeige()
    const vorher = gezeichnetBei()

    fasse()
    const gezogen = gezeichnetBei()

    expect(fireEvent.keyDown(window, { key: 'a' })).toBe(true)
    expect(gezeichnetBei()).toEqual(gezogen)
    expect(gezogen).not.toEqual(vorher)
  })

  it('nimmt Escape ohne Zug nicht weg', async () => {
    await zeige()

    expect(fireEvent.keyDown(window, { key: 'Escape' })).toBe(true)
  })
})

describe('das Tischbild während eines Zugs', () => {
  it('entsteht nicht, und der Grund steht da', async () => {
    const { maler, fasse } = await zeige()

    fasse()

    // **Über die Tastatur ausgelöst** — mit der Maus ginge es nicht, die
    // hält gerade die Karte.
    fireEvent.click(screen.getByRole('button', { name: eigene.abzugTitel }))
    fireEvent.click(
      await screen.findByRole('button', { name: eigene.abzugErzeugen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      eigene.abzugZieht,
    )
    expect(maler.abzug).not.toHaveBeenCalled()
  })

  it('entsteht nach dem Loslassen wieder', async () => {
    const { maler, fasse } = await zeige()

    fasse()
    fireEvent.pointerUp(document.body)

    fireEvent.click(screen.getByRole('button', { name: eigene.abzugTitel }))
    fireEvent.click(
      await screen.findByRole('button', { name: eigene.abzugErzeugen }),
    )

    expect(await screen.findByAltText(eigene.abzugVorschau)).toBeInTheDocument()
    expect(maler.abzug).toHaveBeenCalled()
  })
})

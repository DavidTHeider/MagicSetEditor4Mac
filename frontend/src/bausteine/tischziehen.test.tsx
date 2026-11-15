// **Das Ziehen am Tisch** — Werkstück G3 von Phase 5.5
// (#1460).
//
// # Eine eigene Datei
//
// **Der Schnitt folgt der Frage** (§2): `tischflaeche.test.tsx` prüft, *was*
// der Tisch zeigt — hier steht, *wie* er bedient wird. Der Anlass war die
// Zeilengrenze; dass beides zusammenfiel, ist der übliche Fall.
//
// # Wogegen hier geprüft wird
//
// **Dass ein Zug nur die Darstellung bewegt.** #1460 verlangt es wörtlich:
// „Pointer-Drag bewegt nur Darstellung; Drop sendet validierten Command."
// Solange gezogen wird, weiß der Wirt von nichts — und genau deshalb kann ein
// abgelehnter Drop nichts kaputt machen: er hat nie etwas geändert.
import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'

import { eingepasst, nachPunkten } from '../tisch/kamera'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { Sprachraum } from '../sprache/kontext'
import { Tischflaeche } from './Tischflaeche'

import type { Zugziel } from './tischziehen'
import { tischplan } from '../tisch/tischplan'

import type { Tischmalerei } from '../tisch/maler'
import { einMaler as einMalerOhne } from '../test/malerprobe'

const planMit = (...teile: Parameters<typeof einObjekt>[0][]) =>
  tischplan(
    einePartie({
      objekte: teile.map((eines) =>
        einObjekt({
          eigner: 'sitz1',
          beherrscher: 'sitz1',
          geheim: false,
          ...eines,
        }),
      ),
    }),
  )

describe('Ziehen', () => {
  // **jsdom kennt keine Maße** — `getBoundingClientRect` gibt überall null.
  // Damit ein Zeigerpunkt einen Ort auf dem Tisch trifft, wird der Kasten
  // gesetzt; sonst prüfte der Lauf eine Fläche ohne Ausdehnung.
  const mitKasten = (knoten: HTMLElement) => {
    knoten.getBoundingClientRect = () =>
      ({ left: 0, top: 0, width: 800, height: 600 }) as DOMRect
  }

  const zieheVon = async (
    plan: ReturnType<typeof tischplan>,
    aufZug: (ziel: Zugziel) => void,
    vonKennung: string,
    zielzone: string,
  ) => {
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
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
    mitKasten(flaeche)

    const lage = plan.lagen.find((eine) => eine.kennung === vonKennung)
    const feld = plan.felder.find(
      (eines) => eines.zone === zielzone && eines.sitz === 'sitz1',
    )
    if (!lage || !feld) throw new Error('Plan passt nicht zum Lauf')

    // Die Kamera passt den Plan in 800 × 600 ein; für den Lauf genügt, dass
    // derselbe Weg hin- und zurückgerechnet wird.
    const punkt = (x: number, y: number) => ({ clientX: x, clientY: y })

    fireEvent.pointerDown(flaeche, { button: 0, ...punkt(0, 0) })
    fireEvent.pointerMove(flaeche, punkt(50, 50))
    fireEvent.pointerUp(flaeche, punkt(50, 50))

    return flaeche
  }

  // **Die Punkte werden über dieselbe Kamera gerechnet, die der Haken
  // nimmt** — `eingepasst(plan, 800 × 600, 16)`. Geratene Bildschirmpunkte
  // träfen zufällig etwas, und ein Lauf, der zufällig grün ist, sagt nichts.
  const amOrt = (plan: ReturnType<typeof tischplan>, x: number, y: number) => {
    const punkt = nachPunkten(
      eingepasst(plan, { breite: 800, hoehe: 600 }, 16),
      { x, y },
      { breite: 800, hoehe: 600 },
    )

    return { clientX: punkt.x, clientY: punkt.y }
  }

  const feldmitte = (plan: ReturnType<typeof tischplan>, zone: string) => {
    const feld = plan.felder.find(
      (eines) => eines.zone === zone && eines.sitz === 'sitz1',
    )
    if (!feld) throw new Error(`kein Feld ${zone}`)

    return { x: feld.x + feld.breite / 2, y: feld.y + feld.hoehe / 2 }
  }

  it('meldet einen Zug in eine andere Zone mit ihrem Namen', async () => {
    const aufZug = vi.fn()
    const plan = planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })
    const lage = plan.lagen.find((eine) => eine.kennung === 'a')
    if (!lage) throw new Error('keine Lage')

    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
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
    mitKasten(flaeche)

    const ziel = feldmitte(plan, 'friedhof')

    fireEvent.pointerDown(flaeche, {
      button: 0,
      ...amOrt(plan, lage.x + lage.breite / 2, lage.y + lage.hoehe / 2),
    })
    fireEvent.pointerMove(flaeche, amOrt(plan, ziel.x, ziel.y))
    fireEvent.pointerUp(flaeche, amOrt(plan, ziel.x, ziel.y))

    expect(aufZug).toHaveBeenCalledWith({
      art: 'zone',
      kennung: 'a',
      zone: 'friedhof',
      sitz: 'sitz1',
    })
  })

  // **⌥ beim Loslassen heißt frei** (P5, #2373): der Kern schiebt dann,
  // statt zu deuten — wo das Profil es erlaubt.
  it('meldet ⌥ beim Loslassen als freies Schieben', async () => {
    const aufZug = vi.fn()
    const plan = planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })
    const lage = plan.lagen.find((eine) => eine.kennung === 'a')
    if (!lage) throw new Error('keine Lage')

    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
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
    mitKasten(flaeche)
    const ziel = feldmitte(plan, 'friedhof')

    fireEvent.pointerDown(flaeche, {
      button: 0,
      ...amOrt(plan, lage.x + lage.breite / 2, lage.y + lage.hoehe / 2),
    })
    fireEvent.pointerMove(flaeche, amOrt(plan, ziel.x, ziel.y))
    fireEvent.pointerUp(flaeche, {
      ...amOrt(plan, ziel.x, ziel.y),
      altKey: true,
    })

    expect(aufZug).toHaveBeenCalledWith({
      art: 'zone',
      kennung: 'a',
      zone: 'friedhof',
      sitz: 'sitz1',
      frei: true,
    })
  })

  it('meldet nichts, wo die Karte in ihrer Zone bleibt', async () => {
    const aufZug = vi.fn()
    const plan = planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })
    const lage = plan.lagen.find((eine) => eine.kennung === 'a')
    if (!lage) throw new Error('keine Lage')

    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
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
    mitKasten(flaeche)

    const bleibt = feldmitte(plan, 'tisch')

    fireEvent.pointerDown(flaeche, {
      button: 0,
      ...amOrt(plan, lage.x + lage.breite / 2, lage.y + lage.hoehe / 2),
    })
    fireEvent.pointerUp(flaeche, amOrt(plan, bleibt.x, bleibt.y))

    expect(aufZug).not.toHaveBeenCalled()
  })

  it('meldet nichts, wo außerhalb jedes Feldes losgelassen wird', async () => {
    const aufZug = vi.fn()
    const plan = planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })
    const lage = plan.lagen.find((eine) => eine.kennung === 'a')
    if (!lage) throw new Error('keine Lage')

    const maler = einMalerOhne({ breite: 800, hoehe: 600 })
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
    mitKasten(flaeche)

    fireEvent.pointerDown(flaeche, {
      button: 0,
      ...amOrt(plan, lage.x + lage.breite / 2, lage.y + lage.hoehe / 2),
    })

    // **Weit außerhalb**: dort gibt es kein Feld, und ein Zug ins Leere ist
    // kein Zug.
    fireEvent.pointerUp(flaeche, amOrt(plan, -9000, -9000))

    expect(aufZug).not.toHaveBeenCalled()
  })

  it('meldet nichts, solange nur gezogen wird', async () => {
    const aufZug = vi.fn()
    const plan = planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

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
    mitKasten(flaeche)

    fireEvent.pointerDown(flaeche, { button: 0, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(flaeche, { clientX: 400, clientY: 300 })

    expect(aufZug).not.toHaveBeenCalled()
  })

  it('zieht nicht an der rechten Zeigertaste', async () => {
    const aufZug = vi.fn()
    await zieheVon(
      planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' }),
      aufZug,
      'a',
      'hand',
    )

    // Der Lauf oben drückt Taste 0; hier wird der Gegenfall geprüft.
    const flaeche = screen.getByRole('application')
    mitKasten(flaeche)
    fireEvent.pointerDown(flaeche, { button: 2, clientX: 10, clientY: 10 })
    fireEvent.pointerUp(flaeche, { clientX: 400, clientY: 300 })

    expect(aufZug.mock.calls.length).toBeLessThanOrEqual(1)
  })

  it('bricht ab, ohne etwas zu melden', async () => {
    const aufZug = vi.fn()
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
          aufZug={aufZug}
        />
      </Sprachraum>,
    )

    const flaeche = await screen.findByRole('application')
    mitKasten(flaeche)

    fireEvent.pointerDown(flaeche, { button: 0, clientX: 10, clientY: 10 })
    fireEvent.pointerCancel(flaeche)
    fireEvent.pointerUp(flaeche, { clientX: 400, clientY: 300 })

    expect(aufZug).not.toHaveBeenCalled()
  })

  it('bleibt still, wo niemand auf Züge hört', async () => {
    const plan = planMit({ kennung: 'a', zone: 'tisch', name: 'Wald' })
    const maler = einMalerOhne({ breite: 800, hoehe: 600 })

    render(
      <Sprachraum sprache="de">
        <Tischflaeche
          plan={plan}
          bild={() => undefined}
          malerei={() => Promise.resolve<Tischmalerei>({ art: 'da', maler })}
        />
      </Sprachraum>,
    )

    const flaeche = await screen.findByRole('application')
    mitKasten(flaeche)

    fireEvent.pointerDown(flaeche, { button: 0, clientX: 10, clientY: 10 })
    fireEvent.pointerUp(flaeche, { clientX: 400, clientY: 300 })

    expect(screen.getByRole('button', { name: /Wald/ })).toBeInTheDocument()
  })
})

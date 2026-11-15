// **Den Markttisch bedienen** — P12
// (#2380),
// Befunde C-14, C-18; Entwurf `Markttisch09`: links das eigene Fach, rechts
// das freigegebene Fach des Gegenübers, dazwischen „Von dir" und „Vom
// Gegenüber". Ziehen mit der Maus und Leertaste, Pfeile, ↩ führen auf
// **dieselbe** Absicht (Gestenvertrag, GESTE #2358).
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Anschlussraum } from '../anschlussraum'
import { Markttisch } from './Markttisch'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { einAbgleich } from '../test/tauschabgleich'

import type { ExemplarDTO } from '../api'
import type { Handelsposition } from '../apihandelsleitung'

const texte = kataloge.de.seiten.handeln
const tisch = texte.markttisch

afterEach(() => {
  raeumeApi()
})

const exemplar = (
  id: number,
  name: string,
  teil: Partial<ExemplarDTO> = {},
): ExemplarDTO =>
  ({
    id,
    fassung: 1,
    printing: `p-${String(id)}`,
    druckID: '',
    vorlaeufig: false,
    menge: 2,
    zustand: 'nearMint',
    sprache: 'de',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name,
    setcode: 'p12',
    nummer: String(id),
    lagerort: '',
    wunsch: false,
    bezeichnung: name,
    herkunft: { art: '', quelle: '', quellProfil: 0, uebernommen: '' },
    ...teil,
  }) as ExemplarDTO

function Buehne({
  gegenueber,
  gesetzt,
}: {
  gegenueber: string
  gesetzt: (positionen: Handelsposition[]) => void
}) {
  const [positionen, setzePositionen] = useState<Handelsposition[]>([])

  return (
    <Sprachraum sprache="de">
      <Anschlussraum
        aktuell="handeln"
        wechseln={() => {
          /* hier wird nicht gewechselt */
        }}
      >
        <Markttisch
          ich="ich-1"
          gegenueber={gegenueber}
          name="Bert"
          positionen={positionen}
          setzePositionen={(neu) => {
            gesetzt(neu)
            setzePositionen(neu)
          }}
        />
      </Anschlussraum>
    </Sprachraum>
  )
}

function zeige(wuensche: ExemplarDTO[] = []) {
  const sicht = vi.fn(() =>
    Promise.resolve({
      zeilen: [
        {
          referenz: 'r1',
          printing: 'p-sturm',
          menge: 2,
          zustand: 'excellent',
          sprache: 'en',
          finish: 'foil',
          bezeichnung: 'Sturmrufer',
        },
      ],
      fehler: null,
    }),
  )
  belegeApi({
    Sammlung: () =>
      Promise.resolve({
        exemplare: [
          exemplar(1, 'Glutläufer'),
          exemplar(2, 'Moosbär'),
          exemplar(3, 'Aschenwacht', { wunsch: true }),
          ...wuensche,
        ],
        fehler: null,
      }),
    HandelSichtAnfragen: sicht,
    // **Der Tisch gleicht beim Aufschlagen ab** (#2494) — hier passt nichts.
    Tauschabgleich: () => Promise.resolve(einAbgleich()),
    HandelPreisanhalt: () =>
      Promise.resolve({
        anhalt: { zeilen: [], seiten: [], differenzen: [], hauptwaehrung: '' },
        fehler: null,
      }),
  })
  const gesetzt = vi.fn()
  const ansicht = render(<Buehne gegenueber="du-1" gesetzt={gesetzt} />)

  return { sicht, gesetzt, ansicht }
}

const karte = (name: string) => screen.findByRole('button', { name })

describe('der Markttisch', () => {
  it('zeigt beide Fächer und fragt das Fach des Gegenübers beim Aufschlagen', async () => {
    const { sicht } = zeige()

    const mein = await screen.findByRole('region', { name: tisch.meinFach })
    expect(
      await within(mein).findByRole('button', { name: 'Glutläufer' }),
    ).toBeVisible()

    const sein = screen.getByRole('region', {
      name: tisch.seinFach.replace('{name}', 'Bert'),
    })
    expect(
      await within(sein).findByRole('button', { name: 'Sturmrufer' }),
    ).toBeVisible()
    expect(sicht).toHaveBeenCalledWith('du-1')

    // **Die Wunschliste ist eine Lasche im eigenen Fach.**
    await userEvent.click(
      within(mein).getByRole('tab', { name: tisch.wunschliste }),
    )
    expect(
      within(mein).getByRole('button', { name: 'Aschenwacht' }),
    ).toBeVisible()
  })

  it('legt mit Leertaste und ↩ ein eigenes Stück nach „Von dir"', async () => {
    const { gesetzt } = zeige()
    const hand = userEvent.setup()

    ;(await karte('Glutläufer')).focus()
    await hand.keyboard(' ')
    expect(screen.getByText(/Gegriffen/)).toBeInTheDocument()
    await hand.keyboard('{Enter}')

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ von: 'ich-1', printing: 'p-1', menge: 1 }),
    ])
  })

  it('legt per Ziehen ein fremdes Stück nach „Vom Gegenüber" — und nichts an ein falsches Ziel', async () => {
    const { gesetzt } = zeige()

    const sturm = await karte('Sturmrufer')
    const bekomme = screen.getByRole('region', {
      name: texte.positionVonGegenueber,
    })
    const gebe = screen.getByRole('region', { name: texte.positionVonIch })

    // **Ein falsches Ziel legt nichts.**
    fireEvent.pointerDown(sturm, { button: 0, pointerId: 1 })
    fireEvent.pointerUp(gebe, { pointerId: 1 })
    expect(gesetzt).not.toHaveBeenCalled()

    fireEvent.pointerDown(sturm, { button: 0, pointerId: 1 })
    fireEvent.pointerUp(bekomme, { pointerId: 1 })

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({
        von: 'du-1',
        printing: 'p-sturm',
        guete: 'excellent',
      }),
    ])

    // **Die Zeile nennt den Namen aus seinem Fach** — gefunden am nativen
    // Fenster (`p12markttisch09`): ohne Vorrat stand dort „Unbekannter
    // Druck" mit der Kennung, obwohl das Fach den Namen zeigte.
    expect(await within(bekomme).findByText('1 × Sturmrufer')).toBeVisible()
  })

  it('nimmt mit ⌘ mehrere Stücke und legt sie in einem Zug', async () => {
    const { gesetzt } = zeige()

    const glut = await karte('Glutläufer')
    const moos = await karte('Moosbär')
    fireEvent.click(glut, { metaKey: true })
    fireEvent.click(moos, { metaKey: true })
    expect(glut).toHaveAttribute('aria-pressed', 'true')
    expect(moos).toHaveAttribute('aria-pressed', 'true')

    fireEvent.pointerDown(moos, { button: 0, pointerId: 1 })
    fireEvent.pointerUp(
      screen.getByRole('region', { name: texte.positionVonIch }),
      { pointerId: 1 },
    )

    expect(gesetzt).toHaveBeenCalledTimes(1)
    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ printing: 'p-1' }),
      expect.objectContaining({ printing: 'p-2' }),
    ])
    // **Gelegt ist gelegt**: die Wahl ist danach leer.
    expect(glut).toHaveAttribute('aria-pressed', 'false')
  })

  it('zählt eine Zeile bis zu allem, was die Fächer von ihrer Variante halten', async () => {
    // **#2497**: der eigene Wunsch derselben Variante zählt mit — zwei
    // Stücke im fremden Fach und einer auf der Wunschliste sind drei.
    const { gesetzt } = zeige([
      exemplar(4, 'Sturmwunsch', {
        wunsch: true,
        printing: 'p-sturm',
        menge: 1,
        zustand: 'excellent',
        sprache: 'en',
        finish: 'foil',
      }),
    ])
    const sturm = await karte('Sturmrufer')
    const bekomme = screen.getByRole('region', {
      name: texte.positionVonGegenueber,
    })

    for (let mal = 0; mal < 4; mal++) {
      fireEvent.pointerDown(sturm, { button: 0, pointerId: 1 })
      fireEvent.pointerUp(bekomme, { pointerId: 1 })
    }

    expect(gesetzt).toHaveBeenLastCalledWith([
      expect.objectContaining({ printing: 'p-sturm', menge: 3 }),
    ])
  })

  it('verwirft einen Griff, wenn das Gegenüber wechselt', async () => {
    const { gesetzt, ansicht } = zeige()
    const hand = userEvent.setup()

    ;(await karte('Glutläufer')).focus()
    await hand.keyboard(' ')

    ansicht.rerender(<Buehne gegenueber="du-2" gesetzt={gesetzt} />)
    await act(async () => {
      await Promise.resolve()
    })
    ;(await karte('Glutläufer')).focus()
    await hand.keyboard('{Enter}')

    expect(gesetzt).not.toHaveBeenCalled()
  })

  it('lässt andere Tasten und die rechte Maustaste durch', async () => {
    const { gesetzt } = zeige()

    const glut = await karte('Glutläufer')
    fireEvent.keyDown(glut, { key: 'a' })
    fireEvent.pointerDown(glut, { button: 2, pointerId: 1 })
    fireEvent.pointerUp(
      screen.getByRole('region', { name: texte.positionVonIch }),
      { pointerId: 1 },
    )

    expect(gesetzt).not.toHaveBeenCalled()
    expect(screen.queryByText(/Gegriffen/)).toBeNull()
  })

  it('nennt eine Absage im fremden Fach und fragt auf Wunsch neu', async () => {
    zeige()
    let anfragen = 0
    belegeApi({
      Sammlung: () =>
        Promise.resolve({
          exemplare: [exemplar(4, 'Rindenwall', { zustand: 'eigenwort' })],
          fehler: null,
        }),
      HandelSichtAnfragen: () => {
        anfragen += 1

        return Promise.resolve({
          zeilen: [],
          fehler: {
            code: 'erlaubnis',
            meldung: 'Das Gegenüber gibt dir zum Handeln nichts frei.',
            schluessel: '',
            feld: '',
          },
        })
      },
      HandelPreisanhalt: () =>
        Promise.resolve({
          anhalt: {
            zeilen: [],
            seiten: [],
            differenzen: [],
            hauptwaehrung: '',
          },
          fehler: null,
        }),
    })
    cleanupUndNeu()

    expect(await screen.findByRole('alert')).toHaveTextContent('nichts frei')
    // **Ein Zustand, den der Katalog nicht kennt, steht, wie er heißt.**
    expect(screen.getByText(/eigenwort/)).toBeVisible()

    await userEvent.click(screen.getByRole('button', { name: tisch.neuFragen }))
    await vi.waitFor(() => {
      expect(anfragen).toBe(2)
    })
  })

  it('nimmt mit ⌘ eine Karte wieder heraus, und ein Klick wählt nur sie', async () => {
    zeige()

    const glut = await karte('Glutläufer')
    const moos = await karte('Moosbär')
    fireEvent.click(glut, { metaKey: true })
    fireEvent.click(moos, { metaKey: true })
    fireEvent.click(glut, { metaKey: true })
    expect(glut).toHaveAttribute('aria-pressed', 'false')
    expect(moos).toHaveAttribute('aria-pressed', 'true')

    fireEvent.click(glut)
    expect(glut).toHaveAttribute('aria-pressed', 'true')
    expect(moos).toHaveAttribute('aria-pressed', 'false')

    fireEvent.click(glut)
    expect(glut).toHaveAttribute('aria-pressed', 'false')
  })
})

function cleanupUndNeu() {
  cleanup()
  render(<Buehne gegenueber="du-1" gesetzt={vi.fn()} />)
}

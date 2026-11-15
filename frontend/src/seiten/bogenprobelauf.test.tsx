// „Wie es wird" vor dem Druckbogen — UXR-F7
// (#2015).
//
// # Was das Issue prüft
//
// *„Der Assistent zeigt »Wie es wird« vor dem Schreiben; ein Test fährt
// Probelauf und echten Lauf über denselben Weg."* Beides steht hier: die Zahlen
// stehen da, **bevor** jemand den Knopf drückt, und der Probelauf fragt mit
// denselben Wahlen, mit denen der Knopf danach legt.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import type { BogenprobelaufAntwort } from '../antwortenprobelauf'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { stillerBogenprobelauf } from '../test/apiprobelauf'
import { Druckbogen } from './Druckbogen'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de
const eigene = texte.seiten.ausgeben.probelauf

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Druckbogen setID="4" />
    </Sprachraum>,
  )
}

function probe(teile: Partial<BogenprobelaufAntwort> = {}) {
  return Promise.resolve({ ...stillerBogenprobelauf, ...teile })
}

function gelegt() {
  return Promise.resolve({
    pfad: '/tmp/druckbogen.pdf',
    seiten: 2,
    karten: 10,
    befunde: [],
    abgebrochen: false,
  })
}

// kasten ist der Rahmen „Wie es wird".
async function kasten() {
  return screen.findByRole('region', { name: eigene.ueberschrift })
}

describe('„Wie es wird" vor dem Druckbogen', () => {
  it('zeigt die Zahlen, bevor jemand legt', async () => {
    const BogenLegen = vi.fn(() => gelegt())

    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf: () => probe(),
      BogenLegen,
    })

    zeige()

    const rahmen = await kasten()

    expect(
      await within(rahmen).findByText('9 je Seite · 2 Seiten'),
    ).toBeVisible()
    expect(
      within(rahmen).getByText('10 Karten · Papier 210 × 297 mm · 300 dpi'),
    ).toBeVisible()
    expect(
      within(rahmen).getByText('Auf der letzten Seite: 1 Karte.'),
    ).toBeVisible()
    expect(within(rahmen).getByText(eigene.hinweis)).toBeVisible()

    // **Geschrieben ist noch nichts.**
    expect(BogenLegen).not.toHaveBeenCalled()
  })

  // **Bis die erste Antwort da ist, steht da, dass gerechnet wird.**
  it('sagt, dass gerechnet wird, solange nichts da ist', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf: () => new Promise(() => undefined),
    })

    zeige()

    expect(
      await within(await kasten()).findByText(eigene.rechnet),
    ).toBeVisible()
  })

  // **Ein Weg, zwei Ausgänge** — der Probelauf fragt mit den Wahlen des Knopfes.
  it('fragt mit denselben Wahlen, mit denen der Knopf legt', async () => {
    const BogenProbelauf = vi.fn(() => probe())
    const BogenLegen = vi.fn(() => gelegt())

    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf,
      BogenLegen,
    })

    zeige()

    // **Im Kasten** — die Fußleiste nennt dieselbe Zahl (AP-10, #2367).
    await within(await kasten()).findByText('9 je Seite · 2 Seiten')
    expect(BogenProbelauf).toHaveBeenLastCalledWith('4', 'a4', 0, 300, true, [])

    await userEvent.selectOptions(
      screen.getByLabelText(texte.seiten.druckbogen.papier),
      'letter',
    )

    const anschnitt = screen.getByLabelText(texte.seiten.druckbogen.anschnitt)
    await userEvent.clear(anschnitt)
    await userEvent.type(anschnitt, '3')

    await userEvent.selectOptions(
      screen.getByLabelText(texte.seiten.druckbogen.aufloesung),
      '600',
    )
    await userEvent.click(
      screen.getByLabelText(texte.seiten.druckbogen.schnittmarken),
    )

    await userEvent.click(
      await screen.findByRole('button', {
        name: texte.seiten.druckbogen.legen,
      }),
    )

    // Die letzte Frage und der Knopf haben dieselben sechs Argumente.
    expect(BogenProbelauf).toHaveBeenLastCalledWith(
      '4',
      'letter',
      3,
      600,
      false,
      [],
    )
    expect(BogenLegen).toHaveBeenCalledWith('4', 'letter', 3, 600, false, [])
    expect(BogenProbelauf.mock.lastCall).toEqual(BogenLegen.mock.lastCall)
  })

  it('fragt nach den gewählten Karten, nicht nach dem ganzen Set', async () => {
    const BogenProbelauf = vi.fn(() => probe({ karten: 1, seiten: 1 }))

    belegeApi({
      Kartenzeilen: () =>
        Promise.resolve({
          zeilen: [
            { id: '11', nummer: 0, name: 'Feuerball', felder: 2 },
            { id: '12', nummer: 1, name: 'Blitz', felder: 2 },
          ],
          fehler: null,
        }),
      BogenProbelauf,
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('radio', { name: /Nur ausgewählte/ }),
    )

    await vi.waitFor(() => {
      expect(BogenProbelauf).toHaveBeenLastCalledWith('4', 'a4', 0, 300, true, [
        '11',
      ])
    })
  })

  // **Eine überholte Antwort trägt sich nicht ein** (§7, Regel 2).
  it('verwirft die Antwort einer überholten Frage', async () => {
    let erste: (antwort: BogenprobelaufAntwort) => void = () => undefined

    const BogenProbelauf = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise<BogenprobelaufAntwort>((fertig) => {
            erste = fertig
          }),
      )
      .mockImplementation(() => probe({ jeSeite: 4, seiten: 3, karten: 10 }))

    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf,
    })

    zeige()

    await userEvent.selectOptions(
      await screen.findByLabelText(texte.seiten.druckbogen.papier),
      'letter',
    )

    expect(
      await within(await kasten()).findByText('4 je Seite · 3 Seiten'),
    ).toBeVisible()

    // Die erste Antwort kommt zu spät: sie gehört zu A4.
    erste({ ...stillerBogenprobelauf, jeSeite: 9, seiten: 7 })
    await new Promise((fertig) => setTimeout(fertig, 0))

    expect(screen.queryByText('9 je Seite · 7 Seiten')).not.toBeInTheDocument()
    expect(
      within(await kasten()).getByText('4 je Seite · 3 Seiten'),
    ).toBeVisible()
  })

  // **Das Vorige bleibt stehen — und der Kasten sagt, dass gerechnet wird.**
  it('sagt neben den alten Zahlen, dass neue unterwegs sind', async () => {
    let zweite: (antwort: BogenprobelaufAntwort) => void = () => undefined

    const BogenProbelauf = vi
      .fn()
      .mockImplementationOnce(() => probe())
      .mockImplementation(
        () =>
          new Promise<BogenprobelaufAntwort>((fertig) => {
            zweite = fertig
          }),
      )

    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf,
    })

    zeige()

    const rahmen = await kasten()
    await within(rahmen).findByText('9 je Seite · 2 Seiten')

    const lebendig = rahmen.querySelector('[aria-live]')
    expect(lebendig).toHaveAttribute('aria-busy', 'false')

    await userEvent.selectOptions(
      screen.getByLabelText(texte.seiten.druckbogen.papier),
      'letter',
    )

    await vi.waitFor(() => {
      expect(lebendig).toHaveAttribute('aria-busy', 'true')
    })
    expect(within(rahmen).getByText('9 je Seite · 2 Seiten')).toBeVisible()
    expect(within(rahmen).getByText(eigene.rechnet)).toBeVisible()

    zweite({ ...stillerBogenprobelauf, jeSeite: 4 })

    await vi.waitFor(() => {
      expect(lebendig).toHaveAttribute('aria-busy', 'false')
    })
    expect(within(rahmen).getByText('4 je Seite · 2 Seiten')).toBeVisible()
    expect(within(rahmen).queryByText(eigene.rechnet)).toBeNull()
  })

  it('malt die Plätze einer Seite, belegt, soweit Karten da sind', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf: () => probe({ karten: 4, seiten: 1, aufLetzterSeite: 4 }),
    })

    zeige()

    const rahmen = await kasten()
    await within(rahmen).findByText('9 je Seite · 1 Seite')

    const raster = rahmen.querySelector('[aria-hidden="true"]')
    expect(raster?.children).toHaveLength(9)
    expect(
      [...(raster?.children ?? [])].filter((platz) =>
        platz.className.includes('bg-marke'),
      ),
    ).toHaveLength(4)

    // Eine einzige Seite hat keine „letzte".
    expect(within(rahmen).queryByText(/Auf der letzten Seite/)).toBeNull()
  })

  it('verschweigt die letzte Seite nicht, wenn sie halb ist', async () => {
    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf: () =>
        probe({ karten: 18, seiten: 2, aufLetzterSeite: 9 }),
    })

    zeige()

    const rahmen = await kasten()
    await within(rahmen).findByText('9 je Seite · 2 Seiten')

    // Zwei volle Seiten haben keine halbe.
    expect(within(rahmen).queryByText(/Auf der letzten Seite/)).toBeNull()
  })
})

describe('Wenn der Probelauf Nein sagt', () => {
  // **Derselbe Satz, den der Knopf gäbe** — am verursachenden Ort (§6).
  it('zeigt den Fehler im Rahmen und lässt den Knopf stehen', async () => {
    const BogenLegen = vi.fn(() => gelegt())

    belegeApi({
      Kartenzeilen: () => Promise.resolve({ zeilen: [], fehler: null }),
      BogenProbelauf: () =>
        Promise.resolve({
          ...stillerBogenprobelauf,
          fehler: {
            code: 'validierung',
            schluessel: 'validierung',
            meldung: 'Dieses Set hat keine Karte.',
            feld: '',
            werte: {},
          },
        }),
      BogenLegen,
    })

    zeige()

    const rahmen = await kasten()
    expect(await within(rahmen).findByRole('alert')).toHaveTextContent(
      /nicht gültig/,
    )
    expect(within(rahmen).queryByText(/je Seite/)).toBeNull()

    // Der Knopf bleibt: wer es trotzdem versucht, bekommt den Satz des Laufs.
    const knopf = screen.getByRole('button', {
      name: texte.seiten.druckbogen.legen,
    })
    expect(knopf).toBeEnabled()
    await userEvent.click(knopf)
    expect(BogenLegen).toHaveBeenCalled()
  })
})

// **Derselbe Druck zweimal ist eine Entscheidung über die Menge** — DOM-03,
// S04.NAT-01 bis 03 (#1698).
//
// Bis hierher kam eine Absage: „Diesen Druck führt die Sammlung schon … Erhöhe
// dort die Menge“ — und wo „dort“ war, musste man suchen. Jetzt fragt die
// Seite: erhöhen, zur Zeile, oder nichts tun.
import { keineHerkunft } from '../apiherkunft'
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { Sammlung } from './Sammlung'
import { Sprachraum } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import { belegeApi, raeumeApi } from '../test/api'
import { druckfeld } from '../test/druckfeld'
import { kataloge } from '../sprache/kataloge'

import type { ExemplarDTO } from '../api'

const texte = kataloge.de.seiten.sammlung

function exemplar(teil: Partial<ExemplarDTO> = {}): ExemplarDTO {
  return {
    id: 1,
    fassung: 1,
    printing: 'probe-druck',
    druckID: '',
    vorlaeufig: false,
    menge: 3,
    zustand: 'nearMint',
    sprache: 'en',
    finish: 'nonfoil',
    erwerbsdatum: '',
    erwerbspreis: '',
    erwerbswaehrung: '',
    notiz: '',
    name: '',
    setcode: '',
    nummer: '',
    lagerort: '',
    wunsch: false,
    bezeichnung: 'probe-druck',
    herkunft: keineHerkunft,
    ...teil,
  }
}

// vorhanden ist die Zeile, die es schon gibt — **nicht in der gezeigten
// Liste**, damit „zur Zeile“ auch dann trägt, wenn ein Filter sie verbirgt.
const vorhanden = exemplar({
  id: 17,
  fassung: 4,
  printing: 'doppelter-druck',
  menge: 2,
  bezeichnung: 'Doppelter Druck · ABC #7',
})

// belege setzt die Bindungen: die Liste mit einer anderen Zeile, und ein
// Anlegen, das die Dublette meldet.
function belege(teil: Parameters<typeof belegeApi>[0] = {}) {
  belegeApi({
    Sammlung: () => Promise.resolve({ exemplare: [exemplar()], fehler: null }),
    Preise: () => Promise.resolve({ staende: [], befunde: [], fehler: null }),
    Portfolio: () =>
      Promise.resolve({ zeilen: [], ohnePreis: 0, fehler: null }),
    MarkenJeObjekt: () => Promise.resolve({ marken: {}, fehler: null }),
    ExemplarAnlegen: () =>
      Promise.resolve({
        id: 0,
        fehler: {
          code: 'konflikt',
          schluessel: 'doppeltesExemplar',
          meldung: 'Diesen Druck führt die Sammlung schon als Exemplar 17.',
        },
        vorhanden,
      }),
    ...teil,
  })
}

function zeige() {
  render(
    <Sprachraum sprache="de">
      <Sammlung bestandGeaendert={vi.fn()} />
    </Sprachraum>,
  )
}

// aufnehmen tippt den doppelten Druck ins Formular und schickt ihn ab.
//
// **Seit UXR-10 (#1918) steht das Formular in einer Schublade** hinter dem
// führenden Knopf im Kopf — vorher in einem Kasten im Navigator.
async function aufnehmen() {
  await userEvent.click(
    await screen.findByRole('button', { name: texte.aufnehmenOeffnen }),
  )

  const druck = await druckfeld()
  await userEvent.type(druck, 'doppelter-druck')
  await userEvent.click(screen.getByRole('button', { name: texte.aufnehmen }))

  return druck
}

// entscheidung gibt die Frage, sobald sie dasteht.
async function entscheidung() {
  return within(
    await screen.findByRole('group', {
      name: einsetzen(texte.dubletteFrage, {
        bezeichnung: vorhanden.bezeichnung,
        menge: '2',
      }),
    }),
  )
}

beforeEach(() => {
  Object.defineProperty(HTMLElement.prototype, 'clientHeight', {
    configurable: true,
    get: () => 600,
  })
})

describe('Ein doppelter Druck', () => {
  afterEach(raeumeApi)

  it('fragt nach der Menge, statt nur abzusagen', async () => {
    belege()
    zeige()
    await aufnehmen()

    const frage = await entscheidung()
    expect(
      frage.getByRole('button', {
        name: einsetzen(texte.dubletteErhoehen, { summe: '3' }),
      }),
    ).toBeInTheDocument()

    // **Und keine zweite Meldung daneben** — die Frage ist die Auskunft.
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()

    // **Der Fokus steht auf dem Abbrechen** — wer zweimal ⏎ drückt, soll
    // nichts ändern (`Entscheidung`).
    expect(
      frage.getByRole('button', { name: texte.dubletteAbbrechen }),
    ).toHaveFocus()
  })

  it('erhöht die vorhandene Zeile gegen ihre Fassung und räumt das Formular', async () => {
    const geaendert = vi.fn<(exemplar: ExemplarDTO) => Promise<null>>(() =>
      Promise.resolve(null),
    )
    belege({ ExemplarAendern: geaendert })
    zeige()
    const druck = await aufnehmen()

    await userEvent.click(
      (await entscheidung()).getByRole('button', {
        name: einsetzen(texte.dubletteErhoehen, { summe: '3' }),
      }),
    )

    await waitFor(() => {
      expect(geaendert).toHaveBeenCalledWith(
        expect.objectContaining({ id: 17, fassung: 4, menge: 3 }),
      )
    })
    await waitFor(() => {
      expect(druck).not.toBeInTheDocument()
    })
    expect(await druckfeld()).toHaveValue('')
    expect(screen.queryByRole('group', { name: /liegt schon/ })).toBeNull()

    // **Und die Zeile ist gewählt** — mit der Menge, die jetzt gilt.
    const rechts = within(
      screen.getByRole('complementary', { name: texte.inspector }),
    )
    expect(
      await rechts.findByRole('heading', { name: vorhanden.bezeichnung }),
    ).toBeInTheDocument()
    expect(
      rechts.getByText((_, knoten) => knoten?.textContent === '3 × Near Mint'),
    ).toBeInTheDocument()
  })

  it('ändert beim Abbrechen nichts und behält die Eingabe', async () => {
    const geaendert = vi.fn(() => Promise.resolve(null))
    belege({ ExemplarAendern: geaendert })
    zeige()
    const druck = await aufnehmen()

    await userEvent.click(
      (await entscheidung()).getByRole('button', {
        name: texte.dubletteAbbrechen,
      }),
    )

    await waitFor(() => {
      expect(screen.queryByRole('group', { name: /liegt schon/ })).toBeNull()
    })
    expect(druck).toHaveValue('doppelter-druck')
    expect(geaendert).not.toHaveBeenCalled()
  })

  it('bricht auch mit Escape ab', async () => {
    const geaendert = vi.fn(() => Promise.resolve(null))
    belege({ ExemplarAendern: geaendert })
    zeige()
    await aufnehmen()
    await entscheidung()

    await userEvent.keyboard('{Escape}')

    await waitFor(() => {
      expect(screen.queryByRole('group', { name: /liegt schon/ })).toBeNull()
    })
    expect(geaendert).not.toHaveBeenCalled()
  })

  it('führt zur vorhandenen Zeile, auch wenn die Liste sie nicht zeigt', async () => {
    const geaendert = vi.fn(() => Promise.resolve(null))
    belege({ ExemplarAendern: geaendert })
    zeige()
    await aufnehmen()

    await userEvent.click(
      (await entscheidung()).getByRole('button', {
        name: texte.dubletteZurZeile,
      }),
    )

    // **Die Schublade geht dabei zu** (UXR-10) — sonst wählte der Knopf eine
    // Zeile in einem Inspector, den sie verdeckt.
    await waitFor(() => {
      expect(screen.queryByRole('dialog')).toBeNull()
    })

    const rechts = within(
      screen.getByRole('complementary', { name: texte.inspector }),
    )
    expect(
      await rechts.findByRole('heading', { name: vorhanden.bezeichnung }),
    ).toBeInTheDocument()
    expect(
      rechts.getByText((_, knoten) => knoten?.textContent === '2 × Near Mint'),
    ).toBeInTheDocument()
    expect(geaendert).not.toHaveBeenCalled()
  })

  it('meldet ein gescheitertes Erhöhen und lässt die Eingabe stehen', async () => {
    const geaendert = vi.fn(() =>
      Promise.resolve({
        code: 'konflikt' as const,
        schluessel: 'konflikt',
        meldung: 'überholt',
      }),
    )
    belege({ ExemplarAendern: geaendert })
    zeige()
    const druck = await aufnehmen()

    await userEvent.click(
      (await entscheidung()).getByRole('button', {
        name: einsetzen(texte.dubletteErhoehen, { summe: '3' }),
      }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.konflikt,
    )
    expect(screen.queryByRole('group', { name: /liegt schon/ })).toBeNull()
    expect(druck).toHaveValue('doppelter-druck')
  })

  // **S04.NAT-02**: dieselbe Karte in einer anderen Sprache ist eine eigene
  // Zeile — und die Liste sagt, welche welche ist.
  it('zeigt dieselbe Karte in zwei Sprachen als zwei unterscheidbare Zeilen', async () => {
    belege({
      Sammlung: () =>
        Promise.resolve({
          exemplare: [
            exemplar({ id: 1, sprache: 'en' }),
            exemplar({ id: 2, sprache: 'de', finish: 'foil' }),
          ],
          fehler: null,
        }),
    })
    zeige()

    const liste = within(
      await screen.findByRole('grid', { name: texte.listeBezeichnung }),
    )
    // **In der Spalte und, schmal, unter der Bezeichnung** (AP-12, #2369) —
    // ohne Set und Lagerort sagen beide dasselbe.
    expect(await liste.findAllByText(`en · ${texte.nonfoil}`)).not.toHaveLength(
      0,
    )
    expect(liste.getAllByText(`de · ${texte.foil}`)).not.toHaveLength(0)
  })

  it('meldet eine Absage ohne Zeile wie bisher', async () => {
    belege({
      ExemplarAnlegen: () =>
        Promise.resolve({
          id: 0,
          fehler: {
            code: 'konflikt',
            schluessel: 'doppeltesExemplar',
            meldung: 'doppelt',
          },
          vorhanden: null,
        }),
    })
    zeige()
    await aufnehmen()

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.doppeltesExemplar,
    )
    expect(screen.queryByRole('group', { name: /liegt schon/ })).toBeNull()
  })
})

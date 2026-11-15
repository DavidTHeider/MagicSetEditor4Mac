// **Eine Zeile über dem Editor** — UXR-18
// (#1926).
//
// Das Issue fragt, ob die Hinweise über dem Editor — Geltungsbereich, eigene
// Kopie — zu einer Zeile zusammenfinden. Vorher standen dort ein umrandeter
// Kasten mit drei Zeilen Begründung, eine Reihe Handgriffe und die Warnung
// untereinander; der Editor begann bei 900 × 820 Punkten erst bei rund 400
// (Hüllenprobe `fmt02b`, Kommentar am Issue). Jetzt stehen Menü, Warnung und
// Kopie in **einer** Zeile vor dem Feld, die Begründung der Kopie hinter dem
// Fragezeichen.
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
import { Sprachraum } from '../sprache/kontext'
import { raeumeApi } from '../test/api'
import { paketDTO } from '../test/vorlagenprobe'
import { belegeWerkstatt } from '../test/werkstattprobe'
import { Vorlagen } from './Vorlagen'

import type { Paketherkunftsart } from '../pakete'

const texte = kataloge.de.seiten.vorlagen

function zeige(herkunftsart: Paketherkunftsart) {
  belegeWerkstatt({
    Pakete: () =>
      Promise.resolve({
        pakete: [{ ...paketDTO('probe.mse-style'), herkunftsart }],
        befunde: [],
        fehler: null,
      }),
  })

  render(
    <Sprachraum sprache="de">
      <Vorlagen />
    </Sprachraum>,
  )
}

// zeile gibt die Zeile über dem Editor — die, in der die Warnung steht.
async function oeffneUndFindeZeile(): Promise<HTMLElement> {
  await userEvent.click(await screen.findByRole('button', { name: 'style' }))
  await screen.findByLabelText(texte.inhalt)

  const zeile = screen.getByText(texte.warnung).parentElement

  if (zeile === null) {
    throw new Error('die Warnung steht in keiner Zeile')
  }

  return zeile
}

describe('Die Zeile über dem Editor', () => {
  afterEach(raeumeApi)

  it('stellt Handgriffe, Geltungsbereich und Kopie zusammen vor das Feld', async () => {
    zeige('bezogen')

    const zeile = await oeffneUndFindeZeile()

    expect(
      within(zeile).getByRole('button', { name: texte.datei }),
    ).toBeInTheDocument()
    expect(
      within(zeile).getByRole('button', { name: texte.kopieren }),
    ).toBeInTheDocument()
    expect(
      within(zeile).getByText(kurzform(texte.kopieFremd)),
    ).toBeInTheDocument()

    // **Die Zeile steht vor dem Feld**, und mit ihr die Warnung —
    // `barrierefreiheit.test.tsx` verlangt es für die Werkstatt.
    expect(
      zeile.compareDocumentPosition(screen.getByLabelText(texte.inhalt)) &
        Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  })

  it('legt die Begründung der Kopie hinter das Fragezeichen, nicht fort', async () => {
    zeige('mitgeliefert')

    const zeile = await oeffneUndFindeZeile()
    const kurz = kurzform(texte.kopieMitgeliefert)

    expect(within(zeile).getByText(kurz)).toBeInTheDocument()
    expect(screen.queryByText(texte.kopieMitgeliefert)).not.toBeInTheDocument()
    expect(
      within(zeile).getByRole('button', {
        name: einsetzen(kataloge.de.bausteine.mehrZu, { satz: kurz }),
      }),
    ).toBeInTheDocument()
  })

  it('nennt bei einem eigenen Paket nur Handgriffe und Geltungsbereich', async () => {
    zeige('eigen')

    const zeile = await oeffneUndFindeZeile()

    expect(within(zeile).getAllByRole('button')).toEqual([
      within(zeile).getByRole('button', { name: texte.datei }),
    ])
  })
})

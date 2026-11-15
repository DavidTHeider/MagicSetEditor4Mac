// **Die Seiten in der Palette tragen ihre Überschrift** — UXR-02
// (#1910).
//
// # Der Befund
//
// Ohne Eingabe zeigt die Palette die vier Handlungen unter *Tun* und darunter
// alle 25 Seiten — **ohne Überschrift**. *Tun*, *Finden* und *Befehle*
// tragen eine; die Seiten nannte nur das `aria-label` ihrer Liste. Fürs Auge
// lasen sie sich als Fortsetzung von *Tun*: „Sicherung erstellen", „Start",
// „Sets", „Karten" … Das Artboard *Befehle* trennt jeden Abschnitt mit einer
// Zeile (`palette-gruppe`).
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Palette } from './Palette'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'

const texte = kataloge.de.allgemein.palette

afterEach(() => {
  raeumeApi()
})

function zeige() {
  belegeApi({
    Befehlskatalog: () => Promise.resolve({ befehle: [], fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Palette schliessen={vi.fn()} waehlen={vi.fn()} waehleBefehl={vi.fn()} />
    </Sprachraum>,
  )
}

describe('die Seiten in der Palette', () => {
  it('stehen unter einer sichtbaren Überschrift', () => {
    zeige()

    const liste = screen.getByRole('list', { name: texte.seiten })
    const ueberschrift = liste.previousElementSibling

    expect(ueberschrift).toHaveTextContent(texte.seiten)
    expect(ueberschrift).toBeVisible()
  })

  // **Gefiltert bleibt sie stehen**, solange eine Seite passt — und geht mit
  // der letzten.
  it('geht mit der letzten passenden Seite', async () => {
    zeige()

    const feld = screen.getByRole('textbox')
    await userEvent.type(feld, 'stell')
    expect(
      screen.getByRole('list', { name: texte.seiten }).previousElementSibling,
    ).toHaveTextContent(texte.seiten)

    await userEvent.type(feld, 'xyzxyz')
    expect(screen.queryByText(texte.seiten)).toBeNull()
  })
})

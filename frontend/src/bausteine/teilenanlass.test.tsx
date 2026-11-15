// **Teilen, ausgelöst von außen** — UXR-06
// (#1914).
//
// # Wozu
//
// Auf der Kartenseite stand „Teilen …" als eigener Knopf neben dem Menü
// „Diese Karte" — eines der Bedienelemente, mit denen die Seite über R2 aus
// `docs/UIREVIEW.md` lag. Jetzt ist es ein Eintrag **in** diesem Menü. Der
// Baustein bleibt, was er war — er räumt beim Verlassen weg, was er
// vorbereitet hat, und sein Aufrufer setzt einen `key` je Bezug —, nur der
// Anstoß kommt von außen: ein **Anlass**, eine Zahl, die der Aufrufer
// hochzählt. Dieselbe Gestalt wie `inspectorAnlass` am Arbeitsplatz.
//
// # Was hier gehalten wird
//
// - Mit einem Anlass zeichnet der Baustein **keinen eigenen Knopf**.
// - Erst ein **neuer** Anlass bereitet vor — der beim Aufbau nicht, auch unter
//   StrictMode nicht. Sonst bereitete jeder Kartenwechsel (neuer `key`) eine
//   Teilung vor, um die niemand gebeten hat.
// - Ein Anlass, während eine Teilung offen steht, bereitet **keine zweite**
//   vor: die erste läge sonst ungeräumt im Teilordner.
import { StrictMode } from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Teilen } from './Teilen'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

import type { TeilenAntwort } from '../apiteilen'

const eigene = kataloge.de.bausteine.teilen
const kennung = 'c'.repeat(32)

const vorbereitet: TeilenAntwort = {
  teilgut: {
    kennung,
    art: 'kartenbild',
    titel: 'Feuerball',
    dateiname: 'Feuerball.png',
    groesse: 4096,
    text: '',
    mitgehendes: [],
    verluste: [],
    zuGrossFuerMail: false,
  },
  fehler: null,
}

afterEach(() => {
  raeumeApi()
})

function mitAnlass(anlass: number) {
  return (
    <StrictMode>
      <Sprachraum sprache="de">
        <Teilen art="kartenbild" bezug="5" anlass={anlass} />
      </Sprachraum>
    </StrictMode>
  )
}

describe('Teilen mit einem Anlass von außen', () => {
  it('zeichnet keinen Knopf und bereitet beim Aufbau nichts vor', () => {
    const vorbereiten = vi.fn(() => Promise.resolve(vorbereitet))
    belegeApi({ TeilenVorbereiten: vorbereiten })

    render(mitAnlass(3))

    expect(
      screen.queryByRole('button', { name: eigene.knopf }),
    ).not.toBeInTheDocument()
    expect(vorbereiten).not.toHaveBeenCalled()
  })

  it('bereitet bei einem neuen Anlass vor und zeigt, was mitgeht', async () => {
    const vorbereiten = vi.fn(() => Promise.resolve(vorbereitet))
    const verwerfen = vi.fn(() =>
      Promise.resolve({ teilgut: null, fehler: null }),
    )
    belegeApi({ TeilenVorbereiten: vorbereiten, TeilenVerwerfen: verwerfen })

    const { rerender, unmount } = render(mitAnlass(0))
    rerender(mitAnlass(1))

    expect(await screen.findByText(/Feuerball\.png/)).toBeInTheDocument()
    expect(vorbereiten).toHaveBeenCalledTimes(1)
    expect(vorbereiten).toHaveBeenCalledWith('kartenbild', '5')

    // **Ein zweiter Anlass bei offener Teilung bereitet keine zweite vor.**
    rerender(mitAnlass(2))
    await screen.findByText(/Feuerball\.png/)
    expect(vorbereiten).toHaveBeenCalledTimes(1)

    // **„Fertig" räumt weg, und danach steht kein Knopf da** — der Anstoß
    // bleibt außen.
    await userEvent.click(screen.getByRole('button', { name: eigene.fertig }))
    await waitFor(() => {
      expect(verwerfen).toHaveBeenCalledWith(kennung)
    })
    expect(screen.queryByText(/Feuerball\.png/)).not.toBeInTheDocument()
    expect(
      screen.queryByRole('button', { name: eigene.knopf }),
    ).not.toBeInTheDocument()

    // **Und der nächste Anlass bereitet wieder vor.**
    rerender(mitAnlass(3))
    expect(await screen.findByText(/Feuerball\.png/)).toBeInTheDocument()
    expect(vorbereiten).toHaveBeenCalledTimes(2)

    // **Wer geht, räumt die offene Teilung weg** — wie ohne Anlass.
    unmount()
    expect(verwerfen).toHaveBeenCalledTimes(2)
  })
})

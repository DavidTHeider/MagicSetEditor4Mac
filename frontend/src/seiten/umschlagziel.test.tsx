// **Karte und Vorlage aus einem Umschlag annehmen** —
// #1639.
//
// # Was hier gemessen wird
//
// **Dass das Ziel mitgeht** — das offene Set bei der Karte, der eingetippte
// Name beim Paket. Ein Ruf ohne Ziel legte die Karte in irgendein Set.
//
// **Dass ein abgebrochener Dialog keine Meldung wert ist** (§6) — er ist an
// der leeren Antwort zu erkennen, und danach steht kein Satz da.
//
// **Und dass ein leerer Paketname gar nicht erst hinausgeht.**
import { act, render, renderHook, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useKartenumschlag } from '../haken/useKartenumschlag'
import { Vorlagenumschlag } from '../bausteine/Vorlagenumschlag'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'

const karten = kataloge.de.seiten.karten
const pakete = kataloge.de.seiten.pakete

// eineAbsage ist die Antwort eines abgebrochenen Dialogs — leer und ohne
// Fehler.
const eineLeereKarte = () => Promise.resolve({ karteID: '', fehler: null })

// einPaket ist die Antwort eines angenommenen Pakets.
function einPaket() {
  return Promise.resolve({
    paket: {
      art: 'mse-game',
      name: 'angenommen',
      anzeigename: 'angenommen',
      version: '2026-09-19',
      pfad: '/pakete/angenommen.mse-game',
      herkunft: '',
      pruefsumme: '',
      braucht: [],
      gruppe: '',
      rang: 0,
      artrang: 0,
      herkunftsart: 'mitgeliefert' as const,
      hauptdatei: 'style',
      geaendert: false,
      gebrauchtVon: [],
      benutzt: 0,
    },
    abgebrochen: false,
    fehler: null,
  })
}

// kartenumschlag hängt den Haken in einen Sprachraum — er liest seine Sätze
// aus dem Katalog.
function kartenumschlag(angenommen: () => void) {
  return renderHook(() => useKartenumschlag('7', angenommen), {
    wrapper: ({ children }) => <Sprachraum sprache="de">{children}</Sprachraum>,
  })
}

describe('Karte und Vorlage aus einem Umschlag', () => {
  afterEach(() => {
    raeumeApi()
    vi.restoreAllMocks()
  })

  // **Die Karte geht seit UXR-05 (#1913) über einen Haken** — der Weg steht
  // im Menü „Mehr" neben „Karte anlegen", und was er meldet, zeichnet die
  // Zeile darüber (`kartenwerkzeugzeile.test.tsx`). Hier steht, was der Haken
  // zusagt: Ziel, Satz, Nachladen — und Schweigen nach einem Abbruch.
  it('nimmt eine Karte in das offene Set und lädt nach', async () => {
    const angenommen = vi.fn()
    const genommen = vi.fn(() =>
      Promise.resolve({ karteID: '42', fehler: null }),
    )

    belegeApi({ UmschlagKarte: genommen })

    const { result } = kartenumschlag(angenommen)
    await act(() => result.current.annehmen())

    expect(genommen).toHaveBeenCalledWith('7')
    expect(result.current.satz).toBe(karten.umschlagAngenommen)
    expect(angenommen).toHaveBeenCalled()
  })

  it('sagt nichts, wenn der Dialog abgebrochen wurde', async () => {
    const angenommen = vi.fn()

    belegeApi({ UmschlagKarte: eineLeereKarte })

    const { result } = kartenumschlag(angenommen)
    await act(() => result.current.annehmen())

    expect(result.current.satz).toBeUndefined()
    expect(result.current.fehler).toBeUndefined()
    expect(angenommen).not.toHaveBeenCalled()
  })

  it('meldet eine Absage der Naht an der Karte', async () => {
    belegeApi({
      UmschlagKarte: () => Promise.reject(new Error('geht nicht')),
    })

    const { result } = kartenumschlag(vi.fn())
    await act(() => result.current.annehmen())

    expect(result.current.fehler).toBeDefined()
  })

  it('nimmt eine Vorlage unter dem eingetippten Namen an', async () => {
    const angenommen = vi.fn()
    const genommen = vi.fn(einPaket)

    belegeApi({ UmschlagVorlage: genommen })

    render(
      <Sprachraum sprache="de">
        <Vorlagenumschlag angenommen={angenommen} />
      </Sprachraum>,
    )

    const hand = userEvent.setup()

    expect(screen.getByLabelText(pakete.umschlagName)).toHaveAttribute(
      'spellcheck',
      'false',
    )
    await hand.click(screen.getByLabelText(pakete.umschlagName))
    await hand.paste('angenommen.mse-game')
    await hand.click(
      screen.getByRole('button', { name: pakete.umschlagAnnehmen }),
    )

    expect(genommen).toHaveBeenCalledWith('angenommen.mse-game')
    expect(
      await screen.findByText(pakete.umschlagAngenommen),
    ).toBeInTheDocument()
    expect(angenommen).toHaveBeenCalled()
  })

  it('geht ohne Namen nicht hinaus', async () => {
    const genommen = vi.fn(einPaket)

    belegeApi({ UmschlagVorlage: genommen })

    render(
      <Sprachraum sprache="de">
        <Vorlagenumschlag angenommen={vi.fn()} />
      </Sprachraum>,
    )

    await userEvent
      .setup()
      .click(screen.getByRole('button', { name: pakete.umschlagAnnehmen }))

    expect(await screen.findByRole('alert')).toHaveTextContent(
      pakete.umschlagOhneNamen,
    )
    expect(genommen).not.toHaveBeenCalled()
  })

  it('sagt nichts, wenn der Vorlagendialog abgebrochen wurde', async () => {
    const angenommen = vi.fn()

    belegeApi({
      UmschlagVorlage: () =>
        Promise.resolve({ paket: null, abgebrochen: true, fehler: null }),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenumschlag angenommen={angenommen} />
      </Sprachraum>,
    )

    const hand = userEvent.setup()

    await hand.click(screen.getByLabelText(pakete.umschlagName))
    await hand.paste('angenommen.mse-game')
    await hand.click(
      screen.getByRole('button', { name: pakete.umschlagAnnehmen }),
    )

    expect(
      screen.queryByText(pakete.umschlagAngenommen),
    ).not.toBeInTheDocument()
    expect(angenommen).not.toHaveBeenCalled()
  })

  it('meldet eine Absage der Naht an der Vorlage', async () => {
    belegeApi({
      UmschlagVorlage: () => Promise.reject(new Error('geht nicht')),
    })

    render(
      <Sprachraum sprache="de">
        <Vorlagenumschlag angenommen={vi.fn()} />
      </Sprachraum>,
    )

    const hand = userEvent.setup()

    await hand.click(screen.getByLabelText(pakete.umschlagName))
    await hand.paste('angenommen.mse-game')
    await hand.click(
      screen.getByRole('button', { name: pakete.umschlagAnnehmen }),
    )

    expect(await screen.findByRole('alert')).toBeInTheDocument()
  })
})

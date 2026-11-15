// **Die Sammeldaten im Fenster** —
// #888.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { Sammeldaten } from '../bausteine/Sammeldaten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de.seiten.pakete.sammeldaten

const fehlt = {
  art: 'oracle_cards',
  pfad: '/wo/auch/immer/oracle_cards.jsonl.gz',
  vorhanden: false,
  mb: 0,
  angebotenMB: 24.5,
  stand: '2026-08-23T21:01:56.596+00:00',
  geholt: '',
  fehler: null,
}

const daliegt = {
  ...fehlt,
  vorhanden: true,
  mb: 24.5,
  geholt: '2026-10-01T08:00:00Z',
}

function zeige(teile = {}) {
  belegeApi({
    ScryfallSammeldaten: () => Promise.resolve(fehlt),
    ...teile,
  })

  return render(
    <Sprachraum sprache="de">
      <Sammeldaten />
    </Sprachraum>,
  )
}

describe('Die Sammeldaten', () => {
  it('sagt, dass noch nichts da ist — und was Scryfall anbietet', async () => {
    zeige()

    expect(await screen.findByText(texte.fehlt)).toBeInTheDocument()
    expect(await screen.findByText(/24\.5 MB/)).toBeInTheDocument()
  })

  // **Der Zeitpunkt wird gelesen, nicht durchgereicht.**
  //
  // Gefunden bei der Sichtprüfung am 2026-08-24: im Fenster stand
  // `2026-08-23T21:01:56.596+00:00`. Die Go-Seite reicht die Angabe
  // **absichtlich** roh durch — sie kennt die Zone des Menschen nicht (§3) —,
  // und der Test hier prüfte nur, dass sie ankommt.
  it('zeigt den Zeitpunkt lesbar und nicht als ISO-Angabe', async () => {
    zeige()

    const zeile = await screen.findByText(/Scryfall bietet/)

    expect(zeile.textContent).not.toContain('T21:01:56')
    expect(zeile.textContent).not.toContain('596')
    expect(zeile.textContent).toContain('2026')
  })

  it('holt die Datei und zeigt danach ihre Größe', async () => {
    const ScryfallSammeldatenHolen = vi.fn(() => Promise.resolve(daliegt))

    zeige({ ScryfallSammeldatenHolen })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.holen }),
    )

    expect(ScryfallSammeldatenHolen).toHaveBeenCalled()

    expect(await screen.findByText(/Vorhanden, 24\.5 MB/)).toBeInTheDocument()
  })

  it('meldet einen Fehler beim Holen', async () => {
    zeige({
      ScryfallSammeldatenHolen: () =>
        Promise.resolve({
          ...fehlt,
          fehler: { code: 'unbekannt', schluessel: '', meldung: 'Kein Netz.' },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', { name: texte.holen }),
    )

    expect(await screen.findByText('Kein Netz.')).toBeInTheDocument()
  })

  // **Ohne Netz sagt der Kasten trotzdem, was liegt.**
  //
  // Der Stand liest zuerst die Platte; das Verzeichnis von Scryfall ist eine
  // Zugabe. Fiele beides zusammen aus, stünde hier gar nichts — und das wäre
  // die Umkehrung von §16, Regel 5.
  it('zeigt die vorhandene Datei auch ohne angebotene Größe', async () => {
    zeige({
      ScryfallSammeldaten: () =>
        Promise.resolve({ ...daliegt, angebotenMB: 0, stand: '' }),
    })

    expect(await screen.findByText(/Vorhanden, 24\.5 MB/)).toBeInTheDocument()

    expect(screen.queryByText(/Scryfall bietet/)).not.toBeInTheDocument()
  })

  // **Ohne Netz sagt die liegende Datei, wie alt sie ist** (F04, #2246) —
  // dort fehlt der Stand, den Scryfall nennt, und bis dahin stand nur ihre
  // Größe da.
  it('nennt, wann die liegende Datei geholt wurde', async () => {
    zeige({
      ScryfallSammeldaten: () =>
        Promise.resolve({ ...daliegt, angebotenMB: 0, stand: '' }),
    })

    expect(await screen.findByText(/Geholt .*2026/)).toBeInTheDocument()
  })

  it('nennt keine Abrufzeit, solange nichts liegt', async () => {
    zeige()

    expect(await screen.findByText(texte.fehlt)).toBeInTheDocument()
    expect(screen.queryByText(/Geholt/)).toBeNull()
  })
})

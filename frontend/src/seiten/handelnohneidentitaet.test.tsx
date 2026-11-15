// **Ohne Identität ist Handeln ein Anfang und kein Fehler** — UXR-23
// (#1931),
// Befund #2027.
//
// Gemessen am nativen Fenster (`uxr23`, frischer Bestand): von 25 Seiten
// zeigte allein *Handeln* einen roten Fehler — „Dieses Profil hat noch keine
// Identität …" — und den Knopf „Erneut laden", der nichts ändern kann. §6:
// leere Zustände erklären den nächsten Schritt. Der nächste Schritt ist, eine
// Identität anzulegen, und das geht jetzt hier.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { Handeln } from './Handeln'
import { Sprachraum } from '../sprache/kontext'
import { eineListe } from './peerkontaktprobe'
import { eineUebersicht, eineZeile } from './handelslisteprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handeln.liste

const ohneIdentitaet = {
  zeilen: [],
  fehler: {
    code: 'nichtGefunden' as const,
    schluessel: 'ohnePeeridentitaet',
    meldung: 'dieses Profil hat noch keine Identität',
  },
}

afterEach(() => {
  raeumeApi()
  vi.restoreAllMocks()
})

function zeige(weitere: Record<string, unknown> = {}) {
  belegeApi({
    Peerkontakte: () => Promise.resolve(eineListe([])),
    ...weitere,
  })

  render(
    <Sprachraum sprache="de">
      <Handeln />
    </Sprachraum>,
  )
}

describe('Handeln ohne Identität', () => {
  it('zeigt den nächsten Schritt statt eines Fehlers', async () => {
    zeige({ HandelUebersicht: () => Promise.resolve(ohneIdentitaet) })

    const leer = await screen.findByText(eigene.ohneIdentitaetTitel)
    const knopf = screen.getByRole('button', {
      name: eigene.ohneIdentitaetAnlegen,
    })

    // **Titel und Knopf stehen im selben Leerzustand.**
    expect(knopf.closest('[data-leerzustand]')).toBe(
      leer.closest('[data-leerzustand]'),
    )
    expect(knopf.closest('[data-leerzustand]')).not.toBeNull()
    expect(screen.queryByRole('alert')).toBeNull()
    expect(
      screen.queryByRole('button', {
        name: kataloge.de.bausteine.zustaende.wiederholen,
      }),
    ).toBeNull()
  })

  it('legt die Identität an und lädt die Liste nach', async () => {
    let angelegt = false
    const anlegen = vi.fn(() => {
      angelegt = true
      return Promise.resolve({ stand: 'bereit', fehler: null })
    })
    zeige({
      HandelUebersicht: () =>
        Promise.resolve(
          angelegt ? eineUebersicht([eineZeile()]) : ohneIdentitaet,
        ),
      PeeridentitaetSicherstellen: anlegen,
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: eigene.ohneIdentitaetAnlegen,
      }),
    )

    await waitFor(() => {
      expect(screen.queryByText(eigene.ohneIdentitaetTitel)).toBeNull()
    })
    expect(anlegen).toHaveBeenCalledTimes(1)
  })

  it('nennt einen gescheiterten Versuch am Leerzustand', async () => {
    zeige({
      HandelUebersicht: () => Promise.resolve(ohneIdentitaet),
      PeeridentitaetSicherstellen: () =>
        Promise.resolve({
          stand: 'ohne',
          fehler: {
            code: 'intern' as const,
            schluessel: 'unbekannt',
            meldung: 'der Schlüsselbund antwortet nicht',
          },
        }),
    })

    await userEvent.click(
      await screen.findByRole('button', {
        name: eigene.ohneIdentitaetAnlegen,
      }),
    )

    const meldung = await screen.findByRole('alert')
    expect(meldung.closest('[data-leerzustand]')?.textContent).toContain(
      eigene.ohneIdentitaetTitel,
    )
  })

  // **Ein führender Schritt, und der steht oben** — gesehen in der
  // Sichtprüfung am Fenster (`uxr23`, 2026-10-03): „Neuer Tausch …" und
  // „Identität anlegen" standen beide im Hauptgewicht in derselben Spalte
  // (R1), und der Leerzustand stand unter Suchfeld und Filtern, die ohne einen
  // Handel nichts filtern.
  it('lässt allein „Identität anlegen" führen und zeigt keine Filter', async () => {
    zeige({ HandelUebersicht: () => Promise.resolve(ohneIdentitaet) })

    const anlegen = await screen.findByRole('button', {
      name: eigene.ohneIdentitaetAnlegen,
    })

    expect(anlegen).toHaveClass('bg-marke')

    // **Auf die Wirkung gewartet** (§3): die Liste meldet die fehlende
    // Identität über einen Effekt, und „Neuer Tausch …" tritt einen Durchgang
    // später zurück. Ohne das Warten scheiterte der Lauf in 2 von 10.
    await waitFor(() => {
      expect(
        screen.getByRole('button', {
          name: kataloge.de.seiten.handeln.neuerTausch,
        }),
      ).not.toHaveClass('bg-marke')
    })
    expect(screen.queryByRole('searchbox')).toBeNull()
    expect(screen.queryByRole('checkbox')).toBeNull()
  })
})

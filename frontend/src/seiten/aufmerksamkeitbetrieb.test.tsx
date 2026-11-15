// **Das Betriebsband steht in „Was Aufmerksamkeit braucht"** — OPS02
// (#2348), Artboard `Start09`: je Lage ein Satz, ein Rat und ein Ziel.
import { render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { AufmerksamkeitProbe } from '../test/aufmerksamkeitprobe'
import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import type { BetriebsbandAntwort } from '../apibetriebsband'

const texte = kataloge.de.seiten.start.aufmerksamkeit

const viere: BetriebsbandAntwort = {
  hinweise: [
    {
      art: 'platzKnapp',
      ziel: 'einstellungen/daten',
      werte: { frei: String(1_500_000_000) },
    },
    { art: 'sicherungAlt', ziel: 'einstellungen/daten', werte: { tage: '20' } },
    {
      art: 'dienstBeendet',
      ziel: 'einstellungen/diagnose',
      werte: { ausgang: '1' },
    },
    {
      art: 'webprozessNeu',
      ziel: 'einstellungen/diagnose',
      werte: { anzahl: '2' },
    },
  ],
  ungemessen: [],
  aufwandMikros: 400,
  fehler: null,
}

function Offen() {
  const offen = useAnschluss().offen

  return (
    <output>
      {offen?.art === 'einstellungsgruppe'
        ? [offen.gruppe, offen.abschnitt ?? ''].join(' ').trim()
        : ''}
    </output>
  )
}

function hinstellen(betrieb: BetriebsbandAntwort, sprache: 'de' | 'en' = 'de') {
  belegeApi({ Betriebsband: vi.fn().mockResolvedValue(betrieb) })
  const wechseln = vi.fn()

  render(
    <Sprachraum sprache={sprache}>
      <Anschlussraum aktuell="start" wechseln={wechseln}>
        <AufmerksamkeitProbe waehleSet={vi.fn()} stand={0} />
        <Offen />
      </Anschlussraum>
    </Sprachraum>,
  )

  return wechseln
}

afterEach(() => {
  raeumeApi()
})

describe('Betriebshinweise', () => {
  it('nennt jede Lage mit Satz und Rat und zählt sie mit', async () => {
    hinstellen(viere)

    const bereich = await screen.findByRole('region', { name: texte.titel })

    await within(bereich).findByRole('button', {
      name: /Letzte Sicherung vor 20 Tagen/,
    })

    for (const satz of [/1,5 GB/, /Code 1/, /2-mal/]) {
      expect(
        within(bereich).getByRole('button', { name: satz }),
      ).toBeInTheDocument()
    }

    expect(within(bereich).getByText('4')).toBeInTheDocument()
    expect(within(bereich).queryByText(/^Nichts/)).not.toBeInTheDocument()
  })

  it('führt zu Daten und Sicherung und zur Diagnose', async () => {
    const wechseln = hinstellen(viere)

    await userEvent.click(
      await screen.findByRole('button', { name: /Letzte Sicherung/ }),
    )

    expect(wechseln).toHaveBeenCalledWith('einstellungen')
    expect(screen.getByRole('status')).toHaveTextContent(/^daten$/)

    await userEvent.click(screen.getByRole('button', { name: /Code 1/ }))

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent(/^diagnose$/)
    })

    // **Und der Webprozess schlägt auf, wo sein Speicher steht** — seit
    // #2465 eingeklappt unter „Gewirkt seit dem Start“.
    await userEvent.click(screen.getByRole('button', { name: /Webansicht/ }))

    await waitFor(() => {
      expect(screen.getByRole('status')).toHaveTextContent(/^diagnose wirkung$/)
    })
  })

  it('spricht die Sprache des Fensters', async () => {
    hinstellen(viere, 'en')

    expect(
      await screen.findByRole('button', { name: /20 days/ }),
    ).toBeInTheDocument()
  })

  it('lässt einen Hinweis mit unbekanntem Ziel ohne Weg stehen', async () => {
    hinstellen({
      ...viere,
      hinweise: [
        { art: 'sicherungAlt', ziel: 'irgendwo', werte: { tage: '20' } },
      ],
    })

    const satz = await screen.findByText(/Letzte Sicherung vor 20 Tagen/)

    expect(satz.closest('button')).toBeNull()
  })
})

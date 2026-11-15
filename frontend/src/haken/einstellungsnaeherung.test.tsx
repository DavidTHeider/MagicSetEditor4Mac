// **Was gilt, rechnet der Bestand und nicht das Fenster** — zur Korrektur
// von N6 im Nachlauf über Phase 6.12
// (#1905).
//
// `wirksam` ist eine Faltung in `internal/domain`: `erben` nimmt die Wahl
// beim Menschen, eine Richtlinie die strengere, „nur lokal" schaltet jeden
// Dienst ab, und ein Proxy steht ohne Kennwort da. Ein Fenster, das nach
// dem Schreiben `wirksam` gleich `wert` setzte, zeigte für einen Augenblick
// eine Sprache namens `erben` — auf dem Erfolgsweg, bei jeder Wahl.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it } from 'vitest'

import { belegeApi, raeumeApi } from '../test/api'
import { useEinstellungen } from './useEinstellungen'

afterEach(() => {
  raeumeApi()
})

const proxy = 'netz.proxy.http'

function Anzeige() {
  const stand = useEinstellungen()
  const gewaehlt = (schluessel: string) =>
    stand.eintraege.find((eintrag) => eintrag.schluessel === schluessel)?.wert

  return (
    <>
      <p>{`gewählt:${gewaehlt('sprache') ?? ''}`}</p>
      <p>{`wirksam:${stand.wert('sprache')}`}</p>
      <p>{`proxy:${gewaehlt(proxy) ?? ''}`}</p>
      <p>{stand.ursache === undefined ? 'ohne Fehler' : 'mit Fehler'}</p>
      <button
        type="button"
        onClick={() => {
          stand.setze('sprache', 'erben')
        }}
      >
        erben
      </button>
      <button
        type="button"
        onClick={() => {
          stand.setze(proxy, 'http://nutzer:geheim@proxy:8080')
        }}
      >
        proxy
      </button>
    </>
  )
}

function anfang() {
  return {
    einstellungen: [
      {
        schluessel: 'sprache',
        wert: 'en',
        wirksam: 'en',
        werte: ['erben', 'de', 'en'],
        bereich: 'anwendung',
      },
      {
        schluessel: proxy,
        wert: '',
        wirksam: '',
        werte: [],
        bereich: 'netz',
      },
    ],
  }
}

// belegeNachladen gibt beim ersten Lesen den Anfang und hält jedes weitere
// fest, bis der Test es ablehnt.
function belegeNachladen() {
  let gelesen = 0
  const offen: ((grund: Error) => void)[] = []

  belegeApi({
    Einstellungen: () => {
      gelesen += 1

      return gelesen === 1
        ? Promise.resolve(anfang())
        : new Promise((_, ablehnen) => {
            offen.push(ablehnen)
          })
    },
    EinstellungSetzen: () => Promise.resolve({}),
  })

  return offen
}

it('zeigt während des Nachladens nicht „erben" als geltende Sprache', async () => {
  const offen = belegeNachladen()

  render(<Anzeige />)

  await screen.findByText('wirksam:en')
  await userEvent.click(screen.getByRole('button', { name: 'erben' }))

  await waitFor(() => {
    expect(offen).toHaveLength(1)
  })
  expect(screen.getByText('wirksam:en')).toBeInTheDocument()
})

it('trägt nach gescheitertem Nachladen die Wahl nach, nicht das Geltende', async () => {
  const offen = belegeNachladen()

  render(<Anzeige />)

  await screen.findByText('wirksam:en')
  await userEvent.click(screen.getByRole('button', { name: 'erben' }))

  await waitFor(() => {
    expect(offen).toHaveLength(1)
  })
  offen[0]?.(new Error('der Bestand ist zu'))

  await screen.findByText('mit Fehler')
  expect(screen.getByText('gewählt:erben')).toBeInTheDocument()
  expect(screen.getByText('wirksam:en')).toBeInTheDocument()
})

it('trägt eine freie Eingabe nicht nach — ihre Anzeige kennt nur der Bestand', async () => {
  const offen = belegeNachladen()

  render(<Anzeige />)

  await screen.findByText('wirksam:en')
  await userEvent.click(screen.getByRole('button', { name: 'proxy' }))

  await waitFor(() => {
    expect(offen).toHaveLength(1)
  })
  expect(screen.queryByText(/geheim/)).not.toBeInTheDocument()

  offen[0]?.(new Error('der Bestand ist zu'))

  await screen.findByText('mit Fehler')
  expect(screen.queryByText(/geheim/)).not.toBeInTheDocument()
})

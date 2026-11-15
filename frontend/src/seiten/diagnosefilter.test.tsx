// **„Diagnose öffnen" führt zur Diagnose und filtert auf den Vorgang** —
// NET-02, S13.NAT-04 (#1706).
//
// # Der Befund
//
// Ein unerwarteter Fehler sagte „Suche dort nach {vorgang}" — wo dort war,
// musste man selbst finden, und suchen konnte man in der Diagnose nicht.
// Jetzt öffnet ein Knopf die Gruppe *Diagnose* mit der Kennung, die Diagnose
// holt ihren Text und zeigt nur die Zeilen des Vorgangs.
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import { Anschlussraum, useAnschluss } from '../anschlussraum'
import { Diagnoseverweis } from '../bausteine/Diagnoseverweis'
import { Diagnose } from './Diagnose'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { belegeApi, raeumeApi } from '../test/api'
import { Bindungsfehler } from '../bindungsfehler'

import type { Seitenschluessel } from '../navigation'

const texte = kataloge.de.seiten.technik.diagnose

afterEach(() => {
  raeumeApi()
})

const derText = [
  'MSE4Mac 0.6.0',
  '2026-09-30T10:00:00Z info start',
  '2026-09-30T10:00:01Z fehler vorgang=a1b2c3d4 nachbarn: kaputt',
  '2026-09-30T10:00:02Z info anderes',
  '2026-09-30T10:00:03Z fehler vorgang=a1b2c3d4 grund: EHOSTUNREACH',
].join('\n')

// Buehne zeigt einen Fehler mit Verweis und — nach dem Sprung — die Diagnose.
function Buehne() {
  const [aktuell, setzeAktuell] = useState<Seitenschluessel>('start')

  return (
    <Sprachraum sprache="de">
      <Anschlussraum aktuell={aktuell} wechseln={setzeAktuell}>
        {aktuell === 'einstellungen' ? (
          <Diagnose />
        ) : (
          <Diagnoseverweis
            ursache={
              new Bindungsfehler({
                code: 'unbekannt',
                schluessel: 'unbekanntMitVorgang',
                meldung: 'kaputt',
                werte: { vorgang: 'a1b2c3d4' },
              })
            }
          />
        )}
        <Anschlussstand />
      </Anschlussraum>
    </Sprachraum>
  )
}

// Anschlussstand hält den Anschluss sichtbar — nur für den Lauf.
function Anschlussstand() {
  const anschluss = useAnschluss()

  return <output>{anschluss.offen?.art ?? ''}</output>
}

it('öffnet die Diagnose und zeigt nur die Zeilen des Vorgangs', async () => {
  const geholt = vi.fn(() => Promise.resolve({ text: derText, fehler: null }))
  belegeApi({ Diagnose: geholt })

  render(<Buehne />)

  await userEvent.click(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.diagnoseverweis.oeffnen,
    }),
  )

  // **Der Filter steht da, mit der Kennung**, und der Text ist geholt.
  expect(await screen.findByLabelText(texte.filter)).toHaveValue('a1b2c3d4')
  expect(geholt).toHaveBeenCalledTimes(1)

  const gezeigt = await screen.findByText(/grund: EHOSTUNREACH/)
  expect(gezeigt.textContent).not.toContain('info anderes')
  expect(gezeigt.textContent).toContain('nachbarn: kaputt')
  expect(screen.getByText('2 von 5 Zeilen')).toBeInTheDocument()
})

it('zeigt ohne Filter alles, und mit einem leeren wieder alles', async () => {
  belegeApi({
    Diagnose: () => Promise.resolve({ text: derText, fehler: null }),
  })

  render(
    <Sprachraum sprache="de">
      <Diagnose />
    </Sprachraum>,
  )

  const hand = userEvent.setup()
  await hand.click(screen.getByRole('button', { name: texte.anzeigen }))
  expect(await screen.findByText(/info anderes/)).toBeInTheDocument()

  await hand.type(screen.getByLabelText(texte.filter), 'EHOSTUNREACH')
  expect(screen.queryByText(/info anderes/)).toBeNull()
  expect(screen.getByText('1 von 5 Zeilen')).toBeInTheDocument()

  await hand.clear(screen.getByLabelText(texte.filter))
  expect(screen.getByText(/info anderes/)).toBeInTheDocument()
})

it('bietet keinen Verweis ohne Vorgangskennung', () => {
  render(
    <Sprachraum sprache="de">
      <Diagnoseverweis ursache={new Error('ohne Kennung')} />
    </Sprachraum>,
  )

  expect(screen.queryByRole('button')).toBeNull()
})

// **Die Meldung des Fensters holt auch die Darstellung ins Band** — Befund G5
// des Zweitblicks über Phase 6.3
// (#1665):
// ein gescheitertes »Darstellung zurücksetzen« aus der Menüleiste stand nur
// unter *Einstellungen › Erscheinungsbild*. Hier steht es im Band, und das
// Kreuz räumt es mit allem anderen.
import { act, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { Gestaltungsraum, useGestaltung } from '../gestaltungsraum'
import { Nutzereinstellungsraum } from '../nutzereinstellungsraum'
import { Meldungsband } from '../rahmen/Meldungsband'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { probeprofil } from '../test/profil'
import { useMeldungsband } from './useMeldungsband'

afterEach(() => {
  raeumeApi()
  delete (globalThis as { runtime?: unknown }).runtime
})

const grund = 'Der Bestand antwortet nicht.'

function Fenster({ raeumen }: { raeumen: () => void }) {
  const band = useMeldungsband(
    {
      sprache: 'de',
      rueckspielrest: '',
      dateimeldung: undefined,
      verlaufsfehler: undefined,
      bestandsfehler: undefined,
      bestandspfad: undefined,
      setID: undefined,
      setname: undefined,
      seite: 'hilfe',
      speicherungFuer: undefined,
      speicherfehler: undefined,
      speicherpfad: undefined,
    },
    raeumen,
  )
  const gestaltung = useGestaltung()

  return (
    <>
      <Meldungsband meldung={band.meldung} schliessen={band.raeumen} />
      <output>{`seite:${gestaltung?.ruecksetzfehler ?? 'still'}`}</output>
    </>
  )
}

it('zeigt ein gescheitertes Zurücksetzen im Band und räumt es mit dem Kreuz', async () => {
  const horcher = new Map<string, () => void>()
  ;(globalThis as { runtime?: unknown }).runtime = {
    EventsOn: (name: string, rueckruf: () => void) => {
      horcher.set(name, rueckruf)

      return () => {
        horcher.delete(name)
      }
    },
  }
  belegeApi({
    GezeigtesProfil: vi.fn(() =>
      Promise.resolve({
        profil: probeprofil({
          id: '10',
          nutzerID: '7',
          name: 'P',
          fassung: '1',
        }),
      }),
    ),
    Nutzereinstellungen: vi.fn(() => Promise.resolve({ einstellungen: [] })),
    Ruecksetzungsvorschau: vi.fn(() =>
      Promise.resolve({
        revision: '',
        art: 'layout',
        betroffen: [],
        hindernisse: [],
        bereit: false,
        nameNoetig: false,
        nachweisNoetig: false,
        sicherungEmpfohlen: false,
        fehler: {
          code: 'unbekannt' as const,
          schluessel: 'unbekannt',
          meldung: grund,
        },
      }),
    ),
  })
  const raeumen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Nutzereinstellungsraum>
        <Gestaltungsraum einstellungen={{ wert: () => '' }} neuLesen={vi.fn()}>
          <Fenster raeumen={raeumen} />
        </Gestaltungsraum>
      </Nutzereinstellungsraum>
    </Sprachraum>,
  )

  await waitFor(() => {
    expect(horcher.has('menue:darstellung-zuruecksetzen')).toBe(true)
  })
  act(() => {
    horcher.get('menue:darstellung-zuruecksetzen')?.()
  })

  // **Der Grund kommt übersetzt an** (`meldungAus`), der Vorgang davor
  // steht fest.
  expect(
    await screen.findByText(
      /^Die Darstellung ließ sich nicht zurücksetzen — ./,
    ),
  ).toBeInTheDocument()

  await userEvent.click(
    screen.getByRole('button', {
      name: kataloge.de.bausteine.aktionen.schliessen,
    }),
  )

  expect(raeumen).toHaveBeenCalledTimes(1)
  expect(await screen.findByText('seite:still')).toBeInTheDocument()
  expect(
    screen.queryByText(/Darstellung ließ sich nicht/),
  ).not.toBeInTheDocument()
})

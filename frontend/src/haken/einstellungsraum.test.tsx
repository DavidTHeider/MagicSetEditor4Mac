// **Ein Einstellungsstand für das ganze Fenster** — Befund G14 des
// Zweitblicks über Phase 6.3
// (#1665):
// `useFeldtiefe` und `useAnsicht` hielten je einen eigenen. Nach jedem
// Kartenwechsel galt kurz die Vorgabe, und die Einstellungsseite zeigte
// danach einen veralteten Wert.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState, type ReactNode } from 'react'
import { afterEach, expect, it, vi } from 'vitest'

import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Einstellungsraum } from './einstellungsraum'
import { useEinstellungen } from './useEinstellungen'
import { useFeldtiefe } from './useFeldtiefe'

afterEach(() => {
  raeumeApi()
})

function tiefe(wert: string) {
  return {
    einstellungen: [
      {
        schluessel: 'feldtiefe',
        wert,
        wirksam: wert,
        werte: ['normal', 'erweitert', 'experte'],
        bereich: 'anwendung',
      },
    ],
  }
}

function Wurzel({ children }: { children: ReactNode }) {
  return (
    <Einstellungsraum stand={useEinstellungen()}>{children}</Einstellungsraum>
  )
}

function Tiefe({ name }: { name: string }) {
  const wahl = useFeldtiefe()

  return (
    <button
      type="button"
      onClick={() => {
        wahl.aufTiefe('erweitert')
      }}
    >{`${name}:${wahl.tiefe}`}</button>
  )
}

function zeige(kinder: ReactNode) {
  render(
    <Sprachraum sprache="de">
      <Wurzel>{kinder}</Wurzel>
    </Sprachraum>,
  )
}

it('liest einmal für alle Haken darunter', async () => {
  const Einstellungen = vi.fn(() => Promise.resolve(tiefe('experte')))
  belegeApi({ Einstellungen })

  zeige(
    <>
      <Tiefe name="a" />
      <Tiefe name="b" />
    </>,
  )

  expect(await screen.findByText('a:experte')).toBeInTheDocument()
  expect(screen.getByText('b:experte')).toBeInTheDocument()
  expect(Einstellungen).toHaveBeenCalledTimes(1)
})

// **Nach einem Kartenwechsel gilt nicht kurz die Vorgabe**: ein Haken, der
// nach dem Laden entsteht, sieht sofort, was gilt.
it('gibt einem später aufgebauten Haken sofort den geladenen Stand', async () => {
  // **Ein zweites Laden kommt nie an** — so zeigt ein Haken mit eigenem
  // Stand die Vorgabe, statt dass eine schnelle Vorrichtung es verdeckt.
  let geladen = false
  belegeApi({
    Einstellungen: () => {
      if (geladen) {
        return new Promise(() => undefined)
      }
      geladen = true

      return Promise.resolve(tiefe('experte'))
    },
  })

  function Wechsel() {
    const [offen, setzeOffen] = useState(false)

    return offen ? (
      <Tiefe name="neu" />
    ) : (
      <button
        type="button"
        onClick={() => {
          setzeOffen(true)
        }}
      >
        wechseln
      </button>
    )
  }

  zeige(
    <>
      <Tiefe name="alt" />
      <Wechsel />
    </>,
  )
  await screen.findByText('alt:experte')

  await userEvent.click(screen.getByRole('button', { name: 'wechseln' }))

  expect(screen.getByText('neu:experte')).toBeInTheDocument()
})

// **Und eine Änderung sieht jeder** — auch die Einstellungsseite, die nicht
// geklickt hat.
it('zeigt eine Änderung über einen Haken auch am anderen', async () => {
  let stand = 'normal'
  belegeApi({
    Einstellungen: () => Promise.resolve(tiefe(stand)),
    EinstellungSetzen: (_: string, wert: string) => {
      stand = wert

      return Promise.resolve({ fehler: null })
    },
  })

  zeige(
    <>
      <Tiefe name="a" />
      <Tiefe name="b" />
    </>,
  )
  await userEvent.click(await screen.findByText('a:normal'))

  await waitFor(() => {
    expect(screen.getByText('b:erweitert')).toBeInTheDocument()
  })
})

// **Ein gescheitertes Schreiben merkt sich, wo es geschah** — damit die
// Meldung an der Einstellung steht und nicht als Ladefehler über der Seite
// (§6; gefunden am nativen Fenster für OPS-02, #1719). Ein geglücktes
// Schreiben derselben Einstellung räumt sie.
it('merkt sich die Einstellung eines gescheiterten Schreibens', async () => {
  let scheitern = true
  belegeApi({
    Einstellungen: () => Promise.resolve(tiefe('normal')),
    EinstellungSetzen: () =>
      Promise.resolve(
        scheitern
          ? {
              fehler: {
                code: 'validierung' as const,
                schluessel: 'ordnerFehlt',
                meldung: 'nicht erreichbar',
              },
            }
          : { fehler: null },
      ),
  })

  function Anzeige() {
    const stand = useEinstellungen()

    return (
      <>
        <button
          type="button"
          onClick={() => {
            stand.setze('feldtiefe', 'erweitert')
          }}
        >
          setzen
        </button>
        <output>{`an:${stand.fehlschlag?.schluessel ?? '-'} laden:${String(stand.ursache !== undefined)}`}</output>
      </>
    )
  }

  render(
    <Sprachraum sprache="de">
      <Anzeige />
    </Sprachraum>,
  )

  await userEvent.click(screen.getByRole('button', { name: 'setzen' }))
  expect(
    await screen.findByText('an:feldtiefe laden:false'),
  ).toBeInTheDocument()

  scheitern = false
  await userEvent.click(screen.getByRole('button', { name: 'setzen' }))
  expect(await screen.findByText('an:- laden:false')).toBeInTheDocument()
})

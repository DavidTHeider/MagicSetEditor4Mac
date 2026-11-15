// Die Stilauswahl, wenn zwei Läufe einander überholen (§7, Regel 2).
//
// **Eine eigene Datei**, weil `stilwahl.test.tsx` an der Zeilengrenze steht
// (§2) — und weil es eine andere Frage ist: dort die Auswahl, hier die
// Reihenfolge.
import { act, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { type Api, type FolgenAntwort } from '../api'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { waehleZeile } from '../test/suchwahl'
import { Stilwahl } from './Stilwahl'

afterEach(() => {
  raeumeApi()
})

const texte = kataloge.de

function belegeStil(teile: Partial<Api>) {
  return belegeApi({
    Stile: () =>
      Promise.resolve({
        stile: [
          {
            paket: 'probe-schlicht',
            anzeigename: 'Probe-schlicht',
            spiel: 'probe',
            version: '2026-08-15',
            eigenesSpiel: true,
            gewaehlt: true,
          },
          {
            paket: 'fremd',
            anzeigename: 'Fremd',
            spiel: 'anderes',
            version: '2026-01-01',
            eigenesSpiel: false,
            gewaehlt: false,
          },
        ],
      }),
    Stiloptionen: () => Promise.resolve({ optionen: [] }),

    // **Die Vorschau der Stilseite lädt mit** (#643) — ohne diese Zeile
    // meldete sie ihren eigenen Befund.
    Karten: () => Promise.resolve({ karten: [] }),
    ...teile,
  })
}

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Stilwahl setID="1" geaendert={vi.fn()} />
    </Sprachraum>,
  )
}

// wechsleSet zeichnet dieselbe Auswahl für ein anderes Set.
function wechsleSet(stand: ReturnType<typeof zeige>) {
  stand.rerender(
    <Sprachraum sprache="de">
      <Stilwahl setID="2" geaendert={vi.fn()} />
    </Sprachraum>,
  )
}

describe('Stilauswahl — überholte Läufe', () => {
  // **Eine verspätete Rückfrage gehört zu dem Set, aus dem sie stammt**
  // (§7, Regel 2).
  //
  // # Woher der Fall kommt
  //
  // Aus dem Zweitblick über Phase 7: `fragen` schützte weder Erfolg noch
  // Fehler mit einem Laufstand. Antwortete eine Anfrage für Set A erst nach
  // dem Wechsel auf B, baute sie **dort** die Rückfrage auf — und ihre
  // Bestätigung rief `StilWechseln` mit der Kennung von B.
  it('trägt eine verspätete Rückfrage nicht am neuen Set ein', async () => {
    let loese: (() => void) | undefined
    const Stilfolgen = vi.fn(
      (): Promise<FolgenAntwort> =>
        new Promise((erfuellen) => {
          loese = () => {
            erfuellen({
              folgen: {
                ungedeuteteFelder: [],
                verloreneOptionen: [],
                spielwechsel: false,
              },
            })
          }
        }),
    )

    belegeStil({ Stilfolgen })
    const stand = zeige()

    // **Über die Auswahl** — seit UXR-07 (#1915) kein Optionsfeld je Stil.
    await waehleZeile(
      await screen.findByRole('combobox', {
        name: texte.seiten.karten.stilWaehlen,
      }),
      'Fremd fremd · anderes',
    )

    // Das Set wechselt, **bevor** die Antwort kommt.
    wechsleSet(stand)

    await act(async () => {
      loese?.()
      await Promise.resolve()
    })

    expect(
      screen.queryByText(texte.seiten.karten.stilFrage),
    ).not.toBeInTheDocument()
  })

  // **Und ein verspäteter Fehler erscheint nicht mehr.**
  it('zeigt einen verspäteten Fehler nicht mehr an', async () => {
    let scheitere: (() => void) | undefined
    const Stilfolgen = vi.fn(
      (): Promise<FolgenAntwort> =>
        new Promise((_, ablehnen) => {
          scheitere = () => {
            ablehnen(new Error('zu spät'))
          }
        }),
    )

    belegeStil({ Stilfolgen })
    const stand = zeige()

    await waehleZeile(
      await screen.findByRole('combobox', {
        name: texte.seiten.karten.stilWaehlen,
      }),
      'Fremd fremd · anderes',
    )

    wechsleSet(stand)

    await act(async () => {
      scheitere?.()
      await Promise.resolve()
    })

    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})

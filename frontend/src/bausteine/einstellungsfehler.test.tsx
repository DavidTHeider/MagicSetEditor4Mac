// **Die Meldung steht an der Einstellung, die sie verursacht hat** — §6,
// gefunden am nativen Fenster für OPS-02
// (#1719):
// ein abgewiesener Ordner stand als „Das ließ sich nicht laden." über der
// Seite. `Einstellungsherkunft` steht an jeder Einstellung — Auswahl,
// Belegung, eigene Farben —, und dort erscheint sie jetzt.
import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'

import { Einstellungsherkunft } from './Einstellungsherkunft'
import { Einstellungsraum } from '../haken/einstellungsraum'
import { Bindungsfehler } from '../bindungsfehler'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { Einstellungsstand } from '../haken/useEinstellungen'

function stand(schluessel: string): Einstellungsstand {
  return {
    wert: () => '',
    eintraege: [],
    imBereich: () => [],
    vorgaben: () => [],
    dienstschalter: () => [],
    fragewahlen: () => [],
    kizeilen: () => [],
    geladen: true,
    ursache: undefined,
    fehlschlag: {
      schluessel,
      ursache: new Bindungsfehler({
        code: 'validierung',
        schluessel: 'ordnerFehlt',
        meldung: 'nicht erreichbar',
        werte: { ordner: '/Volumes/Weg' },
      }),
    },
    setze: vi.fn(),
    laeuft: () => false,
  }
}

function zeige(schluessel: string) {
  render(
    <Sprachraum sprache="de">
      <Einstellungsraum stand={stand('vorgabe.sicherungspfad')}>
        <Einstellungsherkunft
          eintrag={{
            schluessel,
            wert: '',
            herkunft: 'vorgabe',
            ebene: 'geraet',
          }}
        />
      </Einstellungsraum>
    </Sprachraum>,
  )
}

it('zeigt die Meldung an der Einstellung, deren Schreiben scheiterte', () => {
  zeige('vorgabe.sicherungspfad')

  expect(screen.getByRole('alert')).toHaveTextContent(
    kataloge.de.fehler.ordnerFehlt.replace('{ordner}', '/Volumes/Weg'),
  )
})

it('zeigt sie an keiner anderen', () => {
  zeige('vorgabe.ausfuhrpfad')

  expect(screen.queryByRole('alert')).toBeNull()
})

// **Die ausdrückliche Rückkehrmeldung** — A05
// (#2219),
// Nachlauf über Phase A.
//
// Scheitert das Zurückspielen, sagt die Seite, was danach dasteht: der
// bisherige Bestand unverändert — oder, wenn auch der Weg zurück scheiterte,
// wo seine Teile liegen. Über die Seite und nicht über den Haken, aus
// demselben Grund wie in `appbestandspflege.test.tsx`.
import type { Rueckspielantwort, Sicherungsauswahl } from './api'
import { Bindungsfehler } from './bindungsfehler'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { zuGebiet } from './test/gebiete'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'

afterEach(() => {
  raeumeApi()
})

const worte = kataloge.de.seiten.bestandspflege

const eineSicherung: Sicherungsauswahl = {
  pfad: '/Users/probe/mse-sicherung.msesicherung',
  angelegtAm: '2026-08-01T09:12:33.000Z',
  schemastand: '0012_paketkurzname.sql',
  sets: 3,
  karten: 107,
  bilder: 'ersetzt',
  abgebrochen: false,
  verschluesselt: false,
  paketeEnthalten: [],
  paketeFehlen: [],
}

// scheitertMit spielt zurück, und das Zurückspielen scheitert — mit einer
// Antwort oder, wie am Tor, mit einer Abweisung.
async function scheitertMit(zurueckspielen: () => Promise<Rueckspielantwort>) {
  belegeApi({
    ...belegeFenster('de'),
    SicherungWaehlen: () => Promise.resolve(eineSicherung),
    SicherungZurueckspielen: zurueckspielen,
  })

  const nutzer = userEvent.setup()

  render(<App />)

  await nutzer.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.einstellungen.titel,
    }),
  )
  await zuGebiet(kataloge.de.seiten.einstellungen.gruppen.daten)
  await screen.findByRole('heading', { name: worte.titel })

  await nutzer.click(screen.getByRole('button', { name: worte.waehlen }))
  await screen.findByText(worte.rueckfrage)
  await nutzer.click(screen.getByRole('button', { name: worte.bestaetigen }))

  return screen.findByRole('alert')
}

// mitAntwort scheitert mit einer Antwort, die sagt, was danach dasteht.
const mitAntwort = (unveraendert: boolean) => () =>
  Promise.resolve({
    ersetzt: false,
    beiseite: '',
    unveraendert,
    fehler: {
      code: 'abgebrochen' as const,
      schluessel: 'vorgangAbgebrochen',
      meldung: 'roh',
    },
  })

describe('die Rückkehrmeldung nach einem gescheiterten Zurückspielen', () => {
  it('sagt, dass der bisherige Bestand unverändert ist', async () => {
    const meldung = await scheitertMit(mitAntwort(true))

    expect(meldung).toHaveTextContent(kataloge.de.fehler.vorgangAbgebrochen)
    expect(meldung).toHaveTextContent(worte.rueckkehrUnveraendert)
    expect(meldung).not.toHaveTextContent('roh')
  })

  it('sagt, wo die Teile liegen, wenn auch der Weg zurück scheiterte', async () => {
    const meldung = await scheitertMit(mitAntwort(false))

    expect(meldung).toHaveTextContent(worte.rueckkehrUnvollstaendig)
    expect(meldung).not.toHaveTextContent(worte.rueckkehrUnveraendert)
  })

  // **Weist das Tor ab, wurde nichts angefasst** — die Meldung des Tors
  // steht allein da, ohne Rückkehrsatz.
  it('zeigt die Abweisung des Tors, wenn gerade getauscht wird', async () => {
    const meldung = await scheitertMit(() =>
      Promise.reject(
        new Bindungsfehler({
          code: 'belegt',
          schluessel: 'bestandWirdGetauscht',
          meldung: 'roh',
        }),
      ),
    )

    expect(meldung).toHaveTextContent(kataloge.de.fehler.bestandWirdGetauscht)
    expect(meldung).not.toHaveTextContent(worte.rueckkehrUnveraendert)
  })
})

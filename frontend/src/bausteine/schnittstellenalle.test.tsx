// **Alles mit einem Knopf** — G08
// (#2272).
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it } from 'vitest'

import type { SchnittstelleDTO, SchnittstellenAlleAntwort } from '../antworten'
import { kataloge } from '../sprache/kataloge'
import { Sprachraum } from '../sprache/kontext'
import { belegeApi, raeumeApi } from '../test/api'
import { Schnittstellen } from './Schnittstellen'

const eigene = kataloge.de.seiten.technik.schnittstellen

const ziel = (name: string, stand: string): SchnittstelleDTO => ({
  name,
  stand,
  art: '',
  datei: `/Users/jemand/.${name}`,
  zeilen: [`# Zeile für ${name}`],
})

const vorher = [
  ziel('befehlszeile', 'offen'),
  ziel('codex', 'fremd'),
  ziel('dienst', 'veraltet'),
  ziel('claudecode', 'nichtGefunden'),
]

function zeige() {
  return render(
    <Sprachraum sprache="de">
      <Schnittstellen />
    </Sprachraum>,
  )
}

afterEach(raeumeApi)

describe('Alles auf einmal', () => {
  it('zeigt vorher jede Datei und Zeile, die hinzukommt, und richtet alles ein', async () => {
    const nachher: SchnittstellenAlleAntwort = {
      schnittstellen: [
        ziel('befehlszeile', 'eingerichtet'),
        ziel('codex', 'fremd'),
        ziel('dienst', 'eingerichtet'),
        ziel('claudecode', 'nichtGefunden'),
      ],
      ausgaenge: [
        { name: 'befehlszeile', ausgang: 'eingerichtet', fehler: null },
        { name: 'codex', ausgang: 'fremd', fehler: null },
        {
          name: 'dienst',
          ausgang: 'gescheitert',
          fehler: {
            code: 'unbekannt',
            schluessel: 'unbekannt',
            meldung: 'roh',
            werte: {},
          },
        },
        { name: 'erfunden', ausgang: 'irgendwie', fehler: null },
      ],
      fehler: null,
    }
    let gerufen = 0
    belegeApi({
      Schnittstellen: () =>
        Promise.resolve({
          schnittstellen: vorher,
          nurEintrag: '',
          fehler: null,
        }),
      SchnittstellenAlleEinrichten: () => {
        gerufen++

        return Promise.resolve(nachher)
      },
    })

    zeige()

    // **Die Vorschau nennt die zwei, die fehlen** — nicht den fremden und
    // nicht den, den es hier nicht gibt.
    const vorschau = await screen.findByText(
      eigene.alle.vorschau.replace('{anzahl}', '2'),
    )
    const liste = vorschau.closest('details')
    if (liste === null) {
      throw new Error('keine Vorschau')
    }

    expect(
      within(liste).getByText('# Zeile für befehlszeile'),
    ).toBeInTheDocument()
    expect(liste).toHaveTextContent('/Users/jemand/.dienst')
    expect(liste).not.toHaveTextContent('/Users/jemand/.codex')

    await userEvent.click(
      screen.getByRole('button', { name: eigene.alle.einrichten }),
    )

    const ergebnis = await screen.findByRole('status')
    expect(gerufen).toBe(1)
    expect(ergebnis).toHaveTextContent(eigene.namen.befehlszeile)
    expect(ergebnis).toHaveTextContent(eigene.alle.ausgaenge.fremd)
    expect(ergebnis).toHaveTextContent(eigene.alle.ausgaenge.gescheitert)
    expect(ergebnis).toHaveTextContent(kataloge.de.fehler.unbekannt)
    expect(ergebnis).toHaveTextContent('erfunden')
    expect(ergebnis).toHaveTextContent('irgendwie')

    // **Danach fehlt nichts mehr** — der Stand ist der aus der Antwort.
    expect(
      await screen.findByText(eigene.alle.vorschauLeer),
    ).toBeInTheDocument()
    expect(
      screen.getByRole('button', { name: eigene.alle.einrichten }),
    ).toBeDisabled()
  })

  it('nimmt alles zurück und meldet einen Fehler der Antwort', async () => {
    belegeApi({
      Schnittstellen: () =>
        Promise.resolve({
          schnittstellen: [ziel('befehlszeile', 'eingerichtet')],
          nurEintrag: '',
          fehler: null,
        }),
      SchnittstellenAlleWiderrufen: () =>
        Promise.resolve({
          schnittstellen: [ziel('befehlszeile', 'eingerichtet')],
          ausgaenge: [
            {
              name: 'befehlszeile',
              ausgang: 'gescheitert',
              fehler: {
                code: 'voll',
                schluessel: 'voll',
                meldung: 'roh',
                werte: {},
              },
            },
          ],
          fehler: {
            code: 'voll',
            schluessel: 'voll',
            meldung: 'roh',
            werte: {},
          },
        }),
    })

    zeige()

    expect(await screen.findByText(eigene.alle.vorschauLeer)).toBeVisible()

    await userEvent.click(
      screen.getByRole('button', { name: eigene.alle.widerrufen }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent(
      kataloge.de.fehler.voll,
    )
    expect(screen.getByRole('status')).toHaveTextContent(
      eigene.alle.ausgaenge.gescheitert,
    )
  })

  it('zeigt ohne Ziele keinen Block und sperrt das Zurücknehmen ohne Eingerichtetes', async () => {
    belegeApi({
      Schnittstellen: () =>
        Promise.resolve({
          schnittstellen: [ziel('befehlszeile', 'offen')],
          nurEintrag: '',
          fehler: null,
        }),
    })

    zeige()

    expect(
      await screen.findByRole('button', { name: eigene.alle.widerrufen }),
    ).toBeDisabled()
  })

  it('meldet, wenn der Ruf selbst scheitert', async () => {
    belegeApi({
      Schnittstellen: () =>
        Promise.resolve({
          schnittstellen: [ziel('befehlszeile', 'offen')],
          nurEintrag: '',
          fehler: null,
        }),
      SchnittstellenAlleEinrichten: () => Promise.reject(new Error('weg')),
    })

    zeige()

    await userEvent.click(
      await screen.findByRole('button', { name: eigene.alle.einrichten }),
    )

    expect(await screen.findByRole('alert')).toHaveTextContent('weg')
  })
})

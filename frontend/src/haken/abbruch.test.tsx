// **Ein abgebrochener Lauf hinterlässt keine Fehlermeldung**
// (#288).
//
// Die Vorschau bricht den vorigen Lauf ab, sobald ein neuer beginnt (§7) — das
// ist die Absicht. Über der gezeichneten Karte stand dabei „Es ist ein
// unerwarteter Fehler aufgetreten": die Meldung des Laufs, der genau so
// abgebrochen wurde, wie es vorgesehen ist.
//
// **Und das Vorige bleibt stehen.** Wer den Stand leerte, zeigte statt der
// Karte eine Ladeanzeige, die nie endet — den abgebrochenen Lauf holt niemand
// nach.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, expect, it } from 'vitest'

import { Bindungsfehler } from '../api'
import { Sprachraum } from '../sprache/kontext'
import { useLaden } from './useLaden'

function abbruch() {
  return new Bindungsfehler({
    code: 'abgebrochen',
    schluessel: 'unbekannt',
    meldung: 'Es ist ein unerwarteter Fehler aufgetreten.',
  })
}

function Probe({ holen }: { holen: () => Promise<string> }) {
  const stand = useLaden<string>(holen, ['fest'])

  return (
    <>
      <p data-testid="daten">{stand.daten ?? '—'}</p>
      <p data-testid="fehler">{stand.fehler ?? '—'}</p>
      <p data-testid="laedt">{stand.laedt ? 'ja' : 'nein'}</p>
    </>
  )
}

describe('useLaden bei einem Abbruch', () => {
  it('meldet ihn nicht als Fehler', async () => {
    render(
      <Sprachraum sprache="de">
        <Probe
          holen={() => {
            return Promise.reject(abbruch())
          }}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('laedt')).toHaveTextContent('nein')
    })

    expect(screen.getByTestId('fehler')).toHaveTextContent('—')
  })

  // **Ein gewöhnlicher Fehler bleibt einer.**
  it('meldet jeden anderen Fehler weiter', async () => {
    render(
      <Sprachraum sprache="de">
        <Probe
          holen={() => {
            return Promise.reject(new Error('Die Platte ist voll.'))
          }}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('fehler')).toHaveTextContent(
        'Die Platte ist voll.',
      )
    })
  })
})

// **Ein Abbruch ohne vorheriges Bild wird nachgeholt.**
//
// # Der Fall
//
// Zwei Läufe gehen gleichzeitig hinaus — beim Wechsel der Karte, oder weil
// React einen Effekt im Entwicklungsmodus zweimal ausführt. `App.Vorschau`
// bricht den **vorigen** ab; welcher das ist, entscheidet die Reihenfolge, in
// der Go sie annimmt. Diese Reihenfolge muss nicht die sein, in der das
// Fenster sie losgeschickt hat.
//
// Trifft es den Lauf, auf den `useLaden` wartet, stand hier bisher: „das
// Vorige bleibt stehen" — und wenn es kein Vorheriges gibt, bleibt **nichts**
// stehen. Keine Ladeanzeige, keine Meldung, kein Bild; für immer, denn den
// abgebrochenen Lauf holte niemand nach.
//
// In der laufenden Anwendung war das ein leeres Kartenfenster nach einem
// Kartenwechsel — nicht jedes Mal, sondern wenn das Rennen so ausging
// (#566).
//
// # Und nur dann
//
// Gibt es ein voriges Bild, bleibt es stehen: ein Nachschlag wäre dort eine
// zweite Antwort auf eine Frage, die schon beantwortet ist.
describe('useLaden bei einem Abbruch ohne Daten', () => {
  it('holt den Lauf nach', async () => {
    let rufe = 0

    render(
      <Sprachraum sprache="de">
        <Probe
          holen={() => {
            rufe++

            return rufe === 1
              ? Promise.reject(abbruch())
              : Promise.resolve('die Karte')
          }}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('daten')).toHaveTextContent('die Karte')
    })

    expect(screen.getByTestId('fehler')).toHaveTextContent('—')
  })

  it('lässt ein vorhandenes Bild stehen und holt nichts nach', async () => {
    let rufe = 0

    function MitNachladen() {
      const stand = useLaden<string>(() => {
        rufe++

        return rufe === 1
          ? Promise.resolve('die Karte')
          : Promise.reject(abbruch())
      }, ['fest'])

      return (
        <>
          <p data-testid="daten">{stand.daten ?? '—'}</p>
          <button type="button" onClick={stand.nachladen}>
            nachladen
          </button>
        </>
      )
    }

    render(
      <Sprachraum sprache="de">
        <MitNachladen />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('daten')).toHaveTextContent('die Karte')
    })

    // **Ein zweiter Lauf, und der wird abgebrochen.** Das Bild bleibt stehen —
    // und es wird **nicht** nachgeholt: die Frage ist beantwortet.
    await userEvent.click(screen.getByRole('button', { name: 'nachladen' }))

    await waitFor(() => {
      expect(rufe).toBe(2)
    })

    expect(screen.getByTestId('daten')).toHaveTextContent('die Karte')

    // **Und es bleibt bei zweien.** Ein dritter Lauf wäre der Nachschlag, den
    // es hier nicht geben darf.
    await new Promise((weiter) => setTimeout(weiter, 50))
    expect(rufe).toBe(2)
  })

  it('holt nicht nach, wenn der erste Lauf durchkommt', async () => {
    let rufe = 0

    render(
      <Sprachraum sprache="de">
        <Probe
          holen={() => {
            rufe++

            return rufe === 1
              ? Promise.resolve('die Karte')
              : Promise.reject(abbruch())
          }}
        />
      </Sprachraum>,
    )

    await waitFor(() => {
      expect(screen.getByTestId('daten')).toHaveTextContent('die Karte')
    })

    // **Kein zweiter Lauf.** Der erste hat geantwortet; es gibt nichts
    // nachzuholen.
    expect(rufe).toBe(1)
  })
})

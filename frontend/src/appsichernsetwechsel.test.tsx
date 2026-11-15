// **Ein Setwechsel räumt die Meldung des Sicherns** —
// #749.
//
// # Der Befund
//
// `useSetSpeichern` trägt Antwort und Fehler ein, **ohne den Stand zu prüfen**.
// `raeumen()` hängt allein am Schließen der Meldung, nicht am Setwechsel. Wer
// ⌘S drückt, das Set wechselt und dann die Antwort bekommt, liest sie unter dem
// neuen.
//
// **Bei der Erfolgsmeldung ist das harmlos** — sie nennt den Pfad und ist damit
// eindeutig. **Bei der Fehlermeldung nicht**: sie trägt keinen Bezug, und wer
// nach dem Wechsel „Der Bestand ist gesperrt" liest, bezieht es auf das Set,
// das er gerade ansieht.
//
// # Warum kein `key` wie auf der Lasche „Ausgeben"
//
// Weil der Haken in `App` sitzt und `App` nicht neu gebaut wird, wenn ein Set
// wechselt — sie **ist** das Fenster. Die Meldung bekommt deshalb ihr Set mit,
// und die Anzeige vergleicht.
//
// **Und kein Aufräumeffekt auf `setID`**: der liefe vor der Antwort, nicht nach
// ihr. Ein Lauf, der eine Sekunde später zurückkommt, schriebe danach wieder
// hinein (§7, Regel 2).
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { leisteneintrag } from './test/leistenweg'

const texte = kataloge.de

function satz(id: string, name: string) {
  return {
    id,
    name,
    spiel: 'probe',
    stil: 'schlicht',
    herkunft: `/pfad/${name}.mse-set`,
    karten: 1,
    herkunftFehlt: false,
  }
}

function band(): Promise<HTMLElement> {
  return screen.findByRole('status', {
    name: texte.allgemein.meldungsband.bereich,
  })
}

async function oeffne(name: string) {
  await userEvent.click(await leisteneintrag('sets'))

  await userEvent.click(
    await screen.findByRole('button', {
      name: `${name} — ${texte.seiten.sets.oeffnen}`,
    }),
  )
}

beforeEach(() => {
  belegeApi({ ...belegeFenster('de') })
})

afterEach(() => {
  raeumeApi()
})

describe('Die Meldung des Sicherns', () => {
  it('steht nicht mehr da, wenn inzwischen ein anderes Set offen ist', async () => {
    let scheitern: (grund: Error) => void = () => undefined

    // **Ein Lauf, dessen Ende der Test bestimmt.** Anders ließe sich der
    // Wechsel nicht **zwischen** Anstoß und Antwort legen — und genau dort
    // liegt der Fall.
    const SetSpeichern = vi.fn(
      (): Promise<never> =>
        new Promise((_, ablehnen) => {
          scheitern = ablehnen
        }),
    )

    belegeApi({
      ...belegeFenster('de'),
      Sets: () =>
        Promise.resolve({ sets: [satz('7', 'Erstes'), satz('8', 'Zweites')] }),
      SetSpeichern,
      Karten: () => Promise.resolve({ karten: [] }),
    })

    render(<App />)

    const meldungen = await band()

    await oeffne('Erstes')

    fireEvent.keyDown(window, { key: 's', metaKey: true })

    await waitFor(() => {
      expect(SetSpeichern).toHaveBeenCalledWith('7')
    })

    // **Der Wechsel, während der Lauf noch läuft.**
    await oeffne('Zweites')

    scheitern(new Error('Der Bestand ist gesperrt.'))

    // **Erst ein Abschlusssignal, dann die Verneinung** (§3): ohne einen Beleg,
    // dass die Antwort angekommen ist, träfe „nichts steht da" auch dann zu,
    // wenn sie noch unterwegs wäre.
    await waitFor(() => {
      expect(SetSpeichern).toHaveBeenCalledTimes(1)
    })

    await waitFor(() => {
      expect(meldungen).toBeEmptyDOMElement()
    })
  })
})

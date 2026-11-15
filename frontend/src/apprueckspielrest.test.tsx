// **Ein unvollendetes Zurückspielen steht im Band** — der zweite Zweitblick
// über die Freigabe 0.3.0, Befund 2
// (#1170).
//
// # Der Befund
//
// Scheitert der beim Start nachgeholte Bildertausch, stand er **nur im
// Protokoll**. Das Fenster öffnete danach wortlos die schon getauschte
// Datenbank neben den alten Bildern — sichtbar zusammengekommen, und niemand
// erfuhr es.
//
// # Was hier geprüft wird
//
// Dass der Satz aus dem Dienst bis ins Band durchkommt. Die Go-Seite hat ihren
// eigenen Lauf; das hier ist die **Naht**, und §3 nennt sie als die Stelle, an
// der beide Prüfebenen aneinander vorbeisehen.
import { render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'

afterEach(raeumeApi)

// band gibt den Bereich, in dem die Meldungen des Fensters stehen.
function band(): Promise<HTMLElement> {
  return screen.findByRole('status', {
    name: kataloge.de.allgemein.meldungsband.bereich,
  })
}

const rest = 'Ein abgebrochenes Zurückspielen ließ sich nicht abschließen.'

describe('ein unvollendetes Zurückspielen', () => {
  it('steht im Band über dem Inhalt', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Startbefund: () =>
        Promise.resolve({
          offen: true,
          grund: '',
          schluessel: '',
          fassung: '0.3.0',
          rueckspielrest: rest,
          rueckweg: null,
        }),
      Sets: () => Promise.resolve({ sets: [] }),
    })

    render(<App />)

    expect(await band()).toHaveTextContent(rest)
  })

  it('bleibt weg, wenn nichts liegen blieb', async () => {
    belegeApi({
      ...belegeFenster('de'),
      Sets: () => Promise.resolve({ sets: [] }),
    })

    render(<App />)

    await screen.findByRole('navigation')

    expect(screen.queryByText(rest)).not.toBeInTheDocument()
  })
})

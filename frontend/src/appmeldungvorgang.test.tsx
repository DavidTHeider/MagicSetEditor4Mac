// **Eine Meldung bleibt bei ihrem Vorgang** — S08.NAT-02 aus UX-08
// (#1680),
// Befund F10 der nativen Sichtprüfung.
//
// # Der Befund
//
// Ein gescheitertes Sichern stand als „Eintrag nicht gefunden." im Band —
// und blieb dort, während jemand Turniere, Einstellungen und Hilfe öffnete.
// Über jeder dieser Seiten sah es aus wie **ihr** Fehler; erst ein
// Profilwechsel räumte es weg.
//
// # Was jetzt gilt
//
// **Die Meldung eines Sets steht nur auf den Seiten eines Sets** — der
// Setliste und den Laschen —, und sie nennt ihren Vorgang und ihr Set. Wer
// zurückkommt, findet sie wieder: verschwunden ist sie nicht, nur nicht an
// fremder Stelle. Ein Zeitgeber räumt sie nicht (S08, „keine nur
// zeitgesteuerte Lösung").
import { fireEvent, render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { leisteneintrag } from './test/leistenweg'

const texte = kataloge.de

function band(): Promise<HTMLElement> {
  return screen.findByRole('status', {
    name: texte.allgemein.meldungsband.bereich,
  })
}

afterEach(() => {
  raeumeApi()
})

it('zeigt das gescheiterte Sichern nur auf den Seiten des Sets', async () => {
  const SetSpeichern = vi
    .fn()
    .mockRejectedValue(new Error('Eintrag nicht gefunden.'))

  belegeApi({
    ...belegeFenster('de'),
    Sets: vi.fn().mockResolvedValue({
      sets: [
        {
          id: '7',
          name: 'Mein Probeset',
          spiel: 'probe',
          stil: 'schlicht',
          herkunft: '/pfad/probeset.mse-set',
          karten: 1,
          herkunftFehlt: false,
        },
      ],
    }),
    Karten: vi.fn().mockResolvedValue({ karten: [] }),
    SetSpeichern,
  })

  render(<App />)

  await userEvent.click(await leisteneintrag('sets'))
  await userEvent.click(
    await screen.findByRole('button', {
      name: `Mein Probeset — ${texte.seiten.sets.oeffnen}`,
    }),
  )

  fireEvent.keyDown(window, { key: 's', metaKey: true })

  // **Der Vorgang und das Set stehen im Satz** — nicht nur der Grund.
  const meldungen = await band()
  await waitFor(() => {
    expect(meldungen).toHaveTextContent(
      'Das Set „Mein Probeset“ wurde nicht gesichert — Eintrag nicht gefunden.',
    )
  })

  // **Auf einer fremden Seite steht sie nicht.**
  await userEvent.click(await leisteneintrag('hilfe'))
  await screen.findAllByRole('heading', { name: texte.seiten.hilfe.titel })
  expect(meldungen).toBeEmptyDOMElement()

  // **Und zurück auf der Setliste steht sie wieder** — der Vorgang ist
  // gescheitert, und das gilt weiter.
  await userEvent.click(await leisteneintrag('sets'))
  await waitFor(() => {
    expect(meldungen).toHaveTextContent('Eintrag nicht gefunden.')
  })
})

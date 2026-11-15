// **Nach einem Neuanfang liest die Oberfläche neu** — #2292
// (#2292).
//
// Nach dem Zurückspielen räumt `Anwendung` offenes Set und Karte und lädt
// alles neu (`appbestandspflege.test.tsx`, „lädt nach dem Zurückspielen
// alles neu"). Ein Neuanfang über die Rücksetzungsseite wechselt den Bestand
// noch vollständiger — die ganze Datei geht beiseite —, und die Seite meldete
// es nicht nach oben: Einstellungen und Bezugsobjekt blieben aus dem alten
// Stand. **Und die Bilanz bleibt stehen**: sie nennt den Ort des alten
// Bestands, und der Weg darf sie nicht abräumen.
//
// **Dieselbe Klasse eine Stufe leichter**: wer auf derselben Seite nur die
// Einstellungen zurücksetzt, sah sie bis zum nächsten Start im alten Stand.
// Die Menüleiste las dafür schon immer neu (UX-09, `gestaltungsraum.tsx`) —
// die Seite nicht.
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { App } from './App'
import { kataloge } from './sprache/kataloge'
import { belegeApi, belegeFenster, raeumeApi } from './test/api'
import { zuAbschnitt } from './test/gebiete'

afterEach(raeumeApi)

const worte = kataloge.de.seiten.ruecksetzung

// zuruecksetzen fährt eine Rücksetzung über die Seite und gibt die Zählung
// der Einstellungsabrufe samt ihrem Stand **vor** dem letzten Klick.
async function zuruecksetzen(art: string, beiseite: string) {
  const Einstellungen = vi.fn(() =>
    Promise.resolve({
      einstellungen: [
        {
          schluessel: 'sprache',
          wert: 'de',
          wirksam: 'de',
          werte: ['de', 'en'],
          bereich: 'anwendung',
        },
      ],
    }),
  )
  const freigabe = art === 'anwendung'

  belegeApi({
    ...belegeFenster('de'),
    Einstellungen,
    Ruecksetzungsarten: () => Promise.resolve([art]),
    Ruecksetzungsvorschau: () =>
      Promise.resolve({
        art,
        betroffen: [],
        hindernisse: [],
        bereit: true,
        nameNoetig: false,
        nachweisNoetig: freigabe,
        sicherungEmpfohlen: true,
        revision: 'marke',
        fehler: null,
      }),
    RuecksetzungFreigabeErteilen: () =>
      Promise.resolve({
        kennung: 'freigabe',
        laeuftab: '2026-10-10T20:00:00Z',
        fehler: null,
      }),
    Zuruecksetzen: () =>
      Promise.resolve({ art, zurueckgesetzt: 0, beiseite, fehler: null }),
  })

  const nutzer = userEvent.setup()
  render(<App />)

  await nutzer.click(
    await screen.findByRole('button', {
      name: kataloge.de.seiten.einstellungen.titel,
    }),
  )
  await zuAbschnitt(
    kataloge.de.seiten.einstellungen.gruppen.daten,
    kataloge.de.seiten.einstellungen.abschnitte.eingriffe,
  )

  await waitFor(() => {
    expect(Einstellungen).toHaveBeenCalled()
  })

  await nutzer.click(
    await screen.findByRole('button', { name: worte.nachsehen }),
  )

  if (freigabe) {
    await nutzer.click(
      await screen.findByRole('button', { name: worte.freigeben }),
    )
  }

  const vorher = Einstellungen.mock.calls.length
  await nutzer.click(await screen.findByRole('button', { name: worte.tun }))

  return { Einstellungen, vorher }
}

describe('Eine Rücksetzung über die Seite', () => {
  it('lädt nach einem Neuanfang alles neu und lässt die Bilanz stehen', async () => {
    const { Einstellungen, vorher } = await zuruecksetzen(
      'anwendung',
      '/Users/probe/vor-dem-neuanfang.db',
    )

    expect(await screen.findByText(/vor-dem-neuanfang\.db/)).toBeInTheDocument()
    await waitFor(() => {
      expect(Einstellungen.mock.calls.length).toBeGreaterThan(vorher)
    })
    expect(screen.getByText(/vor-dem-neuanfang\.db/)).toBeInTheDocument()
  })

  it('liest nach zurückgesetzten Einstellungen die Einstellungen neu', async () => {
    const { Einstellungen, vorher } = await zuruecksetzen(
      'profileinstellungen',
      '',
    )

    await waitFor(() => {
      expect(Einstellungen.mock.calls.length).toBeGreaterThan(vorher)
    })
  })
})

// **Der Beginn am Tisch** — Starthand, Mulligan, Unterlegen und Behalten,
// CR 103.5 (#1839).
//
// # Wogegen hier geprüft wird
//
// **Gegen den Tisch vom 2026-09-24**: mit dem Set „Green“ lag jede Bibliothek
// in der Reihenfolge der Deckliste, niemand hatte eine Hand, und einen
// Mulligan gab es nicht. Wer spielen wollte, mischte von Hand und schob
// sieben Karten einzeln hinüber.
//
// **Die Tafel steht über dem Tisch**, wie die Legendenwahl: der Beginn hält
// das Spiel an, und wer ihn unter den Befehlen suchen müsste, hielte den
// Tisch für hängengeblieben.
import { screen, within } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { einLauf, einObjekt, einePartie, einSitz, element } from './partieprobe'
import { kataloge } from '../sprache/kataloge'
import { zeige } from './partietischprobe'

import type { Sitz } from '../apipartie'

const eigene = kataloge.de.seiten.spielen
const namen = kataloge.de.seiten.handlungen.handlungsname

// imBeginn gibt eine Partie, deren eigener Sitz so steht, wie verlangt — mit
// drei Karten auf der Hand.
function imBeginn(anna: Partial<Sitz>, bert: Partial<Sitz> = {}) {
  return einePartie({
    sitze: [
      einSitz({ kennung: 'sitz1', name: 'Anna', ...anna }),
      einSitz({
        kennung: 'sitz2',
        name: 'Bert',
        eroeffnung: 'behalten',
        ...bert,
      }),
    ],
    objekte: ['Wald', 'Bär', 'Riese'].map((name, stelle) =>
      einObjekt({
        kennung: `sitz1/h${String(stelle)}`,
        name,
        zone: 'hand',
        beherrscher: 'sitz1',
        geheim: false,
      }),
    ),
  })
}

// tafel gibt die Tafel des Beginns.
function tafel() {
  return screen.getByRole('region', { name: eigene.beginnTitel })
}

describe('der Beginn am Tisch', () => {
  test('bietet vor der Starthand genau sie an', async () => {
    const befehlen = vi.fn()
    const nutzer = zeige(
      einLauf({ befehlen }),
      imBeginn({ eroeffnung: 'ausstehend' }),
    )

    expect(tafel()).toHaveTextContent(eigene.beginnStarthand)
    expect(
      within(tafel()).queryByRole('button', {
        name: namen['spieler.behalten'],
      }),
    ).toBeNull()

    await nutzer.click(
      within(tafel()).getByRole('button', {
        name: namen['spieler.starthand'],
      }),
    )

    expect(befehlen).toHaveBeenCalledWith('spieler.starthand', {})
  })

  test('lässt eine offene Hand behalten oder neu ziehen', async () => {
    const befehlen = vi.fn()
    const nutzer = zeige(
      einLauf({ befehlen }),
      imBeginn({ eroeffnung: 'offen' }),
    )

    expect(tafel()).toHaveTextContent('Deine Starthand hat 3 Karten.')
    expect(tafel()).toHaveTextContent(eigene.beginnMulligan)

    await nutzer.click(
      within(tafel()).getByRole('button', { name: namen['spieler.mulligan'] }),
    )
    await nutzer.click(
      within(tafel()).getByRole('button', { name: namen['spieler.behalten'] }),
    )

    expect(befehlen).toHaveBeenNthCalledWith(1, 'spieler.mulligan', {})
    expect(befehlen).toHaveBeenNthCalledWith(2, 'spieler.behalten', {})
  })

  // **Eine Karte ist eine Karte** — nach sechs Mulligans zu zweit bleibt
  // genau eine (CR 103.5).
  test('und sagt es richtig, wenn nur eine Karte bleibt', () => {
    const partie = imBeginn({ eroeffnung: 'offen', mulligans: 6 })

    zeige(einLauf(), { ...partie, objekte: partie.objekte.slice(0, 1) })

    expect(tafel()).toHaveTextContent('Deine Starthand hat eine Karte.')
  })

  // **Aus `erlaubt` und nicht aus einer Rechnung daneben**: wann kein
  // Mulligan mehr geht, weiß der Kern (CR 103.5, CR 103.5c).
  test('und zeigt keinen Mulligan, den der Kern nicht annähme', () => {
    zeige(
      einLauf({ erlaubt: ['spieler.behalten'] }),
      imBeginn({ eroeffnung: 'offen', mulligans: 7 }),
    )

    expect(
      within(tafel()).queryByRole('button', {
        name: namen['spieler.mulligan'],
      }),
    ).toBeNull()
    expect(
      within(tafel()).getByRole('button', { name: namen['spieler.behalten'] }),
    ).toBeInTheDocument()
  })

  test('legt genau die geschuldeten Karten unter, in der gewählten Reihenfolge', async () => {
    const befehlen = vi.fn()
    const nutzer = zeige(
      einLauf({ befehlen }),
      imBeginn({ eroeffnung: 'offen', mulligans: 2, unterlegen: 2 }),
    )

    expect(tafel()).toHaveTextContent('Lege 2 Karten unter deine Bibliothek.')

    const unterlegen = within(tafel()).getByRole('button', {
      name: namen['spieler.unterlegen'],
    })

    // **Solange nicht genau zwei gewählt sind, geht es nicht.**
    expect(unterlegen).toBeDisabled()

    await nutzer.click(within(tafel()).getByRole('checkbox', { name: 'Riese' }))
    expect(unterlegen).toBeDisabled()
    expect(tafel()).toHaveTextContent('1 von 2 gewählt')

    await nutzer.click(within(tafel()).getByRole('checkbox', { name: 'Wald' }))
    expect(unterlegen).toBeEnabled()

    await nutzer.click(unterlegen)

    expect(befehlen).toHaveBeenCalledWith('spieler.unterlegen', {
      objekte: ['sitz1/h2', 'sitz1/h0'],
    })
  })

  test('und nimmt eine abgewählte Karte wieder heraus', async () => {
    const nutzer = zeige(
      einLauf(),
      imBeginn({ eroeffnung: 'offen', mulligans: 1, unterlegen: 1 }),
    )

    expect(tafel()).toHaveTextContent('Lege eine Karte unter deine Bibliothek.')

    const baer = within(tafel()).getByRole('checkbox', { name: 'Bär' })

    await nutzer.click(baer)
    await nutzer.click(baer)

    expect(baer).not.toBeChecked()
    expect(tafel()).toHaveTextContent('0 von 1 gewählt')
  })

  test('nennt, wer am Tisch noch nicht so weit ist', () => {
    zeige(
      einLauf(),
      imBeginn({ eroeffnung: 'behalten' }, { eroeffnung: 'offen' }),
    )

    expect(tafel()).toHaveTextContent('Bert entscheidet über die Hand.')
    expect(within(tafel()).queryAllByRole('button')).toHaveLength(0)
  })

  test('und sagt es für jeden Stand', () => {
    zeige(
      einLauf(),
      einePartie({
        sitze: [
          einSitz({ kennung: 'sitz1', name: 'Anna', eroeffnung: 'behalten' }),
          einSitz({ kennung: 'sitz2', name: 'Bert', eroeffnung: 'ausstehend' }),
          einSitz({
            kennung: 'sitz3',
            name: 'Cleo',
            eroeffnung: 'offen',
            unterlegen: 1,
          }),
        ],
      }),
    )

    expect(tafel()).toHaveTextContent('Bert hat noch keine Starthand gezogen.')
    expect(tafel()).toHaveTextContent('Cleo legt Karten unter die Bibliothek.')
  })

  test('und schweigt, wo niemand mehr beginnt', () => {
    zeige(einLauf(), imBeginn({ eroeffnung: 'behalten' }))

    expect(
      screen.queryByRole('region', { name: eigene.beginnTitel }),
    ).toBeNull()
  })

  // **Eine Partie von vorher beginnt nicht noch einmal** — ihre Sitze tragen
  // keinen Stand.
  test('und in einer Partie von vorher ebenso', () => {
    zeige(einLauf(), imBeginn({ eroeffnung: '' }, { eroeffnung: '' }))

    expect(
      screen.queryByRole('region', { name: eigene.beginnTitel }),
    ).toBeNull()
  })

  test('und hält die Knöpfe an, solange ein Befehl läuft', () => {
    zeige(
      einLauf({ laeuft: () => true }),
      imBeginn({ eroeffnung: 'ausstehend' }),
    )

    expect(
      element(
        within(tafel()).getByRole('button', {
          name: namen['spieler.starthand'],
        }),
      ),
    ).toBeDisabled()
  })
})

// **Nach dem Ende fragt der Tisch nichts mehr** — P4
// (#2372):
// endet eine Partie mitten im Beginn (jemand gibt auf), stand die
// Starthand-Frage weiter da. Das Ergebnis steht oben.
test('zeigt an einer beendeten Partie keine Frage des Beginns', () => {
  zeige(einLauf(), {
    ...imBeginn({ eroeffnung: 'ausstehend' }),
    stand: 'beendet',
  })

  expect(screen.queryByRole('region', { name: eigene.beginnTitel })).toBeNull()
})

// **Wie der Tisch angeordnet ist** — Befund F09 der nativen Prüfung von
// 0.5.0 (#1675,
// #1680)
// und das Artboard *Tisch*.
//
// # Wogegen hier geprüft wird
//
// **Dass die Leinwand vorn steht und nicht am unteren Fensterrand.** Vor
// ihr standen bis zum 2026-09-24 zwei Erklärabsätze, die Lebensleiste, die
// Zugleiste und der Darstellungsschalter. Geprüft wird die **Reihenfolge**
// im Dokument und die **Höhe**, die der Tisch bekommt — was jsdom nicht
// zeigt, ist, wie es aussieht; das zeigt die WebKit-Probe
// (`scripts/webkitprobe/abnahme/play06-tisch.probe`).
import { screen } from '@testing-library/react'
import { afterEach, expect, it } from 'vitest'

import { raeumeApi } from '../test/api'
import { kataloge } from '../sprache/kataloge'
import { kurzform } from '../sprache/kurzform'
import { einsetzen } from '../sprache/einsetzen'

import { einLauf, einePartie, einZug, element } from './partieprobe'
import { zeige } from './partietischprobe'

const eigene = kataloge.de.seiten.spielen
const bausteine = kataloge.de.bausteine

afterEach(() => {
  raeumeApi()
})

// einLaufenderZug ist eine Partie, deren Zug läuft — sonst zeichnet die
// Zugleiste nichts, und die Reihenfolge wäre nicht zu prüfen.
function einLaufenderZug() {
  return einePartie({
    zug: einZug({
      nummer: 3,
      aktiv: 'sitz1',
      prioritaet: 'sitz1',
      phase: 'kampf',
      schritt: 'blocker',
    }),
  })
}

it('stellt die Leinwand vor den Zug und die Spieler', () => {
  zeige(einLauf(), einLaufenderZug(), false)

  const tisch = screen.getByRole('application', { name: eigene.tischBild })

  for (const danach of [
    screen.getByRole('region', { name: eigene.zugLeiste }),
    screen.getByRole('region', { name: eigene.standLeiste }),
  ]) {
    expect(
      tisch.compareDocumentPosition(danach) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  }
})

// **jsdom rechnet kein Layout**: das Fenster ist 768 hoch, und der Tisch
// beginnt bei null. Bis an den Innenrand von `main` (32) bleiben 736.
it('gibt dem Tisch die Höhe bis an den Fensterrand', async () => {
  zeige(einLauf(), einLaufenderZug(), false)

  const tisch = screen.getByRole('application', { name: eigene.tischBild })
  const traeger = tisch.closest<HTMLElement>('[style*="--tischhoehe"]')

  await expect
    .poll(() => traeger?.style.getPropertyValue('--tischhoehe'))
    .toBe('736px')
})

// **Der Wirt und das Aufgeben werden weiter erklärt** — hinter einem
// Fragezeichen bei dem, was sie erklären, und nicht mehr als Absatz, der
// jedem Tisch zwei Zeilen Leinwand nahm.
it('erklärt Wirt und Aufgeben hinter einem Fragezeichen', () => {
  zeige(einLauf(), einLaufenderZug(), false)

  expect(
    screen.queryByText(kurzform(eigene.wirtHinweis)),
  ).not.toBeInTheDocument()
  expect(
    screen.queryByText(kurzform(eigene.aufgebenHinweis)),
  ).not.toBeInTheDocument()

  for (const feld of [eigene.wirt, eigene.aufgeben]) {
    expect(
      screen.getByRole('button', {
        name: einsetzen(bausteine.feldhilfe, { feld }),
      }),
    ).toBeInTheDocument()
  }
})

// **Die Ausgabe des Tischbilds steht hinter einem Knopf** — offen nahm sie
// dem Tisch vier Zeilen, auch wenn niemand ein Bild wollte.
it('öffnet die Ausgabe des Tischbilds erst auf Zuruf', async () => {
  const hand = zeige(einLauf(), einLaufenderZug(), false)

  expect(
    screen.queryByRole('heading', { name: eigene.abzugTitel }),
  ).not.toBeInTheDocument()

  const knopf = screen.getByRole('button', { name: eigene.abzugTitel })
  expect(knopf).toHaveAttribute('aria-expanded', 'false')

  await hand.click(knopf)

  expect(
    await screen.findByRole('heading', { name: eigene.abzugTitel }),
  ).toBeInTheDocument()
  expect(knopf).toHaveAttribute('aria-expanded', 'true')
})

// **Was man nicht während jedes Zuges braucht, steht unter dem Tisch** —
// Anhalten, Beenden, der Verlauf und die Freigabe. Das Artboard *Tisch* hat
// in seiner Spalte Zug, Karte und Handlungen und sonst nichts; bis zum
// 2026-09-24 standen die vier dort und schoben die Ereignisse aus dem Blick.
it('stellt Anhalten, Verlauf und Freigabe unter den Tisch', () => {
  zeige(einLauf(), einLaufenderZug(), false)

  const spalte = element(
    screen.getByRole('region', { name: eigene.zugLeiste }).parentElement,
  )

  for (const knopf of [
    screen.getByRole('button', { name: eigene.pausieren }),
    screen.getByRole('button', { name: eigene.beenden }),
    screen.getByRole('button', { name: eigene.verlaufAnsehen }),
    screen.getByRole('heading', { name: eigene.freigabe }),
  ]) {
    expect(spalte).not.toContainElement(knopf)
    expect(
      spalte.compareDocumentPosition(knopf) & Node.DOCUMENT_POSITION_FOLLOWING,
    ).toBeTruthy()
  }

  // **Die Ereignisse bleiben in der Spalte**: dort steht, was eine Suche in
  // der Bibliothek ergeben hat, und das braucht man im selben Zug.
  expect(spalte).toContainElement(
    screen.getByRole('heading', { name: eigene.ereignisse }),
  )
})

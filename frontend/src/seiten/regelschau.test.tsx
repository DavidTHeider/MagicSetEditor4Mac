// **Die Regelschau vor dem Anfangen** — Werkstück G4 von Phase 5.5
// (#1461).
//
// # Wogegen hier geprüft wird
//
// **Dass jeder Bereich dasteht — auch die stummen.** Eine Liste, die nur die
// Zusagen zeigte, sähe aus wie eine vollständige; der Unterschied zwischen
// „das setzen wir durch" und „das macht ihr selbst" ist die ganze Auskunft.
//
// **Dass die höchste Stufe auch bei einem Ja dasteht.** Wer „begleitet"
// bekommt, soll sehen, dass „streng" nicht zur Wahl steht — sonst erführe er
// es erst beim Versuch.
//
// **Und dass eine Absage als Warnung erscheint** und nicht als stille Zeile.
import { render, screen } from '@testing-library/react'
import { expect, it } from 'vitest'

import { Regelschau } from './Regelschau'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'

import type { RegelfreigabeDTO } from '../apiregeln'
import { eineFreigabe } from '../test/regelprobe'

const eigene = kataloge.de.seiten.spielen

function zeige(freigabe: RegelfreigabeDTO) {
  render(
    <Sprachraum sprache="de">
      <Regelschau freigabe={freigabe} />
    </Sprachraum>,
  )
}

// **Jeder Bereich steht da — der zugesagte mit Quelle und Grenze, der
// stumme ohne.**
it('zeigt jeden Bereich, auch den stummen', () => {
  zeige(eineFreigabe())

  const eintraege = screen.getAllByRole('listitem')

  expect(eintraege.length).toBeGreaterThanOrEqual(2)

  // **Der zugesagte trägt seine Quelle und seine Grenze** — und seinen
  // **Namen** und nicht seinen Schlüssel (PLAY-02.AK-01, #1710).
  expect(screen.getByText(eigene.regelbereichZonen)).toBeInTheDocument()
  expect(screen.getByText(/CR 400/)).toBeInTheDocument()
  expect(
    screen.getByText('kein Auslöser, wenn eine Karte die Zone wechselt'),
  ).toBeInTheDocument()

  // **Der stumme steht auch da — und nennt keine Quelle.**
  //
  // **Geprüft wird die Zeile und nicht der Name**: `getByText` trifft das
  // innere Element, und das trägt nur den Bereich. Die Aussage steht in der
  // Zeile darum.
  const stumm = screen.getByText(eigene.regelbereichCombat).closest('li')

  expect(stumm).toHaveTextContent(eigene.regelgradManuell)
  expect(stumm).not.toHaveTextContent('CR')
})

// **Das Regelwerk steht mit Fassung und Wirksamkeitsdatum da.**
//
// **Alle drei zusammen** — eine Fassungsnummer sagt, welche es ist; das
// Datum sagt, ab wann; und erst beides erklärt, warum zwei Partien
// verschieden entschieden haben.
it('nennt Regelwerk, Fassung und Wirksamkeitsdatum', () => {
  zeige(eineFreigabe())

  const kopf = screen.getByText(new RegExp(eigene.regelwerk))

  expect(kopf).toHaveTextContent('sandkasten')
  expect(kopf).toHaveTextContent('1')
  expect(kopf).toHaveTextContent('2026-08-19')
})

// **Die höchste Stufe steht auch bei einem Ja da.**
it('nennt die höchste Strenge auch ohne Hindernis', () => {
  zeige(eineFreigabe())

  expect(
    screen.getByText(new RegExp(eigene.hoechsteStrenge)),
  ).toHaveTextContent('begleitet')

  expect(screen.queryByRole('alert')).toBeNull()
})

// **Ein Baum, der jeden Bereich führt, lässt nichts übrig.**
//
// **Die Gegenprobe zur Zeile darüber**: der Lauf „zeigt jeden Bereich, auch
// den stummen" lebt davon, dass `combat` im Baum **fehlt** — hier steht er
// drin, und dann darf keine zweite Liste danebenstehen.
it('zeigt keine zweite Liste, wenn der Baum vollständig ist', () => {
  zeige(
    eineFreigabe({
      baum: [{ kennung: 'ablauf', bereiche: ['zonen', 'combat'] }],
    }),
  )

  // **Jeder Bereich steht genau einmal da** — der Baum führt ihn, und die
  // Liste der Übrigen bleibt leer. Zweimal derselbe Name hieße: er steht in
  // beiden.
  expect(screen.getAllByText(eigene.regelbereichCombat)).toHaveLength(1)
  expect(screen.getAllByText(eigene.regelbereichZonen)).toHaveLength(1)
})

// **„Automatisch" heißt hier: mit Grenzen** — PLAY-01.AK-02
// (#1709):
// „die GUI darf ‚automatisch' nicht für bloß teilweise kontrollierte
// Gesamtbereiche behaupten." Jeder automatische Bereich **muss** Grenzen
// nennen (`Matrixeintrag.pruefe`), und das Wort sagt, dass sie darunter
// stehen, statt einen ganzen Bereich zuzusagen.
it('sagt bei einem automatischen Bereich, dass Grenzen folgen', () => {
  zeige(eineFreigabe())

  const zeile = screen.getByText(eigene.regelbereichZonen).closest('li')

  expect(zeile).toHaveTextContent(eigene.regelgradAutomatisch)
  expect(eigene.regelgradAutomatisch).toMatch(/Grenzen/)
  expect(zeile).toHaveTextContent(
    'kein Auslöser, wenn eine Karte die Zone wechselt',
  )
})

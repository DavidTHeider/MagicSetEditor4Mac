// **Die gewählte Karte neben dem Tisch** — das Artboard *Tisch*: unter „Mit
// dieser Karte" die Karte klein, daneben ihr Name und ihr Zustand, darunter
// die Knöpfe (UX-08,
// #1680).
//
// # Wogegen hier geprüft wird
//
// **Dass die Karte klein dasteht, bis jemand sie groß sehen will.** Bis zum
// 2026-09-24 stand sie in voller Betrachtungsgröße über den Handlungen, 312
// Punkte hoch, und schob sie unter die Falz. Klein heißt: die Tischgüte, die
// schon daliegt; groß heißt: die Betrachtungsgüte, und erst dann wird sie
// geholt — „Hohe Auflösung bei Inspektion/Zoom nachladen" (#1459).
//
// **Und dass „Groß ansehen" für die nächste Karte gilt.** Wer Karten liest,
// liest mehr als eine; nach jeder Wahl wieder umzuschalten wäre eine Hürde,
// die das Artboard nicht meint.
import { render, screen } from '@testing-library/react'
import { expect, it, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Partiebefehle } from './Partiebefehle'
import { Sprachraum } from '../sprache/kontext'
import { kataloge } from '../sprache/kataloge'
import { einLauf, einObjekt, einePartie, ruhenderLauf } from './partieprobe'

import type { Spielobjekt } from '../apipartie'
import type { Texturlauf } from '../haken/useSpieltexturen'

const eigene = kataloge.de.seiten.spielen

// befehle ist die Spalte mit einer gewählten Karte.
function befehle(objekt: Spielobjekt, texturen: Texturlauf = ruhenderLauf()) {
  return (
    <Sprachraum sprache="de">
      <Partiebefehle
        lauf={einLauf()}
        partie={einePartie()}
        objekt={objekt}
        entwaehle={() => undefined}
        texturen={texturen}
      />
    </Sprachraum>
  )
}

const wald = einObjekt({
  zone: 'tisch',
  name: 'Wald',
  karteID: '77',
  geheim: false,
  getappt: true,
})

it('zeigt die gewählte Karte klein, mit Name und Zustand', () => {
  const belege = vi.fn()

  render(befehle(wald, { ...ruhenderLauf(), belege }))

  expect(screen.getByText('Wald')).toBeInTheDocument()
  expect(
    screen.getByText(`${eigene.zoneTisch} · ${eigene.getappt}`),
  ).toBeInTheDocument()

  expect(
    screen.getByRole('button', { name: eigene.grossAnsehen }),
  ).toHaveAttribute('aria-pressed', 'false')

  // **Klein ist die Tischgüte**, und die liegt schon da.
  expect(belege).toHaveBeenCalledWith(
    expect.objectContaining({ karteID: '77', guete: 'tisch' }),
  )
  expect(belege).not.toHaveBeenCalledWith(
    expect.objectContaining({ guete: 'betrachten' }),
  )
})

it('lädt groß erst auf Zuruf und bleibt groß für die nächste Karte', async () => {
  const belege = vi.fn()
  const texturen = { ...ruhenderLauf(), belege }

  const { rerender } = render(befehle(wald, texturen))
  const hand = nutzerereignis.setup()

  await hand.click(screen.getByRole('button', { name: eigene.grossAnsehen }))

  expect(
    screen.getByRole('button', { name: eigene.grossAnsehen }),
  ).toHaveAttribute('aria-pressed', 'true')
  expect(belege).toHaveBeenCalledWith(
    expect.objectContaining({ karteID: '77', guete: 'betrachten' }),
  )

  rerender(
    befehle(
      einObjekt({
        kennung: 'sitz1/k2/0',
        zone: 'hand',
        name: 'Berg',
        karteID: '78',
        geheim: false,
      }),
      texturen,
    ),
  )

  expect(
    screen.getByRole('button', { name: eigene.grossAnsehen }),
  ).toHaveAttribute('aria-pressed', 'true')
  expect(belege).toHaveBeenCalledWith(
    expect.objectContaining({ karteID: '78', guete: 'betrachten' }),
  )
})

// **Eine verdeckte Karte trägt auch hier keinen Namen.**
it('nennt eine verdeckte Karte nicht beim Namen', () => {
  render(befehle(einObjekt({ zone: 'hand', name: 'Wald', geheim: true })))

  expect(screen.queryByText('Wald')).not.toBeInTheDocument()
  expect(screen.getByText(eigene.geheim)).toBeInTheDocument()
})

// **Und eine Karte ohne Namen zeigt ihre Kartenkennung.**
it('zeigt die Kartenkennung, wo kein Name steht', () => {
  render(
    befehle(einObjekt({ zone: 'tisch', karte: 'k7', name: '', geheim: false })),
  )

  expect(screen.getByText('k7')).toBeInTheDocument()
})

// **Und der Name kommt aus den Eigenschaften**, wie der Kern ihn schickt —
// am Objekt steht er nur bei einem Spielstein. Hier stand bis zum
// 2026-09-24 „Gewählt: p0-1" für eine Karte aus dem Set „Green".
it('nennt die Karte beim Namen aus ihren Eigenschaften', () => {
  render(
    befehle(
      einObjekt({
        zone: 'hand',
        karte: 'p0-1',
        name: '',
        geheim: false,
        eigenschaften: {
          name: 'A-Canopy Tactician',
          farben: ['G'],
          manawert: 4,
          supertypen: [],
          typen: ['creature'],
          untertypen: ['Elf', 'Warrior'],
          staerke: { steht: true, stern: false, zahl: 3 },
          widerstandskraft: { steht: true, stern: false, zahl: 4 },
          loyalitaet: { steht: false, stern: false, zahl: 0 },
          verteidigung: { steht: false, stern: false, zahl: 0 },
          unbekannteTypen: [],
        },
      }),
    ),
  )

  expect(screen.getByText('A-Canopy Tactician')).toBeInTheDocument()
  expect(screen.queryByText('p0-1')).not.toBeInTheDocument()
})

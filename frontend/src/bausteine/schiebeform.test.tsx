// **Was die Zonenwahl mit der Stellenwahl macht** — Befund **F04**
// (#1748).
//
// # Der Befund
//
// Die ausgelieferte 0.5.0 antwortete auf „Karte schieben" mit *„in hand gibt
// es keine Stelle, an die sich etwas legen ließe"* — **im
// Auslieferungszustand des Formulars**: die Vorgabe der Zone ist `hand`, die
// Vorgabe der Stelle war `oben`, und die Auswahlliste kannte keinen neutralen
// Wert. Wer den Befehl unverändert stellte, bekam einen Fehler; wer ihn
// beheben wollte, fand nichts, womit.
//
// # Warum diese Läufe eine eigene Datei bekommen
//
// **Weil `handlungsleiste.test.tsx` damals bei 391 Zeilen stand** und die
// Zeilengrenze aus §2 keine Näherung kennt. Der Schnitt folgt dabei der Frage
// und nicht der Zeilenzahl: dort, *ob beim Klick das Richtige hinausgeht*,
// hier, *ob ein Feld überhaupt gefragt wird*.
//
// **Und weil die Gestalt eine andere ist**: `spiel/nutzlast.test.ts` hält
// fest, was hinausgeht; hier steht, was ein Mensch sieht. Beides ist nötig —
// eine Nutzlast ohne Stelle nützt niemandem, wenn das Feld weiter danebensteht
// und Erwartungen weckt.
import { fireEvent, render, screen, within } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'
import nutzerereignis from '@testing-library/user-event'

import { Handlungsleiste } from './Handlungsleiste'
import { Sprachraum } from '../sprache/kontext'
import { einObjekt, einePartie } from '../seiten/partieprobe'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.handlungen

// zeige zeichnet die Leiste mit der Schiebeform und einer gewählten Karte.
function zeige() {
  const befehlen = vi.fn()

  render(
    <Sprachraum sprache="de">
      <Handlungsleiste
        partie={einePartie()}
        erlaubt={['objekt.schieben']}
        objekt={einObjekt({
          kennung: 'sitz1/k1/0',
          zone: 'bibliothek',
          name: 'Wald',
          geheim: false,
        })}
        befehlen={befehlen}
        angehalten={false}
      />
    </Sprachraum>,
  )

  // **Aufgeklappt, bevor geprüft wird**: die Felder stehen seit UX-08 erst
  // auf Zuruf da (#1680),
  // und gefragt wird hier, was *in* ihnen steht.
  fireEvent.click(
    screen.getByRole('button', {
      name: eigene.handlungsname['objekt.schieben'],
    }),
  )

  return { hand: nutzerereignis.setup(), befehlen }
}

// schiebezeile gibt die aufgeklappten Felder des Schiebebefehls.
function schiebezeile(): HTMLElement {
  return screen.getByRole('group', {
    name: eigene.handlungsname['objekt.schieben'],
  })
}

// zonenwahl gibt die Auswahlliste der Zielzone.
function zonenwahl(): HTMLElement {
  return within(schiebezeile()).getByLabelText(eigene.handlungsfeld.zone)
}

describe('die Stellenwahl beim Schieben', () => {
  test('bei der Vorgabe „Hand" steht sie gar nicht da', () => {
    zeige()

    expect(
      within(schiebezeile()).queryByLabelText(eigene.handlungsfeld.stelle),
    ).toBeNull()
  })

  test('für eine geordnete Zone kommt sie', async () => {
    const { hand } = zeige()

    await hand.selectOptions(zonenwahl(), 'bibliothek')

    expect(
      within(schiebezeile()).getByLabelText(eigene.handlungsfeld.stelle),
    ).toBeInTheDocument()
  })

  test('und der Befehl trägt sie dann mit', async () => {
    const { hand, befehlen } = zeige()

    await hand.selectOptions(zonenwahl(), 'bibliothek')
    await hand.selectOptions(
      within(schiebezeile()).getByLabelText(eigene.handlungsfeld.stelle),
      'unten',
    )

    await hand.click(
      within(schiebezeile()).getByRole('button', {
        name: eigene.handlungStellen,
      }),
    )

    expect(befehlen).toHaveBeenCalledWith('objekt.schieben', {
      objekt: 'sitz1/k1/0',
      zone: 'bibliothek',
      stelle: 'unten',
    })
  })

  // **S05.NAT-02.** Der Wechsel nimmt die gewählte Stelle aus dem Auftrag —
  // nicht bloß aus dem Bild. Ein Feld, das verschwindet, während sein Wert
  // weiter mitgeschickt wird, ist die schlimmere Hälfte: der Fehler bliebe,
  // und niemand sähe mehr, woher er kommt.
  test('der Wechsel auf eine ungeordnete Zone nimmt sie wieder mit', async () => {
    const { hand, befehlen } = zeige()

    await hand.selectOptions(zonenwahl(), 'bibliothek')
    await hand.selectOptions(
      within(schiebezeile()).getByLabelText(eigene.handlungsfeld.stelle),
      'unten',
    )

    await hand.selectOptions(zonenwahl(), 'hand')

    expect(
      within(schiebezeile()).queryByLabelText(eigene.handlungsfeld.stelle),
    ).toBeNull()

    await hand.click(
      within(schiebezeile()).getByRole('button', {
        name: eigene.handlungStellen,
      }),
    )

    expect(befehlen).toHaveBeenCalledWith('objekt.schieben', {
      objekt: 'sitz1/k1/0',
      zone: 'hand',
    })
  })
})

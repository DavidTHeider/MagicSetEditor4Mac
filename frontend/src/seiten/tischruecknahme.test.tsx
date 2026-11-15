// **⌘Z am Tisch nimmt den letzten eigenen Zug zurück** — P8
// (#2376),
// Vermerk 0197, Artboard `Tisch09`.
//
// # Wogegen
//
// **Gegen den Tisch vor P8**: ⌘Z traf dort die Rücknahme des **Sets** — wer
// sich vertappt hatte und ⌘Z drückte, nahm die letzte Änderung an einer
// Karte zurück, die er gar nicht sah. Jetzt meint ⌘Z am Tisch den Tisch, und
// ob der Zug zurückgeht, entscheidet der Kern.
import { render, screen } from '@testing-library/react'
import nutzerereignis from '@testing-library/user-event'
import { expect, it, vi } from 'vitest'

import { Partietisch } from './Partietisch'
import { Ruecknahmeraum, useRuecknahmeziel } from '../ruecknahme'
import { ruecknahmeweg } from '../ruecknahmeweg'
import { Sprachraum } from '../sprache/kontext'
import { einLauf, einePartie, einZug } from './partieprobe'
import { kataloge } from '../sprache/kataloge'

import type { Partielauf } from '../haken/usePartie'

const eigene = kataloge.de.seiten.spielen

// Menue steht für den Menübefehl „Widerrufen“ der Anwendung.
function Menue({
  amBestand,
  vorBestand,
}: {
  amBestand: () => void
  vorBestand: () => void
}) {
  const weg = ruecknahmeweg(useRuecknahmeziel(), {
    zuruecknehmen: amBestand,
    wiederherstellen: vorBestand,
  })

  return (
    <>
      <button type="button" onClick={weg.zuruecknehmen}>
        widerrufen
      </button>
      <button type="button" onClick={weg.wiederherstellen}>
        wiederholen
      </button>
    </>
  )
}

function zeige(lauf: Partielauf) {
  const amBestand = vi.fn()
  const vorBestand = vi.fn()
  render(
    <Sprachraum sprache="de">
      <Ruecknahmeraum>
        <Partietisch lauf={lauf} partie={lauf.partie ?? einePartie()} />
        <Menue amBestand={amBestand} vorBestand={vorBestand} />
      </Ruecknahmeraum>
    </Sprachraum>,
  )

  return { amBestand, vorBestand, hand: nutzerereignis.setup() }
}

it('nimmt mit ⌘Z am Tisch den letzten Zug zurück, nicht das Set', async () => {
  const befehlen = vi.fn()
  const { amBestand, hand } = zeige(einLauf({ befehlen }))

  await hand.click(screen.getByRole('button', { name: 'widerrufen' }))

  expect(befehlen).toHaveBeenCalledExactlyOnceWith('spieler.zuruecknehmen', {})
  expect(amBestand).not.toHaveBeenCalled()
})

// **Ohne Rücknahme im Angebot tut ⌘Z am Tisch nichts** — und greift auch
// nicht auf das Set durch.
it('tut ohne Angebot nichts', async () => {
  const befehlen = vi.fn()
  const { amBestand, hand } = zeige(
    einLauf({ befehlen, erlaubt: ['objekt.tappen'] }),
  )

  await hand.click(screen.getByRole('button', { name: 'widerrufen' }))

  expect(befehlen).not.toHaveBeenCalled()
  expect(amBestand).not.toHaveBeenCalled()
  expect(
    screen.queryByRole('button', { name: eigene.zuruecknehmenKnopf }),
  ).toBeNull()
})

it('nimmt über den Knopf neben dem Tisch zurück und nennt ⌘Z', async () => {
  const befehlen = vi.fn()
  const { hand } = zeige(einLauf({ befehlen }))

  const knopf = screen.getByRole('button', { name: eigene.zuruecknehmenKnopf })
  expect(knopf).toHaveAttribute('aria-keyshortcuts', 'Meta+Z')

  await hand.click(knopf)

  expect(befehlen).toHaveBeenCalledExactlyOnceWith('spieler.zuruecknehmen', {})
})

// **Der Zug steht oben in der Spalte, der Knopf darunter** — G9 (#1592) und
// die Entscheidung zu #2526: über der Zugleiste schob der Knopf den Zug im
// schmalen Fenster unter die Falz.
it('steht unter der Zugleiste', () => {
  // **Mit laufendem Zug** — ohne zeichnet die Leiste nichts.
  zeige(
    einLauf({
      partie: einePartie({
        zug: einZug({ nummer: 3, aktiv: 'sitz1', prioritaet: 'sitz1' }),
      }),
    }),
  )

  const zug = screen.getByRole('region', { name: eigene.zugLeiste })
  const knopf = screen.getByRole('button', { name: eigene.zuruecknehmenKnopf })

  expect(
    zug.compareDocumentPosition(knopf) & Node.DOCUMENT_POSITION_FOLLOWING,
  ).toBeTruthy()
})

// **Wiederherstellen gibt es am Tisch nicht** (Vermerk 0197) — ⇧⌘Z tut dort
// nichts und greift nicht auf das Set durch.
it('stellt mit ⇧⌘Z nichts wieder her', async () => {
  const befehlen = vi.fn()
  const { vorBestand, hand } = zeige(einLauf({ befehlen }))

  await hand.click(screen.getByRole('button', { name: 'wiederholen' }))

  expect(befehlen).not.toHaveBeenCalled()
  expect(vorBestand).not.toHaveBeenCalled()
})

// **Auch beim Sichtwechsel nicht** — Zweitblick über Phase 7, Befund 9. Die
// Spalte geht mit der Sicht (und hinter den Vorhang); hing die Anmeldung an
// ihr, fiel ⌘Z in dieser Zeit auf das Set zurück und nahm unsichtbar eine
// Kartenänderung zurück. Der Tisch bleibt Ziel, solange er dasteht — und tut
// dann nichts.
it('greift beim Sichtwechsel nicht auf das Set durch', async () => {
  const befehlen = vi.fn()
  const { amBestand, hand } = zeige(einLauf({ befehlen, wechselt: true }))

  await hand.click(screen.getByRole('button', { name: 'widerrufen' }))

  expect(befehlen).not.toHaveBeenCalled()
  expect(amBestand).not.toHaveBeenCalled()
})

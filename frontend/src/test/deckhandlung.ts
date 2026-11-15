// **Ein Eintrag aus dem Menü eines Decks** — AP-6
// (#2363):
// Umbenennen, Duplizieren, Teilen und Löschen stehen seitdem hinter dem
// Knopf „…“ am offenen Deck. Der Weg dorthin steht hier einmal und nicht in
// jedem Test, der ein Deck umbenennt.
import { screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { einsetzen } from '../sprache/einsetzen'
import { kataloge } from '../sprache/kataloge'

const eigene = kataloge.de.seiten.decks

// deckmenue ist der Knopf „…“ am offenen Deck dieses Namens.
export async function deckmenue(deck: string): Promise<HTMLElement> {
  return screen.findByRole('button', {
    name: einsetzen(eigene.handlungenAn, { name: deck }),
  })
}

// deckhandlung wählt einen Eintrag aus diesem Menü.
export async function deckhandlung(
  deck: string,
  eintrag: string,
): Promise<void> {
  await userEvent.click(await deckmenue(deck))
  await userEvent.click(
    screen.getByRole('menuitem', { name: new RegExp(`^${eintrag}`) }),
  )
}

// platzeintrag öffnet das Menü „…“ eines Platzes und gibt seinen Eintrag —
// AP-6: Menge ±, Bereich und Herausnehmen stehen dort und nicht mehr als
// Knöpfe in der Zeile. **Der Zeiger fährt zuerst über die Zeile**: das Menü
// steht nur an der aktiven und an der unter dem Zeiger.
export async function platzeintrag(
  platz: string,
  eintrag: string,
): Promise<HTMLElement> {
  const liste = await screen.findByRole('list', { name: eigene.plaetze })
  const zeile = within(liste)
    .getAllByRole('listitem')
    .find((kandidat) => within(kandidat).queryByText(platz) !== null)
  if (zeile === undefined) {
    throw new Error(`kein Platz »${platz}« in der Liste`)
  }

  await userEvent.hover(zeile)
  await userEvent.click(
    within(zeile).getByRole('button', {
      name: `${eigene.plaetze}: ${platz}`,
    }),
  )

  return screen.getByRole('menuitem', { name: eintrag })
}

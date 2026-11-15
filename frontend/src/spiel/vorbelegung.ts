// **Die Sitze stehen schon da** — P6
// (#2374),
// Befund C-04, Artboard `PartieStart09`.
//
// # Woher jeder Sitz kommt
//
// **Der erste** aus dem Profil, in dem gearbeitet wird, und aus dem Deck, aus
// dem man kommt — mitgebracht („Mit diesem Deck spielen“) oder, ohne eines,
// das zuletzt geöffnete. **Der zweite** aus dem letzten Gegenüber am Tisch und
// dem Deck, das davor geöffnet war; ohne eines spiegelt er den ersten.
//
// # Was sie nie tut
//
// **Ein fremdes Deck wählen.** Gewählt wird nur, was in der eigenen Liste
// steht — ein mitgebrachtes oder zuletzt geöffnetes Deck, das es dort nicht
// (mehr) gibt, wird übergangen. Eine Leihe setzt sie nie: die nimmt man an.
// **Und keinen Sitz an ein Profil binden** — leer bleibt der Normalfall (G8).
import type { Sitzwunsch } from '../seiten/Sitzliste'

export interface Vorbelegungsquelle {
  decks: readonly { id: string }[]
  mitgebracht: string | undefined

  // zuletzt sind die zuletzt geöffneten Decks, das jüngste zuerst.
  zuletzt: readonly string[]
  profilname: string

  // gegenueber ist der Name am zweiten Sitz der letzten Partie.
  gegenueber: string | undefined
  ersatzname: string

  // eingeladen setzt den zweiten Sitz auf einen Gast, der sein Deck
  // mitbringt — „Bo einladen“ (P9c, #2377).
  eingeladen?: { kennung: string; name: string } | undefined
}

export function vorbelegung(quelle: Vorbelegungsquelle): Sitzwunsch[] {
  const eigen = (id: string | undefined): id is string =>
    id !== undefined && quelle.decks.some((deck) => deck.id === id)

  const zuletzt = quelle.zuletzt.filter(eigen)
  const erstes =
    [quelle.mitgebracht, ...zuletzt].find(eigen) ?? quelle.decks[0]?.id ?? ''
  const zweites = zuletzt.find((id) => id !== erstes) ?? erstes

  const erster = {
    name: quelle.profilname,
    deckID: erstes,
    profilID: '',
    leiheID: '',
  }

  if (quelle.eingeladen !== undefined) {
    return [erster, eingeladenerSitz(quelle.eingeladen)]
  }

  return [
    erster,
    {
      name: quelle.gegenueber ?? quelle.ersatzname,
      deckID: zweites,
      profilID: '',
      leiheID: '',
    },
  ]
}

// eingeladenerSitz ist der Sitz eines Gastes, der sein Deck mitbringt.
export function eingeladenerSitz(gast: {
  kennung: string
  name: string
}): Sitzwunsch {
  return {
    name: gast.name,
    deckID: '',
    profilID: '',
    leiheID: '',
    eingeladen: gast.kennung,
  }
}

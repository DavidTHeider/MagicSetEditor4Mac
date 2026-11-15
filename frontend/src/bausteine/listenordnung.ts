// **Wie Sets und Decks sich ordnen lassen** — UX-11
// (#1683).
//
// **Die erste Ordnung ist die, die schon dastand**: bei den Sets die des
// Bestands, bei den Decks die zuletzt benutzten zuerst. Ein Wechsel der
// Ansicht ändert daran nichts; erst eine Wahl hier ordnet um.
import type { SetDTO } from '../api'
import type { DeckDTO } from '../apidecks'
import type { Listenform, Sortierung } from './Listenansicht'
import type { Katalog } from '../sprache/kataloge'
import type { Erlaubt } from '../haken/useAnsicht'

// Die Schlüssel stehen in `internal/domain/ansicht.go`.
export const setansichtsschluessel = 'arbeitsplatz.ansicht.sets'
export const deckansichtsschluessel = 'arbeitsplatz.ansicht.decks'

// Die erste Form ist die Vorgabe — dieselbe wie in der Domäne.
export const setformen: Erlaubt<Listenform> = ['kacheln', 'liste']
export const deckformen: Erlaubt<Listenform> = ['liste', 'kacheln']

// **Menschentext, in der Gegend des Fensters** (`docs/LOKALISIERUNG.md`).
function nachName(sprache: string) {
  return (eines: { name: string }, anderes: { name: string }) =>
    eines.name.localeCompare(anderes.name, sprache)
}

export function setsortierungen(
  texte: Katalog,
  sprache: string,
): Sortierung<SetDTO>[] {
  const worte = texte.bausteine.listenansicht

  return [
    { name: 'bestand', bezeichnung: worte.wieImBestand, vergleich: () => 0 },
    { name: 'name', bezeichnung: worte.nachName, vergleich: nachName(sprache) },
    {
      name: 'karten',
      bezeichnung: worte.nachKarten,
      vergleich: (eines, anderes) => anderes.karten - eines.karten,
    },
  ]
}

export function decksortierungen(
  texte: Katalog,
  sprache: string,
  rang: (deckID: string) => number,
): Sortierung<DeckDTO>[] {
  const worte = texte.bausteine.listenansicht

  return [
    {
      name: 'zuletzt',
      bezeichnung: worte.zuletztBenutzt,
      vergleich: (eines, anderes) => rang(eines.id) - rang(anderes.id),
    },
    { name: 'name', bezeichnung: worte.nachName, vergleich: nachName(sprache) },
    {
      name: 'karten',
      bezeichnung: worte.nachKarten,
      vergleich: (eines, anderes) => anderes.hauptdeck - eines.hauptdeck,
    },
  ]
}

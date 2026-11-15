// **Eine Handlungsliste an der Menüleiste, als Bauteil** — Zweitblick über
// Phase 6 von 0.9.0, zweiter Teil, Befund 6. Ein Haken lässt sich nicht
// bedingt rufen; die Deckzeile steht je Deck da, und nur das offene meldet
// seine Liste an. Gezeichnet wird nichts.
import {
  useHandlungsliste,
  type Handlungslistenangaben,
} from '../haken/useHandlungsliste'

export function Handlungsanmeldung(angaben: Handlungslistenangaben) {
  useHandlungsliste(angaben)

  return null
}

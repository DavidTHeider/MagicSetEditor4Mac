// **Was am offenen Deck zu tun ist — eine Ebene über dem Navigator** —
// Rest R6 aus dem Nachlauf über Phase 6 von 0.9.0
// (#2469).
// Die Handlungen des offenen Decks hingen samt Zustand an seiner Zeile im
// Navigator: traf die Suche das Deck nicht oder war der Navigator
// eingeklappt, fehlte das Menü „Deck“ in der Menüleiste, und ein offenes
// Popover oder die Schublade der Einfuhr ging mit zu.
//
// # Eine Liste und ein Zustand, zwei Wege
//
// **Hier stehen beide, und die Zeile zeigt sie nur.** Überlauf, Rechtsklick
// und ⇧F10 an der Zeile und die Menüleiste nehmen dieselbe Liste (U2, #1413).
// **Die Menüleiste klappt den Navigator vorher auf** (`zeigen`):
// Umbenennen, Kopieren, Teilen und die Rückfrage beim Löschen stehen an der
// Zeile (UX-08.AK-01, AK-03), und die Zeile bleibt in der gefilterten Liste
// stehen (`Listenansicht`, `festgehalten`). Die Schublade der Einfuhr steht
// über der Seite und braucht ihn nicht. Entschieden in
// `docs/PROTOTYP-0.9.0.md` (Weg 2 aus dem Issue).
//
// # Warum der Zustand hier liegt
//
// **Weil eine Zeile, die erst aufgeht, ihn sonst nicht kennte.** Wählt jemand
// in der Menüleiste „Umbenennen“, während der Navigator zu ist, entsteht die
// Zeile im selben Bild wie der Wunsch — und zeigt das Popover, weil der Wunsch
// hier steht und nicht in ihr.
import { useState } from 'react'

import { useAnschluss } from '../anschlussraum'
import type { DeckDTO } from '../apidecks'
import type { Menueeintrag } from '../bausteine/Menue'
import { useTexte } from '../sprache/kontext'
import { useDeckkopie, type Deckkopierlauf } from './useDeckkopie'
import type { Deckstand } from './useDecks'

export interface Deckumbenennung {
  id: string
  name: string
}

export interface Deckhandlung {
  umbenannt: Deckumbenennung | undefined
  setzeUmbenannt: (umbenannt: Deckumbenennung | undefined) => void

  // gefragt ist das Deck, dessen Löschen gerade nachfragt.
  gefragt: string
  fragen: (deckID: string) => void

  // teilanlass zählt jede Wahl von „Teilen“ (`Teilen`, `anlass`).
  teilanlass: number

  einfuhrOffen: boolean
  schliesseEinfuhr: () => void

  kopie: Deckkopierlauf

  // eintraege ist die Liste am offenen Deck. **`zeigen` gibt die Menüleiste
  // mit**: es macht die Zeile sichtbar und sagt, ob sie es vorher nicht war.
  eintraege: (deck: DeckDTO, zeigen?: () => boolean) => Menueeintrag[]
}

export function useDeckhandlung(stand: Deckstand): Deckhandlung {
  const texte = useTexte()
  const eigene = texte.seiten.decks
  const anschluss = useAnschluss()

  const [umbenannt, setzeUmbenannt] = useState<Deckumbenennung>()
  const [gefragt, fragen] = useState('')
  const [teilanlass, setzeTeilanlass] = useState(0)
  const [einfuhrOffen, setzeEinfuhrOffen] = useState(false)
  const kopie = useDeckkopie(stand.offen, stand)

  // **Ein Wechsel des Decks räumt alles** (§7, Regel 1) — beim Zeichnen und
  // nicht in einem Effekt: der geräumte Stand steht im selben Durchgang da,
  // in dem das Deck wechselt.
  const offenID = stand.offen?.id
  const [offenBei, setzeOffenBei] = useState(offenID)
  if (offenBei !== offenID) {
    setzeOffenBei(offenID)
    setzeUmbenannt(undefined)
    fragen('')
    setzeEinfuhrOffen(false)
    kopie.abbrechen()
  }

  const eintraege = (deck: DeckDTO, zeigen = () => false): Menueeintrag[] => {
    const laeuft = stand.vorgang.laeuft(`deck:${deck.id}`)

    // **An der Zeile, also mit sichtbarer Zeile** — `zeigen` ist ohne die
    // Menüleiste ein Nichts.
    const anDerZeile = (tun: () => void) => () => {
      zeigen()
      tun()
    }

    return [
      // **Schnell spielen** — P6 (#2374), Artboard `PartieStart09`: mit der
      // Vorbelegung gleich an den Tisch, derselbe Anfang wie aus der Palette.
      {
        schluessel: 'schnellspielen',
        name: texte.seiten.spielen.schnellSpielen,
        aufWahl: () => {
          anschluss.uebergib({
            art: 'partie',
            name: deck.name,
            deckID: deck.id,
            sofort: true,
          })
        },
      },
      {
        schluessel: 'umbenennen',
        name: eigene.umbenennen,
        aktion: 'bearbeiten',
        schreibt: true,
        angehalten: laeuft,
        aufWahl: anDerZeile(() => {
          kopie.abbrechen()
          setzeUmbenannt({ id: deck.id, name: deck.name })
        }),
      },
      {
        schluessel: 'duplizieren',
        name: eigene.duplizieren,
        aktion: 'duplizieren',
        schreibt: true,
        angehalten: laeuft,
        aufWahl: anDerZeile(() => {
          setzeUmbenannt(undefined)
          kopie.beginnen()
        }),
      },
      // **Das Deck als Liste teilen** — NET-04 (#1708). **Geht die Zeile
      // erst auf, zählt der Anlass einen Takt später**: `Teilen` nimmt den
      // Anlass, mit dem es entsteht, als schon gesehen.
      {
        schluessel: 'teilen',
        name: texte.bausteine.teilen.knopf,
        aufWahl: () => {
          const zaehle = () => {
            setzeTeilanlass((bisher) => bisher + 1)
          }

          if (zeigen()) {
            window.setTimeout(zaehle, 0)
          } else {
            zaehle()
          }
        },
      },
      // **Eine Deckliste einlesen** — Werkstück T8 (#1473); seit AP-6
      // (#2363) in einer Schublade: es füllt dieses Deck, und es ist ein
      // seltener Weg. **Die Schublade steht über der Seite**, nicht an der
      // Zeile — sie braucht keinen Navigator.
      {
        schluessel: 'einlesen',
        name: eigene.einfuhrMenue,
        schreibt: true,
        angehalten: laeuft,
        aufWahl: () => {
          setzeEinfuhrOffen(true)
        },
      },
      // **Löschen fragt unter der Zeile** — UX-08.AK-03 (#1680).
      {
        schluessel: 'loeschen',
        name: eigene.loeschen,
        aktion: 'loeschen',
        schreibt: true,
        angehalten: gefragt === deck.id,
        aufWahl: anDerZeile(() => {
          fragen(deck.id)
        }),
      },
    ]
  }

  return {
    umbenannt,
    setzeUmbenannt,
    gefragt,
    fragen,
    teilanlass,
    einfuhrOffen,
    schliesseEinfuhr: () => {
      setzeEinfuhrOffen(false)
    },
    kopie,
    eintraege,
  }
}

// **Die Plätze eines Decks — mit Menge, Bereichswechsel und Kontextmenü** —
// Werkstück U10 von Phase 5.1
// (#1421).
//
// **Eine eigene Datei**, damit `Deckwerkbank.tsx` unter der Zeilengrenze aus
// §2 bleibt. Der Schnitt folgt der Frage: dort steht, **welches Deck** offen
// ist und was daran hängt, hier, **was darin liegt**.
//
// # Jede Geste hat einen Tastenweg
//
// Das ist die Abnahme aus dem Issue, und sie ist der Grund, warum hier keine
// Zieh-Geste steht, die man nur mit der Maus hat: Menge, Bereich und
// Herausnehmen sind Einträge eines Menüs, das ein Knopf öffnet, und das
// Kontextmenü öffnet sich auch über die Kontextmenütaste.
//
// # Die Knöpfe stehen an einer Zeile, nicht an jeder
//
// **UXR-11** (#1919).
// Ein Commanderdeck hat 76 Plätze, und jeder trug vier Knöpfe: 304
// Bedienelemente in der Mitte, und wer mit der Tastatur ans Ende wollte,
// drückte dreihundertmal Tabulator (R2 und R6 in `docs/UIREVIEW.md`).
//
// Jetzt trägt sie die Zeile mit der **Tabstelle** — es ist eine, und die
// Pfeile wandern — und die unter dem **Zeiger**. Erreichbar bleibt jede
// Handlung an jedem Platz: mit der Maus über die Zeile, mit der Tastatur über
// Pfeile und Tabulator, und über das Kontextmenü wie bisher. Dasselbe Muster
// wie an der Paketliste (UXR-17, #1925).
//
// # Und es ist ein Knopf, nicht vier
//
// **AP-6** (#2363,
// `Bereiche09`): −, +, Bereich und Herausnehmen standen als vier Knöpfe an
// der aktiven Zeile — drei Bedienelemente mehr im Bild eines offenen Decks,
// das höchstens zwanzig tragen soll. Sie stehen jetzt hinter „…“, in
// derselben Liste wie der Rechtsklick.
import { useState, type KeyboardEvent } from 'react'

import { Menue } from './Menue'
import { Kontextmenue } from './Kontextmenue'
import { Vorlaeufig } from './Vorlaeufig'
import { useRovierendeZeile } from '../haken/useRovierendeZeile'
import { zielzeile } from '../seiten/gitterzeilentaste'
import { useTexte } from '../sprache/kontext'

import type { DeckDTO, DeckplatzDTO } from '../apidecks'
import type { Menueeintrag } from './Menueliste'

export function Deckplaetze({
  deck,
  angehalten,
  unzulaessig,
  aufMenge,
  aufBereich,
  aufNehmen,
}: {
  deck: DeckDTO
  angehalten: boolean

  // unzulaessig sagt zu einem Platz, ob sein Posten abgewiesen ist — der
  // Befund **am Posten**, den das Issue verlangt.
  unzulaessig: (platzID: string) => boolean
  aufMenge: (platz: DeckplatzDTO, menge: number) => void
  aufBereich: (platz: DeckplatzDTO, bereich: string) => void
  aufNehmen: (platzID: string) => void
}) {
  const eigene = useTexte().seiten.decks
  const rovierend = useRovierendeZeile(deck.plaetze.length)
  const [zeiger, setzeZeiger] = useState<number>()

  if (deck.plaetze.length === 0) {
    return <p>{eigene.plaetzeLeer}</p>
  }

  // wandern führt die Tabstelle mit den Pfeilen, Pos1 und Ende weiter.
  //
  // **Nur ohne Umschalt- und Befehlstaste** — ⇧F10 gehört dem Kontextmenü,
  // und ⌘↓ dem Fenster.
  const wandern = (ereignis: KeyboardEvent<HTMLElement>, stelle: number) => {
    const ziel = zielzeile(ereignis.key, stelle, deck.plaetze.length)

    if (
      ziel === undefined ||
      ereignis.shiftKey ||
      ereignis.metaKey ||
      ereignis.altKey ||
      ereignis.ctrlKey
    ) {
      return
    }

    ereignis.preventDefault()
    rovierend.rueckenAuf(ziel)

    // **Die Geschwister und nicht eine Suche im Dokument** — es gibt nur
    // diese Liste, und ihre Kinder sind genau die Zeilen.
    const liste = ereignis.currentTarget.parentNode as Element
    ;(liste.children[ziel] as HTMLElement).focus()
  }

  return (
    <ul
      aria-label={eigene.plaetze}
      className="grid grid-cols-[auto_1fr_auto_auto] items-center gap-x-2 gap-y-1"
    >
      {deck.plaetze.map((platz, stelle) => {
        const tabstelle = rovierend.annehmen(stelle)
        const bedienbar = tabstelle || zeiger === stelle
        const eintraege = handlungen(
          platz,
          eigene,
          angehalten,
          aufMenge,
          aufBereich,
          aufNehmen,
        )

        return (
          /* **Der Kasten ist das `li` selbst** — Befund 3 des Nachlaufs zum
             Zweitblick über Phase 5.1. Zuerst stand er darum herum, und weil
             er ein `div` erzeugt, hieß der Baum `ul > div > li`: für ein
             Vorleseprogramm keine Liste mehr (axe, Regel `list`), und das
             `li` war kein direktes Kind des Gitters, womit
             `grid-cols-subgrid` ins Leere ging. Seit UXR-11 ist es auch die
             **Fokusstelle** der Zeile. */
          <Kontextmenue
            key={platz.id}
            element="li"
            klasse="col-span-4 grid min-h-8 grid-cols-subgrid items-center rounded-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            bezeichnung={`${eigene.plaetze}: ${beschriftung(platz)}`}
            eintraege={eintraege}
            zeile={{
              tabIndex: tabstelle ? 0 : -1,
              onFocus: () => {
                rovierend.rueckenAuf(stelle)
              },
              onKeyDown: (ereignis) => {
                wandern(ereignis, stelle)
              },
              onPointerEnter: () => {
                setzeZeiger(stelle)
              },
              onPointerLeave: () => {
                setzeZeiger(undefined)
              },
            }}
          >
            <span className="w-6 text-right tabular-nums">{platz.menge}</span>

            {/* **Der Name und nicht die Kennung** (Werkstück I). Solange die
                Karte noch lädt — oder es sie nicht mehr gibt — steht die
                Kennung: sie ist die einzige Angabe, die dann stimmt. */}
            <span className="flex min-w-0 items-center gap-2">
              <span className="truncate">{beschriftung(platz)}</span>
              <Vorlaeufig vorlaeufig={platz.vorlaeufig} />

              {/* **Der Befund steht am Posten und im Kopf** — die Abnahme
                  aus dem Issue. `ungeprüft` steht dem nicht entgegen
                  (Vermerk 0061): eine eigene Karte ist in keiner Liste, und
                  sie rot zu färben hieße, jedem eigenen Deck zu sagen, es
                  sei kaputt. */}
              {unzulaessig(platz.id) && (
                <span
                  className="rounded-sm border border-current px-1 text-xs text-fehler"
                  title={eigene.postenUnzulaessigErklaerung}
                >
                  {eigene.postenUnzulaessig}
                </span>
              )}
            </span>

            {/* **Ein Wort fürs Sideboard**: „Hauptdeck" an siebzig Zeilen
                sagte nichts. */}
            <span className="text-sm text-leise">
              {platz.bereich === 'sideboard' && eigene.sideboard}
            </span>

            {/* **Menge ±, Bereich und Herausnehmen hinter „…“** — AP-6
                (#2363), `Bereiche09`: vier Knöpfe an der aktiven Zeile waren
                drei Bedienelemente mehr im Bild. Dieselbe Liste wie der
                Rechtsklick; eine Mengenänderung nimmt weiter nichts heraus
                (DOM-03.AK-03), das tut der eigene Eintrag. */}
            {bedienbar ? (
              <Menue
                bezeichnung={`${eigene.plaetze}: ${beschriftung(platz)}`}
                gewicht="leise"
                ausrichtung="rechts"
                eintraege={eintraege}
              >
                {eigene.handlungenKurz}
              </Menue>
            ) : (
              <span />
            )}
          </Kontextmenue>
        )
      })}
    </ul>
  )
}

// beschriftung nennt einen Platz beim Namen — oder bei seiner Kennung.
//
// **Der Name kommt aus der Antwort** und nicht aus einem zweiten Ruf
// (#1376):
// der Dienst holt ihn beim Lesen, für eigene Karten aus dem Bestand und für
// Drucke von Scryfall. Ein Fenster, das ihn selbst holte, ließe MCP und die
// Kommandozeile mit UUIDs zurück (§16).
//
// **Leer heißt: nicht zu haben** — ohne Netz bleibt der Name eines Drucks
// aus, und dann steht die Kennung da; ein Druck ohne Alias hat nur seine
// Zeile (A2, #1441).
export function beschriftung(platz: DeckplatzDTO): string {
  if (platz.bezeichnung !== '') {
    return platz.bezeichnung
  }

  if (platz.printing !== '') {
    return platz.printing
  }

  return platz.druckID !== '' ? platz.druckID : platz.karteID
}

// handlungen sind die vier Dinge, die man mit einem Platz tun kann.
//
// **Eine Liste für beide Wege** — das ist die Regel aus [Kontextmenue]: keine
// Aktion, die es nur im Kontextmenü gibt. Was hier steht, steht auch im Menü
// „…“ der Zeile (AP-6, #2363); der Rechtsklick ist die Abkürzung für die, die
// ihn kennen.
function handlungen(
  platz: DeckplatzDTO,
  eigene: Deckwortliste,
  angehalten: boolean,
  aufMenge: (platz: DeckplatzDTO, menge: number) => void,
  aufBereich: (platz: DeckplatzDTO, bereich: string) => void,
  aufNehmen: (platzID: string) => void,
): Menueeintrag[] {
  const insSideboard = platz.bereich === 'sideboard'

  return [
    {
      schluessel: 'mehr',
      schreibt: true,
      name: eigene.mehrVon,
      angehalten,
      aufWahl: () => {
        aufMenge(platz, platz.menge + 1)
      },
    },
    {
      schluessel: 'weniger',
      schreibt: true,
      name: eigene.wenigerVon,
      // **Bei eins ruht es** — wie der Knopf in der Zeile (DOM-03.AK-03).
      angehalten: angehalten || platz.menge <= 1,
      aufWahl: () => {
        aufMenge(platz, platz.menge - 1)
      },
    },
    {
      schluessel: 'bereich',
      schreibt: true,
      name: insSideboard ? eigene.insHauptdeck : eigene.insSideboard,
      angehalten,
      aufWahl: () => {
        aufBereich(platz, insSideboard ? 'hauptdeck' : 'sideboard')
      },
    },
    {
      schluessel: 'nehmen',
      schreibt: true,
      name: eigene.nehmen,
      angehalten,
      aufWahl: () => {
        aufNehmen(platz.id)
      },
    },
  ]
}

// Deckwortliste ist der Ausschnitt des Katalogs, den die Handlungen brauchen.
type Deckwortliste = ReturnType<typeof useTexte>['seiten']['decks']

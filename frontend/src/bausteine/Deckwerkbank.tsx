// **Das offene Deck — der mittlere Bereich des Deckbauplatzes** — Werkstück
// U10 von Phase 5.1
// (#1421).
//
// # Was hier steht und was nicht
//
// Hier steht, **was man am Deck tut**: Name, Herkunft, Format, die Plätze mit
// +/− und Bereichswechsel. Die **Suche** steht im Navigator (dort wird eine
// Karte gefunden), die **Analyse** im Inspector (dort wird nichts geändert).
// Bis U10 war das alles eine Kachel unter der Deckliste; die Seite war
// 1.318 Punkte hoch, und wer die Auswertung sehen wollte, scrollte an allem
// vorbei.
import { useId, useRef, type ReactNode } from 'react'

import { Deckleiste } from './Deckleiste'
import { Deckplaetze } from './Deckplaetze'
import { Herkunftszeile } from './Herkunftszeile'
import { useTexte } from '../sprache/kontext'

import type { Deckaktionen } from '../haken/useDeckaktionen'
import type { Deckpruefung } from '../haken/useDeckpruefung'
import type { Deckstand } from '../haken/useDecks'
import type { DeckDTO, HabenstandAntwort } from '../apidecks'
import type { Ladezustand } from '../haken/useLaden'

export function Deckwerkbank({
  stand,
  deck,
  aktionen,
  pruefung,
  abgleich,
  zeigePruefung,
  children,
}: {
  stand: Deckstand
  deck: DeckDTO
  aktionen: Deckaktionen
  pruefung: Deckpruefung
  abgleich: Ladezustand<HabenstandAntwort | null>

  // zeigePruefung schlägt die Lasche „Prüfung“ im Inspector auf — das Ziel
  // der Leiste für Urteil und Fehlendes (AP-9).
  zeigePruefung: () => void

  // children ist, was Karten ins Deck bringt — die Kartenwahl und das
  // Einlesen einer Liste. **Zwischen Format und Plätzen** (UXR-11, #1919):
  // über dem, was sie füllen, wie im Artboard *Decks*.
  children?: ReactNode
}) {
  const eigene = useTexte().seiten.decks
  const formathinweis = useId()
  const plaetze = useRef<HTMLHeadingElement>(null)

  return (
    <section aria-label={deck.name} className="flex flex-col gap-4">
      <h2 className="text-lg font-medium">{deck.name}</h2>

      {/* **Woher es kam, gleich unter dem Namen** — Werkstück 12 (#1351).
          Bei einem hier entstandenen Deck steht dort nichts; der Grund
          steht in `Herkunftszeile`. */}
      <Herkunftszeile herkunft={deck.herkunft} />

      {/* **Die Leiste über dem Deck** — AP-9 (#2366), `Bereiche09`. */}
      <Deckleiste
        deck={deck}
        pruefung={pruefung}
        abgleich={abgleich}
        zeigePlaetze={() => {
          plaetze.current?.scrollIntoView({ block: 'start' })
          plaetze.current?.focus()
        }}
        zeigePruefung={zeigePruefung}
      />

      <div className="flex max-w-md flex-col gap-1">
        <label className="flex flex-col gap-1">
          <span className="text-sm">{eigene.formatWaehlen}</span>

          {/* **Der Hinweis hängt über `aria-describedby` daran und steht
              nicht *in* der Aufschrift.** Sonst hieße das Feld „Format wählen
              Ein Deck ohne Format ist ungeprüft, nicht falsch …" — und
              genauso läse eine Vorlesehilfe es vor. */}
          {/* **Angehalten, solange ein Vorgang an diesem Deck läuft** (§7,
              Regel 4). Ein Formatwechsel während des Legens ginge gegen einen
              Stand, den der Bestand noch nicht kennt — angestrichen von
              Befund 5 des Zweitblicks über Phase 4.6. */}
          <select
            aria-describedby={formathinweis}
            disabled={aktionen.laeuft}
            value={deck.formatID}
            onChange={(ereignis) => {
              void aktionen.formatWaehlen(ereignis.target.value)
            }}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            <option value="">{eigene.ohneFormat}</option>

            {stand.formate.map((format) => (
              <option key={format.id} value={format.id}>
                {format.name}
              </option>
            ))}
          </select>
        </label>

        <span id={formathinweis} className="text-sm text-leise">
          {eigene.formatHinweis}
        </span>
      </div>

      {children}

      {/* **Ein Ziel der Leiste** — der Fokus landet hier, damit die Tastatur
          dort weitergeht, wohin der Klick geführt hat. */}
      <h3 ref={plaetze} tabIndex={-1} className="font-medium">
        {eigene.plaetze}
      </h3>

      <Deckplaetze
        deck={deck}
        angehalten={aktionen.laeuft}
        unzulaessig={pruefung.unzulaessig}
        aufMenge={(platz, menge) => {
          void aktionen.legen(
            platz.karteID,
            platz.printing,
            menge,
            platz.bereich,
          )
        }}
        aufBereich={(platz, bereich) => {
          void aktionen.bereichWechseln(platz, bereich)
        }}
        aufNehmen={(platzID) => {
          void aktionen.nehmen(platzID)
        }}
      />
    </section>
  )
}

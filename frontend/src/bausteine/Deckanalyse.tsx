// **Die Analyse eines Decks** — Werkstück U10 von Phase 5.1
// (#1421).
//
// **Sie steht im Inspector und ändert nichts.** Das ist die Zusage aus dem
// Issue: Kurve, Farben, Typen, Kennzahlen, Urteil und Zuordnung sind Blicke
// auf dasselbe Deck — geschrieben wird im Zentrum.
//
// **Eine eigene Datei**, damit `Deckwerkbank.tsx` unter der Zeilengrenze aus
// §2 bleibt; der Schnitt folgt der Frage: dort steht, was man **tut**, hier,
// was man **sieht**.
import { Deckauswertung } from './Deckauswertung'
import { Deckkennzahlen } from './Deckkennzahlen'
import { Deckurteil } from './Deckurteil'
import { Deckzuordnung } from './Deckzuordnung'
import { Laschen } from './Laschen'
import { useTexte } from '../sprache/kontext'

import type { Deckpruefung } from '../haken/useDeckpruefung'
import type { Deckstand } from '../haken/useDecks'
import type { DeckDTO, HabenstandAntwort } from '../apidecks'
import type { Ladezustand } from '../haken/useLaden'
import type { Statistikposten } from './deckstatistik'

export function Deckanalyse({
  stand,
  deck,
  pruefung,
  abgleich,
  oeffne,
}: {
  stand: Deckstand
  deck: DeckDTO
  pruefung: Deckpruefung
  abgleich: Ladezustand<HabenstandAntwort | null>

  // oeffne schlägt eine Lasche von außen auf — die Prüfung, wenn die
  // Aufmerksamkeit wegen fehlender Karten hinführt (AP-8, #2365).
  oeffne?: string | undefined
}) {
  const eigene = useTexte().seiten.decks

  // **Die Posten für die Rechnung** — aus der Antwort, ohne zweiten Ruf
  // (#1376).
  //
  // **Ungefiltert**: wer nicht gezählt wird, sagt `wertBekannt` in
  // `deckrechnung.ts`, und zwar an **einer** Stelle für Kurve und Schnitt —
  // warum ein Filter über leere Kosten zu grob ist, steht dort.
  //
  // **Ein Landfilter steht hier ausdrücklich nicht.** Der erste Anlauf hatte
  // einen — und die Gegenprobe blieb grün: ein Land trägt nie Manakosten und
  // nie ein `cmc` über null, `wertBekannt` hält es also ohnehin heraus. Ein
  // Filter, der nichts ändert, ist eine Zusage ohne Deckung (§10); was
  // wirklich gilt, sagt der Hinweis über der Kurve.
  const posten: Statistikposten[] = deck.plaetze.map((platz) => ({
    typzeile: platz.typzeile,
    kosten: platz.manakosten.trim(),
    menge: platz.menge,

    // **Nur ein Wert über null reist mit.** Die Naht führt `manawert` als
    // Zahl, und eine Null heißt dort beides — „kostet nichts" und „nicht
    // gefragt". Für eine eigene Karte ist sie immer null, und dort rechnet
    // [wertVon] aus den Kosten; für einen Druck mit `cmc` 0 stehen die Kosten
    // als `{0}` da und geben dieselbe Null.
    ...(platz.manawert > 0 ? { manawert: platz.manawert } : {}),
  }))

  return (
    <Laschen
      bezeichnung={eigene.gebieteDeck}
      oeffne={oeffne}
      gebiete={[
        {
          schluessel: 'auswertung',
          name: eigene.gebietAuswertung,
          inhalt: (
            <div className="flex flex-col gap-6">
              <Deckauswertung posten={posten} />
              <Deckkennzahlen posten={posten} />
            </div>
          ),
        },
        {
          schluessel: 'pruefung',
          name: eigene.gebietPruefung,
          inhalt: (
            <Deckurteil
              stand={stand}
              deck={deck}
              pruefung={pruefung}
              abgleich={abgleich}
            />
          ),
        },
        {
          schluessel: 'zuordnung',
          name: eigene.gebietZuordnung,
          inhalt: <Deckzuordnung stand={stand} deck={deck} />,
        },
      ]}
    />
  )
}

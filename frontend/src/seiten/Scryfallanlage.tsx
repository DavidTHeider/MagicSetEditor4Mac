// **Der zweite Weg zu einer neuen Karte: aus Scryfall** —
// #1240.
//
// # Warum er nicht offen dasteht
//
// **Weil er zwei Eingabefelder hat und der erste Weg einen Knopf.** Neben
// „Karte anlegen" in einer Umbruchzeile wären das fünf Bedienelemente für
// den selteneren Weg, dauerhaft ausgebreitet — und der Umbruch risse Feld und
// Knopf auseinander. **Ein Feld und der Knopf, der es abschickt, gehören
// zusammen** — hier stehen sie in einem eigenen Kasten und können gar nicht
// getrennt werden.
//
// **Seit UXR-05 (#1913) öffnet ihn das Menü „Mehr"** neben „Karte anlegen"
// (`Kartenwerkzeuge.tsx`), nicht mehr ein eigener Aufklapper: über dem
// Arbeitsplatz steht eine Handlung, und die seltenen Wege teilen sich einen
// Knopf. Diese Datei hieß bis dahin `Kartenanlegen.tsx` und trug auch den
// Hauptknopf; was sie jetzt trägt, ist der Inhalt des Kastens.
import type { SuchtrefferDTO } from '../apiscryfall'
import type { Vorgang } from '../haken/useVorgang'
import { Scryfallkarte } from '../bausteine/Scryfallkarte'
import { useTexte } from '../sprache/kontext'
import { Scryfallsuche } from './Scryfallsuche'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface AnlageEigenschaften {
  vorgang: Vorgang

  // ausScryfall legt eine Karte aus einem Scryfall-Namen an (#1032).
  ausScryfall: (name: string, art: string) => void

  // gewaehlt ist der Name, den die Suche in das Namensfeld gelegt hat.
  gewaehlt: string

  waehlen: (treffer: SuchtrefferDTO | undefined) => void
}

export function Scryfallanlage({
  vorgang,
  ausScryfall,
  gewaehlt,
  waehlen,
}: AnlageEigenschaften) {
  const eigene = useTexte().seiten.karten

  return (
    <div className="flex flex-col gap-3">
      <Erklaerungsabsatz
        text={eigene.ausScryfallErklaerung}
        klasse="text-sm text-leise"
      />

      {/* **Der zweite Weg zur Karte** (#1032). Bis dahin musste man erst
          eine leere anlegen und dann abgleichen — zwei Schritte für einen
          Vorgang, und den ersten musste man kennen. */}
      <Scryfallkarte
        key={gewaehlt}
        name={gewaehlt}
        anlegen={ausScryfall}
        laeuft={vorgang.laeuft('scryfallkarte')}
      />

      {/* **Und die Suche darunter** (#1167). Bis dahin musste man den
          Namen kennen, wie Scryfall ihn führt — »Lightning Bolt«, nicht
          »Blitzschlag«. Wer ihn nicht kannte, kam nicht weiter.

          **Was sie tut**: sie füllt den Namen darüber — und seit
          Werkstück 2 von Phase 3.6 hält sie die **Kennung des Drucks**
          fest, damit die aufgeschlagene Karte sich darauf beziehen kann. */}
      <Scryfallsuche waehlen={waehlen} />
    </div>
  )
}

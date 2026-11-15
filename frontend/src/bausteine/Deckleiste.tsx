// **Die Kennzahlen des offenen Decks** — AP-9
// (#2366),
// Entwurf `Bereiche09`: „Hauptdeck 60 · Sideboard 15“, „Prüfung gültig ·
// Format Konstruiert“, „Fehlt in der Sammlung 4“.
//
// # Jede Zahl mit Ziel
//
// **Hauptdeck führt zu den Plätzen**, Prüfung und Fehlendes zur Lasche
// „Prüfung“, wo Urteil und Abgleich im Einzelnen stehen. Die Zahlen sind
// dieselben wie dort: die Prüfung ist derselbe Haken, der Abgleich dieselbe
// Antwort (`useDeckabgleich`).
//
// # Was die Leiste nicht trägt
//
// **„Wert“ und „Gespielt“ aus dem Entwurf** — einen Wert je Deck rechnet
// heute nichts, und die Partien zählen nicht nach Deck. Eine Zahl dafür wäre
// erfunden; die Abweichung steht in `docs/PROTOTYP-0.9.0.md`.
import type { DeckDTO, HabenstandAntwort } from '../apidecks'
import type { Deckpruefung } from '../haken/useDeckpruefung'
import type { Ladezustand } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Kennzahlleiste } from './Kennzahlleiste'

// urteilswort sagt, wie die Prüfung steht — **ungeprüft ist kein „nicht
// zulässig“** (Vermerk 0061): ohne Format und nach einem Fehler steht ein
// Satz, keine Behauptung über das Deck.
function urteilswort(
  pruefung: Deckpruefung,
  eigene: {
    legal: string
    postenUnzulaessig: string
    leiste: { ohneFormat: string; nichtGeprueft: string }
  },
): string | undefined {
  if (pruefung.ohneFormat) {
    return eigene.leiste.ohneFormat
  }

  if (pruefung.urteil !== null) {
    return pruefung.urteil.zulaessig ? eigene.legal : eigene.postenUnzulaessig
  }

  return pruefung.fehler === undefined ? undefined : eigene.leiste.nichtGeprueft
}

export function Deckleiste({
  deck,
  pruefung,
  abgleich,
  zeigePlaetze,
  zeigePruefung,
}: {
  deck: DeckDTO
  pruefung: Deckpruefung
  abgleich: Ladezustand<HabenstandAntwort | null>
  zeigePlaetze: () => void
  zeigePruefung: () => void
}) {
  const eigene = useTexte().seiten.decks
  const bestand = abgleich.daten

  return (
    <Kennzahlleiste
      bezeichnung={eigene.leiste.bezeichnung}
      kennzahlen={[
        {
          schluessel: 'hauptdeck',
          was: eigene.hauptdeck,
          wert: String(deck.hauptdeck),
          unter: einsetzen(eigene.leiste.sideboard, {
            zahl: String(deck.sideboard),
          }),
          hin: zeigePlaetze,
        },
        {
          schluessel: 'pruefung',
          was: eigene.leiste.pruefung,
          wert: urteilswort(pruefung, eigene),
          unter: pruefung.urteil?.formatname,
          hin: zeigePruefung,
        },
        {
          schluessel: 'fehlt',
          was: eigene.leiste.fehlt,
          wert:
            abgleich.fehler !== undefined
              ? eigene.leiste.nichtAbgeglichen
              : bestand?.fehlend.toString(),
          unter:
            bestand !== undefined && bestand !== null && bestand.unbekannt > 0
              ? einsetzen(eigene.leiste.unbekannt, {
                  zahl: String(bestand.unbekannt),
                })
              : undefined,
          hin: zeigePruefung,
        },
      ]}
    />
  )
}

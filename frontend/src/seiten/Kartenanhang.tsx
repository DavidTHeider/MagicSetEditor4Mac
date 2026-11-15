// **Was unter der Kartenlasche steht** — die Massenänderung und der Verlauf.
//
// # Warum eine eigene Datei
//
// Weil `Karten.tsx` über die Zeilengrenze ging (§2). Der Schnitt folgt der
// Frage: dort steht, **wie** die Lasche angeordnet ist, hier, **was unter ihr
// hängt**.
//
// # Was hier nicht mehr steht
//
// **Die Vorbehalte des Imports** — Werkstück I von Phase 4.10
// (#1325).
// Sie handeln vom **Set** und standen am Ende einer 1.728 Punkte hohen
// Kartenseite, unter drei ungleich langen Spalten; jetzt stehen sie bei den
// Setangaben.

import type { KartenzeileDTO } from '../api'
import { Flaeche } from '../bausteine/Flaeche'
import { useTexte } from '../sprache/kontext'
import { Massenaenderung } from './Massenaenderung'
import { Verlaufsliste } from './Verlaufsliste'

export function Kartenanhang({
  setID,
  gewaehlteKarten,
  nachladen,
  karten,
  verlaufsstand,
  zuruecknehmen,
}: {
  setID: string

  // gewaehlteKarten sind die **angehakten** — die Massenänderung meint sie.
  gewaehlteKarten: readonly KartenzeileDTO[]

  nachladen: () => void

  // karten sind alle Karten des Sets — die Verlaufsliste nennt sie beim
  // Namen (U6, #1417).
  karten: readonly KartenzeileDTO[]

  // verlaufsstand steigt nach jeder Bewegung; die Liste lädt dann neu.
  verlaufsstand: number

  zuruecknehmen?: ((vorgang: string) => void) | undefined
}) {
  const texte = useTexte()

  return (
    <>
      {/* **Die Massenänderung erscheint, wenn etwas angehakt ist.** Ein
        Abschnitt, der bei jedem Blick dasteht und meistens leer ist, kostet
        Platz und Aufmerksamkeit. */}
      {gewaehlteKarten.length > 0 && (
        <Flaeche ueberschrift={texte.seiten.karten.massen.ueberschrift}>
          <Massenaenderung
            setID={setID}
            karten={[...gewaehlteKarten]}
            nachladen={nachladen}
          />
        </Flaeche>
      )}

      {/* **Der Verlauf steht immer da** (U6, #1417): was zuletzt geschah,
          gehört zu jedem Blick auf das Set — und ⌘Z bewegt genau seinen
          obersten Schritt. */}
      <Flaeche ueberschrift={texte.seiten.karten.verlauf.ueberschrift}>
        <Verlaufsliste
          setID={setID}
          karten={karten}
          verlaufsstand={verlaufsstand}
          zuruecknehmen={zuruecknehmen}
        />
      </Flaeche>
    </>
  )
}

// **Die Sets, die das Bündel mitbringt** —
// #702.
//
// # Wofür
//
// Der erste Weg einer frisch geladenen Anwendung führte durch einen
// Dateidialog — auf einen Rechner, auf dem noch kein `.mse-set` liegt. Wer
// nachsehen wollte, was dieses Programm kann, musste sich erst eines besorgen.
//
// # Warum kein Dialog davor
//
// **Ein Dialog stellt eine Frage, und hier ist keine offen.** Ein Klick auf den
// Namen hat genau eine Antwort.
//
// # Eigene Datei
//
// Weil `Sets.tsx` an der Zeilengrenze steht (§2) — und weil der Schnitt einer
// Frage folgt: dort steht, was mit den **offenen** Sets geschieht, hier, wie
// eines dazukommt, das schon mitkam.
import { useCallback } from 'react'

import { api, auspacken, type BeispielsetDTO } from '../api'
import { useLaden } from '../haken/useLaden'
import { useTexte } from '../sprache/kontext'
import { Flaeche } from '../bausteine/Flaeche'
import { Knopf } from '../bausteine/Knopf'

interface BeispielsetEigenschaften {
  laeuft: boolean
  oeffnen: (pfad: string) => void
}

export function Beispielsets({ laeuft, oeffnen }: BeispielsetEigenschaften) {
  const texte = useTexte()

  const holen = useCallback(async () => {
    const antwort = await api().Beispielsets()

    return auspacken(antwort, () => antwort.sets)
  }, [])

  const { daten } = useLaden<BeispielsetDTO[]>(holen, [])

  // **Ohne Bündel steht der Abschnitt nicht da.** Ein leerer Kasten
  // „Beispielsets" wäre ein Versprechen ohne Inhalt — und beim Entwickeln der
  // Normalfall.
  //
  // **Ein Ladefehler ebenso wenig.** Er kommt aus einem Verzeichnis, das der
  // Freigabeweg angelegt hat; wer ihn hier meldete, stellte einen Hinweis über
  // eine Seite, auf der jemand etwas ganz anderes tut. Er steht im Lauf.
  if (daten === undefined || daten.length === 0) {
    return null
  }

  return (
    <Flaeche ueberschrift={texte.seiten.sets.beispielsets}>
      <div className="flex flex-wrap gap-3">
        {daten.map((satz) => (
          <Knopf
            key={satz.pfad}
            angehalten={laeuft}
            aufKlick={() => {
              oeffnen(satz.pfad)
            }}
          >
            {satz.name}
          </Knopf>
        ))}
      </div>
    </Flaeche>
  )
}

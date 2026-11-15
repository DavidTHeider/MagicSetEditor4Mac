// Die bekannten Bezugsquellen —
// #263.
//
// **Was sie tut, steht hier; wie sie aussieht, in den Klassen** (§2).
//
// # Wofür
//
// Der Paketverwalter kann seit Phase 7 beziehen; was fehlte, ist die letzte
// Meile. Bis hierher musste jemand eine Adresse abtippen **und** einen
// vierzigstelligen Commit dazu — oder die eine Vorgabe nehmen, die im Formular
// stand und deren Herkunft nirgends erklärt war.
//
// # Was ein Klick **nicht** überspringt
//
// **Die Zustimmung.** Ein Klick trägt Adresse, Stand und Namen ein — mehr
// nicht. Danach steht dieselbe Frage wie vorher: „Herkunft anzeigen", dann
// „Beziehen". Das ist die Zusage aus #139 und #56, und sie ist der Grund,
// warum hier kein Knopf „jetzt holen" steht.
import { api, auspacken, type QuelleDTO } from '../api'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { useCallback } from 'react'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface Eigenschaften {
  // uebernehmen trägt die Quelle ins Formular — es bezieht nicht.
  uebernehmen: (quelle: QuelleDTO) => void
}

export function Bezugsquellen({ uebernehmen }: Eigenschaften) {
  const texte = useTexte()

  const holen = useCallback(async () => {
    const antwort = await api().Bezugsquellen()

    return auspacken(antwort, () => antwort.quellen)
  }, [])
  const { daten, fehler } = useLaden<QuelleDTO[]>(holen, [])

  // **Der Satz steht im Katalog, der Schlüssel kommt von Go** (§9). Ein
  // deutscher Fließtext aus der Bindung stünde in der englischen Oberfläche
  // deutsch; ein Wächter in `internal/werkzeuge` hält beide Seiten zusammen.
  const beschreibung = (schluessel: string) =>
    texte.seiten.pakete.quellen[
      schluessel as keyof typeof texte.seiten.pakete.quellen
    ]

  return (
    <div className="mt-3">
      <Erklaerungsabsatz
        text={texte.seiten.pakete.quellenErklaerung}
        klasse="max-w-prose text-leise"
      />

      {fehler !== undefined && (
        <p role="alert" className="mt-2 text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      <ul className="mt-3 grid gap-2 sm:grid-cols-2">
        {(daten ?? []).map((quelle) => (
          <li key={quelle.name}>
            {/* **Ein Knopf über die ganze Kachel.** Wer mit der Tastatur
                kommt, hält einmal an und hat alles gehört, was die Kachel
                sagt — Name, Größe, Satz (§6). */}
            <button
              type="button"
              onClick={() => {
                uebernehmen(quelle)
              }}
              className="flex h-full w-full flex-col gap-1 rounded-md border border-kante bg-flaeche p-3 text-left focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            >
              <span className="flex flex-wrap items-baseline gap-2">
                <span className="font-medium">{quelle.name}</span>
                <span className="text-sm text-leise">
                  {einsetzen(texte.seiten.pakete.quellenGroesse, {
                    megabyte: String(quelle.megabyte),
                  })}
                </span>
              </span>

              <span className="text-sm text-leise">
                {beschreibung(quelle.schluessel)}
              </span>
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}

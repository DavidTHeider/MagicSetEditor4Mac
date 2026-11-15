// **Einen Nutzer anlegen** — Phase 4.12
// (#1351).
//
// **Der Fehler steht am verursachenden Feld** (§6) — die Lehre aus dem
// Nachlauf über Phase 4.10, wo die Paketanlage ihren Feldbezug wegwarf.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Bindungsfehler } from '../bindungsfehler'
import { Knopf } from '../bausteine/Knopf'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

export function Nutzeranlage({ angelegt }: { angelegt: () => void }) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const [name, setzeName] = useState('')
  const [laeuft, setzeLaeuft] = useState(false)
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)
  const [feldfehler, setzeFeldfehler] = useState<string | undefined>(undefined)
  const [erfolg, setzeErfolg] = useState<string | undefined>(undefined)

  const anlegen = async () => {
    setzeLaeuft(true)
    setzeFehler(undefined)
    setzeFeldfehler(undefined)
    setzeErfolg(undefined)

    try {
      const antwort = await api().NutzerAnlegen(name, '')

      auspacken(antwort, () => antwort.nutzer)

      setzeErfolg(einsetzen(eigene.nutzerAngelegt, { name }))
      setzeName('')

      // **Geschrieben ist geschrieben** (§7, Regel 3): das Nachladen ist
      // Bestenfalls-Arbeit und macht den Nutzer nicht ungeschehen.
      angelegt()
    } catch (grund) {
      const satz = meldungAus(grund, texte)

      if (grund instanceof Bindungsfehler && grund.feld === 'name') {
        setzeFeldfehler(satz)

        return
      }

      setzeFehler(satz)
    } finally {
      setzeLaeuft(false)
    }
  }

  return (
    <section aria-label={eigene.neuerNutzer} className="flex flex-col gap-3">
      {/* **Ohne eigene Überschrift** — seit UXR-20 (#1928) steht der
          Abschnitt in einem Aufklappkasten, der heißt, wie sie hieß. */}
      <div className="flex flex-wrap items-end gap-3">
        <Pflichtfeld
          beschriftung={eigene.neuerNutzerName}
          wert={name}
          befund={feldfehler}
          hinweis={eigene.neuerNutzerNameHinweis}
          geaendert={setzeName}
        />

        <Knopf
          gewicht="haupt"
          angehalten={laeuft || name.trim() === ''}
          aufKlick={() => {
            void anlegen()
          }}
        >
          {eigene.neuerNutzerKnopf}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      {erfolg !== undefined && (
        <p role="status" className="text-leise">
          {erfolg}
        </p>
      )}
    </section>
  )
}

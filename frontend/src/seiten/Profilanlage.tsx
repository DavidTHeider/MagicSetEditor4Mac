// **Ein weiteres Profil anlegen** — Phase 4.12
// (#1351).
//
// **Seit UXR-20 in der Schublade hinter *+ Profil anlegen …*** im Kopf der
// Nutzerseite (#1928).
// Eine eigene Überschrift trägt der Abschnitt nicht mehr: die Schublade
// heißt, wie er hieß, und zwei gleiche Titel untereinander wären einer zu
// viel.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { Bindungsfehler } from '../bindungsfehler'
import { Knopf } from '../bausteine/Knopf'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { NutzerDTO } from '../api'

interface Eigenschaften {
  nutzer: readonly NutzerDTO[]
  angelegt: () => void
}

export function Profilanlage({ nutzer, angelegt }: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer

  const [fuer, setzeFuer] = useState('')
  const [name, setzeName] = useState('')
  const [laeuft, setzeLaeuft] = useState(false)
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)
  const [feldfehler, setzeFeldfehler] = useState<string | undefined>(undefined)
  const [erfolg, setzeErfolg] = useState<string | undefined>(undefined)

  // **Der erste Nutzer ist vorgewählt** — es gibt immer einen (Wanderung
  // 0034), und eine leere Wahl wäre ein Schritt, den niemand auslassen kann.
  const gewaehlt = fuer === '' ? (nutzer[0]?.id ?? '') : fuer

  const anlegen = async () => {
    setzeLaeuft(true)
    setzeFehler(undefined)
    setzeFeldfehler(undefined)
    setzeErfolg(undefined)

    try {
      const antwort = await api().ProfilAnlegen(gewaehlt, name)

      auspacken(antwort, () => antwort.profil)

      setzeErfolg(einsetzen(eigene.profilAngelegt, { name }))
      setzeName('')
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
    <section aria-label={eigene.neuesProfil} className="flex flex-col gap-3">
      <div className="flex flex-wrap items-end gap-3">
        <label className="flex flex-col gap-1">
          <span className="text-sm">{eigene.neuesProfilNutzer}</span>

          <select
            value={gewaehlt}
            onChange={(ereignis) => {
              setzeFuer(ereignis.target.value)
            }}
            className="rounded-md border border-kante px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {nutzer.map((einer) => (
              <option key={einer.id} value={einer.id}>
                {einer.name}
              </option>
            ))}
          </select>
        </label>

        <Pflichtfeld
          beschriftung={eigene.neuesProfilName}
          wert={name}
          befund={feldfehler}
          hinweis={eigene.neuesProfilNameHinweis}
          geaendert={setzeName}
        />

        <Knopf
          gewicht="haupt"
          angehalten={laeuft || name.trim() === '' || gewaehlt === ''}
          aufKlick={() => {
            void anlegen()
          }}
        >
          {eigene.neuesProfilKnopf}
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

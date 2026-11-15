// **Einen Code eingeben** — NET-01
// (#1705).
//
// **Prüfen schreibt nichts.** Erst der Vergleich darunter entscheidet, und er
// hat drei Ausgänge: abbrechen (nichts), später vergleichen (aufgenommen,
// unbestätigt) und „stimmt überein" (bestätigt). Einen halb bestätigten
// Kontakt gibt es damit nicht (AK-04).
import { useId } from 'react'

import { Einladungsvergleich } from './Einladungsvergleich'
import { Identitaetanlegen } from './Identitaetanlegen'
import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'

import type { Einladungstausch } from '../haken/useEinladungstausch'
import { kennungseingabe } from './kennungseingabe'

export function Einladungsannahme({ tausch }: { tausch: Einladungstausch }) {
  const texte = useTexte()
  const eigene = texte.seiten.technik.einladung
  const codefeld = useId()

  // ausloeser ist der Fehler dieses Schritts — ihn bekommt das Anlegen der
  // Identität mit, damit sein Ende nur zu ihm gehört (§7, Regel 2).
  const ausloeser = tausch.fehler?.ort === 'pruefen' ? tausch.fehler : undefined
  const fehler = ausloeser?.text

  return (
    <div className="flex flex-col gap-3">
      <label className="flex flex-col gap-1 text-sm" htmlFor={codefeld}>
        {eigene.eingabe}
        <textarea
          {...kennungseingabe}
          id={codefeld}
          rows={3}
          placeholder={eigene.eingabeBeispiel}
          value={tausch.eingabe}
          aria-invalid={fehler !== undefined}
          className="rounded-sm border border-kante bg-transparent p-2 font-mono text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          onChange={(ereignis) => {
            tausch.setzeEingabe(ereignis.target.value)
          }}
        />
      </label>

      {/* **Fehlt die eigene Identität, steht ihr Anlegen hier** (G01): der
          Kasten, den die Meldung nannte, steht nicht in dieser Schublade. */}
      {ausloeser?.ohneIdentitaet === true ? (
        <Identitaetanlegen
          angelegt={() => {
            tausch.identitaetAngelegt(ausloeser)
          }}
        />
      ) : (
        fehler !== undefined && (
          <p role="alert" className="text-sm text-fehler">
            {fehler}
          </p>
        )
      )}

      {tausch.vorschau === undefined ? (
        <div>
          <Knopf
            gewicht="haupt"
            angehalten={
              tausch.laeuft('pruefen') || tausch.eingabe.trim() === ''
            }
            aufKlick={tausch.pruefen}
          >
            {eigene.pruefen}
          </Knopf>
        </div>
      ) : (
        // **Der `key` hängt am Code** (§7, Regel 1): eine andere Einladung
        // bekommt einen frischen Vergleich, samt Name und Häkchen.
        <Einladungsvergleich
          key={tausch.vorschau.kennung + tausch.vorschau.ablauf}
          vorschau={tausch.vorschau}
          tausch={tausch}
        />
      )}
    </div>
  )
}

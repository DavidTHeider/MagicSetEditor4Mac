// **Apples Gerätemodell für diese Sitzung** — AI-05.AK-04
// (#2042).
//
// **Ein Schalter, weil er sofort wirkt** (R7): aus beendet ein laufendes
// `fm serve`, und bis zum Neustart fragt der Vorrang es nicht. **Er ist keine
// Einstellung** — `ki.geraetemodell` darüber gilt über Neustarts hinweg; dieser
// hier lebt im Arbeitsspeicher und steht deshalb nicht mit `data-einstellung`
// da.
import { useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'

interface KIGeraetSitzungEigenschaften {
  // an ist der Stand, wie der Katalog ihn meldet.
  an: boolean

  // gemeldet lädt den Katalog neu — **nur nach einem geglückten Schalten**
  // (§7, Regel 3).
  gemeldet: () => void
}

export function KIGeraetSitzung({
  an,
  gemeldet,
}: KIGeraetSitzungEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  const vorgang = useVorgang()
  const [fehler, setzeFehler] = useState<string>()

  async function schalten(neu: boolean) {
    setzeFehler(undefined)

    await vorgang.starte('geraetesitzung', async () => {
      try {
        auspacken(await api().KIGeraetemodellFuerSitzung(neu), () => null)

        gemeldet()
      } catch (ursache) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <div className="flex flex-col gap-1">
      <label className="flex items-center gap-2">
        <input
          type="checkbox"
          role="switch"
          checked={an}
          disabled={vorgang.laeuft('geraetesitzung')}
          onChange={(ereignis) => {
            void schalten(ereignis.target.checked)
          }}
          className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
        />
        <span>{eigene.kiGeraetSitzung}</span>
      </label>

      <span className="text-leise">{eigene.kiGeraetSitzungHinweis}</span>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}
    </div>
  )
}

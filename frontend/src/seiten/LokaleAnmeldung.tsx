// **Anmelden nach dem Abmelden** — ID-02
// (#1701).
//
// **Ein Nutzer dieses Rechners, sein Kennwort, dann sein zweiter Faktor** —
// keine Adresse, keine Registrierung (AK-01). Wer kein Kennwort hat, kommt
// ohne herein; das ist seine Wahl und keine Lücke dieser Fläche.
import { useId, useState } from 'react'

import { api } from '../fassade'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { Kennwortfeld } from '../bausteine/Kennwortfeld'
import { Knopf } from '../bausteine/Knopf'
import { Nachweisfeld } from '../bausteine/Nachweisfeld'
import { useLaden } from '../haken/useLaden'
import { useSitzungsschritt } from '../haken/useSitzungsschritt'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'

export function LokaleAnmeldung({ angemeldet }: { angemeldet: () => void }) {
  const texte = useTexte()
  const eigene = texte.seiten.nutzer.sitzung
  const schritt = useSitzungsschritt(angemeldet)
  const wahlfeld = useId()

  const nutzer = useLaden(async () => {
    const antwort = await api().Nutzer()

    return auspacken(antwort, () => antwort.nutzer)
  }, [])

  const [gewaehlt, setzeGewaehlt] = useState('')
  const [kennwort, setzeKennwort] = useState('')

  const liste = nutzer.daten ?? []
  const wer = gewaehlt === '' ? (liste[0]?.id ?? '') : gewaehlt

  const anmelden = (nachweis = '') => {
    void schritt.fuehre(() => api().LokalAnmelden(wer, kennwort, nachweis))
  }

  return (
    <main
      data-sitzung="abgemeldet"
      className="mx-auto flex max-w-md flex-col gap-4 p-8"
    >
      <h1 className="text-2xl font-semibold">{eigene.abgemeldetTitel}</h1>
      <Erklaerungsabsatz
        text={eigene.abgemeldetErklaerung}
        klasse="text-leise"
      />

      {nutzer.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {nutzer.fehler}
        </p>
      )}

      <div className="flex flex-col gap-1">
        <label htmlFor={wahlfeld} className="text-sm">
          {eigene.wer}
        </label>
        <select
          id={wahlfeld}
          value={wer}
          onChange={(ereignis) => {
            setzeGewaehlt(ereignis.target.value)
            setzeKennwort('')
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        >
          {liste.map((einer) => (
            <option key={einer.id} value={einer.id}>
              {einer.name}
            </option>
          ))}
        </select>
      </div>

      <Kennwortfeld
        beschriftung={eigene.kennwort}
        wert={kennwort}
        befund={schritt.befund.kennwort}
        geaendert={setzeKennwort}
      />

      {schritt.faktorGefragt && (
        <Nachweisfeld angehalten={schritt.laeuft} abgeschickt={anmelden} />
      )}

      <div>
        <Knopf
          gewicht="haupt"
          angehalten={wer === '' || schritt.laeuft}
          aufKlick={() => {
            anmelden()
          }}
        >
          {eigene.anmelden}
        </Knopf>
      </div>

      {schritt.fehler !== undefined && (
        <p role="alert" className="text-fehler">
          {texte.bausteine.zustaende.fehler} {schritt.fehler}
        </p>
      )}
    </main>
  )
}

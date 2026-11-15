// **Eigene Übersetzungen in den Einstellungen** — LOC-02
// (#1688).
//
// Unter der Sprache, in *Allgemein*: welche eigene gilt, welche es gibt,
// eine neue anlegen oder einlesen. **Bearbeitet wird im selben Bereich** —
// die Liste weicht dem Editor, kein Modal über einem Modal (§6). Der `key`
// am Editor räumt beim Wechsel des Pakets alles ab (§7, Regeln 1 und 2).
import { useState } from 'react'

import { useTexte } from '../sprache/kontext'
import { useSprachpakete } from '../haken/useSprachpakete'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Neuanlage, Paketzeile } from './Eigenuebersetzungzeilen'
import { Uebersetzungseditor } from './Uebersetzungseditor'

import type { EinstellungDTO } from '../api'

// anlagefelder sind die Felder der Neuanlage — ihr Befund steht dort.
const anlagefelder = ['kennung', 'name', 'basis']

export function Eigenuebersetzungen({
  eintrag,
  angehalten,
  setze,
  eingebettet = false,
}: {
  eintrag: EinstellungDTO | undefined
  angehalten: boolean
  setze: (wert: string) => void

  // eingebettet lässt die Überschrift weg — in den Einstellungen steht die
  // Liste in einem Aufklapper, dessen Knopf schon so heißt (UXR-19, #1927).
  eingebettet?: boolean
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen.eigenuebersetzung
  const [bearbeitet, setzeBearbeitet] = useState<string>()
  const stand = useSprachpakete(setzeBearbeitet)

  if (bearbeitet !== undefined) {
    return (
      <Uebersetzungseditor
        key={bearbeitet}
        kennung={bearbeitet}
        zurueck={() => {
          setzeBearbeitet(undefined)
          stand.liste.nachladen()
        }}
      />
    )
  }

  const pakete = stand.liste.daten ?? []

  return (
    <section
      aria-label={eigene.titel}
      className="flex max-w-3xl flex-col gap-4"
    >
      {!eingebettet && <h3 className="font-medium">{eigene.titel}</h3>}
      <Erklaerungsabsatz
        text={eigene.erklaerung}
        klasse="max-w-prose text-sm text-leise"
      />

      {eintrag !== undefined && (
        <label className="flex flex-col gap-1 text-sm">
          {eigene.verwenden}
          <select
            value={eintrag.wert}
            disabled={angehalten}
            onChange={(ereignis) => {
              setze(ereignis.target.value)
            }}
            className="max-w-sm rounded-md border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            <option value="">{eigene.keine}</option>
            {pakete.map((paket) => (
              <option key={paket.kennung} value={paket.kennung}>
                {`${paket.name} (${paket.kennung})`}
              </option>
            ))}
          </select>
        </label>
      )}

      {stand.liste.daten !== undefined && pakete.length === 0 && (
        <p className="text-sm text-leise">{eigene.leer}</p>
      )}

      <ul className="flex flex-col divide-y divide-kante">
        {pakete.map((paket) => (
          <Paketzeile
            key={paket.kennung}
            paket={paket}
            angehalten={stand.laeuft(paket.kennung)}
            bearbeiten={() => {
              setzeBearbeitet(paket.kennung)
            }}
            ausgeben={() => void stand.ausgeben(paket)}
            loeschen={() => void stand.loeschen(paket)}
          />
        ))}
      </ul>

      <Neuanlage
        angehalten={stand.laeuft('+neu')}
        feldbefund={(name) =>
          stand.befund?.feld === name ? stand.befund.text : undefined
        }
        anlegen={(kennung, name, basis) =>
          void stand.anlegen(kennung, name, basis)
        }
        einlesen={(ersetzen) => void stand.einlesen(ersetzen)}
      />

      {stand.befund !== undefined &&
        !anlagefelder.includes(stand.befund.feld) && (
          <p role="alert" className="text-sm text-fehler">
            {stand.befund.text}
          </p>
        )}

      {stand.meldung !== undefined && (
        <p role="status" className="text-sm text-leise">
          {stand.meldung}
        </p>
      )}
    </section>
  )
}

// **Die Wahl eines Modells** — herausgezogen für Werkstück A12 von Phase 5.3
// (#1451).
//
// # Warum ein eigenes Bauteil
//
// **Weil der Bildvorschlag dieselbe Wahl braucht wie der Textvorschlag** —
// und eine zweite Fassung daneben liefe beim ersten Sonderfall auseinander.
// Sie stand bis dahin in `KIVorschlag.tsx`; hier steht sie unverändert, nur
// an einer Stelle statt an zweien.
//
// # `gefragt` steht im Aufruf und nicht hier drin
//
// **Weil der Aufrufer weiß, wann jemand etwas gewollt hat.** `KIModelle`
// fragen geht nach draußen — ist ein Cloudanbieter eingeschaltet, führt der
// Ruf zu seinem Dienst. Das Aufschlagen einer Karte darf das nicht auslösen
// (§16, Regel 6); das Aufklappen eines Kastens schon, denn das hat jemand
// getan.
import { KIEinrichtungsverweis } from './KIEinrichtungsverweis'
import { useTexte } from '../sprache/kontext'

import type { KIModellDTO } from '../antworten'

interface KIModellwahlEigenschaften {
  // gefragt sagt, ob schon gefragt wurde — davor ist eine leere Liste nicht
  // leer, sondern **ungelesen**, und das sind zwei Auskünfte.
  gefragt: boolean

  // modelle, laedt und fehler kommen aus [useKIModelle] beim Aufrufer.
  //
  // **Der Haken hängt dort und nicht hier**: dieses Bauteil wird abgeräumt,
  // wenn jemand den Kasten zuklappt — und ein Katalog darin fragte beim
  // nächsten Aufklappen erneut.
  modelle: KIModellDTO[]
  laedt: boolean
  fehler: string | undefined

  // gewaehlt ist `anbieter/alias` — leer heißt: nichts gewählt.
  //
  // **Ein Schlüssel und nicht zwei Zustände**: zwei Felder daneben liefen
  // auseinander.
  gewaehlt: string

  waehlen: (neu: string) => void
}

export function KIModellwahl({
  gefragt,
  modelle,
  laedt,
  fehler,
  gewaehlt,
  waehlen,
}: KIModellwahlEigenschaften) {
  const eigene = useTexte().seiten.einstellungen

  return (
    <>
      <label className="flex flex-col gap-1 text-sm">
        {eigene.kiVorschlagAnbieter}
        <select
          className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
          value={gewaehlt}
          onChange={(ereignis) => {
            waehlen(ereignis.target.value)
          }}
        >
          <option value="">{eigene.kiVorschlagModellWaehlen}</option>

          {modelle.map((modell) => (
            <option
              key={`${modell.anbieter}/${modell.alias}`}
              value={`${modell.anbieter}/${modell.alias}`}
            >
              {modell.anbieter}/{modell.alias}
            </option>
          ))}
        </select>
      </label>

      {/* **Eine leere Wahl erklärt sich** (§6): »Bitte wählen« ohne
          Möglichkeit ist eine Sackgasse, und der Satz sagt, wo es weitergeht.
          **Erst nach dem Fragen**: davor ist die Wahl nicht leer, sondern
          ungelesen, und das sind zwei verschiedene Auskünfte. */}
      {/* **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742): der Satz
          sagt, was fehlt, der Knopf öffnet den Bereich, in dem es fehlt. */}
      {gefragt && !laedt && modelle.length === 0 && fehler === undefined && (
        <div className="flex flex-col items-start gap-1">
          <p className="text-sm text-leise">{eigene.kiVorschlagOhneModell}</p>
          <KIEinrichtungsverweis />
        </div>
      )}

      {fehler !== undefined && (
        <div className="flex flex-col items-start gap-1">
          <p role="alert" className="text-sm text-fehler">
            {fehler}
          </p>
          <KIEinrichtungsverweis />
        </div>
      )}
    </>
  )
}

// **Preise einer ganzen Quelle holen** — Werkstück A3 von Phase 5.3
// (#1442).
//
// # Warum das neben dem Portfolio steht und nicht an einer Karte
//
// `PreiseHolen` fragt **einen** Druck. MTGJSON antwortet nicht je Karte,
// sondern mit einer Datei — und was daraus gebraucht wird, sagt der ganze
// Bestand. Es ist eine Handlung an der **Sammlung**, nicht an einer Karte.
//
// # Die Wahl steht da und ist nicht voreingestellt
//
// Quelle und Umfang sind zwei Listen, und beide zeigen, was gewählt ist. Eine
// stille Vorgabe wäre eine Preisstrategie, die aussieht wie keine — genau das
// verbietet das Issue.
import { useState } from 'react'

import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Bulkbezugsantwort } from '../antworten'

interface PreisbulkEigenschaften {
  // aufGeholt läuft nach einem geglückten Bezug — das Portfolio rechnet neu.
  aufGeholt: () => void
}

// dieQuelle ist die einzige Familie, die es heute gibt.
//
// **Als Konstante und nicht als Text im JSX**: sie ist ein **Name** der
// Anwendung und kein Oberflächentext — `keintextimjsx` weist einen solchen zu
// Recht ab, denn er ließe sich nicht übersetzen und stünde in keinem Katalog.
const dieQuelle = 'mtgjson'

export function Preisbulk({ aufGeholt }: PreisbulkEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  const [umfang, setzeUmfang] = useState('heute')
  const [bilanz, setzeBilanz] = useState<Bulkbezugsantwort>()
  const [meldung, setzeMeldung] = useState<string>()

  const vorgang = useVorgang()

  // **Ein Vorgang je Quelle** (§7, Regel 4): ein zweiter Ruf, während der
  // erste lädt, holte dieselbe Datei ein zweites Mal.
  const laeuft = vorgang.laeuft(dieQuelle)

  const holen = () => {
    void vorgang.starte(dieQuelle, async () => {
      setzeMeldung(undefined)

      try {
        const antwort = await api().Preisbulk(dieQuelle, umfang)

        // **Die Bilanz wird gesetzt, bevor der Fehler geworfen wird**
        // (§7, Regel 3): was abgelegt ist, bleibt abgelegt.
        setzeBilanz(antwort)
        auspacken(antwort, () => undefined)

        aufGeholt()
      } catch (ursache: unknown) {
        setzeMeldung(meldungAus(ursache, texte))
      }
    })
  }

  return (
    // **Die Überschrift trägt der Kasten darum** — der Bulkbezug steht seit
    // UXR-10 eingeklappt im Navigator der Sammlung.
    <section aria-label={eigene.bulkTitel} className="flex flex-col gap-2">
      <Erklaerungsabsatz
        text={eigene.bulkErklaerung}
        klasse="text-sm text-leise"
      />

      <div className="flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm">
          {eigene.bulkUmfang}
          <select
            className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
            value={umfang}
            onChange={(ereignis) => {
              setzeUmfang(ereignis.target.value)
            }}
          >
            <option value="heute">{eigene.bulkHeute}</option>
            <option value="neunzigTage">{eigene.bulkNeunzigTage}</option>
          </select>
        </label>

        <Knopf
          bezeichnung={einsetzen(eigene.bulkHolenBei, { quelle: dieQuelle })}
          aufKlick={holen}
          angehalten={laeuft}
        >
          {laeuft ? eigene.bulkLaeuft : eigene.bulkHolen}
        </Knopf>
      </div>

      {meldung !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {meldung}
        </p>
      )}

      {bilanz?.fehler === null && (
        <p className="text-sm text-leise">
          {bilanz.drucke === 0
            ? eigene.bulkOhneDrucke
            : einsetzen(eigene.bulkBilanz, {
                staende: String(bilanz.staende),
                drucke: String(bilanz.drucke),
              })}
        </p>
      )}

      {bilanz !== undefined && bilanz.befunde.length > 0 && (
        <ul className="flex flex-col gap-1 text-xs text-leise">
          {bilanz.befunde.map((befund) => (
            <li key={befund}>{befund}</li>
          ))}
        </ul>
      )}
    </section>
  )
}

// **Die belegte Einschätzung im Editor** — Werkstück A13.2a von Phase 5.3
// (#1452).
//
// # Warum sie neben dem Vorschlag steht und nicht in ihm
//
// **Weil sie das Gegenteil tut.** Ein Vorschlag will in die Karte; eine
// Einschätzung will nie dorthin — es gibt keine Übernahme, und ein Kasten,
// der beides führte, hätte einen Knopf, der manchmal schreibt.
//
// # Und warum der Vorbehalt über der Antwort steht
//
// **Weil man ihn vorher liest.** Er kommt aus `domain` und reist mit jeder
// Antwort mit; wer ihn ans Ende setzte, hätte ihn dort, wo eine lange Antwort
// abgeschnitten wird.
import { useState } from 'react'

import { Bestandsfreigabe } from './Bestandsfreigabe'
import { KIEinschaetzungsbelege } from './KIEinschaetzungsbelege'
import { KIKostenfreigabe } from './KIKostenfreigabe'
import { KIModellwahl } from './KIModellwahl'
import { aliasAus, anbieterAus, kostetGeld } from './kiwahl'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useKIModelle } from '../haken/useKIModelle'
import { useTexte } from '../sprache/kontext'
import { useVorgang } from '../haken/useVorgang'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { KIEinrichtungsverweis } from './KIEinrichtungsverweis'

import type { KIEinschaetzungAntwort } from '../antworteneinschaetzung'

interface KIEinschaetzungEigenschaften {
  karteID: string
}

// dieArten sind die zwei, die `domain` kennt.
//
// **Als Liste und nicht als zwei Knöpfe**: eine dritte Art käme sonst an drei
// Stellen dazu, und eine davon bliebe stehen.
const dieArten = ['regel', 'balance'] as const

export function KIEinschaetzung({ karteID }: KIEinschaetzungEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen
  const vorgang = useVorgang()

  const [offen, setzeOffen] = useState(false)
  const [gefragt, setzeGefragt] = useState(false)

  const katalog = useKIModelle(gefragt)

  const [art, setzeArt] = useState<string>(dieArten[0])
  const [gewaehlt, setzeGewaehlt] = useState('')
  const [frage, setzeFrage] = useState('')
  const [format, setzeFormat] = useState('')
  const [freigabe, setzeFreigabe] = useState(false)

  // **Leer, und das ist die Vorgabe** (A13.4, #1452): ein Vergleich sieht
  // keine einzige eigene Karte, solange niemand das Kreuz setzt.
  const [freigegeben, setzeFreigegeben] = useState<string[]>([])

  const [antwort, setzeAntwort] = useState<KIEinschaetzungAntwort>()
  const [fehler, setzeFehler] = useState<string>()

  const laeuft = vorgang.laeuft(`einschaetzung/${karteID}`)

  // ausloesen entscheidet, ob vorher nach einer Freigabe zu fragen ist.
  //
  // **Aus dem Katalog und nicht aus dem Namen** ([kostetGeld]) — dieselbe
  // Entscheidung wie beim Vorschlag, und aus demselben Grund: wo ein Anbieter
  // rechnet, sagt er selbst.
  function ausloesen() {
    if (kostetGeld(katalog.modelle, gewaehlt)) {
      setzeFreigabe(true)

      return
    }

    void beurteilen('')
  }

  async function beurteilen(nachweis: string) {
    setzeFehler(undefined)

    await vorgang.starte(`einschaetzung/${karteID}`, async () => {
      try {
        const gelesen = await api().KIEinschaetzen({
          art,
          karteID,
          anbieter: anbieterAus(gewaehlt),
          alias: aliasAus(gewaehlt),
          frage,

          // **Das Format nur beim Vergleich** (A13.2b, #1452): ein
          // Regelhinweis gilt nach den Regeln und nicht nach einem Format.
          format: art === 'balance' ? format : '',

          // **Die Freigabe nur beim Vergleich** (A13.4, #1452): ein
          // Regelhinweis sucht nicht in den eigenen Karten, und eine
          // Freigabe dafür wird abgewiesen.
          freigegeben: art === 'balance' ? freigegeben : [],
          nachweis,
        })

        setzeAntwort(auspacken(gelesen, () => gelesen))
      } catch (ursache) {
        setzeAntwort(undefined)
        setzeFehler(meldungAus(ursache, texte))
      }
    })
  }

  return (
    <details
      className="rounded-sm border border-kante px-3 py-2"
      onToggle={(ereignis) => {
        const auf = ereignis.currentTarget.open

        setzeOffen(auf)

        if (auf) {
          setzeGefragt(true)
        }
      }}
    >
      {/* **Zugeklappt, bis jemand ihn aufschlägt** — erst dann wird der
          Modellkatalog gelesen, und erst dann geht überhaupt etwas hinaus. */}
      <summary className="cursor-pointer text-sm font-medium focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring">
        {eigene.kiEinschaetzungTitel}
      </summary>

      {offen && (
        <div className="mt-3 flex flex-col gap-3">
          <Erklaerungsabsatz
            text={eigene.kiEinschaetzungErklaerung}
            klasse="max-w-prose text-sm text-leise"
          />

          <label className="flex flex-col gap-1 text-sm">
            {eigene.kiEinschaetzungArt}
            <select
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              value={art}
              onChange={(ereignis) => {
                setzeArt(ereignis.target.value)
              }}
            >
              <option value="regel">{eigene.kiEinschaetzungArtRegel}</option>
              <option value="balance">
                {eigene.kiEinschaetzungArtBalance}
              </option>
            </select>
          </label>

          {/* **Das Kreuz steht nur beim Vergleich da** — aus demselben
              Grund wie das Feld darunter: bei einer Regelfrage bedeutete es
              nichts, und `domain` weist eine Freigabe ohne Wirkung ab. */}
          {art === 'balance' && (
            <Bestandsfreigabe
              freigegeben={freigegeben}
              waehlen={(neu) => {
                setzeFreigegeben(neu)
                setzeAntwort(undefined)
              }}
            />
          )}

          {/* **Das Feld steht nur beim Vergleich da** — bei einer Regelfrage
              bedeutete es nichts, und ein Feld, das manchmal nichts bedeutet,
              beantwortet man irgendwann falsch (§6). */}
          {art === 'balance' && (
            <label className="flex flex-col gap-1 text-sm">
              {eigene.kiEinschaetzungFormat}
              <input
                className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                value={format}
                onChange={(ereignis) => {
                  setzeFormat(ereignis.target.value)
                }}
              />
            </label>
          )}

          <KIModellwahl
            gefragt={gefragt}
            modelle={katalog.modelle}
            laedt={katalog.laedt}
            fehler={katalog.fehler}
            gewaehlt={gewaehlt}
            waehlen={setzeGewaehlt}
          />

          <label className="flex flex-col gap-1 text-sm">
            {eigene.kiEinschaetzungFrage}
            <textarea
              className="rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
              rows={2}
              value={frage}
              onChange={(ereignis) => {
                setzeFrage(ereignis.target.value)
              }}
            />
          </label>

          <button
            type="button"
            className="self-start rounded-sm border border-kante px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60"
            disabled={laeuft || gewaehlt === '' || frage === ''}
            onClick={ausloesen}
          >
            {eigene.kiEinschaetzungFragen}
          </button>

          {freigabe && (
            <KIKostenfreigabe
              anbieter={anbieterAus(gewaehlt)}
              freigegeben={(nachweis) => {
                setzeFreigabe(false)
                void beurteilen(nachweis)
              }}
              verzichtet={() => {
                setzeFreigabe(false)
              }}
            />
          )}

          {laeuft && (
            <p role="status" className="text-sm text-leise">
              {eigene.kiEinschaetzungLaeuft}
            </p>
          )}

          {/* **Mit dem Weg zur Einrichtung** (AI-01.AK-04, #1742): eine

              KI-Meldung heißt meist, dass ein Anbieter oder Modell fehlt. */}

          {fehler !== undefined && (
            <div className="flex flex-col items-start gap-1">
              <p role="alert" className="text-sm text-fehler">
                {fehler}
              </p>
              <KIEinrichtungsverweis />
            </div>
          )}

          {antwort === undefined ? (
            <p className="text-sm text-leise">{eigene.kiEinschaetzungLeer}</p>
          ) : (
            <KIEinschaetzungsbelege antwort={antwort} />
          )}
        </div>
      )}
    </details>
  )
}

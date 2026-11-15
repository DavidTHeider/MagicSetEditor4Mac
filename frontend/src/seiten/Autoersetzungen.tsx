// **Die Autoersetzungsregeln eines Spiels** —
// #1341,
// Vermerk [0060](../../docs/entscheidungen/0060-eigene-autoersetzungen-liegen-im-bestand.md).
//
// # Was MSE2 hier tut
//
// Ein eigenes Fenster mit Liste, zwei Textfeldern, zwei Kästchen und drei
// Knöpfen (`src/gui/auto_replace_window.cpp`, Zeilen 157–210, Fassung f9d9356
// vom 2021-07-06). Hier steht ein **Kasten** auf der Kartenseite: §6 verbietet
// verschachtelte Dialoge, und die Regeln greifen beim Tippen in genau den
// Feldern, die daneben stehen.
//
// # Was übernommen ist
//
// **Dass eine Regel des Pakets sich nicht löschen lässt** und ihr Muster fest
// ist — MSEs `remove->Enable(ar && ar->custom)` und
// `match->Enable(ar && ar->custom)` (ebd., Zeilen 258 und 262). Was sich
// ändern lässt, ist ihr Ersatz, ihr Schalter und ihre Wortgrenze.
import { useState } from 'react'

import { Aktionsknopf } from '../bausteine/Aktionsknopf'
import { Entscheidung } from '../bausteine/Entscheidung'
import { Knopf } from '../bausteine/Knopf'
import { Pflichtfeld } from '../bausteine/Pflichtfeld'
import { useAutoersetzungen } from '../haken/useAutoersetzungen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import type { AutoersetzungsregelDTO } from '../antwortenautoersetzung'
import { useLesegrund } from '../haken/schreibrecht'

interface AutoersetzungenEigenschaften {
  setID: string
}

export function Autoersetzungen({ setID }: AutoersetzungenEigenschaften) {
  const texte = useTexte().seiten.karten.autoersetzung
  // **Ein Leser pflegt keine Regeln** (#2290): jede Zeile schreibt sofort.
  // **Mit Grund** (H01, #2255): jede gesperrte Eingabe sagt einem Leser,
  // warum — und die Knöpfe sperren über `schreibt`, wie jeder Knopf.
  const lesegrund = useLesegrund()
  const lesend = lesegrund.gesperrt
  const grund = lesegrund.beschreibe()
  const allgemein = useTexte().bausteine.aktionen

  const griff = useAutoersetzungen(setID)

  const [muster, setzeMuster] = useState('')
  const [ersatz, setzeErsatz] = useState('')

  // **Alle eigenen Regeln erst nach einer Frage** — UX-10.AK-04 (#1682):
  // keine kommt zurück. Eine einzelne löscht ohne Frage; sie ist in einer
  // Zeile wieder angelegt.
  const [fragt, setzeFragt] = useState(false)

  return (
    // **Hinter dem Menü „Mehr", wie die Skripte daneben** — seit UXR-05
    // (#1913); davor seit Werkstück C von Phase 4.10 hinter einem eigenen
    // Aufklapper. Ein Erklärabsatz, ein Schalter und zwei Felder standen mit
    // 200 Punkten über der Kartenliste; die Regeln stellt man je Spiel ein,
    // und danach greifen sie beim Tippen von selbst. **Den Kasten mit Titel
    // und Schließen trägt `Kartenwerkzeuge.tsx`**, hier steht sein Inhalt.
    <div className="flex flex-col gap-2">
      {lesegrund.satz}
      <Erklaerungsabsatz
        text={texte.erklaerung}
        klasse="max-w-3xl text-xs text-leise"
      />

      <label className="flex items-center gap-2 text-sm">
        <input
          type="checkbox"
          className={kaestchen}
          checked={griff.an}
          disabled={griff.laeuft || lesend}
          aria-describedby={grund}
          onChange={(ereignis) => {
            griff.schalten(ereignis.target.checked)
          }}
        />
        {texte.schalter}
      </label>

      <ul className="flex flex-col gap-1">
        {griff.regeln.map((regel) => (
          <Regelzeile
            key={regel.muster}
            regel={regel}
            angehalten={griff.laeuft || lesend}
            beschriebenVon={grund}
            setzen={griff.setzen}
            loeschen={griff.loeschen}
          />
        ))}
      </ul>

      <div className="flex flex-wrap items-end gap-2">
        {/* **Das Muster ist Pflicht, und das Feld sagt es** — Werkstück U3
            von Phase 5.1
            (#1414).
            `domain.PruefeAutoersetzung` weist ein leeres Muster ab („eine
            Regel ohne Muster greift nie"); bis dahin sagte das nur der
            ausgegraute Knopf, und wer ihn sieht, weiß nicht, **welches** der
            beiden Felder fehlt (§6).

            **Die Zeile steht mit `items-end`** — ein Hinweis unter dem Feld
            verschöbe sie sonst. So richtet sie sich an den Unterkanten aus,
            und der Satz wächst nach unten, statt die Nachbarn zu heben. */}
        <Pflichtfeld
          beschriftung={texte.muster}
          wert={muster}
          angehalten={griff.laeuft || lesend}
          beschriebenVon={grund}
          geaendert={setzeMuster}
        />

        <label className="flex flex-col gap-1 text-xs">
          {texte.ersatz}
          <input
            type="text"
            className={eingabe}
            value={ersatz}
            disabled={griff.laeuft || lesend}
            aria-describedby={grund}
            onChange={(ereignis) => {
              setzeErsatz(ereignis.target.value)
            }}
          />
        </label>

        {/* **Ohne Muster kein Knopf** — eine Regel ohne Muster greift nie,
            und der Dienst wiese sie ab. Ein Knopf, der nur eine Fehlermeldung
            auslöst, ist keine Hilfe (§6). */}
        <Aktionsknopf
          aktion="anlegen"
          gewicht="neben"
          bezeichnung={texte.anlegen}
          angehalten={griff.laeuft || muster.trim() === ''}
          schreibt
          aufKlick={() => {
            griff.setzen({ muster, ersatz, ganzesWort: true, aktiv: true })
            setzeMuster('')
            setzeErsatz('')
          }}
        >
          {texte.anlegen}
        </Aktionsknopf>

        <Knopf
          gewicht="neben"
          bezeichnung={texte.zuruecksetzen}
          angehalten={griff.laeuft || fragt}
          schreibt
          aufKlick={() => {
            setzeFragt(true)
          }}
        >
          {texte.zuruecksetzen}
        </Knopf>
      </div>

      {fragt && (
        <Entscheidung
          frage={texte.zuruecksetzenFrage}
          erklaerung={texte.zuruecksetzenErklaerung}
          bestaetigen={texte.zuruecksetzenJa}
          abbrechen={allgemein.abbrechen}
          laeuft={griff.laeuft || lesend}
          aufBestaetigen={() => {
            setzeFragt(false)
            griff.zuruecksetzen()
          }}
          aufAbbrechen={() => {
            setzeFragt(false)
          }}
        />
      )}

      {griff.fehler !== undefined && (
        <p className="text-sm text-hinweis">{griff.fehler}</p>
      )}
    </div>
  )
}

// kaestchen sind die Klassen eines Kästchens — der sichtbare Fokusring (§6).
//
// **Er steht am Element und nicht an der Beschriftung**: der Wächter in
// `barrierefreiheit.test.tsx` prüft jedes fokussierbare Element, und ein Ring
// an der Zeile darüber wäre für die Tastatur die halbe Auskunft.
const kaestchen =
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring '

// eingabe sind die gemeinsamen Klassen — dieselben wie an einem Kartenfeld.
const eingabe =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring disabled:opacity-60 '

interface RegelzeileEigenschaften {
  regel: AutoersetzungsregelDTO
  angehalten: boolean
  beschriebenVon: string | undefined
  setzen: (regel: Omit<AutoersetzungsregelDTO, 'eigen'>) => void
  loeschen: (muster: string) => void
}

// Regelzeile zeigt eine Regel — und was sich an ihr ändern lässt.
function Regelzeile({
  regel,
  angehalten,
  beschriebenVon,
  setzen,
  loeschen,
}: RegelzeileEigenschaften) {
  const texte = useTexte().seiten.karten.autoersetzung

  return (
    <li className="flex flex-wrap items-center gap-2 text-sm">
      <label className="flex items-center gap-1">
        <input
          type="checkbox"
          className={kaestchen}
          checked={regel.aktiv}
          disabled={angehalten}
          aria-describedby={beschriebenVon}
          onChange={(ereignis) => {
            setzen({ ...regel, aktiv: ereignis.target.checked })
          }}
        />
        <span className="sr-only">{texte.aktiv}</span>
      </label>

      <code className="rounded-sm bg-betont px-1">{regel.muster}</code>

      <span aria-hidden="true">→</span>

      <input
        type="text"
        className={eingabe + ' flex-1 basis-32'}
        aria-label={`${texte.ersatz} — ${regel.muster}`}
        defaultValue={regel.ersatz}
        disabled={angehalten}
        aria-describedby={beschriebenVon}
        onBlur={(ereignis) => {
          if (ereignis.target.value !== regel.ersatz) {
            setzen({ ...regel, ersatz: ereignis.target.value })
          }
        }}
      />

      <label className="flex items-center gap-1 text-xs">
        <input
          type="checkbox"
          className={kaestchen}
          checked={regel.ganzesWort}
          disabled={angehalten}
          aria-describedby={beschriebenVon}
          onChange={(ereignis) => {
            setzen({ ...regel, ganzesWort: ereignis.target.checked })
          }}
        />
        {texte.ganzesWort}
      </label>

      {/* **Nur eine eigene lässt sich löschen** — eine des Pakets gehört dem
          Paket; was geht, ist sie abzuschalten. */}
      {regel.eigen ? (
        <Aktionsknopf
          aktion="loeschen"
          gewicht="neben"
          bezeichnung={`${texte.loeschen} — ${regel.muster}`}
          angehalten={angehalten}
          aufKlick={() => {
            loeschen(regel.muster)
          }}
        >
          {texte.loeschen}
        </Aktionsknopf>
      ) : (
        <span className="text-xs text-leise">{texte.ausDemPaket}</span>
      )}
    </li>
  )
}

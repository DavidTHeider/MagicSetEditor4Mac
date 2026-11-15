// **Eine Zeile, die in eine Spalte kommt** — TRADE-02b
// (#1951),
// herausgelöst aus `Handelspositionen.tsx`.
//
// **Die Seite steht nicht mehr darin**: die Spalte, deren Knopf das
// Formular öffnete, sagt sie (AK-01). Eine Auswahl „Von wem" war eine zweite
// Stelle für dieselbe Angabe.
//
// **Druck und Menge sind offen, die übrigen drei nicht** — UX-06.AK-01
// (#1678): geschlossene Wertemengen werden gewählt, offene bekommen
// Vorschläge und bleiben Freitext.
//
// **Der Druck kommt über seinen Namen** — P10
// (#2378),
// Befund C-13: dieselbe Suche wie beim Aufnehmen eines Exemplars (AP-10),
// der Fokus steht beim Öffnen darin. **Die Kennung steht im Kasten** „Kennung
// von Hand" — für den, der sie zur Hand hat, und für eigene Drucke, die
// Scryfall nicht kennt.
import { useEffect, useId, useRef, useState } from 'react'

import { drucksprachen } from '../drucksprachen'
import type { Listentreffer } from '../antwortenlisteneinfuhr'
import { Aktionsknopf } from './Aktionsknopf'
import { Aufklappkasten } from './Aufklappkasten'
import { Drucksuche } from './Drucksuche'
import { Ganzzahlfeld, ganzzahlImFeld } from './Ganzzahlfeld'
import { Knopf } from './Knopf'
import { Vorschlagsfeld } from './Vorschlagsfeld'
import { zustaende } from './Zustandswahl'
import { useTexte } from '../sprache/kontext'
import { kennungseingabe } from './kennungseingabe'

// dieLeereZeile ist der Anfangszustand des Formulars.
//
// **`nearMint` und nicht `NM`** — UX-06.AK-01 (#1678). Verbucht wird eine
// Position gegen die sieben Zustände der Sammlung (`domain.IstZustand`), und
// `NM` ist keiner: eine Position mit der alten Vorgabe scheiterte erst am
// Ende eines Handels.
export const dieLeereZeile = {
  printing: '',
  menge: '1',
  guete: 'nearMint',
  sprache: 'de',
  finish: 'nonfoil',
}

export type Formularzeile = typeof dieLeereZeile

// ausfuehrungen sind die drei aus `domain.Finish` — dieselbe Reihenfolge wie
// in der Sammlung (`Ausfuehrungswahl`).
const ausfuehrungen = ['nonfoil', 'foil', 'etched'] as const

// feldstil ist der Rahmen eines Eingabefeldes — **mit sichtbarem Fokusring**
// (§6).
const feldstil =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

// schmalstil ist derselbe Rahmen für die kurzen Felder.
//
// **Ausgeschrieben und nicht zusammengesetzt**: der Wächter über die
// Fokusringe liest die Klassen am Element
// (`internal/werkzeuge/fokusring_test.go`).
const schmalstil =
  'w-28 rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

interface Eigenschaften {
  // vorlage füllt das Formular vor — etwa mit einer Karte aus der Sammlung.
  vorlage: Formularzeile

  uebernehmen: (zeile: Formularzeile) => void
  verwerfen: () => void
}

export function Handelspositionsform({
  vorlage,
  uebernehmen,
  verwerfen,
}: Eigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const sammlung = texte.seiten.sammlung

  const [zeile, setzeZeile] = useState(vorlage)
  const [pruefen, setzePruefen] = useState(false)

  // gewaehlt ist der Druck, den die Suche gebracht hat — hebt sie ihre Wahl
  // auf, geht er wieder aus der Zeile, **es sei denn, jemand hat die
  // Kennung danach von Hand geändert** (wie `useDruckwahl`).
  const [gewaehlt, setzeGewaehlt] = useState<string>()

  // **Der Fokus beim Öffnen steht in der Suche** — wer eine Zeile anlegt,
  // tippt zuerst den Namen. Bringt die Vorlage schon eine Kennung mit (aus
  // der Sammlung), steht der Kasten offen und behält ihn.
  const rahmen = useRef<HTMLDivElement | null>(null)
  const mitKennung = vorlage.printing !== ''
  useEffect(() => {
    if (!mitKennung) {
      rahmen.current?.querySelector<HTMLElement>('[data-autofocus]')?.focus()
    }
  }, [mitKennung])

  const printingFeld = useId()
  const gueteFeld = useId()
  const finishFeld = useId()

  const sprachvorschlaege = drucksprachen.map((code) => ({
    wert: code,
    bezeichnung: texte.bausteine.drucksprachen.namen[code],
  }))

  const setze = (teil: Partial<Formularzeile>) => {
    setzeZeile((vorher) => ({ ...vorher, ...teil }))
  }

  const waehle = (treffer: Listentreffer | undefined) => {
    if (treffer !== undefined) {
      setzeGewaehlt(treffer.printing)
      setze({ printing: treffer.printing })

      return
    }

    setzeGewaehlt(undefined)
    setzeZeile((vorher) =>
      vorher.printing === gewaehlt ? { ...vorher, printing: '' } : vorher,
    )
  }

  return (
    <div ref={rahmen} className="flex flex-col gap-2">
      <Drucksuche waehle={waehle} laeuft={false} />

      <div className="flex flex-wrap items-end gap-2">
        <Ganzzahlfeld
          beschriftung={eigene.positionMenge}
          wert={zeile.menge}
          setzeWert={(wert) => {
            setze({ menge: wert })
          }}
          pruefen={pruefen}
          kleinstens={1}
        />

        <label className="flex flex-col gap-1 text-sm" htmlFor={gueteFeld}>
          {eigene.positionGuete}
          <select
            id={gueteFeld}
            className={feldstil}
            value={zeile.guete}
            onChange={(ereignis) => {
              setze({ guete: ereignis.target.value })
            }}
          >
            {zustaende.map((zustand) => (
              <option key={zustand} value={zustand}>
                {sammlung.zustaende[zustand]}
              </option>
            ))}
          </select>
        </label>

        <Vorschlagsfeld
          beschriftung={eigene.positionSprache}
          wert={zeile.sprache}
          setzeWert={(wert) => {
            setze({ sprache: wert })
          }}
          vorschlaege={sprachvorschlaege}
          beispiel={texte.bausteine.drucksprachen.beispiel}
          breite={5}
        />

        <label className="flex flex-col gap-1 text-sm" htmlFor={finishFeld}>
          {eigene.positionFinish}
          <select
            id={finishFeld}
            className={feldstil}
            value={zeile.finish}
            onChange={(ereignis) => {
              setze({ finish: ereignis.target.value })
            }}
          >
            {ausfuehrungen.map((eine) => (
              <option key={eine} value={eine}>
                {sammlung[eine]}
              </option>
            ))}
          </select>
        </label>
      </div>

      <Aufklappkasten
        titel={eigene.kennungVonHand}
        aufschlagen={mitKennung ? vorlage.printing : undefined}
      >
        <label className="flex flex-col gap-1 text-sm" htmlFor={printingFeld}>
          {eigene.positionPrinting}
          <input
            {...kennungseingabe}
            id={printingFeld}
            type="text"
            className={schmalstil}
            value={zeile.printing}
            onChange={(ereignis) => {
              setze({ printing: ereignis.target.value })
            }}
          />
        </label>
      </Aufklappkasten>

      <div className="flex flex-wrap items-end gap-2">
        {/* **Ohne Prüfung auf Vollständigkeit**: was eine Position ist,
          entscheidet der Vertrag, und er sagt es mit einem Satz, der das
          Feld nennt (§6). **Nur die Menge wird gelesen** (F03, #2245): eine
          Kommazahl käme sonst an ein `int` und als Satz über JSON zurück. */}
        <Aktionsknopf
          aktion="anlegen"
          gewicht="leise"
          angehalten={zeile.printing.trim() === ''}
          aufKlick={() => {
            setzePruefen(true)

            const menge = ganzzahlImFeld(zeile.menge, false, 1)

            if (menge !== undefined) {
              uebernehmen({ ...zeile, menge: String(menge) })
            }
          }}
        >
          {eigene.positionHinzufuegen}
        </Aktionsknopf>

        <Knopf gewicht="leise" aufKlick={verwerfen}>
          {eigene.assistent.zeileVerwerfen}
        </Knopf>
      </div>
    </div>
  )
}

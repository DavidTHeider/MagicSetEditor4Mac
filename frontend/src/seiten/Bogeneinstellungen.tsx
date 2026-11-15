// Die Einstellungen des Druckbogens: Papier, Anschnitt, Auflösung, Marken.
//
// **Was hier nicht steht, ist das Kartenmaß.** 63 × 88 mm ist die Zusage aus
// #50 und keine Einstellung: eine Karte, die anders herauskommt, passt in keine
// Hülle. Wer ein anderes Spiel mit anderen Karten druckt, braucht dafür eine
// Entscheidung und kein Eingabefeld.
import { useId } from 'react'

import { Zahlenfeld } from '../bausteine/Zahlenfeld'
import { useTexte } from '../sprache/kontext'

interface BogeneinstellungenEigenschaften {
  papier: string
  anschnitt: number
  aufloesung: number
  marken: boolean
  laeuft: boolean

  setzePapier: (papier: string) => void
  setzeAnschnitt: (anschnitt: number) => void
  setzeAufloesung: (aufloesung: number) => void
  setzeMarken: (marken: boolean) => void
}

export function Bogeneinstellungen({
  papier,
  anschnitt,
  aufloesung,
  marken,
  laeuft,
  setzePapier,
  setzeAnschnitt,
  setzeAufloesung,
  setzeMarken,
}: BogeneinstellungenEigenschaften) {
  const texte = useTexte()
  const anschnitthilfe = useId()

  const feld =
    'mt-1 w-full rounded-sm border border-kante bg-transparent px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-60'

  return (
    <div className="mt-4 space-y-4">
      <div className="grid max-w-2xl gap-4 sm:grid-cols-3">
        <label className="block">
          <span className="text-sm">{texte.seiten.druckbogen.papier}</span>
          <select
            value={papier}
            disabled={laeuft}
            onChange={(ereignis) => {
              setzePapier(ereignis.target.value)
            }}
            className={feld}
          >
            <option value="a4">{texte.seiten.druckbogen.papierA4}</option>
            <option value="letter">
              {texte.seiten.druckbogen.papierLetter}
            </option>
          </select>
        </label>

        {/* **Dasselbe Zahlenfeld wie bei der Bildbreite** (Werkstück F von
            Phase 4.10): `min` und `max` allein melden nichts. */}
        <Zahlenfeld
          beschriftung={texte.seiten.druckbogen.anschnitt}
          wert={anschnitt}
          kleinstens={0}
          groesstens={20}
          schritt={0.5}
          angehalten={laeuft}
          beschreibung={anschnitthilfe}
          geaendert={setzeAnschnitt}
        />

        <label className="block">
          <span className="text-sm">{texte.seiten.druckbogen.aufloesung}</span>
          <select
            value={String(aufloesung)}
            disabled={laeuft}
            onChange={(ereignis) => {
              setzeAufloesung(Number(ereignis.target.value))
            }}
            className={feld}
          >
            <option value="150">{texte.seiten.druckbogen.dpi150}</option>
            <option value="300">{texte.seiten.druckbogen.dpi300}</option>
            <option value="600">{texte.seiten.druckbogen.dpi600}</option>
          </select>
        </label>
      </div>

      <label className="flex items-center gap-2">
        <input
          type="checkbox"
          checked={marken}
          disabled={laeuft}
          onChange={(ereignis) => {
            setzeMarken(ereignis.target.checked)
          }}
          className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
        />
        <span className="text-sm">{texte.seiten.druckbogen.schnittmarken}</span>
      </label>

      <p id={anschnitthilfe} className="max-w-prose text-sm text-leise">
        {texte.seiten.druckbogen.anschnittHilfe}
      </p>
    </div>
  )
}

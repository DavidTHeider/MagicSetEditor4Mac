// **Die Bedienelemente der Sammlungssicht** —
// #1200.
//
// **Ein eigener Baustein**, weil `Sammlung.tsx` mit ihm über die Zeilengrenze
// ginge (§2). Der Schnitt folgt der Frage: dort steht, **was** die Seite tut,
// hier, **was** die Sicht ausschneidet — Suche und Ausführung.
//
// **Wie die Liste aussieht** — Ansicht, Sortierung, Gruppierung — steht seit
// UXR-10 (#1918) über der Liste (`Sammlungsleiste`), wie im Artboard
// `Sammlung`. Vorher standen hier fünf Felder in 240 Punkten untereinander.
import { useId } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { einsetzen } from '../sprache/einsetzen'
import type { Ausfuehrung, Sicht } from '../seiten/sammlungssicht'

interface SammlungssichtEigenschaften {
  sicht: Sicht
  setze: (teil: Partial<Sicht>) => void

  // gezeigt und gesamt sagen, wie viel die Sicht wegnimmt.
  //
  // **Beides und nicht nur eines**: „12" allein verrät nicht, dass 300
  // ausgeblendet sind — und eine Sicht, die stillschweigend filtert, sieht aus
  // wie eine kurze Sammlung.
  gezeigt: number
  gesamt: number

  // zuruecksetzen nimmt Suche und Filter zurück — DOM-03.AK-03 (#1698).
  zuruecksetzen: () => void
}

export function Sammlungssicht({
  sicht,
  setze,
  gezeigt,
  gesamt,
  zuruecksetzen,
}: SammlungssichtEigenschaften) {
  const eigene = useTexte().seiten.sammlung

  const sucheID = useId()
  const ausfuehrungID = useId()

  return (
    <section
      aria-label={eigene.sicht}
      className="flex flex-wrap items-end gap-3 rounded-sm border border-kante p-3"
    >
      {/* **Die Beschriftung darf schrumpfen, das Feld füllt sie** — `size`
          ist ein Wunsch und keine Grenze. WebKit bemisst ihn mit der
          mittleren Zeichenbreite der Anwendungsschrift, und im Navigator
          (240 Punkte) ragte das Feld 82 Punkte in die Liste daneben;
          gemessen mit der WebKit-Probe (#1828), jsdom hat kein Layout. */}
      <label className="flex min-w-0 max-w-full flex-col gap-1">
        <span className="text-sm font-medium" id={sucheID}>
          {eigene.suche}
        </span>
        <input
          type="search"
          aria-labelledby={sucheID}
          value={sicht.suche}
          placeholder={eigene.suchePlatzhalter}
          onChange={(ereignis) => {
            setze({ suche: ereignis.target.value })
          }}
          className="w-full rounded-sm border border-kante bg-transparent px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          size={28}
        />
      </label>

      {/* **Die Ausführung bleibt eine Auswahlliste** — UXR-10 (#1918) hat
          sie als vier Segmente gebaut (R7: Segmente bis vier Werte) und
          wieder zurückgenommen: im Navigator von 240 Punkten brachen sie um,
          und mit einem echten Bestand standen dann 24 Bedienelemente im
          Bild statt 20 (R2). Die Ansicht über der Liste hat den Platz. */}
      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium" id={ausfuehrungID}>
          {eigene.ausfuehrung}
        </span>
        <select
          aria-labelledby={ausfuehrungID}
          value={sicht.ausfuehrung}
          onChange={(ereignis) => {
            setze({ ausfuehrung: ereignis.target.value as Ausfuehrung })
          }}
          className={wahlfeld}
        >
          <option value="alle">{eigene.ausfuehrungAlle}</option>
          <option value="foil">{eigene.nurFoil}</option>
          <option value="nonfoil">{eigene.nurNonfoil}</option>
          <option value="etched">{eigene.nurEtched}</option>
        </select>
      </label>

      {/* **Wie viel die Sicht wegnimmt, steht dabei** — eine Liste, die
          stillschweigend filtert, sieht aus wie eine kurze Sammlung. **Und
          als `status`** (UX-10.AK-02, #1682), wie in der Paketliste: wer
          tippt, hört die Zahl, ohne die Liste durchzugehen. */}
      <p role="status" className="text-sm text-leise">
        {einsetzen(eigene.gezeigt, {
          gezeigt: String(gezeigt),
          gesamt: String(gesamt),
        })}
      </p>
      {/* **Was die Sicht wegnimmt, nimmt ein Knopf zurück** — auch mit
          Treffern (DOM-03.AK-03, #1698). Ohne Treffer steht derselbe Knopf
          im leeren Zustand der Seite; hier stünde er dann doppelt. */}
      {gezeigt > 0 && gezeigt < gesamt && (
        <Knopf gewicht="leise" aufKlick={zuruecksetzen}>
          {eigene.ohneTrefferZuruecksetzen}
        </Knopf>
      )}
    </section>
  )
}

const wahlfeld =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

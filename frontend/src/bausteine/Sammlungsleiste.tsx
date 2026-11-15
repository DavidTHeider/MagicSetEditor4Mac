// **Die Leiste über der Liste** — UXR-10
// (#1918),
// nach dem Artboard `Sammlung`: der Ansichtswechsel steht über der Liste,
// nicht im Navigator.
//
// # Warum Segmente und keine Auswahlliste
//
// **R7 in `docs/UIREVIEW.md`**: Segmente bis vier Werte. Drei Ansichten in
// einer Auswahlliste verstecken zwei davon hinter einem Klick, und Sets und
// Decks tragen denselben Wechsel längst als Knopfgruppe (`Listenansicht`).
//
// # Warum Sortierung und Gruppierung mitkommen
//
// **Weil sie dieselbe Frage beantworten**: wie die Liste aussieht. Was sie
// **ausschneidet** — Suche, Ausführung, Marken —, bleibt im Navigator. Die
// Sortierung bleibt eine Auswahlliste wie in `Listenansicht`: zwei
// Segmentgruppen nebeneinander läsen sich als eine Reihe von sechs
// gleichrangigen Knöpfen. Die Gruppierung hat fünf Werte und gilt nur in der
// Gruppenansicht; sonst ruht sie, statt zu verschwinden — wer sie sucht,
// findet sie neben dem Knopf, der sie weckt.
import { useId } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import type {
  Ansicht,
  Gruppierung,
  Sicht,
  Sortierung,
} from '../seiten/sammlungssicht'

interface LeistenEigenschaften {
  sicht: Sicht
  setze: (teil: Partial<Sicht>) => void
}

const ansichten: readonly Ansicht[] = ['tabelle', 'galerie', 'gruppen']

export function Sammlungsleiste({ sicht, setze }: LeistenEigenschaften) {
  const eigene = useTexte().seiten.sammlung

  const sortierID = useId()
  const gruppierungID = useId()

  const namen: Record<Ansicht, string> = {
    tabelle: eigene.ansichtTabelle,
    galerie: eigene.ansichtGalerie,
    gruppen: eigene.ansichtGruppen,
  }

  return (
    <div className="flex flex-wrap items-end gap-3">
      <div role="group" aria-label={eigene.ansicht} className="flex gap-1">
        {ansichten.map((form) => (
          <Knopf
            key={form}
            klein
            gedrueckt={sicht.ansicht === form}
            aufKlick={() => {
              // **Ein zweiter Druck ändert nichts** — eine neue Sicht räumt
              // die Haken (`useSammlung`).
              if (form !== sicht.ansicht) {
                setze({ ansicht: form })
              }
            }}
          >
            {namen[form]}
          </Knopf>
        ))}
      </div>

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium" id={sortierID}>
          {eigene.sortierung}
        </span>
        <select
          aria-labelledby={sortierID}
          value={sicht.sortierung}
          onChange={(ereignis) => {
            setze({ sortierung: ereignis.target.value as Sortierung })
          }}
          className={wahlfeld}
        >
          <option value="bezeichnung">{eigene.sortierBezeichnung}</option>
          <option value="menge">{eigene.sortierMenge}</option>
          <option value="erwerbsdatum">{eigene.sortierErwerbsdatum}</option>
        </select>
      </label>

      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium" id={gruppierungID}>
          {eigene.gruppierung}
        </span>
        <select
          aria-labelledby={gruppierungID}
          value={sicht.gruppierung}
          disabled={sicht.ansicht !== 'gruppen'}
          onChange={(ereignis) => {
            setze({ gruppierung: ereignis.target.value as Gruppierung })
          }}
          className={wahlfeld}
        >
          <option value="keine">{eigene.gruppierungKeine}</option>
          <option value="set">{eigene.gruppierungSet}</option>
          <option value="lagerort">{eigene.gruppierungLagerort}</option>
          <option value="zustand">{eigene.gruppierungZustand}</option>
          <option value="sprache">{eigene.gruppierungSprache}</option>
        </select>
      </label>
    </div>
  )
}

const wahlfeld =
  'rounded-sm border border-kante bg-transparent px-2 py-1 text-sm ' +
  'focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 ' +
  'focus-visible:outline-ring'

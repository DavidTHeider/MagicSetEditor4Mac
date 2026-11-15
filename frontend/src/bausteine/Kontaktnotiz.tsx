// **Was man sich zu einem Gegenüber notiert** — Werkstück T3 von Phase 5.6
// (#1468).
//
// # Warum ein eigener Baustein
//
// **Dieselbe Begründung wie bei `Kontaktadresse.tsx`**: `Peerkontaktzeile.tsx`
// ginge mit ihm über die Zeilengrenze aus §2, und er beantwortet eine andere
// Frage — dort steht, **was** ein Gegenüber darf, hier, was man **über** es
// denkt.
//
// # Und warum der Satz darunter steht
//
// **Weil eine Notiz über einen Menschen aussieht wie etwas, das ihn
// erreicht.** Sie tut es nicht: keine Sichtkarte trägt sie, keine Freigabe,
// keine Leitung. Das steht dabei, statt es offen zu lassen.
import { useId, useState } from 'react'

import { Knopf } from './Knopf'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

interface Kontaktnotizeigenschaften {
  kennung: string
  notiz: string
  laeuft: boolean
  setzen: (kennung: string, notiz: string) => void
}

export function Kontaktnotiz({
  kennung,
  notiz,
  laeuft,
  setzen,
}: Kontaktnotizeigenschaften) {
  const eigene = useTexte().seiten.technik
  const feld = useId()

  // **Geräumt über den `key` des Aufrufers** (§7, Regel 1) — derselbe Weg
  // wie bei der Adresse, und aus demselben Grund: nach einem geglückten
  // Schreiben kommt die ganze Liste zurück.
  const [entwurf, setzeEntwurf] = useState(notiz)

  return (
    <div className="flex flex-wrap items-end gap-2">
      <label className="flex grow flex-col gap-1 text-sm" htmlFor={feld}>
        {eigene.kontakteNotiz}
        <input
          id={feld}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          placeholder={eigene.kontakteNotizBeispiel}
          value={entwurf}
          onChange={(ereignis) => {
            setzeEntwurf(ereignis.target.value)
          }}
        />
      </label>

      <Erklaerungsabsatz
        text={eigene.kontakteNotizHinweis}
        klasse="w-full text-sm text-leise"
      />

      <Knopf
        gewicht="leise"
        angehalten={laeuft || entwurf === notiz}
        aufKlick={() => {
          setzen(kennung, entwurf)
        }}
      >
        {eigene.kontakteNotizMerken}
      </Knopf>
    </div>
  )
}

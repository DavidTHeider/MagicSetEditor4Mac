// **Der Kampf in Zeilen** — Artboard *Tisch*: „Aschenwächter → Hallenwart",
// „drei weitere ungeblockt → Bo", unter der Überschrift „Kampf".
//
// # Warum im Zugkasten und nicht neben dem Stapel
//
// **Weil die Leinwand keine Zeilen trägt, die man vorlesen kann.** Das
// Artboard setzt den Kasten neben den Stapel, also **in** die Tischfläche;
// hier ist sie eine WebGL-Leinwand, und Text darauf erreicht kein
// Vorleseprogramm. Der Zugkasten steht in derselben Blickhöhe, und dort
// fragt man ohnehin, wo der Zug steht.
//
// # Wann zusammengefasst wird
//
// **Ab drei ungeblockten gegen dasselbe Ziel** — darunter lesen sich Namen
// schneller als eine Zahl, darüber füllten sie die Spalte. „Weitere" steht
// nur, wo geblockte darüber stehen; sonst gibt es nichts, wozu sie weitere
// wären.
import { useId } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { kampfzeilen } from '../spiel/kampfzeilen'
import { useSprache, useTexte } from '../sprache/kontext'

import type { Partie } from '../apipartie'

export function Kampfliste({ partie }: { partie: Partie }) {
  const eigene = useTexte().seiten.spielen
  const sprache = useSprache()
  const titel = useId()

  const zeilen = kampfzeilen(partie)

  if (zeilen.length === 0) {
    return undefined
  }

  // **Mehrere Blocker in der Sprache des Fensters aufgezählt** — „und" hier,
  // „and" dort, wie beim Namen einer Partie.
  const liste = new Intl.ListFormat(sprache, { type: 'conjunction' })
  const geblockte = zeilen.some((zeile) => zeile.art === 'geblockt')

  const texte = zeilen.flatMap((zeile) => {
    if (zeile.art === 'geblockt') {
      return [
        einsetzen(eigene.kampfGegen, {
          wer: zeile.angreifer,
          gegen: liste.format(zeile.blocker),
        }),
      ]
    }

    if (zeile.angreifer.length < 3) {
      return zeile.angreifer.map((wer) =>
        einsetzen(eigene.kampfGegen, { wer, gegen: zeile.ziel }),
      )
    }

    return [
      einsetzen(
        geblockte ? eigene.kampfWeitereUngeblockt : eigene.kampfUngeblockt,
        { zahl: String(zeile.angreifer.length), ziel: zeile.ziel },
      ),
    ]
  })

  return (
    <div className="flex flex-col gap-0.5">
      <span
        id={titel}
        className="text-xs font-medium tracking-wide text-leise uppercase"
      >
        {eigene.kampfTitel}
      </span>

      <ul aria-labelledby={titel} className="text-sm">
        {texte.map((text, stelle) => (
          // **Die Stelle gehört zum Schlüssel**: zwei gleichnamige Karten
          // greifen dasselbe Ziel an und geben dieselbe Zeile.
          <li key={`${String(stelle)}:${text}`}>{text}</li>
        ))}
      </ul>
    </div>
  )
}

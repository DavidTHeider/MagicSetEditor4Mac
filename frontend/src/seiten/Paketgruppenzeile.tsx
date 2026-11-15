// Die Gruppenzeile der Paketliste (#244) — herausgelöst aus
// `Paketzeilen.tsx`, das sonst über die Zeilengrenze aus §2 ginge. Der
// Schnitt folgt der Frage: dort die Pakete, hier die Gruppe darüber.
import type { KeyboardEvent } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { Baumzeile } from './paketbaum'
import { einzug, gruppenaufschrift } from './paketeinzug'

const zeilenklasse =
  'border-b border-kante bg-flaeche focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'
const schaltzeile = zeilenklasse + ' cursor-pointer'

interface Gruppeneigenschaften {
  zeile: Baumzeile
  offen: boolean
  schaltbar: boolean
  umschalten: () => void

  // tabstelle sagt, ob **diese** Zeile den Tabulator annimmt (#1290).
  tabstelle: boolean
  aufTaste: (ereignis: KeyboardEvent<HTMLElement>) => void
}

// **Eine Gruppe ohne Inhalt gibt es nicht**, und deshalb steht hier kein
// Zweig dafür. Ein Knoten entsteht in [paketbaum] nur, weil ein Paket ihn
// nennt; gefiltert wird **vor** dem Bauen (siehe [sicht]). Der erste Anlauf
// hatte den Zweig samt Test — und schaltete jede zugeklappte Gruppe ab, weil
// er ihre Kinder in der sichtbaren Liste suchte, wo sie gerade nicht stehen.
//
// **Die Gruppenzeile liegt auf `bg-flaeche` / `bg-slate-900`, nicht auf
// `bg-betont`.** Der erste Anlauf nahm die 100er — und sie ist im hellen
// Erscheinungsbild genau der Anwendungsgrund, auf dem die Paketseite steht
// (`Rahmen.tsx`, Zeile 162). Eine Hervorhebung in der Farbe des Untergrunds
// ist keine.
export function Gruppenzeile({
  zeile,
  offen,
  schaltbar,
  umschalten,
  tabstelle,
  aufTaste,
}: Gruppeneigenschaften) {
  const texte = useTexte()

  // **Der Pfeil ist Zier und wird nicht vorgelesen.** Was er sagt, sagt
  // `aria-expanded` schon — zweimal dasselbe zu hören ist schlechter als
  // einmal.
  const inhalt = (
    <>
      <span aria-hidden="true">{offen ? '▾' : '▸'}</span>
      <span>{zeile.aufschrift}</span>
    </>
  )

  return (
    <tr
      tabIndex={tabstelle ? 0 : -1}
      // **Die Zeile ist der Schalter, kein Knopf darin** — UXR-17 (#1925).
      // Ein Knopf je Gruppe zählte als eigenes Bedienelement, bei Hunderten
      // Gruppen (R6, `docs/UIREVIEW.md`); bedienbar war ohnehin die ganze
      // Breite. Klick, `Enter` und Leertaste bleiben dieselben, und
      // `aria-expanded` sagt am Gitter, was der Knopf sagte.
      aria-label={einsetzen(texte.seiten.pakete.katalogGruppeAn, {
        name: zeile.aufschrift,
      })}
      aria-expanded={schaltbar ? offen : undefined}
      aria-level={zeile.stufe + 1}
      onClick={schaltbar ? umschalten : undefined}
      onKeyDown={(ereignis) => {
        // **Eingabe und Leertaste schalten die Gruppe um** — auf der Zeile
        // selbst und nicht erst hinter einem →. Ein Gitter gibt sie
        // ohnehin: `Enter` bedient die Zeile.
        if (schaltbar && (ereignis.key === 'Enter' || ereignis.key === ' ')) {
          ereignis.preventDefault()
          umschalten()

          return
        }

        aufTaste(ereignis)
      }}
      className={schaltbar ? schaltzeile : zeilenklasse}
    >
      <th scope="colgroup" colSpan={6} className="py-1 text-left">
        <span className={gruppenaufschrift} style={einzug(zeile.stufe)}>
          {inhalt}
        </span>
      </th>
    </tr>
  )
}

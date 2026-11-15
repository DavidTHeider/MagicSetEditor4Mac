// **Die Karte einer Handelszeile** — P10
// (#2378),
// Befund C-13, Entwurf `Markttisch09`: Name, Set, Nummer und Zustand, links
// das Bild.
//
// # Woher das Bild kommt, und wann keines
//
// **Aus dem, was Scryfall schon einmal beantwortet hat** — gefragt wird
// dafür nicht (§16, Regel 5). **Ein eigener Druck hat nie eines**: er
// bekommt einen gestrichelten Platzhalter, und das Bild einer anderen Karte
// an seiner Stelle wäre eine falsche Auskunft.
//
// # Und warum das Bild für ein Vorleseprogramm nicht da ist
//
// **Weil der Name daneben steht** — wie bei `Kartenbild`: ein `alt` mit
// demselben Text sagte ihn zweimal (WCAG 1.1.1).
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Handelsdruck, Handelsposition } from '../apihandelsleitung'

interface Eigenschaften {
  position: Handelsposition

  // druck ist `undefined`, solange der Anhalt nicht da ist.
  druck: Handelsdruck | undefined

  // merkmale sind Güte, Sprache und Ausführung als Wörter.
  merkmale: string
}

export function Handelsdruckzeile({
  position,
  druck,
  merkmale,
}: Eigenschaften) {
  const texte = useTexte().seiten.handeln.druck

  const unbekannt = druck?.name === '' && !druck.eigen

  const unterzeile = [
    druck?.eigen === true ? texte.eigen : '',
    bezugVon(druck, texte.bezug),
    // **Wo nichts bekannt ist, bleibt die Kennung** — sie sagt niemandem
    // viel, aber sie ist eine Auskunft und keine erfundene.
    unbekannt ? position.printing : '',
    merkmale,
  ]
    .filter((teil) => teil !== '')
    .join(' · ')

  return (
    <span className="flex min-w-0 items-center gap-2">
      {druck !== undefined && druck.bild !== '' ? (
        <img
          src={druck.bild}
          alt=""
          loading="lazy"
          className="h-14 w-10 shrink-0 rounded-sm object-contain"
        />
      ) : (
        <span aria-hidden="true" className={platzhalterklassen(druck)} />
      )}

      <span className="flex min-w-0 flex-col">
        <span>
          {position.menge} × {nameVon(druck, texte)}
        </span>
        {unterzeile !== '' && (
          <span className="text-xs text-leise">{unterzeile}</span>
        )}
      </span>
    </span>
  )
}

// nameVon nennt die Karte — oder sagt, warum sie keinen Namen hat.
function nameVon(
  druck: Handelsdruck | undefined,
  texte: { ohneNamen: string; unbekannt: string; laedt: string },
): string {
  if (druck === undefined) {
    return texte.laedt
  }

  if (druck.name !== '') {
    return druck.name
  }

  return druck.eigen ? texte.ohneNamen : texte.unbekannt
}

// bezugVon gibt Set und Nummer — `M10 #146`, oder nur das Set.
function bezugVon(druck: Handelsdruck | undefined, vorlage: string): string {
  if (druck === undefined || druck.setcode === '') {
    return ''
  }

  const code = druck.setcode.toUpperCase()

  return druck.nummer === ''
    ? code
    : einsetzen(vorlage, { code, nummer: druck.nummer })
}

// platzhalterklassen unterscheidet den eigenen Druck vom unbekannten.
function platzhalterklassen(druck: Handelsdruck | undefined): string {
  return (
    'h-14 w-10 shrink-0 rounded-sm ' +
    (druck?.eigen === true ? 'border border-dashed border-kante' : 'bg-betont')
  )
}

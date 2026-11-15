// **Wer wen angreift und wer wen blockt** — CR 508.1a, CR 509.1a, Werkstück
// zu #1623.
//
// # Warum eine Zeile je Kreatur und kein Textfeld
//
// **Weil ein Kampf aus Paaren besteht** und nicht aus einer Aufzählung: der
// Kern nimmt `{kreatur, ziel}` entgegen (CR 508.1a). Eine Zeile je eigener
// Kreatur mit einer Auswahl daneben ist dieselbe Gestalt — und sie zeigt
// nebenbei, **welche** Kreaturen überhaupt in Frage kommen.
//
// # Was sie nicht prüft
//
// **Ob diese Kreatur angreifen darf.** Einsatzverzögerung (CR 302.6),
// getappt (CR 508.1a) und jede Einschränkung aus einer Fähigkeit
// (CR 508.1c) entscheidet der Kern; das Fenster bietet an, und der Kern
// weist ab. Wer hier prüfte, hätte die Regel an zwei Stellen.
//
// # Was sie doch auswählt
//
// **Nur Kreaturen** (CR 508.1a, CR 509.1a, Fassung 2026-09-25, #2390): das
// ist keine Ermessensfrage des Kerns, sondern was eine Zeile hier überhaupt
// ist. Hier standen alle eigenen Bleibenden, und ein Land war die erste
// Zeile. **Wo die Eigenschaften verborgen sind** (`null`), steht die Zeile
// weiter da — der Kern weiß, was dort liegt. Liefert der Kern einmal die
// Objektaktionen selbst (P3, #2371), kommt die Auswahl von dort.
import { kartenwort, zielVon } from '../spiel/feldwerte'
import { useTexte } from '../sprache/kontext'

import type { Partie } from '../apipartie'

const eingabeklassen =
  'rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Kampffeld({
  art,
  wert,
  setze,
  partie,
}: {
  art: 'angriffe' | 'bloecke'
  wert: string[]
  setze: (wert: string[]) => void
  partie: Partie
}) {
  const texte = useTexte()
  const eigene = texte.seiten.handlungen

  const eigeneKreaturen = partie.objekte.filter(
    (eines) =>
      eines.zone === 'tisch' &&
      eines.beherrscher === partie.fuer &&
      (eines.eigenschaften === null ||
        eines.eigenschaften.typen.includes('creature')),
  )

  if (eigeneKreaturen.length === 0) {
    return <p className="text-sm text-leise">{eigene.kampfKeineKreatur}</p>
  }

  const ziele = art === 'angriffe' ? gegner(partie) : angreifer(partie)

  if (ziele.length === 0) {
    return <p className="text-sm text-leise">{eigene.kampfKeinAngreifer}</p>
  }

  return (
    <ul className="flex flex-col gap-1">
      {eigeneKreaturen.map((kreatur) => (
        <li key={kreatur.kennung} className="flex items-baseline gap-2">
          <span className="text-sm">{kartenwort(kreatur)}</span>

          <span className="text-sm text-leise">
            {art === 'angriffe' ? eigene.kampfGreiftAn : eigene.kampfBlockt}
          </span>

          <select
            aria-label={kartenwort(kreatur)}
            value={zielVon(wert, kreatur.kennung)}
            onChange={(ereignis) => {
              setze(
                ohneKreatur(wert, kreatur.kennung).concat(
                  ereignis.target.value === ''
                    ? []
                    : [`${kreatur.kennung}|${ereignis.target.value}`],
                ),
              )
            }}
            className={eingabeklassen}
          >
            <option value="">{eigene.handlungKeineWahl}</option>

            {ziele.map((eines) => (
              <option key={eines.wert} value={eines.wert}>
                {eines.text}
              </option>
            ))}
          </select>
        </li>
      ))}
    </ul>
  )
}

// Wahl ist ein Eintrag in der Zielauswahl.
interface Wahl {
  wert: string
  text: string
}

// gegner sind die Sitze, die angegriffen werden können — CR 508.1a.
//
// **Ohne den eigenen**: niemand greift sich selbst an, und der Kern weist es
// ab (`verteidigerSteht`). Ein Eintrag, der immer abgewiesen würde, ist eine
// Einladung, die man zurücknimmt.
function gegner(partie: Partie): Wahl[] {
  return partie.sitze
    .filter((sitz) => sitz.kennung !== partie.fuer)
    .map((sitz) => ({ wert: sitz.kennung, text: sitz.name }))
}

// angreifer sind die erklärten Angreifer — CR 509.1a.
function angreifer(partie: Partie): Wahl[] {
  return partie.zug.angriffe.map((angriff) => ({
    wert: angriff.kreatur,
    text: kennungswort(partie, angriff.kreatur),
  }))
}

// kennungswort gibt den Namen zu einer Objektkennung.
function kennungswort(partie: Partie, kennung: string): string {
  const objekt = partie.objekte.find((eines) => eines.kennung === kennung)

  return objekt === undefined ? kennung : kartenwort(objekt)
}

// ohneKreatur gibt die Zeilen ohne die dieser Kreatur.
function ohneKreatur(wert: string[], kreatur: string): string[] {
  return wert.filter((eine) => !eine.startsWith(`${kreatur}|`))
}

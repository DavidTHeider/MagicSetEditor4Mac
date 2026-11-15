// **Derselbe Druck zweimal ist eine Entscheidung über die Menge** — DOM-03,
// S04.NAT-01 bis 03 (#1698).
//
// # Was vorher geschah
//
// **Eine Absage**: „Diesen Druck führt die Sammlung schon … Erhöhe dort die
// Menge“ (nativer Befund F08, ENG-05). Sie stimmte, und sie ließ offen, wo
// „dort“ ist — die Zeile konnte hinter einem Filter liegen oder außerhalb
// des geholten Ausschnitts.
//
// # Was jetzt dasteht
//
// **Drei Antworten an Ort und Stelle** (`Entscheidung`, §6): die Menge der
// vorhandenen Zeile erhöhen, zu ihr springen, oder nichts tun. Der Fokus
// steht auf dem Abbrechen — wer zweimal ⏎ drückt, ändert keine Menge.
import { Entscheidung } from './Entscheidung'
import { Knopf } from './Knopf'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Mengenschritte } from '../haken/mengenentscheidung'
import type { Dublette } from '../haken/useSammlung'

interface MengenentscheidungEigenschaften {
  dublette: Dublette
  schritte: Mengenschritte
}

export function Mengenentscheidung({
  dublette,
  schritte,
}: MengenentscheidungEigenschaften) {
  const eigene = useTexte().seiten.sammlung
  const { vorhanden, neu } = dublette

  return (
    <Entscheidung
      frage={einsetzen(eigene.dubletteFrage, {
        bezeichnung: vorhanden.bezeichnung,
        menge: String(vorhanden.menge),
      })}
      erklaerung={eigene.dubletteErklaerung}
      bestaetigen={einsetzen(eigene.dubletteErhoehen, {
        summe: String(vorhanden.menge + neu.menge),
      })}
      abbrechen={eigene.dubletteAbbrechen}
      laeuft={schritte.laeuft}
      aufBestaetigen={schritte.erhoehen}
      aufAbbrechen={schritte.abbrechen}
    >
      <div>
        <Knopf angehalten={schritte.laeuft} aufKlick={schritte.zurZeile}>
          {eigene.dubletteZurZeile}
        </Knopf>
      </div>
    </Entscheidung>
  )
}

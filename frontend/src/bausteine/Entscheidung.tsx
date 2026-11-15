// **Eine Rückfrage an Ort und Stelle** — Werkstück U2 von Phase 5.1
// (#1413),
// nach [Vermerk 0077](../../../docs/entscheidungen/0077-eigene-primitive-statt-radix.md).
//
// # Wozu
//
// Zehn Stellen fragen, bevor sie etwas tun, das nicht zurückkommt — Set,
// Karte, Deck, Turnier, Nutzer, Profil, Stil, Ersetzen, Sicherung,
// Verschlüsselung —, und jede war bis hierher eine eigene Handvoll Knöpfe.
// Was gleich ist, steht jetzt hier: die Frage benennt die Gruppe, der Fokus
// springt hinein, Escape ist Abbrechen, und danach steht der Fokus wieder
// dort, wo er war.
//
// # Warum kein Dialog
//
// §6: keine verschachtelten Dialoge, Bearbeitung im Seitenbereich. Die Frage
// steht dort, wo geklickt wurde, und nimmt dem Rest der Seite nichts weg —
// `role="group"` und nicht `alertdialog`, denn nichts ist modal. Das ist
// dieselbe Entscheidung, die `Kartenloeschen` seit #621 trifft; sie gilt jetzt
// für alle.
//
// # Wohin der Fokus springt
//
// **Auf das Abbrechen, nicht auf das Bestätigen.** Wer mit ⏎ „Löschen" drückt
// und gleich noch einmal ⏎, soll damit nichts löschen — die WAI-ARIA-Muster
// setzen den Fokus bei einer zerstörenden Frage auf die ungefährliche
// Antwort, und hier auch.
import { useId, useRef, type ReactNode } from 'react'

import type { Aktion } from '../aktionen'
import { Aktionsknopf } from './Aktionsknopf'
import { useRueckkehr } from '../haken/useRueckkehr'
import { Knopf } from './Knopf'

interface EntscheidungEigenschaften {
  // frage ist der Satz, der die Gruppe benennt — und die eine Zeile, die
  // jemand liest, bevor er drückt.
  frage: string

  // erklaerung steht unter der Frage: was geschieht, und was nicht.
  erklaerung?: string

  // bestaetigen und abbrechen sind die Aufschriften — aus dem Katalog des
  // Aufrufers (§9), denn „Löschen" und „Zurückspielen" sind zweierlei.
  bestaetigen: string
  abbrechen: string

  // aktion ist die Katalogaktion, die das Bestätigen auslöst — UX-10.AK-01
  // (#1682). **Das Ja einer Löschfrage ist ein Löschen**, und es trägt dann
  // dasselbe Symbol wie der Knopf, der die Frage gestellt hat. Ohne Angabe
  // ist die Bestätigung keine Katalogaktion — ein Token erneuern etwa.
  aktion?: Aktion

  // laeuft hält Regel 4 aus §7: während der erste Klick arbeitet, ruht der
  // zweite.
  laeuft?: boolean

  aufBestaetigen: () => void
  aufAbbrechen: () => void

  // children steht zwischen Erklärung und Knöpfen — ein Fehler, ein Feld
  // für den Namen, ein „Trotzdem".
  children?: ReactNode
}

export function Entscheidung({
  frage,
  erklaerung,
  bestaetigen,
  abbrechen,
  aktion,
  laeuft = false,
  aufBestaetigen,
  aufAbbrechen,
  children,
}: EntscheidungEigenschaften) {
  const frageID = useId()
  const erklaerungID = useId()
  const abbruchknopf = useRef<HTMLButtonElement>(null)

  // **Der Auslöser wird beim Erscheinen gemerkt, nicht beim Verschwinden.**
  // Beim Verschwinden steht der Fokus schon auf einem Knopf dieser Gruppe —
  // und der ist gleich weg.
  //
  // **Und nur beim echten Erscheinen und Verschwinden** — nicht, wenn die
  // Seite aufbewahrt und wieder eingeblendet wird (`useRueckkehr`, G11).
  // Gibt es den Auslöser nicht mehr — nach einem Löschen ist die Zeile weg,
  // in der er stand —, sagt `fokus.ts`, was geschieht.
  useRueckkehr(abbruchknopf, () => abbruchknopf.current?.focus())

  return (
    // **Die Gruppe hört Escape von ihren Knöpfen.** `jsx-a11y` sieht ein
    // `div` mit Tastenhorcher und weiß nicht, dass die Rolle `group` samt
    // Namen dasteht und der Fokus auf einem Knopf darin liegt — die Ausnahme
    // steht hier, damit sie begründet ist.
    // eslint-disable-next-line jsx-a11y/no-noninteractive-element-interactions
    <div
      role="group"
      aria-labelledby={frageID}
      aria-describedby={erklaerung === undefined ? undefined : erklaerungID}
      onKeyDown={(ereignis) => {
        // **Nicht, wenn ein Hinweis darin sie schon genommen hat** —
        // das Innerste zuerst (G12, WCAG 1.4.13).
        if (ereignis.key === 'Escape' && !ereignis.defaultPrevented) {
          // **Und sagt, dass es die Taste verbraucht hat** — eine Schublade
          // darum horcht in der Capture-Phase, also vor dem
          // `stopPropagation`, und bliebe sonst nicht offen (UX-08, #1680).
          ereignis.preventDefault()
          ereignis.stopPropagation()
          aufAbbrechen()
        }
      }}
      className="flex flex-col gap-2 rounded-sm border border-kante bg-flaeche p-3"
    >
      <p id={frageID} className="font-medium">
        {frage}
      </p>

      {erklaerung !== undefined && (
        <p id={erklaerungID} className="max-w-prose text-sm text-leise">
          {erklaerung}
        </p>
      )}

      {children}

      <div className="flex flex-wrap gap-2">
        {aktion === undefined ? (
          <Knopf gewicht="haupt" angehalten={laeuft} aufKlick={aufBestaetigen}>
            {bestaetigen}
          </Knopf>
        ) : (
          <Aktionsknopf
            aktion={aktion}
            gewicht="haupt"
            angehalten={laeuft}
            aufKlick={aufBestaetigen}
          >
            {bestaetigen}
          </Aktionsknopf>
        )}

        <Knopf griff={abbruchknopf} angehalten={laeuft} aufKlick={aufAbbrechen}>
          {abbrechen}
        </Knopf>
      </div>
    </div>
  )
}

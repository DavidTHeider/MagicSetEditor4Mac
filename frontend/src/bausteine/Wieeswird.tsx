// **„Wie es wird"** — der Schritt vor dem Schreiben, UXR-F7
// (#2015),
// nach dem Artboard *Ausgeben* (`docs/prototyp/teile/Ausgeben.html`).
//
// # Was er ist
//
// Der Rahmen um den Probelauf eines Ziels: eine Überschrift, das, was der
// Dienst gerechnet hat, und der Satz, dass erst der Knopf darunter schreibt.
// **Was darin steht, bestimmt das Ziel** (`children`); der Rahmen weiß nur,
// ob noch gerechnet wird, ob die Rechnung scheiterte und ob die Zahlen schon
// zu einer früheren Wahl gehören.
//
// # Warum eine Zusage an eine Bildschirmleserin
//
// **`aria-live` ohne `role`**: die Zahl ändert sich, wenn jemand Papier oder
// Auswahl wechselt, und wer nichts sieht, soll es hören. `role="status"` wäre
// dafür der übliche Weg — aber die Fläche trägt schon einen Stand für das
// Ergebnis des Schreibens, und zwei davon machten aus „dem Status" eine
// Mehrdeutigkeit, die kein Test mehr auflöst.
import { useId, type ReactNode } from 'react'

import { useTexte } from '../sprache/kontext'

interface WieeswirdEigenschaften {
  // wartet sagt, dass es noch keine Zahlen gibt und auch keinen Fehler.
  wartet: boolean

  // veraltet sagt, dass die Zahlen zu einer früheren Wahl gehören und neue
  // unterwegs sind — sie bleiben stehen, und neben der Überschrift steht,
  // dass gerechnet wird. **Nicht gedämpft**: Text mit weniger Deckkraft
  // verfehlte den Kontrast, den der Rest der Fläche hält.
  veraltet: boolean

  // fehler ist der Satz, wenn die Rechnung scheiterte — **derselbe, den der
  // Knopf gäbe** (ein Weg, zwei Ausgänge).
  fehler: string | undefined

  children: ReactNode
}

export function Wieeswird({
  wartet,
  veraltet,
  fehler,
  children,
}: WieeswirdEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.ausgeben.probelauf
  const titel = useId()

  return (
    <section
      aria-labelledby={titel}
      className="mt-4 flex max-w-2xl flex-col gap-2 rounded-sm border border-kante bg-flaeche p-3"
    >
      <div className="flex items-baseline gap-2">
        <h4 id={titel} className="text-sm font-medium">
          {eigene.ueberschrift}
        </h4>

        {veraltet && !wartet && (
          <span className="text-sm text-leise">{eigene.rechnet}</span>
        )}
      </div>

      <div
        aria-live="polite"
        aria-busy={veraltet || wartet}
        className="flex flex-col gap-2"
      >
        {wartet ? (
          <p className="text-sm text-leise">{eigene.rechnet}</p>
        ) : (
          children
        )}
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {texte.bausteine.zustaende.fehler} {fehler}
        </p>
      )}

      <p className="text-sm text-leise">{eigene.hinweis}</p>
    </section>
  )
}

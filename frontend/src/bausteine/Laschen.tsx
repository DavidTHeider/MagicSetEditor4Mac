// **Laschen für eine Seite mit mehreren Gebieten** — Werkstück I von Phase
// 4.10 (#1325).
//
// # Wozu, wo der Setarbeitsplatz doch schon eine Leiste hat
//
// Weil seine an einem **Set** hängt: er kennt Setschlüssel, führt die
// Laschengruppen des Spiels und wechselt über den Anwendungszustand. Eine
// Deckseite hat kein Set und keine Gruppen — sie hat zwei Gebiete, und die
// Turnierwerkbank hat drei.
//
// **Was gleich bleibt, ist das ARIA-Muster**: eine `tablist` ist **eine**
// Tabstelle, zwischen den Laschen wandern die Pfeile, und das Feld nennt seine
// Lasche. Das ist der Teil, der hier steht.
//
// # Der Befund, für den es sie gibt
//
// Die Deckseite legte vier Blöcke untereinander — Anlegen, Liste, das offene
// Deck, Formate —, und im offenen Deck noch einmal fünf. Am 2026-09-03 im
// Fenster nachgemessen: 1.614 Punkte, und wer ein Format anlegen wollte,
// scrollte an allem vorbei.
//
// # Was sie nicht tut
//
//   - **Sie merkt sich nichts.** Die gewählte Lasche steht im Zustand der
//     Komponente; wer die Seite verlässt, kommt beim ersten Gebiet wieder an.
//     Ein Gedächtnis wäre eine Einstellung, und dafür gibt es keinen Anlass.
//   - **Sie kennt keine Gruppen.** Fünf Laschen mit Überschriften darüber sind
//     die Antwort des Setarbeitsplatzes auf **dreizehn** Gebiete; hier sind es
//     zwei bis drei.
//   - **Sie führt die Wahl nicht selbst.** Welches Gebiet offen ist und wie
//     die Pfeile wandern, steht seit UX-04.AK-01 in
//     [useGebietswahl](../haken/useGebietswahl.ts) — geteilt mit der
//     senkrechten [Gruppenliste](./Gruppenliste.tsx), die das Artboard
//     `Einstellungen` zeichnet.
import { useGebietswahl, type Gebiet } from '../haken/useGebietswahl'

import { lasche } from './laschenstil'

export type { Gebiet }

interface Lascheneigenschaften {
  // bezeichnung benennt die Leiste für Bedienhilfen.
  bezeichnung: string

  gebiete: readonly Gebiet[]

  // oeffne schlägt ein Gebiet von außen auf.
  //
  // **Der Anlass ist eine Regression, die dieser Umbau selbst erzeugt hat**:
  // der Menüeintrag „Über MagicSetEditor4Mac" rollt auf der Hilfeseite zum
  // Abschnitt „Über" — und seit die Hilfe Laschen führt, stand der Abschnitt
  // gar nicht mehr im Baum, solange sein Gebiet zu war. Wer eine Seite
  // gliedert, muss jeden Weg dorthin mitnehmen.
  //
  // **Der Wert wird bei jeder Änderung gelesen, nicht nur beim ersten Mal.**
  // Ein Aufrufer, der denselben Weg zweimal geht, hängt deshalb eine Nummer
  // an — sonst wirkte der Menüeintrag genau einmal.
  //
  // **Ein Schlüssel, den es nicht gibt, ändert nichts.** Er kann nur aus einem
  // Tippfehler im Aufrufer kommen; das offene Gebiet stehen zu lassen ist
  // besser als eine leere Seite.
  oeffne?: string | undefined
}

export function Laschen({
  bezeichnung,
  gebiete,
  oeffne,
}: Lascheneigenschaften) {
  const { offen, kennung, leiste, waehle, aufTaste } = useGebietswahl(
    gebiete,
    oeffne,
    'ArrowRight',
    'ArrowLeft',
  )

  if (offen === undefined) {
    return null
  }

  return (
    <div className="flex flex-col gap-4">
      <div
        ref={leiste}
        role="tablist"
        aria-label={bezeichnung}
        className="flex flex-wrap gap-1 border-b border-kante pb-1"
      >
        {gebiete.map((gebiet) => {
          const dieses = gebiet.schluessel === offen.schluessel

          return (
            <button
              key={gebiet.schluessel}
              type="button"
              role="tab"
              id={`${kennung}-${gebiet.schluessel}`}
              data-lasche={gebiet.schluessel}
              aria-selected={dieses}
              aria-controls={`${kennung}-feld`}
              tabIndex={dieses ? 0 : -1}
              onClick={() => {
                waehle(gebiet.schluessel)
              }}
              onKeyDown={aufTaste}
              className={lasche}
            >
              {gebiet.name}
            </button>
          )
        })}
      </div>

      <div
        id={`${kennung}-feld`}
        role="tabpanel"
        aria-labelledby={`${kennung}-${offen.schluessel}`}
        tabIndex={0}
      >
        {offen.inhalt}
      </div>
    </div>
  )
}

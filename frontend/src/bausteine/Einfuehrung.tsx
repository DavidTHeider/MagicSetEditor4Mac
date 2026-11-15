// **Die Einführung als Fenster** —
// #1325,
// Werkstück C von Phase 4.10.
//
// # Warum sie kein Teil der Startseite mehr ist
//
// Sie war eine **Hälfte** davon: die Seite zeigte *Erste Schritte* **oder**
// *Dein Überblick*, je nachdem, ob der Bestand ein Set führt. Zwei Dinge
// gingen dabei schief:
//
//   - **Sie sprang.** Solange die Zahlen unterwegs waren, galt „kein Set",
//     und die Seite baute sich um, sobald die Antwort kam. Der Kommentar an
//     dieser Stelle sagte ausdrücklich, das dürfe nicht geschehen — und der
//     Code tat es (§10: ein Kopf, der mehr verspricht als sein Rumpf hält).
//   - **Sie war weg, sobald man etwas hatte.** Eine Einführung, die
//     verschwindet, sobald man ein Set anlegt, fehlt genau an dem Tag, an dem
//     man sie zum zweiten Mal bräuchte.
//
// # Und warum ein Fenster kein verschachtelter Dialog ist
//
// §6 verbietet Dialoge übereinander. Dieses steht beim Start über der
// Startseite und über nichts sonst; wer einem Schritt folgt, schließt es
// dabei.
import { useId, useState } from 'react'

import { Dialog } from './Dialog'
import { Knopf } from './Knopf'
import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import type { Seitenschluessel } from '../navigation'

interface EinfuehrungEigenschaften {
  // setID ist das offene Set — der mittlere Schritt hängt daran.
  setID: string | undefined

  wechseln: (seite: Seitenschluessel) => void
  schliessen: () => void

  // wiederZeigen ist die gespeicherte Wahl. Ohne Angabe: ja — das ist die
  // Vorgabe der Domäne (`ErsteSchritteAn`).
  wiederZeigen?: boolean

  // gewaehlt meldet die neue Wahl nach oben, damit die Einstellungsseite sie
  // ohne Neuladen sieht.
  gewaehlt?: (wiederZeigen: boolean) => void
}

// schritte ist die Reihenfolge, in der jemand sie geht.
const schritte: {
  schluessel: 'setOeffnen' | 'kartenBearbeiten' | 'paketeHolen'
  ziel: Seitenschluessel
}[] = [
  { schluessel: 'setOeffnen', ziel: 'sets' },
  { schluessel: 'kartenBearbeiten', ziel: 'karten' },
  { schluessel: 'paketeHolen', ziel: 'pakete' },
]

export function Einfuehrung({
  setID,
  wechseln,
  schliessen,
  wiederZeigen = true,
  gewaehlt,
}: EinfuehrungEigenschaften) {
  const texte = useTexte()
  const start = texte.seiten.start
  const schalterID = useId()

  // **Der Haken hat einen eigenen Stand, und das ist der ganze Punkt.**
  //
  // # Warum die Eigenschaft allein nicht reicht
  //
  // **Weil sie sich nicht ändert.** `checked={wiederZeigen}` machte daraus
  // ein kontrolliertes Feld, dessen Wert von oben kommt — und oben stand
  // `wiederZeigen` fest. Der Klick schrieb die Einstellung, React zeichnete
  // den Haken sofort wieder hin, und für den Menschen davor war der Schalter
  // kaputt. Gemeldet vom Auftraggeber am 2026-09-06.
  //
  // **Optimistisch, mit Rücknahme bei Fehlern** (§6): der Haken folgt dem
  // Klick sofort; scheitert das Schreiben, springt er zurück, und die Meldung
  // steht daneben.
  //
  // **Die Eigenschaft bleibt der Anfangswert** — sie sagt, was gespeichert
  // ist, wenn das Fenster aufgeht.
  const [angehakt, setzeAngehakt] = useState(wiederZeigen)
  const [fehler, setzeFehler] = useState<string | undefined>(undefined)

  const waehle = async (an: boolean) => {
    setzeAngehakt(an)
    setzeFehler(undefined)

    try {
      const antwort = await api().EinstellungSetzen(
        'ersteschritte',
        an ? 'an' : 'aus',
      )

      auspacken(antwort, () => antwort)

      // **Erst nach der Antwort nach oben melden** (§7, Regel 3): wer es
      // davor täte, ließe die Einstellungsseite gegen den alten Stand neu
      // lesen — ein Wettlauf, den sie meistens verliert.
      gewaehlt?.(an)
    } catch (grund) {
      setzeAngehakt(!an)
      setzeFehler(meldungAus(grund, texte))
    }
  }

  // **Escape horcht am Fenster, nicht am Kasten** — derselbe Grund wie bei
  // der Befehlspalette: an einem `div` mit `role="dialog"` wäre es ein
  // Ereignishorcher an einem nicht bedienbaren Element, und es griffe nur,
  // solange der Fokus drinnen steht.
  // **Escape und Fokus gehören dem Dialog** — seit Werkstück U2 von Phase
  // 5.1 (#1413), mit derselben Zusage wie die Palette und erstmals mit der
  // Rückgabe des Fokus.

  return (
    <Dialog
      bezeichnung={start.einfuehrung.bezeichnung}
      ausrichtung="mitte"
      schliessen={schliessen}
    >
      {/* **`data-einfuehrung` für die Hüllenprobe** — sie schließt das
          Fenster vor jedem Drehbuch, und bis zum 2026-09-30 erkannte sie es
          an „Erste Schritte": auf Englisch oder Japanisch blieb es stehen. */}
      <div
        data-einfuehrung
        className="max-h-full w-full max-w-3xl overflow-y-auto rounded-lg border-2 border-kante bg-flaeche p-6"
      >
        <h2 className="text-lg font-semibold">{start.ersteSchritte}</h2>

        <p className="mt-1 mb-4 max-w-prose text-sm text-leise">
          {start.einfuehrung.willkommen}
        </p>

        <ul className="grid gap-4 sm:grid-cols-3">
          {schritte.map(({ schluessel, ziel }) => {
            const schritt = start.schritte[schluessel]

            // **Ein Schritt, den man noch nicht gehen kann, bleibt stehen**
            // (#673).
            // Ihn auszublenden nähme dem Anfänger die Landkarte — und weil es
            // der mittlere ist, sähe es aus, als fehlte ein anderer.
            const fehltDasSet =
              schluessel === 'kartenBearbeiten' && setID === undefined

            return (
              <li
                key={schluessel}
                className="rounded-md border border-kante p-3"
              >
                <h3 className="font-semibold">{schritt.titel}</h3>

                <p className="mt-1 mb-4 text-sm text-leise">
                  {fehltDasSet
                    ? start.schritte.kartenBearbeiten.ersteinSet
                    : schritt.erklaerung}
                </p>

                {/* **Er führt dorthin, wo die Voraussetzung entsteht** —
                    nicht ins Leere, und nicht ausgegraut: ein abgeschalteter
                    Knopf fiele aus der Tastaturreihenfolge (§6). */}
                <Knopf
                  gewicht={fehltDasSet ? 'neben' : 'haupt'}
                  aufKlick={() => {
                    wechseln(fehltDasSet ? 'sets' : ziel)
                    schliessen()
                  }}
                >
                  {fehltDasSet
                    ? start.schritte.kartenBearbeiten.stattdessen
                    : schritt.handlung}
                </Knopf>
              </li>
            )
          })}
        </ul>

        <div className="mt-6 flex flex-wrap items-center justify-between gap-3">
          {/* **Der Schalter steht hier und in den Einstellungen** (§6):
              niemand sucht die Einstellungen, um etwas loszuwerden, das gerade
              vor ihm steht.

              **Und das Fenster bleibt dabei offen.** Wer den Haken wegnimmt,
              sagt etwas über das nächste Mal, nicht über dieses. */}
          <label
            htmlFor={schalterID}
            className="flex items-center gap-2 text-sm"
          >
            <input
              id={schalterID}
              type="checkbox"
              checked={angehakt}
              className="size-4 accent-marke focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              onChange={(ereignis) => {
                void waehle(ereignis.target.checked)
              }}
            />
            {start.einfuehrung.wiederzeigen}
          </label>

          {fehler !== undefined && (
            <p role="alert" className="text-fehler">
              {fehler}
            </p>
          )}

          <Knopf gewicht="haupt" aufKlick={schliessen}>
            {start.einfuehrung.schliessen}
          </Knopf>
        </div>
      </div>
    </Dialog>
  )
}

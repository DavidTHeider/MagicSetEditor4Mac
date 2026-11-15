// **Typen, Länder, Durchschnitt und Quellen** — Werkstück U10 von Phase 5.1
// (#1421).
//
// **Eine eigene Datei neben `Deckauswertung.tsx`**, damit beide unter der
// Zeilengrenze aus §2 bleiben — und weil der Schnitt der Frage folgt: dort
// stehen die zwei Bilder (Kurve, Farben), hier die vier Zahlen.
//
// **Die Analyse ändert nichts.** Sie liest, was das Deck ohnehin mitbringt,
// und schreibt nirgends hin — das ist die Zusage aus dem Issue.
import {
  durchschnitt,
  laenderzahl,
  quellen,
  typverteilung,
} from './deckstatistik'
import { farbenVon } from './deckrechnung'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

import type { Farbe } from './deckrechnung'
import type { Statistikposten } from './deckstatistik'

export function Deckkennzahlen({
  posten,
}: {
  posten: readonly Statistikposten[]
}) {
  const texte = useTexte().seiten.decks

  const farbname: Record<Farbe, string> = {
    weiss: texte.farbeWeiss,
    blau: texte.farbeBlau,
    schwarz: texte.farbeSchwarz,
    rot: texte.farbeRot,
    gruen: texte.farbeGruen,
  }

  const typen = typverteilung(posten)
  const laender = laenderzahl(posten)
  const mittel = durchschnitt(posten)

  // **Die Farben eines Landes stehen in seinen Kosten nicht** — ein Wald
  // kostet nichts. Sie kommen aus dem **Untertyp**: `Land — Forest Island`
  // gibt Grün und Blau. Was ein Untertyp gibt, steht in den Comprehensive
  // Rules 305.6 und ist hier für die fünf Grundtypen eingetragen; alles
  // andere zählt nicht mit, statt geraten zu werden.
  const quellenzeilen = quellen(
    posten.map((eintrag) => ({ ...eintrag, farben: landfarben(eintrag) })),
  )

  if (typen.length === 0) {
    return null
  }

  return (
    <div className="flex flex-wrap gap-8">
      <section className="min-w-56 flex-1">
        <h3 className="font-medium">{texte.typen}</h3>

        <ul aria-label={texte.typen} className="flex flex-col gap-1">
          {typen.map(([typ, zahl]) => (
            <li key={typ} className="flex items-center gap-2">
              <span className="w-40 text-leise">{typ}</span>
              <span className="tabular-nums">{zahl}</span>
            </li>
          ))}
        </ul>
      </section>

      <section className="min-w-56 flex-1">
        <h3 className="font-medium">{texte.kennzahlen}</h3>

        <dl aria-label={texte.kennzahlen} className="flex flex-col gap-1">
          <div className="flex items-center gap-2">
            <dt className="w-40 text-leise">{texte.laender}</dt>
            <dd className="tabular-nums">{laender}</dd>
          </div>

          <div className="flex items-center gap-2">
            <dt className="w-40 text-leise">{texte.durchschnitt}</dt>
            <dd className="tabular-nums">
              {mittel === undefined ? texte.keineZahl : mittel.toFixed(2)}
            </dd>
          </div>
        </dl>

        {quellenzeilen.length > 0 && (
          <>
            <h4 className="mt-3 font-medium">{texte.quellen}</h4>

            <Erklaerungsabsatz
              text={texte.quellenHinweis}
              klasse="mb-1 max-w-prose text-sm text-leise"
            />

            <ul aria-label={texte.quellen} className="flex flex-col gap-1">
              {quellenzeilen.map(([farbe, zahl]) => (
                <li key={farbe} className="flex items-center gap-2">
                  <span className="w-40 text-leise">{farbname[farbe]}</span>
                  <span className="tabular-nums">{zahl}</span>
                </li>
              ))}
            </ul>
          </>
        )}
      </section>
    </div>
  )
}

// grundtypen ordnet den fünf Landgrundtypen ihre Farbe zu.
//
// **Nachgelesen in den Comprehensive Rules 305.6**: die fünf Grundtypen
// tragen die eingebaute Fähigkeit, ihr Mana zu geben. Alles andere — ein Land
// mit Regeltext — zählt hier **nicht** mit; wer es zählte, müsste den Text
// deuten, und das ist geraten (§10).
const grundtypen: Record<string, Farbe> = {
  plains: 'weiss',
  island: 'blau',
  swamp: 'schwarz',
  mountain: 'rot',
  forest: 'gruen',
}

// landfarben gibt die Farben, die ein Land nach seinen Untertypen gibt.
//
// **Für alles, was kein Land ist, sind es keine** — `quellen` zählt sie
// ohnehin nicht, und eine Farbe aus den Kosten wäre hier die falsche Frage.
function landfarben(eintrag: Statistikposten): readonly Farbe[] {
  const [, unter = ''] = eintrag.typzeile.split('—')

  const ausTypen = unter
    .trim()
    .split(/\s+/u)
    .flatMap((wort): Farbe[] => {
      const farbe = grundtypen[wort.toLowerCase()]

      return farbe === undefined ? [] : [farbe]
    })

  // **Ohne Grundtyp bleibt der Rückfall auf die Kosten** — er trifft für ein
  // Land nie zu (Länder kosten nichts) und hält die Zusage trotzdem: was
  // hier steht, sind die Farben **dieses** Postens.
  return ausTypen.length > 0 ? ausTypen : farbenVon(eintrag.kosten)
}

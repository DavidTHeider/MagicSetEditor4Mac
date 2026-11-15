// **Eine einzelne Zeile der Deutung** — und was aus ihr entschieden ist.
//
// **Eine eigene Datei**, seit `Listeneinfuhr.tsx` an der Zeilengrenze aus §2
// stand. Der Schnitt folgt der Frage: dort steht, *wie der Kasten sich
// verhält* — deuten, einlesen, was er anzeigt —, hier, *wie eine einzelne
// Zeile aussieht* und *welche Zeilen entschieden sind*.
import { useId } from 'react'

import { Bindungsfehler } from '../api'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'

import type { Katalog } from '../sprache/kataloge'
import type {
  Listendeutung,
  Listendeutungszeile,
  Listeneinfuhrzeile,
} from '../antwortenlisteneinfuhr'

interface DeutungszeileEigenschaften {
  zeile: Listendeutungszeile
  gewaehlt: string
  setze: (printing: string) => void
}

// Deutungszeile zeigt eine Zeile und, wenn nötig, ihre Auswahl.
export function Deutungszeile({
  zeile,
  gewaehlt,
  setze,
}: DeutungszeileEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  const wahlID = useId()
  const befund = befundtext(zeile, texte)

  return (
    <div className="flex flex-wrap items-center gap-3">
      <span className="font-mono text-sm">{zeile.roh}</span>

      {befund !== '' && <span className="text-sm text-hinweis">{befund}</span>}

      {zeile.eindeutig && zeile.treffer[0] !== undefined && (
        <span className="text-sm text-leise">
          {beschriftung(zeile.treffer[0])}
        </span>
      )}

      {/* **Die Auswahl nur, wo es etwas zu wählen gibt.** Ein Feld mit einer
          einzigen Möglichkeit ist eine Tabulatorstelle ohne Zweck. */}
      {!zeile.eindeutig && zeile.treffer.length > 0 && (
        <label className="flex min-w-0 max-w-full items-center gap-2">
          <span className="text-sm" id={wahlID}>
            {eigene.listeDruckWahl}
          </span>
          <select
            aria-labelledby={wahlID}
            value={gewaehlt}
            onChange={(ereignis) => {
              setze(ereignis.target.value)
            }}
            className="min-w-0 max-w-full rounded-sm border border-kante bg-transparent px-2 py-1 text-sm focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {/* **Nichts ist vorgewählt.** Ein vorgewählter Druck wäre eine
                Entscheidung, die niemand getroffen hat (§16, Regel 6). */}
            <option value="">{eigene.listeOhneWahl}</option>
            {zeile.treffer.map((treffer) => (
              <option key={treffer.printing} value={treffer.printing}>
                {beschriftung(treffer)}
              </option>
            ))}
          </select>

          {/* **Und wie viele es insgesamt sind** (§10): die Liste ist bei
              zwanzig gekappt, und das darf nicht wie Vollständigkeit aussehen. */}
          {zeile.drucke > zeile.treffer.length && (
            <span className="text-xs text-leise">
              {einsetzen(eigene.listeDrucke, { drucke: String(zeile.drucke) })}
            </span>
          )}
        </label>
      )}
    </div>
  )
}

// befundtext sagt, warum eine Zeile nicht aufgeht — **in der Sprache des
// Fensters** (G01, `g01einsteiger`). Bis dahin stand hier der `befund` aus Go:
// deutsch in jeder Sprache, und bei einem Ausfall der rohe Fehler samt seiner
// Klasse. Was Go nur als Satz kennt — die Befunde der Quelle —, bleibt der Satz.
function befundtext(zeile: Listendeutungszeile, texte: Katalog): string {
  const eigene = texte.seiten.sammlung

  if (zeile.fehler !== null) {
    return meldungAus(new Bindungsfehler(zeile.fehler, []), texte)
  }

  if (zeile.drucke === 0) {
    return eigene.listeKeinDruck
  }

  if (zeile.weitere) {
    return einsetzen(eigene.listeZuViele, { drucke: String(zeile.drucke) })
  }

  return zeile.befund
}

// Deutungsbilanz sagt, wie viele Zeilen eine Entscheidung brauchen — und
// **getrennt davon**, wie viele keine Karte finden: die lassen sich nicht
// entscheiden und werden übergangen (G01).
export function Deutungsbilanz({
  zeilen,
}: {
  zeilen: readonly Listendeutungszeile[]
}) {
  const eigene = useTexte().seiten.sammlung

  const offen = zeilen.filter(
    (zeile) => !zeile.eindeutig && zeile.treffer.length > 0,
  ).length
  const ohneKarte = zeilen.filter(
    (zeile) => !zeile.eindeutig && zeile.treffer.length === 0,
  ).length
  const gesamt = String(zeilen.length)

  return (
    <>
      <p className="text-sm text-leise">
        {offen === 0 && ohneKarte === 0
          ? einsetzen(eigene.listeAlleEindeutig, { gesamt })
          : offen > 0 &&
            einsetzen(eigene.listeOffen, { offen: String(offen), gesamt })}
      </p>
      {ohneKarte > 0 && (
        <p className="text-sm text-leise">
          {ohneKarte === 1
            ? eigene.listeOhneKarteEins
            : einsetzen(eigene.listeOhneKarte, { anzahl: String(ohneKarte) })}
        </p>
      )}
    </>
  )
}

// beschriftung gibt den lesbaren Bezug eines Drucks.
function beschriftung(treffer: {
  name: string
  setname: string
  nummer: string
}): string {
  return `${treffer.name} · ${treffer.setname} #${treffer.nummer}`
}

// entschieden gibt die Zeilen, für die ein Druck feststeht.
//
// **Eine eindeutige Zeile ist entschieden, ohne dass jemand klickt** — es gibt
// dort nichts zu wählen. Eine offene ohne Wahl fällt weg; sie zu raten wäre
// genau das, was der zweite Schritt verhindert.
export function entschieden(
  deutung: Listendeutung | null,
  wahl: Record<number, string>,
): Listeneinfuhrzeile[] {
  const zeilen: Listeneinfuhrzeile[] = []

  for (const zeile of deutung?.zeilen ?? []) {
    const printing = zeile.eindeutig
      ? (zeile.treffer[0]?.printing ?? '')
      : (wahl[zeile.zeilennummer] ?? '')

    if (printing === '') {
      continue
    }

    const treffer = zeile.treffer.find((einer) => einer.printing === printing)
    if (treffer === undefined) {
      continue
    }

    zeilen.push({
      printing,
      menge: zeile.menge,
      // **Der Zustand ist die Vorgabe.** Wer eine Liste einfügt, hat keine
      // Zustandsangabe darin; wer sie braucht, ändert die Zeile danach.
      zustand: 'nearMint',
      sprache: 'en',
      finish: zeile.finish,
      name: treffer.name,
      setcode: treffer.setcode,
      nummer: treffer.nummer,
    })
  }

  return zeilen
}

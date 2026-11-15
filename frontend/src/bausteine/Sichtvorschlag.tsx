// **Was dein Gegenüber freigibt, und sein Vorschlag als Zeilen** — P11
// (#2379),
// Befund C-15. **Was es tut, steht im Haken** (`useSichtvorschlag`), hier
// nur, wie es aussieht — und ein Fehler steht am Ort (§6).
import { useSichtvorschlag } from '../haken/useSichtvorschlag'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Knopf } from './Knopf'

import type { Handelsposition } from '../apihandelsleitung'
import type { Tauschkandidat } from '../apitauschabgleich'
import type { Geholt } from '../haken/useSichtvorschlag'

interface Eigenschaften {
  ich: string
  gegenueber: string
  positionen: Handelsposition[]
  setzePositionen: (positionen: Handelsposition[]) => void

  // geholt ist die Sicht, die der Markttisch schon geholt hat (#2494) —
  // **dann gibt es nichts anzufragen**: der Knopf entfällt, der Abgleich
  // steht da, sobald sie kommt.
  geholt?: Geholt | undefined
}

export function Sichtvorschlag(eigenschaften: Eigenschaften) {
  const texte = useTexte().seiten.handeln.sicht
  const {
    laeuft,
    fehler,
    ergebnis,
    vorgeschlagen,
    uebernommen,
    anfragen,
    uebernehmen,
  } = useSichtvorschlag(eigenschaften)

  const liste = (titel: string, zeilen: Tauschkandidat[]) =>
    zeilen.length > 0 && (
      <div className="flex flex-col gap-1">
        <p className="text-sm font-medium">{titel}</p>
        <ul className="flex flex-col text-sm">
          {zeilen.map((zeile) => (
            <li key={`${zeile.angebotDruck}/${zeile.wunsch}`}>
              {einsetzen(texte.zeile, {
                menge: String(zeile.menge),
                name: zeile.angebot,
              })}
            </li>
          ))}
        </ul>
      </div>
    )

  const abgleich = ergebnis?.abgleich
  const gebe =
    abgleich?.hatVorschlag === true ? abgleich.vorschlag.gebe.zeilen : []
  const bekomme =
    abgleich === undefined
      ? []
      : abgleich.hatVorschlag
        ? abgleich.vorschlag.bekomme.zeilen
        : abgleich.bekomme

  // **Am Markttisch eine Zeile** (#2494) — so hoch wie der Knopf, an dessen
  // Stelle sie steht. Als Kasten schob sie „Vom Gegenüber“ unter den
  // Fensterrand, und ein Zug aus Berts Fach fand sein Ziel nicht mehr
  // (gemessen mit `p12markttisch09` bei 1.440 × 900).
  if (eigenschaften.geholt !== undefined) {
    if (fehler !== undefined) {
      return (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )
    }

    if (ergebnis === undefined) {
      return null
    }

    const inZeile = (titel: string, zeilen: Tauschkandidat[]) =>
      zeilen.length > 0 && (
        <div className="flex items-center gap-1">
          <span className="font-medium">{titel}</span>
          <ul className="flex flex-wrap gap-2">
            {zeilen.map((zeile) => (
              <li key={`${zeile.angebotDruck}/${zeile.wunsch}`}>
                {einsetzen(texte.zeile, {
                  menge: String(zeile.menge),
                  name: zeile.angebot,
                })}
              </li>
            ))}
          </ul>
        </div>
      )

    return (
      <div
        role="group"
        aria-label={texte.titel}
        className="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm"
      >
        {vorgeschlagen.length === 0 ? (
          <p className="text-leise">{texte.keinVorschlag}</p>
        ) : (
          <>
            {inZeile(texte.gibst, gebe)}
            {inZeile(texte.bekommst, bekomme)}
            <Knopf
              gewicht="leise"
              angehalten={uebernommen !== undefined}
              aufKlick={() => {
                uebernehmen(ergebnis)
              }}
            >
              {texte.uebernehmen}
            </Knopf>
          </>
        )}

        {uebernommen !== undefined && (
          <p aria-live="polite" className="text-leise">
            {einsetzen(texte.uebernommen, { zahl: String(uebernommen) })}
          </p>
        )}
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-2">
      <div>
        <Knopf
          gewicht="leise"
          angehalten={laeuft}
          aufKlick={() => {
            void anfragen()
          }}
        >
          {texte.anfragen}
        </Knopf>
      </div>

      {laeuft && <p className="text-sm text-leise">{texte.laeuft}</p>}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {ergebnis !== undefined && (
        <div
          role="group"
          aria-label={texte.titel}
          className="flex flex-col gap-2 rounded-sm border border-kante p-3"
        >
          <p className="text-sm">
            {einsetzen(texte.zeilen, {
              zahl: String(ergebnis.sicht.length),
              wuensche: String(
                ergebnis.sicht.filter((zeile) => zeile.wunsch === true).length,
              ),
            })}
          </p>

          {vorgeschlagen.length === 0 ? (
            <p className="text-sm text-leise">{texte.keinVorschlag}</p>
          ) : (
            <>
              {liste(texte.gibst, gebe)}
              {liste(texte.bekommst, bekomme)}
              <div>
                <Knopf
                  gewicht="haupt"
                  angehalten={uebernommen !== undefined}
                  aufKlick={() => {
                    uebernehmen(ergebnis)
                  }}
                >
                  {texte.uebernehmen}
                </Knopf>
              </div>
            </>
          )}

          {uebernommen !== undefined && (
            <p aria-live="polite" className="text-sm text-leise">
              {einsetzen(texte.uebernommen, { zahl: String(uebernommen) })}
            </p>
          )}
        </div>
      )}
    </div>
  )
}

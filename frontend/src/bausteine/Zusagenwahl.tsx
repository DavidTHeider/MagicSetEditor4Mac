// **Welches deiner Exemplare weggeht — gewählt aus der Sammlung** — UX-08.AK-04
// (#1680).
//
// # Was vorher war
//
// **Drei Felder**: die Kennung einer Besitzzeile, die Stückzahl und ihre
// Fassung — zwei davon Zahlen, die niemand weiß, der nicht in den Bestand
// schaut. Sie stehen jetzt unter „Für Fachleute" in `Handelszusagen`.
//
// # Was jetzt ist
//
// **Je eigener Position die Zeilen der Sammlung, die sie erfüllen** (#2492:
// Druck, Zustand, Sprache und Ausführung — bis dahin jede mit diesem Druck) —
// Lagerort, Zustand, Sprache, Ausführung und Stückzahl, wie die Sammlung sie
// zeigt. Gewählt wird eine Zeile; **Kennung und Fassung kommen aus ihr**, und
// die Stückzahl aus der Position. **Ein Wunsch ist kein Besitz** und steht
// nicht zur Wahl.
import { useId, useState } from 'react'

import { api } from '../fassade'
import { auspacken } from '../umschlag'
import { useLaden } from '../haken/useLaden'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { erfuellt } from '../zusagepassung'
import { Knopf } from './Knopf'
import { Ladefehler, Wirdgeladen } from './Ladehinweise'

import type { Handelsposition } from '../apihandelsleitung'

// wahlstil ist ein Optionsfeld — **mit sichtbarem Fokusring** (§6).
const wahlstil =
  'mt-1 accent-marke focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

interface Zusagenwahleigenschaften {
  position: Handelsposition
  angehalten: boolean
  festhalten: (exemplarID: number, menge: number, fassung: number) => void
}

export function Zusagenwahl({
  position,
  angehalten,
  festhalten,
}: Zusagenwahleigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.handeln
  const sammlung = texte.seiten.sammlung
  const gruppe = useId()

  const [gewaehlt, setzeGewaehlt] = useState<number>()

  const besitz = useLaden(async () => {
    const antwort = await api().SammlungBesitz(position.printing)

    return auspacken(antwort, () => antwort.besitz)
  }, [position.printing])

  const besessen = (besitz.daten?.zeilen ?? []).filter((zeile) => !zeile.wunsch)
  // **Nur was die Zeile erfüllt** (#2492) — der Dienst hielte nichts
  // anderes fest, und ein anderes Exemplar ginge weg, als drüben ankommt.
  const zeilen = besessen.filter((zeile) => erfuellt(zeile, position))
  const zeile = zeilen.find((einer) => einer.id === gewaehlt)

  // **Eine Karte ohne Namen heißt nach ihrem Druck** — gesehen in der
  // WebKit-Probe: ein Exemplar der Saat trug keinen, und die Legende stand
  // als „1 ×" da.
  const bezeichnung = besessen[0]?.bezeichnung ?? ''

  // **Die Wörter der Sammlung** — und der rohe Wert, wo es keines gibt: ein
  // Zustand, den dieses Fenster nicht kennt, ist immer noch eine Auskunft.
  const zustaende = sammlung.zustaende as Record<string, string | undefined>
  const sprachen = texte.bausteine.drucksprachen.namen as Record<
    string,
    string | undefined
  >
  const ausfuehrungen: Record<string, string | undefined> = {
    foil: sammlung.foil,
    nonfoil: sammlung.nonfoil,
    etched: sammlung.etched,
  }

  return (
    <div className="flex flex-col gap-2">
      <fieldset className="flex flex-col gap-1">
        <legend className="mb-1 text-sm">
          {einsetzen(eigene.zusagenPosition, {
            menge: String(position.menge),
            name: bezeichnung === '' ? position.printing : bezeichnung,
          })}
        </legend>

        {besitz.daten === undefined ? (
          besitz.fehler === undefined ? (
            <Wirdgeladen />
          ) : (
            <Ladefehler fehler={besitz.fehler} wiederholen={besitz.nachladen} />
          )
        ) : zeilen.length === 0 ? (
          <p className="text-sm text-leise">
            {besessen.length === 0
              ? eigene.zusagenKeinExemplar
              : eigene.zusagenKeinePassende}
          </p>
        ) : (
          zeilen.map((einer) => (
            <label
              key={einer.id}
              className="flex cursor-pointer items-start gap-2 rounded-sm border border-kante px-2 py-1.5 has-[:checked]:border-marke"
            >
              <input
                type="radio"
                name={gruppe}
                value={einer.id}
                checked={einer.id === gewaehlt}
                onChange={() => {
                  setzeGewaehlt(einer.id)
                }}
                className={wahlstil}
              />
              <span>
                {einsetzen(eigene.zusagenZeile, {
                  lagerort:
                    einer.lagerort === ''
                      ? eigene.zusagenOhneLagerort
                      : einer.lagerort,
                  zustand: zustaende[einer.zustand] ?? einer.zustand,
                  sprache: sprachen[einer.sprache] ?? einer.sprache,
                  finish: ausfuehrungen[einer.finish] ?? einer.finish,
                  stueck: zaehlwort(
                    einer.menge,
                    texte.allgemein.zaehlwoerter.exemplare,
                  ),
                })}
              </span>
            </label>
          ))
        )}

        {/* **Ohne Wahl kein Festhalten** — und ein Knopf, der dann nichts
            tut, bekommt keine Behandlung (`Knopf`, `aufKlick`). **In der
            Gruppe**, denn er gehört zu dieser Wahl und keiner anderen. */}
        <div className="mt-1">
          {zeile === undefined ? (
            <Knopf gewicht="leise" angehalten>
              {eigene.zusagenFesthalten}
            </Knopf>
          ) : (
            <Knopf
              gewicht="leise"
              angehalten={angehalten}
              aufKlick={() => {
                festhalten(zeile.id, position.menge, zeile.fassung)
              }}
            >
              {eigene.zusagenFesthalten}
            </Knopf>
          )}
        </div>
      </fieldset>
    </div>
  )
}

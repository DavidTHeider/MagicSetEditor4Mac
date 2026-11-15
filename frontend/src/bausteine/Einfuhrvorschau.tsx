// **Eine Einfuhr zeigt vorher, was sie zusammenführt** — DOM-03.AK-01
// (#1698).
//
// # Warum es sie gibt
//
// Eine Zeile, die die Sammlung schon führte, wurde beim Einlesen **still
// erhöht**. Wer die eigene Ausfuhr ein zweites Mal einlas, hatte danach jede
// Menge doppelt und erfuhr es aus dem Bericht — hinterher.
//
// # Was sie zeigt
//
// **Nur, wenn etwas zusammengeführt würde** (`braucheVorschau`): eine
// Einfuhr aus lauter neuen Zeilen geht ohne Zwischenschritt durch. Sonst
// stehen die getroffenen Zeilen mit *vorher → nachher* da, und wer einliest,
// wählt: Mengen addieren — wie bisher, und vorgewählt — oder Vorhandene
// auslassen.
import { useId, useState } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'
import { Aktionsknopf } from './Aktionsknopf'
import { Knopf } from './Knopf'

import type {
  SammelvorschauDTO,
  Zusammenfuehrung,
} from '../antwortenlisteneinfuhr'

// braucheVorschau sagt, ob eine Einfuhr etwas zusammenführte — ohne
// Vorschau nicht.
export function braucheVorschau(
  vorschau: SammelvorschauDTO | null,
): vorschau is SammelvorschauDTO {
  return vorschau?.posten.some((posten) => posten.art !== 'neu') ?? false
}

export function Einfuhrvorschau({
  vorschau,
  laeuft,
  aufEinlesen,
  aufAbbrechen,
}: {
  vorschau: SammelvorschauDTO
  laeuft: boolean
  aufEinlesen: (modus: Zusammenfuehrung) => void
  aufAbbrechen: () => void
}) {
  const eigene = useTexte().seiten.sammlung
  const gruppe = useId()
  const [modus, setzeModus] = useState<Zusammenfuehrung>('addieren')

  const zaehle = (art: string) =>
    String(vorschau.posten.filter((posten) => posten.art === art).length)

  const getroffen = vorschau.posten.filter((posten) => posten.art !== 'neu')

  // **Ein Zustand, den der Katalog nicht kennt, steht, wie er kam** — eine
  // Zeile mit neuem Druck geht in der Vorschau ungeprüft durch (sie wird erst
  // beim Einlesen geprüft), und ein leeres Wort sagte weniger als das rohe.
  const zustandsname = (zustand: string) =>
    zustand in eigene.zustaende
      ? eigene.zustaende[zustand as keyof typeof eigene.zustaende]
      : zustand

  return (
    <section
      aria-label={eigene.vorschauTitel}
      className="flex flex-col gap-3 rounded-sm border border-kante p-3"
    >
      <h4 className="text-sm font-medium">{eigene.vorschauTitel}</h4>

      <Erklaerungsabsatz
        text={eigene.vorschauErklaerung}
        klasse="max-w-3xl text-xs text-leise"
      />

      <p className="text-sm tabular-nums">
        {einsetzen(eigene.vorschauZaehlung, {
          neu: zaehle('neu'),
          zusammen: zaehle('zusammen'),
          wiederholt: zaehle('wiederholt'),
        })}
      </p>

      <ul
        aria-label={eigene.vorschauTitel}
        className="flex max-h-48 flex-col gap-1 overflow-auto text-sm"
      >
        {getroffen.map((posten) => (
          <li key={posten.zeile} className="tabular-nums">
            {einsetzen(eigene.vorschauZusammen, {
              bezeichnung: posten.bezeichnung,
              zustand: zustandsname(posten.zustand),
              vorher: String(posten.vorher),
              nachher: String(posten.nachher),
            })}
          </li>
        ))}
      </ul>

      {vorschau.befunde.map((befund) => (
        <p key={befund} className="text-sm text-fehler">
          {befund}
        </p>
      ))}

      <fieldset className="flex flex-col gap-1">
        <legend className="text-sm font-medium">{eigene.vorschauModus}</legend>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="radio"
            name={gruppe}
            checked={modus === 'addieren'}
            disabled={laeuft}
            onChange={() => {
              setzeModus('addieren')
            }}
            className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          {eigene.vorschauAddieren}
        </label>

        <label className="flex items-center gap-2 text-sm">
          <input
            type="radio"
            name={gruppe}
            checked={modus === 'auslassen'}
            disabled={laeuft}
            onChange={() => {
              setzeModus('auslassen')
            }}
            className="focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          />
          {eigene.vorschauAuslassen}
        </label>
      </fieldset>

      <div className="flex flex-wrap gap-2">
        <Aktionsknopf
          aktion="anlegen"
          gewicht="haupt"
          angehalten={laeuft}
          aufKlick={() => {
            aufEinlesen(modus)
          }}
        >
          {eigene.vorschauEinlesen}
        </Aktionsknopf>

        <Knopf angehalten={laeuft} aufKlick={aufAbbrechen}>
          {eigene.vorschauAbbrechen}
        </Knopf>
      </div>
    </section>
  )
}

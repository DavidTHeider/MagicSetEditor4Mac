// **Einen Eintrag aus einer Wortliste wählen.**
//
// Ein eigener Haken aus demselben Grund wie `useFeldschreiben` daneben (§2) —
// und mit denselben Regeln aus §7:
//
//   - **Regel 4**: ein Vorgang je Datensatz, und der Schlüssel ist Karte
//     **und** Feld. Er ist **derselbe** wie beim Tippen: eine Wahl und ein
//     getippter Wert gehen in dasselbe Feld, und beide gleichzeitig gingen
//     gegen einen Stand, den das Backend noch nicht kennt.
//   - **Regel 3**: nachgeladen wird nur nach einem geglückten Schreiben.
//
// **Der neue Wert wird nicht hier gerechnet** — er kommt aus `service`. Wer
// ihn im Fenster zusammensetzte, baute MSEs Auswahllogik ein zweites Mal nach
// (`DropDownWordList::select`, `src/gui/value/text.cpp`, Zeilen 276–297,
// Fassung f9d9356 vom 2021-07-06), mit den Präfixen in der Reihenfolge des
// Klickens statt in der der Liste.
import type { Katalog } from '../sprache/kataloge'
import { api } from '../api'
import { auspacken } from '../umschlag'
import { meldungAus, verursachendesFeld } from '../meldung'
import type { Vorgang } from './useVorgang'

interface Wahlwege {
  setID: string
  karteID: string
  vorgang: Vorgang
  texte: Katalog

  melden: (befund: { feld: string; meldung: string } | undefined) => void
  geglueckt: () => void
}

// useWortwahl gibt die Handlung, die einen Eintrag wählt.
export function useWortwahl({
  setID,
  karteID,
  vorgang,
  texte,
  melden,
  geglueckt,
}: Wahlwege): (
  feld: string,
  bereich: number,
  weg: number[],
  bekannt: readonly string[],
) => Promise<void> {
  return async (
    feld: string,
    bereich: number,
    weg: number[],
    bekannt: readonly string[],
  ) => {
    const ergebnis = await vorgang.starte(`${karteID}/${feld}`, async () => {
      try {
        auspacken(
          await api().WortWaehlen(setID, karteID, feld, bereich, weg, bekannt),
          () => undefined,
        )

        return true
      } catch (ursache: unknown) {
        melden({
          feld: verursachendesFeld(ursache, feld),
          meldung: meldungAus(ursache, texte),
        })

        return false
      }
    })

    if (ergebnis === true) {
      geglueckt()
    }
  }
}

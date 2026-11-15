// Die **Auswahllisten** eines Feldes
// (#1074).
//
// # Was MSE hier tut
//
// Es setzt einen kleinen Pfeil **in** den Text und klappt beim Klick ein Menü
// auf: Kästchen für die Präfixe, Punkte für die Typen
// (`DropDownWordList`, `src/gui/value/text.cpp`, Zeilen 109–300, Fassung
// f9d9356 vom 2021-07-06).
//
// # Warum es hier anders aussieht — `adapted`
//
// Ein Pfeil im Text setzt voraus, dass das Textfeld weiß, wo seine Zeichen
// stehen; ein `<textarea>` weiß das nicht. Wichtiger noch: ein selbstgebautes
// Menü müsste Tastaturbedienung, Fokusring und Ansage für Vorleseprogramme
// nachbauen — und §6 verlangt sie alle drei.
//
// Deshalb steht hier, **was das Menü ist**: ein Kästchen ist eine
// Mehrfachwahl, ein Punkt eine Einfachwahl. Beides gibt es fertig, bedienbar
// und angesagt.
//
// **Die Verschachtelung wird eingerückt, nicht verschachtelt.** Im Pack geht
// sie vier Ebenen tief (`race-de`, gemessen am 2026-08-24); `<optgroup>` kann
// eine. Ein Gruppenkopf ist deshalb ein abgeschalteter Eintrag — er öffnet bei
// MSE ein Untermenü und wird selbst nie gewählt.
import type { WortlistenbereichDTO, WortmenueeintragDTO } from '../formenworte'
import { eingabeklassen } from './Feldeingabe'
import { einsetzen } from '../sprache/einsetzen'
import { useLesegrund } from '../haken/schreibrecht'
import { useTexte } from '../sprache/kontext'

interface WortlistenwahlEigenschaften {
  bereiche: WortlistenbereichDTO[]
  angehalten: boolean

  // waehlen bekommt den Bereich und den Weg zum Eintrag.
  waehlen: (bereich: number, weg: number[]) => void
}

// flacherEintrag ist eine Zeile des Menüs samt ihrem Weg und ihrer Tiefe.
interface flacherEintrag {
  eintrag: WortmenueeintragDTO
  weg: number[]
  tiefe: number
}

export function Wortlistenwahl({
  bereiche,
  angehalten,
  waehlen,
}: WortlistenwahlEigenschaften) {
  const texte = useTexte()

  // **Ein Leser wählt nicht** (#2290; Befund 5 des Zweitblicks über Phase
  // G, zweiter Teil): die Wahl schreibt sofort, wie eine Auswahl im Feld.
  // **Mit Grund** (H01, #2255): jede Stelle sagt einem Leser, warum sie
  // nichts wählt.
  const lesegrund = useLesegrund()

  if (bereiche.length === 0) {
    return null
  }

  return (
    <div className="flex flex-col gap-2">
      {lesegrund.satz}
      {bereiche.map((bereich, nummer) => (
        <Bereichswahl
          key={`${bereich.liste}-${String(nummer)}`}
          bereich={bereich}
          angehalten={angehalten || lesegrund.gesperrt}
          beschriebenVon={lesegrund.beschreibe()}
          beschriftung={einsetzen(texte.seiten.karten.wortlisten.liste, {
            liste: bereich.liste,
          })}
          keine={texte.seiten.karten.wortlisten.keine}
          waehlen={(weg) => {
            waehlen(nummer, weg)
          }}
        />
      ))}
    </div>
  )
}

interface BereichswahlEigenschaften {
  bereich: WortlistenbereichDTO
  angehalten: boolean
  beschriebenVon: string | undefined
  beschriftung: string
  keine: string
  waehlen: (weg: number[]) => void
}

// Bereichswahl ist eine Stelle des Feldes: Kästchen und eine Liste.
function Bereichswahl({
  bereich,
  angehalten,
  beschriebenVon,
  beschriftung,
  keine,
  waehlen,
}: BereichswahlEigenschaften) {
  const flach = flachgelegt(bereich.eintraege)

  // **Die Kästchen stehen nur auf der obersten Ebene.** Eine Wahl in einem
  // Untermenü ersetzt den Bereich **ganz** — so tut MSE es (`select` sieht nur
  // die Einträge seines eigenen Menüs) —, und ein Präfix darin gehörte zu
  // einer Auswahl, die es hier gar nicht gibt. Im echten Pack tritt der Fall
  // nicht auf: die 100 Präfixe stehen alle in Listen ohne Gruppen.
  const kaestchen = flach.filter(
    (zeile) => zeile.tiefe === 0 && zeile.eintrag.istPraefix,
  )

  const punkte = flach.filter((zeile) => !zeile.eintrag.istPraefix)

  return (
    <fieldset className="flex flex-wrap items-center gap-2">
      <legend className="sr-only">{beschriftung}</legend>

      {kaestchen.map((zeile) => (
        <label
          key={zeile.weg.join('.')}
          className="flex items-center gap-1 text-sm"
        >
          <input
            type="checkbox"
            checked={zeile.eintrag.aktiv}
            disabled={angehalten}
            aria-describedby={beschriebenVon}
            onChange={() => {
              waehlen(zeile.weg)
            }}
          />
          {zeile.eintrag.name}
        </label>
      ))}

      {punkte.length > 0 && (
        <select
          aria-label={beschriftung}
          className={eingabeklassen + ' text-sm'}
          value={gewaehlterWeg(punkte)}
          disabled={angehalten}
          aria-describedby={beschriebenVon}
          onChange={(ereignis) => {
            waehlen(ereignis.target.value.split('.').map(Number))
          }}
        >
          {/* **Ein leerer Eintrag, wenn nichts passt.** Ein `select` zeigt
              sonst den ersten Eintrag an, und der sähe aus wie eine Wahl, die
              niemand getroffen hat. */}
          {gewaehlterWeg(punkte) === '' && <option value="">{keine}</option>}

          {punkte.map((zeile) => (
            <option
              key={zeile.weg.join('.')}
              value={zeile.weg.join('.')}
              disabled={zeile.eintrag.kinder !== null}
            >
              {' '.repeat(zeile.tiefe * 2) + zeile.eintrag.name}
            </option>
          ))}
        </select>
      )}
    </fieldset>
  )
}

// flachgelegt macht aus dem Menübaum eine Liste mit Weg und Tiefe.
function flachgelegt(
  eintraege: WortmenueeintragDTO[],
  weg: number[] = [],
  tiefe = 0,
): flacherEintrag[] {
  return eintraege.flatMap((eintrag, nummer) => {
    const hierhin = [...weg, nummer]

    return [
      { eintrag, weg: hierhin, tiefe },
      ...flachgelegt(eintrag.kinder ?? [], hierhin, tiefe + 1),
    ]
  })
}

// gewaehlterWeg gibt den Weg des Eintrags, den der Wert enthält.
//
// **Der erste aktive gewinnt.** Tiefer im Baum kann ein zweiter denselben
// Namen tragen; welcher gemeint ist, entscheidet die Reihenfolge — dieselbe,
// in der MSE sein Menü durchgeht.
function gewaehlterWeg(zeilen: flacherEintrag[]): string {
  const aktiv = zeilen.find((zeile) => zeile.eintrag.aktiv)

  return aktiv === undefined ? '' : aktiv.weg.join('.')
}

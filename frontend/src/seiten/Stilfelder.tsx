// **Die Felder des Stils** — die Zusatzfelder, die der Stil einer Karte
// mitbringt und die sich von Hand setzen lassen
// (#2298).
//
// MSE2 bearbeitet sie auf der Karte selbst, neben ihren Feldern
// (`setData(card->data, &card->extraDataFor(*stylesheet))`,
// `src/render/card/viewer.cpp`, Zeile 139, Fassung f9d9356 vom 2021-07-06).
// Hier ist die Karte keine Eingabefläche; die Felder stehen deshalb in einem
// eigenen Kasten unter dem Formular — **mit denselben Bauteilen** wie die
// Kartenfelder ([Feldzeile]), denn die Form ist dieselbe.
//
// **Ein Kasten, kein zweites Formular dahinter**: ohne Felder steht nichts da.
// Bild- und Symbolfelder fehlen — sie brauchen Wege, die an einem Kartenfeld
// hängen (Zuschnitt, Zeichner), und kein Stil des Packs bringt eines mit, das
// sich von Hand setzen ließe.
import { api } from '../api'
import { useLaden } from '../haken/useLaden'
import { useZusatzschreiben } from '../haken/useZusatzschreiben'
import type { Vorgang } from '../haken/useVorgang'
import type { Katalog } from '../sprache/kataloge'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Feldzeile } from './Feldzeile'
import type { Feldfehler } from './feldkennung'

interface StilfelderEigenschaften {
  // wege sind dieselben wie die des Kartenformulars — Vorgang, Meldung und
  // was nach einem geglückten Schreiben geschieht.
  wege: {
    setID: string
    karteID: string
    vorgang: Vorgang
    texte: Katalog
    melden: (befund: Feldfehler | undefined) => void
    geglueckt: () => void
  }

  // stand steigt, wenn sich die Karte anderswo geändert hat.
  stand: number

  fehler: Feldfehler | undefined
}

const ohneBild = new Set(['image', 'symbol'])

export function Stilfelder({ wege, stand, fehler }: StilfelderEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten

  const geladen = useLaden(async () => {
    const antwort = await api().Kartenzusatzfelder(wege.setID, wege.karteID)

    return auspacken(antwort, () => antwort.felder)
  }, [wege.setID, wege.karteID, stand])

  const schreiben = useZusatzschreiben({
    ...wege,
    geglueckt: () => {
      geladen.nachladen()
      wege.geglueckt()
    },
  })

  const felder = (geladen.daten ?? []).filter(
    (feld) => !ohneBild.has(feld.form.art),
  )

  // **„Keine Felder" und „nicht lesbar" sind zwei Antworten** (Befund 7 des
  // Nachlaufs über E/F): ohne Meldung verschwand der Kasten bei einem
  // Lesefehler wie bei einem Stil ohne Felder, und ein gescheitertes
  // Nachladen nach dem Schreiben blieb ungesagt.
  if (felder.length === 0 && geladen.fehler === undefined) {
    return null
  }

  return (
    <section
      aria-label={eigene.stilfelder}
      className="flex w-full flex-col gap-1 rounded-sm border border-kante px-3 py-2"
    >
      <h3 className="text-sm font-medium">{eigene.stilfelder}</h3>

      <p className="max-w-prose text-sm text-leise">{eigene.stilfelderHilfe}</p>

      {geladen.fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {texte.bausteine.zustaende.fehler} {geladen.fehler}
        </p>
      )}

      <form
        className="mt-2 flex flex-col gap-3"
        onSubmit={(ereignis) => {
          ereignis.preventDefault()
        }}
      >
        {felder.map((feld) => (
          <Feldzeile
            key={feld.form.name}
            feld={feld.form.name}
            wert={feld.wert}
            form={feld.form}
            fehler={
              fehler?.feld === feld.form.name ? fehler.meldung : undefined
            }
            laeuft={wege.vorgang.laeuft(`${wege.karteID}/extra data`)}
            palette={undefined}
            wortlisten={[]}
            hergeleitet={undefined}
            gewaehlt={nichts}
            getippt={nichts}
            schreiben={(wert) => {
              void schreiben(feld.form.name, wert)
            }}
            einfuegen={nichts}
            bildWaehlen={nichts}
            wortWaehlen={nichts}
          />
        ))}
      </form>
    </section>
  )
}

// nichts ist der Rückruf für Wege, die ein Zusatzfeld nicht hat — Palette,
// Wortlisten, Bild und die getippte Vorschau gehören zu einem Kartenfeld.
function nichts() {
  return undefined
}

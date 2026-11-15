// Die **Symbolpalette** eines Feldes
// (#564).
//
// # Wofür
//
// Im Feld für die Manakosten steht `2RR`. Auf der Karte werden daraus Symbole;
// im Feld bleiben es Buchstaben, und wer nicht weiß, dass `T` das Tap-Symbol
// ist, findet es nicht. MSE bietet dafür ein Menü an, und es steht **im
// Paket** — `insert symbol menu:` in der `.mse-symbol-font`.
//
// # Als Palette und nicht als Menü
//
// MSE hängt die Einträge in ein Kontextmenü, weil ein wxWidgets-Textfeld keine
// Kacheln neben sich stellen kann. Hier stehen sie unter dem Feld: sichtbar,
// ohne Klick, und mit dem Bild, das eingefügt wird. Der Baum aus dem Paket
// bleibt derselbe — ein Untermenü wird eine Gruppe.
import type { PalettenEintragDTO } from '../api'
import { useLesegrund } from '../haken/schreibrecht'
import { useTexte } from '../sprache/kontext'

interface SymbolpaletteEigenschaften {
  eintraege: readonly PalettenEintragDTO[]

  // einfuegen bekommt den **rohen** Code. Ob er ausgezeichnet wird, entscheidet
  // der Aufrufer: `always symbol:` ist eine Aussage über das Feld, nicht über
  // das Symbol.
  einfuegen: (code: string) => void

  angehalten: boolean
}

export function Symbolpalette({
  eintraege,
  einfuegen,
  angehalten,
}: SymbolpaletteEigenschaften) {
  const texte = useTexte()

  // **Ein Leser fügt nichts ein** (#2290; Befund 5 des Zweitblicks über
  // Phase G, zweiter Teil): die Kachel übernähme einen neuen Entwurf in die
  // Vorschau eines Felds, das für ihn schreibgeschützt ist.
  // **Mit Grund** (H01, #2255): jede Kachel sagt einem Leser, warum sie
  // nichts einfügt.
  const lesegrund = useLesegrund()
  const gesperrt = angehalten || lesegrund.gesperrt

  if (eintraege.length === 0) {
    return null
  }

  return (
    <div
      role="group"
      aria-label={texte.seiten.karten.symbole}
      className="flex flex-wrap items-center gap-1"
    >
      {lesegrund.satz}
      {eintraege.map((eintrag, stelle) => (
        <Eintrag
          // **Die Stelle als Schlüssel**, denn ein Trenner hat keinen Code und
          // ein Paket darf denselben Code zweimal anbieten — `magic-mana-large`
          // führt `T` dreimal, je unter einer anderen Bedingung.
          key={`${eintrag.art}-${eintrag.code}-${String(stelle)}`}
          eintrag={eintrag}
          einfuegen={einfuegen}
          angehalten={gesperrt}
          beschriebenVon={lesegrund.beschreibe()}
        />
      ))}
    </div>
  )
}

function Eintrag({
  eintrag,
  einfuegen,
  angehalten,
  beschriebenVon,
}: {
  eintrag: PalettenEintragDTO
  einfuegen: (code: string) => void
  angehalten: boolean
  beschriebenVon: string | undefined
}) {
  // **Ein Trenner ist eine Linie, kein Zwischenraum.** Er sagt im Paket, dass
  // hier eine Gruppe endet — `magic-mana-large` trennt so die Sondersymbole von
  // den fünf Farben.
  if (eintrag.art === 'line') {
    return <span aria-hidden="true" className="mx-1 h-5 w-px bg-kante" />
  }

  if (eintrag.art === 'submenu') {
    return (
      <span
        role="group"
        aria-label={eintrag.beschriftung}
        className="flex flex-wrap items-center gap-1 rounded-sm border border-kante px-1"
      >
        {eintrag.kinder.map((kind, stelle) => (
          <Eintrag
            key={`${kind.art}-${kind.code}-${String(stelle)}`}
            eintrag={kind}
            einfuegen={einfuegen}
            angehalten={angehalten}
            beschriebenVon={beschriebenVon}
          />
        ))}
      </span>
    )
  }

  // **Ein `custom` steht noch nicht.** MSE öffnet dafür ein Eingabefeld
  // (`wxGetTextFromUser`, `InsertSymbolMenu::getCode`, Zeile 462, Fassung
  // f9d9356 vom 2021-07-06); über das ganze Pack sind es 62 von 5.219
  // Einträgen. Sie wegzulassen ist ehrlicher als eine Schaltfläche, die etwas
  // anderes einfügt als das, was dransteht.
  if (eintrag.art !== 'code') {
    return null
  }

  return (
    <button
      type="button"
      aria-label={eintrag.beschriftung}
      title={eintrag.beschriftung}
      disabled={angehalten}
      aria-describedby={beschriebenVon}
      onClick={() => {
        einfuegen(eintrag.code)
      }}
      className="flex h-7 w-7 items-center justify-center rounded-sm border border-kante bg-flaeche hover:bg-betont focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring disabled:opacity-50"
    >
      {/* **Ein leeres Bild bleibt leer, der Code steht daneben.** Ein Menü
          nennt Codes, die eine Bedingung gerade weggenommen hat; MSE zeigt
          dann einen Eintrag ohne Kachel (`makeMenuItem`, Zeilen 508–518). */}
      {eintrag.bild === '' ? (
        <span aria-hidden="true" className="text-xs">
          {eintrag.code}
        </span>
      ) : (
        <img src={eintrag.bild} alt="" className="h-5 w-5 object-contain" />
      )}
    </button>
  )
}

// **Ein Stil aus vielen, als Auswahl mit Suche** — UXR-07
// (#1915),
// nach dem Artboard *Eigenschaften*: ein Auswahlfeld „Stilbogen“, gesucht
// „nach sichtbarem Namen **und** technischer Kennung“ (UX-05, S07.NAT-02,
// #1677).
//
// # Wo sie steht
//
// In der Lasche *Stil* eines offenen Sets **und** im zweiten Schritt eines
// neuen Sets — dieselbe Frage an zwei Stellen, und zwei Gestalten liefen
// auseinander. Befund F07 der nativen Sichtprüfung: das Stilmenü führte
// **355 technische Kennungen** und keinen Namen; deshalb steht in jeder Zeile
// der Name vorn und die Kennung dahinter.
//
// # Warum keine Optionsfelder mehr — und seit UXR-F1 auch kein natives Feld
//
// **Bis UXR-04 zeichnete die `Stilliste` je Stil eines.** In der Schublade
// *Eigenschaften* waren das gemessen **382 Felder**, zehn davon im Bild — R3
// aus `docs/UIREVIEW.md` erlaubt zwölf; im neuen Set mit dem Full-Magic-Pack
// **355** (UXR-04, #1912).
// R7 sagt, was an ihre Stelle gehört: über fünfzehn Werte eine **Auswahl mit
// Suche**.
//
// **UXR-07 baute sie aus einem Suchfeld und einem nativen `select`**, weil es
// keinen Baustein gab, der in WKWebView belegt war. Den gibt es seit UXR-F1
// (#2007):
// die `Suchauswahl`. Sie ist **ein** Feld statt zweier, und die Suche steht
// dort, wo gewählt wird — nicht darüber, getrennt von der Liste, die sie
// eingrenzt.
import { useId } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { Suchauswahl } from './Suchauswahl'

// Stileintrag ist ein Stil, wie die Auswahl ihn zeigt.
export interface Stileintrag {
  // kennung ist der Name des Pakets ohne Endung — `magic-m15`.
  kennung: string

  // anzeigename kommt fertig aus Go (`domain.Paket.Anzeigename`).
  anzeigename: string

  // zusatz steht hinter der Kennung — etwa das fremde Spiel eines Stils.
  zusatz?: string
}

interface StilauswahlEigenschaften {
  // beschriftung benennt das Auswahlfeld.
  beschriftung: string
  eintraege: readonly Stileintrag[]

  // gewaehlt ist die Kennung des gewählten Stils — leer, wenn das Set
  // keinen der installierten führt.
  gewaehlt: string
  waehlen: (kennung: string) => void
}

export function Stilauswahl({
  beschriftung,
  eintraege,
  gewaehlt,
  waehlen,
}: StilauswahlEigenschaften) {
  const worte = useTexte().bausteine.stilliste
  const kennung = useId()

  return (
    <div className="flex max-w-md flex-col gap-2">
      <label htmlFor={kennung} className="text-sm font-medium">
        {beschriftung}
      </label>

      {/* **Ohne Wahl ein leeres Feld mit Platzhalter** — nicht der erste
          Stil: das Set sähe sonst aus, als führte es ihn. */}
      <Suchauswahl
        id={kennung}
        bezeichnung={beschriftung}
        eintraege={eintraege.map((eintrag) => ({
          wert: eintrag.kennung,
          name: eintrag.anzeigename,
          zusatz:
            eintrag.zusatz === undefined
              ? eintrag.kennung
              : einsetzen(worte.mitZusatz, {
                  kennung: eintrag.kennung,
                  zusatz: eintrag.zusatz,
                }),
        }))}
        wert={gewaehlt}
        aendern={waehlen}
        platzhalter={worte.keinerGewaehlt}
      />
    </div>
  )
}

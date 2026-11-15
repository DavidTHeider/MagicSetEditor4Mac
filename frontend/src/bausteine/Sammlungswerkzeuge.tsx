// **Was man mit einer Sammlung tut, in zwei Schubladen** — UXR-10
// (#1918),
// nach dem Artboard `Sammlung`: *+ Aufnehmen* und *Einlesen und ausgeben*
// stehen im Kopf der Seite, und jeder öffnet seine Schublade.
//
// # Was vorher war
//
// **Bis Phase 4.10 standen die Werkzeuge über der Liste** (#1325) — zehn von
// elf Textfeldern über der Falz. Danach standen sie in einem Kasten im
// Navigator, **bei leerer Sammlung offen**: gemessen am 2026-10-02 am
// nativen Fenster achtzehn Felder in 240 Punkten (R3 in
// `docs/UIREVIEW.md`). Und mit dem ersten Exemplar klappte der Kasten zu —
// samt dem Bericht der Einfuhr, die ihn gefüllt hatte.
//
// # Gezeichnet beim ersten Öffnen, danach nur verborgen
//
// **Wie die Schubladen des Sets** (`Setbereich.tsx`): eine halb getippte
// Eingabe, eine Deutung, ein Bericht überstehen das Schließen. Was eine
// Schublade fragt — der Modellkatalog der Erkennung —, fragt sie erst, wenn
// jemand sie aufzieht.
//
// # Was in welcher steht
//
// **Aufnehmen**: das Formular, die Frage nach einer Dublette und die
// Erkennung über ein Foto — sie beantwortet dieselbe Frage (welcher Druck
// ist das?) und nimmt so wenig auf wie die Deutung einer Liste (A13.3,
// #1452). **Einlesen und ausgeben**: die Listeneinfuhr und, eingeklappt, der
// Rückweg als CSV (#1292) — er liest auch wieder ein.
import { Activity, useState, type ReactNode } from 'react'

import { Aufklappkasten } from './Aufklappkasten'
import { KIKartenerkennung } from './KIKartenerkennung'
import { Listeneinfuhr } from './Listeneinfuhr'
import { Sammlungsausfuhr } from './Sammlungsausfuhr'
import { Sammlungsformular } from './Sammlungsformular'
import { Schublade } from './Schublade'
import { useTexte } from '../sprache/kontext'
import type { ExemplarDTO } from '../api'

// Werkzeug ist eine der zwei Schubladen.
export type Werkzeug = 'aufnehmen' | 'einfuhr'

interface WerkzeugEigenschaften {
  // offen ist die Schublade, die steht — keine, solange niemand eine öffnet.
  offen: Werkzeug | undefined
  schliessen: () => void

  // meldung ist der Fehler der Seite. **Er steht in der Schublade**, solange
  // eine steht: darunter sähe ihn niemand, und ein gescheitertes Aufnehmen
  // meldet sich dort, wo geklickt wurde (§6).
  meldung?: ReactNode

  anlegen: (exemplar: ExemplarDTO) => Promise<boolean>
  laeuft: boolean
  nachladen: () => void

  // geleert zählt, wie oft das Formular von außen geräumt wurde — nach einer
  // zusammengeführten Dublette (DOM-03); entscheidung steht darunter.
  geleert: number
  entscheidung?: ReactNode
}

// sichtbar macht aus einem Wahrheitswert die Art einer `<Activity>`.
const sichtbar = (ja: boolean) => (ja ? 'visible' : 'hidden')

export function Sammlungswerkzeuge({
  offen,
  schliessen,
  meldung,
  anlegen,
  laeuft,
  nachladen,
  geleert,
  entscheidung,
}: WerkzeugEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.sammlung

  // **Beim Zeichnen gesetzt und nicht in einem Effekt**, wie im
  // `Aufklappkasten`: sonst stünde die Schublade erst einen Zug später da.
  const [gezeichnet, setzeGezeichnet] = useState<readonly Werkzeug[]>([])
  if (offen !== undefined && !gezeichnet.includes(offen)) {
    setzeGezeichnet([...gezeichnet, offen])
  }

  const inhalt: Record<Werkzeug, { titel: string; teile: ReactNode }> = {
    aufnehmen: {
      titel: eigene.neuesTitel,
      teile: (
        <>
          <Sammlungsformular
            aufAufnehmen={anlegen}
            laeuft={laeuft}
            geleert={geleert}
          />

          {/* **Die Frage steht, wo geklickt wurde** (§6): unter dem
              Formular, nicht in einem Dialog darüber. */}
          {entscheidung}

          <KIKartenerkennung />
        </>
      ),
    },
    einfuhr: {
      titel: eigene.einfuhrTitel,
      teile: (
        <>
          <Listeneinfuhr aufAufgenommen={nachladen} />

          {/* **Der Rückweg eingeklappt** — offen ist eine Gruppe, nicht
              zwei (R3). */}
          <Aufklappkasten titel={eigene.ausfuhrTitel}>
            <Sammlungsausfuhr aufEingelesen={nachladen} />
          </Aufklappkasten>
        </>
      ),
    },
  }

  return gezeichnet.map((werkzeug) => (
    <Activity key={werkzeug} mode={sichtbar(werkzeug === offen)}>
      <Schublade
        titel={inhalt[werkzeug].titel}
        schliessenBezeichnung={
          texte.allgemein.arbeitsplatz.schubladen.schliessen
        }
        schliessen={schliessen}
      >
        <div className="flex flex-col gap-4 px-4 pb-5 pt-3">
          {werkzeug === offen && meldung}
          {inhalt[werkzeug].teile}
        </div>
      </Schublade>
    </Activity>
  ))
}

// **Der zweite Schritt des Ausgabeassistenten: wie?** — UXR-09
// (#1917),
// nach dem Artboard *Ausgeben*: rechts die Fläche „Umfang und Maße" des
// gewählten Ziels, und **nur** seine.
//
// **Was darin steht, gab es schon** — die Bildausgabe, den Druckbogen, die
// Exportvorlage, MTG JSON. Neu ist der Ort: bis hierher standen zwei davon
// auf eigenen Laschen und zwei auf dieser, und die Einstellungen des
// Druckbogens standen da, ob man drucken wollte oder nicht. Jede Fläche hat
// **eine** führende Handlung — den Knopf, der schreibt (R1).
import { useId, type ReactNode } from 'react'

import type { SetDTO } from '../api'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import type { Vorgang } from '../haken/useVorgang'
import { useTexte } from '../sprache/kontext'
import type { Ausgabeziel } from '../ausgabe'
import { zielerklaerung, zielname } from './ausgabetexte'
import { Bilder } from './Bilder'
import { Druckbogen } from './Druckbogen'
import { Ladefeld } from './Ladefeld'
import { MTGJSONSichern } from './MTGJSONSichern'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'

interface AusgabeflaecheEigenschaften {
  ziel: Ausgabeziel
  setID: string

  // satz ist das Set, solange es geladen ist — Exportvorlage und MTG JSON
  // brauchen es ganz: Name, Spiel, Herkunft.
  satz: SetDTO | undefined
  laedt: boolean
  fehler: string | undefined

  // vorgang ist der des Assistenten — §7 Regel 4 je Set, nicht je Fläche.
  vorgang: Vorgang
}

export function Ausgabeflaeche({
  ziel,
  setID,
  satz,
  laedt,
  fehler,
  vorgang,
}: AusgabeflaecheEigenschaften) {
  const texte = useTexte()
  const titel = useId()

  // mitSatz baut, was das Set braucht — und zeigt bis dahin, dass es kommt.
  const mitSatz = (bau: (satz: SetDTO) => ReactNode) =>
    satz === undefined ? (
      <Ladefeld laedt={laedt} fehler={fehler} leer={false}>
        <></>
      </Ladefeld>
    ) : (
      bau(satz)
    )

  const inhalt = {
    // **Die Setdatei schreibt ⌘S** — ein zweiter Sichernknopf hier wäre ein
    // zweiter Vorgang für dasselbe Set neben dem der Setliste und der
    // Menüleiste (§7, Regel 4). Die Fläche sagt, was hineinkommt und was
    // nicht.
    setdatei: null,
    bilder: <Bilder setID={setID} />,
    druckbogen: <Druckbogen setID={setID} />,
    vorlage: mitSatz((gesetzt) => (
      <SetVorlagenausfuhr satz={gesetzt} imAssistenten />
    )),
    mtgjson: mitSatz((gesetzt) => (
      <MTGJSONSichern satz={gesetzt} vorgang={vorgang} />
    )),
  }

  return (
    <section
      aria-labelledby={titel}
      className="flex flex-col gap-3 rounded-lg border border-kante bg-flaeche p-5"
    >
      <h3 id={titel} className="text-lg font-semibold">
        {zielname(texte, ziel)}
      </h3>

      <Erklaerungsabsatz text={zielerklaerung(texte, ziel)} />

      {inhalt[ziel]}
    </section>
  )
}

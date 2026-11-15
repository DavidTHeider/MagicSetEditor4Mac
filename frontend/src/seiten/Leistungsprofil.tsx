// **Ein Leistungsprofil, gezielt** — OPS-05.AK-03
// (#1722).
//
// **Nur an diesem Fenster**, und warum, steht in
// `internal/app/leistungsprofil.go`: ein Profil misst den Prozess, in dem es
// läuft. Die Dauer ist eine Zahl von 1 bis 60; was Go dazu sagt, steht am
// Knopf, wenn es nicht geht.
import { useId, useState } from 'react'

import { api } from '../fassade'
import { Knopf } from '../bausteine/Knopf'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { alsGroesse } from './Diagnosepaket'

export function Leistungsprofil() {
  const texte = useTexte()
  const eigene = texte.seiten.technik.leistungsprofil
  const darstellung = useDarstellung()
  const vorgang = useVorgang()

  const [sekunden, setzeSekunden] = useState(10)
  const [ergebnis, setzeErgebnis] = useState<string>()
  const [fehler, setzeFehler] = useState<string>()

  const aufzeichnen = () =>
    vorgang.starte('leistungsprofil', async () => {
      try {
        const antwort = await api().Leistungsprofil(sekunden)
        const profil = auspacken(antwort, () => antwort)

        setzeErgebnis(
          einsetzen(eigene.fertig, {
            datei: profil.datei,
            groesse: alsGroesse(profil.bytes, darstellung),
          }),
        )
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        setzeFehler(meldungAus(ursache, texte))
      }
    })

  const laeuft = vorgang.laeuft('leistungsprofil')

  // **Die Grenze steht am Feld** (§6): Go prüft dieselbe noch einmal, aber
  // wer 0 tippt, soll es sehen, bevor er drückt.
  const gueltig = Number.isInteger(sekunden) && sekunden >= 1 && sekunden <= 60
  const hinweis = useId()

  return (
    <section className="flex flex-col gap-3 border-t border-kante pt-4">
      <h3 className="text-lg font-semibold">{eigene.ueberschrift}</h3>

      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      <div className="flex flex-wrap items-end gap-2">
        <label className="flex flex-col gap-1 text-sm">
          {eigene.sekunden}
          <input
            type="number"
            min={1}
            max={60}
            value={sekunden}
            aria-invalid={!gueltig}
            aria-describedby={hinweis}
            onChange={(ereignis) => {
              setzeSekunden(Number(ereignis.target.value))
            }}
            className="w-24 rounded-sm border border-kante bg-flaeche px-2 py-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-ring"
          />
        </label>

        <p id={hinweis} className="text-sm text-leise">
          {eigene.grenze}
        </p>

        <Knopf
          bezeichnung={eigene.aufzeichnen}
          angehalten={laeuft || !gueltig}
          aufKlick={() => {
            void aufzeichnen()
          }}
        >
          {eigene.aufzeichnen}
        </Knopf>
      </div>

      {laeuft && <p className="text-sm text-leise">{eigene.laeuft}</p>}

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {ergebnis !== undefined && (
        <p className="font-mono text-sm break-all">{ergebnis}</p>
      )}
    </section>
  )
}

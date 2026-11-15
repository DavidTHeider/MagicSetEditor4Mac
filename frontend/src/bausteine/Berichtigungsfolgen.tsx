// **Was eine Berichtigung verschöbe** — DOM-04.AK-02
// (#1699):
// *„Ergebniskorrekturen zeigen Auswirkungen auf folgende Runden und
// Ranglisten."*
//
// Muster *Rückfrage* aus `docs/UX-ARCHITEKTUR.md`: die Frage steht an der
// Paarung, an der geklickt wurde — wer platzt, wer steigt, und welche Runden
// nach dem alten Stand gepaart wurden und bleiben.
import { Entscheidung } from './Entscheidung'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import type { KorrekturfolgenDTO } from '../apiturniere'

interface FolgenEigenschaften {
  folgen: KorrekturfolgenDTO
  laeuft: boolean
  aufEintragen: () => void
  aufVerwerfen: () => void
}

export function Berichtigungsfolgen({
  folgen,
  laeuft,
  aufEintragen,
  aufVerwerfen,
}: FolgenEigenschaften) {
  const eigene = useTexte().seiten.turniere

  return (
    <Entscheidung
      frage={eigene.folgenFrage}
      bestaetigen={eigene.folgenEintragen}
      abbrechen={eigene.folgenNicht}
      laeuft={laeuft}
      aufBestaetigen={aufEintragen}
      aufAbbrechen={aufVerwerfen}
    >
      {folgen.verschoben.length > 0 && (
        <ul className="flex flex-col gap-1 text-sm">
          {folgen.verschoben.map((wechsel) => (
            <li key={wechsel.teilnehmerID}>
              {einsetzen(eigene.folgenPlatz, {
                name: wechsel.name,
                vorher: String(wechsel.platzVorher),
                nachher: String(wechsel.platzNachher),
                punkteVorher: String(wechsel.punkteVorher),
                punkteNachher: String(wechsel.punkteNachher),
              })}
            </li>
          ))}
        </ul>
      )}

      {folgen.spaetereRunden.length > 0 && (
        <p className="text-sm text-leise">
          {einsetzen(eigene.folgenSpaeter, {
            runden: folgen.spaetereRunden.join(', '),
          })}
        </p>
      )}
    </Entscheidung>
  )
}

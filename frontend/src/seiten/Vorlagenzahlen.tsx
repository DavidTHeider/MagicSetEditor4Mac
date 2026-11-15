// **Welche Datei daraus wird** — der Inhalt von „Wie es wird" für die
// Exportvorlage, UXR-F7
// (#2015).
//
// Der Name und die Größe der Hauptdatei — **genau**, denn die Vorlage ist
// ausgewertet —, die Karten, über die sie lief, und was die Vorlage daneben
// ablegt. Die Größe der Dateien daneben ist eine Untergrenze, wo ein Bild
// erst beim Rastern entsteht; die Anzeige sagt es.
import type { VorlagenprobelaufAntwort } from '../antwortenprobelauf'
import { einsetzen } from '../sprache/einsetzen'
import { useDarstellung, useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { alsGroesse } from './Diagnosepaket'

interface VorlagenzahlenEigenschaften {
  probe: VorlagenprobelaufAntwort
}

export function Vorlagenzahlen({ probe }: VorlagenzahlenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.ausgeben.probelauf.vorlage
  const darstellung = useDarstellung()
  const beigaben = alsGroesse(probe.beigabenBytes, darstellung)

  return (
    <>
      <p className="font-medium">
        {einsetzen(eigene.datei, {
          name: probe.dateiname,
          groesse: alsGroesse(probe.bytes, darstellung),
        })}
      </p>

      <p className="text-sm text-leise">
        {einsetzen(eigene.karten, {
          karten: zaehlwort(probe.karten, texte.allgemein.zaehlwoerter.karten),
        })}
      </p>

      {probe.beigaben > 0 && (
        <p className="text-sm text-leise">
          {einsetzen(eigene.beigaben, {
            zahl: String(probe.beigaben),
            ordner: probe.verzeichnis,
            groesse: probe.geschaetzt
              ? einsetzen(eigene.mindestens, { groesse: beigaben })
              : beigaben,
          })}
        </p>
      )}

      {probe.befunde.length > 0 && (
        <>
          <p className="text-sm text-leise">{eigene.befunde}</p>
          <ul className="space-y-1 font-mono text-sm">
            {probe.befunde.map((befund) => (
              <li key={befund}>{befund}</li>
            ))}
          </ul>
        </>
      )}
    </>
  )
}

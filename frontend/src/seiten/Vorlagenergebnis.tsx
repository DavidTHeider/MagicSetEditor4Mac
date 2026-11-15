// **Was eine Vorlage hinterlassen hat** — aus `SetVorlagenausfuhr.tsx`
// gelöst, als die Wahl der Vorlage eine Suche bekam (UXR-09,
// #1917):
// die Datei liefe sonst über die Zeilengrenze aus §2. Der Schnitt folgt der
// Frage — dort, was die Ausfuhr **tut**, hier, was danach **dasteht**:
// Fehler, Befunde, der Laufstand und die geschriebene Datei.
import { Dateiverweis } from '../bausteine/Dateiverweis'
import { Laufstand } from '../bausteine/Laufstand'
import { useTexte } from '../sprache/kontext'
import type { VorlagenausfuhrDTO } from '../vorlagen'

interface VorlagenergebnisEigenschaften {
  fehler: string | undefined

  // gescheitert sind die Befunde eines gescheiterten Laufs — der Grund.
  gescheitert: readonly string[]

  // befunde sind die Befunde beim Lesen der Vorlagen.
  befunde: readonly string[]

  ausfuhr: VorlagenausfuhrDTO | undefined
}

export function Vorlagenergebnis({
  fehler,
  gescheitert,
  befunde,
  ausfuhr,
}: VorlagenergebnisEigenschaften) {
  const texte = useTexte()

  return (
    <>
      {/* **Warum eine Vorlage nicht durchläuft, gehört hierher** (#209): sie
        ist der Normalfall, nicht die Ausnahme. */}
      {fehler !== undefined && (
        <div role="alert" className="mt-3 text-fehler">
          <p>{texte.seiten.sets.vorlageGescheitert}</p>
          <p className="mt-1 font-mono text-sm">{fehler}</p>

          {/* **Der Grund kommt als Befund**, nicht als übersetzte Meldung:
            „der Name … ist hier nicht bekannt" ist eine Aussage über die
            Vorlage, und ohne sie stünde nur „Die Eingabe ist nicht gültig".
            Siehe `app.schreibeVorlage`. */}
          {gescheitert.length > 0 && (
            <ul className="mt-1 space-y-1 font-mono text-sm">
              {gescheitert.map((grund) => (
                <li key={grund}>{grund}</li>
              ))}
            </ul>
          )}
        </div>
      )}

      {befunde.length > 0 && (
        <ul className="mt-3 space-y-1 font-mono text-sm">
          {befunde.map((befund) => (
            <li key={befund}>{befund}</li>
          ))}
        </ul>
      )}

      <Laufstand art="vorlage" />

      {ausfuhr && (
        <div role="status" className="mt-3">
          <p className="text-leise">
            {`${texte.seiten.sets.vorlageGeschrieben} ${ausfuhr.pfad}`}
          </p>

          <Dateiverweis pfad={ausfuhr.pfad} />

          {ausfuhr.befunde.length > 0 && (
            <>
              <p className="mt-2 max-w-prose text-leise">
                {texte.seiten.sets.vorlageBefunde}
              </p>
              <ul className="mt-1 space-y-1 font-mono text-sm">
                {ausfuhr.befunde.map((befund) => (
                  <li key={befund}>{befund}</li>
                ))}
              </ul>
            </>
          )}
        </div>
      )}
    </>
  )
}

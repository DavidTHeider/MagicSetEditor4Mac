// **Was am Tisch auffiel, ohne ein Fehler zu sein** — herausgelöst am
// 2026-09-24 aus `Partietisch.tsx`, als die Anordnung nach dem Artboard
// *Tisch* (Befund F09) die Datei über die Zeilengrenze aus §2 trug.
//
// # Der Schnitt folgt der Frage
//
// `Partietisch.tsx` ordnet an, was während des Spiels gebraucht wird. Hier
// steht, **was dabei nebenher auffiel**: ein Sitz, dessen Profil es nicht
// mehr gibt, und eine Ressource, die sich nicht zeichnen ließ.
import { useTexte } from '../sprache/kontext'

interface Befundeigenschaften {
  // sitzbefunde sind die Befunde der Partie über ihre Sitze (G8, #1465).
  sitzbefunde: readonly string[]

  // texturfehler und texturbefunde kommen aus dem Speicher der Bilder
  // (G2, #1459).
  texturfehler: string | undefined
  texturbefunde: readonly string[]
}

export function Tischbefunde({
  sitzbefunde,
  texturfehler,
  texturbefunde,
}: Befundeigenschaften) {
  const eigene = useTexte().seiten.spielen

  return (
    <>
      {/* **Ein Befund ist kein Fehler** (G8,
          #1465):
          ein Sitz, dessen Profil es nicht mehr gibt, bleibt ein Sitz.
          Gespielt wird weiter, und die Bindung wird benannt statt
          verschwiegen (§8). Deshalb `status` und nicht `alert`. */}
      {sitzbefunde.length > 0 && (
        <div role="status" className={befundlisteklassen}>
          <p>{eigene.sitzprofilBefunde}</p>

          <ul className="flex flex-col gap-1">
            {sitzbefunde.map((befund) => (
              <li key={befund} className={befundklassen}>
                {befund}
              </li>
            ))}
          </ul>
        </div>
      )}

      {texturfehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {texturfehler}
        </p>
      )}

      {/* **Was beim Zeichnen auffiel, steht da** — eine fehlende Ressource
          hat eine sichtbare Diagnose (#1459). Je Satz einmal: sechzig
          Karten mit derselben fehlenden Schrift sind ein Befund. */}
      {/* **Die Rolle am Rahmen, nicht an der Liste** (P4, #2372, axe in
          WebKit): eine `ul` mit fremder Rolle ist keine Liste mehr. */}
      {texturbefunde.length > 0 && (
        <div role="status" className={befundlisteklassen}>
          <ul className="flex flex-col gap-1">
            {texturbefunde.map((befund) => (
              <li key={befund} className={befundklassen}>
                {befund}
              </li>
            ))}
          </ul>
        </div>
      )}
    </>
  )
}

// befundlisteklassen und befundklassen halten eine Diagnose im Rahmen.
//
// # Wofür die Klassen da sind
//
// **Weil ein Befund lang sein kann — sehr lang.** Gesehen in der
// Sichtprüfung am 2026-09-15 an `magic.mse-game`: „init script des Spiels,
// Zeile 8425 … match timeout after 1s on input `Altmetall|Attraction|…`" —
// über zweitausend Zeichen, ein einziges Wort ohne Leerzeichen, und die
// Seite lief nach rechts aus dem Fenster.
//
// **Drei Klassen, drei Gründe**: `max-w-prose` hält die Zeile lesbar,
// `break-words` bricht auch ein Wort ohne Leerzeichen um, und
// `line-clamp-3` schneidet nach drei Zeilen ab — wer alles lesen will,
// findet es im Protokoll, und wer am Tisch sitzt, will wissen **dass**
// etwas fehlt.
const befundlisteklassen = 'flex max-w-prose flex-col gap-1 text-sm text-leise'

const befundklassen = 'line-clamp-3 break-words'

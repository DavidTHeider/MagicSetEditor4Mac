// **Die Beispiele der Skriptkonsole** — Werkstück 1 von Phase 3.3
// (#1164).
//
// # Warum zwei Gruppen
//
// Weil die eine Hälfte ohne eingebundenes Stilpaket nicht läuft: `card` und
// `set` legt die Vorlage in die Umgebung. Der Grund steht bei den Listen in
// `beispiele.ts`, samt der Messung — und `internal/service/beispiele_test.go`
// hält beide Zusagen: was hier als „läuft überall" steht, wird dort
// ausgewertet, und was als „braucht eine Vorlage" steht, scheitert dort ohne.
import { Knopf } from '../bausteine/Knopf'
import {
  immerLaufende,
  mitVorlage,
  type Beispiel,
  type Beispielschluessel,
} from './beispiele'
import { useTexte } from '../sprache/kontext'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'

interface SkriptbeispieleEigenschaften {
  // einsetzen schreibt einen Ausdruck ins Feld.
  einsetzen: (ausdruck: string) => void
}

export function Skriptbeispiele({ einsetzen }: SkriptbeispieleEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.konsole

  // **Ohne eigene Überschrift** — sie steht im Kasten der Konsole, und sein
  // Schalter trägt sie (UXR-08, #1916). Die Gruppen sind deshalb die erste
  // Ebene darunter.
  return (
    <div className="flex flex-col gap-2">
      <Gruppe
        titel={eigene.beispieleImmer}
        erklaerung={eigene.beispieleImmerErklaerung}
        eintraege={immerLaufende}
        saetze={eigene.beispielListe}
        einsetzen={einsetzen}
      />

      <Gruppe
        titel={eigene.beispieleVorlage}
        erklaerung={eigene.beispieleVorlageErklaerung}
        eintraege={mitVorlage}
        saetze={eigene.beispielListe}
        einsetzen={einsetzen}
      />
    </div>
  )
}

interface GruppeEigenschaften {
  titel: string
  erklaerung: string
  eintraege: readonly Beispiel[]
  saetze: Record<Beispielschluessel, string>
  einsetzen: (ausdruck: string) => void
}

function Gruppe({
  titel,
  erklaerung,
  eintraege,
  saetze,
  einsetzen,
}: GruppeEigenschaften) {
  return (
    <div className="flex flex-col gap-1">
      <h3 className="text-sm font-medium">{titel}</h3>
      <Erklaerungsabsatz text={erklaerung} klasse="text-sm text-leise" />

      <ul className="flex flex-col gap-1">
        {eintraege.map((beispiel) => (
          <li key={beispiel.ausdruck}>
            {/* **Der Satz steht am Knopf und der Ausdruck darin.** Wer die
                Sprache nicht kennt, liest den Satz; wer sie kennt, sieht den
                Ausdruck — und ein Vorleseprogramm bekommt beides. */}
            <Knopf
              bezeichnung={`${saetze[beispiel.schluessel]} — ${beispiel.ausdruck}`}
              aufKlick={() => {
                einsetzen(beispiel.ausdruck)
              }}
            >
              <span className="font-mono text-xs">{beispiel.ausdruck}</span>
            </Knopf>

            <span className="ml-2 text-sm text-leise">
              {saetze[beispiel.schluessel]}
            </span>
          </li>
        ))}
      </ul>
    </div>
  )
}

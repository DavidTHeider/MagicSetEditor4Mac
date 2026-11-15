// **Die lokale KI-Laufzeit einrichten** — Werkstück A7 von Phase 5.3
// (#1446).
//
// # Zwei Fragen, zwei Bedienelemente
//
// **„Darf gefragt werden" gehört dem Profil, „wo lauscht sie" diesem
// Rechner.** Ein einziger Schalter für beides hieße, dass ein Rechner ohne
// laufendes Ollama auch seine Adresse vergisst.
//
// # Über [Wahl] und nicht selbst gebaut
//
// **Weil dieselbe Zeile sonst zweimal dasteht**: Name, Hinweis, Herkunft, der
// Fokusring und die Entscheidung „ohne Werteliste ein Textfeld" — all das
// steht dort, und ein zweites Bedienelement daneben liefe beim nächsten Umbau
// auseinander. Die Adresse hat **keine** Werte und wird damit von selbst ein
// Textfeld.
//
// # Was hier **nicht** steht
//
// **Kein Knopf, der etwas installiert.** Das Issue verlangt es wörtlich:
// „keine automatische Installation/Downloads/Modellpulls". Wer ein Modell
// braucht, lädt es in Ollama — diese Anwendung liest nur, was da ist.
import { useTexte } from '../sprache/kontext'
import { Wahl } from '../seiten/Wahl'

import type { EinstellungDTO } from '../api'

interface KILaufzeitEigenschaften {
  zeilen: EinstellungDTO[]
  angehalten: (schluessel: string) => boolean
  setze: (schluessel: string, wert: string) => void
}

export function KILaufzeit({
  zeilen,
  angehalten,
  setze,
}: KILaufzeitEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  return (
    <section
      aria-label={eigene.kiLaufzeitTitel}
      className="flex flex-col gap-3"
    >
      <h3 className="text-sm font-medium">{eigene.kiLaufzeitTitel}</h3>

      {/* **Der Hinweis auf Ollama steht hier und nicht bei den fünf
          Ständen** (A8, #1447): die gelten seit dem zweiten Anbieter für
          **jeden**, und „Läuft Ollama?" traf bei Anthropic auf nichts zu.
          Gefunden hat es die Sichtprüfung am laufenden Fenster. */}
      <p className="max-w-xl text-sm text-leise">{eigene.kiLaufzeitOllama}</p>

      <div className="grid max-w-xl gap-5">
        {zeilen.map((eintrag) => (
          <Wahl
            key={eintrag.schluessel}
            eintrag={eintrag}
            angehalten={angehalten(eintrag.schluessel)}
            aendern={(wert) => {
              setze(eintrag.schluessel, wert)
            }}
          />
        ))}
      </div>
    </section>
  )
}

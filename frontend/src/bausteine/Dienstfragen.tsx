// **Wer welche Frage beantwortet** — Werkstück 8 von Phase 4.8
// (#1321),
// entschieden in [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md),
// Entscheidung 1.
//
// **Ein eigener Bereich neben den Schaltern**, weil es zwei verschiedene
// Aussagen sind: „darf gefragt werden" und „beantwortet diese Frage". Wer sie
// in eine Liste legte, hätte zwei Zeilen, die gleich aussehen und
// Verschiedenes tun.
import { Wahl } from '../seiten/Wahl'
import { dienstnamen } from './dienstnamen'
import { useTexte } from '../sprache/kontext'
import type { EinstellungDTO } from '../api'
import { Erklaerungsabsatz } from './Erklaerungsabsatz'

export function Dienstfragen({
  fragen,
  dienste,
  angehalten,
  setze,
}: {
  fragen: EinstellungDTO[]

  // dienste sind die Schalterstände — sie entscheiden, wer hier angeboten wird.
  //
  // **Befund 4 des Zweitblicks über Phase 4.8.**
  // [0063](../../docs/entscheidungen/0063-was-ein-dienst-hier-ist.md) verlangt
  // eine Wahl „aus den Diensten, die eingeschaltet **und** für die Frage
  // geeignet sind". Angeboten wurden alle geeigneten — auch die
  // abgeschalteten, unmittelbar unter den Schaltern, die sie ausschalten.
  dienste: EinstellungDTO[]

  angehalten: (schluessel: string) => boolean
  setze: (schluessel: string, wert: string) => void
}) {
  const texte = useTexte()
  const eigene = texte.seiten.einstellungen

  if (fragen.length === 0) {
    return null
  }

  const namen = eigene as unknown as Record<string, string | undefined>

  const an = new Set(
    dienste
      .filter((schalter) => schalter.wert === 'an')
      .map((schalter) => schalter.schluessel.replace(/^dienst\./, '')),
  )

  // angebot gibt die Dienste, die für diese Frage zur Wahl stehen.
  //
  // **Der eingestellte bleibt darin, auch wenn er aus ist.** Sonst zeigte die
  // Auswahl einen anderen Dienst als den gespeicherten — sie behauptete eine
  // Wahl, die niemand getroffen hat. Dass er nicht antwortet, sagt die Suche
  // mit Namen und Weg.
  const angebot = (eintrag: EinstellungDTO) =>
    eintrag.werte.filter((wert) => an.has(wert) || wert === eintrag.wert)

  return (
    <section aria-label={eigene.fragen} className="mt-8 flex flex-col gap-3">
      <h2 className="text-lg font-semibold">{eigene.fragen}</h2>

      {/* **Über `Satz` und nicht roh** — die Kataloge benutzen `**…**`, und
            roh im JSX stehen die Sternchen wörtlich im Fenster. Dasselbe ist
            schon in #1100
            geschehen; behoben wurde damals die Stelle, nicht die Klasse.
            Jetzt hält `src/auszeichnung.test.tsx` jede Seite dagegen. */}
      <Erklaerungsabsatz
        text={eigene.fragenHinweis}
        klasse="max-w-prose text-sm text-leise"
      />

      <div className="grid max-w-xl gap-5">
        {fragen.map((eintrag) => (
          <Wahl
            key={eintrag.schluessel}
            eintrag={{ ...eintrag, werte: angebot(eintrag) }}
            beschriftung={namen[eintrag.schluessel] ?? eintrag.schluessel}
            // **Die Dienste heißen, wie sie heißen** (0063, Entscheidung 5).
            werteNamen={dienstnamen}
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

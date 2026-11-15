// **Die Kartensuche bei Scryfall** — Werkstück 1 von Phase 3.6
// (#1167).
//
// # Der Befund
//
// Zehn Bindungen sprachen mit Scryfall, und keine davon suchte: man tippte
// einen Kartennamen und bekam **einen** Vorschlag. Wer nicht weiß, wie eine
// Karte genau heißt, kam nicht weiter — und `scryfall.Klient.Suche` gab es
// seit Phase V, mit Tests und ohne Aufrufer.
//
// # Die Frage geht unverändert hinaus
//
// Scryfalls Suchsyntax ist die des Suchfelds seiner Website — `t:creature c:r
// cmc<=3`. Sie hier nachzubauen wäre eine zweite Sprache, die niemand kennt
// und die bei jeder Erweiterung auseinanderliefe. **Wer sie nicht kennt,
// tippt einen Namen**, und das ist eine gültige Frage; die Seite sagt es.
//
// # Gesucht wird auf Anweisung
//
// **Nicht beim Tippen.** Scryfall nennt eine Grenze, und §16 verlangt, sie
// **unter** der genannten Zahl einzuhalten; eine Suche je Tastendruck ginge
// dagegen. Der Knopf ist die Anweisung.
import { useState } from 'react'

import { api } from '../fassade'
import { Knopf } from '../bausteine/Knopf'
import { useVorgang } from '../haken/useVorgang'
import { einsetzen } from '../sprache/einsetzen'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { Erklaerungsabsatz } from '../bausteine/Erklaerungsabsatz'
import type { SuchtrefferDTO } from '../apiscryfall'

interface ScryfallsucheEigenschaften {
  // waehlen bekommt den gewählten Treffer — die Seite darüber entscheidet,
  // was damit geschieht.
  //
  // **`undefined` nimmt die Wahl zurück** — Befund 1 des Nachlaufs über
  // Phase 3.6. Eine gescheiterte Suche räumte bis dahin nur die **eigene**
  // Trefferliste; der Treffer, den sie vorher nach oben gegeben hatte, blieb
  // dort stehen und wurde weiter angeboten.
  waehlen: (treffer: SuchtrefferDTO | undefined) => void
}

export function Scryfallsuche({ waehlen }: ScryfallsucheEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.scryfallsuche
  const vorgang = useVorgang()

  const [frage, setzeFrage] = useState('')
  const [treffer, setzeTreffer] = useState<SuchtrefferDTO[]>()
  const [weitere, setzeWeitere] = useState(false)

  // **Alter und Befunde der Antwort** — Befund 4 des Zweitblicks über
  // Phase 3.6. Sie reisen durch `service` und `app` bis hierher und wurden
  // hier weggeworfen: ein Stand aus dem Zwischenspeicher sah aus wie eine
  // frische Auskunft. §16, Regel 5: der letzte gute Stand bleibt, **und er
  // sagt, wie alt er ist**.
  const [alterStunden, setzeAlter] = useState(0)
  const [befunde, setzeBefunde] = useState<readonly string[]>([])

  const [fehler, setzeFehler] = useState<string>()

  const suchen = () =>
    vorgang.starte('scryfallsuche', async () => {
      try {
        const antwort = await api().ScryfallSuche(frage)
        const gefunden = auspacken(antwort, () => antwort)

        setzeTreffer(gefunden.treffer)
        setzeWeitere(gefunden.weitere)
        setzeAlter(gefunden.alterStunden)
        setzeBefunde(gefunden.befunde)
        setzeFehler(undefined)
      } catch (ursache: unknown) {
        // **Und die alten Treffer gehen weg** — Befund 5 des Zweitblicks über
        // Phase 3.6. Sie stehen zu lassen hieße, eine gescheiterte Suche mit
        // dem Ergebnis der vorigen zu beantworten; wer den Fehler übersieht,
        // klickt einen Treffer an, den er gar nicht gesucht hat.
        setzeTreffer(undefined)
        setzeWeitere(false)
        setzeAlter(0)
        setzeBefunde([])

        // **Und die gehobene Wahl geht mit.** Sie stammt aus einer
        // Trefferliste, die es nicht mehr gibt.
        waehlen(undefined)

        setzeFehler(meldungAus(ursache, texte))
      }
    })

  return (
    <section className="flex flex-col gap-2">
      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.frage}</span>

        <input
          type="text"
          value={frage}
          placeholder={eigene.platzhalter}
          onChange={(ereignis) => {
            setzeFrage(ereignis.target.value)
          }}
          onKeyDown={(ereignis) => {
            // **Die Eingabetaste sucht.** Wer tippt, will nicht zur Maus
            // greifen — und §6 verlangt, dass jede Handlung mit der Tastatur
            // erreichbar ist.
            if (ereignis.key === 'Enter') {
              ereignis.preventDefault()
              void suchen()
            }
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 placeholder:text-leise focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring dark:placeholder:text-leise"
        />
      </label>

      <Erklaerungsabsatz text={eigene.erklaerung} klasse="text-sm text-leise" />

      <div>
        <Knopf
          bezeichnung={eigene.suchen}
          aufKlick={() => {
            void suchen()
          }}
        >
          {eigene.suchen}
        </Knopf>
      </div>

      {fehler !== undefined && (
        <p role="alert" className="text-sm text-fehler">
          {fehler}
        </p>
      )}

      {/* **Der Stand sagt, wie alt er ist** (§16, Regel 5). Eine Antwort aus
          dem Zwischenspeicher sah bis hierher aus wie eine frische — Befund 4
          des Zweitblicks über Phase 3.6. */}
      {treffer !== undefined && (
        <p className="text-xs text-leise">
          {alterStunden === 0
            ? eigene.frisch
            : einsetzen(eigene.alter, {
                stunden: String(Math.round(alterStunden)),
              })}
        </p>
      )}

      {/* **Und die Befunde reisen bis hierher.** Ein unbekannter
          Aufzählungswert soll den sehen, der wählt — nicht nur den, der die
          Antwort gelesen hat (§8). */}
      {befunde.map((einer) => (
        <p key={einer} className="text-xs text-hinweis">
          {einsetzen(eigene.befund, { text: einer })}
        </p>
      ))}

      {/* **Keine Treffer ist eine Antwort und kein leerer Bereich** (§6). */}
      {treffer?.length === 0 && (
        <p className="text-sm text-leise">{eigene.nichtsGefunden}</p>
      )}

      {treffer !== undefined && treffer.length > 0 && (
        <>
          <ul className="flex flex-col gap-1">
            {treffer.map((einer) => (
              <li key={einer.kennung}>
                {/* **Die ganze Zeile wählt** — und das Bild steht dabei, denn
                    eine Karte erkennt man am Bild, nicht am Setcode. */}
                <button
                  type="button"
                  onClick={() => {
                    waehlen(einer)
                  }}
                  className="flex w-full items-center gap-3 rounded-md border border-kante px-2 py-1 text-left hover:bg-betont focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                >
                  <span className="flex h-14 w-10 shrink-0 items-center justify-center overflow-hidden rounded-sm bg-betont">
                    {einer.bild !== '' && (
                      <img
                        src={einer.bild}
                        alt=""
                        className="max-h-full max-w-full object-contain"
                      />
                    )}
                  </span>

                  <span className="min-w-0">
                    <span className="block truncate font-medium">
                      {einer.name}
                    </span>

                    <span className="block truncate text-sm text-leise">
                      {`${einer.setname} · ${einer.setcode.toUpperCase()} ${einer.nummer} · ${einer.seltenheit}`}
                    </span>
                  </span>
                </button>
              </li>
            ))}
          </ul>

          {/* **»Es gibt mehr« steht dabei.** Eine Liste, die schweigend endet,
              sieht aus wie eine vollständige — und nachgeladen wird nicht. */}
          {weitere && <p className="text-sm text-leise">{eigene.mehr}</p>}
        </>
      )}
    </section>
  )
}

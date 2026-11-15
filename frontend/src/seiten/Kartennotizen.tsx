// **Die Notizen einer Karte** — Werkstück 3 von Phase 3.5
// (#1166).
//
// # ported, und die Fundstelle ist genau
//
//	notes = new TextCtrl(nodes_panel, ID_NOTES, true);
//	…
//	sc->Add(new wxStaticText(nodes_panel, wxID_ANY, _LABEL_("card notes")), …);
//	sn->Add(notes, 1, wxEXPAND | wxTOP, 2);
//	splitter->SplitHorizontally(card_list, nodes_panel, -40);
//
// `src/gui/set/cards_panel.cpp`, Zeilen 39–55, Fassung `a7336d6` vom
// 2020-06-14. Drei Dinge stehen darin, und alle drei gelten hier:
//
//   - **Ein eigener Kasten unter der Kartenliste**, nicht ein Feld zwischen
//     den Feldwerten. `notes` ist keines: `String notes;` steht in `Card`
//     neben `data` (`src/data/card.hpp`, Zeile 39).
//   - **Mehrzeilig** — das `true` ist `multi_line`
//     (`TextCtrl::TextCtrl(Window*, int, bool multi_line, long)`).
//   - **An jeder Karte**, auch an einer ohne Notizen. Der Kasten hängt am
//     Panel und nicht am Wert.
//
// **Nicht portiert ist der Zuklappknopf** (`collapse_notes`, Zeile 40). Er
// löst ein Platzproblem eines festen Fensterteilers; hier rollt die Spalte,
// und ein Knopf, der etwas verbirgt, was ohnehin unten steht, verstellte den
// Weg mehr, als er spart.
//
// # Geschrieben wird beim Verlassen
//
// **Und nicht bei jedem Tastendruck.** Eine Notiz ist Fließtext; bei drei
// Zeilen wären das hundert Aufrufe, und jeder ginge gegen einen Stand, den der
// Bestand gerade erst bekommen hat (§7, Regel 4). Der Karteneditor daneben
// entscheidet dieselbe Frage anders und aus gutem Grund: dort ist die
// **Vorschau** die Antwort, und §6 verlangt sie ohne Bestätigung. Eine Notiz
// steht auf keiner Karte.
import { useRef } from 'react'

import { api } from '../fassade'
import { Entwurfskonflikt } from '../bausteine/Entwurfskonflikt'
import { useEntwurf } from '../haken/useEntwurf'
import { useVorgang } from '../haken/useVorgang'
import { meldungAus } from '../meldung'
import { useTexte } from '../sprache/kontext'
import { auspacken } from '../umschlag'
import { notizenfeld } from './angebotenefelder'
import { schreibenBeginnt, schreibenEndet } from '../speicherstand'
import { useLesegrund } from '../haken/schreibrecht'

interface KartennotizenEigenschaften {
  setID: string
  karteID: string

  // werte sind die Feldwerte der Karte — `notes` ist einer davon.
  werte: readonly { feld: string; wert: string }[]

  // geaendert meldet, dass geschrieben wurde.
  geaendert: () => void

  // aufFehler meldet einen Schreibfehler **nach oben**.
  //
  // # Warum nicht hier
  //
  // Der Kasten wird beim Kartenwechsel abgebaut (der `key` in
  // `Kartenspalte`). Ein Schreibvorgang, den das Verlassen ausgelöst hat,
  // lebt in seiner Closure weiter — und meldete seinen Fehler an eine
  // Komponente, die es nicht mehr gibt. Der gewöhnliche Weg »Notiz tippen und
  // gleich die nächste Karte anklicken« verlor damit bei einem Schreibfehler
  // die Änderung **ohne sichtbare Auskunft**.
  //
  // Angestrichen vom Nachlauf über Phase 3.5 als Befund 3. Der Aufrufer
  // überlebt den Wechsel; deshalb hält er die Meldung.
  aufFehler: (meldung: string | undefined) => void
}

export function Kartennotizen({
  setID,
  karteID,
  werte,
  geaendert,
  aufFehler,
}: KartennotizenEigenschaften) {
  const texte = useTexte()
  const eigene = texte.seiten.karten.notizen
  const vorgang = useVorgang()
  // **Ein Leser schreibt nicht** (#2290) — die Notiz bleibt lesbar.
  // **Mit Grund** (H01, #2255): der Kasten sagt einem Leser, warum er nur
  // liest — wie ein gesperrter Knopf.
  const lesegrund = useLesegrund()
  const lesend = lesegrund.gesperrt

  const gespeichert =
    werte.find((wert) => wert.feld === notizenfeld)?.wert ?? ''

  // ausstehend hält den Text, der noch geschrieben werden muss.
  //
  // # Warum es ihn braucht
  //
  // `useVorgang.starte` gibt für einen laufenden Schlüssel still `undefined`
  // zurück (§7, Regel 4) — richtig für einen Knopf, falsch für einen Kasten,
  // in dem jemand weitertippt. Wer den Kasten ein zweites Mal verlässt,
  // während der erste Ruf noch unterwegs ist, verlor seine Änderung, und
  // niemand sagte es. Angestrichen vom Zweitblick über Phase 3.5 als Befund 9.
  //
  // **Ein nachlaufender Schreibvorgang und keine Warteschlange**: was zählt,
  // ist der **letzte** Text, nicht jeder Zwischenstand.
  const ausstehend = useRef<
    { text: string; bekannt: readonly string[] } | undefined
  >(undefined)

  // **Der Kasten gehört zur Karte.** Ein Wechsel räumt ihn (§7, Regel 1) —
  // das tut der `key` am Aufrufer, und deshalb steht hier keine Umschaltung.

  // **Ein unveränderter Kasten schreibt gar nicht** — das prüft
  // `useEntwurf`, gegen den gespeicherten Wert und gegen das, was dieser
  // Entwurf schon geschickt hat. Bis U04 (#2121) hielt ein eigener Merker
  // hier den zuletzt geschriebenen Text; ohne Neuaufbau bei jedem
  // Verlaufsstand wäre er nach ⌘Z veraltet gewesen.
  // **Die bekannten Werte reisen mit dem Text** (U02b, #2179): wer nachholt,
  // schreibt mit der Bedingung des Entwurfs, der ihn schickte.
  const schreiben = (neuerText: string, bekannt: readonly string[]) => {
    // **Läuft schon einer, wird nur gemerkt.** Der laufende Ruf holt ihn am
    // Ende nach; ohne das fiele die zweite Änderung still weg (Befund 9).
    ausstehend.current = { text: neuerText, bekannt }

    void vorgang.starte(karteID, async () => {
      // **Solange etwas aussteht, wird geschrieben** — höchstens ein Ruf zur
      // Zeit, aber keiner geht verloren.
      while (ausstehend.current !== undefined) {
        const dieser = ausstehend.current
        ausstehend.current = undefined

        // **Der Speicherstand hört mit** (UX-05.AK-03) — die Notizen
        // schrieben bis zum Zweitblick über Phase 6.3 an ihm vorbei (M8).
        schreibenBeginnt(setID)

        try {
          const antwort = await api().FeldSetzen(
            setID,
            karteID,
            notizenfeld,
            dieser.text,
            dieser.bekannt,
          )

          // **`FeldSetzen` gibt den Umschlag selbst zurück**, nicht einen Wert
          // darin — ihn noch einmal einzupacken ergäbe einen Fehler, den
          // `auspacken` nicht deuten kann, und die Meldung käme leer an.
          auspacken(antwort, () => null)

          schreibenEndet(setID, karteID, notizenfeld, true)
          aufFehler(undefined)
          geaendert()
        } catch (ursache: unknown) {
          // **Ein gescheiterter Schreibvorgang beendet die Schleife.** Sonst
          // liefe sie gegen dieselbe Ursache weiter, und die Meldung
          // flackerte.
          ausstehend.current = undefined
          schreibenEndet(setID, karteID, notizenfeld, false)
          aufFehler(meldungAus(ursache, texte))
        }
      }
    })
  }

  const entwurf = useEntwurf(gespeichert, schreiben)

  return (
    <div className="mt-2 flex flex-col gap-1">
      {lesegrund.satz}
      <label className="flex flex-col gap-1">
        <span className="text-sm font-medium">{eigene.ueberschrift}</span>

        {/* **Der Entwurf übersteht fremde Änderungen** (U04, #2121): solange
            jemand tippt, bleibt der Schlüssel; sonst trägt er den
            gespeicherten Text, und eine Rücknahme erreicht den Kasten. */}
        <textarea
          key={entwurf.schluessel}
          defaultValue={gespeichert}
          rows={3}
          placeholder={eigene.platzhalter}
          readOnly={lesend}
          aria-describedby={lesegrund.beschreibe()}
          onChange={(ereignis) => {
            entwurf.aendern(ereignis.target.value)
          }}
          onBlur={(ereignis) => {
            entwurf.verlassen(ereignis.target.value)
          }}
          className="rounded-sm border border-kante bg-transparent px-2 py-1 font-mono text-sm placeholder:text-leise focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring dark:placeholder:text-leise"
        />
      </label>

      <Entwurfskonflikt
        griff={entwurf}
        fremd={gespeichert}
        angehalten={vorgang.laeuft(karteID)}
      />
    </div>
  )
}

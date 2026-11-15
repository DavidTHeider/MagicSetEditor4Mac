// Ein Set in der Liste: sein Name, seine Kennzahlen, seine Handlungen.
//
// # Warum keine Tabellenzeile mehr
//
// Bis #288
// stand die Setliste als Tabelle mit sieben Spalten, **vier davon Knöpfe**. In
// einem Fenster von 1.280 Punkten brachen sie über fünf Zeilen um: „Als / MTG /
// JSON / sichern / …". Eine Tabelle ist zum Vergleichen da; hier gibt es
// nichts zu tun als das eine: das Set aufschlagen.
//
// # Der Name führt, das Seltene steht im Menü
//
// **UXR-04** (#1912).
// Bis dahin trug jede Karte fünf gleichrangige Knöpfe — Sichern, Sichern
// unter …, Als Umschlag ablegen, Über eine Exportvorlage …, Löschen … —, mit
// *+ Marke* sieben Bedienelemente je Set: sechzig Sets auf einer Seite der
// Liste waren 420 (R6 aus `docs/UIREVIEW.md`). Das Artboard *Sets* zeichnet je
// Set eine Zeile, die öffnet. **Jetzt öffnet der Name, und die fünf stehen in
// einem Menü** — das Löschen zuletzt, wie bei den Paketen (UXR-17).
//
// **Was ein Eintrag aufschlägt, steht unter der Zeile** und nicht im Menü:
// die Rückfrage zum Löschen und die Exportvorlage. Beide sind unabhängig
// voneinander offen; zuklappen räumt ihren Zustand mit (§7, Regel 1).
import { useState } from 'react'

import { type SetDTO } from '../api'
import { type Gestalt } from '../bausteine/Listenansicht'
import { Kontextmenue } from '../bausteine/Kontextmenue'
import { Menue, type Menueeintrag } from '../bausteine/Menue'
import { type Speicherlauf } from '../haken/useSetSpeichern'
import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { zaehlwort } from '../sprache/zaehlwort'
import { SetVorlagenausfuhr } from './SetVorlagenausfuhr'
import { Setloeschen } from './Setloeschen'
import { type Setloeschung } from '../haken/useSetloeschung'
import { type Umschlaglauf } from '../haken/useUmschlag'

interface SetkarteEigenschaften {
  // gestalt ist die Anordnung — Kachel oder Zeile (UX-11, #1683). **Nur die
  // Anordnung**: Handlungen und Rechte sind in beiden dieselben.
  gestalt?: Gestalt

  satz: SetDTO
  waehlen: (setID: string) => void

  // speichern, loeschung und umschlag sind die Haken der **Seite** — einer
  // je Zeile hieße §7 Regel 4 je Zeile statt je Set (#162, #1218, #1457).
  speichern: Speicherlauf
  loeschung: Setloeschung
  umschlag: Umschlaglauf
}

export function Setkarte({
  gestalt = 'kachel',
  satz,
  waehlen,
  speichern,
  loeschung,
  umschlag,
}: SetkarteEigenschaften) {
  const texte = useTexte()
  const worte = texte.seiten.sets

  const [fragt, setzeFragt] = useState(false)
  const [vorlageOffen, setzeVorlageOffen] = useState(false)

  const befunde = loeschung.befunde(satz.id)

  // **Die Meldung des Umschlags steht an der Karte, aus der er kam** (§6) —
  // `fuer` sagt, welche. Bis UXR-04 stand sie nur in der Annahme unter „Ein
  // Set von woanders einlesen", und die war zugeklappt, also gar nicht
  // gezeichnet: wer ablegte, sah weder Erfolg noch Fehler.
  const umschlagHier = umschlag.fuer === satz.id

  const eintraege: Menueeintrag[] = [
    {
      schluessel: 'sichern',
      name: worte.speichern,
      aktion: 'sichern',
      angehalten: speichern.laeuft(satz.id),
      aufWahl: () => {
        speichern.speichern(satz.id, false)
      },
    },
    {
      schluessel: 'sichernUnter',
      name: worte.speichernUnter,
      aktion: 'sichern',
      angehalten: speichern.laeuft(satz.id),
      aufWahl: () => {
        speichern.speichern(satz.id, true)
      },
    },
    {
      schluessel: 'umschlag',
      name: worte.umschlagAblegen,
      angehalten: umschlag.laeuft(`schnueren-${satz.id}`),
      aufWahl: () => {
        umschlag.schnueren(satz.id)
      },
    },
    {
      schluessel: 'vorlage',
      name: worte.vorlageAusfuhr,
      aufWahl: () => {
        setzeVorlageOffen(true)
      },
    },
    {
      schluessel: 'loeschen',
      schreibt: true,
      name: worte.loeschen,
      aktion: 'loeschen',
      angehalten: loeschung.laeuft(satz.id),
      aufWahl: () => {
        setzeFragt(true)
      },
    },
  ]

  const bezeichnung = einsetzen(worte.handlungenAn, { name: satz.name })

  return (
    // **Ein Abstand nach unten**: darunter stehen die Marken des Eintrags
    // (`Listenansicht`), und ohne ihn klebte *+ Marke* an der Kante —
    // angestrichen in der Hüllenprobe `fmt03a`.
    <article
      className={
        gestalt === 'kachel'
          ? 'mb-2 h-full rounded-lg border border-kante bg-flaeche p-4'
          : 'mb-1 rounded-md border border-kante bg-flaeche px-3 py-2'
      }
    >
      {/* **Rechtsklick und ⇧F10 an der ganzen Karte** — AP-5 (#2362),
          Befund B-01: dieselben Einträge wie im Menü daneben. */}
      <Kontextmenue
        bezeichnung={bezeichnung}
        eintraege={eintraege}
        objekt={`set/${satz.id}`}
        klasse="h-full"
      >
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            {/* **h2 und nicht h3** (#2249): die Liste hat keine eigene
              Überschrift, und neben ihr steht „Ein neues Set" als h2 — eine
              h3 darunter sprang von der h1 der Seite auf die dritte Ebene. */}
            <h2 className="text-lg font-semibold">
              <button
                type="button"
                aria-label={`${satz.name} — ${worte.oeffnen}`}
                onClick={() => {
                  waehlen(satz.id)
                }}
                className="rounded-sm text-left text-marke underline underline-offset-2 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
              >
                {satz.name}
              </button>
            </h2>

            <p className="mt-1 text-leise">
              {`${satz.spiel} · ${zaehlwort(satz.karten, texte.allgemein.zaehlwoerter.karten)} · ${satz.stil}`}
            </p>

            {/* **Der Dateiname, der Pfad im Titel** — AP-11 (#2368), B-27: ein
              Pfad aus `/var/folders/…` an der Karte ist Technik, gebraucht
              wird sein Ende. Vorgelesen wird der ganze. */}
            <p
              title={satz.herkunft}
              className="mt-1 truncate text-sm text-leise"
            >
              <span aria-hidden="true">{satz.herkunft.split('/').pop()}</span>
              <span className="sr-only">{satz.herkunft}</span>
            </p>
          </div>

          <Menue
            bezeichnung={bezeichnung}
            gewicht="leise"
            ausrichtung="rechts"
            eintraege={eintraege}
          >
            {worte.handlungenKurz}
          </Menue>
        </div>

        {umschlagHier && umschlag.geschnuert !== undefined && (
          <p role="status" className="mt-2 max-w-prose text-sm text-leise">
            {worte.umschlagGeschnuert}
          </p>
        )}

        {umschlagHier && umschlag.fehler !== undefined && (
          <p role="alert" className="mt-2 text-sm text-fehler">
            {texte.bausteine.zustaende.fehler} {umschlag.fehler}
          </p>
        )}

        {vorlageOffen && (
          <SetVorlagenausfuhr
            satz={satz}
            schliessen={() => {
              setzeVorlageOffen(false)
            }}
          />
        )}

        {/* **Die Rückfrage steht unter der Zeile** — das einzige Unumkehrbare
          bekommt Platz und den Fokus, nicht eine Stelle in einer Reihe. Die
          Reste eines geglückten Löschens stehen auch ohne offene Frage. */}
        {(fragt || befunde.length > 0) && (
          <div className="mt-3 border-t border-kante pt-3">
            <Setloeschen
              satz={satz}
              laeuft={loeschung.laeuft(satz.id)}
              loeschen={(setID, name) => {
                void loeschung.loeschen(setID, name)
              }}
              abbrechen={() => {
                setzeFragt(false)
              }}
              befunde={befunde}
              fehler={loeschung.fehler(satz.id)}
            />
          </div>
        )}
      </Kontextmenue>
    </article>
  )
}

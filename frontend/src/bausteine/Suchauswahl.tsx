// **Eine Auswahl mit Suche** — UXR-F1
// (#2007),
// R7 aus `docs/UIREVIEW.md`: *„Auswahl mit Suche ab 15 Werten"*; Schritt 2
// aus [0125](../../../docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md).
// Wer nur **entscheiden** will, ob eine Suche dazugehört, nimmt das
// [`Auswahlfeld`](./Auswahlfeld.tsx).
//
// # Das Muster
//
// Das **Kombinationsfeld** der Praktiken zur Barrierefreiheit, in der Form
// „bearbeitbar, mit Liste": ein Eingabefeld mit `role="combobox"`, das eine
// Liste (`role="listbox"`, Zeilen mit `role="option"`) öffnet. **Der Fokus
// bleibt im Feld**; die hervorgehobene Zeile steht in `aria-activedescendant`.
//
// # Was es zusagt
//
//   - **Tippen filtert**, ohne Rücksicht auf Groß und klein, über Teilwörter,
//     über Name **und** Zusatz — und `_`, `-` und `/` zählen wie ein
//     Leerzeichen, damit „new york" `America/New_York` findet.
//   - **Pfeil hoch und runter** wandern, rundherum; **Enter** wählt;
//     **Escape** schließt und stellt den vorigen Wert her; **Klick** wählt.
//   - **Escape wird nur verbraucht, solange die Liste offen ist.** Ist sie
//     zu, gehört die Taste den anderen — sonst ließe sich eine Schublade, in
//     der das Feld steht, mit ihm nie schließen (`Schublade.tsx` fragt nach
//     `defaultPrevented`).
//   - **Ein leeres Ergebnis sagt es**: ein Satz aus dem Katalog in der Liste,
//     und derselbe in einer Live-Region, die Vorleseprogramme ansagen. Dort
//     steht auch die Zahl der Treffer. **Gezeichnet werden höchstens
//     `zeilengrenze` Zeilen**; gesucht wird in allen.
//   - **Was die Liste nicht kennt, bleibt sichtbar gewählt.** Der Wert, der
//     gilt, steht im Feld, wie er ist — er verschwindet nicht und fällt nicht
//     auf die erste Zeile.
//   - **Eine Wahl, die schon gilt, ist keine Änderung**: `aendern` läuft nicht.
//
// # Wer was liefert — **Mantine** und **eigen**
//
// Mantine liefert **die Fläche**: das Portal, die Lage am Feld (`flip`,
// `shift`, die Breite des Feldes), das Schließen auf einen Klick daneben, die
// Zeilen und den Satz für „nichts gefunden". **Die Tastatur und die Rollen
// sind eigen**, und der Grund ist nachgelesen: Mantines Pfeiltasten merken die
// Kennung der Zeile in einem Zustand des Ziels, den nur ein Pfeil
// fortschreibt (`use-combobox-target-props.mjs`, Mantine 9.6.1). Filtert die
// Suche die Liste, zeigt `aria-activedescendant` auf eine Zeile, die es nicht
// mehr gibt. Hier steht die Stelle als Zahl im Baustein; Attribut und Rolle
// folgen ihr. **Die Liste ist nur da, solange sie offen ist**
// (`keepMounted={false}`): Mantine hielte sie sonst verborgen im Baum.
//
// # Was der Aufrufer liefert
//
//   - **Die Beschriftung**: ein `<label htmlFor={id}>` oder die Beschriftung
//     eines `Feld`. Der Baustein malt sie nicht selbst, denn zwei Aufrufer
//     (`Feld`, die Zeile einer Einstellung) haben sie schon — eine zweite
//     stünde doppelt da. Dazu `bezeichnung` als Name der Liste.
//   - **Eine Zeile mit leerem Wert**, wenn „nichts" eine Wahl ist („wie das
//     System", „kein Spiel"). Ohne sie zeigt das Feld seinen Platzhalter.
//
// **Farben aus den gemessenen Merkmalen** (`docs/KONTRASTE.md`), nicht aus
// Mantines Palette, und deshalb `unstyled` — wie beim `Popover`.
import { Combobox, Portal, useCombobox } from '@mantine/core'
import { useId, useState, type KeyboardEvent } from 'react'

import { einsetzen } from '../sprache/einsetzen'
import { useTexte } from '../sprache/kontext'
import { suchtreffer, useSuchfeld, zeigen, type Suchwahl } from './suchwahl'

interface SuchauswahlEigenschaften {
  // id ist die Kennung des Feldes — die der Aufrufer mit seiner Beschriftung
  // verbindet.
  id: string

  // bezeichnung benennt die Liste. Meist der Text der Beschriftung.
  bezeichnung: string

  eintraege: readonly Suchwahl[]

  // wert ist die gewählte Zeile — kontrolliert, wie bei jedem Feld hier.
  wert: string
  aendern: (wert: string) => void

  // platzhalter steht im Feld, solange nichts gewählt ist und keine Zeile
  // dazu gehört.
  platzhalter?: string

  // angehalten sperrt das Feld, solange ein Vorgang läuft (§7, Regel 4).
  angehalten?: boolean

  // beschriebenVon und ungueltig reichen die Verbindung eines `Feld` an das
  // Element weiter.
  beschriebenVon?: string | undefined
  ungueltig?: boolean

  // einstellung kennzeichnet das Feld als das einer Einstellung; gezählt wird
  // darüber (`einstellungsdoppelung.test.tsx`).
  einstellung?: string

  // benennen setzt `bezeichnung` auch als `aria-label` an das Feld — für die
  // Aufrufer ohne `<label>`, wie die Kartenfelder.
  benennen?: boolean
}

// **Der Pfeil ist der des nativen Felds** (`auswahlpfeil`, AP-3, #2360):
// bis dahin stand ein eigenes Zeichen daneben, und zwei Auswahlen sahen
// verschieden aus. `pr-7` hält ihm den Platz frei, wie das native Feld.
const feldklassen =
  'auswahlpfeil w-full rounded-sm border border-kante bg-flaeche py-1 pl-2 pr-7 text-text ' +
  'placeholder:text-leise disabled:opacity-60 ' +
  'focus-visible:outline focus-visible:outline-2 ' +
  'focus-visible:outline-offset-2 focus-visible:outline-ring'

const zeilenklassen =
  'cursor-pointer px-3 py-1.5 text-text ' +
  'data-combobox-active:font-semibold ' +
  'data-combobox-selected:bg-betont ' +
  'data-combobox-selected:outline data-combobox-selected:outline-2 ' +
  'data-combobox-selected:-outline-offset-2 ' +
  'data-combobox-selected:outline-ring'

export function Suchauswahl({
  id,
  bezeichnung,
  eintraege,
  wert,
  aendern,
  platzhalter,
  angehalten = false,
  beschriebenVon,
  ungueltig = false,
  einstellung,
  benennen = false,
}: SuchauswahlEigenschaften) {
  const worte = useTexte().bausteine.suchauswahl
  const listenID = useId()
  // Fokus über das Schreiben und ruhender Zeiger — am Fenster gemessen.
  const { feld: feldRef, merkeFokus, zeiger } = useSuchfeld(angehalten)

  // entwurf ist, was jemand getippt hat — `undefined`, solange er nicht
  // tippt. Er filtert, und er ist nie die Wahl: erst eine Zeile ist es.
  const [entwurf, setzeEntwurf] = useState<string>()

  // aktiv ist die Stelle der hervorgehobenen Zeile in den Treffern, −1 ohne.
  const [aktiv, setzeAktiv] = useState(-1)

  // **Jedes Schließen verwirft den Entwurf**: Escape, ein Klick daneben, das
  // Verlassen des Feldes, eine Wahl. Eine Stelle, damit keiner vergessen wird.
  const combobox = useCombobox({
    onDropdownClose: () => {
      setzeEntwurf(undefined)
    },
  })
  const offen = combobox.dropdownOpened

  const gewaehlt = eintraege.find((eintrag) => eintrag.wert === wert)
  const suche = entwurf ?? ''
  const { gefunden, gezeigt: treffer } = suchtreffer(eintraege, suche)
  const hervorgehoben = treffer[aktiv]

  // **Aufgeklappt heißt: die Liste steht da.** Passt nichts, gibt es keine
  // Liste, nur den Satz — und ein Feld, das `aria-expanded` und
  // `aria-controls` auf nichts richtete, bräche das Muster (axe:
  // `aria-required-attr`).
  const listeDa = offen && treffer.length > 0

  const zeilenID = (stelle: number) => `${listenID}-${String(stelle)}`

  const oeffnen = (mitTaste: boolean) => {
    const stelle = treffer.findIndex((eintrag) => eintrag.wert === wert)

    combobox.openDropdown()
    // **Die Wahl, die gilt, ist die erste hervorgehobene.** Ohne sie beginnt
    // ein Pfeil oben; ein Klick hebt nichts hervor.
    setzeAktiv(stelle >= 0 ? stelle : mitTaste ? 0 : -1)
  }

  const waehlen = (neu: string) => {
    combobox.closeDropdown()

    if (neu !== wert) {
      merkeFokus()
      aendern(neu)
    }
  }

  const aufTaste = (ereignis: KeyboardEvent<HTMLInputElement>) => {
    // **Während einer Umwandlung (Japanisch, Chinesisch) gehört die Taste der
    // Eingabemethode**: ihr Enter bestätigt den Text und wählt keine Zeile.
    if (ereignis.nativeEvent.isComposing) {
      return
    }

    if (ereignis.key === 'ArrowDown' || ereignis.key === 'ArrowUp') {
      ereignis.preventDefault()

      if (!offen) {
        oeffnen(true)

        return
      }

      const schritt = ereignis.key === 'ArrowDown' ? 1 : -1
      const anzahl = treffer.length

      // **Wo nichts passt, gibt es nichts anzusteuern** — und `% 0` wäre
      // keine Stelle.
      if (anzahl === 0) {
        return
      }

      // **Rundherum**; ohne Hervorhebung führt ein Pfeil nach oben ans Ende.
      setzeAktiv(
        aktiv === -1 && schritt === -1
          ? anzahl - 1
          : (aktiv + schritt + anzahl) % anzahl,
      )
    }

    if (ereignis.key === 'Enter' && offen && hervorgehoben !== undefined) {
      ereignis.preventDefault()
      waehlen(hervorgehoben.wert)
    }

    // **Verbraucht wird Escape nur bei offener Liste** — siehe oben.
    if (ereignis.key === 'Escape' && offen) {
      ereignis.preventDefault()
      ereignis.stopPropagation()
      combobox.closeDropdown()
    }
  }

  // meldung ist, was die Live-Region ansagt — nichts, solange die Liste zu
  // ist, sonst die Zahl der Treffer oder dass keiner passt.
  // **Gibt es mehr Treffer, als gezeichnet sind**, sagt der Satz beides: wie
  // viele es sind und wie viele da stehen.
  const begrenzt = einsetzen(worte.begrenzt, {
    gezeigt: String(treffer.length),
    anzahl: String(gefunden),
  })
  const meldung = !offen
    ? ''
    : gefunden === 0
      ? einsetzen(worte.keiner, { suche: suche.trim() })
      : gefunden > treffer.length
        ? begrenzt
        : einsetzen(worte.treffer, { anzahl: String(gefunden) })

  return (
    <div className="relative">
      <Combobox
        store={combobox}
        onOptionSubmit={waehlen}
        keepMounted={false}
        unstyled
        position="bottom-start"
        offset={2}
      >
        <Combobox.Target
          withKeyboardNavigation={false}
          withAriaAttributes={false}
        >
          <input
            id={id}
            type="text"
            role="combobox"
            autoComplete="off"
            spellCheck={false}
            aria-expanded={listeDa}
            aria-haspopup="listbox"
            aria-autocomplete="list"
            aria-controls={listeDa ? listenID : undefined}
            aria-activedescendant={
              listeDa && hervorgehoben !== undefined
                ? zeilenID(aktiv)
                : undefined
            }
            aria-label={benennen ? bezeichnung : undefined}
            aria-describedby={beschriebenVon}
            aria-invalid={ungueltig ? true : undefined}
            data-einstellung={einstellung}
            // **Offen und ungetippt ist das Feld leer**, und die Wahl steht
            // als Platzhalter da: Tippen beginnt neu. Eine Markierung per
            // `select()` hielt WKWebView nach dem Klick nicht (`fmt02d`).
            placeholder={
              offen && entwurf === undefined
                ? (gewaehlt?.name ?? platzhalter)
                : platzhalter
            }
            disabled={angehalten}
            value={entwurf ?? (offen ? '' : (gewaehlt?.name ?? wert))}
            onChange={(ereignis) => {
              const text = ereignis.target.value

              setzeEntwurf(text)
              combobox.openDropdown()
              // **Der erste Treffer ist hervorgehoben, damit Enter ihn
              // nimmt** — bei leerer Suche keiner, sonst bekäme jeder, der
              // den Text löscht und Enter drückt, die erste Zeile.
              setzeAktiv(text.trim() === '' ? -1 : 0)
            }}
            onKeyDown={aufTaste}
            ref={feldRef}
            onClick={(ereignis) => {
              zeiger.merke(ereignis)
              if (!offen) {
                oeffnen(false)
              }
            }}
            onBlur={() => {
              combobox.closeDropdown()
            }}
            className={feldklassen}
          />
        </Combobox.Target>

        <Combobox.Dropdown className="absolute rounded-md border-2 border-kante bg-flaeche text-text">
          {treffer.length > 0 ? (
            <Combobox.Options
              id={listenID}
              aria-label={bezeichnung}
              className="max-h-64 overflow-y-auto py-1"
            >
              {treffer.map((eintrag, stelle) => (
                <Combobox.Option
                  key={eintrag.wert}
                  id={zeilenID(stelle)}
                  value={eintrag.wert}
                  active={eintrag.wert === wert}
                  selected={stelle === aktiv}
                  aria-selected={stelle === aktiv}
                  ref={stelle === aktiv ? zeigen : undefined}
                  onMouseMove={(ereignis) => {
                    if (zeiger.bewegt(ereignis)) {
                      setzeAktiv(stelle)
                    }
                  }}
                  className={zeilenklassen}
                >
                  {eintrag.name}
                  {eintrag.zusatz !== undefined && (
                    <>
                      {' '}
                      <span className="text-xs text-leise">
                        {eintrag.zusatz}
                      </span>
                    </>
                  )}
                </Combobox.Option>
              ))}
            </Combobox.Options>
          ) : (
            // **Stumm in der Liste, ansagend in der Live-Region**: derselbe Satz
            // zweimal vorgelesen hülfe niemandem.
            <Combobox.Empty
              aria-hidden="true"
              className="px-3 py-2 text-sm text-leise"
            >
              {meldung}
            </Combobox.Empty>
          )}

          {/* **Unter der Liste und nicht in ihr**: ein Satz ist keine Zeile,
              die man wählen könnte, und er rollt nicht mit. Angesagt wird er
              von der Live-Region. */}
          {gefunden > treffer.length && (
            <p
              aria-hidden="true"
              className="border-t border-kante px-3 py-2 text-sm text-leise"
            >
              {begrenzt}
            </p>
          )}
        </Combobox.Dropdown>
      </Combobox>

      {/* **Die Live-Region steht am Ende von `body`, nicht im Baum des
          Feldes** — über das Portal von Mantine, das die Combobox ohnehin
          mitbringt, statt `react-dom` an eine weitere Stelle zu holen
          (`docs/ABHAENGIGKEITEN.md`): umschließt ein `<label>` das Feld (die Einstellungen, die
          Suche), wäre ihr Text ein Teil von dessen Namen — aus „Zeitzone"
          würde bei offener Liste „Zeitzone Treffer: 420".

          **Ohne Rolle `status`**: ein leeres `status` auf jeder Seite mit
          einer Auswahl träfe jede Abfrage nach einer Meldung der Seite
          (`werkstattkopie.test.tsx` fragt genau so). */}
      <Portal>
        <p
          aria-live="polite"
          aria-atomic="true"
          data-suchauswahl=""
          className="sr-only"
        >
          {meldung}
        </p>
      </Portal>
    </div>
  )
}

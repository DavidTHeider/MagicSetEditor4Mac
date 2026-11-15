// **Die Schublade** — ein Seitenbereich, der von rechts über die
// Arbeitsfläche fährt: UX-08.AK-01
// (#1680),
// nach den Artboards *Eigenschaften* und *Werkzeuge* des Entwurfs
// (`docs/PROTOTYP-0.6.0.md`), auf Mantine wie der `Dialog`
// ([0125](../../../docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md)).
//
// # Wofür sie da ist — und wofür nicht
//
// **Für Details zum offenen Objekt**, die man selten braucht und nie neben
// der Arbeit stehen haben will: die Eigenschaften des Sets, seine Werkzeuge.
// Der Musterkatalog sagt: *Drawer für Details*, *Modal nur bei begründeter
// Unterbrechung*. Die Arbeitsfläche bleibt darunter stehen — die Schublade
// verdeckt sie, sie ersetzt sie nicht.
//
// # Was sie zusagt
//
//   - Der Fokus geht beim Öffnen hinein, Tab und ⇧Tab bleiben darin —
//     **Mantine**. Wohin er geht, sagt der Inhalt über `data-autofocus`.
//   - **Der Fokus kehrt zum Auslöser zurück** (S06.NAT-02) — **eigen**, aus
//     demselben Grund wie beim `Dialog`: die Aufrufer hängen sie aus, statt
//     sie zu schließen, und dann läuft Mantines Rückgabe nie.
//   - **Escape schließt das Innerste zuerst** — **eigen**, und das ist ein
//     Messergebnis: Mantine schließt in der Capture-Phase am Fenster
//     (`use-modal.mjs`, Zeile 15, Mantine 9.6.1), also **bevor** die
//     Suchleiste der Konsole oder eine Rückfrage die Taste überhaupt sieht.
//     Ein Druck schlösse beide. Hier gilt: die Schublade schließt, wenn der
//     Fokus in ihr steht **und** niemand darin das Escape mit
//     `preventDefault` für sich genommen hat.
//   - Ein Knopf zum Schließen steht im Kopf, benannt und per Tastatur
//     erreichbar — **eigen**, weil Mantines `CloseButton` Mantines Farben
//     trägt und keinen der gemessenen Fokusringe (`docs/KONTRASTE.md`).
//   - **Genannt wird sie über ihren Titel, beschrieben wird sie nicht.**
//     Mantines `Drawer.Body` hängt sich als `aria-describedby` an den Kasten
//     (`ModalBaseContent.mjs`, Zeile 37, Mantine 9.6.1) — dann läse
//     VoiceOver beim Öffnen die ganze Stilseite als Beschreibung vor. Der
//     Inhalt steht deshalb ohne `Drawer.Body` im Kasten.
//
// # Und was sie nicht zusagt
//
// **Kein `aria-hidden` auf dem Rest der Seite.** Mantine setzt es nicht —
// nachgelesen in `ModalBase`, `Drawer` und `Portal`; es gibt dort keine
// Stelle, die es schreibt. Was darunter liegt, sperrt der Aufrufer mit
// `inert` (`Seiteninhalt.tsx`); ob WKWebView und VoiceOver das achten, zeigt
// erst der native Pflichtpilot (UX-07.AK-01).
import { Drawer } from '@mantine/core'
import { useEffect, useRef } from 'react'

import { useRueckkehr } from '../haken/useRueckkehr'

import type { ReactNode } from 'react'

interface SchubladenEigenschaften {
  // titel ist die Überschrift der Schublade — und ihr zugänglicher Name.
  titel: string

  // schliessenBezeichnung benennt den Knopf, der sie schließt.
  schliessenBezeichnung: string

  schliessen: () => void

  // breite gibt die Schublade weiter auf — der Symboleditor im größeren
  // Zustand (AP-12, #2369). Ohne Angabe gilt die eine Breite des Entwurfs.
  breite?: string | undefined

  children: ReactNode
}

// **Die Breite des Entwurfs**: 460 von 1.280 Punkten, gut ein Drittel. Unter
// dem Fenster, das die Anwendung mindestens hat (1.024), bleiben daneben
// knapp 560 Punkte Arbeitsfläche — genug, dass man sieht, worüber die
// Schublade spricht.
const schmal = '28.75rem'

// kreuz ist das Zeichen auf dem Schließknopf — Gestaltung, kein Wort: der
// Name steht im `aria-label`.
const kreuz = '×'

const schliessklassen =
  'flex h-8 w-8 items-center justify-center rounded-md text-leise hover:bg-betont hover:text-text focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring'

export function Schublade({
  titel,
  schliessenBezeichnung,
  schliessen,
  breite = schmal,
  children,
}: SchubladenEigenschaften) {
  // **Die Rückgabe des Fokus bleibt eigen** — der Grund steht beim `Dialog`
  // und gilt hier genauso.
  // **Nur beim echten Erscheinen und Verschwinden** (`useRueckkehr`, G11).
  const kasten = useRef<HTMLDivElement>(null)
  useRueckkehr(kasten)

  useEffect(() => {
    const aufTaste = (ereignis: KeyboardEvent) => {
      if (ereignis.key !== 'Escape' || ereignis.isComposing) {
        return
      }

      // **Der Fokus steht woanders** — in der Befehlspalette, die über der
      // Schublade aufgegangen ist. Dann gehört ihr das Escape. `body` zählt
      // nicht als woanders: dorthin fällt der Fokus, wenn das Element unter
      // ihm verschwindet, und dann soll Escape trotzdem schließen.
      const fokus = document.activeElement
      if (
        fokus !== null &&
        fokus !== document.body &&
        kasten.current?.contains(fokus) !== true
      ) {
        return
      }

      // **Erst, wenn das Ereignis durch ist.** Diese Zeile läuft in der
      // Capture-Phase und damit vor allen anderen; ob jemand darin die Taste
      // verbraucht hat, steht erst fest, wenn alle sie gesehen haben. Das
      // Ereignis behält `defaultPrevented` über seine Zustellung hinaus.
      setTimeout(() => {
        if (!ereignis.defaultPrevented) {
          schliessen()
        }
      }, 0)
    }

    window.addEventListener('keydown', aufTaste, true)

    return () => {
      window.removeEventListener('keydown', aufTaste, true)
    }
  }, [schliessen])

  return (
    <Drawer.Root
      opened
      onClose={schliessen}
      closeOnEscape={false}
      position="right"
      size={breite}
      padding={0}
      styles={stile}
      // **Ohne Mantines Übergang**: die Schublade wird offen gezeichnet, und
      // Mantine spielt dann keinen. Sie schiebt sich über `einschieben`
      // herein (AP-4, #2361) — in `--dauer-mittel`, unter wenig Bewegung 0 ms.
      transitionProps={{ duration: 0 }}
    >
      {/* **Derselbe Schleier wie beim Dialog**: die Textfarbe bei vierzig
          Hundertsteln, eine gemessene Farbe (`docs/KONTRASTE.md`). */}
      <Drawer.Overlay
        backgroundOpacity={0.4}
        color="var(--mse-text)"
        className="auftauchen"
      />

      <Drawer.Content ref={kasten} aria-label={titel} className="einschieben">
        <Drawer.Header>
          <Drawer.Title>{titel}</Drawer.Title>

          <button
            type="button"
            aria-label={schliessenBezeichnung}
            onClick={schliessen}
            className={schliessklassen}
          >
            <span aria-hidden="true" className="text-lg leading-none">
              {kreuz}
            </span>
          </button>
        </Drawer.Header>

        {children}
      </Drawer.Content>
    </Drawer.Root>
  )
}

// **Über `styles` je Teil und nicht als `style` am Kasten** — und beides ist
// nachgelesen, nicht vermutet:
//
//   - Mantines Regeln standen bis #2451 ungeschichtet in `styles.css`, und
//     Ungeschichtetes gewann gegen jede Schicht — eine Klasse `text-lg` am
//     Titel verlor gegen Mantines `font-size`. Seither liegen sie in
//     `@layer mantine` unter den Utilities; die Stile bleiben trotzdem hier,
//     wegen des nächsten Punkts. Ein Stil am Element gewinnt ohnehin.
//   - **`Drawer.Content` reicht sein `style` auch an den Innenrahmen weiter**
//     (`DrawerContent.mjs`, Zeilen 15–26: `getStyles("inner", { style })`),
//     und der liegt über dem ganzen Fenster. Eine Fläche dort deckte die
//     Karten zu, statt sie abzudunkeln. `styles` trifft genau den Teil, den
//     es nennt.
//
// **Eine Kante statt eines Schattens** — Vermerk 0065, Festlegung 3. Die
// Fläche ist die der Anwendung, nicht Mantines.
const stile = {
  content: {
    background: 'var(--mse-flaeche)',
    color: 'var(--mse-text)',
    borderLeft: '1px solid var(--mse-kante)',
    boxShadow: 'none',
  },
  header: {
    background: 'var(--mse-flaeche)',
    borderBottom: '1px solid var(--mse-kante)',
    padding: '0.75rem 1rem',
  },
  // **Die Stimme der Überschriften** — die Anzeigeschrift, wie im Entwurf.
  title: {
    fontFamily: 'var(--font-anzeige)',
    fontSize: '1.125rem',
    fontWeight: 600,
    lineHeight: 1.3,
  },
}

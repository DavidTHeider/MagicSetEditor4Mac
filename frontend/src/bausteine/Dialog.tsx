// **Der eine modale Dialog** — Werkstück U2 von Phase 5.1
// (#1413),
// seit UX-07 auf Mantine
// (#1679,
// [0125](../../../docs/entscheidungen/0125-mantine-als-bauteilbibliothek.md)).
//
// # Wofür er da ist — und wofür nicht
//
// §6: keine verschachtelten Dialoge, Bearbeitung im Seitenbereich. Modal ist
// in dieser Anwendung genau dreierlei: die Befehlspalette, die Einführung und
// die Kommandozentrale.
//
// **Eine Rückfrage ist kein Dialog** — dafür gibt es `Entscheidung`.
//
// # Was er zusagt
//
//   - Der Fokus geht beim Öffnen in den Kasten — **Mantine**.
//   - Tab und ⇧Tab bleiben darin — **Mantine**.
//   - Escape schließt, auch wenn der Fokus den Kasten verlassen hat —
//     **Mantine**.
//   - **Der Fokus kehrt zum Auslöser zurück** — **eigen**, und warum, steht
//     unten am Effekt.
//   - **Neu seit Mantine**: das Rollen dahinter steht still
//     (`react-remove-scroll`, `ModalBase.mjs`, Zeile 49). Der Kasten trägt
//     `aria-modal`; **`aria-hidden` auf dem Rest setzt Mantine nicht.**
//
//     Nachgelesen in `ModalBase`, `Modal`, `Drawer` und `Portal` (Mantine
//     9.6.1) und in ihren Abhängigkeiten: keine Stelle schreibt es. Ob
//     WKWebView und VoiceOver `aria-modal` allein achten, zeigt erst der
//     native Pflichtpilot (UX-07.AK-01).
//
// # Was die eigene Fassung war, und warum sie geht
//
// **Sie hat gehalten, was sie zusagte** — und §15, Satz 12 sagt: was
// funktioniert, wird nicht ohne Grund neu geschrieben. Der Grund steht in
// UX-07.AK-01: die nativen Pflichtprüfungen gelten ausdrücklich „Palette" und
// „Overlays". Man erfährt nur an einem **echten** Overlay in WKWebView, ob
// Mantine dort trägt — und dieses hier ist das einzige, das die Anwendung
// hat.
//
// **Was dabei gewonnen ist**, steht oben in der letzten Zeile: der
// stillstehende Hintergrund, den die eigene Fassung nicht hatte.
//
// **Und was verloren geht**, gehört dazu: die Fokusfalle liegt jetzt in einer
// fremden Bibliothek. Dass sie trägt, prüft weiterhin `dialog.test.tsx` —
// derselbe Lauf, derselbe Anspruch. **Die Rückgabe blieb eigen**, und das ist
// kein Versehen: der Grund steht am Effekt.
import { Modal } from '@mantine/core'
import { useEffect, useRef } from 'react'

import { fokusZurueck, merkeRueckkehr } from './fokus'

import type { ReactNode } from 'react'

// Ausrichtung sagt, wo der Kasten auf dem Schleier steht: die Palette oben,
// wo man eine Eingabe erwartet; die Einführung in der Mitte, wo man liest.
export type Ausrichtung = 'oben' | 'mitte'

// **Der Abstand vom oberen Rand je Ausrichtung.**
//
// Mantine kennt dafür `yOffset`; „mitte" entsteht über `centered` und braucht
// keinen. Die Zahlen sind die der eigenen Fassung: `p-16` waren 4 rem, `p-8`
// zwei — und die zweite trägt hier `centered`, weshalb sie als Rand bleibt.
const abstaende: Record<Ausrichtung, string> = {
  oben: '4rem',
  mitte: '2rem',
}

interface DialogEigenschaften {
  // bezeichnung benennt den Dialog — `aria-label`, denn nicht jeder hat eine
  // Überschrift im Kasten (die Palette hat ein Suchfeld).
  bezeichnung: string

  ausrichtung?: Ausrichtung

  schliessen: () => void

  children: ReactNode
}

export function Dialog({
  bezeichnung,
  ausrichtung = 'oben',
  schliessen,
  children,
}: DialogEigenschaften) {
  // **Die Rückgabe des Fokus bleibt eigen, und das ist ein Messergebnis.**
  //
  // Mantine gibt ihn zurück, wenn ein Modal **schließt** — hier wird es
  // **ausgehängt**: die drei Aufrufer schreiben `{offen && <Dialog …>}`, und
  // beim Aushängen läuft Mantines Effekt mit `opened=false` nie. Gemessen am
  // 2026-09-20: ohne diese sieben Zeilen steht der Fokus nach Escape
  // nirgends, und der Lauf daneben lief fünfzehn Sekunden in seine Frist.
  //
  // **Die Alternative wäre eine andere Fläche** — `offen` als Eigenschaft
  // statt der Bedingung beim Aufrufer. Sie ist größer, berührt drei Dateien
  // und gewänne nichts, was diese sieben Zeilen nicht auch geben.
  const rueckkehr = useRef<Element | null>(null)

  useEffect(() => {
    rueckkehr.current = merkeRueckkehr()

    return () => {
      fokusZurueck(rueckkehr.current)
    }
  }, [])

  // **Escape schließt, wenn niemand darin es genommen hat** — G12 aus dem
  // Zweitblick über Phase 6.3. Mantine schloss in der Capture-Phase am
  // Fenster, vor einem Hinweis oder einer Rückfrage im Kasten; ein Druck
  // schloss alles. Dasselbe Muster wie die `Schublade`: horchen in der
  // Capture-Phase, entscheiden, wenn das Ereignis durch ist. Der Fokus darf
  // dabei woanders stehen — das war schon Mantines Zusage.
  useEffect(() => {
    const aufTaste = (ereignis: KeyboardEvent) => {
      if (ereignis.key !== 'Escape' || ereignis.isComposing) {
        return
      }

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
    // **Die zusammengesetzte Form und nicht das kurze `<Modal>`** — und das
    // ist ein Fund, kein Geschmack: `Modal` reicht alles Übrige an seine
    // **Wurzel** weiter, nicht an den Kasten. Ein `aria-label` landete dort
    // am Schleiercontainer, und der Knoten mit `role="dialog"` stand ohne
    // Namen da. Nachgelesen in `Modal.mjs`, Zeile 34, und `ModalBaseContent.mjs`,
    // Zeile 33.
    <Modal.Root
      opened
      onClose={schliessen}
      closeOnEscape={false}
      centered={ausrichtung === 'mitte'}
      yOffset={abstaende[ausrichtung]}
      padding={0}
      // **Die Breite bestimmt der Aufrufer und nicht Mantine.**
      //
      // `Modal` bringt eine eigene mit (`size="md"`, rund 400 Punkte), und
      // die schnitt den Einführungskasten von `max-w-3xl` auf ein Drittel
      // zusammen — **gesehen in der Sichtprüfung am 2026-09-21**, nicht in
      // einem Lauf: ein DOM-Test misst keine Breite, und jsdom rechnet kein
      // Layout (§3).
      //
      // `100%` gibt sie zurück: die drei Aufrufer tragen `max-w-md`,
      // `max-w-2xl` und `max-w-3xl`, und jeder weiß selbst, wie breit sein
      // Inhalt sein will.
      size="100%"
      // **Ohne Mantines Übergang**: der Dialog wird offen gezeichnet, und
      // Mantine spielt dann keinen. Er taucht über `auftauchen` auf
      // (AP-4, #2361) — in `--dauer-mittel`, unter wenig Bewegung 0 ms.
      transitionProps={{ duration: 0 }}
    >
      {/* **Der Schleier ist derselbe wie vorher**: die Textfarbe bei vierzig
          Hundertsteln. Eine fremde Farbe stünde in keiner Messung
          (`docs/KONTRASTE.md`). */}
      <Modal.Overlay
        backgroundOpacity={0.4}
        color="var(--mse-text)"
        className="auftauchen"
      />

      {/* **Der Kasten ist durchsichtig und trägt keine Kante.** Die drei
          Aufrufer bringen ihren eigenen mit — gerundet, mit Kante, auf
          `bg-flaeche` —, und der ist gemessen. Ein zweiter darum wäre der
          Zierrahmen, den Vermerk 0065, Festlegung 3 verwirft. */}
      <Modal.Content
        aria-label={bezeichnung}
        bg="transparent"
        style={durchsichtig}
        className="auftauchen"
      >
        {children}
      </Modal.Content>
    </Modal.Root>
  )
}

// durchsichtig nimmt dem Kasten Fläche und Schatten — und stellt ihn mittig.
//
// **Als Stil und nicht als Klasse**: Mantine setzt Fläche und Schatten über
// eigene Merkmale am Element; eine Tailwind-Klasse daneben verlöre gegen sie,
// und `shadow-none` wäre außerdem ein Fund für
// `internal/werkzeuge/schlagschatten_test.go`.
//
// **Und die Mitte ist ein Fund der Sichtprüfung vom 2026-09-21.** Die eigene
// Fassung war ein `flex justify-center` über die ganze Breite; mit
// `size="100%"` füllt Mantines Kasten sie zwar, stellt seinen Inhalt aber
// nach links. Die Befehlspalette (`max-w-md`) hing damit in der linken Ecke
// — **kein DOM-Lauf sieht das**, jsdom rechnet kein Layout (§3).
const durchsichtig = {
  background: 'transparent',
  boxShadow: 'none',
  display: 'flex',
  justifyContent: 'center',
}

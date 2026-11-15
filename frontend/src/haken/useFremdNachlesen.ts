// **Das Aufgeschlagene liest nach, wenn ein anderer schrieb** — Befund 2 des
// Nachlaufs zum Gesamtreview über 0.6.0 (#1660).
//
// Der fremde Stand (`useBestandsstand`, #156) erreichte seit #2062 die Listen
// der Seiten. Ein offenes Deck oder Turnier las nur, wer es aufschlug — die
// Liste zeigte den neuen Namen, die Werkbank daneben den alten.
//
// # Warum ein eigener Haken und nicht `useLaden`
//
// **Weil das Aufgeschlagene keinen Ladeeffekt hat.** Es kommt aus einem Klick
// (`waehle`), und sein Bezug steht in einer Referenz. Der Haken sagt nur
// **wann**; **was** nachgelesen wird, weiß die Seite.
import { useEffect, useRef } from 'react'

// useFremdNachlesen ruft `nachlesen`, sobald der fremde Stand steigt —
// **nicht** beim ersten Zeichnen: dann liest die Seite ohnehin.
export function useFremdNachlesen(
  fremdstand: number,
  nachlesen: () => void,
): void {
  const gesehen = useRef(fremdstand)

  useEffect(() => {
    if (gesehen.current === fremdstand) {
      return
    }

    gesehen.current = fremdstand
    nachlesen()
  }, [fremdstand, nachlesen])
}

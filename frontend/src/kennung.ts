// neueKennung gibt eine Kennung in der Gestalt `7f3a-21c9` — dieselbe für
// die Fehlerfläche (F01) und für einen unerwarteten Fehler in einer Meldung
// oder einem Ruf (F03, #2342). Im Protokoll steht sie als `fehlerkennung`.
export function neueKennung(): string {
  const zahl = [...crypto.getRandomValues(new Uint8Array(4))]
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')

  return `${zahl.slice(0, 4)}-${zahl.slice(4)}`
}

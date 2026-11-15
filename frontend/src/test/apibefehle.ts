// **Die Vorbelegungen der Kommandozentrale** (#1423): kein Befehl, ein stilles
// Ergebnis, ein stiller Abbruch — wer die Zentrale meint, belegt sie selbst.
export function ohneBefehle() {
  return Promise.resolve({ befehle: [], fehler: null })
}

export function stillesErgebnis(name: string) {
  return Promise.resolve({ name, text: '', dauerMs: 0, fehler: null })
}

export function stillerAbbruch() {
  return Promise.resolve({ fehler: null })
}

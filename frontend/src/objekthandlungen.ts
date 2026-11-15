// **Die Handlungsliste des gewählten Objekts, wie Go sie entgegennimmt** —
// GESTE (#2358).
// Dieselben Felder wie `app.Objekthandlungen` und `app.Objektausloesung` in
// `internal/app/objektmenue.go`; der Vertragswächter hält beide zusammen.
//
// Was sie ist und warum, steht dort und in `haken/useHandlungsliste.ts`;
// wie sie nach Go kommt, in `objekthandlungsraum.tsx`.

// Objekthandlung ist ein Eintrag des Menüs.
export interface Objekthandlung {
  schluessel: string
  name: string
  gesperrt: boolean

  // kuerzel ist die Taste in der Leiste — AP-7 (#2364); Go nimmt nur die
  // Wörter aus [objektkuerzel] an (`objektkuerzel` in `objektmenue.go`).
  kuerzel?: Objektkuerzel
}

// **Die Kürzel eines Objektmenüs**, wie der Entwurf `Kontext09` sie für das
// Menü „Karte“ zeichnet — mit der Schreibweise von `aria-keyshortcuts` und
// dem Zeichen, wie macOS es im Menü setzt. Dieselbe Liste steht in Go; der
// Vertragstest hält die Wörter zusammen.
export const objektkuerzel = {
  'befehl+return': { tasten: 'Meta+Enter', zeichen: '⌘↩' },
  'befehl+d': { tasten: 'Meta+D', zeichen: '⌘D' },
  'befehl+backspace': { tasten: 'Meta+Backspace', zeichen: '⌘⌫' },
  'bild auf': { tasten: 'PageUp', zeichen: '⇞' },
  'bild ab': { tasten: 'PageDown', zeichen: '⇟' },
} as const

export type Objektkuerzel = keyof typeof objektkuerzel

// Objekthandlungen ist die Liste eines Objekts — leer, wenn keines gewählt
// ist.
export interface Objekthandlungen {
  titel: string
  art: string
  kennung: string
  epoche: number
  eintraege: Objekthandlung[]
}

// Objektausloesung ist, was ein Klick im Menü der Leiste ins Fenster meldet.
export interface Objektausloesung {
  art: string
  kennung: string
  epoche: number
  schluessel: string
}

// keinObjekt ist die Meldung, wenn nichts gewählt ist.
export const keinObjekt: Objekthandlungen = {
  titel: '',
  art: '',
  kennung: '',
  epoche: 0,
  eintraege: [],
}

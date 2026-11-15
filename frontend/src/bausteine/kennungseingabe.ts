// **Ein Feld, dessen Wert genau so ankommen muss, wie er getippt wurde** —
// #2389.
//
// Gefunden bei der Ausgangsmessung P0 (#2336): im Feld „Druck" schrieb
// macOS aus `p0-blitz` ein `P0-blitz`, und das eigene Exemplar passte im
// Handel nicht mehr. Eine Kennung, ein Code, eine Adresse, ein Name zur
// Bestätigung — die Liste steht in `docs/FELDPRUEFUNG.md` („Kennungsfelder"),
// und `kennungsfelder.test.ts` hält sie gegen den Quelltext.
//
// **Alle drei, und nicht eines**: `autocapitalize` ist die große erste
// Stelle, `autocorrect` die Ersetzung durch ein Wort, das das System kennt,
// und `spellcheck` die Unterringelung, über die eine Ersetzung angeboten
// wird. Welche WebKit unter macOS beachtet, misst die Hüllenprobe.
export const kennungseingabe = {
  autoCapitalize: 'off',
  autoCorrect: 'off',
  spellCheck: false,
} as const

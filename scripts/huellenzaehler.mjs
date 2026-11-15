// **Der Zähler der Handlungen** — gezogen aus `huellenprobe/ausgangswerte.mjs`
// für G01 (#2248),
// damit die vier Personawege **dieselbe** Zählung haben wie die fünf
// Kernwege von UX-01.AK-03 — eine zweite daneben wäre eine zweite Wahrheit.
//
// # Was gezählt wird
//
// **Eine Handlung ist, was ein Mensch einmal tut**: ein Klick, ein Kürzel,
// ein getippter Wert (samt ⇥, das ihn schreibt), eine Wahl in einer Liste,
// eine Wahl im Dateidialog. **Rollen zählt nicht** — die Probe rollt ein
// Ziel in den Blick, bevor sie klickt. Was vorher angelegt wird, damit es den
// Weg überhaupt gibt, entsteht über `mse` und zählt nicht.
//
// **Die Dauer ist Maschinenzeit** und wird so genannt: vom ersten Handgriff
// bis die Wirkung im Bestand steht. **Sie ist keine Bearbeitungszeit eines
// Menschen** — ein Mensch liest, sucht und zögert, und das misst sie nicht.
// **Fehlversuche** hat die Probe keine, weil sie den Weg kennt; die Spalte
// steht trotzdem da, damit die Tafel dieselbe Gestalt hat wie die eines
// Menschen.

// sichtbarerKnopf ist der Ausdruck für den sichtbaren Knopf mit diesem Wort —
// **auch mit einem Zeichen davor** (»+ Karte anlegen«, »⎘ Kopieren«).
export const sichtbarerKnopf = (text) => `[...document.querySelectorAll('button, a, [role=button], [role=tab], summary')].find((e) => e.checkVisibility() && (e.innerText ?? '').replace(/\\s+/g, ' ').trim().replace(/^[^\\p{L}\\p{N}]+/u, '') === ${JSON.stringify(text)})`

// zaehler zählt die Handlungen eines Wegs und stoppt die Maschinenzeit.
//
// `uhr` ist austauschbar, damit sich die Zählung ohne Fenster prüfen lässt.
export function zaehler(p, uhr = Date.now) {
  let handlungen = 0
  let anfang = 0

  return {
    beginne() {
      handlungen = 0
      anfang = uhr()
    },
    async klick(text) {
      handlungen += 1
      await p.seite(`(${sichtbarerKnopf(text)})?.scrollIntoView({ block: 'center' }); return true`)
      await p.klickeAuf(sichtbarerKnopf(text), `»${text}«`)
    },
    // text klickt, was `p.klickeText` findet — auch eine Beschriftung oder
    // einen Eintrag mit einer zweiten Zeile, die `klick` genau nimmt (G01).
    async text(text) {
      handlungen += 1
      await p.klickeText(text)
    },
    async klickAuf(ausdruck, beschreibung, art = {}) {
      handlungen += 1
      await p.seite(`(${ausdruck})?.scrollIntoView({ block: 'center' }); return true`)
      await p.klickeAuf(ausdruck, beschreibung, art)
    },
    async taste(taste, ...mit) {
      handlungen += 1
      await p.taste(taste, ...mit)
    },
    async tippe(text) {
      handlungen += 1
      await p.tippeInsFeld(text)
    },
    async menue(...pfad) {
      handlungen += 1
      await p.klick(...pfad)
    },
    // wahl zählt eine Wahl, die die Probe über die Seite setzt — ein Mensch
    // wählt dort aus einer Liste, einmal.
    async wahl(ausdruck) {
      handlungen += 1
      await p.seite(ausdruck)
    },
    // dialog zählt die Wahl im Dateidialog — die Probe legt sie bereit.
    async dialog(pfad) {
      handlungen += 1
      await p.dialogantwort(pfad)
    },
    ende(weg) {
      const dauer = uhr() - anfang
      p.pruefe(`${weg}: ${String(handlungen)} Handlungen, 0 Fehlversuche, ${(dauer / 1000).toFixed(1).replace('.', ',')} s Maschinenzeit`, true)

      return { handlungen, dauer }
    },
  }
}

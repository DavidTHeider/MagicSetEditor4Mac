// **Der Weg in die Zwischenablage — mit seinem Nein** — Befund 5 des
// Nachlaufs zum Zweitblick über Phase 5.1
// (#1405).
//
// # Warum es diese Datei gibt
//
// **Weil `navigator.clipboard.writeText` ablehnen kann und es an zwei Stellen
// niemand aufgefangen hat.** Ohne Erlaubnis — und das ist der Normalfall, wenn
// das Fenster den Fokus nicht hat — blieb eine Zusage ohne Empfänger stehen,
// und die Oberfläche zeigte **nichts**: kein „Kopiert", keine Meldung. Wer
// klickte, wusste nicht, ob es geklappt hat.
//
// **Und sie kann fehlen.** In einem Fenster ohne sicheren Kontext gibt es
// `navigator.clipboard` gar nicht; ein Zugriff darauf wäre ein Fehler und
// keine Ablehnung.

// inDieZwischenablage legt einen Text ab und sagt, ob es geklappt hat.
//
// **Ein `boolean` und keine Ausnahme**: der Aufrufer zeigt „Kopiert" oder eine
// Meldung, und beides ist ein normaler Ausgang.
export async function inDieZwischenablage(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text)

    return true
  } catch {
    return false
  }
}

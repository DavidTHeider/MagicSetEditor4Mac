// **Die Sicherung mit Kennwort, am nativen Fenster** — OPS-06
// (#1723):
//
// - AK-02: unter *Einstellungen › Daten und Sicherung* stehen Kennwort und
//   Wiederholung vor dem Sichern; zwei verschiedene halten den Knopf an und
//   sagen es am Feld,
// - zwei gleiche geben ihn frei, und er öffnet den Speichern-Dialog von
//   macOS — **abgebrochen**, denn eine Probe legt keine Datei an einem Ort
//   ab, den ein Mensch gewählt hätte.
//
// Dass eine Sicherung mit Kennwort hin und zurück geht, belegen
// `internal/app/sicherungsrundlauf_test.go` (leerer Bestand, zwei Profile,
// Tabelle für Tabelle) und `internal/flaechen/sicherungskennwort_test.go`.

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

const feld = (name) => `[...document.querySelectorAll('label')].find((l) => l.innerText.trim().startsWith(${JSON.stringify(name)}))?.querySelector('input')`

const knopf = "[...document.querySelectorAll('button')].find((b) => b.innerText.includes('Bestand sichern'))"

// imFeld wartet, **bis der Fokus angekommen ist** — der Klick geht über die
// Ereignisschlange und kommt später an als der nächste Ruf in die Seite.
// Bis zum 2026-10-05 las `tippeInsFeld` den Ausgangswert dann noch aus dem
// ersten Feld, und in drei von fünf Läufen stimmte im Fenster alles, nur die
// Erwartung nicht.
async function imFeld(p, name) {
  await p.warte(`der Fokus steht in »${name}«`, () => p.seite(`return document.activeElement === (${feld(name)})`))
}

export default async function ops06(p) {
  await p.starte()

  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)
  await p.klickeText('Daten und Sicherung')
  await p.ruhe()
  await p.warte('das Kennwortfeld steht da', () => p.seite(`return !!(${feld('Kennwort (freiwillig)')})`), 20000)

  p.notiz('AK-02 — zwei verschiedene Kennwörter halten den Knopf an')
  await p.klickeAuf(feld('Kennwort (freiwillig)'), 'das Kennwortfeld')
  await imFeld(p, 'Kennwort (freiwillig)')
  await p.tippeInsFeld('korrekt pferd batterie')
  await p.klickeAuf(feld('Kennwort wiederholen'), 'die Wiederholung')
  await imFeld(p, 'Kennwort wiederholen')
  await p.tippeInsFeld('korrekt pferd')

  p.pruefe('die Meldung steht am Feld', await p.seite("return document.body.innerText.includes('Die beiden Kennwörter stimmen nicht überein.')"))
  p.pruefe('der Knopf ist angehalten', await p.seite(`return ${knopf}?.disabled === true`))
  await p.seite(`${knopf}.scrollIntoView({ block: 'center' })`)
  await p.bild('kennwort-verschieden')

  p.notiz('AK-02 — zwei gleiche geben ihn frei, und er öffnet den Dialog von macOS')
  await p.tippeInsFeld(' batterie')
  p.pruefe('die Meldung ist fort', await p.seite("return !document.body.innerText.includes('Die beiden Kennwörter stimmen nicht überein.')"))
  p.pruefe('der Knopf ist frei', await p.seite(`return ${knopf}?.disabled === false`))
  p.pruefe('der Hinweis sagt, dass es keinen Rückweg gibt', await p.seite("return document.body.innerText.includes('Ein vergessenes Kennwort lässt sich nicht wiederherstellen')"))
  await p.bild('kennwort-gleich')

  await p.klickeAuf(knopf, 'Bestand sichern …')
  await p.dialogZeigtSich('sichern', '»Bestand sichern …« mit Kennwort')

  p.pruefe('⌘Q beendet', await p.beende())
}

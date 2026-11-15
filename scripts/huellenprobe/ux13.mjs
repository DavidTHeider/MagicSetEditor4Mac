// **Eine Datei ansehen, am nativen Fenster** — UX-13
// (#1685):
//
// - AK-01/03: der Bestand — eine Datei ohne bekannte Art — zeigt Name, Art,
//   Größe und Ort, **keine** Vorschau und **kein** *Öffnen*, nur *Im Finder
//   zeigen*,
// - AK-04: der Knopf dafür steht da; gedrückt wird er nicht, denn er risse
//   dem Prüfenden ein Finderfenster auf,
// - AK-01: das Protokoll der Probe — ein Text — erscheint als Text, mit
//   *Öffnen*.
//
// Welche Art was zeigt, belegen `internal/service/dateiansicht_test.go` und
// `frontend/src/bausteine/dateiansicht.test.tsx`; dieselbe Auskunft an jeder
// Fläche `internal/flaechen/dateiansicht_test.go`.

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

const knopf = (name) => `[...document.querySelectorAll('button')].find((b) => b.getAttribute('aria-label') === ${JSON.stringify(name)})`

export default async function ux13(p) {
  await p.starte()

  await p.klick('Hilfe', 'MagicSetEditor4Mac-Hilfe')
  await p.warte('die Hilfe steht da', async () => (await hier(p)) === 'Hilfe', 20000)
  await p.klickeText('Betrieb und Vertrauen')
  await p.wartetext('Was geprüft ist', 20000)

  p.notiz('AK-01/03 — der Bestand zeigt seine Angaben und nichts sonst')
  p.pruefe('neben dem Ort steht Ansehen', await p.seite(`return !!(${knopf('bestand.db ansehen')})`))
  await p.klickeAuf(knopf('bestand.db ansehen'), 'bestand.db ansehen')
  await p.wartetext('Unbekannte Art', 20000)

  const ansicht = await p.seite("return document.querySelector('section[aria-label=\"Die Datei\"]')?.innerText ?? ''")
  p.pruefe('Name und Art', ansicht.includes('bestand.db') && ansicht.includes('Unbekannte Art'))
  p.pruefe('eine Größe', /Größe\s+[\d.,]+ (B|KB|MB)/.test(ansicht), ansicht.split('\n').find((z) => z.startsWith('Größe')) ?? '')
  p.pruefe('keine Vorschau, mit Grund', ansicht.includes('sie könnte ein Programm sein'))

  p.notiz('AK-04 — im Finder zeigen ja, öffnen nein')
  p.pruefe('Im Finder zeigen steht da', await p.seite(`return !!(${knopf('bestand.db im Finder zeigen')})`))
  p.pruefe('Öffnen fehlt', await p.seite(`return !(${knopf('bestand.db mit seinem Programm öffnen')})`))
  await p.seite("document.querySelector('section[aria-label=\"Die Datei\"]').scrollIntoView({ block: 'center' })")
  await p.bild('bestand-angesehen')

  await p.klickeAuf(knopf('Ansicht schließen'), 'Ansicht schließen')
  await p.warte('die Ansicht geht wieder zu', () => p.seite("return !document.querySelector('section[aria-label=\"Die Datei\"]')"), 10000)

  // **Das Protokoll dieser Probe liegt in ihrem eigenen Ordner** — sein
  // Anfang als Text zeigt die Vorschau, ohne eine Datei des Menschen zu lesen.
  p.notiz('AK-01 — ein Text erscheint als Text')
  p.pruefe('das Protokoll liegt im Ordner der Probe', await p.seite("return document.body.innerText.includes('huelle-') && !!(" + knopf('mse.log ansehen') + ")"))
  await p.klickeAuf(knopf('mse.log ansehen'), 'mse.log ansehen')
  await p.warte('der Text steht da', () => p.seite("return !!document.querySelector('section[aria-label=\"Die Datei\"] pre')"), 20000)

  const protokoll = await p.seite("return document.querySelector('section[aria-label=\"Die Datei\"]').innerText")
  p.pruefe('Art Text, mit Öffnen', protokoll.includes('Text') && await p.seite(`return !!(${knopf('mse.log mit seinem Programm öffnen')})`))
  p.pruefe('der Anfang des Protokolls', /level=|"level"|INFO|DEBUG/.test(protokoll), protokoll.split('\n').find((z) => /level|INFO/.test(z))?.slice(0, 80) ?? '')
  await p.seite("document.querySelector('section[aria-label=\"Die Datei\"]').scrollIntoView({ block: 'center' })")
  await p.bild('protokoll-angesehen')

  p.pruefe('⌘Q beendet', await p.beende())
}

// **Einladen und vergleichen am nativen Fenster** — NET-01
// (#1705):
//
// - wo gehandelt wird, stehen „Meinen Code zeigen" und „Code eingeben" vorn
//   (AK-01), die Kennung von Hand dahinter,
// - der Code eines zweiten Rechners (Bert, nur Kommandozeile) wird geprüft,
//   und das Fenster zeigt **dieselben dreißig Ziffern**, die Bert für Annas
//   Code sieht,
// - „Stimmt überein" mit Haken bestätigt und gibt den Handel frei; dieselbe
//   Einladung ein zweites Mal wird benannt abgewiesen (AK-04).
//
// **Der Code kommt über JavaScript ins Feld und nicht über Tasten**: er ist
// fünfhundert Zeichen lang, und die Zwischenablage darf eine Probe nicht
// anfassen. Das Tippen in ein Textfeld prüfen die Oberflächenläufe.
const eingabefeld = '[data-einladungstausch] textarea:not([readonly])'
const setze = (text) => `(() => { const feld = document.querySelector('${eingabefeld}'); const setzer = Object.getOwnPropertyDescriptor(HTMLTextAreaElement.prototype, 'value').set; setzer.call(feld, ${JSON.stringify(text)}); feld.dispatchEvent(new Event('input', { bubbles: true })); return feld.value.length })()`
const vergleichscode = "return document.querySelector('[data-einladungsvergleich] [data-vergleichscode]')?.innerText ?? ''"
const eigenerCode = "return document.querySelector('[data-einladungstausch] textarea[readonly]')?.value ?? ''"
const zeilenwert = (ausgabe, name) => ausgabe.split('\n').find((z) => z.startsWith(`${name}\t`))?.split('\t')[1] ?? ''

export default async function net01(p) {
  const anna = p.ort()
  const bert = p.ort()

  p.notiz('1 — zwei Rechner mit je eigener Identität; Bert lädt an der Kommandozeile ein')
  for (const ort of [anna, bert]) {
    p.mse(ort, 'einstellungen', 'setzen', 'sprache', 'de')
    p.mse(ort, 'peer-anlegen')
  }
  const bertsCode = zeilenwert(p.mse(bert, 'einladung-erstellen', 'Bert', 'handel'), 'code')
  p.zeige(`Berts Code: ${bertsCode.length} Zeichen`)

  await p.starte({ ort: anna })
  await p.taste('5', 'befehl')

  p.notiz('2 — im Handeln stehen die zwei Handlungen vorn')
  // **Seit UXR-15 in der Schublade hinter „Neuer Tausch …"** (#1923).
  await p.klickeText('Neuer Tausch …')
  await p.klickeText('Gegenüber aufnehmen oder bestätigen')
  await p.wartetext('Meinen Code zeigen')
  p.pruefe('Code eingeben steht da', await p.seite("return [...document.querySelectorAll('[data-einladungstausch] button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Code eingeben')"))
  p.pruefe('die Kennung von Hand steht dahinter', !(await p.seite("return [...document.querySelectorAll('main label')].some((l) => l.checkVisibility() && l.innerText.trim().startsWith('Kennung des Gegenübers'))")))
  await p.bild('vorn')

  p.notiz('3 — Anna zeigt ihren Code, Bert prüft ihn an der Kommandozeile')
  await p.klickeText('Meinen Code zeigen')
  await p.klickeAuf("document.querySelector('[data-einladungstausch] input')", 'das Namensfeld')
  await p.tippeInsFeld('Anna')
  await p.klickeText('Code erzeugen')
  await p.warte('Annas Code steht da', async () => (await p.seite(eigenerCode)).startsWith('mse4mac-kontakt:1:'))
  const annasCode = await p.seite(eigenerCode)
  const bertSieht = zeilenwert(p.mse(bert, 'einladung-pruefen', annasCode), 'vergleich')
  p.zeige(`Bert sieht: ${bertSieht}`)
  await p.bild('eigener-code')

  p.notiz('4 — Anna gibt Berts Code ein und sieht dieselben Ziffern')
  await p.klickeText('Code eingeben')
  await p.warte('das Eingabefeld steht', () => p.seite(`return document.querySelector('${eingabefeld}') !== null`))
  p.pruefe('der Code steht im Feld', (await p.seite(`return ${setze(bertsCode)}`)) === bertsCode.length)
  await p.klickeText('Prüfen')
  await p.wartetext('Bert bestätigen')
  const annaSieht = await p.seite(vergleichscode)
  p.zeige(`Anna sieht: ${annaSieht}`)
  p.pruefe('dieselben dreißig Ziffern', annaSieht !== '' && annaSieht === bertSieht, `${annaSieht} gegen ${bertSieht}`)
  p.pruefe('der Haken ist nicht gesetzt', await p.seite("return document.querySelector('[data-einladungsvergleich] input[type=checkbox]')?.checked === false"))
  await p.bild('vergleich')

  p.notiz('5 — stimmt überein, mit Freigabe')
  await p.klickeText('Danach erlauben: Karten tauschen')
  await p.warte('der Haken ist gesetzt', () => p.seite("return document.querySelector('[data-einladungsvergleich] input[type=checkbox]')?.checked === true"), 5000)
  await p.klickeText('Stimmt überein')
  await p.wartetext('Bert ist bestätigt.')
  const kontakte = p.mse(anna, 'kontakte')
  p.pruefe('bestätigt und darf handeln', /^bestaetigt\t/m.test(kontakte) && kontakte.includes('darf\thandel'), kontakte)
  await p.bild('bestaetigt')

  p.notiz('6 — dieselbe Einladung ein zweites Mal')
  p.pruefe('der Code steht wieder im Feld', (await p.seite(`return ${setze(bertsCode)}`)) === bertsCode.length)
  await p.klickeText('Prüfen')
  await p.wartetext('schon angenommen')
  p.pruefe('kein Vergleich angeboten', !(await p.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && b.innerText.trim() === 'Stimmt überein')")))
  await p.bild('schon-angenommen')

  p.pruefe('⌘Q beendet', await p.beende())
}

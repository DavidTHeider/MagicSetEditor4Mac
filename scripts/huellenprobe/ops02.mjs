// **Ein Ordner für Sicherungen, am nativen Fenster** — OPS-02
// (#1719):
//
// - AK-01: ein Ordner, den es nicht gibt, wird beim Setzen abgewiesen — mit
//   einem Satz, der das Laufwerk nennt —, und im Bestand steht danach nichts,
// - ein Ordner, den es gibt, steht danach im Bestand,
// - das Textfeld schreibt mit ⏎ und nicht bei jedem Buchstaben: solange nur
//   getippt ist, steht im Bestand nichts.
//
// **Der Wert kommt über das Feld selbst** — gesetzt wie eine Eingabe, damit
// React ihn als Änderung sieht; die Taste ⏎ kommt nativ.

import { mkdirSync } from 'node:fs'
import { join } from 'node:path'

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

const feld = "document.querySelector('[data-einstellung=\"vorgabe.sicherungspfad\"]')"

// eintragen setzt den Entwurf wie eine Eingabe und fokussiert das Feld.
// **Erst ein nativer Klick ins Feld** — in WebKit fokussiert er nicht, aber
// er macht die Webansicht zum Ersten Antwortenden; ohne ihn erreicht das
// native ⏎ die Seite nicht (`docs/SICHTPRUEFUNG.md`).
async function eintragen(p, wert) {
  await p.klickeAuf(feld, 'das Feld für Sicherungen')
  return p.seite(`const f = ${feld}; f.focus(); Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ${JSON.stringify(wert)}); f.dispatchEvent(new Event('input', { bubbles: true })); return f.value`)
}

const gesetzt = (p, ort) => p.mse(ort, 'einstellungen').split('\n').find((z) => z.startsWith('vorgabe.sicherungspfad')) ?? ''

export default async function ops02(p) {
  const ort = p.ort()
  const vorhanden = join(ort.ort, 'Sicherungen')
  mkdirSync(vorhanden)
  await p.starte({ ort })

  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)
  await p.klickeText('Daten und Sicherung')
  await p.ruhe()
  // **Seit UXR-19 (#1927) in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Ordner und Dateinamen')
  await p.warte('das Feld für Sicherungen steht da', () => p.seite(`return !!${feld}`))

  p.notiz('Ein Ordner, den es nicht gibt — abgewiesen, mit Satz')
  const fehlt = join(ort.ort, 'abgezogen')
  await eintragen(p, fehlt)
  p.pruefe('solange nur getippt ist, steht im Bestand nichts', !gesetzt(p, ort).includes(fehlt), gesetzt(p, ort))
  await p.taste('Enter')
  // **Am Feld und nicht über der Seite** (§6) — die Meldung steht im Kasten
  // der Einstellung, unter ihrem Feld.
  const satz = await p.warte('der Satz steht am Feld', () => p.seite(`const a = ${feld}.closest('label').parentElement.querySelector('[role=alert]'); return a && a.innerText.includes('nicht erreichbar') ? a.innerText : false`), 20000).catch(async (grund) => {
    p.zeige(`Fokus: ${await p.seite('return document.activeElement?.getAttribute(\'data-einstellung\') ?? document.activeElement?.tagName')}, Feld: ${await p.seite(`return ${feld}.value`)}`)
    p.zeige(`Meldungen: ${await p.seite("return [...document.querySelectorAll('[role=alert],[role=status]')].map((e) => e.innerText.trim()).filter(Boolean).join(' | ').slice(0, 400)")}`)
    p.zeige(`im Bestand: ${gesetzt(p, ort)}`)
    throw grund
  })
  p.pruefe('er nennt das Laufwerk', satz.includes('Laufwerk'), satz.replaceAll('\n', ' '))
  p.pruefe('über der Seite steht kein Ladefehler', !(await p.seite("return document.body.innerText.includes('Das ließ sich nicht laden')")))
  p.pruefe('und im Bestand steht der Ordner nicht', !gesetzt(p, ort).includes(fehlt), gesetzt(p, ort))
  await p.bild('ordner-fehlt')

  p.notiz('Ein Ordner, den es gibt — steht danach im Bestand')
  await eintragen(p, vorhanden)
  await p.taste('Enter')
  const danach = await p.warte('der Ordner steht im Bestand', async () => {
    const zeile = gesetzt(p, ort)
    return zeile.includes(vorhanden) ? zeile : false
  }, 20000)
  p.pruefe('der Ordner steht im Bestand', true, danach.replaceAll('\t', ' '))
  p.pruefe('und die Meldung ist fort', !(await p.seite(`return !!${feld}.closest('label').parentElement.querySelector('[role=alert]')`)))
  // **Seit UXR-F2 (#2008) vier Segmente**: der Name steht an der Gruppe
  // (`aria-labelledby`), nicht an einem Beschriftungsfeld. Er muss dastehen
  // und darf nicht der rohe Schlüssel sein.
  p.pruefe('der Namenskonflikt daneben trägt seinen Namen', await p.seite("const g = document.querySelector('[data-einstellung=\"bildkonflikt\"]'); const n = (g?.labels?.[0]?.innerText ?? document.getElementById(g?.getAttribute('aria-labelledby') ?? '')?.innerText ?? '').trim(); return n !== '' && !n.startsWith('bildkonflikt')"))
  await p.bild('ordner-gesetzt')

  p.pruefe('⌘Q beendet', await p.beende())
}

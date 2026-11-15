// **Das Protokoll nach seinen Einstellungen, am nativen Fenster** — OPS-04
// (#1721):
//
// - AK-01: Ort, Stufe und Form des Protokolls sind eingestellt — über `mse`
//   vor dem Start, denn sie wirken ab dem nächsten —, und die Datei liegt
//   danach dort, wo die Einstellung sagt,
// - AK-02/03: die Diagnose nennt die Kanäle und sagt, dass das ausführliche
//   Protokoll an ist,
// - die Einstellungen stehen in der Gruppe *Diagnose* mit ihren Namen und
//   nicht als rohe Schlüssel,
// - AK-04: im Diagnosepaket lässt sich ein Teil abwählen, und im Paket auf
//   der Platte fehlt er dann.

import { execFileSync } from 'node:child_process'
import { existsSync, mkdirSync } from 'node:fs'
import { join } from 'node:path'

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

const kasten = (name) => `[...document.querySelectorAll('input[type=checkbox]')].find((k) => k.getAttribute('aria-label') === ${JSON.stringify(`${name} ins Paket nehmen`)})`

export default async function ops04(p) {
  const ort = p.ort()
  const protokolle = join(ort.ort, 'Protokolle')
  mkdirSync(protokolle)
  p.mse(ort, 'einstellungen', 'setzen', 'protokoll.ort', protokolle)
  p.mse(ort, 'einstellungen', 'setzen', 'protokoll.stufe', 'ausfuehrlich')
  p.mse(ort, 'einstellungen', 'setzen', 'protokoll.format', 'json')
  await p.starte({ ort })

  p.notiz('AK-01 — die Datei liegt am eingestellten Ort')
  p.pruefe('die Protokolldatei liegt im eingestellten Ordner', existsSync(join(protokolle, 'mse.log')), protokolle)

  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)
  await p.klickeText('Diagnose')
  // **Seit UXR-19 (#1927) in einem Abschnitt** — erst aufklappen.
  await p.klickeText('Protokoll')
  await p.wartetext('Wie viel das Protokoll aufschreibt', 20000)
  // **Seit UXR-F2 (#2008) vier Segmente** statt einer Auswahl: die Stufe hat
  // vier Werte (R7). Gewählt ist das Optionsfeld, dessen Segment den Namen trägt.
  p.pruefe('die Stufe steht mit ihrem Namen da', await p.seite("const r = document.querySelector('[role=radiogroup][data-einstellung=\"protokoll.stufe\"] input:checked'); return !!r && r.closest('label').innerText.trim().startsWith('Ausführlich')"))
  // **Der Name eines Segments steht an der Gruppe** (`aria-labelledby`), der eines Schalters oder einer Auswahl am Beschriftungsfeld.
  p.pruefe('kein roher Schlüssel steht als Beschriftung', await p.seite("return ![...document.querySelectorAll('[data-einstellung]')].some((e) => (e.labels?.[0]?.innerText ?? document.getElementById(e.getAttribute('aria-labelledby') ?? '')?.innerText ?? '').trim().startsWith('protokoll.'))"))
  await p.bild('einstellungen')

  p.notiz('AK-02/03 — die Diagnose nennt Kanäle und Stufe')
  await p.klickeText('Diagnose anzeigen')
  const diagnose = await p.warte('die Diagnose nennt die Kanäle', () => p.seite("const t = document.body.innerText; const z = t.split('\\n').find((l) => l.startsWith('Kanäle:')); return z ?? false"), 20000)
  p.pruefe('Datei und Unified Log an, Standardfehlerausgabe aus', diagnose.includes('Datei an') && diagnose.includes('Unified Log an') && diagnose.includes('Standardfehlerausgabe aus'), diagnose)
  p.pruefe('Stufe und Form stehen dabei', diagnose.includes('Stufe ausfuehrlich, json'), diagnose)
  p.pruefe('und dass das ausführliche Protokoll an ist', await p.seite("return document.body.innerText.includes('Das ausführliche Protokoll ist an')"))
  p.pruefe('der Ort des Protokolls ist der eingestellte', await p.seite(`return document.body.innerText.includes(${JSON.stringify('Protokolle/mse.log')})`))
  await p.bild('diagnose')

  p.notiz('AK-04 — ein Teil lässt sich vor dem Sichern abwählen')
  await p.klickeText('Zusammenstellen')
  await p.warte('die Vorschau zeigt die Teile', () => p.seite(`return !!${kasten('Protokollzeilen')}`), 20000)
  await p.klickeAuf(kasten('Protokollzeilen'), 'das Häkchen der Protokollzeilen')
  await p.warte('das Häkchen ist aus', () => p.seite(`return ${kasten('Protokollzeilen')}.checked === false`))
  await p.bild('abgewaehlt')

  const paket = join(ort.ort, 'bericht.msediagnose')
  await p.dialogantwort(paket)
  // **Das Zeichen vor dem Text** (⤓) lässt `klickeText` vorbeigreifen —
  // der Knopf wird am Ende seines Textes erkannt.
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith('Paket sichern'))", '»Paket sichern«')
  await p.warte('das Paket liegt auf der Platte', () => existsSync(paket), 20000)
  const inhalt = execFileSync('unzip', ['-Z1', paket], { encoding: 'utf8' }).split('\n').filter(Boolean)
  p.pruefe('im Paket fehlen die Protokollzeilen', !inhalt.includes('protokoll.txt'), inhalt.join(', '))
  p.pruefe('die übrigen Teile sind drin', inhalt.includes('fassung.txt') && inhalt.includes('metriken.json'), inhalt.join(', '))

  p.pruefe('⌘Q beendet', await p.beende())
}

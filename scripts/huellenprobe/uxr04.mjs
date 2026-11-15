// **Die Setliste nach dem Artboard *Sets*, am nativen Fenster** — UXR-04
// (#1912),
// AK-02 bis AK-04:
//
// - im Kopf steht *Neues Set* neben *Set öffnen …*, beide im Nebengewicht
//   — seit AP-3 (#2360) führt im Kopf keiner (`PROTOTYP-0.9.0.md`),
// - bei 1280 × 800 steht die zweite Spalte **neben** der Liste, und nichts
//   rollt waagerecht (R5); **das Anlegen steht seit AP-10 (#2367) mit einem
//   Set in einer Schublade**, daneben steht die Einfuhr,
// - eine Setkarte trägt ihren Namen und ein Menü — die fünf seltenen
//   Handlungen stehen darin, das Löschen zuletzt, und seine Frage nimmt den
//   Fokus,
// - *Neues Set* springt zum Namen; der Stil ist eine Auswahl mit Suche.
//
// **Was sie nicht belegt**: ob jemand die Handlungen hinter „…" findet — das
// sieht nur ein Mensch (UXR-04.AK-04, Sichtprüfung).
const knopf = (name) => `[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.trim().endsWith(${JSON.stringify(name)}))`
const region = (name) => `[...document.querySelectorAll('main section, [role=dialog] section')].find((s) => s.checkVisibility() && (s.getAttribute('aria-label') === ${JSON.stringify(name)} || document.getElementById(s.getAttribute('aria-labelledby') ?? '')?.innerText.trim() === ${JSON.stringify(name)}))`
const menueknopf = (setname) => `document.querySelector('main button[aria-label=${JSON.stringify(`Handlungen — ${setname}`)}]')`

// lage gibt die Kanten der Liste und des Neuen.
const lage = `const l = (${region('Sets')})?.getBoundingClientRect(); const n = [...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Ein Set von woanders einlesen')?.getBoundingClientRect(); return l && n ? { liste: { links: l.left, rechts: l.right, unten: l.bottom }, neu: { links: n.left, oben: n.top } } : null`

async function starte(p, ort, breite, hoehe) {
  p.mse(ort, 'einstellungen', 'setzen', 'fensterstart', 'zuletzt')
  p.mse(ort, 'einstellungen', 'setzen', 'fensterbreite', String(breite))
  p.mse(ort, 'einstellungen', 'setzen', 'fensterhoehe', String(hoehe))
  await p.starte({ ort })
  await p.taste('2', 'befehl')
  await p.warte('die Sets stehen', () => p.seite("return !!document.querySelector('main article')"), 20000)
}

export default async function uxr04(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  const setname = 'Werkstatt-Testkarten'

  p.notiz('1280 × 800 — der Kopf und die zwei Spalten')
  await starte(p, ort, 1280, 800)
  const kopf = await p.seite("return [...document.querySelectorAll('main header button')].filter((b) => b.checkVisibility()).map((b) => ({ text: b.innerText.trim(), haupt: /(^|\\s)bg-marke(\\s|$)/.test(b.className) }))")
  p.zeige(`Kopf: ${JSON.stringify(kopf)}`)
  // **Das Zeichen steht vor dem Text** („+ Neues Set"), deshalb das Ende.
  const haupt = kopf.filter((k) => k.haupt)
  p.pruefe('im Kopf führt keiner, »Neues Set« steht darin (AP-3)', haupt.length === 0 && kopf.some((k) => k.text.endsWith('Neues Set')))
  p.pruefe('»Set öffnen …« steht daneben, im Nebengewicht', kopf.some((k) => k.text === 'Set öffnen …' && !k.haupt))
  const breit = await p.seite(lage)
  p.zeige(`Lage: ${JSON.stringify(breit)}`)
  p.pruefe('die zweite Spalte steht neben der Liste', !!breit && breit.neu.links >= breit.liste.rechts)
  p.pruefe('nichts rollt waagerecht', await p.seite('return document.scrollingElement.scrollWidth <= innerWidth + 1'))
  await p.bild('setliste-breit')

  p.notiz('Die Setkarte — der Name und ein Menü')
  const kartenknoepfe = await p.seite("return [...document.querySelectorAll('main article button')].filter((b) => b.checkVisibility()).map((b) => b.getAttribute('aria-label') ?? b.innerText.trim())")
  p.zeige(`Knöpfe der Karte: ${JSON.stringify(kartenknoepfe)}`)
  p.pruefe('die Karte trägt zwei Knöpfe: den Namen und das Menü', kartenknoepfe.length === 2)
  await p.klickeAuf(menueknopf(setname), 'das Menü der Setkarte')
  const eintraege = await p.warte('das Menü steht', () => p.seite("const m = [...document.querySelectorAll('[role=menu]')].find((e) => e.checkVisibility()); return m ? [...m.querySelectorAll('[role=menuitem]')].map((e) => e.innerText.replace(/\\s+/g, ' ').trim()) : false"))
  p.zeige(`Menü: ${JSON.stringify(eintraege)}`)
  p.pruefe('fünf Einträge, das Löschen zuletzt', eintraege.length === 5 && eintraege[4].endsWith('Löschen …'))
  p.pruefe('das Menü bleibt im Fenster', await p.seite("const m = [...document.querySelectorAll('[role=menu]')].find((e) => e.checkVisibility()); const r = m.getBoundingClientRect(); return r.left >= 0 && r.right <= innerWidth"))
  await p.bild('setkarte-menue')

  p.notiz('Löschen — die Frage nimmt den Fokus, Abbrechen nimmt sie weg')
  await p.klickeAuf("[...document.querySelectorAll('[role=menuitem]')].find((e) => e.checkVisibility() && e.innerText.trim().endsWith('Löschen …'))", 'Löschen …')
  await p.warte('die Frage steht, der Fokus im Feld', () => p.seite("return (document.activeElement?.labels?.[0]?.innerText ?? '').includes('Namen des Sets eintippen')"))
  p.pruefe('die Frage nimmt den Fokus', true)
  await p.bild('setkarte-loeschen')
  await p.klickeAuf("[...document.querySelectorAll('main article button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Abbrechen')", 'Abbrechen')
  await p.warte('die Frage ist weg', () => p.seite("return ![...document.querySelectorAll('main input')].some((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').includes('Namen des Sets eintippen'))"))
  p.pruefe('nichts gelöscht', JSON.parse(p.mse(ort, 'sets', '--json')).sets.length === 1)

  p.notiz('Neues Set — zum Namen, und der Stil als Auswahl')
  await p.klickeAuf(knopf('Neues Set'), 'Neues Set')
  await p.warte('der Fokus steht im Namen', () => p.seite("return (document.activeElement?.labels?.[0]?.innerText ?? '').includes('Name des Sets')"))
  p.pruefe('»Neues Set« springt zum Namen, in der Schublade', await p.seite("return !!document.activeElement?.closest('[role=dialog]')"))
  // **»Zurück« aus dem leeren Namen heraus** — die Meldung, die beim
  // Verlassen kommt, schob die Knöpfe hinab, und der erste Klick landete
  // daneben (#2024). Seitdem steht ihr Platz vorher frei; dieser Schritt ist
  // der Beleg am Fenster.
  await p.klickeAuf(`[...((${region('Ein neues Set')})?.querySelectorAll('button') ?? [])].find((b) => b.checkVisibility() && b.innerText.trim() === 'Zurück')`, 'Zurück')
  // **Seit UXR-F1 (#2007) ein Kombinationsfeld mit Suche** — vorher ein
  // natives `select`. Die Zeilen stehen erst, wenn es offen ist; `uxrf1`
  // fährt Öffnen, Tippen und Wählen.
  const stil = await p.warte('der Stil steht als Auswahl', () => p.seite(`const n = ${region('Ein neues Set')}; const s = [...(n?.querySelectorAll('input[role=combobox]') ?? [])].find((e) => (e.labels?.[0]?.innerText ?? '').includes('Stil des neuen Sets')); return s ? { rolle: s.getAttribute('role'), wert: s.value, optionsfelder: n.querySelectorAll('input[type=radio]').length } : false`))
  p.zeige(`Stil: ${JSON.stringify(stil)}`)
  p.pruefe('kein Optionsfeld je Stil', stil.optionsfelder === 0)
  p.pruefe('der Stil ist ein Kombinationsfeld mit gewähltem Namen', stil.rolle === 'combobox' && stil.wert !== '')
  await p.bild('neues-set-stil')
  p.pruefe('⌘Q beendet (breit)', await p.beende())

  // **Das Fenster wird nicht schmaler als 1024 Punkte** — `uirevue` maß
  // „eng" deshalb bei 1024 × 700, und dort greift `lg`. Die beiden stehen
  // nebeneinander; geprüft wird, dass die Liste dabei lesbar breit bleibt.
  p.notiz('das schmalste Fenster — nebeneinander, die Liste lesbar breit')
  await starte(p, ort, 1000, 700)
  const eng = await p.seite(lage)
  p.zeige(`Lage: ${JSON.stringify(eng)}`)
  p.pruefe('die zweite Spalte steht neben der Liste, die Liste mindestens 18rem breit', !!eng && eng.neu.links >= eng.liste.rechts && eng.liste.rechts - eng.liste.links >= 288)
  p.pruefe('nichts rollt waagerecht', await p.seite('return document.scrollingElement.scrollWidth <= innerWidth + 1'))
  await p.bild('setliste-eng')
  p.pruefe('⌘Q beendet (eng)', await p.beende())
}

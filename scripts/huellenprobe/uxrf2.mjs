// **Schalter und Segmente in den Einstellungen, am nativen Fenster** — UXR-F2
// (#2008),
// Regel R7 aus `docs/UIREVIEW.md`:
//
// 1. eine Einstellung mit zwei Werten, die an und aus sind, ist ein Schalter,
//    eine mit drei oder vier Werten eine Reihe von Segmenten, eine mit mehr
//    eine Auswahlliste,
// 2. **die Leertaste legt den Schalter um** und der Fokus bleibt auf ihm —
//    während des Schreibens ist er `disabled`, und WebKit nimmt einem
//    fokussierten Element den Fokus (#1990),
// 3. **die Pfeiltasten wählen in den Segmenten**, der Fokus bleibt in der
//    Reihe, und ⇥ verlässt sie mit **einem** Druck,
// 4. der Fokusring ist zu sehen — am Schalter und am Segment,
// 5. `Wie beim Nutzer` ist ein Segment unter anderen und erreichbar,
// 6. nichts rollt waagerecht.
//
// Gelesen wird der Bestand über `mse einstellungen`, nicht am Fenster: was
// das Fenster zeigt, ist die Frage, was es schreibt, die andere.

const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

// art sagt, was unter einem Schlüssel gezeichnet ist.
const art = (p, schluessel) => p.seite(`const e = document.querySelector('[data-einstellung="${schluessel}"]'); if (!e || !e.checkVisibility()) return 'fehlt'; return e.getAttribute('role') ?? e.tagName.toLowerCase()`)

// fokusAn fragt, ob der Fokus in dem Element mit diesem Schlüssel steht.
const fokusAn = (p, schluessel) => p.seite(`const e = document.querySelector('[data-einstellung="${schluessel}"]'); return !!e && (e === document.activeElement || e.contains(document.activeElement))`)

// gewaehlt ist der Wert des gewählten Segments.
const gewaehlt = (p, schluessel) => p.seite(`return document.querySelector('[data-einstellung="${schluessel}"] input:checked')?.value ?? ''`)

// freigegeben wartet, bis nichts mehr geschrieben wird — die Felder sind es
// nicht mehr `disabled`.
const freigegeben = (p, schluessel) => p.warte(`${schluessel} ist geschrieben`, () => p.seite(`const e = document.querySelector('[data-einstellung="${schluessel}"]'); if (!e) return false; const felder = e.matches('input') ? [e] : [...e.querySelectorAll('input')]; return felder.length > 0 && felder.every((f) => !f.disabled)`), 20000)

// ring sagt, ob das fokussierte Element einen Fokusring zeichnet. Beim Segment
// liegt er um das Beschriftungsfeld (has-focus-visible), nicht um das
// Optionsfeld, das man nicht sieht.
const ring = (p) => p.seite("const a = document.activeElement; const e = a?.type === 'radio' ? a.closest('label') : a; if (!e) return false; const s = getComputedStyle(e); return s.outlineStyle !== 'none' && parseFloat(s.outlineWidth) >= 2")

export default async function uxrf2(p) {
  const ort = p.ort()
  await p.starte({ ort })

  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)
  await p.wartetext('Nach neuen Fassungen sehen', 20000)

  p.notiz('1 — das Bedienelement nach der Zahl der Werte')
  p.pruefe('Einführung beim Start (zwei Werte, an und aus) ist ein Schalter', (await art(p, 'ersteschritte')) === 'switch', await art(p, 'ersteschritte'))
  p.pruefe('Nach neuen Fassungen sehen (drei Werte) ist eine Reihe von Segmenten', (await art(p, 'aktualisierung')) === 'radiogroup', await art(p, 'aktualisierung'))
  p.pruefe('Sprache (neun Werte) bleibt eine Auswahlliste', (await art(p, 'sprache')) === 'select', await art(p, 'sprache'))
  p.pruefe('die Segmente sind drei Optionsfelder einer Gruppe', await p.seite("const r = [...document.querySelectorAll('[data-einstellung=\"aktualisierung\"] input[type=radio]')]; return r.length === 3 && new Set(r.map((x) => x.name)).size === 1"))
  await p.bild('allgemein')

  p.notiz('2 — die Leertaste legt den Schalter um, und der Fokus bleibt')
  await p.seite("document.querySelector('input[role=switch][data-einstellung=\"ersteschritte\"]').focus(); return true")
  await p.taste(' ')
  await p.warte('der Schalter steht auf aus', () => p.seite("return document.querySelector('input[role=switch][data-einstellung=\"ersteschritte\"]').checked === false"), 20000)
  await freigegeben(p, 'ersteschritte')
  p.pruefe('im Bestand steht „aus“', /ersteschritte\s+aus/.test(p.mse(ort, 'einstellungen')))
  p.pruefe('der Fokus steht noch auf dem Schalter', await fokusAn(p, 'ersteschritte'), await p.seite("return document.activeElement?.tagName ?? ''"))
  p.pruefe('der Fokusring ist zu sehen', await ring(p))
  await p.taste(' ')
  await p.warte('der Schalter steht wieder auf an', () => p.seite("return document.querySelector('input[role=switch][data-einstellung=\"ersteschritte\"]').checked === true"), 20000)
  await freigegeben(p, 'ersteschritte')
  p.pruefe('im Bestand steht nicht mehr „aus“', !/ersteschritte\s+aus/.test(p.mse(ort, 'einstellungen')))

  p.notiz('3 — die Pfeile wählen in den Segmenten, und ⇥ verlässt die Reihe mit einem Druck')
  const vorher = await gewaehlt(p, 'aktualisierung')
  await p.seite("document.querySelector('[data-einstellung=\"aktualisierung\"] input:checked').focus(); return true")
  await p.taste('ArrowRight')
  await p.warte('„Nur melden“ ist gewählt', async () => (await gewaehlt(p, 'aktualisierung')) === 'melden', 20000)
  await freigegeben(p, 'aktualisierung')
  p.pruefe('im Bestand steht „melden“', /aktualisierung\s+melden/.test(p.mse(ort, 'einstellungen')))
  p.pruefe('der Fokus steht noch in der Reihe', await fokusAn(p, 'aktualisierung'))
  await p.taste('ArrowRight')
  await p.warte('„Melden und herunterladen“ ist gewählt', async () => (await gewaehlt(p, 'aktualisierung')) === 'holen', 20000)
  await freigegeben(p, 'aktualisierung')
  p.pruefe('der zweite Pfeil kam an: im Bestand steht „holen“', /aktualisierung\s+holen/.test(p.mse(ort, 'einstellungen')))
  // **Der Ring nach dem zweiten Pfeil**: das Drehbuch setzt den ersten Fokus
  // per Skript, und WebKit zeigt `:focus-visible` erst, wenn eine Taste den
  // Fokus bewegt hat — ein Mensch kommt mit ⇥ in die Reihe. Gemessen am
  // 2026-10-03: nach dem ersten Pfeil ohne Ring, nach dem zweiten mit.
  p.zeige(`Fokus nach dem zweiten Pfeil: ${await p.seite("const a = document.activeElement; const l = a?.closest('label'); return JSON.stringify({ art: a?.type, fokus: a?.matches(':focus'), sichtbar: a?.matches(':focus-visible'), labelHat: l?.matches(':has(:focus-visible)'), labelHatFokus: l?.matches(':has(:focus)'), rahmen: l ? getComputedStyle(l).outlineStyle + ' ' + getComputedStyle(l).outlineWidth : null, klassen: l?.className.slice(0, 160) })")}`)
  p.pruefe('der Fokusring ist zu sehen', await ring(p))
  await p.bild('segmente')
  // **Jeder Pfeil einzeln abgewartet**: ein Pfeil, der ankommt, während die Reihe
  // noch schreibt, ginge ins Leere — und das wäre die Probe, nicht das Fenster.
  await p.taste('ArrowLeft')
  await p.warte('„Nur melden“ ist wieder gewählt', async () => (await gewaehlt(p, 'aktualisierung')) === 'melden', 20000)
  await freigegeben(p, 'aktualisierung')
  await p.taste('ArrowLeft')
  await p.warte(`der Anfangswert „${vorher}“ ist wieder gewählt`, async () => (await gewaehlt(p, 'aktualisierung')) === vorher, 20000)
  await freigegeben(p, 'aktualisierung')
  // **Auf die Wirkung gewartet, mit einem Druck** — im Gesamtlauf der Abnahme
  // von 0.7.0 war diese Zeile einmal rot: sie fragte sofort nach der Taste,
  // bevor die Webansicht sie verarbeitet hatte. Bild für Bild gemessen blieb
  // der Fokus während jedes Schreibens auf dem gewählten Segment, nie auf
  // `body` und nie auf einem abgeschalteten Feld — es war die Probe.
  await p.taste('Tab')
  p.pruefe('⇥ verlässt die Reihe mit einem Druck', Boolean(await p.warte('⇥ hat die Reihe verlassen', async () => !(await fokusAn(p, 'aktualisierung')) && (await p.seite('return document.activeElement !== document.body')), 5000)))

  p.notiz('4 — „Wie beim Nutzer“ ist ein Segment unter anderen')
  await p.klickeText('Erscheinungsbild')
  await p.warte('das Thema steht da', async () => (await art(p, 'erscheinungsbild')) !== 'fehlt', 20000)
  p.pruefe('das Thema (vier Werte mit dem Erbwert) ist eine Reihe von Segmenten', (await art(p, 'erscheinungsbild')) === 'radiogroup', await art(p, 'erscheinungsbild'))
  const namen = await p.seite("return [...document.querySelectorAll('[data-einstellung=\"erscheinungsbild\"] label')].map((l) => l.innerText.trim()).join('|')")
  p.pruefe('sie führt „Wie beim Nutzer“, „Wie das System“, „Hell“ und „Dunkel“', namen === 'Wie beim Nutzer|Wie das System|Hell|Dunkel', namen)
  p.pruefe('Farbwelt, Fensterbild und Anzeigeschrift gilt sind ebenfalls Segmente', (await art(p, 'farbwelt')) === 'radiogroup' && (await art(p, 'hintergrund')) === 'radiogroup' && (await art(p, 'schriftbereich')) === 'radiogroup')
  p.pruefe('die Schriftgröße (fünf Werte) bleibt eine Auswahlliste', (await art(p, 'schriftgroesse')) === 'select', await art(p, 'schriftgroesse'))
  await p.seite("document.querySelector('[data-einstellung=\"erscheinungsbild\"] input[value=hell]').click(); return true")
  await p.warte('„hell“ steht im Bestand', async () => /erscheinungsbild\s+hell/.test(p.mse(ort, 'einstellungen')), 20000)
  await freigegeben(p, 'erscheinungsbild')
  await p.seite("document.querySelector('[data-einstellung=\"erscheinungsbild\"] input[value=erben]').click(); return true")
  await p.warte('„Wie beim Nutzer“ ist wieder gewählt und „hell“ nicht mehr im Bestand', async () => (await gewaehlt(p, 'erscheinungsbild')) === 'erben' && !/erscheinungsbild\s+hell/.test(p.mse(ort, 'einstellungen')), 20000)
  await p.bild('erscheinungsbild')

  p.notiz('5 — nichts rollt waagerecht, und jede Reihe liegt im Fenster')
  p.pruefe('die Seite rollt nicht waagerecht', await p.seite('return document.scrollingElement.scrollWidth <= innerWidth + 1'))
  p.pruefe('jede Reihe endet vor dem Fensterrand', await p.seite("return [...document.querySelectorAll('[role=radiogroup]')].filter((r) => r.checkVisibility()).every((r) => r.getBoundingClientRect().right <= innerWidth)"))

  p.pruefe('⌘Q beendet', await p.beende())
}

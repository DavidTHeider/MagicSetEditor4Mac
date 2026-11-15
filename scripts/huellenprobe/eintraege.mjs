// **Jeder übrige Eintrag der Leiste tut, was er sagt** — ENG-02.AK-04,
// Kernweg 2, und Kernweg 3 mit dem Ordner für MTG JSON
// (#1727).
//
// Neu, Öffnen, Sichern und Sichern unter stehen in `fix-f02`, das
// Anwendungs-, Bearbeiten- und Fenstermenü samt Sprachwechsel in `leiste`.
// Hier der Rest, und jeder mit seiner **Wirkung**, nicht mit dem Ruf:
//
// - die Einträge, die auf eine Seite führen — Ausgeben (drei), Drucken,
//   Suchen, Ersetzen, Hilfe —: die Brotkrume nennt danach diese Seite; und
//   »Über«, das auf den Abschnitt mit Lizenz und Rechten Dritter rollt,
// - Zuletzt benutzt: schlägt das Set auf,
// - Alle Karten anhaken: jede Karte der Liste ist angehakt,
// - Rückgängig und Wiederherstellen: der Bestand verliert eine Änderung und
//   bekommt sie wieder — gelesen über `mse`, nicht über das Fenster,
// - Seitenleiste, Zurück, Vorwärts,
// - ein Fensterbild wählen (UX-09.AK-01) — der Dialog, der Fokus nach dem
//   Abbruch (UX-08.AK-03), die Vorschau, das Bild hinter der Oberfläche —,
// - Darstellung zurücksetzen: Farbwelt und Fensterbild fallen auf die
//   Vorgabe (UX-09.AK-03),
// - Bestand sichern: der Dateidialog, dann eine Sicherung am gewählten Ort,
// - Schließen (⌘W): das Set ist zu, die Liste der Sets steht da,
// - Kernweg 3: „Verzeichnis öffnen …" für MTG JSON — der Ordnerdialog, dann
//   die Sets aus dem Verzeichnis im Bestand.
import { execFileSync } from 'node:child_process'
import { existsSync, statSync } from 'node:fs'
import { homedir } from 'node:os'
import { join, resolve } from 'node:path'

const wurzel = resolve(import.meta.dirname, '..', '..')

// hier ist die Seite, auf der die Brotkrume steht.
const hier = (p) => p.seite("return document.querySelector('nav ol [aria-current=page]')?.innerText.trim() ?? ''")

// fuehrtAuf löst einen Eintrag aus und wartet, bis die Brotkrume ihn nennt.
async function fuehrtAuf(p, pfad, erwartet) {
  await p.klick(...pfad)
  const seite = await p.warte(`»${pfad.at(-1)}« führt auf ${erwartet}`, async () => {
    const jetzt = await hier(p)
    return jetzt === erwartet ? jetzt : false
  }).catch(async (grund) => {
    p.zeige(`die Brotkrume nennt »${await hier(p)}«`)
    throw grund
  })
  p.pruefe(`»${pfad.join(' › ')}« führt auf »${seite}«`, true)
}

const kartennamen = (p, ort) => p.mse(ort, 'karten', '1').trim().split('\n')

const angehakt = (p) => p.seite("return [...document.querySelectorAll('main input[type=checkbox]')].filter((k) => k.checkVisibility() && k.checked).length")

const seitenleiste = (p) => p.seite("const k = [...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /Seitenleiste (ein|aus)blenden/.test(b.getAttribute('aria-label') ?? b.innerText)); return k ? (k.getAttribute('aria-label') ?? k.innerText).trim() : ''")

export default async function eintraege(p) {
  const ort = p.ort()
  p.mse(ort, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(ort, 'einstellungen', 'setzen', 'farbwelt', 'schiefer')
  p.mse(ort, 'einstellungen', 'setzen', 'hintergrund', 'fuellen')
  await p.starte({ ort })

  p.notiz('Ein Fensterbild wählen — UX-09.AK-01, und der Fokus nach dem Abbruch (UX-08.AK-03)')
  const bild = join(ort.ort, 'fensterbild.png')
  execFileSync('cp', [join(wurzel, 'build/appicon.png'), bild])
  await p.taste(',', 'befehl')
  await p.warte('die Einstellungen stehen da', async () => (await hier(p)) !== '', 20000)
  const bildknopf = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && b.innerText.trim() === 'Bild wählen …')"
  if (!(await p.seite(`return !!${bildknopf}`))) await p.klickeText('Erscheinungsbild')
  await p.warte('»Bild wählen …« steht da', () => p.seite(`return !!${bildknopf}`))
  // **Mit der Tastatur ausgelöst** (UX-08.AK-03): ein Klick fokussiert in
  // WebKit keinen Knopf, und die Frage ist, wohin der Fokus nach dem Abbruch
  // zurückkehrt — für den, der mit der Tastatur kam.
  // **Erst ein Klick in die Seite**, damit sie Tasten bekommt wie bei einem
  // Menschen, der im Fenster arbeitet — ein Klick fokussiert den Knopf in
  // WebKit nicht, er macht nur die Webansicht zum Ersten Antwortenden.
  await p.klickeAuf(bildknopf, '»Bild wählen …«')
  await p.dialogZeigtSich('oeffnen', '»Bild wählen …« (Klick)')
  await p.seite(`${bildknopf}.focus(); return true`)
  await p.taste('Enter')
  await p.dialogZeigtSich('oeffnen', '»Bild wählen …« (⏎)')
  // **Wohin der Fokus nach dem Abbruch geht, belegt die Probe nicht**: ihr
  // Fenster wird nie Schlüsselfenster (die Anwendung startet aus einem
  // Hintergrundprozess und kommt seit macOS 14 nicht nach vorn), und damit
  // gibt AppKit der Webansicht den Fokus nicht so zurück wie einem Menschen,
  // der vor dem Fenster sitzt. Was sie sieht, steht im Beleg; die Prüfung
  // bleibt eine Handprüfung (`docs/HANDPRUEFUNGEN.md`, UX-08.AK-03).
  p.zeige(`der Fokus nach dem Abbruch, ohne Schlüsselfenster: »${(await p.seite("return document.activeElement?.innerText?.trim() ?? ''")).slice(0, 40)}«`)
  await p.dialogantwort(bild)
  await p.klickeAuf(bildknopf, '»Bild wählen …«')
  await p.warte('die Vorschau zeigt das Bild', () => p.seite("return document.querySelector('[data-vorschau=fensterbild]') !== null"), 20000)
  const hinter = await p.warte('das Bild steht hinter der Oberfläche', () => p.seite("const w = document.documentElement.style.getPropertyValue('--mse-fensterbild'); return w !== '' ? w : false"), 20000)
  p.pruefe('das gewählte Bild steht in der Vorschau und hinter der Oberfläche', true, hinter.slice(0, 60))
  await p.bild('fensterbild')

  p.notiz('Darstellung zurücksetzen — Farbwelt und Fensterbild fallen auf die Vorgabe')
  const vorher = await p.seite("return document.documentElement.getAttribute('data-farbwelt') ?? 'pergament'")
  p.pruefe('vorher steht die Farbwelt aus dem Profil', vorher === 'schiefer', vorher)
  await p.klick('Darstellung', 'Darstellung zurücksetzen')
  const nachher = await p.warte('die Farbwelt fällt zurück', () => p.seite("return document.documentElement.getAttribute('data-farbwelt') === null ? 'pergament' : false"))
  p.pruefe('danach steht die Vorgabe', nachher === 'pergament', nachher)
  await p.warte('das Fensterbild ist aus', () => p.seite("return document.documentElement.style.getPropertyValue('--mse-fensterbild') === ''"))
  p.pruefe('und hinter der Oberfläche steht kein Bild mehr', true)
  p.pruefe('und im Profil steht keine Farbwelt mehr', !/farbwelt\s+schiefer/.test(p.mse(ort, 'einstellungen')))

  p.notiz('Ein Set aufschlagen — die Kartenliste steht da')
  await p.taste('2', 'befehl')
  await p.klickeText('Werkstatt-Testkarten')
  await p.warte('die Kartenliste steht', async () => (await p.seite("return document.querySelectorAll('main input[type=checkbox]').length")) >= 5, 20000)
  p.zeige(`die Brotkrume nach dem Aufschlagen: »${await hier(p)}«`)

  p.notiz('Rückgängig und Wiederherstellen — am Bestand gelesen')
  const alt = kartennamen(p, ort)
  await p.klickeAuf("[...document.querySelectorAll('main button')].find((b) => b.checkVisibility() && b.innerText.startsWith('Langtext'))", 'die Karte »Langtext«')
  // **Ein Klick macht die Webansicht zum Ersten Antwortenden, der Fokus
  // kommt über die Seite** — hier ist die Frage Rückgängig und
  // Wiederherstellen, nicht der Klick; getippt wird nativ.
  await p.klickeAuf("[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && (e.labels?.[0]?.querySelector(':scope > span')?.innerText ?? e.getAttribute('aria-label') ?? '').trim() === 'Name')", 'das Namensfeld')
  // **Die Aufschrift und nicht das ganze Label**: seit AI-05 (#2048) steht
  // im Label auch der leise Knopf »Vorschlagen«, und bis zum 2026-10-05
  // verglich dieses Drehbuch den ganzen Text mit »Name«.
  //
  // **Erst, wenn die Karte steht** — das Formular wird nach dem Aufschlagen
  // noch einmal gezeichnet, und ein Fokus davor ginge mit dem alten Feld.
  const namensfeld = "[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && (e.labels?.[0]?.querySelector(':scope > span')?.innerText ?? e.getAttribute('aria-label') ?? '').trim() === 'Name')"
  await p.warte('das Namensfeld trägt »Langtext«', () => p.seite(`return (${namensfeld})?.value === 'Langtext'`))
  await new Promise((los) => setTimeout(los, 800))
  await p.warte('das Namensfeld hat den Fokus', () => p.seite(`const f = ${namensfeld}; if (!f) return false; f.focus(); f.setSelectionRange(f.value.length, f.value.length); return document.activeElement === f && f.value === 'Langtext'`))
  await p.tippeInsFeld(' Neu')
  // **Geschrieben wird beim Verlassen des Felds** — ⇥, wie in
  // `ux10-tastatur`.
  await p.taste('Tab')
  const neu = await p.warte('der neue Name steht im Bestand', () => {
    const jetzt = kartennamen(p, ort)
    return jetzt.some((z) => z.includes('Langtext Neu')) ? jetzt : false
  }, 20000)
  await p.klick('Set', 'Rückgängig')
  await p.warte('»Rückgängig« nimmt ihn zurück', () => !kartennamen(p, ort).some((z) => z.includes('Langtext Neu')), 20000)
  p.pruefe('»Rückgängig« nimmt die Änderung im Bestand zurück', true, kartennamen(p, ort).find((z) => z.includes('Langtext')) ?? '')
  // **Bis zum alten Stand**: ein Eingriff je Tastendruck wäre eine Folge,
  // und ihr Ende ist der Name von vorher.
  for (let schritt = 0; schritt < 5 && kartennamen(p, ort).join('\n') !== alt.join('\n'); schritt += 1) {
    await p.klick('Set', 'Rückgängig')
    await new Promise((los) => setTimeout(los, 700))
  }
  p.zeige(`nach dem Rücknehmen: ${kartennamen(p, ort).find((z) => z.includes('Langtext')) ?? ''}`)
  await p.klick('Set', 'Wiederherstellen')
  await p.warte('»Wiederherstellen« bringt eine Änderung wieder', () => kartennamen(p, ort).join('\n') !== alt.join('\n'), 20000)
  p.pruefe('»Wiederherstellen« bringt sie wieder', true, kartennamen(p, ort).find((z) => z.includes('Langtext')) ?? '')
  p.zeige(`der Name vor dem Tippen war »${alt.find((z) => z.includes('Langtext')) ?? ''}«, danach »${neu.find((z) => z.includes('Langtext')) ?? ''}«`)

  p.notiz('Alle Karten anhaken')
  const vorAnhaken = await angehakt(p)
  await p.klick('Set', 'Alle Karten anhaken')
  const nachAnhaken = await p.warte('jede Karte ist angehakt', async () => {
    const zahl = await angehakt(p)
    return zahl >= vorAnhaken + 5 ? zahl : false
  })
  p.pruefe('»Alle Karten anhaken« hakt die fünf Karten an', nachAnhaken - vorAnhaken === 5, `${String(vorAnhaken)} → ${String(nachAnhaken)}`)

  p.notiz('Die Einträge, die auf eine Seite führen')
  const ziele = [
    [['Ablage', 'Ausgeben', 'Über eine Exportvorlage …'], 'Ausgeben'],
    [['Ablage', 'Ausgeben', 'Als Bilder …'], 'Bilder'],
    [['Ablage', 'Ausgeben', 'Druckbogen …'], 'Druckbogen'],
    [['Ablage', 'Drucken …'], 'Druckbogen'],
    [['Set', 'Suchen …'], 'Suche'],
    [['Set', 'Ersetzen …'], 'Ersetzen'],
    [['Hilfe', 'MagicSetEditor4Mac-Hilfe'], 'Hilfe'],
  ]
  for (const [pfad, erwartet] of ziele) await fuehrtAuf(p, pfad, erwartet)

  p.notiz('»Über« führt auf den Abschnitt mit Lizenz und Rechten Dritter')
  await p.klick('MagicSetEditor4Mac', 'Über MagicSetEditor4Mac')
  const ueber = await p.warte('der Abschnitt »Über« steht im Blick', () => p.seite("const h = [...document.querySelectorAll('h2,h3')].find((e) => e.innerText.trim() === 'Über MagicSetEditor4Mac'); if (!h) return false; const r = h.getBoundingClientRect(); const abschnitt = h.closest('section') ?? h.parentElement; return r.top >= 0 && r.top < innerHeight ? abschnitt.innerText : false"))
  p.pruefe('er nennt die Lizenz und die Inhalte Dritter', ueber.includes('GNU General Public License') && ueber.includes('Inhalte Dritter'), ueber.split('\n').slice(0, 3).join(' · '))
  // **Zurück auf die Hilfe als Ausgangspunkt** für den Gang der Seiten.
  await p.klick('Set', 'Ersetzen …')
  await p.warte('wieder bei Ersetzen', async () => (await hier(p)) === 'Ersetzen')
  await p.klick('Hilfe', 'MagicSetEditor4Mac-Hilfe')
  await p.warte('wieder in der Hilfe', async () => (await hier(p)) === 'Hilfe')

  p.notiz('Zurück und Vorwärts gehen den Gang der Seiten')
  await p.klick('Darstellung', 'Zurück')
  await p.warte('»Zurück« führt auf Ersetzen', async () => (await hier(p)) === 'Ersetzen')
  p.pruefe('»Zurück« führt auf die Seite davor', true, await hier(p))
  await p.klick('Darstellung', 'Vorwärts')
  await p.warte('»Vorwärts« führt auf die Hilfe', async () => (await hier(p)) === 'Hilfe')
  p.pruefe('»Vorwärts« führt wieder hin', true, await hier(p))

  p.notiz('Die Seitenleiste ein- und ausblenden')
  const offen = await seitenleiste(p)
  // **Der Eintrag heißt nach dem Stand** (UXR-F14, #2029): steht die Leiste,
  // „Seitenleiste ausblenden", sonst „Seitenleiste einblenden" — dieselben
  // Wörter wie am Knopf, den `seitenleiste` liest. Das Menü selbst misst
  // `darstellung`.
  await p.klick('Darstellung', offen)
  const zu = await p.warte('die Seitenleiste klappt', async () => {
    const jetzt = await seitenleiste(p)
    return jetzt !== offen ? jetzt : false
  })
  p.pruefe('der Eintrag klappt die Seitenleiste', true, `${offen} → ${zu}`)
  await p.taste('s', 'wahl', 'befehl')
  await p.warte('⌥⌘S klappt sie zurück', async () => (await seitenleiste(p)) === offen)
  p.pruefe('⌥⌘S klappt sie zurück', true)

  p.notiz('Zuletzt benutzt schlägt das Set auf')
  await p.klick('Ablage', 'Schließen')
  await p.warte('⌘W schließt das Set', async () => (await hier(p)) === 'Sets')
  p.pruefe('»Schließen« schließt das Set, die Liste der Sets steht da', true, await hier(p))
  const leiste = await p.menue()
  const zuletzt = leiste.find((m) => m.titel === 'Ablage')?.unter?.find((m) => m.titel === 'Zuletzt benutzt')
  p.pruefe('»Zuletzt benutzt« nennt das Set', (zuletzt?.unter ?? []).some((m) => m.titel.includes('Werkstatt')), (zuletzt?.unter ?? []).map((m) => m.titel).join(' · '))
  const eintrag = (zuletzt?.unter ?? []).find((m) => m.titel.includes('Werkstatt'))?.titel ?? 'Werkstatt-Testkarten'
  await p.klick('Ablage', 'Zuletzt benutzt', eintrag)
  await p.warte('das Set ist wieder offen', async () => !['Sets', ''].includes(await hier(p)))
  p.pruefe('»Zuletzt benutzt« schlägt es auf', true, await hier(p))

  p.notiz('Bestand sichern — der Dateidialog, dann die Sicherung')
  const sicherung = join(ort.ort, 'bestand-sicherung.mse-bestand')
  await p.klick('Ablage', 'Bestand sichern …')
  const blatt = await p.dialogZeigtSich('sichern', '»Bestand sichern …«')
  p.zeige(`vorgeschlagen: »${blatt.name ?? ''}«`)
  await p.dialogantwort(sicherung)
  await p.klick('Ablage', 'Bestand sichern …')
  await p.warte('die Sicherung liegt da', () => existsSync(sicherung) && statSync(sicherung).size > 0, 30000).catch(async (grund) => {
    p.zeige(`Meldungen: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('[role=alert],[role=status]')].map((e) => e.innerText.trim()).filter(Boolean).slice(0, 6)"))}`)
    throw grund
  })
  p.pruefe('»Bestand sichern …« schreibt an den gewählten Ort', true, `${String(statSync(sicherung).size)} Bytes`)

  p.notiz('Kernweg 3 — ein Verzeichnis für MTG JSON wählen')
  const setsVorher = p.mse(ort, 'sets').trim().split('\n').length
  await p.taste('2', 'befehl')
  await p.klickeText('Ein Set von woanders einlesen')
  await p.wartetext('Verzeichnis öffnen')
  await p.klickeText('Verzeichnis öffnen …')
  const ordner = await p.dialogZeigtSich('oeffnen', '»Verzeichnis öffnen …«')
  // **Der Ort ohne das Heimatverzeichnis** — er ist der von macOS, und der
  // Beleg gehört nicht einem Rechner.
  p.pruefe('der Dialog wählt Ordner', ordner.ordnerwahl === true, JSON.stringify(ordner).replaceAll(homedir(), '~'))
  await p.dialogantwort(join(wurzel, 'testdaten/mtgjson/einzeln'))
  await p.klickeText('Verzeichnis öffnen …')
  const setsNachher = await p.warte('die Sets aus dem Verzeichnis sind da', () => {
    const zahl = p.mse(ort, 'sets').trim().split('\n').length
    return zahl > setsVorher ? zahl : false
  }, 30000)
  p.pruefe('das Verzeichnis kommt als Sets in den Bestand', true, `${String(setsVorher)} → ${String(setsNachher)} Zeilen`)
  p.pruefe('⌘Q beendet', await p.beende())
}

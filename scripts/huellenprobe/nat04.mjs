// **Zwei Fenster, ein Tisch** — S12.NAT-04 aus dem Abschluss von Phase 6.11
// (#1672).
//
// Zwei Anwendungen, je mit eigenem Bestand, eigenem Bund und eigener
// Identität, auf diesem Mac — gestartet wie zwei Programme über
// LaunchServices (`open -n`). Dazwischen liegt alles, was der Betrieb hat:
// Horchposten, TCP, TLS 1.3 festgenagelt auf die Kennung, Handschlag, Wirt
// und Gast.
//
// - **Verdeckte Informationen**: jedes Fenster sieht die eigene Hand offen
//   und die fremde verdeckt.
// - **Der Wirt sieht, was der Gast tut** — ohne selbst etwas zu tun. Bis zum
//   Zweitblick über Phase 6.11 kam das bei Go an und nicht im Fenster (Z1).
// - **Ein Satz geht vom Gast zum Wirt.**
//
// **Was sie nicht belegt**: zwei Geräte in zwei Netzen (#1998) — beide Enden
// laufen auf diesem Mac über die Schleife. Und den Handel zwischen zwei
// Fenstern; ihn tragen zwei Prozesse mit getrennten Beständen
// (`internal/flaechen/handelsdraht_test.go`).
//
// **Eingerichtet wird über die Bindungen** (`window.go.app.Bindungen`, die
// Gestalt aus `anbindung.ts`): Identität, Kontakt, Zweck und Horchposten
// sind ein eigener Weg mit eigenen Proben (`net01`, `id02`). **Beobachtet
// wird über die Oberfläche** — der Beitritt, die Hände, der Zug, der Satz.

import { rufbuchHolen } from '../huellenrufbuch.mjs'

const bindung = (p, ausdruck) => p.seite(`const b = window.go.app.Bindungen; return await ${ausdruck}`)

const knoepfe = (p) => p.seite("return [...document.querySelectorAll('[role=application] li button')].map((b) => b.getAttribute('aria-label') || b.textContent)")

// haende zählt die Handkarten: die eigenen tragen ihren Namen, die fremden
// stehen verdeckt da (dieselbe Lesart wie `fix-f04`).
async function haende(p) {
  const hand = (await knoepfe(p)).filter((t) => t.endsWith('· Hand'))
  const verdeckt = hand.filter((t) => t.startsWith('verdeckt')).length

  return { offen: hand.length - verdeckt, verdeckt }
}

// meldungen nennt, was als Fehler dasteht — ein Fenster, das still klagt,
// ist ein Befund, auch wenn der Weg gelingt.
const meldungen = (p) => p.seite("return [...document.querySelectorAll('[role=alert]')].filter((e) => e.checkVisibility()).map((e) => e.innerText.trim()).filter(Boolean)")

const protokoll = (p) => p.seite("return [...document.querySelectorAll('[role=log] li')].map((l) => l.innerText.replace(/\\s+/g, ' ').trim())")

// koppeln macht zwei Anwendungen einander bekannt — bestätigt, mit Zweck.
async function koppeln(hier, dort, name) {
  await bindung(hier, `b.PeerkontaktAufnehmen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(name)})`)
  const bestaetigt = await bindung(hier, `b.PeerkontaktBestaetigen(${JSON.stringify(dort.kennung)}, ${JSON.stringify(dort.fingerabdruck)})`)
  hier.pruefe(`${name} ist bestätigt`, !bestaetigt?.fehler, JSON.stringify(bestaetigt?.fehler ?? ''))
  const erlaubt = await bindung(hier, `b.PeerzweckErlauben(${JSON.stringify(dort.kennung)}, 'partie')`)
  hier.pruefe(`${name} darf eine Partie spielen`, !erlaubt?.fehler, JSON.stringify(erlaubt?.fehler ?? ''))
}

export default async function nat04(p) {
  // **Dieselbe Klasse, eine zweite Probe** — ein Import von
  // `huellenprobe.mjs` hinge an dessen eigenem `await main()`. **Und
  // dieselbe Zeilenliste**: der Beleg steht in der Reihenfolge, in der es
  // geschah.
  const gast = new p.constructor('nat04-gast')
  gast.zeilen = p.zeilen

  try {
    await zweiFenster(p, gast)
  } finally {
    p.fehler += gast.fehler
    gast.aufraeumen()
  }
}

async function zweiFenster(p, gast) {
  const wirtOrt = p.ort()
  p.mse(wirtOrt, 'oeffnen', 'saat/probe-werkstatt.mse-set')
  p.mse(wirtOrt, 'deck', 'neu', 'Probe-Deck', 'probe')
  for (const karte of ['1', '2', '3', '4', '5']) p.mse(wirtOrt, 'deck', 'legen', '1', `--karte=${karte}`, '8')
  const angelegt = p.mse(wirtOrt, 'partie-anlegen', 'probe', 'ada:Ada:1', 'bo:Bo:1')
  const partie = /Partie ([0-9a-f]+)/.exec(angelegt)?.[1] ?? ''
  p.pruefe('die Partie ist angelegt', partie !== '', angelegt.split('\n')[0])

  // **Der Beginn geht vorweg** (#1851): Starthand und Behalten für beide
  // Sitze, über die Kommandozeile — der Beginn ist nicht, was hier geprüft
  // wird.
  for (const sitz of ['ada', 'bo']) {
    for (const art of ['spieler.starthand', 'spieler.behalten']) {
      p.mse(wirtOrt, 'partie-befehl', partie, sitz, `${art}/${sitz}`, art)
    }
  }

  p.notiz('Zwei Anwendungen, zwei Bestände, einander bekannt')
  await p.starte({ ort: wirtOrt })
  await gast.starte({ ort: gast.ort() })

  const wirt = await bindung(p, 'b.PeeridentitaetSicherstellen()')
  const bert = await bindung(gast, 'b.PeeridentitaetSicherstellen()')
  p.pruefe('zwei verschiedene Identitäten', wirt.kennung && bert.kennung && wirt.kennung !== bert.kennung, `…${wirt.kennung?.slice(-8)} · …${bert.kennung?.slice(-8)}`)

  await koppeln(p, bert, 'Gast')
  await koppeln(gast, wirt, 'Wirt')

  const horch = await bindung(p, "b.HorchStarten('127.0.0.1:0')")
  const port = horch?.stand?.adresse?.split(':').pop()
  p.pruefe('der Wirt horcht', Boolean(horch?.stand?.laeuft && port), horch?.stand?.adresse ?? JSON.stringify(horch?.fehler))
  await bindung(gast, `b.PeerkontaktAdresse(${JSON.stringify(wirt.kennung)}, '127.0.0.1:${port}')`)

  p.notiz('Der Wirt öffnet seinen Tisch und gibt ihn frei')
  await p.taste('4', 'befehl')
  await p.wartetext('Partie anfangen', 20000)
  await p.klickeText('Partie anfangen')
  await p.wartetext('Weiterspielen')
  await p.klickeText('Weiterspielen')
  await p.klickeText('Weiterspielen als Ada')
  await p.warte('die Leinwand des Wirts zeichnet', () => p.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 20000)

  // **Über die Oberfläche**, wie ein Mensch: eine Freigabe über die Bindung
  // sähe der Tisch erst nach »Nachsehen«, und ohne sie stünde kein Gespräch.
  await p.klickeText('Tisch öffnen')
  await p.wartetext('Dieser Tisch steht offen.', 10000)
  p.pruefe('der Tisch steht für Bo offen', true)
  p.zeige(`Meldungen beim Wirt: ${JSON.stringify(await meldungen(p))}`)

  p.notiz('Der Gast setzt sich — über die Schublade, wie ein Mensch')
  // **Das handelnde Fenster nach vorn**: Tasten und Klicke gehen an das
  // Fenster, das den Fokus hat — zwei Programme teilen sich einen Bildschirm.
  await gast.aktiviere()
  await gast.taste('4', 'befehl')
  await gast.wartetext('Partie anfangen', 20000)
  await gast.klickeText('Partie anfangen')
  await gast.wartetext('Bei jemandem mitspielen', 20000)
  await gast.klickeText('Bei jemandem mitspielen')
  await gast.klickeAuf("[...document.querySelectorAll('label, [role=radio]')].find((e) => e.checkVisibility() && /^Wirt · /.test((e.innerText ?? '').trim()))", 'den Wirt')
  await gast.klickeText('Mitspielen')
  await gast.wartetext('Du sitzt am Tisch von', 20000)
  await gast.warte('die Leinwand des Gastes zeichnet', () => gast.seite("return !!document.querySelector('canvas')?.checkVisibility()"), 20000)
  await gast.bild('nat04-gast-sitzt')
  // **Und ohne Klage**: bis S12.NAT-04 zeigte der fremde Tisch die Führung
  // des Wirts, und Rückschau und Protokoll fragten den eigenen Bestand nach
  // einer Partie, die dort nicht liegt — „Der gesuchte Eintrag wurde nicht
  // gefunden".
  await gast.warte('der Tisch des Gastes steht', () => gast.seite("return !!document.querySelector('[role=application]')"), 20000)
  gast.pruefe('am fremden Tisch steht keine Fehlermeldung', (await meldungen(gast)).length === 0, JSON.stringify(await meldungen(gast)))
  // **Und der Wirt sieht, dass jemand sitzt** — ohne »Nachsehen«. Bis dahin
  // sagte seine Freigabe „Noch sitzt niemand", während Bo schon zog.
  await p.warte('der Wirt sieht Bo an seinem Tisch', () => p.seite("return document.body.innerText.includes('Es sitzen:')"), 10000)
  p.pruefe('der Wirt sieht, dass Bo sitzt, ohne nachzusehen', true)
  gast.pruefe('und nichts, was dem Wirt gehört', !(await gast.seite("return [...document.querySelectorAll('button')].some((b) => b.checkVisibility() && ['Anhalten', 'Tisch öffnen', 'Verlauf ansehen'].includes(b.innerText.trim())) || [...document.querySelectorAll('label')].some((l) => l.checkVisibility() && l.innerText.trim().startsWith('Sicht für'))")))
  // **Und die Bilder kommen vom Wirt** — P9 (#2377): bis dahin stand hier
  // ein Satz, warum sie fehlen. Der Gast hat die Werkstatt gar nicht.
  // Gleiche Karten teilen ein Bild — wie viele es sind, sagt die Sicht des
  // Wirts für Bo.
  const sicht = await bindung(p, `b.Partie(${JSON.stringify(partie)}, 'bo')`)
  const soll = new Set((sicht?.partie?.objekte ?? []).filter((o) => o.eigner === 'bo' && o.zone === 'hand' && o.bild).map((o) => o.bild)).size
  const bilder = async () => (await rufbuchHolen(gast)).Fremdkartenbild?.ausgaenge?.gelungen ?? 0
  await gast.warte('der Gast hat seine Handkarten beim Wirt erfragt', async () => (await bilder()) >= soll, 20000).catch(() => false)
  gast.pruefe('der Gast zeigt die Kartenbilder des Wirts', soll > 0 && (await bilder()) === soll, `${String(await bilder())} von ${String(soll)}`)

  p.notiz('Verdeckte Informationen — jedes Fenster sieht nur die eigene Hand')
  const beimWirt = await haende(p)
  const beimGast = await haende(gast)
  p.pruefe('der Wirt sieht seine sieben Karten offen und die des Gastes verdeckt', beimWirt.offen === 7 && beimWirt.verdeckt === 7, JSON.stringify(beimWirt))
  gast.pruefe('der Gast sieht seine sieben Karten offen und die des Wirts verdeckt', beimGast.offen === 7 && beimGast.verdeckt === 7, JSON.stringify(beimGast))
  await gast.seite("document.querySelector('[role=application]')?.scrollIntoView({ block: 'start', behavior: 'instant' })")
  await gast.bild('nat04-gast-am-tisch')

  p.notiz('Der Gast wirft eine Münze — und der Wirt sieht es, ohne etwas zu tun')
  const vorher = await protokoll(p)
  await gast.klickeText('Münze werfen')
  await gast.klickeAuf("[...document.querySelectorAll('[role=group], fieldset')].find((g) => g.checkVisibility() && (g.getAttribute('aria-label') === 'Münze werfen' || g.querySelector('legend')?.innerText.trim() === 'Münze werfen'))?.querySelector('button[type=submit], button:last-of-type')", '»Stellen« am Münzwurf')
  await gast.warte('der Gast sieht seinen Münzwurf', async () => (await protokoll(gast)).some((z) => /Münze/.test(z)), 10000)
  // **Der Wirt bleibt hinten und rührt nichts an** — gefragt wird nur, was
  // sein Fenster zeigt.
  await p.warte('der Wirt sieht den Münzwurf des Gastes', async () => {
    const jetzt = await protokoll(p)

    return jetzt.length > vorher.length && jetzt.some((z) => /Münze/.test(z))
  }, 10000)
  p.zeige(`Protokoll beim Wirt: ${JSON.stringify(await protokoll(p))}`)
  p.pruefe('der Wirt sieht den Münzwurf, ohne etwas getan zu haben', (await protokoll(p)).some((z) => /Münze/.test(z)))
  gast.pruefe('auch nach dem Wurf keine Fehlermeldung beim Gast', (await meldungen(gast)).length === 0, JSON.stringify(await meldungen(gast)))
  await p.aktiviere()
  await p.bild('nat04-wirt-sieht-den-zug')

  p.notiz('Ein Satz geht vom Gast zum Wirt')
  await gast.aktiviere()
  // **Der Fokus per Skript**: ein Klick fokussiert in WebKit nicht
  // verlässlich (`fokus.ts`), und getippt wird in das fokussierte Feld.
  const feld = "[...document.querySelectorAll('input, textarea')].find((e) => e.checkVisibility() && (e.getAttribute('aria-label') === 'Etwas sagen' || e.labels?.[0]?.innerText.trim() === 'Etwas sagen'))"
  await gast.klickeAuf(feld, 'das Feld »Etwas sagen«')
  await gast.seite(`(${feld})?.focus()`)
  await gast.tippeInsFeld('Hallo vom Gast')
  await gast.klickeText('Sagen')
  await p.wartetext('Hallo vom Gast', 10000)
  p.pruefe('der Wirt liest den Satz des Gastes', true)

  // **Der Beitritt gehört seinem eigenen Stop** — BK-01.AK-03 (#2109): seit
  // jede Bindung den Kontext ihres Rufs nimmt, endet der mit der
  // Beitrittsantwort; die Partie lief trotzdem bis hierher weiter.
  p.notiz('Der Gast steht auf — der Beitritt endet beim zuständigen Stop und nicht vorher')
  gast.pruefe('nach Münzwurf und Satz läuft der Beitritt noch', (await bindung(gast, 'b.PartieBeitrittStand()')).laeuft === true)
  const aufgestanden = await bindung(gast, 'b.PartieBeitrittBeenden()')
  gast.pruefe('»Beenden« antwortet ohne Fehler', !aufgestanden.fehler, JSON.stringify(aufgestanden.fehler ?? ''))
  await gast.warte('der Beitritt läuft nicht mehr', async () => (await bindung(gast, 'b.PartieBeitrittStand()')).laeuft === false, 10000)
}

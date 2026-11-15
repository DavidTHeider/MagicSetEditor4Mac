// **Profilwechsel mit zweitem Faktor** — ENG-02.AK-04, Kernweg 9
// (#1727).
//
// Ein zweiter Mensch, Bob, mit einem TOTP-Faktor — eingerichtet über `mse`
// im Prüfbestand und in dessen **eigenem Prüfbund**, nicht im Anmeldebund
// dieses Rechners. Der Code entsteht hier nach RFC 6238 aus dem Geheimnis,
// das `faktor-neu` genau einmal ausgibt: ein Testwert dieses Laufs.
//
// - Wechseln ohne Code wird nicht angenommen, sondern fragt nach ihm,
// - ein falscher Code wird abgewiesen, und gearbeitet wird weiter im alten
//   Profil,
// - der richtige Code wechselt.
//
// **Wohin er wechselt**: in Bobs Profil — **als der Mensch, der vorher
// arbeitete**, mit der Rolle, die er dort hat. „Person A im Profil von B
// bleibt A" (`frontend/src/rahmen/Kontextleiste.tsx`); ein frisch
// angelegter Bob hat niemandem mehr als das Lesen gegeben, und so steht es
// dann in der Kontextleiste: „Mein Profil von Bob · nur lesen".
import { createHmac } from 'node:crypto'

// code rechnet ein Einmalkennwort nach RFC 6238 — mit den Angaben aus der
// Adresse, die `faktor-neu` ausgibt, und nicht mit angenommenen.
function code(adresse, zeit) {
  const angaben = new URL(adresse).searchParams
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567'
  const bits = [...(angaben.get('secret') ?? '').replace(/=+$/, '').toUpperCase()]
    .map((z) => alphabet.indexOf(z).toString(2).padStart(5, '0')).join('')
  const schluessel = Buffer.from((bits.match(/.{8}/g) ?? []).map((b) => parseInt(b, 2)))
  const schritt = Number(angaben.get('period') ?? '30') * 1000
  const stellen = Number(angaben.get('digits') ?? '6')
  const zaehler = Buffer.alloc(8)
  zaehler.writeBigUInt64BE(BigInt(Math.floor(zeit / schritt)))
  const hmac = createHmac((angaben.get('algorithm') ?? 'SHA1').toLowerCase(), schluessel).update(zaehler).digest()
  const stelle = hmac[hmac.length - 1] & 0xf

  return String((hmac.readUInt32BE(stelle) & 0x7fffffff) % 10 ** stellen).padStart(stellen, '0')
}

const fenster = (adresse) => Number(new URL(adresse).searchParams.get('period') ?? '30') * 1000

const wechselKnopf = "[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && (b.getAttribute('aria-label') ?? '').startsWith('Zu Bob'))"

// hierBei sagt, in wessen Zeile „Hier arbeitest du gerade" steht — seit
// UXR-20 der Name des Profils; der Mensch steht in der zweiten Zeile.
const hierBei = (p) => p.seite("const li = [...document.querySelectorAll('li')].find((e) => e.checkVisibility() && e.innerText.includes('Hier arbeitest du gerade')); return li ? li.innerText.split('\\n')[0].trim() : ''")

const codefeld = "[...document.querySelectorAll('input')].find((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim() === 'Code')"

// codeEingeben leert das Feld und tippt — **über das Feld selbst**, nicht
// über den Fokus: nach einer Abweisung steht er auf der Meldung.
async function codeEingeben(p, wert) {
  await p.klickeAuf(codefeld, 'das Feld »Code«')
  await p.seite(`const f = ${codefeld}; f.focus(); Object.getOwnPropertyDescriptor(HTMLInputElement.prototype, 'value').set.call(f, ''); f.dispatchEvent(new Event('input', { bubbles: true })); return true`)
  await p.tippeInsFeld(wert)
  await p.klickeText('Bestätigen')
}

export default async function profilwechsel(p) {
  const ort = p.ort()
  p.mse(ort, 'nutzer-neu', 'Bob')
  const nutzer = p.mse(ort, 'nutzer')
  const bob = nutzer.split('\n').find((z) => z.includes('Bob'))?.trim().split(/\s+/)[0] ?? ''
  p.zeige(`Bob ist Nutzer ${bob}`)
  const [adresse] = p.mse(ort, 'faktor-neu', bob, 'totp').trim().split('\n')
  const eingerichtet = Date.now()
  p.mse(ort, 'faktor-bestaetigen', bob, code(adresse, eingerichtet))
  p.pruefe('Bobs zweiter Faktor steht', /totp/i.test(p.mse(ort, 'faktoren', bob)))
  await p.starte({ ort })

  p.notiz('Wechseln ohne Code fragt nach ihm')
  // **Über die Leiste** — die Profilwahl steht seit UXR-20 (#1928) neben den
  // Laschen und nicht mehr in *Profile*; eine Lasche braucht es nicht.
  await p.klickeAuf("[...document.querySelectorAll('button')].find((b) => b.checkVisibility() && /Nutzer und Profile/.test(b.getAttribute('aria-label') ?? b.innerText))", '»Nutzer und Profile«')
  await p.wartetext('Hierhin wechseln', 20000)
  const vorher = await hierBei(p)
  p.zeige(`gearbeitet wird bei »${vorher}«`)
  await p.klickeAuf(wechselKnopf, 'Zu Bob wechseln')
  await p.warte('das Feld »Code« steht da', () => p.seite("return [...document.querySelectorAll('input')].some((e) => e.checkVisibility() && (e.labels?.[0]?.innerText ?? '').trim() === 'Code')"))
  p.pruefe('der Wechsel verlangt den zweiten Faktor und bleibt, wo er war', (await hierBei(p)) === vorher, await hierBei(p))

  p.notiz('Ein falscher Code wird abgewiesen')
  const richtig = code(adresse, Date.now())
  await codeEingeben(p, richtig === '000000' ? '111111' : '000000')
  const abweisung = await p.warte('eine Meldung weist ihn ab', () => p.seite("return [...document.querySelectorAll('[role=alert]')].map((e) => e.innerText.trim()).find(Boolean) ?? false"))
  p.pruefe('der falsche Code wird abgewiesen', true, abweisung.slice(0, 160))
  p.pruefe('und gearbeitet wird weiter im alten Profil', (await hierBei(p)) === vorher)

  p.notiz('Der richtige Code wechselt')
  // **Ein neues Zeitfenster**: der Code vom Einrichten ist verbraucht, und
  // derselbe noch einmal wäre ein Wiederholungsversuch.
  const schritt = fenster(adresse)
  while (Math.floor(Date.now() / schritt) <= Math.floor(eingerichtet / schritt)) {
    await new Promise((los) => setTimeout(los, 1000))
  }
  await codeEingeben(p, code(adresse, Date.now()))
  // **Nach dem Wechsel steht die Anwendung wieder am Anfang** (§7, Regel 1;
  // `profilwechsel.test.tsx`) — gelesen wird deshalb die Kontextleiste und
  // nicht die Nutzerseite.
  const nachher = await p.warte('gearbeitet wird in Bobs Profil', async () => {
    const jetzt = await p.seite("return [...document.querySelectorAll('[aria-label]')].find((e) => e.getAttribute('aria-label') === 'Arbeitskontext')?.innerText.replace(/\\s+/g, ' ').trim() ?? ''")
    return jetzt.includes('von Bob') ? jetzt : false
  }, 20000).catch(async (grund) => {
    p.zeige(`Meldungen: ${JSON.stringify(await p.seite("return [...document.querySelectorAll('[role=alert],[role=status]')].map((e) => e.innerText.trim()).filter(Boolean).slice(0, 6)"))}`)
    throw grund
  })
  p.pruefe('der richtige Code wechselt in Bobs Profil', true, `${vorher} → ${nachher}`)
  p.pruefe('als derselbe Mensch, mit der Rolle aus Bobs Profil', nachher.includes('Ich') && nachher.includes('nur lesen'))
  p.zeige(`mse profil: ${p.mse(ort, 'profil').trim().split('\n')[0]}`)
  p.pruefe('⌘Q beendet', await p.beende())
}

// **Hat jeder Ruf des Vertragsrundgangs so viele Argumente wie seine
// Bindung?** — #2528.
//
// # Der Anlass
//
// Zwei Rufe folgten ihren neuen Argumenten nicht (`HandelSchritt` seit
// #2496, `PartieBeitreten` seit #2377). Wails wies sie am Tor ab, und der
// Rundgang (`huellenprobe/vertrag.mjs`) zählte das als fachliche Absage —
// gefunden erst in der Gesamtabnahme (QA01, #2382), als jemand die Absagen
// las. Eine neue Bindung zwingt dazu, sie im Rundgang einzutragen; ein
// geänderter Aufruf zwang zu nichts.
//
// # Wie
//
// Jedes `argumente(k)` läuft gegen einen **Stellvertreter** für den Kontext:
// jede Eigenschaft, jeder Aufruf und jede Umwandlung in Text gibt wieder ihn.
// Gezählt wird, was zurückkommt — die Länge der Liste, nicht ihr Inhalt.
//
// # Was er nicht hält
//
// **Die Typen und Werte der Argumente** — dafür ruft der Rundgang am Fenster.
// Und einen Ruf, dessen Zahl vom Kontext abhängt: der Stellvertreter ist in
// jeder Bedingung wahr, also zählt der erste Zweig.

// stellvertreter gibt einen Kontext, der jede Rechnung trägt.
export function stellvertreter() {
  const knoten = new Proxy(function () {}, {
    get: (_, schluessel) => {
      if (schluessel === Symbol.toPrimitive || schluessel === 'toJSON') return () => 'stellvertreter'
      if (schluessel === Symbol.iterator) return function* () {}
      // **Kein `then`** — sonst hielte `await` ihn für ein Versprechen.
      if (schluessel === 'then') return undefined
      if (schluessel === 'length') return 0

      return knoten
    },
    apply: () => knoten,
  })

  return knoten
}

// abweichungen nennt jeden Ruf, dessen Argumentzahl nicht zur Bindung passt,
// und jeden auf eine Bindung, die es nicht gibt.
export async function abweichungen(rundgang, signaturen) {
  const zahl = new Map(signaturen.map((methode) => [methode.name, methode.parameter.length]))
  const funde = []

  for (const gruppe of rundgang) {
    for (const eintrag of gruppe.rufe) {
      if (!zahl.has(eintrag.name)) {
        funde.push(`${eintrag.name}: keine Bindung dieses Namens`)
        continue
      }

      let gezaehlt
      try {
        gezaehlt = (eintrag.argumente ? await eintrag.argumente(stellvertreter()) : []).length
      } catch (fehler) {
        // **Was beim Auswerten wirklich rechnet**, zählt der Quelltext.
        gezaehlt = stellenImQuelltext(eintrag.argumente)
        if (gezaehlt === undefined) {
          funde.push(`${eintrag.name}: argumente lassen sich nicht auswerten (${fehler.message})`)
          continue
        }
      }

      const erwartet = zahl.get(eintrag.name)
      if (gezaehlt !== erwartet) {
        funde.push(`${eintrag.name}: ${String(gezaehlt)} Argumente, die Bindung hat ${String(erwartet)}`)
      }
    }
  }

  return funde
}

// stellenImQuelltext zählt die Stellen der Liste, die eine Pfeilfunktion
// unmittelbar zurückgibt — `undefined`, wenn ihr Rumpf keine Liste ist.
// Text, Vorlagen und alles Geschachtelte zählen als **eine** Stelle.
export function stellenImQuelltext(funktion) {
  const text = funktion.toString()
  const pfeil = text.indexOf('=>')
  if (pfeil < 0) return undefined

  let stelle = pfeil + 2
  while (/\s/.test(text[stelle])) stelle++
  if (text[stelle] !== '[') return undefined

  let tiefe = 0
  let stellen = 0
  let etwas = false
  for (; stelle < text.length; stelle++) {
    const zeichen = text[stelle]
    if (zeichen === '"' || zeichen === "'" || zeichen === '`') {
      stelle = textende(text, stelle)
      etwas = true
    } else if ('([{'.includes(zeichen)) {
      tiefe++
      if (tiefe > 1) etwas = true
    } else if (')]}'.includes(zeichen)) {
      tiefe--
      if (tiefe === 0) return stellen + (etwas ? 1 : 0)
    } else if (zeichen === ',' && tiefe === 1) {
      if (etwas) stellen++
      etwas = false
    } else if (!/\s/.test(zeichen)) {
      etwas = true
    }
  }

  return undefined
}

// textende gibt die Stelle des Zeichens, das einen Text oder eine Vorlage
// schließt — in einer Vorlage über jedes `${…}` hinweg.
function textende(text, anfang) {
  const zeichen = text[anfang]
  for (let stelle = anfang + 1; stelle < text.length; stelle++) {
    if (text[stelle] === '\\') {
      stelle++
    } else if (text[stelle] === zeichen) {
      return stelle
    } else if (zeichen === '`' && text[stelle] === '$' && text[stelle + 1] === '{') {
      let tiefe = 0
      for (stelle += 1; stelle < text.length; stelle++) {
        if (text[stelle] === '"' || text[stelle] === "'" || text[stelle] === '`') stelle = textende(text, stelle)
        else if (text[stelle] === '{') tiefe++
        else if (text[stelle] === '}' && --tiefe === 0) break
      }
    }
  }

  return text.length
}

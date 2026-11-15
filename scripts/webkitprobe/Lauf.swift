// **Ein Lauf fährt ein Drehbuch und hinterlässt seinen Beleg** — Schritt für
// Schritt mit Zeile, Ergebnis und Bild, und am Ende einen Rückgabewert
// (AGENTS.md §3: ein Test belegt sich mit seinem Lauf).
//
// **Beim ersten roten Schritt hält er an**: was danach käme, stünde auf einem
// Stand, den niemand bestellt hat. Ein Bild vom Augenblick des Fehlers kommt
// trotzdem mit — es sagt meist mehr als die Meldung.
import AppKit

struct Schrittfehler: Error, CustomStringConvertible {
  let description: String
}

@MainActor final class Lauf {
  let argumente: Argumente
  let motor: Motor
  let fenster: Probefenster
  private var frist = 10_000
  private var bericht: [String] = []

  init(argumente: Argumente, motor: Motor, fenster: Probefenster) {
    self.argumente = argumente
    self.motor = motor
    self.fenster = fenster
  }

  func fahren(_ schritte: [(zeile: Int, schritt: Schritt, roh: String)]) async -> Bool {
    bericht += kopf()
    var bestanden = true

    for (zeile, schritt, roh) in schritte {
      do {
        let auskunft = try await ausfuehren(schritt)
        bericht.append("- ✓ Z\(zeile) `\(roh)`")
        if !auskunft.isEmpty { bericht.append(contentsOf: block(auskunft)) }
      } catch {
        bericht.append("- ✗ Z\(zeile) `\(roh)` — **\(error)**")
        if let seite = try? await motor.auswerten("window.__probeFehler || []"), seite != "[]" {
          bericht.append("  Fehler der Seite:")
          bericht.append(contentsOf: block(seite))
        }
        if let bild = try? await motor.abzug() {
          try? bild.write(to: ausgabe("fehler.png"))
          bericht.append("  Bild im Augenblick des Fehlers: `fehler.png`")
        }
        bestanden = false
        break
      }
    }

    bericht.append("")
    bericht.append(bestanden ? "**Ergebnis: bestanden** — \(schritte.count) Schritte."
                             : "**Ergebnis: rot.**")
    try? (bericht.joined(separator: "\n") + "\n").write(to: ausgabe("bericht.md"), atomically: true, encoding: .utf8)
    print(bericht.joined(separator: "\n"))
    return bestanden
  }

  private func ausgabe(_ datei: String) -> URL {
    URL(fileURLWithPath: argumente.ausgabe).appendingPathComponent(datei)
  }

  private func kopf() -> [String] {
    let system = ProcessInfo.processInfo.operatingSystemVersionString
    let webkit = Bundle(identifier: "com.apple.WebKit")?.infoDictionary?["CFBundleVersion"] as? String ?? "?"
    return [
      "# WebKit-Probe — \(URL(fileURLWithPath: argumente.drehbuch).lastPathComponent)", "",
      "| | |", "| --- | --- |",
      "| Stand | `\(argumente.stand)` |",
      "| Motor | \(motor.name == "wk2" ? "wk2 — WKWebView, eingerichtet wie in Wails 3" : "wk1 — WebCore im eigenen Prozess") |",
      "| Basis | \(argumente.basis) |",
      "| Fenster | \(Int(argumente.breite)) × \(Int(argumente.hoehe)) Punkte, \(argumente.dunkel ? "dunkel" : "hell") |",
      "| macOS | \(system) |", "| WebKit | \(webkit) |",
      "| Beginn | \(ISO8601DateFormatter().string(from: Date())) |", "",
    ]
  }

  private func block(_ text: String) -> [String] {
    ["  ```"] + text.components(separatedBy: "\n").map { "  " + $0 } + ["  ```"]
  }

  // MARK: - Schritte

  private func ausfuehren(_ schritt: Schritt) async throws -> String {
    switch schritt {
    case .motor, .notiz:
      return ""
    case .lade(let pfad):
      guard let adresse = URL(string: pfad, relativeTo: URL(string: argumente.basis)) else {
        throw Schrittfehler(description: "»\(pfad)« ist keine Adresse")
      }
      try await motor.lade(adresse.absoluteURL)
      return ""
    case .warte(let css):
      try await warten("sichtbar \(css)", bis: sichtbar(css))
      return ""
    case .warteText(let gesucht):
      try await warten("Text »\(gesucht)«", bis: "document.body.innerText.includes(\(json(gesucht)))")
      return ""
    case .bis(let ausdruck):
      try await warten("»\(ausdruck)«", bis: "Boolean(\(ausdruck))")
      return ""
    case .klicke(let css):
      try await klicken(mitte: ersterSichtbarer(css), wer: css)
      return ""
    case .klickeText(let aufschrift):
      try await klicken(mitte: bedienelement(aufschrift), wer: "»\(aufschrift)«")
      return ""
    case .tippe(let eingabe):
      textSenden(an: fenster, eingabe)
      try await Task.sleep(nanoseconds: 150_000_000)
      return ""
    case .taste(let kombination):
      let (umschalter, code, zeichen) = try tasteLesen(kombination)
      tasteSenden(an: fenster, umschalter: umschalter, code: code, zeichen: zeichen,
                  alsAequivalent: motor.name == "wk2")
      try await Task.sleep(nanoseconds: 200_000_000)
      return ""
    case .pruefe(let ausdruck):
      let ergebnis = try await motor.auswerten(ausdruck)
      guard ergebnis == "true" else { throw Schrittfehler(description: "ergab \(ergebnis)") }
      return ""
    case .gleich(let ausdruck, let erwartet):
      let ergebnis = try await motor.auswerten(ausdruck)
      let gelesen = (try? JSONSerialization.jsonObject(with: Data(ergebnis.utf8), options: .fragmentsAllowed) as? String) ?? ergebnis
      guard gelesen == erwartet else {
        throw Schrittfehler(description: "erwartet »\(erwartet)«, bekommen »\(gelesen)«")
      }
      return ""
    case .zeige(let ausdruck):
      return try await motor.auswerten(ausdruck)
    case .sichere(let name, let ausdruck):
      let ergebnis = try await motor.auswerten(ausdruck)
      try Data(ergebnis.utf8).write(to: ausgabe("\(name).json"))
      return "gesichert: \(name).json (\(ergebnis.utf8.count) Bytes)"
    case .abzug(let name):
      try await motor.abzug().write(to: ausgabe("\(name).png"))
      return "Bild: \(name).png"
    case .baum(let css):
      return try await baum(css)
    case .fokus:
      return try await fokus()
    case .fenster(let name):
      return try fensterBeschreiben(name)
    case .erscheinung(let art):
      let aussehen = NSAppearance(named: art == "dunkel" ? .darkAqua : .aqua)
      fenster.appearance = aussehen
      motor.ansicht.appearance = aussehen
      try await Task.sleep(nanoseconds: 300_000_000)
      return ""
    case .groesse(let breite, let hoehe):
      // **Das Fenster und nicht die Ansicht**: die Ansicht ist sein Inhalt
      // und folgt ihm; WebKit meldet danach `resize`, wie im Wails-Fenster.
      fenster.setContentSize(NSSize(width: breite, height: hoehe))
      try await Task.sleep(nanoseconds: 500_000_000)
      return "\(Int(breite)) × \(Int(hoehe)) Punkte"
    case .pause(let ms):
      try await Task.sleep(nanoseconds: UInt64(ms) * 1_000_000)
      return ""
    case .frist(let ms):
      frist = ms
      return ""
    }
  }

  private func json(_ text: String) -> String {
    let daten = try? JSONSerialization.data(withJSONObject: [text])
    let feld = daten.flatMap { String(data: $0, encoding: .utf8) } ?? "[\"\"]"
    return String(feld.dropFirst().dropLast())
  }

  private func sichtbar(_ css: String) -> String {
    "(\(ersterSichtbarer(css)) !== null)"
  }

  // ersterSichtbarer findet den ersten **sichtbaren** Treffer eines Wählers.
  //
  // **Nicht den ersten im Baum**: seit die Anwendung verlassene Seiten
  // verborgen aufhebt (Vermerk 0133, UX-03), steht dieselbe Überschrift, derselbe
  // Dialog, derselbe Knopf mehrfach im Baum — sichtbar ist einer. Der erste
  // Treffer war dann oft ein verborgener, und `warte` lief in seine Frist.
  private func ersterSichtbarer(_ css: String) -> String {
    "([...document.querySelectorAll(\(json(css)))].find((e) => { " +
      "const r = e.getBoundingClientRect(); const s = getComputedStyle(e); " +
      "return r.width > 0 && r.height > 0 && s.visibility !== 'hidden' }) ?? null)"
  }

  // bedienelement findet ein Bedienelement über das, was darauf steht — den
  // sichtbaren Text, `aria-label` oder, bei einem Feld, die Beschriftung —,
  // ohne Rücksicht auf Leerraum. **Über die Beschriftung**, weil ein Mensch
  // ein Feld so findet: „Sprache", nicht `input[list]`.
  //
  // **Und auch ohne das stumme Zeichen davor** (UX-10.AK-01, #1682): ein
  // Knopf aus dem Aktionskatalog trägt „✕ Löschen", und wer ihn meint, sagt
  // „Löschen" — wie ein Vorleseprogramm, dem das Zeichen verborgen ist.
  private func bedienelement(_ aufschrift: String) -> String {
    "[...document.querySelectorAll('button, a, [role=button], [role=tab], [role=menuitem], " +
      "[role=option], [role=link], summary, input, select, textarea')].find(e => { " +
      "const r = e.getBoundingClientRect(); if (r.width === 0 || r.height === 0) return false; " +
      "const glatt = (s) => (s || '').replace(/\\s+/g, ' ').trim(); " +
      "const lesbar = [...e.querySelectorAll('[aria-hidden=\"true\"]')].reduce((t, x) => t.replace(x.innerText, ''), e.innerText || ''); " +
      "const beschriftet = e.labels ? [...e.labels].some(l => glatt(l.innerText) === \(json(aufschrift))) : false; " +
      "return glatt(e.innerText) === \(json(aufschrift)) || glatt(lesbar) === \(json(aufschrift)) || e.getAttribute('aria-label') === \(json(aufschrift)) || beschriftet })"
  }

  private func warten(_ was: String, bis bedingung: String) async throws {
    let ende = Date().addingTimeInterval(Double(frist) / 1000)
    var letztes = ""
    while Date() < ende {
      letztes = try await motor.auswerten(bedingung)
      if letztes == "true" { return }
      try await Task.sleep(nanoseconds: 100_000_000)
    }
    throw Schrittfehler(description: "\(was) nach \(frist) ms nicht erreicht (zuletzt \(letztes))")
  }

  private func klicken(mitte element: String, wer: String) async throws {
    let ort = try await motor.auswerten(
      "(() => { const e = \(element); if (!e) return null; e.scrollIntoView({ block: 'center' }); " +
        "const r = e.getBoundingClientRect(); return [r.left + r.width / 2, r.top + r.height / 2] })()")
    guard let paar = try? JSONSerialization.jsonObject(with: Data(ort.utf8)) as? [Double], paar.count == 2 else {
      throw Schrittfehler(description: "\(wer) steht nicht auf der Seite")
    }
    klickSenden(an: fenster, ansicht: motor.ansicht, punkt: NSPoint(x: paar[0], y: paar[1]))
    try await Task.sleep(nanoseconds: 200_000_000)
  }

  private func baum(_ css: String?) async throws -> String {
    guard motor.name == "wk1" else { throw Schrittfehler(description: "»baum« braucht den Motor wk1") }
    guard let flaeche = webflaeche(in: motor.ansicht) else { throw Schrittfehler(description: "keine Webfläche im Baum") }

    var wurzel: AnyObject = flaeche
    if let css {
      let kennung = try await motor.auswerten(
        "(() => { const e = \(ersterSichtbarer(css)) ?? document.querySelector(\(json(css))); if (!e) return null; " +
          "if (!e.id) e.id = '__probe' + Math.random().toString(36).slice(2); return e.id })()")
      let id = (try? JSONSerialization.jsonObject(with: Data(kennung.utf8), options: .fragmentsAllowed) as? String) ?? ""
      guard !id.isEmpty, let element = suche(ab: flaeche, { text($0, "AXDOMIdentifier") == id }) else {
        throw Schrittfehler(description: "\(css) steht nicht im Zugänglichkeitsbaum")
      }
      wurzel = element
    }

    var zeilen: [String] = []
    baumzeilen(wurzel, zeilen: &zeilen)
    return zeilen.joined(separator: "\n")
  }

  private func fokus() async throws -> String {
    guard motor.name == "wk1" else { throw Schrittfehler(description: "»fokus« braucht den Motor wk1") }
    let dom = try await motor.auswerten(
      "(() => { const e = document.activeElement; return e ? e.tagName.toLowerCase() + " +
        "(e.id ? '#' + e.id : '') + ' ' + (e.getAttribute('aria-label') || '') : null })()")
    guard let flaeche = webflaeche(in: motor.ansicht),
          let element = suche(ab: flaeche, { text($0, "AXFocused") == "1" && text($0, "AXRole") != "AXWebArea" }) else {
      return "Ansage: (nichts fokussiert) — DOM: \(dom)"
    }
    // **Und jedes Attribut, das WebCore dazu kennt** — die Ansage oben ist
    // eine Auswahl, und was sie nicht nennt, steht hier.
    let namen = (element.accessibilityAttributeNames?() ?? []).map(\.rawValue).sorted()
    let alle = namen.compactMap { name -> String? in
      let wert = text(element, name)
      return wert.isEmpty || wert.count > 200 ? nil : "  \(name) = \(wert)"
    }
    return (["Ansage: \(text(element, "AXRole")) \(ansage(element)) — DOM: \(dom)"] + alle)
      .joined(separator: "\n")
  }

  private func fensterBeschreiben(_ name: String?) throws -> String {
    guard motor.name == "wk2" else { throw Schrittfehler(description: "»fenster« braucht den Motor wk2") }
    var zeilen: [String] = []
    for (nummer, eintrag) in fensterzeilen(ohne: fenster).enumerated() {
      zeilen += eintrag.zeilen
      if let name, let inhalt = eintrag.fenster.contentView,
         let rep = inhalt.bitmapImageRepForCachingDisplay(in: inhalt.bounds) {
        inhalt.cacheDisplay(in: inhalt.bounds, to: rep)
        try pngAus(rep).write(to: ausgabe("\(name)-\(nummer + 1).png"))
        zeilen.append("  Bild: \(name)-\(nummer + 1).png")
      }
    }
    return zeilen.isEmpty ? "(kein weiteres Fenster)" : zeilen.joined(separator: "\n")
  }
}

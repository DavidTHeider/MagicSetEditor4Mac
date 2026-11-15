// **Was ein Vorleser sieht** — WebCores Zugänglichkeitsbaum, gelesen im
// eigenen Prozess (Motor `wk1`), und die nativen Fenster, die WebKit neben
// die Seite stellt (Motor `wk2`).
//
// # Warum das die VoiceOver-Stichprobe trägt
//
// **VoiceOver liest genau diese Attribute** — `AXRole`, `AXRoleDescription`,
// `AXTitle`, `AXDescription`, `AXValue` und die Zustände — über dieselbe
// Schnittstelle, nur aus einem anderen Prozess. Was hier fehlt, fehlt dort
// auch. **Was hier nicht steht**, ist VoiceOvers eigene Satzbildung und seine
// Navigationsbefehle: die Reihenfolge der Ansage und Kürzel wie VO-→.
import AppKit

@MainActor func attribut(_ element: AnyObject, _ name: String) -> Any? {
  element.accessibilityAttributeValue?(NSAccessibility.Attribute(rawValue: name))
}

@MainActor func text(_ element: AnyObject, _ name: String) -> String {
  guard let wert = attribut(element, name) else { return "" }
  if let zeichen = wert as? String { return zeichen }
  if let zahl = wert as? NSNumber { return zahl.stringValue }
  if let attributiert = wert as? NSAttributedString { return attributiert.string }
  return "\(wert)"
}

@MainActor func kinder(_ element: AnyObject) -> [AnyObject] {
  (attribut(element, "AXChildren") as? [AnyObject]) ?? []
}

// suche geht breit durch den Baum und gibt das erste passende Element.
@MainActor func suche(ab wurzel: AnyObject, grenze: Int = 20000,
                      _ passt: (AnyObject) -> Bool) -> AnyObject? {
  var offen = [wurzel]
  var gesehen = 0

  while !offen.isEmpty, gesehen < grenze {
    let element = offen.removeFirst()
    gesehen += 1
    if passt(element) { return element }
    offen.append(contentsOf: kinder(element))
  }

  return nil
}

@MainActor func webflaeche(in ansicht: NSView) -> AnyObject? {
  suche(ab: ansicht) { text($0, "AXRole") == "AXWebArea" }
}

// ansage ist eine Zeile so, wie ein Vorleser sie zusammensetzt: Name, Rolle,
// Wert und die Zustände, die er nennt.
@MainActor func ansage(_ element: AnyObject) -> String {
  let name = [text(element, "AXTitle"), text(element, "AXDescription")]
    .first { !$0.isEmpty } ?? ""
  var teile = ["«\(name)»", text(element, "AXRoleDescription")]

  let wert = text(element, "AXValue")
  if !wert.isEmpty, wert != name { teile.append("Wert «\(wert)»") }

  let zustaende: [(String, String)] = [
    ("AXRequired", "Pflicht"), ("AXInvalid", "ungültig"), ("AXExpanded", "aufgeklappt"),
    ("AXSelected", "gewählt"), ("AXFocused", "fokussiert"),
  ]
  for (schluessel, wort) in zustaende where text(element, schluessel) == "1" || text(element, schluessel) == "true" {
    teile.append(wort)
  }
  if text(element, "AXEnabled") == "0" { teile.append("nicht verfügbar") }

  let hilfe = text(element, "AXHelp")
  if !hilfe.isEmpty { teile.append("Hilfe «\(hilfe)»") }

  return teile.filter { !$0.isEmpty }.joined(separator: " · ")
}

// baumzeilen schreibt den Baum ab einem Element — eine Ansage je Zeile,
// eingerückt nach Tiefe. Namenlose Gruppen mit nur einem Kind fallen weg:
// sie sagt kein Vorleser an, und der Baum würde unlesbar tief.
@MainActor func baumzeilen(_ element: AnyObject, tiefe: Int = 0, zeilen: inout [String]) {
  guard tiefe < 40, zeilen.count < 4000 else { return }
  let unter = kinder(element)
  let rolle = text(element, "AXRole")
  let namenlos = text(element, "AXTitle").isEmpty && text(element, "AXDescription").isEmpty

  if rolle == "AXGroup", namenlos, unter.count == 1 {
    baumzeilen(unter[0], tiefe: tiefe, zeilen: &zeilen)
    return
  }

  zeilen.append(String(repeating: "  ", count: tiefe) + "\(rolle) " + ansage(element))
  for kind in unter { baumzeilen(kind, tiefe: tiefe + 1, zeilen: &zeilen) }
}

// fensterzeilen beschreibt die nativen Fenster dieses Prozesses außer dem
// Probefenster — und liest aus einer Tabelle ihre Zeilen.
@MainActor func fensterzeilen(ohne haupt: NSWindow) -> [(fenster: NSWindow, zeilen: [String])] {
  NSApp.windows.filter { $0 !== haupt && $0.isVisible }.map { fenster in
    var zeilen = ["\(type(of: fenster)) \(Int(fenster.frame.width))×\(Int(fenster.frame.height))"]
    if let inhalt = fenster.contentView {
      for tabelle in tabellen(in: inhalt) {
        for zeile in 0..<tabelle.numberOfRows {
          let zelle = tabelle.view(atColumn: 0, row: zeile, makeIfNecessary: true)
          let worte = zelle.map { beschriftungen(in: $0) } ?? []
          let gewaehlt = tabelle.selectedRowIndexes.contains(zeile) ? " (gewählt)" : ""
          zeilen.append("  Zeile \(zeile + 1): \(worte.joined(separator: " — "))\(gewaehlt)")
        }
      }
    }
    return (fenster, zeilen)
  }
}

@MainActor func tabellen(in ansicht: NSView) -> [NSTableView] {
  var gefunden: [NSTableView] = []
  if let tabelle = ansicht as? NSTableView { gefunden.append(tabelle) }
  for unter in ansicht.subviews { gefunden.append(contentsOf: tabellen(in: unter)) }
  return gefunden
}

@MainActor func beschriftungen(in ansicht: NSView) -> [String] {
  var worte: [String] = []
  if let feld = ansicht as? NSTextField, !feld.stringValue.isEmpty { worte.append(feld.stringValue) }
  for unter in ansicht.subviews { worte.append(contentsOf: beschriftungen(in: unter)) }
  return worte
}

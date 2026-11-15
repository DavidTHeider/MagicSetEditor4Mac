// **Das Drehbuch einer Probe** — eine Zeile, ein Schritt.
//
// # Warum eine eigene kleine Sprache und kein JavaScript
//
// **Weil die Schritte keine Seitenskripte sind, sondern Handlungen eines
// Menschen**: ein Klick, ein Tastendruck, ein Blick auf den Bildschirm. Sie
// gehen als echte AppKit-Ereignisse an das Fenster — ein `element.click()`
// aus JavaScript wäre ein synthetisches Ereignis, und WebKit behandelt es
// anders (ein Klick fokussiert dort keinen Knopf, ein Skriptklick schon).
//
// Jede Zeile beginnt mit einem Wort; `#` leitet einen Kommentar ein.
//
//   motor wk2|wk1          wk2: WKWebView wie in Wails — wk1: WebCore im
//                           eigenen Prozess, für den Zugänglichkeitsbaum
//   lade <pfad>            relativ zur Basis, oder eine volle Adresse
//   warte <css>            bis ein sichtbares Element passt
//   wartetext <text>       bis der Text auf der Seite steht
//   bis <js>               bis der Ausdruck wahr ist — etwa eine Messung fertig
//   klicke <css>           echter Mausklick in die Mitte des Elements
//   klicketext <text>      dasselbe über die Aufschrift eines Bedienelements
//                           oder die Beschriftung eines Felds
//   tippe <text>           echte Tastenereignisse, Zeichen für Zeichen
//   taste <kombination>    tab, shift+tab, escape, enter, down, cmd+k …
//   pruefe <js>            der Ausdruck muss wahr sein
//   gleich <js> => <wert>  der Ausdruck muss genau diesen Text ergeben
//   zeige <js>             schreibt das Ergebnis in den Bericht
//   sichere <name> <js>    schreibt das Ergebnis als <name>.json daneben
//   abzug <name>           Bild der Seite als <name>.png
//   baum [<css>]           Zugänglichkeitsbaum (wk1), ganz oder ab Element
//   fokus                  was ein Vorleser beim Fokus ansagt (wk1)
//   fenster [<name>]       native Fenster der Probe — etwa WebKits
//                           Vorschlagsliste — samt Zeilen und Bild (wk2)
//   erscheinung hell|dunkel
//   groesse <b> <h>        das Fenster auf so viele Punkte stellen — eine
//                           Anordnung, die an der Breite hängt, mitten im Lauf
//   pause <ms>
//   frist <ms>             wie lange `warte` wartet (Vorgabe 10000)
//   notiz <text>           Überschrift im Bericht
import Foundation

enum Schritt {
  case motor(String)
  case lade(String)
  case warte(String)
  case warteText(String)
  case bis(String)
  case klicke(String)
  case klickeText(String)
  case tippe(String)
  case taste(String)
  case pruefe(String)
  case gleich(String, String)
  case zeige(String)
  case sichere(String, String)
  case abzug(String)
  case baum(String?)
  case fokus
  case fenster(String?)
  case erscheinung(String)
  case groesse(Double, Double)
  case pause(Int)
  case frist(Int)
  case notiz(String)
}

struct Drehbuchfehler: Error, CustomStringConvertible {
  let zeile: Int
  let grund: String
  var description: String { "Zeile \(zeile): \(grund)" }
}

// drehbuchLesen zerlegt eine Probe in Schritte — mit der Zeilennummer, damit
// ein Befund im Bericht auf seine Zeile zeigt.
func drehbuchLesen(_ text: String) throws -> [(zeile: Int, schritt: Schritt, roh: String)] {
  var schritte: [(Int, Schritt, String)] = []

  for (index, rohzeile) in text.components(separatedBy: "\n").enumerated() {
    let zeile = rohzeile.trimmingCharacters(in: .whitespaces)
    if zeile.isEmpty || zeile.hasPrefix("#") { continue }

    let teile = zeile.split(separator: " ", maxSplits: 1, omittingEmptySubsequences: true)
    let wort = String(teile[0]).lowercased()
    let rest = teile.count > 1 ? String(teile[1]).trimmingCharacters(in: .whitespaces) : ""
    let nummer = index + 1

    func verlangt(_ was: String) throws -> String {
      if rest.isEmpty { throw Drehbuchfehler(zeile: nummer, grund: "»\(wort)« verlangt \(was)") }
      return rest
    }

    func zahl() throws -> Int {
      guard let wert = Int(try verlangt("eine Zahl")) else {
        throw Drehbuchfehler(zeile: nummer, grund: "»\(rest)« ist keine Zahl")
      }
      return wert
    }

    let schritt: Schritt
    switch wort {
    case "motor": schritt = .motor(try verlangt("wk1 oder wk2"))
    case "lade": schritt = .lade(try verlangt("einen Pfad"))
    case "warte": schritt = .warte(try verlangt("einen CSS-Wähler"))
    case "wartetext": schritt = .warteText(try verlangt("einen Text"))
    case "bis": schritt = .bis(try verlangt("einen Ausdruck"))
    case "klicke": schritt = .klicke(try verlangt("einen CSS-Wähler"))
    case "klicketext": schritt = .klickeText(try verlangt("eine Aufschrift"))
    case "tippe": schritt = .tippe(try verlangt("einen Text"))
    case "taste": schritt = .taste(try verlangt("eine Taste"))
    case "pruefe": schritt = .pruefe(try verlangt("einen Ausdruck"))
    case "gleich":
      // **Getrennt wird am letzten » => «** — der Ausdruck davor darf selbst
      // Pfeilfunktionen enthalten, der erwartete Wert dahinter nicht.
      let teile = try verlangt("»<js> => <wert>«").components(separatedBy: " => ")
      guard teile.count >= 2 else {
        throw Drehbuchfehler(zeile: nummer, grund: "»gleich« braucht ein » => «")
      }
      schritt = .gleich(teile.dropLast().joined(separator: " => "), teile[teile.count - 1])
    case "zeige": schritt = .zeige(try verlangt("einen Ausdruck"))
    case "sichere":
      let paar = try verlangt("»<name> <js>«").split(separator: " ", maxSplits: 1).map(String.init)
      guard paar.count == 2 else {
        throw Drehbuchfehler(zeile: nummer, grund: "»sichere« braucht einen Namen und einen Ausdruck")
      }
      schritt = .sichere(paar[0], paar[1])
    case "abzug": schritt = .abzug(try verlangt("einen Namen"))
    case "baum": schritt = .baum(rest.isEmpty ? nil : rest)
    case "fokus": schritt = .fokus
    case "fenster": schritt = .fenster(rest.isEmpty ? nil : rest)
    case "erscheinung": schritt = .erscheinung(try verlangt("hell oder dunkel"))
    case "groesse":
      let masse = try verlangt("»<breite> <hoehe>«").split(separator: " ").compactMap { Double($0) }
      guard masse.count == 2, masse.allSatisfy({ $0 >= 200 }) else {
        throw Drehbuchfehler(zeile: nummer, grund: "»groesse« braucht Breite und Höhe in Punkten, je ab 200")
      }
      schritt = .groesse(masse[0], masse[1])
    case "pause": schritt = .pause(try zahl())
    case "frist": schritt = .frist(try zahl())
    case "notiz": schritt = .notiz(try verlangt("einen Text"))
    default:
      throw Drehbuchfehler(zeile: nummer, grund: "unbekannter Schritt »\(wort)«")
    }

    schritte.append((nummer, schritt, zeile))
  }

  return schritte
}

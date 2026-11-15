// Misst, wie Quartz' Schriftglättung einen Umriss verdickt — die Probe hinter
// `internal/render/glaettung.go` (#1806) und `docs/SCHRIFTUNTERSUCHUNG.md`.
//
// Gezeichnet wird in einen Bitmapkontext wie den von wxWidgets (24 Bit, weiß auf
// schwarz), einmal ohne Glättung und einmal ohne Angabe — wxWidgets schaltet sie
// nicht um, und ohne Angabe ist gleich mit.
//
// Aufruf:  swift scripts/coretext/glaettung.swift kanten <schrift> <groessen>
//          swift scripts/coretext/glaettung.swift winkel <schrift> <groesse> <winkel>
//
// `kanten` gibt je Größe (Bildpunkte je Geviert) die Tinte, die jede Kante eines
// Stamms zulegt, in Bildpunkten: links und rechts am senkrechten Stamm von `l`,
// oben und unten am waagerechten eines um 90° gedrehten `l`.
// `winkel` gibt je Drehung die Tinte von `l` ohne und mit Glättung.
import CoreGraphics
import CoreText
import Foundation

let argumente = CommandLine.arguments
guard argumente.count >= 4,
      let deskriptor = (CTFontManagerCreateFontDescriptorsFromURL(
          URL(fileURLWithPath: argumente[2]) as CFURL) as? [CTFontDescriptor])?.first
else {
    FileHandle.standardError.write("Aufruf siehe Kopf von glaettung.swift\n".data(using: .utf8)!)
    exit(2)
}

/// Zeichnet `l` in ein quadratisches Feld und gibt die Deckung, Zeile für Zeile von oben.
func zeichne(_ groesse: Double, _ glatt: Bool, _ winkel: Double) -> ([UInt8], Int) {
    let seite = Int(groesse * 3)
    let ctx = CGContext(data: nil, width: seite, height: seite, bitsPerComponent: 8, bytesPerRow: seite * 4,
                        space: CGColorSpaceCreateDeviceRGB(),
                        bitmapInfo: CGImageAlphaInfo.noneSkipFirst.rawValue | CGBitmapInfo.byteOrder32Little.rawValue)!
    ctx.setFillColor(CGColor(red: 0, green: 0, blue: 0, alpha: 1))
    ctx.fill(CGRect(x: 0, y: 0, width: seite, height: seite))
    if !glatt { ctx.setShouldSmoothFonts(false) }
    ctx.setFillColor(CGColor(red: 1, green: 1, blue: 1, alpha: 1))
    ctx.translateBy(x: CGFloat(seite / 2), y: CGFloat(seite / 2))
    ctx.rotate(by: CGFloat(winkel * .pi / 180))
    let schrift = CTFontCreateWithFontDescriptor(deskriptor, CGFloat(groesse), nil)
    let text = NSAttributedString(string: "l", attributes: [
        kCTFontAttributeName as NSAttributedString.Key: schrift,
        kCTForegroundColorFromContextAttributeName as NSAttributedString.Key: true,
    ])
    ctx.textPosition = .zero
    CTLineDraw(CTLineCreateWithAttributedString(text), ctx)
    let daten = ctx.data!.bindMemory(to: UInt8.self, capacity: seite * seite * 4)
    return ((0..<(seite * seite)).map { daten[$0 * 4 + 2] }, seite)
}

/// Die Tinte, die jede Kante quer durch den Stamm zulegt: je Linie quer zum Stamm
/// die Deckung vor und hinter der Mitte, gemittelt über die mittlere Hälfte des Stamms.
func kanten(_ groesse: Double, gedreht: Bool) -> (Double, Double) {
    let (ohne, seite) = zeichne(groesse, false, gedreht ? 90 : 0)
    let (mit, _) = zeichne(groesse, true, gedreht ? 90 : 0)
    func wert(_ bild: [UInt8], _ quer: Int, _ laengs: Int) -> Double {
        Double(gedreht ? bild[quer * seite + laengs] : bild[laengs * seite + quer]) / 255
    }
    // die Linien quer zum Stamm, die ohne Glättung Tinte tragen
    let linien = (0..<seite).filter { l in (0..<seite).contains { wert(ohne, $0, l) > 0 } }
    let mitte = linien[linien.count / 4 ..< linien.count * 3 / 4]
    var vorn = 0.0, hinten = 0.0
    for l in mitte {
        let tinte = (0..<seite).filter { wert(ohne, $0, l) > 0 }
        let halb = (tinte.first! + tinte.last!) / 2
        for q in 0..<seite {
            let zuwachs = wert(mit, q, l) - wert(ohne, q, l)
            if q <= halb { vorn += zuwachs } else { hinten += zuwachs }
        }
    }
    return (vorn / Double(mitte.count), hinten / Double(mitte.count))
}

switch argumente[1] {
case "kanten":
    print("Geviert\tlinks\trechts\toben\tunten")
    for groesse in argumente[3].split(separator: ",").compactMap({ Double($0) }) {
        let (links, rechts) = kanten(groesse, gedreht: false)
        let (oben, unten) = kanten(groesse, gedreht: true)
        print(String(format: "%.1f\t%.3f\t%.3f\t%.3f\t%.3f", groesse, links, rechts, oben, unten))
    }
case "winkel":
    guard argumente.count == 5, let groesse = Double(argumente[3]) else { exit(2) }
    print("Winkel\tohne\tmit\tZuwachs")
    for winkel in argumente[4].split(separator: ",").compactMap({ Double($0) }) {
        let ohne = zeichne(groesse, false, winkel).0.reduce(0) { $0 + Double($1) } / 255
        let mit = zeichne(groesse, true, winkel).0.reduce(0) { $0 + Double($1) } / 255
        print(String(format: "%.1f\t%.2f\t%.2f\t%.3f", winkel, ohne, mit, mit - ohne))
    }
default:
    exit(2)
}

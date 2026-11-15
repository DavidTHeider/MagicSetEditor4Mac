// Rastert eine Zeile auf MSE2s Weg — Core Text in vierfacher Größe, weiß auf schwarz, dann
// der 4×4-Kastenfilter aus `downsample_to_alpha` — und schreibt die Deckung als PGM.
// Die Probe hinter docs/SCHRIFTUNTERSUCHUNG.md (QA-03.AK-02, #1656).
//
// `draw_resampled_text`, `src/gfx/resample_text.cpp`, Zeilen 166–211, Fassung f9d9356
// vom 2021-07-06: der Puffer ist 24 Bit tief, die Stelle darin rastet auf Viertelpunkte
// (`xsub`), und gemittelt wird zweimal ganzzahlig — erst waagerecht, dann senkrecht.
//
// Aufruf:  swift scripts/coretext/raster.swift <schrift> <groesse> <x> <grundlinie>
//              <breite> <hoehe> <glaettung: an|aus|vorgabe> <ausgabe.pgm> <text>
// `vorgabe` lässt den Kontext, wie wxWidgets ihn anlegt: es schaltet die Glättung nicht.
import CoreGraphics
import CoreText
import Foundation

let a = CommandLine.arguments
guard a.count == 10,
      let deskriptor = (CTFontManagerCreateFontDescriptorsFromURL(
          URL(fileURLWithPath: a[1]) as CFURL) as? [CTFontDescriptor])?.first,
      let groesse = Double(a[2]), let x = Double(a[3]), let grund = Double(a[4]),
      let w = Int(a[5]), let h = Int(a[6])
else {
    FileHandle.standardError.write("Aufruf siehe Kopf von raster.swift\n".data(using: .utf8)!)
    exit(2)
}

let faktor = 4 // `text_scaling`, Zeile 22
let schrift = CTFontCreateWithFontDescriptor(deskriptor, CGFloat(groesse * Double(faktor)), nil)
let breite = w * faktor, hoehe = h * faktor

guard let ctx = CGContext(data: nil, width: breite, height: hoehe, bitsPerComponent: 8,
                          bytesPerRow: breite * 4, space: CGColorSpaceCreateDeviceRGB(),
                          bitmapInfo: CGImageAlphaInfo.noneSkipFirst.rawValue | CGBitmapInfo.byteOrder32Little.rawValue)
else { exit(1) }

ctx.setFillColor(CGColor(red: 0, green: 0, blue: 0, alpha: 1))
ctx.fill(CGRect(x: 0, y: 0, width: breite, height: hoehe))
switch a[7] {
case "an": ctx.setShouldSmoothFonts(true)
case "aus": ctx.setShouldSmoothFonts(false)
default: break
}
ctx.setFillColor(CGColor(red: 1, green: 1, blue: 1, alpha: 1))

let text = NSAttributedString(string: a[9], attributes: [
    kCTFontAttributeName as NSAttributedString.Key: schrift,
    kCTForegroundColorFromContextAttributeName as NSAttributedString.Key: true,
])
// `xsub = int(text_scaling * (pos.x - xi))`
let xi = x.rounded(.towardZero), xsub = Double(Int(Double(faktor) * (x - xi)))
ctx.textPosition = CGPoint(x: xi * Double(faktor) + xsub, y: Double(hoehe) - grund * Double(faktor))
CTLineDraw(CTLineCreateWithAttributedString(text), ctx)

guard let daten = ctx.data?.bindMemory(to: UInt8.self, capacity: breite * hoehe * 4) else { exit(1) }
var pgm = "P5\n\(w) \(h)\n255\n".data(using: .ascii)!
for y in 0..<h {
    for xx in 0..<w {
        // zweimal ganzzahlig geteilt, wie `downsample_to_alpha`
        var spalte = 0
        for j in 0..<faktor {
            var zeile = 0
            for i in 0..<faktor { zeile += Int(daten[((y * faktor + j) * breite + xx * faktor + i) * 4 + 2]) }
            spalte += zeile / faktor
        }
        pgm.append(UInt8(spalte / faktor))
    }
}
try pgm.write(to: URL(fileURLWithPath: a[8]))

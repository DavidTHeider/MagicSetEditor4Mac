// Fragt Core Text, wie breit es eine Zeile setzt und wie viele Glyphen es dafür nimmt —
// die Probe hinter docs/SCHRIFTUNTERSUCHUNG.md (QA-03.AK-02, #1656).
//
// **Wie wxWidgets fragt**: die Zeile trägt nur die Schrift, kein Merkmal ist umgestellt
// (`kCTLigatureAttributeName`, `kCTKernAttributeName` und `kCTFontFeatureSettingsAttribute`
// kommen in wxWidgets nicht vor), gemessen mit `CTLineGetTypographicBounds` wie in
// `wxMacCoreGraphicsContext::GetTextExtent`.
//
// Aufruf:  swift scripts/coretext/breiten.swift <schriftdatei> <groesse> < zeilen.txt
// Ausgabe: je Zeile  <breite> TAB <glyphen> TAB <zeile>
// Weniger Glyphen als Zeichen heißt: Core Text hat eine Ligatur gesetzt.
import CoreText
import Foundation

let argumente = CommandLine.arguments
guard argumente.count == 3,
      let deskriptoren = CTFontManagerCreateFontDescriptorsFromURL(
          URL(fileURLWithPath: argumente[1]) as CFURL) as? [CTFontDescriptor],
      let deskriptor = deskriptoren.first,
      let groesse = Double(argumente[2])
else {
    FileHandle.standardError.write("Aufruf: breiten.swift <schriftdatei> <groesse> < zeilen\n".data(using: .utf8)!)
    exit(2)
}

let schrift = CTFontCreateWithFontDescriptor(deskriptor, CGFloat(groesse), nil)

while let zeile = readLine(strippingNewline: true) {
    let text = NSAttributedString(string: zeile,
                                  attributes: [kCTFontAttributeName as NSAttributedString.Key: schrift])
    let gesetzt = CTLineCreateWithAttributedString(text)
    var glyphen = 0
    for lauf in CTLineGetGlyphRuns(gesetzt) as! [CTRun] {
        glyphen += CTRunGetGlyphCount(lauf)
    }
    print(String(format: "%.4f", CTLineGetTypographicBounds(gesetzt, nil, nil, nil)) + "\t\(glyphen)\t" + zeile)
}

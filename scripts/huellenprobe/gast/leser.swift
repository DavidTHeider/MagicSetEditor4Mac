// **Ein Leser über die native API** — I01
// (#2259).
//
// Er liest ein Geheimnis mit `SecItemCopyMatching` und sagt nur, **ob**: die
// Länge oder den Status, nie den Wert. **Ohne Oberfläche** (der Gast wird
// über `lume ssh` gefahren) kann macOS nicht fragen; wo ein Zugriff eine
// Bestätigung bräuchte, antwortet die API mit `-25308` („User interaction is
// not allowed"). So trennt der Lauf »liest still« von »müsste fragen«.
//
// Übersetzt wird er am Wirt in zwei Fassungen (`FASSUNG` verschieden), damit
// ihre cdhashes verschieden sind — wie zwei Fassungen eines Programms.
//
//     swiftc -O -D FASSUNG_ZWEI leser.swift -o leser-v2
import Foundation
import Security

#if FASSUNG_ZWEI
let fassung = "v2"
#else
let fassung = "v1"
#endif

let dienst = CommandLine.arguments.count > 1 ? CommandLine.arguments[1] : "i01"
let anfrage: [String: Any] = [
  kSecClass as String: kSecClassGenericPassword,
  kSecAttrService as String: dienst,
  kSecReturnData as String: true,
  kSecMatchLimit as String: kSecMatchLimitOne,
]
var ergebnis: CFTypeRef?
let status = SecItemCopyMatching(anfrage as CFDictionary, &ergebnis)
if status == errSecSuccess, let daten = ergebnis as? Data {
  print("\(fassung) GELESEN \(daten.count) Bytes")
} else {
  let satz = SecCopyErrorMessageString(status, nil) as String? ?? "?"
  print("\(fassung) STATUS \(status) \(satz)")
}

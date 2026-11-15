// **KI-Wege ohne Modellrechnung** — eine Gruppe des Vertragsrundgangs (B01, #2220).
//
// Hier steht, was an einer KI-Bindung **ohne** einen Lauf des Modells geht:
// die Freigabe eines Menschen für einen kostenpflichtigen Auftrag. **Alles
// andere setzt einen Vorschlag oder eine Messung voraus, und die entstehen nur
// aus einem Modell** — und das gibt es im Rundgang nicht: keine Laufzeit
// (`ki.laufzeit` ist aus, die Modelle der KI-Phase sind gelöscht), kein
// Cloudkonto, kein Netz. Diese Rufe bekommen `fachfehler` mit dem Grund; sie
// gehen mit gültigen Argumenten bis dorthin, wo die Voraussetzung fehlt.
//
// Es wird nichts gerechnet, nichts gebucht und nichts über das Netz gesendet:
// der Anbieter `anthropic` ist ein Dienst, der **aus** ist (`dienst.anthropic`),
// und `quelleFuer` weist ihn ab, bevor eine Anfrage entsteht
// (`internal/app/kiwerkzeug.go`).

// vorschlag ist eine Kennung, unter der nichts wartet.
const vorschlag = 'vertrag-ohne-vorschlag'

export const aKI = {
  name: 'KI ohne Modellrechnung (A)',
  rufe: [
    {
      // **Die Freigabe ist ein Beleg ohne Auftrag**: sie gilt fünf Minuten
      // (`KIFreigabefrist`) und läuft ungenutzt ab. Ausgestellt wird sie am
      // Fenster, von einem Menschen mit der Rolle Mitarbeit — `Nachweise.Ausstellen`
      // weist jeden anderen Modus ab. Betrag in der kleinsten Einheit.
      name: 'KIFreigabeErteilen', argumente: () => ['anthropic', 100, 'EUR'],
      nimm: 'a.kennung',
    },
    {
      name: 'KIModellPruefen', argumente: () => ['anthropic', 'vertrag-modell', ''], erwartet: 'fachfehler',
      grund: 'eine Messung ruft das Modell; der Anbieter ist aus und es gibt kein Konto (kimodellprobe.go, quelleFuer)',
    },
    {
      name: 'KIVorschlagUebernehmen', argumente: () => [vorschlag, ['name']], erwartet: 'fachfehler',
      grund: 'es wartet kein Vorschlag; er entstünde nur aus einem Modellauf (KIVorschlagen)',
    },
    {
      name: 'KIGruppeUebernehmen', argumente: () => [[{ kennung: vorschlag, felder: ['name'] }]], erwartet: 'fachfehler',
      grund: 'es wartet kein Vorschlag; er entstünde nur aus einem Modellauf (KIBatchVorschlagen)',
    },
  ],
}

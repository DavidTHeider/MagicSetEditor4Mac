// **Was geschrieben ist, und was nicht** — S08.NAT-03
// (#1680):
// *„Teilerfolg benennt geschriebene Daten und verbleibenden Fehler
// getrennt."*
//
// # Warum zwei Teile und nicht ein Satz
//
// **Weil sie Verschiedenes verlangen.** Was geschrieben ist, steht im Bestand
// und braucht nichts mehr (§7, Regel 3: geschrieben ist geschrieben); was
// nicht mitkam, braucht vielleicht eine Handlung. Ein Satz, der beides
// vermischte — „zwölf Karten kamen nicht mit" allein —, ließe offen, ob das
// Deck überhaupt da ist.
//
// # Und warum der erste Teil ein `status` ist
//
// **Er ist eine gute Nachricht**: ein Vorleser sagt ihn, sobald er Luft
// holt, und unterbricht nichts. Die Liste darunter bleibt stehen, bis der
// nächste Vorgang sie ablöst — ein Teilerfolg verschwindet nicht von selbst
// (UX-08.AK-02).
interface Teilerfolgeigenschaften {
  // geschrieben nennt, was jetzt im Bestand steht.
  geschrieben: string

  // offen sind die Reste, und offenTitel ist die Zeile über ihnen.
  offen: readonly string[]
  offenTitel: string
}

export function Teilerfolg({
  geschrieben,
  offen,
  offenTitel,
}: Teilerfolgeigenschaften) {
  return (
    <div className="flex flex-col gap-2">
      <p role="status" className="text-sm">
        {geschrieben}
      </p>

      {offen.length > 0 && (
        <div className="flex flex-col gap-1">
          <p className="text-sm font-medium">{offenTitel}</p>
          <ul className="flex flex-col gap-1 text-sm text-leise">
            {offen.map((rest) => (
              <li key={rest}>{rest}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

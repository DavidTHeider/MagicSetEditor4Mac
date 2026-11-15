// Platzhalter in einem Satz aus dem Sprachkatalog ersetzen.
//
// **In einem Zug und nicht nacheinander.** Zwei aufeinanderfolgende `replace`
// werten aus, was der erste eingesetzt hat: enthält eine Repositoriumsadresse
// die Zeichenfolge `{stand}`, ersetzte der zweite Durchgang sie — und der
// eigentliche Platzhalter bliebe stehen. Der Nachlauf zum Zweitblick über
// Phase 9 hat es gefunden.
//
// Fremde Daten sind der Normalfall (§8), und eine Adresse ist fremde Eingabe.
export function einsetzen(
  vorlage: string,
  werte: Record<string, string>,
): string {
  return vorlage.replace(/\{(\w+)\}/g, (ganz: string, name: string) => {
    return werte[name] ?? ganz
  })
}

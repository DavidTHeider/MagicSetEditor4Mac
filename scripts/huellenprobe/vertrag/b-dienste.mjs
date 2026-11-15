// **Die zwei Dienste und der Platz** — eine Gruppe des Vertragsrundgangs
// (B01, #2220).
//
// # Auf der Schleife, und wieder zu
//
// `RESTStarten` und `MCPStarten` starten je ein **Unterprozess** neben dem
// Fenster (`internal/app/dienstwerk.go`): `mse-rest` und `mse-mcp` aus dem
// Ordner des Bündels. Die Adresse ist die Schleife (`127.0.0.1`), der Port
// `0` — ein freier vom Betriebssystem, damit nichts mit einem Lauf davor
// streitet. **Beide werden sofort wieder beendet**: ein laufender Dienst
// hielte den Bestand mit offen, und `SicherungZurueckspielen` und
// `VerschluesselungUmschalten` verlangen, allein zu sein
// (`internal/service/bestandsrueckspielen.go`).
//
// **Was dieser Eintrag voraussetzt, ohne es zu prüfen**: dass im Bündel
// `mse-rest` und `mse-mcp` liegen. Fehlen sie, antwortet der Start mit einem
// fachlichen Fehler („starten"), und der Eintrag zeigt es als Abweichung —
// das ist der Befund, nicht ein Fehler dieser Datei.
//
// **Der Platz**: `PlatzFreigeben` räumt die Vorsicherungen (Sicherungen vor
// Migrationen) bis auf den Weg zurück. Es nimmt nur weg, was dieser Lauf in
// seinem eigenen Zuhause selbst angelegt hat.
export const bDienste = {
  name: 'Dienste (REST, MCP) und Platz',
  rufe: [
    { name: 'RESTStarten', argumente: () => ['127.0.0.1', 0] },
    { name: 'RESTBeenden' },
    { name: 'MCPStarten', argumente: () => ['127.0.0.1', 0] },
    { name: 'MCPBeenden' },

    // `vorsicherungen` und `protokoll` sind die zwei räumbaren Posten. Das
    // Protokoll gehört dem laufenden Fenster, und die Probe liest es nach dem
    // Lauf; geräumt werden deshalb nur die Vorsicherungen.
    { name: 'PlatzFreigeben', argumente: () => ['vorsicherungen'] },
  ],
}

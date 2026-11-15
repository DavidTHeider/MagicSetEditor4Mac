// **Die Netzeinstellungen** — Werkstück N5 von Phase 5.4
// (#1571).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenzugaenge.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die drei Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// netznamen sind die Beschriftungen der sechs Einstellungen.
export const netznamen = {
  'netz.proxy.http': 'Proxy für http',
  'netz.proxy.https': 'Proxy für https',
  'netz.proxy.aus': 'Ohne Proxy erreichen',
  'netz.frist.verbindung': 'Frist: Verbindungsaufbau (s)',
  'netz.frist.kopf': 'Frist: Antwortkopf (s)',
  'netz.frist.stillstand': 'Frist: Stillstand (s)',
  'netz.schnittstelle': 'Schnittstelle für die Suche im eigenen Netz',
  'netz.lauschen': 'Wo die eigenen Dienste lauschen',
  'dienst.tls.zertifikat': 'Eigenes Zeugnis der Dienste (Zertifikat)',
  'dienst.tls.schluessel': 'Schlüssel zum eigenen Zeugnis',
  'dienst.tls.kette': 'Zwischenzertifikate des Zeugnisses',
  'dienst.tls.anker': 'Vertrauensanker des Zeugnisses',
  'grenzen.kontingente': 'Kontingente der Dienste',
}

// netzhilfe ist der erklärende Satz je Einstellung.
export const netzhilfe = {
  'netz.proxy.http':
    'Über welchen Proxy `http://`-Adressen gehen. **Leer heißt: `HTTP_PROXY` aus der Umgebung gilt** — steht hier etwas, schlägt es sie. Mit Schema, also `http://proxy.firma:3128`. **Ohne Kennwort** — es läge unverschlüsselt auf der Platte; ein Proxy mit Anmeldung kommt über `HTTPS_PROXY` aus der Umgebung.',
  'netz.proxy.https':
    'Dasselbe für `https://`-Adressen, aus `HTTPS_PROXY`. **Getrennt**, weil viele Netze nur für eines von beiden einen Proxy haben.',
  'netz.proxy.aus':
    'Was **nicht** über den Proxy geht — Rechnernamen und Adressbereiche, mit Komma getrennt, wie `NO_PROXY`. Die Schleifenadresse bleibt ohnehin außen vor; ein Ollama auf einem **anderen** Rechner im eigenen Netz gehört hierher.',
  'netz.frist.verbindung':
    'Wie lange ein Verbindungsaufbau samt TLS dauern darf. Leer heißt: 15 Sekunden.',
  'netz.frist.kopf':
    'Wie lange eine Gegenstelle für ihren Antwortkopf hat. Leer heißt: 30 Sekunden.',
  'netz.frist.stillstand':
    'Wie lange eine laufende Antwort ohne Fortschritt bleiben darf. **Diese ist die, die man anhebt**, wenn ein großer Bezug über eine langsame Leitung abbricht — nicht die beiden darüber. Leer heißt: 60 Sekunden.',
  'netz.schnittstelle':
    'Über welche Netzwerkschnittstelle die Suche im eigenen Netz fragt und die Ankündigung hinausgeht — etwa `en0`. **Leer heißt: automatisch**, jede taugliche Schnittstelle mit IPv4 und Multicast. Ein Name, den es gerade nicht gibt, gilt wie leer und steht als Befund im Protokoll.',
  'netz.lauschen':
    'Adressen, auf denen REST und MCP lauschen, mit Komma getrennt. **Leer heißt: nur dieser Mac** (`127.0.0.1`). `0.0.0.0` heißt jede Schnittstelle und steht allein. **Verlässt eine Adresse den Rechner, gelten TLS und Token**; gewählt wird bequemer unter „Werkzeuge ohne Fenster“. Wirkt beim nächsten Start eines Dienstes.',
  'dienst.tls.zertifikat':
    'Das Zertifikat, das REST- und MCP-Dienst jenseits der Schleife vorzeigen, als PEM. **Leer heißt: das Zeugnis auf der Identität dieses Profils.** Mit einer Stelle deines Hauses prüfen Klienten nach Name und Kette statt nach dem Fingerabdruck. Gilt beim nächsten Start des Dienstes.',
  'dienst.tls.schluessel':
    'Der private Schlüssel zum Zertifikat, als PEM. **Nur für dich lesbar (0600)**, sonst startet der Dienst nicht. Ein Austausch beider Dateien gilt ohne Neustart am nächsten Handschlag.',
  'dienst.tls.kette':
    'Zwischenzertifikate, falls sie nicht im Zertifikat mitstehen. Leer heißt: keine.',
  'dienst.tls.anker':
    'Die Wurzel deiner Stelle, gegen die die Kette vor dem Start geprüft wird. Leer heißt: die Anker des Systems.',
  'grenzen.kontingente':
    'Wie viel ein Maschinenzugang über REST und MCP in der Minute darf — je Zugang und für alle zusammen. **Aus ist für eine kontrollierte Umgebung**, in der niemand einem anderen etwas wegnimmt, etwa einen Lastlauf. Zugang, TLS, die Plätze je Arbeit, die Grenzen für Rumpf und Antwort, die Fristen und die Sperre gegen das Raten bleiben. Gilt beim nächsten Start eines Dienstes.',
}

// netzbereich ist die Überschrift des Bereichs samt Hinweis.
export const netzbereich = {
  netz: 'Wie dieses Programm hinausgeht',
  netzHinweis:
    'Diese Angaben gelten für **jeden** Dienst zugleich — auch für die Kommandozeile, den MCP-Server und den Dienst. Leer heißt: es gilt, was in der Umgebung steht. Was gerade gilt und woher es kommt, steht unter Netz und Erreichbarkeit.',
}

// netzwerte sind die Namen der Werte — QA-08.AK-03 (#1739).
export const netzwerte = {
  'grenzen.kontingente.an': 'An',
  'grenzen.kontingente.aus': 'Aus',
}

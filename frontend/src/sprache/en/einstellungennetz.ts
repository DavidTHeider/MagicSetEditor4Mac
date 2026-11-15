// **The network settings** — work item N5 of phase 5.4
// (#1571).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenzugaenge.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die drei Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// netznamen sind die Beschriftungen der sechs Einstellungen.
export const netznamen = {
  'netz.proxy.http': 'Proxy for http',
  'netz.proxy.https': 'Proxy for https',
  'netz.proxy.aus': 'Reach without proxy',
  'netz.frist.verbindung': 'Timeout: connect (s)',
  'netz.frist.kopf': 'Timeout: response header (s)',
  'netz.frist.stillstand': 'Timeout: no progress (s)',
  'netz.schnittstelle': 'Interface for the local network search',
  'netz.lauschen': 'Where the own services listen',
  'dienst.tls.zertifikat': 'Services’ own certificate',
  'dienst.tls.schluessel': 'Key for the own certificate',
  'dienst.tls.kette': 'Intermediate certificates',
  'dienst.tls.anker': 'Trust anchors for the certificate',
  'grenzen.kontingente': 'Rate budgets of the services',
}

// netzhilfe ist der erklärende Satz je Einstellung.
export const netzhilfe = {
  'netz.proxy.http':
    'Which proxy `http://` addresses go through. **Empty means `HTTP_PROXY` from the environment applies** — anything here overrides it. With a scheme, so `http://proxy.company:3128`. **No password** — it would sit unencrypted on disk; a proxy that needs a login comes from `HTTPS_PROXY` in the environment.',
  'netz.proxy.https':
    'The same for `https://` addresses, from `HTTPS_PROXY`. **Kept apart**, because many networks have a proxy for only one of the two.',
  'netz.proxy.aus':
    'What does **not** go through the proxy — host names and address ranges, comma separated, like `NO_PROXY`. Loopback stays out anyway; an Ollama on **another** machine in your own network belongs here.',
  'netz.frist.verbindung':
    'How long connecting, including TLS, may take. Empty means 15 seconds.',
  'netz.frist.kopf':
    'How long a server has to send its response header. Empty means 30 seconds.',
  'netz.frist.stillstand':
    'How long a running response may stall. **This is the one to raise** when a large download over a slow line breaks off — not the two above. Empty means 60 seconds.',
  'netz.schnittstelle':
    'Which network interface the local network search asks on and the announcement goes out through — for example `en0`. **Empty means automatic**: every suitable interface with IPv4 and multicast. A name that does not exist right now counts as empty and is noted in the log.',
  'netz.lauschen':
    'Addresses REST and MCP listen on, separated by commas. **Empty means: this Mac only** (`127.0.0.1`). `0.0.0.0` means every interface and stands alone. **If an address leaves this computer, TLS and a token apply**; it is easier to choose under “Tools without a window”. Takes effect the next time a service starts.',
  'dienst.tls.zertifikat':
    'The certificate the REST and MCP services present beyond loopback, as PEM. **Empty means: the certificate on this profile’s identity.** With your own authority, clients check name and chain instead of the fingerprint. Applies the next time the service starts.',
  'dienst.tls.schluessel':
    'The private key for the certificate, as PEM. **Readable only by you (0600)**, otherwise the service does not start. Replacing both files takes effect at the next handshake without a restart.',
  'dienst.tls.kette':
    'Intermediate certificates, if they are not included in the certificate. Empty means: none.',
  'dienst.tls.anker':
    'The root of your authority, against which the chain is checked before start. Empty means: the system’s anchors.',
  'grenzen.kontingente':
    'How much a machine access may do per minute over REST and MCP — per access and for all together. **Off is for a controlled environment** where nobody takes anything from anyone else, such as a load run. Access, TLS, the slots per kind of work, the limits on request and response bodies, the timeouts and the lock against guessing stay. Takes effect the next time a service starts.',
}

// netzbereich ist die Überschrift des Bereichs samt Hinweis.
export const netzbereich = {
  netz: 'How this program goes out',
  netzHinweis:
    'These apply to **every** service at once — including the command line, the MCP server and the service. Empty means whatever the environment says applies. What currently applies, and where it comes from, is shown under Network and reachability.',
}

// netzwerte sind die Namen der Werte — QA-08.AK-03 (#1739).
export const netzwerte = {
  'grenzen.kontingente.an': 'On',
  'grenzen.kontingente.aus': 'Off',
}

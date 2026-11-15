// Die Texte der Technikseite über **Dienste und Wege** — welches Zeugnis
// ein Dienst vorzeigt (NET-03, #1707) und wie ein Gegenüber über fremde
// Netze angebunden ist (NET-02, #1706). Aus `seitentechnik.ts` herausgelöst,
// weil die Datei an der Zeilengrenze aus §2 stand.

export const dienstzeugnis = {
  eigenes: 'Certificado propio de {aussteller}, válido hasta {bis}',
  identitaet:
    'Certificado sobre la identidad de este perfil: quien se conecte compara la huella.',
}

export const fernwege = {
  titel: 'Cómo están conectados los interlocutores',
  keine: 'Todavía nadie ha usado esta invitación.',
  direkt: 'Directo, a través de {adresse}',
  relais: 'A través del relay {adresse} — aún sin camino directo',
  aufbau: 'Conocido, pero todavía no accesible por ningún camino',
  unbestimmt: 'Conectados: {anzahl} — la ruta aún no se conoce.',
}

// **Command line and tool server with one click** (OPS-09, #1907).
export const schnittstellen = {
  titel: 'Configurar la línea de comandos y el servidor de herramientas',
  satz: 'Un botón añade lo que un programa de este Mac necesita para encontrar la línea de comandos o el servidor de herramientas; un segundo botón quita exactamente eso. Antes ves qué archivo y qué líneas se tocan; las entradas ajenas se quedan como están.',
  namen: {
    befehle: 'Los enlaces a mse y mse-mcp en ~/.local/bin',
    befehlszeile:
      'La línea de comandos en el PATH del shell de inicio de sesión',
    codex: 'El servidor de herramientas para Codex',
    claudedesktop: 'Claude Desktop',
    claudecode: 'Claude Code',
    dienst: 'El servicio para scripts como LaunchAgent',
  },
  staende: {
    offen: 'sin configurar',
    eingerichtet: 'configurado',
    veraltet: 'desactualizado',
    fremd: 'entrada ajena',
    nichtGefunden: 'no está en este Mac',
    unnoetig: 'no hace falta',
  },
  staendeSaetze: {
    veraltet: 'La aplicación está ahora en otro sitio; Configurar lo repara.',
    fremd: 'Ya hay una entrada ajena — se queda como está.',
    unnoetig: 'Ya está en el PATH.',
  },
  zeilen: 'Se añaden estas líneas',
  vorschau: {
    json: 'Esta entrada se añade al archivo',
    befehl: 'Se ejecuta este comando',
    agent: 'Se crea este archivo y este comando lo carga',
  },
  nurEintrag:
    '{name} había cambiado desde entonces: solo se quitó la entrada propia; el resto queda como lo dejaste.',
  einrichten: 'Configurar',
  widerrufen: 'Revocar',
  alle: {
    titel: 'Todo a la vez',
    satz: 'Configura cada destino que existe en este Mac y que aún falta, o retira todo lo configurado. Las entradas ajenas se quedan como están.',
    vorschau: 'Lo que se escribirá ({anzahl} destinos)',
    vorschauLeer: 'No falta nada: está configurado lo que existe aquí.',
    einrichten: 'Configurar todo',
    widerrufen: 'Retirar todo',
    ergebnis: 'Lo que ocurrió',
    ausgaenge: {
      eingerichtet: 'configurado',
      widerrufen: 'retirado',
      fremd: 'entrada ajena, sin cambios',
      unveraendert: 'sin cambios',
      gescheitert: 'fallido',
      abgebrochen: 'cancelado',
    },
  },
  lauschwahl: {
    titel: 'Dónde escuchan los servicios',
    satz: 'Se aplica al servicio para scripts —iniciado en la ventana y como LaunchAgent— y al servidor de herramientas por HTTP, a partir de su próximo inicio.',
    wahl: 'Accesible a través de',
    schleife: 'Solo este Mac (127.0.0.1)',
    auswahl: 'Direcciones seleccionadas',
    alle: 'Todas las interfaces (0.0.0.0)',
    alleSatz: 'Cada interfaz, también una que se añada después.',
    aussen:
      'Una dirección elegida sale de este Mac: los servicios exigen entonces TLS y un acceso (token), y es posible que macOS pida el permiso de red local.',
    inaktiv: 'apagada',
    ohneAdresse: 'sin dirección',
    liste: 'Interfaces de este Mac',
    arten: {
      schleife: 'Bucle local',
      netz: 'Red',
      tunnel: 'Túnel (VPN)',
      bruecke: 'Puente',
      direkt: 'Conexión inalámbrica directa',
    },
    speichern: 'Aplicar elección',
    gespeichert: 'Aplicado: surte efecto en el próximo inicio de un servicio.',
    keineGewaehlt: 'Elige al menos una dirección.',
    platzhalter: 'como en «Dónde escuchan los servicios»',
    weitere: 'Escucha además en',
    ohne: 'Sin {adresse}: {grund}',
  },
}

// **The network settings** — work item N5 of phase 5.4
// (#1571).
//
// **Eine eigene Datei**, weil `seiteneinstellungen.ts` mit ihnen über die
// Zeilengrenze aus §2 ginge — dieselbe Gestalt wie `einstellungenzugaenge.ts`
// daneben. **Ein Katalog bleibt es trotzdem**: die drei Teile werden dort
// hineingestreut, und der Vertragstest gegen die Go-Seite sieht einen.

// netznamen sind die Beschriftungen der sechs Einstellungen.
export const netznamen = {
  'netz.proxy.http': 'Proxy para http',
  'netz.proxy.https': 'Proxy para https',
  'netz.proxy.aus': 'Acceder sin proxy',
  'netz.frist.verbindung': 'Tiempo límite: conexión (s)',
  'netz.frist.kopf': 'Tiempo límite: cabecera de respuesta (s)',
  'netz.frist.stillstand': 'Tiempo límite: sin avance (s)',
  'netz.schnittstelle': 'Interfaz para la búsqueda en la red local',
  'netz.lauschen': 'Dónde escuchan los servicios propios',
  'dienst.tls.zertifikat': 'Certificado propio de los servicios',
  'dienst.tls.schluessel': 'Clave del certificado propio',
  'dienst.tls.kette': 'Certificados intermedios',
  'dienst.tls.anker': 'Anclas de confianza del certificado',
  'grenzen.kontingente': 'Cuotas de los servicios',
}

// netzhilfe ist der erklärende Satz je Einstellung.
export const netzhilfe = {
  'netz.proxy.http':
    'Por qué proxy pasan las direcciones `http://`. **Vacío significa que rige `HTTP_PROXY` del entorno**: si aquí hay algo, tiene prioridad sobre él. Con esquema, es decir, `http://proxy.company:3128`. **Sin contraseña**: se guardaría sin cifrar en el disco; un proxy con autenticación se toma de `HTTPS_PROXY` del entorno.',
  'netz.proxy.https':
    'Lo mismo para las direcciones `https://`, tomado de `HTTPS_PROXY`. **Por separado**, porque muchas redes tienen proxy solo para uno de los dos.',
  'netz.proxy.aus':
    'Lo que **no** pasa por el proxy: nombres de equipo e intervalos de direcciones, separados por comas, como en `NO_PROXY`. La dirección de bucle local queda fuera de todos modos; un Ollama en **otro** equipo de tu propia red va aquí.',
  'netz.frist.verbindung':
    'Cuánto puede tardar el establecimiento de la conexión, incluido TLS. Vacío significa 15 segundos.',
  'netz.frist.kopf':
    'Cuánto tiempo tiene un servidor para enviar la cabecera de su respuesta. Vacío significa 30 segundos.',
  'netz.frist.stillstand':
    'Cuánto tiempo puede quedarse parada una respuesta en curso. **Este es el que conviene subir** cuando se corta una descarga grande por una línea lenta, no los dos anteriores. Vacío significa 60 segundos.',
  'netz.schnittstelle':
    'Por qué interfaz de red pregunta la búsqueda en la red local y sale el anuncio, por ejemplo `en0`. **Vacío significa automático**: cualquier interfaz adecuada con IPv4 y multicast. Un nombre que ahora mismo no existe cuenta como vacío y queda anotado como hallazgo en el registro.',
  'netz.lauschen':
    'Direcciones en las que escuchan REST y MCP, separadas por comas. **Vacío significa: solo este Mac** (`127.0.0.1`). `0.0.0.0` significa todas las interfaces y va sola. **Si una dirección sale de este equipo, se aplican TLS y token**; es más cómodo elegirla en «Herramientas sin ventana». Surte efecto en el próximo inicio de un servicio.',
  'dienst.tls.zertifikat':
    'El certificado que los servicios REST y MCP presentan fuera del bucle local, en formato PEM. **Vacío significa: el certificado sobre la identidad de este perfil.** Con una autoridad de certificación propia, los clientes comprueban el nombre y la cadena en lugar de la huella. Rige desde el próximo inicio del servicio.',
  'dienst.tls.schluessel':
    'La clave privada del certificado, en formato PEM. **Legible solo por ti (0600)**; de lo contrario, el servicio no se inicia. Sustituir ambos archivos surte efecto sin reiniciar, en el próximo handshake.',
  'dienst.tls.kette':
    'Certificados intermedios, si no están incluidos en el certificado. Vacío significa: ninguno.',
  'dienst.tls.anker':
    'La raíz de tu autoridad, con la que se comprueba la cadena antes del inicio. Vacío significa: las anclas del sistema.',
  'grenzen.kontingente':
    'Cuánto puede hacer por minuto un acceso de máquina a través de REST y MCP — por acceso y para todos juntos. **Desactivadas es para un entorno controlado** en el que nadie le quita nada a nadie, como una prueba de carga. El acceso, TLS, los puestos por tipo de trabajo, los límites del cuerpo y de la respuesta, los plazos y el bloqueo contra quien adivina se mantienen. Surte efecto en el próximo inicio de un servicio.',
}

// netzbereich ist die Überschrift des Bereichs samt Hinweis.
export const netzbereich = {
  netz: 'Cómo sale este programa a la red',
  netzHinweis:
    'Estos ajustes rigen para **todos** los servicios a la vez, también para la línea de comandos, el servidor MCP y el servicio. Vacío significa que rige lo que diga el entorno. Lo que rige ahora mismo y de dónde procede se muestra en Red y conectividad.',
}

// netzwerte sind die Namen der Werte — QA-08.AK-03 (#1739).
export const netzwerte = {
  'grenzen.kontingente.an': 'Activadas',
  'grenzen.kontingente.aus': 'Desactivadas',
}

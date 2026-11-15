import { dateiansicht } from './bausteinedatei'
import { ablage } from './bausteineablage'
import { geste } from './bausteinegeste'
import { laufstand } from './bausteinelauf'
import { objektmarken } from './bausteinemarken'
import { teilen } from './bausteineteilen'

export const bausteine = {
  dateiansicht,
  ablage,
  geste,
  laufstand,
  objektmarken,
  teilen,
  entwurf: {
    fremdGeaendert:
      'Mientras tanto, este campo dice «{wert}»: se cambió en otro lugar. Tu borrador aún no está guardado.',
    behalten: 'Guardar mi borrador',
    nehmen: 'Tomar el valor nuevo',
  },
  listenansicht: {
    suche: 'Buscar',
    suchePlatzhalter: 'Nombre o juego',
    sortierung: 'Orden',
    ansicht: 'Vista',
    liste: 'Lista',
    kacheln: 'Mosaico',
    zahl: '{gezeigt} de {gesamt}',
    nichtsPasst: 'Nada coincide con «{suche}».',
    sucheLeeren: 'Borrar la búsqueda',
    filterZuruecksetzen: 'Restablecer búsqueda y filtros',
    weitere: 'Mostrar {anzahl} más',
    wieImBestand: 'Como en la biblioteca',
    zuletztBenutzt: 'Usados recientemente primero',
    nachName: 'Por nombre',
    nachKarten: 'Por número de cartas',
  },
  speicherstand: {
    schreibt: 'Guardando …',
    geschrieben: 'Todos los cambios están guardados.',
    fehler: '«{feld}» no está guardado — el motivo aparece en el campo.',
  },
  gegenueberTitel: 'Añadir o confirmar una contraparte',
  kontaktwahlEintrag: '{name} · {fingerabdruck}',
  kontaktwahlSuchen: 'Buscar contraparte …',
  feldhilfe: 'Explicación de {feld}',
  mehrZu: 'Más sobre «{satz}»',

  // **Where something came from** — Werkstück 12 of Phase 4.12
  // (#1351).
  herkunftAusProfil: 'Tomado del perfil {quelle}.',
  herkunftAusEinfuhr: 'Importado de {quelle}.',

  pflichtfeldLeer: '{feld} necesita un valor.',

  waehrung: {
    beispiel: 'p. ej. EUR',
    keineGestalt:
      '«{wert}» no es un código de moneda: tiene tres letras, como EUR.',
    grossSchreiben:
      '«{wert}»: un código de moneda se escribe en mayúsculas: {gross}.',
  },

  drucksprachen: {
    beispiel: 'p. ej. en',
    namen: {
      en: 'Inglés',
      es: 'Español',
      fr: 'Francés',
      de: 'Alemán',
      it: 'Italiano',
      pt: 'Portugués',
      ja: 'Japonés',
      ko: 'Coreano',
      ru: 'Ruso',
      zhs: 'Chino simplificado',
      zht: 'Chino tradicional',
      he: 'Hebreo',
      la: 'Latín',
      grc: 'Griego antiguo',
      ar: 'Árabe',
      sa: 'Sánscrito',
      ph: 'Phyrexiano',
      qya: 'Quenya',
      dw: 'Enánico',
      tlh: 'Klingon',
    },
  },

  // **The password field** — piece 6 of phase 4.12 (#1351).
  kennwort: {
    zeigen: 'Mostrar la contraseña en texto claro',
    zeigenKurz: 'Mostrar',
    verbergen: 'Volver a ocultar la contraseña',
    verbergenKurz: 'Ocultar',
  },

  // **The field for the second factor** — piece 7 of phase 4.12 (#1351).
  nachweis: {
    beschriftung: 'Código',
    abschicken: 'Confirmar',
  },

  aktionen: {
    abbrechen: 'Cancelar',
    oeffnen: 'Abrir',
    schliessen: 'Cerrar',
    speichern: 'Guardar',
  },
  stilliste: {
    mitZusatz: '{kennung} · {zusatz}',
    keinerGewaehlt: '— Elige un estilo —',
  },
  druckwahl: {
    drucke: 'Impresiones',
    hinweis:
      'El interlocutor solo ve estas impresiones, con la cantidad que tienes.',
    suchen: 'Buscar impresiones',
    gewaehlt: 'Selección: {anzahl}',
    nurSuche:
      'Tu colección tiene {anzahl}. Busca un nombre para elegir algunas.',
    keiner: 'Ninguna impresión coincide con «{suche}».',
    genauer:
      'Coinciden más impresiones de las que se muestran. Afina la búsqueda.',
    leer: 'Tu colección aún no tiene ninguna impresión que puedas compartir. Añade primero ejemplares.',
    fehlt:
      'Elige al menos una impresión: una selección sin impresiones no comparte nada.',
    behalten: 'Ejemplares de cada impresión que me quedo',
    behaltenHinweis:
      'El interlocutor ve lo que supera esa cantidad: con 1 y tres ejemplares, dos.',
    detailBehalten: 'te quedas con {behalten}',
  },
  suchauswahl: {
    keiner: 'Nada coincide con «{suche}».',
    treffer: 'Resultados: {anzahl}',
    begrenzt:
      'Se muestran {gezeigt} de {anzahl} resultados — acota la búsqueda …',
  },
  zeitzonenwahl: {
    system: 'Como el sistema',
  },
  aktionskatalog: {
    aktionen: {
      anlegen: { kurz: 'Crear', heisst: 'Crea algo nuevo.' },
      bearbeiten: { kurz: 'Editar', heisst: 'Lo modifica en el mismo sitio.' },
      sichern: { kurz: 'Guardar', heisst: 'Conserva el estado actual.' },
      zuruecknehmen: {
        kurz: 'Deshacer',
        heisst: 'Deshace el último paso.',
      },
      duplizieren: {
        kurz: 'Duplicar',
        heisst: 'Pone una segunda copia al lado.',
      },
      kopieren: { kurz: 'Copiar', heisst: 'Lo pone en el portapapeles.' },
      entfernen: {
        kurz: 'Quitar',
        heisst: 'Lo quita de esta vista; en otros sitios se conserva.',
      },
      loeschen: {
        kurz: 'Eliminar',
        heisst: 'Lo elimina definitivamente, en todas partes.',
      },
    },
  },
  nurlesen: {
    hinweis:
      'Estás viendo este perfil como lector: puedes verlo todo y no puedes cambiar nada. Cambia a tu propio perfil para trabajar, o pide a quien administra este que te dé el rol «colaborador».',
    knopf: 'Solo lectura: en este perfil no puedes cambiar nada.',
  },

  // **The context strip** — piece U4 of phase 5.1 (#1415).
  kontext: {
    bezeichnung: 'Contexto de trabajo',
    profilVon: 'de {eigner}',
    rollen: {
      leser: 'Lector',
      mitarbeiter: 'Colaborador',
      verwalter: 'Administrador',
    },
    nurLesen: 'solo lectura',
    zuDenNutzern: 'Cambiar de perfil …',
    zeile: '{name} · {profil}',
  },

  zustaende: {
    laedt: 'Cargando …',
    leer: 'No hay nada',
    fehler: 'No se pudo cargar.',
    fehlerMitGrund: 'No se pudo cargar. {grund}',
    wiederholen: 'Volver a cargar',
    neuLaden: 'Recargar',
  },
  // Find and replace **inside the field** (#1164).
  suche: {
    oeffnen: 'Buscar en el texto',
    schliessen: 'Cerrar la búsqueda',
    suchen: 'Buscar',
    ersetzen: 'Reemplazar por',
    treffer: '{bei} de {ganz}',
    keinTreffer: 'Sin coincidencias',
    weiter: 'Coincidencia siguiente',
    zurueck: 'Coincidencia anterior',
    einsetzen: 'Reemplazar',
    alleEinsetzen: 'Reemplazar todo',
    grossKlein:
      'Se distinguen mayúsculas y minúsculas: en MSE-Script «card» y «Card» son cosas distintas.',
  },
  // Completion **inside** the field (#1164).
  // **Welche Karten hinausgehen** (#1167).
  ausgabeauswahl: {
    aufschrift: 'Qué cartas',
    ganzesSet: 'El set completo ({zahl} cartas)',
    ganzesSetOhneZahl: 'El set completo',
    teilmenge: 'Solo las seleccionadas ({zahl})',
    ohneNamen: 'sin nombre',
    eingrenzen: 'Filtrar cartas',
  },

  vorschlaege: {
    bezeichnung: 'Sugerencias',
    uebernehmen: 'Aceptar con ↹',
  },

  // **The regions of a workspace** — piece U2 of phase 5.1 (#1413).
  arbeitsplatz: {
    breiteVon: 'Anchura de «{region}»',
    breiteWert: '{punkte} puntos',
    navigatorAuf: 'Mostrar el navegador',
    navigatorZu: 'Ocultar el navegador',
    inspectorAuf: 'Mostrar el inspector',
    inspectorZu: 'Ocultar el inspector',
    navigatorEng:
      'Demasiado estrecho para las tres columnas: el navegador se abre sobre el centro',
    inspectorEng:
      'Demasiado estrecho para las tres columnas: el inspector se abre sobre el centro',
  },
  // The command centre (#1423).
  druck: {
    vorlaeufig: 'provisional',
    vorlaeufigErklaerung:
      'Esta impresión aún no tiene identificador de Scryfall. Los precios y las imágenes no aparecerán hasta que se resuelva.',
  },
  kommandozentrale: {
    titel: 'Centro de comandos',
    erklaerung:
      'Un comando es una herramienta del servidor MCP, llamada aquí en la ventana: la misma respuesta que por MCP y en la línea de comandos. Nada pasa por un shell: lo que escribes sigue siendo dato.',
    lesend: 'lee',
    schreibend: 'modifica',
    gefaehrlich: 'difícil de deshacer',
    faehigkeit: 'Capacidad',
    cli: 'Línea de comandos',
    rest: 'Servicio',
    parameter: 'Parámetros',
    keineParameter: 'Este comando no necesita parámetros.',
    pflicht: 'Obligatorio',
    jsonHinweis:
      'Como JSON: una lista entre corchetes, un objeto entre llaves.',
    befundPflicht: 'Este campo necesita un valor.',
    befundZahl: 'Eso no es un número.',
    befundJson: 'Eso no es JSON.',
    vorschau: 'Lo que se enviará',
    bestaetigen: 'He leído la vista previa y quiero ejecutarlo tal cual.',
    ausfuehren: 'Ejecutar',
    abbrechen: 'Cancelar',
    laeuft: 'El comando se está ejecutando …',
    ergebnis: 'Resultado',
    ausgabe: 'Salida del comando',
    alsText: 'Texto',
    alsJson: 'JSON',
    dauer: '{ms} ms',
    verlaufLeer:
      'Aún no se ha ejecutado nada. Los secretos (tokens, contraseñas) nunca aparecen aquí en texto claro.',
    verlauf: 'Historial de esta sesión',
    wiederholen: 'Volver a cargar {name}',
    gelungen: 'correcto',
    gescheitert: 'fallido',
    abgebrochen: 'cancelado',
    schliessen: 'Cerrar',
  },
  // „Diagnose öffnen“ an einem Fehler mit Vorgangskennung (NET-02,
  // S13.NAT-04, #1706).
  diagnoseverweis: {
    oeffnen: 'Abrir diagnóstico',
  },
  // „KI einrichten“ an einer KI-Meldung (AI-01.AK-04, #1742).
  kieinrichtung: {
    oeffnen: 'Configurar la IA',
  },
  kikasten: {
    titel: 'IA para esta carta',
  },
} as const

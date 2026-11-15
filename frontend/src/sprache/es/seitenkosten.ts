// The **cost ledger** — work item O12 of phase 5.2
// (#1439).

export const kosten = {
  titel: 'Lo que puede costar un trabajo externo',
  hinweis:
    'Los trabajos enviados a servicios externos pueden costar dinero. Aquí ' +
    'se muestra lo que está aprobado y cuánto de ello ya está comprometido.',

  kontoWaehlen: 'Cuenta',
  zeitraumWaehlen: 'Periodo',

  zeitraeume: {
    tag: 'Hoy',
    monat: 'Este mes',
  } as Record<string, string>,

  grenze: 'Aprobado',
  reserviert: 'Comprometido',
  verbraucht: 'Gastado',
  unklar: 'Sin aclarar',
  verfuegbar: 'Disponible',

  unklarHinweis:
    'Sin aclarar significa que el trabajo superó un tiempo límite cuando el ' +
    'proveedor ya podía haberlo aceptado. Cuenta como gastado hasta que ' +
    'alguien lo compruebe.',

  ohneGrenze:
    'No hay nada aprobado para esta cuenta. Sin aprobación no sale ningún ' +
    'trabajo de pago.',

  lokalHinweis:
    'Este es un límite de este equipo. Quien use las mismas credenciales en ' +
    'otro sitio consume cuota que este registro de costes desconoce.',

  grenzeSetzen: 'Cambiar la aprobación',
  betragTippen: 'Importe en la unidad más pequeña (céntimos para EUR)',
  waehrungTippen: 'Moneda',
  setzen: 'Aprobar',
  setztGerade: 'Estableciendo …',
  nullNimmtWeg: 'Cero retira la aprobación.',

  postenTitel: 'Los trabajos más recientes',
  keinePosten: 'Todavía no hay nada contabilizado.',
  postenZeile: '{aufgabe} · {stand} · {betrag}',

  staende: {
    reserviert: 'comprometido',
    abgerechnet: 'liquidado',
    freigegeben: 'sin cargo',
    unklar: 'sin aclarar',
  } as Record<string, string>,
}

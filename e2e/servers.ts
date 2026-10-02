// The two production-built servers the browser tests run against.

/** The default configuration: the pile of poo is switched off. */
export const origin = 'http://127.0.0.1:4324';

/** The same build with PLANNINGPOKER_POO_THROWS=true; tests opt in with test.use. */
export const pooOrigin = 'http://127.0.0.1:4325';

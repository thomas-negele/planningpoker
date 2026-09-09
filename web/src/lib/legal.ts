interface LegalAvailability {
  enabled: boolean;
}

/**
 * Ask the running server whether this installation publishes legal notices.
 *
 * The answer comes from the backend rather than from a build-time setting,
 * because the same built frontend is served by every installation — one that has
 * notices and one that does not — and only the process that read the operator's
 * documents knows which of the two it is.
 */
export async function legalNoticesEnabled(): Promise<boolean> {
  const response = await fetch('/api/legal');
  if (!response.ok) {
    throw new Error(`the server did not answer about legal notices (status ${response.status})`);
  }

  const body = (await response.json()) as LegalAvailability;
  if (typeof body.enabled !== 'boolean') {
    throw new Error('the server sent an unusable answer about legal notices');
  }
  return body.enabled;
}

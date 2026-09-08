

interface CreateGameResponse {
  roomId: string;
}

/** Create a room without seating anyone; return its invitation identifier. */
export async function createGame(): Promise<string> {
  const response = await fetch('/api/games', { method: 'POST' });
  if (response.status === 503) {
    // Report capacity separately from other HTTP failures.
    throw new Error(
      'This server is running as many games as it can right now. Please try again in a few minutes.',
    );
  }
  if (!response.ok) {
    throw new Error(`the server refused to start a game (status ${response.status})`);
  }
  const body = (await response.json()) as CreateGameResponse;
  if (!body.roomId) {
    throw new Error('the server started a game but returned no room');
  }
  return body.roomId;
}

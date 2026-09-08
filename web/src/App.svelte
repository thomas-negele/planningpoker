<script lang="ts">
  import EntryScreen from './components/EntryScreen.svelte';
  import RoomView from './components/RoomView.svelte';
  import { router } from './lib/router.svelte';

  const route = $derived(router.route);
</script>

{#if route.name === 'room'}
  <!-- Recreate the room connection when navigation changes the room ID. -->
  {#key route.roomId}
    <RoomView roomId={route.roomId} />
  {/key}
{:else}
  <EntryScreen />
{/if}

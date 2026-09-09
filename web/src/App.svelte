<script lang="ts">
  import EntryScreen from './components/EntryScreen.svelte';
  import LegalFooter from './components/LegalFooter.svelte';
  import RoomView from './components/RoomView.svelte';
  import { router } from './lib/router.svelte';

  const route = $derived(router.route);
</script>

<!--
  The screen fills the space above the footer, and the footer follows it in
  normal flow. Every screen — entry, room, waiting and the refusal screens — is
  inside this one element, so the footer is present on all of them without any
  of them knowing about it.
-->
<div class="screen">
  {#if route.name === 'room'}
    <!-- Recreate the room connection when navigation changes the room ID. -->
    {#key route.roomId}
      <RoomView roomId={route.roomId} />
    {/key}
  {:else}
    <EntryScreen />
  {/if}
</div>

<LegalFooter />

<style>
  /*
   * Grow into the space the footer leaves, but never shrink below the screen's
   * own content: a long seat list on a narrow window makes the page scroll
   * instead of being squeezed into the remaining height.
   */
  .screen {
    flex: 1 0 auto;
    display: flex;
    align-items: center;
    justify-content: center;
  }
</style>

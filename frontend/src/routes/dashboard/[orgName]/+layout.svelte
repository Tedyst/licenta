<script lang="ts">
	import type { LayoutData } from './$types';
	export let data: LayoutData;
</script>

{#if data.organizations === null || data.organizations.length === 0}
	<div class="skeleton max-w-screen-lg h-20 m-2" />
	<div class="skeleton max-w-screen-lg h-20 m-2" />
	<div class="skeleton max-w-screen-lg h-20 m-2" />
	<div class="skeleton max-w-screen-lg h-20 m-2" />
{:else if data.organization === null}
	This organization does not exist or you do not have permission to see it
{:else}
	<slot />
{/if}

<svelte:head>
	<title>Dashboard | Licenta</title>
</svelte:head>

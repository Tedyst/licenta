<script lang="ts">
	import type { LayoutData } from './$types';
	export let data: LayoutData;
</script>

{#if data.organization === null}
	This organization does not exist or you do not have permission to see it
{:else if data.project === null}
	This project does not exist or you do not have permission to see it
{:else}
	<slot />
{/if}

<svelte:head>
	<title>{data.organization?.name} | Dashboard | Licenta</title>
</svelte:head>

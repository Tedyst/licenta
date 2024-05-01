<script lang="ts">
	import type { LayoutData } from './$types';
	export let data: LayoutData;
	import { organizations, currentOrganization } from '$lib/stores';

	$: $currentOrganization = $organizations?.filter((v) => v.name == data.orgName).at(0) || null;
</script>

{#if $organizations === null || $organizations.length === 0}
	<div class="skeleton max-w-screen-lg h-20 m-2" />
	<div class="skeleton max-w-screen-lg h-20 m-2" />
	<div class="skeleton max-w-screen-lg h-20 m-2" />
	<div class="skeleton max-w-screen-lg h-20 m-2" />
{:else if $currentOrganization === null}
	This organization does not exist or you do not have permission to see it
{:else}
	<slot />
{/if}

<svelte:head>
	<title>Dashboard | Licenta</title>
</svelte:head>

<script lang="ts">
	import type { LayoutData } from '../$types';
	import { currentOrganization, currentProject } from '$lib/stores';
	export let data: LayoutData;

	$: $currentProject =
		$currentOrganization?.projects.filter((v) => v.name == data.projectName).at(0) || null;
</script>

{#if $currentOrganization === null}
	This organization does not exist or you do not have permission to see it
{:else if $currentProject === null}
	This project does not exist or you do not have permission to see it
{:else}
	<slot />
{/if}

<svelte:head>
	<title>{$currentOrganization?.name} | Dashboard | Licenta</title>
</svelte:head>

<script lang="ts">
	import { currentOrganization } from '$lib/stores';
</script>

<svelte:head>
	<title>{$currentOrganization.name} | Dashboard | Licenta</title>
</svelte:head>

{#each $currentOrganization.projects as project, i}
	{project.name} {i}
{/each}

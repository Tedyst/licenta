<script lang="ts">
	import type { LayoutData } from '../$types';
	import {
		currentMysqlDatabases,
		currentOrganization,
		currentPostgresDatabases,
		currentProject
	} from '$lib/stores';
	import client from '$lib/client';
	export let data: LayoutData;

	$: $currentProject =
		$currentOrganization?.projects.filter((v) => v.name == data.projectName).at(0) || null;

	$: if ($currentProject)
		client
			.GET('/mysql', { params: { query: { project: $currentProject?.id } } })
			.then((response) => {
				$currentMysqlDatabases = response.data?.mysql_databases || [];
			});
	$: if ($currentProject)
		client
			.GET('/postgres', { params: { query: { project: $currentProject?.id } } })
			.then((response) => {
				$currentPostgresDatabases = response.data?.postgres_databases || [];
			});
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

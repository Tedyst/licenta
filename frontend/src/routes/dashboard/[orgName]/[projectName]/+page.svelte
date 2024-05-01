<script lang="ts">
	import client from '$lib/client';
	import { currentProject, currentMysqlDatabases } from '$lib/stores';
	import { onMount } from 'svelte';

	import MysqlListItem from '$lib/dashboard/MysqlListItem.svelte';
	import CreateMysqlDatabase from '$lib/dashboard/CreateMysqlDatabase.svelte';

	let promise: Promise<void> | null = null;
	onMount(() => {
		if (!$currentProject) return;
		promise = client
			.GET('/mysql', { params: { query: { project: $currentProject?.id } } })
			.then((response) => {
				$currentMysqlDatabases = response.data?.mysql_databases || [];
			});
	});
</script>

project

{#await promise}
	<p>Loading...</p>
{:then}
	{#if $currentMysqlDatabases.length > 0}
		{#each $currentMysqlDatabases as mysqlDatabase}
			<MysqlListItem {mysqlDatabase} />
		{/each}
	{:else}
		<p>No MySQL databases found</p>
	{/if}
{:catch error}
	<p>{error.message}</p>
{/await}

<CreateMysqlDatabase />

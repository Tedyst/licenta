<script lang="ts">
	import { currentMysqlDatabases } from '$lib/stores';

	import MysqlListItem from '$lib/dashboard/MysqlListItem.svelte';
	import CreateMysqlDatabase from '$lib/dashboard/CreateMysqlDatabase.svelte';
</script>

project

{#if $currentMysqlDatabases.length > 0}
	{#each $currentMysqlDatabases as mysqlDatabase}
		<MysqlListItem {mysqlDatabase} />
	{/each}
{:else}
	<p>No MySQL databases found</p>
{/if}

<CreateMysqlDatabase />

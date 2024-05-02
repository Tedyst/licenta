<script lang="ts">
	import { currentMysqlDatabases } from '$lib/stores';

	import MysqlListItem from '$lib/dashboard/MysqlListDatabase.svelte';
	import CreateMysqlDatabase from '$lib/dashboard/MysqlCreateDatabase.svelte';

	let selectedDatabaseType: 'MySQL' | 'PostgreSQL' | 'none' = 'none';
</script>

project

{#each $currentMysqlDatabases as mysqlDatabase}
	<MysqlListItem {mysqlDatabase} />
{/each}

<div class="card w-full md:w-96 bg-base-100 shadow-xl">
	<div class="card-body gap-1">
		<select class="select select-bordered w-full" bind:value={selectedDatabaseType}>
			<option disabled selected value="none">Select database type</option>
			<option value="MySQL">MySQL</option>
			<option value="PostgreSQL">PostgreSQL</option>
		</select>
		{#if selectedDatabaseType === 'MySQL'}
			<CreateMysqlDatabase />
		{/if}
	</div>
</div>

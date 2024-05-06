<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import MysqlListDatabase from '$lib/dashboard/MysqlListDatabase.svelte';
	import MysqlCreateDatabase from '$lib/dashboard/MysqlCreateDatabase.svelte';
	import PostgresCreateDatabase from '$lib/dashboard/PostgresCreateDatabase.svelte';
	import PostgresListDatabase from '$lib/dashboard/PostgresListDatabase.svelte';

	let selectedDatabaseType: 'MySQL' | 'PostgreSQL' | 'none' = 'none';
</script>

project

{#if data.project}
	{#each data?.mysqlDatabases || [] as mysqlDatabase}
		<MysqlListDatabase {mysqlDatabase} project={data.project} />
	{/each}
	{#each data?.postgresDatabases || [] as postgresDatabase}
		<PostgresListDatabase {postgresDatabase} project={data.project} />
	{/each}

	<div class="card w-full md:w-96 bg-base-100 shadow-xl">
		<div class="card-body gap-1">
			<select class="select select-bordered w-full" bind:value={selectedDatabaseType}>
				<option disabled selected value="none">Select database type</option>
				<option value="MySQL">MySQL</option>
				<option value="PostgreSQL">PostgreSQL</option>
			</select>
			{#if selectedDatabaseType === 'MySQL'}
				<MysqlCreateDatabase project={data.project} />
			{:else if selectedDatabaseType === 'PostgreSQL'}
				<PostgresCreateDatabase project={data.project} />
			{/if}
		</div>
	</div>
{/if}

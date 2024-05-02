<script lang="ts">
	import type { components } from '$lib/api/v1';

	import PostgresIcon from '$lib/images/postgresql-icon.svg';
	import PostgresEditDatabase from './PostgresEditDatabase.svelte';
	import client from '$lib/client';
	import { currentPostgresDatabases } from '$lib/stores';
	import { toast } from 'svelte-daisy-toast';
	import BaseListItem from './BaseListItem.svelte';

	export let postgresDatabase: components['schemas']['PostgresDatabase'];

	const deleteDatabase = (id: number) => {
		client.DELETE('/mysql/{id}', { params: { path: { id } } }).then((response) => {
			if (response.error) {
				console.error(response.error);
				return;
			}
			$currentPostgresDatabases = $currentPostgresDatabases.filter((db) => db.id !== id);
			toast({
				closable: true,
				duration: 5000,
				message: 'Database deleted successfully',
				type: 'success'
			});
		});
	};
</script>

<BaseListItem
	databaseUrl={`postgres://${postgresDatabase.username}@****:${postgresDatabase.host}:${postgresDatabase.port}/${postgresDatabase.database_name}`}
	databaseIcon={PostgresIcon}
	databaseID={postgresDatabase.id}
	databaseType="PostgreSQL"
	deleteAction={deleteDatabase}
>
	<div slot="editbox">
		<PostgresEditDatabase {postgresDatabase} />
	</div>
</BaseListItem>

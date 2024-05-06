<script lang="ts">
	export let project: components['schemas']['Project'] | null;
	import type { components } from '$lib/api/v1';

	import PostgresIcon from '$lib/images/postgresql-icon.svg';
	import PostgresEditDatabase from './PostgresEditDatabase.svelte';
	import client from '$lib/client';
	import { toast } from 'svelte-daisy-toast';
	import BaseListItem from './BaseListItem.svelte';
	import { invalidate } from '$app/navigation';

	export let postgresDatabase: components['schemas']['PostgresDatabase'];

	const deleteDatabase = (id: number) => {
		client.DELETE('/mysql/{id}', { params: { path: { id } } }).then((response) => {
			if (response.error) {
				console.error(response.error);
				return;
			}
			invalidate('app:postgres');
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
		<PostgresEditDatabase {postgresDatabase} {project} />
	</div>
</BaseListItem>

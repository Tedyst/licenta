<script lang="ts">
	import type { components } from '$lib/api/v1';
	import client from '$lib/client';
	import { currentProject, currentPostgresDatabases } from '$lib/stores';

	import { toast } from 'svelte-daisy-toast';
	import BaseEditItem from './BaseEditItem.svelte';
	import type { Field } from './BaseEditItem.svelte';

	export let postgresDatabase: components['schemas']['PostgresDatabase'];

	const editAction = (id: number, data: Record<Field, string>) => {
		if (!$currentProject) return;

		client
			.PATCH('/postgres/{id}', {
				params: { path: { id: id } },
				body: {
					database_name: data.database,
					host: data.hostname,
					password: data.password,
					port: parseInt(data.port),
					username: data.username
				}
			})
			.then((response) => {
				if (response.error) {
					toast({
						closable: true,
						duration: 5000,
						message: response.error.message,
						type: 'error'
					});
					return;
				}
				let postgresDatabasesCopy = [...$currentPostgresDatabases];
				let index = postgresDatabasesCopy.findIndex((md) => md.id === postgresDatabase.id);
				postgresDatabasesCopy[index] = response.data?.postgres_database;
				$currentPostgresDatabases = postgresDatabasesCopy;
				toast({
					closable: true,
					duration: 5000,
					message: 'Database edited successfully',
					type: 'success'
				});
			});
	};
</script>

<BaseEditItem
	databaseID={postgresDatabase.id}
	{editAction}
	defaultValues={{
		hostname: postgresDatabase.host,
		port: postgresDatabase.port.toString(),
		username: postgresDatabase.username,
		password: postgresDatabase.password,
		database: postgresDatabase.database_name
	}}
/>

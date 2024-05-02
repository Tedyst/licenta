<script lang="ts">
	import type { components } from '$lib/api/v1';
	import client from '$lib/client';
	import { currentProject, currentMysqlDatabases } from '$lib/stores';

	import { toast } from 'svelte-daisy-toast';
	import BaseEditItem from './BaseEditItem.svelte';
	import type { Field } from './BaseEditItem.svelte';

	export let mysqlDatabase: components['schemas']['MysqlDatabase'];

	const editAction = (id: number, data: Record<Field, string>) => {
		if (!$currentProject) return;

		client
			.PATCH('/mysql/{id}', {
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
				let mysqlDatabasesCopy = [...$currentMysqlDatabases];
				let index = mysqlDatabasesCopy.findIndex((md) => md.id === mysqlDatabase.id);
				mysqlDatabasesCopy[index] = response.data?.mysql_database;
				$currentMysqlDatabases = mysqlDatabasesCopy;
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
	databaseID={mysqlDatabase.id}
	{editAction}
	defaultValues={{
		hostname: mysqlDatabase.host,
		port: mysqlDatabase.port.toString(),
		username: mysqlDatabase.username,
		password: mysqlDatabase.password,
		database: mysqlDatabase.database_name
	}}
/>

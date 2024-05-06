<script lang="ts">
	export let project: components['schemas']['Project'] | null;

	import type { components } from '$lib/api/v1';
	import client from '$lib/client';

	import { toast } from 'svelte-daisy-toast';
	import BaseEditItem from './BaseEditItem.svelte';
	import type { Field } from './BaseEditItem.svelte';
	import { invalidate } from '$app/navigation';

	export let mysqlDatabase: components['schemas']['MysqlDatabase'];

	const editAction = (id: number, data: Record<Field, string>) => {
		if (!project) return;

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
				invalidate('app:mysql');
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

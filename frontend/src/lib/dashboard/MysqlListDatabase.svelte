<script lang="ts">
	export let project: components['schemas']['Project'] | null;
	import type { components } from '$lib/api/v1';

	import MysqlIcon from '$lib/images/mysql-icon.svg';
	import EditMysqlDatabase from './MysqlEditDatabase.svelte';
	import client from '$lib/client';
	import { toast } from 'svelte-daisy-toast';
	import BaseListItem from './BaseListItem.svelte';
	import { invalidate } from '$app/navigation';

	export let mysqlDatabase: components['schemas']['MysqlDatabase'];

	const deleteDatabase = (id: number) => {
		client.DELETE('/mysql/{id}', { params: { path: { id } } }).then((response) => {
			if (response.error) {
				console.error(response.error);
				return;
			}
			invalidate('app:mysql');
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
	databaseUrl={`mysql://${mysqlDatabase.username}@****:${mysqlDatabase.host}:${mysqlDatabase.port}/${mysqlDatabase.database_name}`}
	databaseIcon={MysqlIcon}
	databaseID={mysqlDatabase.id}
	databaseType="MySQL"
	deleteAction={deleteDatabase}
>
	<div slot="editbox">
		<EditMysqlDatabase {mysqlDatabase} {project} />
	</div>
</BaseListItem>

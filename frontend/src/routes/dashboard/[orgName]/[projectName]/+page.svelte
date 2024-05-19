<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import MysqlIcon from '$lib/images/mysql-icon.svg';
	import PostgresIcon from '$lib/images/postgresql-icon.svg';

	import BaseListItem from '$lib/dashboard/BaseListItem.svelte';
</script>

project

{#each data?.mysqlDatabases || [] as mysqlDatabase}
	<BaseListItem
		databaseUrl={`mysql://${mysqlDatabase.username}@****:${mysqlDatabase.host}:${mysqlDatabase.port}/${mysqlDatabase.database_name}`}
		databaseIcon={MysqlIcon}
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-scanner/mysql/?id=${mysqlDatabase.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-scanner/mysql/?id=${mysqlDatabase.id}`}
		databaseType="MySQL"
	/>
{/each}

{#each data?.postgresDatabases || [] as postgresDatabase}
	<BaseListItem
		databaseUrl={`postgres://${postgresDatabase.username}@****:${postgresDatabase.host}:${postgresDatabase.port}/${postgresDatabase.database_name}`}
		databaseIcon={PostgresIcon}
		databaseType="PostgreSQL"
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-scanner/postgresql/?id=${postgresDatabase.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-scanner/postgresql/?id=${postgresDatabase.id}`}
	/>
{/each}

<a
	href="/dashboard/{data.organization?.name}/{data.project?.name}/add-scanner"
	class="btn btn-primary">Add a scanner to the project</a
>

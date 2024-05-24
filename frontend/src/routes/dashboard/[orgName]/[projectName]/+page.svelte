<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import MysqlIcon from '$lib/images/mysql-icon.svg';
	import PostgresIcon from '$lib/images/postgresql-icon.svg';
	import GitIcon from '$lib/images/git-icon.svg';
	import DockerIcon from '$lib/images/docker-icon.svg';

	import BaseListItem from '$lib/dashboard/BaseListItem.svelte';
	import { enhance } from '$app/forms';
	import ScanGroupItem from '$lib/dashboard/ScanGroupItem.svelte';
	import ScanItem from '$lib/dashboard/ScanItem.svelte';
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

{#each data?.gitRepositories || [] as gitRepository}
	<BaseListItem
		databaseUrl={`${gitRepository.git_repository}`}
		databaseIcon={GitIcon}
		databaseType="Git"
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-scanner/postgresql/?id=${gitRepository.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-scanner/postgresql/?id=${gitRepository.id}`}
		viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/view-source/git/?id=${gitRepository.id}`}
	/>
{/each}

{#each data?.dockerImages || [] as dockerImage}
	<BaseListItem
		databaseUrl={dockerImage.docker_image}
		databaseIcon={DockerIcon}
		databaseType="Docker"
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-scanner/postgresql/?id=${dockerImage.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-scanner/postgresql/?id=${dockerImage.id}`}
		viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/view-source/docker/?id=${dockerImage.id}`}
	/>
{/each}

{#each data?.scanGroups || [] as scanGroup}
	<ScanGroupItem {scanGroup}>
		{#each scanGroup?.scans as scan}
			{#if scan.scan_type === 0}
				<ScanItem
					{scan}
					scanIcon={PostgresIcon}
					scanName="Postgres"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 1}
				<ScanItem
					{scan}
					scanIcon={MysqlIcon}
					scanName="MySQL"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 2}
				<ScanItem
					{scan}
					scanIcon={GitIcon}
					scanName="Git"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 3}
				<ScanItem
					{scan}
					scanIcon={DockerIcon}
					scanName="Docker"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{/if}
		{/each}
	</ScanGroupItem>
{/each}

<a
	href="/dashboard/{data.organization?.name}/{data.project?.name}/add-scanner"
	class="btn btn-primary">Add a scanner to the project</a
>

<a
	href="/dashboard/{data.organization?.name}/{data.project?.name}/add-source"
	class="btn btn-primary">Add a source to the project</a
>

<form method="POST" use:enhance action="?/run">
	<input type="hidden" name="projectId" value={data.project?.id} />
	<button type="submit" class="btn btn-primary">Run all sources and scanners</button>
</form>

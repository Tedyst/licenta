<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import MysqlIcon from '$lib/images/mysql-icon.svg';
	import MongoIcon from '$lib/images/mongo-icon.svg';
	import RedisIcon from '$lib/images/redis-icon.svg';
	import PostgresIcon from '$lib/images/postgresql-icon.svg';
	import GitIcon from '$lib/images/git-icon.svg';
	import DockerIcon from '$lib/images/docker-icon.svg';

	import BaseListItem from '$lib/dashboard/BaseListItem.svelte';
	import { enhance } from '$app/forms';
	import ScanGroupItem from '$lib/dashboard/ScanGroupItem.svelte';
	import ScanItem from '$lib/dashboard/ScanItem.svelte';
</script>

<div class="hero bg-base-200">
	<div class="hero-content text-center">
		<div class="max-w-md">
			<h1 class="text-5xl font-bold">Projects</h1>
			<p class="py-6">
				Projects are the main way to organize your scans. You can add scanned databases and secret
				sources to a project, and then run scans on them. Scans are grouped by the type of source
				they were run on. You can view the results of each scan by clicking on the scan.
			</p>
		</div>
	</div>
</div>

<div class="divider">Project Actions</div>
<div class="flex flex-col gap-2 lg:justify-center lg:flex-row justify-stretch flex-wrap">
	<a
		href="/dashboard/{data.organization?.name}/{data.project?.name}/add-scanner"
		class="btn btn-primary">Add a Scanned Database to the project</a
	>

	<a
		href="/dashboard/{data.organization?.name}/{data.project?.name}/add-source"
		class="btn btn-primary">Add a Secret Source to the project</a
	>

	<form method="POST" use:enhance action="?/run" class="grow flex lg:grow-0">
		<input type="hidden" name="projectId" value={data.project?.id} />
		<button type="submit" class="btn btn-primary grow lg:grow-0"
			>Run all sources and scanners</button
		>
	</form>

	<form method="POST" use:enhance action="?/toggle_remote" class="grow flex lg:grow-0">
		<input type="hidden" name="projectId" value={data.project?.id} />
		<input type="hidden" name="remote" value={!data.project?.remote} />
		<button type="submit" class="btn btn-primary grow lg:grow-0"
			>{!data.project?.remote ? 'Enable' : 'Disable'} the external workers for this project</button
		>
	</form>
</div>

<div class="divider">Scanned Databases</div>
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

{#each data?.mongoDatabases || [] as mongoDatabase}
	<BaseListItem
		databaseUrl={`mongodb://${mongoDatabase.username}@****:${mongoDatabase.host}:${mongoDatabase.port}/${mongoDatabase.database_name}`}
		databaseIcon={MongoIcon}
		databaseType="MongoDB"
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-scanner/mongo/?id=${mongoDatabase.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-scanner/mongo/?id=${mongoDatabase.id}`}
	/>
{/each}

{#each data?.redisDatabases || [] as redisDatabase}
	<BaseListItem
		databaseUrl={`redis://${redisDatabase.username}@****:${redisDatabase.host}:${redisDatabase.port}/0}`}
		databaseIcon={RedisIcon}
		databaseType="Redis"
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-scanner/redis/?id=${redisDatabase.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-scanner/redis/?id=${redisDatabase.id}`}
	/>
{/each}

<div class="divider">Secret Sources</div>
{#each data?.gitRepositories || [] as gitRepository}
	<BaseListItem
		databaseUrl={`${gitRepository.git_repository}`}
		databaseIcon={GitIcon}
		databaseType="Git"
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-source/git/?id=${gitRepository.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-source/git/?id=${gitRepository.id}`}
		viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/view-source/git/?id=${gitRepository.id}`}
	/>
{/each}

{#each data?.dockerImages || [] as dockerImage}
	<BaseListItem
		databaseUrl={dockerImage.docker_image}
		databaseIcon={DockerIcon}
		databaseType="Docker"
		deleteURL={`/dashboard/${data.organization?.name}/${data.project?.name}/delete-source/docker/?id=${dockerImage.id}`}
		editURL={`/dashboard/${data.organization?.name}/${data.project?.name}/edit-source/docker/?id=${dockerImage.id}`}
		viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/view-source/docker/?id=${dockerImage.id}`}
	/>
{/each}

<div class="divider">Scans</div>
{#each data?.scanGroups || [] as scanGroup}
	<ScanGroupItem {scanGroup}>
		{#each scanGroup?.scans as scan}
			{#if scan.scan_type === 1}
				<ScanItem
					{scan}
					scanIcon={PostgresIcon}
					scanName="Postgres"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 2}
				<ScanItem
					{scan}
					scanIcon={MysqlIcon}
					scanName="MySQL"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 3}
				<ScanItem
					{scan}
					scanIcon={GitIcon}
					scanName="Git"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 4}
				<ScanItem
					{scan}
					scanIcon={DockerIcon}
					scanName="Docker"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 5}
				<ScanItem
					{scan}
					scanIcon={RedisIcon}
					scanName="Redis"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{:else if scan.scan_type === 6}
				<ScanItem
					{scan}
					scanIcon={MongoIcon}
					scanName="MongoDB"
					viewURL={`/dashboard/${data.organization?.name}/${data.project?.name}/scan/?id=${scan.id}`}
				/>
			{/if}
		{/each}
	</ScanGroupItem>
{/each}

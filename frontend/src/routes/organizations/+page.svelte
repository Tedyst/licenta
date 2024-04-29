<script lang="ts">
	import { organizations } from '$lib/stores';
</script>

{#if $organizations === null}
	<div class="skeleton h-36 m-0 ml-4 mt-2" />
	<div class="skeleton h-36 m-0 ml-4 mt-2" />
{:else}
	{#each $organizations as organization}
		<a href="/dashboard/{organization.name}">
			<div class="card ml-4 h-30 max-w-full mt-2 bg-base-100">
				<div class="card-body flex-row pr-4">
					<div class="flex flex-grow align-middle items-start flex-col self-center">
						Organization
						<div class="ml-4 text-2xl text-info">{organization.name}</div>
					</div>
					<div
						class="mr-0 inline-grid grid-cols-1 grid-rows-3 sm:grid-cols-2 sm:grid-rows-2 md:grid-cols-4 md:grid-rows-1 grow-0 gap-2"
					>
						<div class="inline-grid w-max gap-4">
							<div class="stat-title text-sm">Users</div>
							<div class="stat-value text-primary text-base">{organization.stats?.users}</div>
						</div>
						<div class="inline-grid w-max gap-4">
							<div class="stat-title text-sm">Projects</div>
							<div class="stat-value text-secondary text-base">{organization.stats?.projects}</div>
						</div>
						<div class="inline-grid w-max gap-4">
							<div class="stat-title text-sm">Scans</div>
							<div class="stat-value text-success text-base">{organization.stats?.scans}</div>
						</div>
					</div>
				</div>
			</div>
		</a>
	{/each}
{/if}

<a href="/organizations/create">
	<div class="card ml-4 h-30 max-w-full bg-base-100 outline-dotted outline-secondary mt-4">
		<div class="card-body flex-row pr-4">
			<div class="flex flex-grow align-middle items-start flex-col self-center">
				<div class="ml-4 text-2xl text-info">Create a new organization</div>
			</div>
		</div>
	</div>
</a>

<svelte:head>
	<title>Organizations | Licenta</title>
</svelte:head>

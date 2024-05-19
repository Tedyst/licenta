<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import ListMembers from '$lib/utils/ListMembers.svelte';
	import OrganizationProject from '$lib/utils/OrganizationProject.svelte';
</script>

<svelte:head>
	<title>{data.organization?.name} | Dashboard | Licenta</title>
</svelte:head>

{#if data.organization === null}
	This organization does not exist or you do not have permission to see it
{:else}
	<div class="hero bg-base-200">
		<div class="hero-content text-center">
			<div class="max-w-md">
				<h1 class="text-5xl font-bold">{data.organization.name}</h1>
				<p class="py-6">
					Here are all the projects that belong to this organization. You can create new projects
					and manage existing ones.
				</p>
			</div>
		</div>
	</div>

	{#each data.organization.projects as project}
		<OrganizationProject organization={data.organization} {project} />
	{/each}

	<a href="/dashboard/{data.organization.name}/create">
		<div class="card ml-4 h-30 max-w-full bg-base-100 outline-dotted outline-secondary mt-4">
			<div class="card-body flex-row pr-4">
				<div class="flex flex-grow align-middle items-start flex-col self-center">
					<div class="ml-4 text-2xl text-info">Create a new project</div>
				</div>
			</div>
		</div>
	</a>

	<div class="divider" />

	<div class="flex-col flex lg:flex-row w-full">
		<div class="grid flex-grow place-content-center h-auto flex-1">
			<h1 class="text-3xl font-bold">Organization Settings</h1>
			<div class="form-control mt-6">
				<a href="/dashboard/{data.organization.name}/delete" class="btn btn-error"
					>Delete organization</a
				>
			</div>
		</div>
		<div class="divider divider-horizontal" />
		<ListMembers organization={data.organization} />
	</div>
{/if}

<script lang="ts">
	import client, { updateOrganizations } from '$lib/client';
	import { currentOrganization } from '$lib/stores';
	import ListMembers from '$lib/utils/ListMembers.svelte';
	import OrganizationProject from '$lib/utils/OrganizationProject.svelte';
	import OrganizationSettings from '$lib/utils/OrganizationSettings.svelte';

	let editRoleAction = (role: 'Owner' | 'Admin' | 'Viewer' | 'None', userId: number) => {
		console.log(role, userId);
	};

	let addUserAction = (email: string) => {
		console.log(email);
	};

	let deleteUserAction = (userId: number) => {
		if (!$currentOrganization) return;
		client
			.DELETE('/organizations/{id}/delete-user', {
				params: { path: { id: $currentOrganization?.id } },
				body: { id: userId }
			})
			.then((res) => {
				if (res.data?.success) {
					updateOrganizations();
				}
			});
	};
</script>

<svelte:head>
	<title>{$currentOrganization?.name} | Dashboard | Licenta</title>
</svelte:head>

{#if $currentOrganization === null}
	This organization does not exist or you do not have permission to see it
{:else}
	<div class="hero bg-base-200">
		<div class="hero-content text-center">
			<div class="max-w-md">
				<h1 class="text-5xl font-bold">{$currentOrganization.name}</h1>
				<p class="py-6">
					Here are all the projects that belong to this organization. You can create new projects
					and manage existing ones.
				</p>
			</div>
		</div>
	</div>

	{#each $currentOrganization.projects as project, i}
		<OrganizationProject organization={$currentOrganization} {project} />
	{/each}

	<a href="/dashboard/{$currentOrganization.name}/create">
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
		<OrganizationSettings organization={$currentOrganization} />
		<div class="divider divider-horizontal" />
		<ListMembers {editRoleAction} {addUserAction} {deleteUserAction} />
	</div>
{/if}

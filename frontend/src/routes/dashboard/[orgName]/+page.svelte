<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import ListMembers from '$lib/utils/ListMembers.svelte';
	import OrganizationProject from '$lib/utils/OrganizationProject.svelte';
	import OrganizationSettings from '$lib/utils/OrganizationSettings.svelte';
	import { toast } from 'svelte-daisy-toast';
	import { invalidate } from '$app/navigation';
	import client from '$lib/client';

	let editRoleAction = (role: 'Owner' | 'Admin' | 'Viewer' | 'None', userId: number) => {
		if (!data.organization) return;
		client
			.POST('/organizations/{id}/edit-user', {
				params: { path: { id: data.organization?.id } },
				body: { id: userId, role }
			})
			.then((res) => {
				if (res.data?.success) {
					invalidate('app:organizationinfo');
					toast({
						closable: true,
						duration: 5000,
						message: 'User role edited successfully',
						title: 'Success',
						type: 'success'
					});
				} else {
					toast({
						closable: true,
						duration: 5000,
						message: res.error?.message || 'Could not edit user role',
						title: 'Error',
						type: 'error'
					});
				}
			});
	};

	let addUserAction = (email: string) => {
		if (!data.organization) return;
		client
			.POST('/organizations/{id}/add-user', {
				params: { path: { id: data.organization?.id } },
				body: { email }
			})
			.then((res) => {
				if (res.data?.success) {
					invalidate('app:organizationinfo');
					toast({
						closable: true,
						duration: 5000,
						message: 'User added successfully',
						title: 'Success',
						type: 'success'
					});
				} else {
					toast({
						closable: true,
						duration: 5000,
						message: res.error?.message || 'Could not add user',
						title: 'Error',
						type: 'error'
					});
				}
			});
	};

	let deleteUserAction = (userId: number) => {
		if (!data.organization) return;
		client
			.DELETE('/organizations/{id}/delete-user', {
				params: { path: { id: data.organization?.id } },
				body: { id: userId }
			})
			.then((res) => {
				invalidate('app:organizationinfo');
				if (res.response.status === 204) {
					toast({
						closable: true,
						duration: 5000,
						message: 'User deleted successfully',
						title: 'Success',
						type: 'success'
					});
				} else {
					toast({
						closable: true,
						duration: 5000,
						message: res.error?.message || 'Could not delete user',
						title: 'Error',
						type: 'error'
					});
				}
			});
	};
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
		<OrganizationSettings organization={data.organization} />
		<div class="divider divider-horizontal" />
		<ListMembers
			{editRoleAction}
			{addUserAction}
			{deleteUserAction}
			members={data.organization.members}
			currentUser={data.user}
		/>
	</div>
{/if}

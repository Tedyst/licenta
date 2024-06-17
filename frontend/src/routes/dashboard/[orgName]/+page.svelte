<script lang="ts">
	import type { PageData, ActionData } from './$types';
	export let data: PageData;
	export let form: ActionData;

	import ListMembers from '$lib/utils/ListMembers.svelte';
	import OrganizationProject from '$lib/utils/OrganizationProject.svelte';
	import { enhance } from '$app/forms';
</script>

<svelte:head>
	<title>{data.organization?.name} | Dashboard | Licenta</title>
</svelte:head>

{#if data.organization === null}
	This organization does not exist or you do not have permission to see it
{:else}
	{#if form?.error}
		<div class="alert alert-error">
			{form.error}
		</div>
	{/if}

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
		<div class="grid flex-grow place-content-center h-auto flex-1 w-full">
			<h1 class="text-3xl font-bold">Organization Settings</h1>
			<div class="form-control mt-6">
				<a href="/dashboard/{data.organization.name}/delete" class="btn btn-error"
					>Delete organization</a
				>
			</div>

			<div class="divider">Workers</div>

			{#if data.workers.length !== 0}
				<table class="table">
					<thead>
						<tr>
							<th>Worker Name</th>
							<th>Token</th>
							<th>Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each data.workers as worker}
							<tr>
								<td>{worker.name}</td>
								<td>{worker.token}</td>
								<td>
									<form method="POST" use:enhance action="?/delete_worker">
										<input type="hidden" name="workerId" value={worker.id} />
										<button type="submit" class="btn btn-error">Delete</button>
									</form>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			{/if}

			<form class="form-control mt-1" method="POST" use:enhance action="?/create_worker">
				<input type="hidden" name="organizationId" value={data.organization.id} />
				<div class="form-control mt-4">
					<input
						type="text"
						name="workerName"
						placeholder="Worker name"
						class="input input-primary"
					/>
				</div>
				<button type="submit" class="btn btn-info mt-1">Create new Worker</button>
			</form>
		</div>

		<div class="divider divider-horizontal" />
		<ListMembers organization={data.organization} />
	</div>
{/if}

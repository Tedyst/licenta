<script lang="ts">
	import type { PageData } from './$types';
	export let data: PageData;

	import client from '$lib/client';
	import { goto, invalidate } from '$app/navigation';

	let projectName = '';
	let error = '';

	function createProject() {
		if (data.organization === null) {
			error = 'Organization not found';
			return;
		}
		client
			.POST('/projects', {
				body: { name: projectName.toLowerCase(), organization_id: data.organization.id }
			})
			.then(async (res) => {
				await invalidate('app:organizationinfo');
				if (res.data?.success) {
					await goto(`/dashboard/${data.organization?.name}/${res.data.project.name}`);
				} else {
					error = res.error?.message || 'Internal server error';
				}
			})
			.catch((err) => {
				console.log(err);
				error = err.toString();
			});
	}
</script>

<svelte:head>
	<title>Create Project | {data.organization?.name} | Dashboard | Licenta</title>
</svelte:head>

<div class="hero bg-base-200">
	<div class="hero-content flex-col">
		<div class="text-center lg:text-left">
			<p class="py-6">
				Projects represent a collection of databases and secret sources. You can create multiple
				projects under an organization.
			</p>
		</div>
		<div class="card shrink-0 w-full max-w-sm shadow-2xl bg-base-100">
			<form
				class="card-body"
				on:submit|preventDefault={createProject}
				on:input={() => (error = '')}
			>
				<div class="form-control">
					<label class="label" for="projectName">
						<span class="label-text">Project Name</span>
					</label>
					<input
						type="text"
						id="projectName"
						placeholder="Name"
						class="lowercase input input-bordered transition-colors duration-300 ease-in-out {error
							? 'wiggle input-error'
							: ''}"
						required
						bind:value={projectName}
					/>
					{#if error}
						<div class="label text-error text-xs">
							{error}
						</div>
					{/if}
				</div>
				<div class="form-control mt-6">
					<button class="btn btn-primary">Create a new Project</button>
				</div>
			</form>
		</div>
	</div>
</div>

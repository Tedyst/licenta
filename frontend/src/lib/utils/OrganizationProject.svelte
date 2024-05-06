<script lang="ts">
	import { goto } from '$app/navigation';
	import type { components } from '$lib/api/v1';
	import client from '$lib/client';
	import { toast } from 'svelte-daisy-toast';
	import TrashCan from 'svelte-material-icons/TrashCan.svelte';

	export let organization: components['schemas']['Organization'];
	export let project: components['schemas']['Project'];

	let dialog: HTMLDialogElement;

	let deleteProject = () => {
		client.DELETE('/projects/{id}', { params: { path: { id: project.id } } }).then(async (res) => {
			if (res.response.status === 204) {
				await goto(`/dashboard/${organization.name}`);
				toast({
					closable: true,
					duration: 5000,
					message: 'Project deleted successfully',
					title: 'Success',
					type: 'success'
				});
			} else {
				toast({
					closable: true,
					duration: 5000,
					message: res.error?.message || 'Could not delete project',
					title: 'Error',
					type: 'error'
				});
			}
		});
	};
</script>

<a href="/dashboard/{organization.name}/{project.name}">
	<div class="card ml-4 h-30 max-w-full mt-2 bg-base-100">
		<div class="card-body flex-row pr-4">
			<div class="flex flex-grow align-middle items-start flex-col self-center">
				Project
				<div class="ml-4 text-2xl text-info">{project.name}</div>
			</div>
			<div class="mr-0 inline-grid grid-cols-2 grid-rows-1 grow-0 gap-2">
				<div class="inline-grid w-max gap-4">
					<div class="stat-title text-sm">Scans</div>
					<div class="stat-value text-success text-base">0</div>
				</div>
			</div>
			<div class="divider divider-horizontal" />
			<button
				type="button"
				class="mr-5 inline place-content-center text-red-500"
				on:click|preventDefault={() => dialog?.showModal()}
			>
				<TrashCan size={25} />
			</button>
		</div>
	</div>
</a>

<dialog class="modal" bind:this={dialog}>
	<div class="modal-box">
		<h3 class="font-bold text-lg">
			Are you sure that you want to delete the project <b>{project.name}</b>?
		</h3>
		<p class="py-4">Press ESC key or click the button below to close</p>
		<div class="modal-action">
			<form method="dialog">
				<button class="btn">No</button>
				<button class="btn bg-red-500 text-black" on:click={deleteProject}>Yes</button>
			</form>
		</div>
	</div>
</dialog>
